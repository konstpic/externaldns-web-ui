import { useQuery } from "@tanstack/react-query";
import { FadeIn, LoadingState, PageHeader, ErrorState } from "@/components/ui";
import { getController } from "@/lib/api";
import { cn, pageShellClass } from "@/lib/utils";

export function ControllerPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["controller"],
    queryFn: getController,
  });

  if (isLoading) return <div className={pageShellClass}><LoadingState /></div>;
  if (error || !data) {
    return (
      <div className={pageShellClass}>
        <ErrorState message={error instanceof Error ? error.message : "Controller не найден"} />
      </div>
    );
  }

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <PageHeader
          title="ExternalDNS Controller"
          subtitle={`Deployment ${data.namespace}/${data.deployment}`}
        />
      </FadeIn>

      <FadeIn delay={0.05}>
        <div className="glass p-6">
          <div className="flex items-center gap-3">
            <span className={data.ready ? "badge-success" : "badge-danger"}>
              {data.ready ? "Ready" : "Not Ready"}
            </span>
            <span className="text-sm text-muted">
              {data.ready_replicas}/{data.replicas} replicas
            </span>
            {data.dry_run && <span className="badge-warning">Dry Run</span>}
          </div>

          <dl className="mt-6 grid gap-4 sm:grid-cols-2">
            <Field label="Image" value={data.image} mono />
            <Field label="Provider" value={data.provider} />
            <Field label="Policy" value={data.policy} />
            <Field label="TXT Owner ID" value={data.txt_owner_id} />
            <Field label="Domain filters" value={data.domain_filters.join(", ")} />
            <Field label="Sources" value={data.sources.join(", ")} />
            <Field label="Interval" value={data.interval} />
            <Field label="Pods" value={data.pod_names.join(", ")} />
          </dl>
        </div>
      </FadeIn>

      <FadeIn delay={0.1}>
        <div className="mt-6 glass p-6">
          <h2 className="text-lg font-semibold text-fg">Как добавить запись</h2>
          <p className="mt-2 text-sm text-muted">
            ExternalDNS читает annotations и hostnames из Kubernetes ресурсов и синхронизирует их с DNS провайдером.
          </p>
          <pre className="mt-4 overflow-x-auto rounded-xl bg-black/30 p-4 text-xs text-fg-secondary">
{`# Service
kubectl annotate service my-svc \\
  external-dns.kubernetes.io/hostname=app.example.com.

# Ingress — hostname в spec.rules или annotation
external-dns.alpha.kubernetes.io/hostname: app.example.com`}
          </pre>
        </div>
      </FadeIn>
    </div>
  );
}

function Field({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wider text-subtle">{label}</dt>
      <dd className={cn("mt-1 text-sm text-fg-secondary", mono && "font-mono text-xs break-all")}>
        {value || "—"}
      </dd>
    </div>
  );
}
