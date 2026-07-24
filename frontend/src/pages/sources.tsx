import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { EmptyState, ErrorState, FadeIn, LoadingState, PageHeader } from "@/components/ui";
import { getSources } from "@/lib/api";
import { cn, pageShellClass } from "@/lib/utils";

const kinds = ["", "Service", "Ingress", "DNSEndpoint"] as const;

export function SourcesPage() {
  const [kind, setKind] = useState<(typeof kinds)[number]>("");
  const { data, isLoading, error } = useQuery({
    queryKey: ["sources", kind],
    queryFn: () => getSources(kind || undefined),
  });

  const items = data?.items ?? [];

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <PageHeader
          title="Источники"
          subtitle="Kubernetes ресурсы, из которых ExternalDNS формирует DNS записи"
        />
      </FadeIn>

      <FadeIn delay={0.05}>
        <div className="mb-6 flex flex-wrap gap-2">
          {kinds.map((k) => (
            <button
              key={k || "all"}
              type="button"
              className={cn(
                "rounded-xl px-4 py-2 text-sm font-medium transition-colors",
                kind === k
                  ? "bg-primary/15 text-primary"
                  : "bg-surface/10 text-muted hover:text-fg-secondary"
              )}
              onClick={() => setKind(k)}
            >
              {k || "Все"}
            </button>
          ))}
        </div>
      </FadeIn>

      {isLoading && <LoadingState />}
      {error && <ErrorState message={error instanceof Error ? error.message : "Ошибка загрузки"} />}
      {!isLoading && !error && items.length === 0 && (
        <EmptyState message="Источники с DNS annotations не найдены." />
      )}

      {!isLoading && !error && items.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2">
          {items.map((s, i) => (
            <FadeIn key={`${s.kind}/${s.namespace}/${s.name}`} delay={i * 0.03}>
              <div className="glass-hover p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <span className="badge">{s.kind}</span>
                    <h3 className="mt-2 font-semibold text-fg">
                      {s.namespace}/{s.name}
                    </h3>
                    <p className="mt-1 text-xs text-subtle">Age: {s.age}</p>
                  </div>
                  {s.target && s.target !== "—" && (
                    <code className="rounded-lg bg-surface/10 px-2 py-1 text-xs text-muted">
                      {s.target}
                    </code>
                  )}
                </div>
                <ul className="mt-4 space-y-1">
                  {s.hostnames.map((h) => (
                    <li key={h} className="font-mono text-sm text-primary">
                      {h}
                    </li>
                  ))}
                </ul>
                {s.annotations && Object.keys(s.annotations).length > 0 && (
                  <dl className="mt-4 space-y-1 border-t border-border/10 pt-3">
                    {Object.entries(s.annotations).map(([k, v]) => (
                      <div key={k} className="grid grid-cols-[1fr,1fr] gap-2 text-xs">
                        <dt className="truncate text-subtle">{k}</dt>
                        <dd className="truncate text-muted">{v}</dd>
                      </div>
                    ))}
                  </dl>
                )}
              </div>
            </FadeIn>
          ))}
        </div>
      )}
    </div>
  );
}
