import { useQuery } from "@tanstack/react-query";
import { getAdminAudit } from "@/lib/admin-api";
import { ErrorState, FadeIn, LoadingState } from "@/components/ui";
import { formatDate } from "@/lib/utils";

export function AdminAuditPage() {
  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ["admin-audit"],
    queryFn: () => getAdminAudit(200),
    refetchInterval: 30_000,
  });

  const items = data?.items ?? [];

  return (
    <div>
      <div className="mb-4 flex justify-end">
        <button type="button" className="btn-secondary" onClick={() => refetch()} disabled={isFetching}>
          {isFetching ? "Обновление..." : "Обновить"}
        </button>
      </div>

      {isLoading && <LoadingState />}
      {error && <ErrorState message={error instanceof Error ? error.message : "Ошибка"} />}

      {!isLoading && !error && (
        <FadeIn>
          <div className="glass overflow-x-auto">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Время</th>
                  <th>Актор</th>
                  <th>Действие</th>
                  <th>Ресурс</th>
                  <th>Детали</th>
                  <th>Статус</th>
                </tr>
              </thead>
              <tbody>
                {items.map((e) => (
                  <tr key={e.id}>
                    <td className="text-xs text-muted whitespace-nowrap">{formatDate(e.timestamp)}</td>
                    <td>{e.actor}</td>
                    <td><span className="badge">{e.action}</span></td>
                    <td className="font-mono text-xs">{e.resource}</td>
                    <td className="text-xs">{e.detail || "—"}</td>
                    <td>
                      <span className={e.success ? "badge-success" : "badge-danger"}>
                        {e.success ? "OK" : "Fail"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {items.length === 0 && (
              <p className="px-4 py-10 text-center text-sm text-muted">Audit log пуст — изменения появятся после admin операций</p>
            )}
          </div>
        </FadeIn>
      )}
    </div>
  );
}
