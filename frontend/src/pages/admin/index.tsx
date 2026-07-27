import { useQuery } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle2, Globe, Layers } from "lucide-react";
import { getAdminOverview } from "@/lib/admin-api";
import { ErrorState, FadeIn, LoadingState, StatCard } from "@/components/ui";

export function AdminOverviewPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["admin-overview"],
    queryFn: getAdminOverview,
  });

  if (isLoading) return <LoadingState />;
  if (error || !data) {
    return <ErrorState message={error instanceof Error ? error.message : "Ошибка загрузки"} />;
  }

  return (
    <div>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <FadeIn>
          <StatCard label="DNS записи" value={data.total_records} icon={<Globe className="h-5 w-5" />} />
        </FadeIn>
        <FadeIn delay={0.05}>
          <StatCard
            label="Без DNS"
            value={data.unmanaged_sources}
            hint="LoadBalancer/Ingress без annotation"
            icon={<AlertTriangle className="h-5 w-5" />}
          />
        </FadeIn>
        <FadeIn delay={0.1}>
          <StatCard label="Namespaces" value={data.namespaces} icon={<Layers className="h-5 w-5" />} />
        </FadeIn>
        <FadeIn delay={0.15}>
          <StatCard
            label="Controller"
            value={data.controller_ready ? "Ready" : "Down"}
            icon={<CheckCircle2 className="h-5 w-5" />}
          />
        </FadeIn>
      </div>

      <FadeIn delay={0.2}>
        <div className="mt-8 glass p-6">
          <h2 className="text-lg font-semibold text-fg">Возможности администратора</h2>
          <ul className="mt-4 space-y-2 text-sm text-muted">
            <li>• Добавление <code className="text-fg-secondary">external-dns.kubernetes.io/hostname</code> к Service/Ingress</li>
            <li>• Создание DNSEndpoint CRD для произвольных записей</li>
            <li>• Удаление DNS annotations и DNSEndpoint</li>
            <li>• Просмотр audit log всех изменений</li>
          </ul>
          <p className="mt-4 text-xs text-subtle">
            Кластер: {data.cluster_name} · Domain filter: {data.domain_filter || "—"}
          </p>
        </div>
      </FadeIn>
    </div>
  );
}
