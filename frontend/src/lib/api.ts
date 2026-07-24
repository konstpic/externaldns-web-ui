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

async function api<T>(path: string): Promise<T> {
  const res = await fetch(path);
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const getOverview = () => api<Overview>("/api/v1/overview");
export const getRecords = (q?: string) =>
  api<{ items: DNSRecord[]; total: number }>(
    `/api/v1/records${q ? `?q=${encodeURIComponent(q)}` : ""}`
  );
export const getSources = (kind?: string) =>
  api<{ items: SourceResource[]; total: number }>(
    `/api/v1/sources${kind ? `?kind=${encodeURIComponent(kind)}` : ""}`
  );
export const getController = () => api<ControllerStatus>("/api/v1/controller");
export const getLogs = (tail = 100) =>
  api<{ items: LogLine[] }>(`/api/v1/logs?tail=${tail}`);
