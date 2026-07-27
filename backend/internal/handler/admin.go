package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/konstpic/externaldns-web-ui/backend/internal/audit"
	"github.com/konstpic/externaldns-web-ui/backend/internal/auth"
	"github.com/konstpic/externaldns-web-ui/backend/internal/k8s"
	"github.com/konstpic/externaldns-web-ui/backend/internal/models"
)

type AdminHandler struct {
	k8s   *k8s.Client
	auth  *auth.Service
	audit *audit.Log
}

func NewAdmin(k *k8s.Client, authSvc *auth.Service, auditLog *audit.Log) *AdminHandler {
	return &AdminHandler{k8s: k, auth: authSvc, audit: auditLog}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	admin := func(pattern string, fn http.HandlerFunc) {
		mux.HandleFunc(pattern, auth.RequireAdmin(h.auth, fn))
	}

	admin("GET /api/v1/admin/overview", h.overview)
	admin("GET /api/v1/admin/settings", h.settings)
	admin("GET /api/v1/admin/audit", h.auditList)
	admin("GET /api/v1/admin/namespaces", h.namespaces)
	admin("GET /api/v1/admin/candidates", h.candidates)
	admin("POST /api/v1/admin/annotate", h.createAnnotate)
	admin("PUT /api/v1/admin/annotate", h.updateAnnotate)
	admin("GET /api/v1/admin/annotate", h.getAnnotate)
	admin("DELETE /api/v1/admin/annotate", h.removeAnnotate)
	admin("POST /api/v1/admin/dnsendpoints", h.createDNSEndpoint)
	admin("GET /api/v1/admin/dnsendpoints/{namespace}/{name}", h.getDNSEndpoint)
	admin("PUT /api/v1/admin/dnsendpoints/{namespace}/{name}", h.updateDNSEndpoint)
	admin("DELETE /api/v1/admin/dnsendpoints/{namespace}/{name}", h.deleteDNSEndpoint)
}

func (h *AdminHandler) overview(w http.ResponseWriter, r *http.Request) {
	data, err := h.k8s.AdminOverview(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AdminHandler) settings(w http.ResponseWriter, _ *http.Request) {
	cfg := h.auth.Config()
	writeJSON(w, http.StatusOK, map[string]any{
		"auth": map[string]any{
			"auth_required": cfg.AuthRequired,
			"oidc_enabled":  cfg.OIDCConfigured(),
			"issuer_url":    cfg.IssuerURL,
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

func (h *AdminHandler) auditList(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": h.audit.List(limit)})
}

func (h *AdminHandler) namespaces(w http.ResponseWriter, r *http.Request) {
	list, err := h.k8s.ListNamespaces(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *AdminHandler) candidates(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	list, err := h.k8s.ListCandidateResources(r.Context(), ns)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *AdminHandler) createAnnotate(w http.ResponseWriter, r *http.Request) {
	var req models.AnnotateRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	has, err := h.k8s.HasDNSAnnotation(r.Context(), req.Kind, req.Namespace, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if has {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "resource already has DNS annotation; use PUT to update"})
		return
	}
	actor := actorEmail(r)
	resource := req.Kind + "/" + req.Namespace + "/" + req.Name
	err = h.k8s.AnnotateResource(r.Context(), req)
	h.audit.Record(actor, "create_annotation", resource, req.Hostname, err == nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *AdminHandler) updateAnnotate(w http.ResponseWriter, r *http.Request) {
	var req models.AnnotateRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	actor := actorEmail(r)
	resource := req.Kind + "/" + req.Namespace + "/" + req.Name
	err := h.k8s.UpdateAnnotateResource(r.Context(), req)
	h.audit.Record(actor, "update_annotation", resource, req.Hostname, err == nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AdminHandler) getAnnotate(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if kind == "" || namespace == "" || name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind, namespace, name required"})
		return
	}
	data, err := h.k8s.GetAnnotateDetail(r.Context(), kind, namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AdminHandler) removeAnnotate(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if kind == "" || namespace == "" || name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind, namespace, name required"})
		return
	}
	actor := actorEmail(r)
	resource := kind + "/" + namespace + "/" + name
	err := h.k8s.RemoveDNSAnnotations(r.Context(), kind, namespace, name)
	h.audit.Record(actor, "remove_annotation", resource, "", err == nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminHandler) createDNSEndpoint(w http.ResponseWriter, r *http.Request) {
	var req models.DNSEndpointRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	actor := actorEmail(r)
	resource := "DNSEndpoint/" + req.Namespace + "/" + req.Name
	err := h.k8s.CreateDNSEndpoint(r.Context(), req)
	h.audit.Record(actor, "create_dnsendpoint", resource, req.DNSName, err == nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *AdminHandler) getDNSEndpoint(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	data, err := h.k8s.GetDNSEndpoint(r.Context(), namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *AdminHandler) updateDNSEndpoint(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	var req models.DNSEndpointRequest
	if err := readBodyJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.Namespace = namespace
	req.Name = name
	actor := actorEmail(r)
	resource := "DNSEndpoint/" + namespace + "/" + name
	err := h.k8s.UpdateDNSEndpoint(r.Context(), req)
	h.audit.Record(actor, "update_dnsendpoint", resource, req.DNSName, err == nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AdminHandler) deleteDNSEndpoint(w http.ResponseWriter, r *http.Request) {
	namespace := r.PathValue("namespace")
	name := r.PathValue("name")
	actor := actorEmail(r)
	resource := "DNSEndpoint/" + namespace + "/" + name
	err := h.k8s.DeleteDNSEndpoint(r.Context(), namespace, name)
	h.audit.Record(actor, "delete_dnsendpoint", resource, "", err == nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func actorEmail(r *http.Request) string {
	if claims := auth.UserFromContext(r.Context()); claims != nil && claims.Email != "" {
		return claims.Email
	}
	return "unknown"
}

func readBodyJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
