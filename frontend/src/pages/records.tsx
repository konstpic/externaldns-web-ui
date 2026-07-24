import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { EmptyState, ErrorState, FadeIn, LoadingState, PageHeader } from "@/components/ui";
import { getRecords } from "@/lib/api";
import { cn, formatDate, pageShellClass } from "@/lib/utils";

export function RecordsPage() {
  const [query, setQuery] = useState("");
  const [debounced, setDebounced] = useState("");

  const { data, isLoading, error } = useQuery({
    queryKey: ["records", debounced],
    queryFn: () => getRecords(debounced || undefined),
  });

  const items = useMemo(() => data?.items ?? [], [data]);

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <PageHeader
          title="DNS записи"
          subtitle="Желаемое состояние из Kubernetes: Services, Ingresses и DNSEndpoint CRD"
        />
      </FadeIn>

      <FadeIn delay={0.05}>
        <div className="mb-6 flex gap-3">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-subtle" />
            <input
              className="input-field pl-10"
              placeholder="Поиск по hostname, namespace, target..."
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                window.clearTimeout((window as unknown as { _dnsSearch?: number })._dnsSearch);
                (window as unknown as { _dnsSearch?: number })._dnsSearch = window.setTimeout(
                  () => setDebounced(e.target.value),
                  300
                );
              }}
            />
          </div>
        </div>
      </FadeIn>

      {isLoading && <LoadingState />}
      {error && <ErrorState message={error instanceof Error ? error.message : "Ошибка загрузки"} />}
      {!isLoading && !error && items.length === 0 && (
        <EmptyState message="DNS записи не найдены. Добавьте annotation external-dns.kubernetes.io/hostname к Service или Ingress." />
      )}

      {!isLoading && !error && items.length > 0 && (
        <FadeIn delay={0.1}>
          <div className="glass overflow-x-auto">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Hostname</th>
                  <th>Type</th>
                  <th>Target</th>
                  <th>Source</th>
                  <th>Namespace</th>
                  <th>Resource</th>
                  <th>TTL</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {items.map((r) => (
                  <tr key={r.id}>
                    <td className="font-medium text-fg">{r.hostname}</td>
                    <td><span className="badge">{r.record_type}</span></td>
                    <td className="font-mono text-xs">{r.target}</td>
                    <td><span className="badge">{r.source_type}</span></td>
                    <td>{r.namespace}</td>
                    <td>{r.resource}</td>
                    <td>{r.ttl ?? "—"}</td>
                    <td className="text-xs text-muted">{formatDate(r.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="mt-3 text-xs text-subtle">Всего: {data?.total ?? items.length}</p>
        </FadeIn>
      )}
    </div>
  );
}
