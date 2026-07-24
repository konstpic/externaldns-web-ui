package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/konstpic/externaldns-web-ui/backend/internal/k8s"
)

type Handler struct {
	k8s *k8s.Client
}

func New(k *k8s.Client) *Handler {
	return &Handler{k8s: k}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("GET /api/v1/overview", h.overview)
	mux.HandleFunc("GET /api/v1/records", h.records)
	mux.HandleFunc("GET /api/v1/sources", h.sources)
	mux.HandleFunc("GET /api/v1/controller", h.controller)
	mux.HandleFunc("GET /api/v1/logs", h.logs)
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
