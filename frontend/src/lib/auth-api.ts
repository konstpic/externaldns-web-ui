export interface User {
  id: string;
  email: string;
  display_name: string;
  roles: string[];
  is_admin?: boolean;
}

interface AuthMethods {
  auth_required: boolean;
  oidc_enabled: boolean;
  login_url: string;
}

interface AdminSettings {
  auth: {
    auth_required: boolean;
    oidc_enabled: boolean;
    issuer_url: string;
    client_id: string;
    redirect_url: string;
    scopes: string[];
    role_claim: string;
    group_claim: string;
    admin_roles: string[];
    frontend_url: string;
  };
  app: {
    cluster_name: string;
    domain_filter: string;
    external_dns_namespace: string;
    external_dns_deploy: string;
  };
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

let refreshPromise: Promise<boolean> | null = null;

function getAccessToken() {
  return localStorage.getItem("access_token");
}

function getRefreshToken() {
  return localStorage.getItem("refresh_token");
}

async function refreshTokens(): Promise<boolean> {
  const rt = getRefreshToken();
  if (!rt) return false;
  const res = await fetch("/api/auth/refresh", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: rt }),
  });
  if (!res.ok) return false;
  const data = await res.json();
  localStorage.setItem("access_token", data.access_token);
  localStorage.setItem("refresh_token", data.refresh_token);
  return true;
}

async function ensureRefreshed(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = refreshTokens().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const token = getAccessToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  let res = await fetch(path, { ...init, headers });

  if (res.status === 401 && getRefreshToken()) {
    const ok = await ensureRefreshed();
    if (ok) {
      headers.set("Authorization", `Bearer ${getAccessToken()}`);
      res = await fetch(path, { ...init, headers });
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const getAuthMethods = () => apiFetch<AuthMethods>("/api/public/auth-methods");
export const getMe = () => apiFetch<User>("/api/auth/me");
export const getAdminSettings = () => apiFetch<AdminSettings>("/api/v1/admin/settings");

export function getLoginUrl() {
  return "/api/auth/login";
}

export async function logout() {
  const at = getAccessToken();
  const rt = getRefreshToken();
  await fetch("/api/auth/logout", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(at ? { Authorization: `Bearer ${at}` } : {}),
    },
    body: JSON.stringify({ refresh_token: rt }),
  }).catch(() => {});
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}
