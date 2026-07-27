import { apiFetch } from "@/lib/auth-api";

export interface AdminOverview {
  cluster_name: string;
  domain_filter: string;
  total_records: number;
  unmanaged_sources: number;
  namespaces: number;
  controller_ready: boolean;
  write_enabled: boolean;
}

export interface AuditEntry {
  id: string;
  timestamp: string;
  actor: string;
  action: string;
  resource: string;
  detail: string;
  success: boolean;
}

export interface K8sResourceRef {
  kind: string;
  namespace: string;
  name: string;
}

export interface AnnotatePayload {
  kind: string;
  namespace: string;
  name: string;
  hostname: string;
  ttl?: number;
  internal?: boolean;
}

export interface AnnotateDetail extends AnnotatePayload {
  has_dns: boolean;
}

export interface DNSEndpointPayload {
  namespace: string;
  name: string;
  dns_name: string;
  record_type: string;
  targets: string[];
  ttl?: number;
}

export type DNSEndpointDetail = DNSEndpointPayload;

export const getAdminOverview = () => apiFetch<AdminOverview>("/api/v1/admin/overview");
export const getAdminAudit = (limit = 100) =>
  apiFetch<{ items: AuditEntry[] }>(`/api/v1/admin/audit?limit=${limit}`);
export const getAdminNamespaces = () =>
  apiFetch<{ items: string[] }>("/api/v1/admin/namespaces");
export const getAdminCandidates = (namespace?: string) =>
  apiFetch<{ items: K8sResourceRef[] }>(
    `/api/v1/admin/candidates${namespace ? `?namespace=${encodeURIComponent(namespace)}` : ""}`
  );

export const createAnnotation = (payload: AnnotatePayload) =>
  apiFetch<{ status: string }>("/api/v1/admin/annotate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

export const updateAnnotation = (payload: AnnotatePayload) =>
  apiFetch<{ status: string }>("/api/v1/admin/annotate", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

export const getAnnotation = (kind: string, namespace: string, name: string) =>
  apiFetch<AnnotateDetail>(
    `/api/v1/admin/annotate?kind=${encodeURIComponent(kind)}&namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`
  );

export const removeAnnotation = (kind: string, namespace: string, name: string) =>
  apiFetch<{ status: string }>(
    `/api/v1/admin/annotate?kind=${encodeURIComponent(kind)}&namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`,
    { method: "DELETE" }
  );

export const createDNSEndpoint = (payload: DNSEndpointPayload) =>
  apiFetch<{ status: string }>("/api/v1/admin/dnsendpoints", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

export const getDNSEndpoint = (namespace: string, name: string) =>
  apiFetch<DNSEndpointDetail>(`/api/v1/admin/dnsendpoints/${namespace}/${name}`);

export const updateDNSEndpoint = (namespace: string, name: string, payload: Omit<DNSEndpointPayload, "namespace" | "name">) =>
  apiFetch<{ status: string }>(`/api/v1/admin/dnsendpoints/${namespace}/${name}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

export const deleteDNSEndpoint = (namespace: string, name: string) =>
  apiFetch<{ status: string }>(`/api/v1/admin/dnsendpoints/${namespace}/${name}`, {
    method: "DELETE",
  });

/** @deprecated use createAnnotation */
export const annotateResource = createAnnotation;

export { getAdminSettings } from "@/lib/auth-api";
