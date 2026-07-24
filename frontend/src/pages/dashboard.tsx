import { useQuery } from "@tanstack/react-query";
import { Globe, Layers, Network, Server } from "lucide-react";
import { Link } from "react-router-dom";
import { FadeIn, LoadingState, PageHeader, StatCard, ErrorState } from "@/components/ui";
import { getOverview } from "@/lib/api";
import { cn, pageShellClass } from "@/lib/utils";

export function DashboardPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["overview"],
    queryFn: getOverview,
  });

  if (isLoading) return <div className={pageShellClass}><LoadingState /></div>;
  if (error || !data) {
    return (
      <div className={pageShellClass}>
        <ErrorState message={error instanceof Error ? error.message : "Не удалось загрузить данные"} />
      </div>
    );
  }

  const { controller: c } = data;

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <PageHeader
          title="ExternalDNS Dashboard"
          subtitle={`Кластер ${data.cluster_name}${data.domain_filter ? ` · зона ${data.domain_filter}` : ""}`}
        />
      </FadeIn>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <FadeIn delay={0.05}>
          <StatCard label="DNS записи" value={data.total_records} icon={<Globe className="h-5 w-5" />} />
        </FadeIn>
        <FadeIn delay={0.1}>
          <StatCard label="Services" value={data.service_sources} icon={<Server className="h-5 w-5" />} />
        </FadeIn>
        <FadeIn delay={0.15}>
          <StatCard label="Ingresses" value={data.ingress_sources} icon={<Network className="h-5 w-5" />} />
        </FadeIn>
        <FadeIn delay={0.2}>
          <StatCard
            label="Namespaces"
            value={data.namespaces}
            hint={`${data.dnsendpoint_crds} DNSEndpoint CRD`}
            icon={<Layers className="h-5 w-5" />}
          />
        </FadeIn>
      </div>

      <FadeIn delay={0.25}>
        <div className="mt-8 glass p-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h2 className="text-lg font-semibold text-fg">ExternalDNS Controller</h2>
              <p className="mt-1 text-sm text-muted">
                {c.namespace}/{c.deployment}
              </p>
            </div>
            <span className={c.ready ? "badge-success" : "badge-danger"}>
              {c.ready ? "Ready" : "Not Ready"}
            </span>
          </div>

          <dl className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <Info label="Provider" value={c.provider || "—"} />
            <Info label="Policy" value={c.policy || "—"} />
            <Info label="TXT Owner ID" value={c.txt_owner_id || "—"} />
            <Info label="Domain filters" value={c.domain_filters.join(", ") || "—"} />
            <Info label="Sources" value={c.sources.join(", ") || "—"} />
            <Info label="Replicas" value={`${c.ready_replicas}/${c.replicas}`} />
            <Info label="Image" value={c.image || "—"} />
            <Info label="Interval" value={c.interval || "—"} />
            <Info label="Dry run" value={c.dry_run ? "Yes" : "No"} />
          </dl>

          <div className="mt-6 flex flex-wrap gap-3">
            <Link to="/records" className="btn-primary">DNS записи</Link>
            <Link to="/controller" className="btn-secondary">Подробнее</Link>
            <Link to="/logs" className="btn-secondary">Логи</Link>
          </div>
        </div>
      </FadeIn>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wider text-subtle">{label}</dt>
      <dd className="mt-1 break-all text-sm text-fg-secondary">{value}</dd>
    </div>
  );
}
