package auth

import (
	"encoding/json"
	"io"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc *Service) {
	mux.HandleFunc("GET /api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		svc.LoginRedirect(w, r)
	})
	mux.HandleFunc("GET /api/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		svc.Callback(w, r)
	})
	mux.HandleFunc("POST /api/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		pair, err := svc.Refresh(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
			return
		}
		writeJSON(w, http.StatusOK, pair)
	})
	mux.HandleFunc("POST /api/auth/logout", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/auth/me", func(w http.ResponseWriter, r *http.Request) {
		claims, err := svc.UserFromRequest(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":           claims.UserID,
			"email":        claims.Email,
			"display_name": claims.DisplayName,
			"roles":        claims.Roles,
			"is_admin":     svc.Config().IsAdmin(claims.Roles),
		})
	})
	mux.HandleFunc("GET /api/public/auth-methods", func(w http.ResponseWriter, _ *http.Request) {
		cfg := svc.Config()
		writeJSON(w, http.StatusOK, map[string]any{
			"auth_required": cfg.AuthRequired,
			"oidc_enabled":  cfg.OIDCConfigured(),
			"login_url":     "/api/auth/login",
		})
	})
}

func RequireAuth(svc *Service, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := svc.Config()
		if !cfg.AuthRequired {
			next(w, r)
			return
		}
		claims, err := svc.UserFromRequest(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		ctx := withClaims(r.Context(), claims)
		next(w, r.WithContext(ctx))
	}
}

func RequireAdmin(svc *Service, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(svc, func(w http.ResponseWriter, r *http.Request) {
		claims := claimsFromContext(r.Context())
		if claims == nil || !svc.Config().IsAdmin(claims.Roles) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin required"})
			return
		}
		next(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}
