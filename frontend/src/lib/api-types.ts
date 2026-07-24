export interface DNSRecord {
  id: string;
  hostname: string;
  record_type: string;
  target: string;
  ttl?: number;
  source_type: string;
  namespace: string;
  resource: string;
  annotations?: Record<string, string>;
  created_at?: string;
}

export interface SourceResource {
  kind: string;
  namespace: string;
  name: string;
  hostnames: string[];
  target: string;
  annotations?: Record<string, string>;
  age: string;
}

export interface ControllerStatus {
  namespace: string;
  deployment: string;
  ready: boolean;
  replicas: number;
  ready_replicas: number;
  image: string;
  provider: string;
  domain_filters: string[];
  txt_owner_id: string;
  policy: string;
  sources: string[];
  dry_run: boolean;
  interval: string;
  pod_names: string[];
}

export interface Overview {
  total_records: number;
  service_sources: number;
  ingress_sources: number;
  dnsendpoint_crds: number;
  namespaces: number;
  controller: ControllerStatus;
  cluster_name: string;
  domain_filter: string;
}

export interface LogLine {
  timestamp: string;
  message: string;
}
