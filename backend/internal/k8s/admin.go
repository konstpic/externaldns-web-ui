package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/konstpic/externaldns-web-ui/backend/internal/models"
)

func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	list, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(list.Items))
	for _, ns := range list.Items {
		out = append(out, ns.Name)
	}
	return out, nil
}

func (c *Client) AdminOverview(ctx context.Context) (*models.AdminOverview, error) {
	records, err := c.ListRecords(ctx, "")
	if err != nil {
		return nil, err
	}
	status, _ := c.ControllerStatus(ctx)
	nss, _ := c.ListNamespaces(ctx)
	unmanaged, _ := c.countUnmanagedSources(ctx)

	return &models.AdminOverview{
		ClusterName:      c.clusterName,
		DomainFilter:     c.domainFilter,
		TotalRecords:     len(records),
		UnmanagedSources: unmanaged,
		Namespaces:       len(nss),
		ControllerReady:  status.Ready,
		WriteEnabled:     true,
	}, nil
}

func (c *Client) countUnmanagedSources(ctx context.Context) (int, error) {
	count := 0
	services, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	for _, svc := range services.Items {
		if len(hostnamesFromAnnotations(svc.Annotations)) == 0 && hasExposePotential(svc) {
			count++
		}
	}
	ingresses, err := c.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return count, err
	}
	for _, ing := range ingresses.Items {
		if len(ingressHostnames(ing)) == 0 {
			count++
		}
	}
	return count, nil
}

func hasExposePotential(svc corev1.Service) bool {
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		return true
	}
	if len(svc.Spec.ExternalIPs) > 0 {
		return true
	}
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		return true
	}
	return false
}

func (c *Client) AnnotateResource(ctx context.Context, req models.AnnotateRequest) error {
	return c.applyAnnotatePatch(ctx, req, false)
}

func (c *Client) UpdateAnnotateResource(ctx context.Context, req models.AnnotateRequest) error {
	return c.applyAnnotatePatch(ctx, req, true)
}

func (c *Client) applyAnnotatePatch(ctx context.Context, req models.AnnotateRequest, isUpdate bool) error {
	if req.Kind != "Service" && req.Kind != "Ingress" {
		return fmt.Errorf("unsupported kind: %s", req.Kind)
	}
	hostname := strings.TrimSpace(req.Hostname)
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	if isUpdate {
		has, err := c.resourceHasDNSAnnotation(ctx, req.Kind, req.Namespace, req.Name)
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("resource has no DNS annotation to update")
		}
	}

	key := annHostname
	if req.Internal {
		key = annInternalHostname
	}

	anns := map[string]any{
		key: hostname,
	}
	if req.TTL != nil {
		anns[annTTL] = fmt.Sprintf("%d", *req.TTL)
	} else if isUpdate {
		anns[annTTL] = nil
	}

	// Clear other hostname keys when updating
	if req.Internal {
		anns[annHostname] = nil
		anns[annHostnameLegacy] = nil
	} else {
		anns[annInternalHostname] = nil
	}
	if isUpdate {
		anns[annHostnameLegacy] = nil
	}

	patch := map[string]any{
		"metadata": map[string]any{
			"annotations": anns,
		},
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	switch req.Kind {
	case "Service":
		_, err = c.clientset.CoreV1().Services(req.Namespace).Patch(ctx, req.Name, types.MergePatchType, raw, metav1.PatchOptions{})
	case "Ingress":
		_, err = c.clientset.NetworkingV1().Ingresses(req.Namespace).Patch(ctx, req.Name, types.MergePatchType, raw, metav1.PatchOptions{})
	}
	return err
}

func (c *Client) HasDNSAnnotation(ctx context.Context, kind, namespace, name string) (bool, error) {
	return c.resourceHasDNSAnnotation(ctx, kind, namespace, name)
}

func (c *Client) resourceHasDNSAnnotation(ctx context.Context, kind, namespace, name string) (bool, error) {
	switch kind {
	case "Service":
		svc, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return len(hostnamesFromAnnotations(svc.Annotations)) > 0, nil
	case "Ingress":
		ing, err := c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return len(hostnamesFromAnnotations(ing.Annotations)) > 0, nil
	default:
		return false, fmt.Errorf("unsupported kind: %s", kind)
	}
}

func (c *Client) GetAnnotateDetail(ctx context.Context, kind, namespace, name string) (*models.AnnotateDetail, error) {
	var anns map[string]string
	switch kind {
	case "Service":
		svc, err := c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		anns = svc.Annotations
	case "Ingress":
		ing, err := c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		anns = ing.Annotations
	default:
		return nil, fmt.Errorf("unsupported kind: %s", kind)
	}

	detail := &models.AnnotateDetail{
		Kind: kind, Namespace: namespace, Name: name,
	}
	if anns == nil {
		return detail, nil
	}

	if v := anns[annInternalHostname]; v != "" {
		detail.Internal = true
		detail.Hostname = strings.TrimSuffix(v, ".")
		detail.HasDNS = true
	} else if v := anns[annHostname]; v != "" {
		detail.Hostname = strings.TrimSuffix(v, ".")
		detail.HasDNS = true
	} else if v := anns[annHostnameLegacy]; v != "" {
		detail.Hostname = strings.TrimSuffix(v, ".")
		detail.HasDNS = true
	}
	detail.TTL = parseTTL(anns)
	return detail, nil
}

func (c *Client) RemoveDNSAnnotations(ctx context.Context, kind, namespace, name string) error {
	patch := map[string]any{
		"metadata": map[string]any{
			"annotations": map[string]any{
				annHostname:         nil,
				annHostnameLegacy:   nil,
				annInternalHostname: nil,
				annTTL:              nil,
			},
		},
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	switch kind {
	case "Service":
		_, err = c.clientset.CoreV1().Services(namespace).Patch(ctx, name, types.MergePatchType, raw, metav1.PatchOptions{})
	case "Ingress":
		_, err = c.clientset.NetworkingV1().Ingresses(namespace).Patch(ctx, name, types.MergePatchType, raw, metav1.PatchOptions{})
	default:
		return fmt.Errorf("unsupported kind: %s", kind)
	}
	return err
}

func (c *Client) CreateDNSEndpoint(ctx context.Context, req models.DNSEndpointRequest) error {
	if req.Namespace == "" || req.Name == "" || req.DNSName == "" {
		return fmt.Errorf("namespace, name and dns_name are required")
	}
	recordType := req.RecordType
	if recordType == "" {
		recordType = "A"
	}
	if len(req.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	endpoint := map[string]any{
		"dnsName":    req.DNSName,
		"recordType": recordType,
		"targets":    req.Targets,
	}
	if req.TTL != nil {
		endpoint["recordTTL"] = *req.TTL
	}

	obj := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "externaldns.k8s.io/v1alpha1",
			"kind":       "DNSEndpoint",
			"metadata": map[string]any{
				"name":      req.Name,
				"namespace": req.Namespace,
			},
			"spec": map[string]any{
				"endpoints": []any{endpoint},
			},
		},
	}

	gvr := schema.GroupVersionResource{Group: "externaldns.k8s.io", Version: "v1alpha1", Resource: "dnsendpoints"}
	_, err := c.dynamic.Resource(gvr).Namespace(req.Namespace).Create(ctx, obj, metav1.CreateOptions{})
	return err
}

func (c *Client) DeleteDNSEndpoint(ctx context.Context, namespace, name string) error {
	gvr := schema.GroupVersionResource{Group: "externaldns.k8s.io", Version: "v1alpha1", Resource: "dnsendpoints"}
	return c.dynamic.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) GetDNSEndpoint(ctx context.Context, namespace, name string) (*models.DNSEndpointDetail, error) {
	gvr := schema.GroupVersionResource{Group: "externaldns.k8s.io", Version: "v1alpha1", Resource: "dnsendpoints"}
	obj, err := c.dynamic.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	detail := &models.DNSEndpointDetail{Namespace: namespace, Name: name}
	endpoints, _, _ := unstructured.NestedSlice(obj.Object, "spec", "endpoints")
	if len(endpoints) == 0 {
		return detail, nil
	}
	m, ok := endpoints[0].(map[string]any)
	if !ok {
		return detail, nil
	}
	detail.DNSName, _ = m["dnsName"].(string)
	detail.RecordType, _ = m["recordType"].(string)
	if detail.RecordType == "" {
		detail.RecordType = "A"
	}
	if targets, ok := m["targets"].([]any); ok {
		for _, t := range targets {
			if s, ok := t.(string); ok {
				detail.Targets = append(detail.Targets, s)
			}
		}
	}
	if v, ok := m["recordTTL"].(int64); ok {
		detail.TTL = &v
	} else if v, ok := m["recordTTL"].(float64); ok {
		i := int64(v)
		detail.TTL = &i
	}
	return detail, nil
}

func (c *Client) UpdateDNSEndpoint(ctx context.Context, req models.DNSEndpointRequest) error {
	if req.Namespace == "" || req.Name == "" || req.DNSName == "" {
		return fmt.Errorf("namespace, name and dns_name are required")
	}
	recordType := req.RecordType
	if recordType == "" {
		recordType = "A"
	}
	if len(req.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	endpoint := map[string]any{
		"dnsName":    req.DNSName,
		"recordType": recordType,
		"targets":    req.Targets,
	}
	if req.TTL != nil {
		endpoint["recordTTL"] = *req.TTL
	}

	patch := map[string]any{
		"spec": map[string]any{
			"endpoints": []any{endpoint},
		},
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{Group: "externaldns.k8s.io", Version: "v1alpha1", Resource: "dnsendpoints"}
	_, err = c.dynamic.Resource(gvr).Namespace(req.Namespace).Patch(ctx, req.Name, types.MergePatchType, raw, metav1.PatchOptions{})
	return err
}

func (c *Client) ListCandidateResources(ctx context.Context, namespace string) ([]models.K8sResourceRef, error) {
	var out []models.K8sResourceRef

	svcList, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, svc := range svcList.Items {
		if len(hostnamesFromAnnotations(svc.Annotations)) > 0 {
			continue
		}
		if !hasExposePotential(svc) {
			continue
		}
		out = append(out, models.K8sResourceRef{Kind: "Service", Namespace: svc.Namespace, Name: svc.Name})
	}

	ingList, err := c.clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, ing := range ingList.Items {
		if len(ingressHostnames(ing)) > 0 {
			continue
		}
		if len(ing.Spec.Rules) == 0 {
			continue
		}
		out = append(out, models.K8sResourceRef{Kind: "Ingress", Namespace: ing.Namespace, Name: ing.Name})
	}
	return out, nil
}
