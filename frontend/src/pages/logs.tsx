import { useQuery } from "@tanstack/react-query";
import { FadeIn, LoadingState, PageHeader, ErrorState, EmptyState } from "@/components/ui";
import { getLogs } from "@/lib/api";
import { cn, pageShellClass } from "@/lib/utils";

export function LogsPage() {
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["logs"],
    queryFn: () => getLogs(200),
    refetchInterval: 15_000,
  });

  const lines = data?.items ?? [];

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
          <PageHeader
            title="Логи ExternalDNS"
            subtitle="Последние строки из pod ExternalDNS controller"
          />
          <button type="button" className="btn-secondary" onClick={() => refetch()} disabled={isFetching}>
            {isFetching ? "Обновление..." : "Обновить"}
          </button>
        </div>
      </FadeIn>

      {isLoading && <LoadingState />}
      {error && <ErrorState message={error instanceof Error ? error.message : "Не удалось получить логи"} />}
      {!isLoading && !error && lines.length === 0 && (
        <EmptyState message="Логи недоступны. Проверьте RBAC и наличие pod ExternalDNS." />
      )}

      {!isLoading && !error && lines.length > 0 && (
        <FadeIn delay={0.05}>
          <div className="glass max-h-[70vh] overflow-auto p-4 font-mono text-xs leading-relaxed text-fg-secondary">
            {lines.map((l, i) => (
              <div key={i} className="whitespace-pre-wrap break-all border-b border-border/5 py-1">
                {l.message}
              </div>
            ))}
          </div>
        </FadeIn>
      )}
    </div>
  );
}
