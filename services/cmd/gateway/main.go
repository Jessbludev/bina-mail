// Command gateway es el borde publico HTTP/WS de Bina Mail (BFF, ADR-001).
package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/jessbludev/bina-mail/services/internal/cache"
	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/gateway"
	"github.com/jessbludev/bina-mail/services/internal/health"
)

func main() {
	var limiter gateway.RateLimiter
	if addr := config.Env("CACHE_ADDR", ""); addr != "" {
		conn, err := cache.Dial(addr)
		if err != nil {
			log.Fatalf("gateway: cache no disponible en %s: %v", addr, err)
		}
		defer func() { _ = conn.Close() }()
		limiter = conn
	}
	rateLimit := 240
	if raw := config.Env("RATE_LIMIT", ""); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			log.Fatal("gateway: RATE_LIMIT debe ser un entero positivo")
		}
		rateLimit = parsed
	}
	api := gateway.New(gateway.Config{
		AuthURL:      config.Env("AUTH_URL", "http://localhost:8081"),
		DirectoryURL: config.Env("DIRECTORY_URL", "http://localhost:8082"),
		MailboxURL:   config.Env("MAILBOX_URL", "http://localhost:8083"),
		Cache:        limiter,
		RateLimit:    rateLimit,
	})
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("gateway"))
	addr := ":" + config.Env("PORT", "8080")
	log.Printf("gateway escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}
