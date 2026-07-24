package models

import "time"

type DNSRecord struct {
	ID          string            `json:"id"`
	Hostname    string            `json:"hostname"`
	RecordType  string            `json:"record_type"`
	Target      string            `json:"target"`
	TTL         *int64            `json:"ttl,omitempty"`
	SourceType  string            `json:"source_type"`
	Namespace   string            `json:"namespace"`
	Resource    string            `json:"resource"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   *time.Time        `json:"created_at,omitempty"`
}

type SourceResource struct {
	Kind        string            `json:"kind"`
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Hostnames   []string          `json:"hostnames"`
	Target      string            `json:"target"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Age         string            `json:"age"`
}

type ControllerStatus struct {
	Namespace      string   `json:"namespace"`
	Deployment     string   `json:"deployment"`
	Ready          bool     `json:"ready"`
	Replicas       int32    `json:"replicas"`
	ReadyReplicas  int32    `json:"ready_replicas"`
	Image          string   `json:"image"`
	Provider       string   `json:"provider"`
	DomainFilters  []string `json:"domain_filters"`
	TXTOwnerID     string   `json:"txt_owner_id"`
	Policy         string   `json:"policy"`
	Sources        []string `json:"sources"`
	DryRun         bool     `json:"dry_run"`
	Interval       string   `json:"interval"`
	LastSyncHint   string   `json:"last_sync_hint"`
	PodNames       []string `json:"pod_names"`
}

type Overview struct {
	TotalRecords    int              `json:"total_records"`
	ServiceSources  int              `json:"service_sources"`
	IngressSources  int              `json:"ingress_sources"`
	DNSEndpointCRDs int              `json:"dnsendpoint_crds"`
	Namespaces      int              `json:"namespaces"`
	Controller      ControllerStatus `json:"controller"`
	ClusterName     string           `json:"cluster_name"`
	DomainFilter    string           `json:"domain_filter"`
}

type LogLine struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}
