// Command gateway es el borde publico HTTP/WS de Bina Mail (BFF, ADR-001).
package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/gateway"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/ratelimit"
	"github.com/jessbludev/bina-mail/services/internal/resp"
)

func main() {
	var counter ratelimit.Counter
	if addr := config.Env("DRAGONFLY_ADDR", ""); addr != "" {
		client, err := resp.Dial(addr, 2*time.Second)
		if err != nil {
			log.Fatalf("gateway: cache no disponible en %s: %v", addr, err)
		}
		defer func() { _ = client.Close() }()
		counter = client
	} else {
		counter = ratelimit.NewMemoryCounter(nil)
	}
	limit := int64(60)
	if raw := config.Env("RATE_LIMIT_PER_MIN", ""); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			log.Fatal("gateway: RATE_LIMIT_PER_MIN debe ser un entero positivo")
		}
		limit = int64(parsed)
	}
	api := gateway.New(gateway.Config{
		AuthURL:       config.Env("AUTH_URL", "http://localhost:8081"),
		DirectoryURL:  config.Env("DIRECTORY_URL", "http://localhost:8082"),
		MailboxURL:    config.Env("MAILBOX_URL", "http://localhost:8083"),
		RealtimeURL:   config.Env("REALTIME_URL", "http://localhost:8084"),
		AllowedOrigin: config.Env("CORS_ALLOWED_ORIGIN", ""),
		Limiter: &ratelimit.FixedWindow{
			Counter: counter,
			Window:  time.Minute,
			Limit:   limit,
		},
	})
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("gateway"))
	addr := ":" + config.Env("PORT", "8080")
	log.Printf("gateway escuchando en %s", addr)
	// Sin WriteTimeout: el proxy WS mantiene conexiones de larga vida;
	// ReadHeaderTimeout e IdleTimeout acotan el abuso de conexiones.
	server := &http.Server{
		Addr:              addr,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
