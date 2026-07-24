package k8s

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/konstpic/externaldns-web-ui/backend/internal/models"
)

const (
	annHostnameLegacy   = "external-dns.alpha.kubernetes.io/hostname"
	annHostname         = "external-dns.kubernetes.io/hostname"
	annInternalHostname = "external-dns.kubernetes.io/internal-hostname"
	annTTL              = "external-dns.kubernetes.io/ttl"
)

type Client struct {
	clientset     kubernetes.Interface
	dynamic       dynamic.Interface
	namespace     string
	deployment    string
	clusterName   string
	domainFilter  string
}

func NewFromEnv() (*Client, error) {
	cfg, err := restConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	ns := envOr("EXTERNALDNS_NAMESPACE", "external-dns")
	return &Client{
		clientset:    clientset,
		dynamic:      dyn,
		namespace:    ns,
		deployment:   envOr("EXTERNALDNS_DEPLOYMENT", "external-dns"),
		clusterName:  envOr("CLUSTER_NAME", "cluster"),
		domainFilter: envOr("DOMAIN_FILTER", ""),
	}, nil
}

func restConfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home, err := os.UserHomeDir(); err == nil {
			kubeconfig = home + "/.kube/config"
		}
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (c *Client) Overview(ctx context.Context) (*models.Overview, error) {
	records, err := c.ListRecords(ctx, "")
	if err != nil {
		return nil, err
	}

	status, err := c.ControllerStatus(ctx)
	if err != nil {
		return nil, err
	}

	svcCount, ingCount, crdCount := 0, 0, 0
	nsSet := map[string]struct{}{}
	for _, r := range records {
		nsSet[r.Namespace] = struct{}{}
		switch r.SourceType {
		case "Service":
			svcCount++
		case "Ingress":
			ingCount++
		case "DNSEndpoint":
			crdCount++
		}
	}

	domainFilter := c.domainFilter
	if domainFilter == "" && len(status.DomainFilters) > 0 {
		domainFilter = strings.Join(status.DomainFilters, ", ")
	}

	return &models.Overview{
		TotalRecords:    len(records),
		ServiceSources:  svcCount,
		IngressSources:  ingCount,
		DNSEndpointCRDs: crdCount,
		Namespaces:      len(nsSet),
		Controller:      *status,
		ClusterName:     c.clusterName,
		DomainFilter:    domainFilter,
	}, nil
}

func (c *Client) ListRecords(ctx context.Context, query string) ([]models.DNSRecord, error) {
	var out []models.DNSRecord

	services, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	for _, svc := range services.Items {
		out = append(out, recordsFromService(svc)...)
	}

	ingresses, err := c.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list ingresses: %w", err)
	}
	for _, ing := range ingresses.Items {
		out = append(out, recordsFromIngress(ing)...)
	}

	crdRecords, err := c.listDNSEndpointRecords(ctx)
	if err == nil {
		out = append(out, crdRecords...)
	}

	if query != "" {
		q := strings.ToLower(query)
		filtered := out[:0]
		for _, r := range out {
			if strings.Contains(strings.ToLower(r.Hostname), q) ||
				strings.Contains(strings.ToLower(r.Namespace), q) ||
				strings.Contains(strings.ToLower(r.Resource), q) ||
				strings.Contains(strings.ToLower(r.Target), q) {
				filtered = append(filtered, r)
			}
		}
		out = filtered
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Hostname != out[j].Hostname {
			return out[i].Hostname < out[j].Hostname
		}
		return out[i].Namespace < out[j].Namespace
	})

	return out, nil
}

func (c *Client) ListSources(ctx context.Context, kind string) ([]models.SourceResource, error) {
	var out []models.SourceResource

	if kind == "" || kind == "Service" {
		services, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, svc := range services.Items {
			hosts := hostnamesFromAnnotations(svc.Annotations)
			if len(hosts) == 0 {
				continue
			}
			out = append(out, models.SourceResource{
				Kind:        "Service",
				Namespace:   svc.Namespace,
				Name:        svc.Name,
				Hostnames:   hosts,
				Target:      serviceTarget(svc),
				Annotations: filterDNSAnnotations(svc.Annotations),
				Age:         formatAge(svc.CreationTimestamp.Time),
			})
		}
	}

	if kind == "" || kind == "Ingress" {
		ingresses, err := c.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, ing := range ingresses.Items {
			hosts := ingressHostnames(ing)
			if len(hosts) == 0 {
				continue
			}
			out = append(out, models.SourceResource{
				Kind:        "Ingress",
				Namespace:   ing.Namespace,
				Name:        ing.Name,
				Hostnames:   hosts,
				Target:      ingressTarget(ing),
				Annotations: filterDNSAnnotations(ing.Annotations),
				Age:         formatAge(ing.CreationTimestamp.Time),
			})
		}
	}

	if kind == "" || kind == "DNSEndpoint" {
		crdSources, err := c.listDNSEndpointSources(ctx)
		if err == nil {
			out = append(out, crdSources...)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].Namespace != out[j].Namespace {
			return out[i].Namespace < out[j].Namespace
		}
		return out[i].Name < out[j].Name
	})

	return out, nil
}

func (c *Client) ControllerStatus(ctx context.Context) (*models.ControllerStatus, error) {
	status := &models.ControllerStatus{
		Namespace:  c.namespace,
		Deployment: c.deployment,
	}

	dep, err := c.clientset.AppsV1().Deployments(c.namespace).Get(ctx, c.deployment, metav1.GetOptions{})
	if err != nil {
		return status, nil
	}

	status.Replicas = dep.Status.Replicas
	status.ReadyReplicas = dep.Status.ReadyReplicas
	status.Ready = dep.Status.ReadyReplicas > 0 && dep.Status.ReadyReplicas == dep.Status.Replicas

	if len(dep.Spec.Template.Spec.Containers) > 0 {
		container := dep.Spec.Template.Spec.Containers[0]
		status.Image = container.Image
		parseExternalDNSArgs(container.Args, container.Env, status)
	}

	pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(dep.Spec.Selector),
	})
	if err == nil {
		for _, p := range pods.Items {
			status.PodNames = append(status.PodNames, p.Name)
		}
	}

	return status, nil
}

func (c *Client) RecentLogs(ctx context.Context, tail int64) ([]models.LogLine, error) {
	status, err := c.ControllerStatus(ctx)
	if err != nil || len(status.PodNames) == 0 {
		return nil, err
	}

	podName := status.PodNames[0]
	if tail <= 0 {
		tail = 100
	}

	raw, err := c.clientset.CoreV1().Pods(c.namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tail,
	}).DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	var lines []models.LogLine
	now := time.Now()
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, models.LogLine{Timestamp: now, Message: line})
	}
	return lines, nil
}

func (c *Client) listDNSEndpointRecords(ctx context.Context) ([]models.DNSRecord, error) {
	gvr := schema.GroupVersionResource{Group: "externaldns.k8s.io", Version: "v1alpha1", Resource: "dnsendpoints"}
	list, err := c.dynamic.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var out []models.DNSRecord
	for _, item := range list.Items {
		ns := item.GetNamespace()
		name := item.GetName()
		created := item.GetCreationTimestamp().Time

		endpoints, found, _ := unstructured.NestedSlice(item.Object, "spec", "endpoints")
		if !found {
			continue
		}
		for i, ep := range endpoints {
			m, ok := ep.(map[string]interface{})
			if !ok {
				continue
			}
			dnsName, _ := m["dnsName"].(string)
			recordType, _ := m["recordType"].(string)
			targets, _ := m["targets"].([]interface{})
			target := ""
			if len(targets) > 0 {
				target, _ = targets[0].(string)
			}
			var ttl *int64
			if v, ok := m["recordTTL"].(int64); ok {
				ttl = &v
			} else if v, ok := m["recordTTL"].(float64); ok {
				i := int64(v)
				ttl = &i
			}
			if dnsName == "" {
				continue
			}
			if recordType == "" {
				recordType = "A"
			}
			out = append(out, models.DNSRecord{
				ID:         fmt.Sprintf("dnsendpoint/%s/%s/%d", ns, name, i),
				Hostname:   dnsName,
				RecordType: recordType,
				Target:     target,
				TTL:        ttl,
				SourceType: "DNSEndpoint",
				Namespace:  ns,
				Resource:   name,
				CreatedAt:  &created,
			})
		}
	}
	return out, nil
}

func (c *Client) listDNSEndpointSources(ctx context.Context) ([]models.SourceResource, error) {
	gvr := schema.GroupVersionResource{Group: "externaldns.k8s.io", Version: "v1alpha1", Resource: "dnsendpoints"}
	list, err := c.dynamic.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var out []models.SourceResource
	for _, item := range list.Items {
		hosts := []string{}
		endpoints, _, _ := unstructured.NestedSlice(item.Object, "spec", "endpoints")
		for _, ep := range endpoints {
			m, ok := ep.(map[string]interface{})
			if !ok {
				continue
			}
			if dnsName, _ := m["dnsName"].(string); dnsName != "" {
				hosts = append(hosts, dnsName)
			}
		}
		if len(hosts) == 0 {
			continue
		}
		out = append(out, models.SourceResource{
			Kind:      "DNSEndpoint",
			Namespace: item.GetNamespace(),
			Name:      item.GetName(),
			Hostnames: hosts,
			Age:       formatAge(item.GetCreationTimestamp().Time),
		})
	}
	return out, nil
}

func recordsFromService(svc corev1.Service) []models.DNSRecord {
	hosts := hostnamesFromAnnotations(svc.Annotations)
	if len(hosts) == 0 {
		return nil
	}

	target := serviceTarget(svc)
	ttl := parseTTL(svc.Annotations)
	created := svc.CreationTimestamp.Time
	anns := filterDNSAnnotations(svc.Annotations)

	var out []models.DNSRecord
	for i, h := range hosts {
		recordType := "A"
		if strings.HasPrefix(h, "*.") {
			recordType = "CNAME"
		}
		out = append(out, models.DNSRecord{
			ID:          fmt.Sprintf("service/%s/%s/%d", svc.Namespace, svc.Name, i),
			Hostname:    h,
			RecordType:  recordType,
			Target:      target,
			TTL:         ttl,
			SourceType:  "Service",
			Namespace:   svc.Namespace,
			Resource:    svc.Name,
			Annotations: anns,
			CreatedAt:   &created,
		})
	}
	return out
}

func recordsFromIngress(ing networkingv1.Ingress) []models.DNSRecord {
	hosts := ingressHostnames(ing)
	if len(hosts) == 0 {
		return nil
	}

	target := ingressTarget(ing)
	ttl := parseTTL(ing.Annotations)
	created := ing.CreationTimestamp.Time
	anns := filterDNSAnnotations(ing.Annotations)

	var out []models.DNSRecord
	for i, h := range hosts {
		out = append(out, models.DNSRecord{
			ID:          fmt.Sprintf("ingress/%s/%s/%d", ing.Namespace, ing.Name, i),
			Hostname:    h,
			RecordType:  "A",
			Target:      target,
			TTL:         ttl,
			SourceType:  "Ingress",
			Namespace:   ing.Namespace,
			Resource:    ing.Name,
			Annotations: anns,
			CreatedAt:   &created,
		})
	}
	return out
}

func hostnamesFromAnnotations(ann map[string]string) []string {
	if ann == nil {
		return nil
	}
	var hosts []string
	for _, key := range []string{annHostname, annHostnameLegacy, annInternalHostname} {
		if v := ann[key]; v != "" {
			hosts = append(hosts, splitHostnames(v)...)
		}
	}
	return uniqueStrings(hosts)
}

func ingressHostnames(ing networkingv1.Ingress) []string {
	hosts := hostnamesFromAnnotations(ing.Annotations)
	for _, rule := range ing.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, rule.Host)
		}
	}
	for _, tls := range ing.Spec.TLS {
		hosts = append(hosts, tls.Hosts...)
	}
	return uniqueStrings(hosts)
}

func serviceTarget(svc corev1.Service) string {
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		ing := svc.Status.LoadBalancer.Ingress[0]
		if ing.IP != "" {
			return ing.IP
		}
		return ing.Hostname
	}
	if len(svc.Spec.ExternalIPs) > 0 {
		return svc.Spec.ExternalIPs[0]
	}
	if svc.Spec.ClusterIP != "" && svc.Spec.ClusterIP != "None" {
		return svc.Spec.ClusterIP
	}
	return "—"
}

func ingressTarget(ing networkingv1.Ingress) string {
	for _, lb := range ing.Status.LoadBalancer.Ingress {
		if lb.IP != "" {
			return lb.IP
		}
		if lb.Hostname != "" {
			return lb.Hostname
		}
	}
	return "—"
}

func parseTTL(ann map[string]string) *int64 {
	if ann == nil {
		return nil
	}
	if v := ann[annTTL]; v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return &n
		}
	}
	return nil
}

func filterDNSAnnotations(ann map[string]string) map[string]string {
	if ann == nil {
		return nil
	}
	out := map[string]string{}
	for k, v := range ann {
		if strings.Contains(k, "external-dns") {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func splitHostnames(v string) []string {
	parts := strings.FieldsFunc(v, func(r rune) bool {
		return r == ',' || r == ';' || r == ' '
	})
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func formatAge(t time.Time) string {
	d := time.Since(t).Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func parseExternalDNSArgs(args []string, env []corev1.EnvVar, status *models.ControllerStatus) {
	envMap := map[string]string{}
	for _, e := range env {
		envMap[e.Name] = e.Value
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--provider" && i+1 < len(args):
			status.Provider = args[i+1]
			i++
		case strings.HasPrefix(arg, "--provider="):
			status.Provider = strings.TrimPrefix(arg, "--provider=")
		case arg == "--txt-owner-id" && i+1 < len(args):
			status.TXTOwnerID = args[i+1]
			i++
		case strings.HasPrefix(arg, "--txt-owner-id="):
			status.TXTOwnerID = strings.TrimPrefix(arg, "--txt-owner-id=")
		case arg == "--policy" && i+1 < len(args):
			status.Policy = args[i+1]
			i++
		case strings.HasPrefix(arg, "--policy="):
			status.Policy = strings.TrimPrefix(arg, "--policy=")
		case arg == "--domain-filter" && i+1 < len(args):
			status.DomainFilters = append(status.DomainFilters, args[i+1])
			i++
		case strings.HasPrefix(arg, "--domain-filter="):
			status.DomainFilters = append(status.DomainFilters, strings.TrimPrefix(arg, "--domain-filter="))
		case arg == "--source" && i+1 < len(args):
			status.Sources = append(status.Sources, args[i+1])
			i++
		case strings.HasPrefix(arg, "--source="):
			status.Sources = append(status.Sources, strings.TrimPrefix(arg, "--source="))
		case arg == "--interval" && i+1 < len(args):
			status.Interval = args[i+1]
			i++
		case strings.HasPrefix(arg, "--interval="):
			status.Interval = strings.TrimPrefix(arg, "--interval=")
		case arg == "--dry-run":
			status.DryRun = true
		}
	}

	if status.Provider == "" {
		status.Provider = envMap["EXTERNAL_DNS_PROVIDER"]
	}
	if status.TXTOwnerID == "" {
		status.TXTOwnerID = envMap["EXTERNAL_DNS_TXT_OWNER_ID"]
	}
	if status.Policy == "" {
		status.Policy = envMap["EXTERNAL_DNS_POLICY"]
	}
	if status.Interval == "" {
		status.Interval = envMap["EXTERNAL_DNS_INTERVAL"]
	}
	if v := envMap["EXTERNAL_DNS_DRY_RUN"]; v == "1" || strings.EqualFold(v, "true") {
		status.DryRun = true
	}
}
