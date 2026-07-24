package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/konstpic/externaldns-web-ui/backend/internal/auth"
	"github.com/konstpic/externaldns-web-ui/backend/internal/k8s"
)

type Handler struct {
	k8s  *k8s.Client
	auth *auth.Service
}

func New(k *k8s.Client, authSvc *auth.Service) *Handler {
	return &Handler{k8s: k, auth: authSvc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.health)

	protected := func(pattern string, fn http.HandlerFunc) {
		mux.HandleFunc(pattern, auth.RequireAuth(h.auth, fn))
	}
	admin := func(pattern string, fn http.HandlerFunc) {
		mux.HandleFunc(pattern, auth.RequireAdmin(h.auth, fn))
	}

	protected("GET /api/v1/overview", h.overview)
	protected("GET /api/v1/records", h.records)
	protected("GET /api/v1/sources", h.sources)
	protected("GET /api/v1/controller", h.controller)
	protected("GET /api/v1/logs", h.logs)
	admin("GET /api/v1/admin/settings", h.adminSettings)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	data, err := h.k8s.Overview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) records(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	data, err := h.k8s.ListRecords(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": data, "total": len(data)})
}

func (h *Handler) sources(w http.ResponseWriter, r *http.Request) {
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	data, err := h.k8s.ListSources(r.Context(), kind)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": data, "total": len(data)})
}

func (h *Handler) controller(w http.ResponseWriter, r *http.Request) {
	data, err := h.k8s.ControllerStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) logs(w http.ResponseWriter, r *http.Request) {
	tail := int64(100)
	if v := r.URL.Query().Get("tail"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			tail = n
		}
	}
	data, err := h.k8s.RecentLogs(r.Context(), tail)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": data})
}

func (h *Handler) adminSettings(w http.ResponseWriter, _ *http.Request) {
	cfg := h.auth.Config()
	writeJSON(w, http.StatusOK, map[string]any{
		"auth": map[string]any{
			"auth_required": cfg.AuthRequired,
			"oidc_enabled":  cfg.OIDCConfigured(),
			"issuer_url":    strings.TrimRight(cfg.IssuerURL, "/"),
			"client_id":     cfg.ClientID,
			"redirect_url":  cfg.RedirectURL,
			"scopes":        cfg.Scopes,
			"role_claim":    cfg.RoleClaim,
			"group_claim":   cfg.GroupClaim,
			"admin_roles":   cfg.AdminRoles,
			"frontend_url":  cfg.FrontendURL,
		},
		"app": map[string]any{
			"cluster_name":           envOr("CLUSTER_NAME", ""),
			"domain_filter":          envOr("DOMAIN_FILTER", ""),
			"external_dns_namespace": envOr("EXTERNALDNS_NAMESPACE", "external-dns"),
			"external_dns_deploy":    envOr("EXTERNALDNS_DEPLOYMENT", "external-dns"),
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenv(key string) string {
	// small helper without importing os in every handler method
	return lookupEnv(key)
}
