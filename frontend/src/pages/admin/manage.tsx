import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Pencil, Plus, Save, Trash2, X } from "lucide-react";
import {
  createAnnotation,
  createDNSEndpoint,
  deleteDNSEndpoint,
  getAdminCandidates,
  getAdminNamespaces,
  getAnnotation,
  getDNSEndpoint,
  removeAnnotation,
  updateAnnotation,
  updateDNSEndpoint,
} from "@/lib/admin-api";
import { getRecords, type DNSRecord } from "@/lib/api";
import { ApiError } from "@/lib/auth-api";
import { ErrorState, FadeIn } from "@/components/ui";
import { cn } from "@/lib/utils";

type Tab = "create" | "edit" | "delete";

export function AdminManagePage() {
  const qc = useQueryClient();
  const [tab, setTab] = useState<Tab>("create");
  const [createMode, setCreateMode] = useState<"annotate" | "crd">("annotate");
  const [namespace, setNamespace] = useState("default");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [editing, setEditing] = useState<DNSRecord | null>(null);

  const { data: namespaces } = useQuery({
    queryKey: ["admin-namespaces"],
    queryFn: getAdminNamespaces,
  });

  const { data: candidates, isLoading: loadingCandidates } = useQuery({
    queryKey: ["admin-candidates", namespace],
    queryFn: () => getAdminCandidates(namespace),
    enabled: tab === "create" && createMode === "annotate",
  });

  const { data: recordsData } = useQuery({
    queryKey: ["records"],
    queryFn: () => getRecords(),
  });

  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ["records"] });
    qc.invalidateQueries({ queryKey: ["admin-audit"] });
    qc.invalidateQueries({ queryKey: ["overview"] });
    qc.invalidateQueries({ queryKey: ["admin-overview"] });
    qc.invalidateQueries({ queryKey: ["admin-candidates"] });
  };

  const mutateOpts = {
    onSuccess: () => {
      setMessage("Изменения применены");
      setError("");
      setEditing(null);
      invalidate();
    },
    onError: (err: Error) => {
      setMessage("");
      setError(err instanceof ApiError ? err.message : "Ошибка операции");
    },
  };

  const createAnnotMut = useMutation({ mutationFn: createAnnotation, ...mutateOpts });
  const updateAnnotMut = useMutation({ mutationFn: updateAnnotation, ...mutateOpts });
  const crdMut = useMutation({ mutationFn: createDNSEndpoint, ...mutateOpts });
  const updateCrdMut = useMutation({
    mutationFn: ({ ns, name, payload }: { ns: string; name: string; payload: Parameters<typeof updateDNSEndpoint>[2] }) =>
      updateDNSEndpoint(ns, name, payload),
    ...mutateOpts,
  });
  const removeMut = useMutation({
    mutationFn: ({ kind, namespace, name }: { kind: string; namespace: string; name: string }) =>
      removeAnnotation(kind, namespace, name),
    ...mutateOpts,
  });
  const deleteCrdMut = useMutation({
    mutationFn: ({ ns, name }: { ns: string; name: string }) => deleteDNSEndpoint(ns, name),
    ...mutateOpts,
  });

  const dnsEndpoints = useMemo(
    () => (recordsData?.items ?? []).filter((r) => r.source_type === "DNSEndpoint"),
    [recordsData]
  );

  const editableRecords = useMemo(
    () => recordsData?.items ?? [],
    [recordsData]
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap gap-2">
        {(
          [
            ["create", "Создать"],
            ["edit", "Редактировать"],
            ["delete", "Удалить"],
          ] as const
        ).map(([t, label]) => (
          <button
            key={t}
            type="button"
            className={cn(
              "rounded-xl px-4 py-2 text-sm font-medium transition-colors",
              tab === t ? "bg-primary/15 text-primary" : "bg-surface/10 text-muted"
            )}
            onClick={() => {
              setTab(t);
              setEditing(null);
              setMessage("");
              setError("");
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {message && <p className="text-sm text-success">{message}</p>}
      {error && <ErrorState message={error} />}

      {tab === "create" && (
        <FadeIn>
          <div className="mb-4 flex gap-2">
            <button
              type="button"
              className={cn("rounded-lg px-3 py-1.5 text-xs font-medium", createMode === "annotate" ? "bg-primary/15 text-primary" : "text-muted")}
              onClick={() => setCreateMode("annotate")}
            >
              Annotation
            </button>
            <button
              type="button"
              className={cn("rounded-lg px-3 py-1.5 text-xs font-medium", createMode === "crd" ? "bg-primary/15 text-primary" : "text-muted")}
              onClick={() => setCreateMode("crd")}
            >
              DNSEndpoint CRD
            </button>
          </div>
          {createMode === "annotate" ? (
            <AnnotateForm
              mode="create"
              namespaces={namespaces?.items ?? []}
              namespace={namespace}
              onNamespace={setNamespace}
              candidates={candidates?.items ?? []}
              loading={loadingCandidates}
              pending={createAnnotMut.isPending}
              onSubmit={(p) => createAnnotMut.mutate(p)}
            />
          ) : (
            <CRDForm mode="create" namespaces={namespaces?.items ?? ["default"]} pending={crdMut.isPending} onSubmit={(p) => crdMut.mutate(p)} />
          )}
        </FadeIn>
      )}

      {tab === "edit" && (
        <FadeIn>
          <div className="glass p-6">
            <h2 className="text-lg font-semibold text-fg">Редактировать DNS запись</h2>
            <p className="mt-1 text-sm text-muted">Выберите запись для изменения hostname, TTL или targets</p>
            <ul className="mt-4 max-h-64 space-y-2 overflow-auto">
              {editableRecords.map((r) => (
                <li
                  key={r.id}
                  className={cn(
                    "flex cursor-pointer items-center justify-between gap-2 rounded-xl border px-3 py-2 text-sm transition-colors",
                    editing?.id === r.id ? "border-primary/40 bg-primary/10" : "border-border/10 hover:bg-surface/5"
                  )}
                  onClick={() => setEditing(r)}
                >
                  <div className="min-w-0">
                    <p className="truncate font-medium text-fg">{r.hostname}</p>
                    <p className="truncate text-xs text-subtle">
                      {r.source_type}/{r.namespace}/{r.resource} · {r.target}
                    </p>
                  </div>
                  <Pencil className="h-4 w-4 shrink-0 text-primary" />
                </li>
              ))}
              {editableRecords.length === 0 && <p className="text-sm text-muted">Нет записей для редактирования</p>}
            </ul>
          </div>

          {editing && editing.source_type !== "DNSEndpoint" && (
            <div className="mt-6">
              <EditAnnotatePanel
                record={editing}
                pending={updateAnnotMut.isPending}
                onCancel={() => setEditing(null)}
                onSave={(p) => updateAnnotMut.mutate(p)}
              />
            </div>
          )}

          {editing && editing.source_type === "DNSEndpoint" && (
            <div className="mt-6">
              <EditCRDPanel
                record={editing}
                pending={updateCrdMut.isPending}
                onCancel={() => setEditing(null)}
                onSave={(payload) =>
                  updateCrdMut.mutate({ ns: editing.namespace, name: editing.resource, payload })
                }
              />
            </div>
          )}
        </FadeIn>
      )}

      {tab === "delete" && (
        <FadeIn>
          <RemovePanel
            records={recordsData?.items ?? []}
            dnsEndpoints={dnsEndpoints}
            pending={removeMut.isPending || deleteCrdMut.isPending}
            onRemoveAnnotation={(kind, ns, name) => removeMut.mutate({ kind, namespace: ns, name })}
            onDeleteCrd={(ns, name) => deleteCrdMut.mutate({ ns, name })}
          />
        </FadeIn>
      )}
    </div>
  );
}

function EditAnnotatePanel({
  record,
  pending,
  onCancel,
  onSave,
}: {
  record: DNSRecord;
  pending: boolean;
  onCancel: () => void;
  onSave: (p: AnnotatePayload) => void;
}) {
  const { data, isLoading } = useQuery({
    queryKey: ["annotate-detail", record.source_type, record.namespace, record.resource],
    queryFn: () => getAnnotation(record.source_type, record.namespace, record.resource),
  });

  const [hostname, setHostname] = useState("");
  const [ttl, setTtl] = useState("");
  const [internal, setInternal] = useState(false);

  useEffect(() => {
    if (data) {
      setHostname(data.hostname);
      setTtl(data.ttl != null ? String(data.ttl) : "");
      setInternal(!!data.internal);
    }
  }, [data]);

  if (isLoading) return <Loader2 className="mx-auto h-6 w-6 animate-spin text-primary" />;

  return (
    <AnnotateForm
      mode="edit"
      namespaces={[record.namespace]}
      namespace={record.namespace}
      onNamespace={() => {}}
      candidates={[]}
      loading={false}
      pending={pending}
      initialKind={record.source_type}
      initialName={record.resource}
      hostname={hostname}
      ttl={ttl}
      internal={internal}
      onHostname={setHostname}
      onTtl={setTtl}
      onInternal={setInternal}
      lockResource
      onCancel={onCancel}
      onSubmit={(p) => onSave(p)}
    />
  );
}

function EditCRDPanel({
  record,
  pending,
  onCancel,
  onSave,
}: {
  record: DNSRecord;
  pending: boolean;
  onCancel: () => void;
  onSave: (p: { dns_name: string; record_type: string; targets: string[]; ttl?: number }) => void;
}) {
  const { data, isLoading } = useQuery({
    queryKey: ["dnsendpoint-detail", record.namespace, record.resource],
    queryFn: () => getDNSEndpoint(record.namespace, record.resource),
  });

  if (isLoading || !data) return <Loader2 className="mx-auto h-6 w-6 animate-spin text-primary" />;

  return (
    <CRDForm
      mode="edit"
      namespaces={[record.namespace]}
      initial={{ ...data, namespace: record.namespace, name: record.resource }}
      pending={pending}
      onCancel={onCancel}
      onSubmit={(p) =>
        onSave({
          dns_name: p.dns_name,
          record_type: p.record_type,
          targets: p.targets,
          ttl: p.ttl,
        })
      }
    />
  );
}

type AnnotatePayload = {
  kind: string;
  namespace: string;
  name: string;
  hostname: string;
  ttl?: number;
  internal?: boolean;
};

function AnnotateForm({
  mode,
  namespaces,
  namespace,
  onNamespace,
  candidates,
  loading,
  pending,
  onSubmit,
  initialKind = "Service",
  initialName = "",
  hostname: hostnameProp,
  ttl: ttlProp,
  internal: internalProp,
  onHostname,
  onTtl,
  onInternal,
  lockResource = false,
  onCancel,
}: {
  mode: "create" | "edit";
  namespaces: string[];
  namespace: string;
  onNamespace: (v: string) => void;
  candidates: { kind: string; namespace: string; name: string }[];
  loading: boolean;
  pending: boolean;
  onSubmit: (p: AnnotatePayload) => void;
  initialKind?: string;
  initialName?: string;
  hostname?: string;
  ttl?: string;
  internal?: boolean;
  onHostname?: (v: string) => void;
  onTtl?: (v: string) => void;
  onInternal?: (v: boolean) => void;
  lockResource?: boolean;
  onCancel?: () => void;
}) {
  const [kind, setKind] = useState(initialKind);
  const [name, setName] = useState(initialName);
  const [hostnameLocal, setHostnameLocal] = useState("");
  const [ttlLocal, setTtlLocal] = useState("");
  const [internalLocal, setInternalLocal] = useState(false);

  const hostname = hostnameProp ?? hostnameLocal;
  const ttl = ttlProp ?? ttlLocal;
  const internal = internalProp ?? internalLocal;
  const setHostname = onHostname ?? setHostnameLocal;
  const setTtl = onTtl ?? setTtlLocal;
  const setInternal = onInternal ?? setInternalLocal;

  return (
    <div className="glass p-6">
      <h2 className="text-lg font-semibold text-fg">
        {mode === "create" ? "Добавить DNS annotation" : "Редактировать annotation"}
      </h2>

      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <Field label="Namespace">
          <select
            className="select-field"
            value={namespace}
            onChange={(e) => onNamespace(e.target.value)}
            disabled={lockResource}
          >
            {namespaces.map((ns) => (
              <option key={ns} value={ns}>{ns}</option>
            ))}
          </select>
        </Field>
        <Field label="Kind">
          <select className="select-field" value={kind} onChange={(e) => setKind(e.target.value)} disabled={lockResource}>
            <option value="Service">Service</option>
            <option value="Ingress">Ingress</option>
          </select>
        </Field>
        <Field label="Resource name">
          <input className="input-field" value={lockResource ? initialName : name} onChange={(e) => setName(e.target.value)} disabled={lockResource} />
        </Field>
        <Field label="Hostname">
          <input className="input-field" value={hostname} onChange={(e) => setHostname(e.target.value)} placeholder="app.example.com" />
        </Field>
        <Field label="TTL (optional)">
          <input className="input-field" value={ttl} onChange={(e) => setTtl(e.target.value)} placeholder="300" />
        </Field>
        <label className="flex items-center gap-2 self-end pb-2 text-sm text-muted">
          <input type="checkbox" checked={internal} onChange={(e) => setInternal(e.target.checked)} className="checkbox-field" />
          Internal hostname
        </label>
      </div>

      {mode === "create" && !loading && candidates.length > 0 && (
        <div className="mt-4">
          <p className="text-xs font-semibold uppercase tracking-wider text-subtle">Кандидаты без DNS</p>
          <div className="mt-2 flex flex-wrap gap-2">
            {candidates.map((c) => (
              <button
                key={`${c.kind}/${c.name}`}
                type="button"
                className="badge hover:border-primary/30"
                onClick={() => {
                  setKind(c.kind);
                  setName(c.name);
                }}
              >
                {c.kind}/{c.name}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="mt-6 flex flex-wrap gap-3">
        <button
          type="button"
          className="btn-primary"
          disabled={pending || !(lockResource ? initialName : name) || !hostname}
          onClick={() =>
            onSubmit({
              kind,
              namespace,
              name: lockResource ? initialName : name,
              hostname,
              ttl: ttl ? Number(ttl) : undefined,
              internal,
            })
          }
        >
          {pending ? <Loader2 className="h-4 w-4 animate-spin" /> : mode === "edit" ? <Save className="h-4 w-4" /> : <Plus className="h-4 w-4" />}
          {mode === "edit" ? "Сохранить" : "Создать"}
        </button>
        {onCancel && (
          <button type="button" className="btn-secondary" onClick={onCancel}>
            <X className="h-4 w-4" />
            Отмена
          </button>
        )}
      </div>
    </div>
  );
}

function CRDForm({
  mode,
  namespaces,
  pending,
  onSubmit,
  initial,
  onCancel,
}: {
  mode: "create" | "edit";
  namespaces: string[];
  pending: boolean;
  onSubmit: (p: {
    namespace: string;
    name: string;
    dns_name: string;
    record_type: string;
    targets: string[];
    ttl?: number;
  }) => void;
  initial?: {
    namespace: string;
    name: string;
    dns_name: string;
    record_type: string;
    targets: string[];
    ttl?: number;
  };
  onCancel?: () => void;
}) {
  const [namespace, setNamespace] = useState(initial?.namespace ?? namespaces[0] ?? "default");
  const [name, setName] = useState(initial?.name ?? "");
  const [dnsName, setDnsName] = useState(initial?.dns_name ?? "");
  const [recordType, setRecordType] = useState(initial?.record_type ?? "A");
  const [targets, setTargets] = useState(initial?.targets?.join(", ") ?? "");
  const [ttl, setTtl] = useState(initial?.ttl != null ? String(initial.ttl) : "");

  useEffect(() => {
    if (initial) {
      setNamespace(initial.namespace);
      setName(initial.name);
      setDnsName(initial.dns_name);
      setRecordType(initial.record_type);
      setTargets(initial.targets.join(", "));
      setTtl(initial.ttl != null ? String(initial.ttl) : "");
    }
  }, [initial]);

  return (
    <div className="glass p-6">
      <h2 className="text-lg font-semibold text-fg">
        {mode === "create" ? "Создать DNSEndpoint CRD" : "Редактировать DNSEndpoint CRD"}
      </h2>
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        <Field label="Namespace">
          <input className="input-field" value={namespace} disabled />
        </Field>
        <Field label="CRD name">
          <input className="input-field" value={name} disabled={mode === "edit"} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Field label="DNS name">
          <input className="input-field" value={dnsName} onChange={(e) => setDnsName(e.target.value)} />
        </Field>
        <Field label="Record type">
          <select className="select-field" value={recordType} onChange={(e) => setRecordType(e.target.value)}>
            {["A", "AAAA", "CNAME", "TXT"].map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </Field>
        <Field label="Targets (comma separated)">
          <input className="input-field sm:col-span-2" value={targets} onChange={(e) => setTargets(e.target.value)} />
        </Field>
        <Field label="TTL">
          <input className="input-field" value={ttl} onChange={(e) => setTtl(e.target.value)} />
        </Field>
      </div>
      <div className="mt-6 flex flex-wrap gap-3">
        <button
          type="button"
          className="btn-primary"
          disabled={pending || !name || !dnsName || !targets}
          onClick={() =>
            onSubmit({
              namespace,
              name,
              dns_name: dnsName,
              record_type: recordType,
              targets: targets.split(/[,;\s]+/).map((t) => t.trim()).filter(Boolean),
              ttl: ttl ? Number(ttl) : undefined,
            })
          }
        >
          {pending ? <Loader2 className="h-4 w-4 animate-spin" /> : mode === "edit" ? <Save className="h-4 w-4" /> : <Plus className="h-4 w-4" />}
          {mode === "edit" ? "Сохранить" : "Создать CRD"}
        </button>
        {onCancel && (
          <button type="button" className="btn-secondary" onClick={onCancel}>
            <X className="h-4 w-4" />
            Отмена
          </button>
        )}
      </div>
    </div>
  );
}

function RemovePanel({
  records,
  dnsEndpoints,
  pending,
  onRemoveAnnotation,
  onDeleteCrd,
}: {
  records: DNSRecord[];
  dnsEndpoints: DNSRecord[];
  pending: boolean;
  onRemoveAnnotation: (kind: string, ns: string, name: string) => void;
  onDeleteCrd: (ns: string, name: string) => void;
}) {
  const annotated = records.filter((r) => r.source_type !== "DNSEndpoint");

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">Удалить annotations</h2>
        <ul className="mt-4 max-h-80 space-y-2 overflow-auto">
          {annotated.map((r) => (
            <li key={r.id} className="flex items-center justify-between gap-2 rounded-xl border border-border/10 px-3 py-2 text-sm">
              <div className="min-w-0">
                <p className="truncate font-medium text-fg">{r.hostname}</p>
                <p className="truncate text-xs text-subtle">{r.source_type}/{r.namespace}/{r.resource}</p>
              </div>
              <button type="button" className="btn-secondary !px-2 !py-1.5" disabled={pending} onClick={() => onRemoveAnnotation(r.source_type, r.namespace, r.resource)}>
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </li>
          ))}
          {annotated.length === 0 && <p className="text-sm text-muted">Нет annotated ресурсов</p>}
        </ul>
      </div>
      <div className="glass p-6">
        <h2 className="text-lg font-semibold text-fg">Удалить DNSEndpoint CRD</h2>
        <ul className="mt-4 max-h-80 space-y-2 overflow-auto">
          {dnsEndpoints.map((r) => (
            <li key={r.id} className="flex items-center justify-between gap-2 rounded-xl border border-border/10 px-3 py-2 text-sm">
              <div className="min-w-0">
                <p className="truncate font-medium text-fg">{r.hostname}</p>
                <p className="truncate text-xs text-subtle">{r.namespace}/{r.resource}</p>
              </div>
              <button type="button" className="btn-secondary !px-2 !py-1.5" disabled={pending} onClick={() => onDeleteCrd(r.namespace, r.resource)}>
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </li>
          ))}
          {dnsEndpoints.length === 0 && <p className="text-sm text-muted">Нет DNSEndpoint CRD</p>}
        </ul>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-subtle">{label}</span>
      {children}
    </label>
  );
}
