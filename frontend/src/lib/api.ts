import { apiFetch } from "@/lib/auth-api";
import type {
  ControllerStatus,
  DNSRecord,
  LogLine,
  Overview,
  SourceResource,
} from "@/lib/api-types";

export type { ControllerStatus, DNSRecord, LogLine, Overview, SourceResource };

export const getOverview = () => apiFetch<Overview>("/api/v1/overview");
export const getRecords = (q?: string) =>
  apiFetch<{ items: DNSRecord[]; total: number }>(
    `/api/v1/records${q ? `?q=${encodeURIComponent(q)}` : ""}`
  );
export const getSources = (kind?: string) =>
  apiFetch<{ items: SourceResource[]; total: number }>(
    `/api/v1/sources${kind ? `?kind=${encodeURIComponent(kind)}` : ""}`
  );
export const getController = () => apiFetch<ControllerStatus>("/api/v1/controller");
export const getLogs = (tail = 100) =>
  apiFetch<{ items: LogLine[] }>(`/api/v1/logs?tail=${tail}`);
