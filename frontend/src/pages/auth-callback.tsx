import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { pageShellClass } from "@/lib/utils";

export function AuthCallbackPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const setAuth = useAuthStore((s) => s.setAuth);
  const loadUser = useAuthStore((s) => s.loadUser);
  const [error, setError] = useState("");

  useEffect(() => {
    const access = searchParams.get("access_token");
    const refresh = searchParams.get("refresh_token");
    if (!access || !refresh) {
      setError("Missing tokens");
      return;
    }
    setAuth(access, refresh);
    loadUser()
      .then(() => navigate("/", { replace: true }))
      .catch(() => setError("Не удалось загрузить профиль"));
  }, [searchParams, setAuth, loadUser, navigate]);

  return (
    <div className={pageShellClass}>
      <div className="flex flex-col items-center justify-center gap-3 py-20">
        {!error ? (
          <>
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
            <p className="text-sm text-muted">Завершение входа...</p>
          </>
        ) : (
          <p className="text-sm text-danger">{error}</p>
        )}
      </div>
    </div>
  );
}
