package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/konstpic/externaldns-web-ui/backend/internal/handler"
	"github.com/konstpic/externaldns-web-ui/backend/internal/k8s"
)

func main() {
	addr := envOr("LISTEN_ADDR", ":8080")

	client, err := k8s.NewFromEnv()
	if err != nil {
		log.Fatalf("kubernetes client: %v", err)
	}

	h := handler.New(client)
	mux := http.NewServeMux()
	h.Register(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("externaldns-web-ui backend listening on %s", addr)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
