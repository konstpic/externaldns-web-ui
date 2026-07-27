package models

type AnnotateRequest struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Hostname  string `json:"hostname"`
	TTL       *int64 `json:"ttl,omitempty"`
	Internal  bool   `json:"internal"`
}

type DNSEndpointRequest struct {
	Namespace  string   `json:"namespace"`
	Name       string   `json:"name"`
	DNSName    string   `json:"dns_name"`
	RecordType string   `json:"record_type"`
	Targets    []string `json:"targets"`
	TTL        *int64   `json:"ttl,omitempty"`
}

type AdminOverview struct {
	ClusterName     string `json:"cluster_name"`
	DomainFilter    string `json:"domain_filter"`
	TotalRecords    int    `json:"total_records"`
	UnmanagedSources int   `json:"unmanaged_sources"`
	Namespaces      int    `json:"namespaces"`
	ControllerReady bool   `json:"controller_ready"`
	WriteEnabled    bool   `json:"write_enabled"`
}

type K8sResourceRef struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type AnnotateDetail struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Hostname  string `json:"hostname"`
	TTL       *int64 `json:"ttl,omitempty"`
	Internal  bool   `json:"internal"`
	HasDNS    bool   `json:"has_dns"`
}

type DNSEndpointDetail struct {
	Namespace  string   `json:"namespace"`
	Name       string   `json:"name"`
	DNSName    string   `json:"dns_name"`
	RecordType string   `json:"record_type"`
	Targets    []string `json:"targets"`
	TTL        *int64   `json:"ttl,omitempty"`
}
