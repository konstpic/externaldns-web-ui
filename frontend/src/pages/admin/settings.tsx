import { useQuery } from "@tanstack/react-query";
import { Settings, Shield, KeyRound } from "lucide-react";
import { getAdminSettings } from "@/lib/auth-api";
import { ErrorState, FadeIn, LoadingState, PageHeader } from "@/components/ui";
import { cn, pageShellClass } from "@/lib/utils";

export function AdminSettingsPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["admin-settings"],
    queryFn: getAdminSettings,
  });

  if (isLoading) return <div className={pageShellClass}><LoadingState /></div>;
  if (error || !data) {
    return (
      <div className={pageShellClass}>
        <ErrorState message={error instanceof Error ? error.message : "Ошибка загрузки настроек"} />
      </div>
    );
  }

  return (
    <div className={cn(pageShellClass, "py-10")}>
      <FadeIn>
        <PageHeader
          title="Настройки"
          subtitle="Конфигурация OIDC/Authentik и параметры приложения (только для администраторов)"
        />
      </FadeIn>

      <div className="grid gap-6 lg:grid-cols-2">
        <FadeIn delay={0.05}>
          <section className="glass p-6">
            <div className="mb-4 flex items-center gap-2">
              <Shield className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">Аутентификация</h2>
            </div>
            <dl className="space-y-4 text-sm">
              <Row label="Auth required" value={data.auth.auth_required ? "Да" : "Нет"} />
              <Row label="OIDC enabled" value={data.auth.oidc_enabled ? "Да" : "Нет"} />
              <Row label="Issuer URL" value={data.auth.issuer_url || "—"} mono />
              <Row label="Client ID" value={data.auth.client_id || "—"} mono />
              <Row label="Redirect URL" value={data.auth.redirect_url || "—"} mono />
              <Row label="Scopes" value={data.auth.scopes.join(" ") || "—"} />
              <Row label="Role claim" value={data.auth.role_claim} />
              <Row label="Group claim" value={data.auth.group_claim} />
              <Row label="Admin roles" value={data.auth.admin_roles.join(", ")} />
              <Row label="Frontend URL" value={data.auth.frontend_url || "—"} mono />
            </dl>
          </section>
        </FadeIn>

        <FadeIn delay={0.1}>
          <section className="glass p-6">
            <div className="mb-4 flex items-center gap-2">
              <Settings className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">Приложение</h2>
            </div>
            <dl className="space-y-4 text-sm">
              <Row label="Cluster" value={data.app.cluster_name || "—"} />
              <Row label="Domain filter" value={data.app.domain_filter || "—"} />
              <Row label="ExternalDNS NS" value={data.app.external_dns_namespace} />
              <Row label="ExternalDNS deploy" value={data.app.external_dns_deploy} />
            </dl>
          </section>
        </FadeIn>
      </div>

      <FadeIn delay={0.15}>
        <section className="mt-6 glass p-6">
          <div className="mb-3 flex items-center gap-2">
            <KeyRound className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold text-fg">Authentik</h2>
          </div>
          <p className="text-sm text-muted">
            Client secret и JWT secret хранятся в Kubernetes Secret и не отображаются в UI.
            OAuth2 provider создаётся через Authentik blueprint в репозитории k8s-konstpic.
          </p>
          <pre className="mt-4 overflow-x-auto rounded-xl bg-black/30 p-4 text-xs text-fg-secondary">
{`Redirect URI: ${data.auth.redirect_url}
Application slug: externaldns-web-ui
Admin groups: ${data.auth.admin_roles.join(", ")}`}
          </pre>
        </section>
      </FadeIn>
    </div>
  );
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wider text-subtle">{label}</dt>
      <dd className={cn("mt-1 text-fg-secondary", mono && "break-all font-mono text-xs")}>{value}</dd>
    </div>
  );
}
