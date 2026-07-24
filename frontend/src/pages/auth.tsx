import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Globe, Loader2, LogIn, Shield } from "lucide-react";
import { ApiError, getAuthMethods, getLoginUrl } from "@/lib/auth-api";
import { useAuthStore } from "@/lib/auth-store";
import { cn, pageShellClass } from "@/lib/utils";
import { FadeIn } from "@/components/ui";

export function AuthPage() {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [oidcEnabled, setOidcEnabled] = useState(true);

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/", { replace: true });
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    getAuthMethods()
      .then((m) => setOidcEnabled(m.oidc_enabled))
      .catch(() => setOidcEnabled(false));
  }, []);

  async function handleOIDCLogin() {
    setLoading(true);
    setError("");
    try {
      window.location.href = getLoginUrl();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "OIDC недоступен");
      setLoading(false);
    }
  }

  return (
    <div className={cn(pageShellClass, "flex min-h-dvh items-center justify-center py-16")}>
      <FadeIn>
        <div className="mx-auto w-full max-w-md">
          <div className="mb-8 text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/15 text-primary">
              <Globe className="h-6 w-6" />
            </div>
            <h1 className="text-3xl font-bold text-fg">ExternalDNS UI</h1>
            <p className="mt-2 text-sm text-muted">Войдите через Authentik для доступа к панели</p>
          </div>

          <div className="glass p-8">
            {loading && (
              <div className="mb-4 flex items-center justify-center gap-2 text-sm text-muted">
                <Loader2 className="h-4 w-4 animate-spin" />
                Авторизация...
              </div>
            )}

            {error && <p className="mb-4 text-center text-sm text-danger">{error}</p>}

            <button
              type="button"
              className="btn-primary w-full"
              onClick={handleOIDCLogin}
              disabled={loading || !oidcEnabled}
            >
              <LogIn className="h-4 w-4" />
              Войти через Authentik
            </button>

            {!oidcEnabled && (
              <p className="mt-4 text-center text-xs text-warning">
                OIDC не настроен. Проверьте переменные OIDC_* и JWT_SECRET на backend.
              </p>
            )}

            <div className="mt-6 flex items-start gap-3 rounded-xl border border-border/10 bg-surface/5 p-4 text-xs text-muted">
              <Shield className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
              <p>
                Доступ только для авторизованных пользователей. Администраторы Authentik получают доступ к
                настройкам через группы, указанные в <code className="text-fg-secondary">ADMIN_ROLES</code>.
              </p>
            </div>
          </div>

          <p className="mt-6 text-center text-xs text-subtle">
            Нет доступа? Обратитесь к администратору{" "}
            <Link to="/" className="text-primary hover:text-primary-hover">
              Authentik
            </Link>
          </p>
        </div>
      </FadeIn>
    </div>
  );
}
