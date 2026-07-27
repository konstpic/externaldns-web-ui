package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/konstpic/externaldns-web-ui/backend/internal/audit"
	"github.com/konstpic/externaldns-web-ui/backend/internal/auth"
	"github.com/konstpic/externaldns-web-ui/backend/internal/handler"
	"github.com/konstpic/externaldns-web-ui/backend/internal/k8s"
)

func main() {
	addr := envOr("LISTEN_ADDR", ":8080")
	authCfg := auth.LoadConfig()

	if authCfg.AuthRequired && !authCfg.JWTConfigured() {
		log.Fatal("JWT_SECRET is required when AUTH_REQUIRED=true")
	}

	authSvc, err := auth.NewService(authCfg)
	if err != nil {
		log.Fatalf("auth service: %v", err)
	}

	client, err := k8s.NewFromEnv()
	if err != nil {
		log.Fatalf("kubernetes client: %v", err)
	}

	auditLog := audit.New(500)

	mux := http.NewServeMux()
	auth.RegisterRoutes(mux, authSvc)
	handler.New(client, authSvc).Register(mux)
	handler.NewAdmin(client, authSvc, auditLog).Register(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("externaldns-web-ui backend listening on %s (auth_required=%v oidc=%v)",
		addr, authCfg.AuthRequired, authCfg.OIDCConfigured())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
