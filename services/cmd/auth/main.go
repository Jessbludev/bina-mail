// Command auth gestiona identidad, sesiones y rotacion de tokens (MAP.md).
package main

import (
	"log"
	"net/http"

	"github.com/jessbludev/bina-mail/services/internal/authapi"
	"github.com/jessbludev/bina-mail/services/internal/cache"
	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

func main() {
	var sessionCache authapi.SessionCache
	if addr := config.Env("CACHE_ADDR", ""); addr != "" {
		conn, err := cache.Dial(addr)
		if err != nil {
			log.Fatalf("auth: cache no disponible en %s: %v", addr, err)
		}
		defer func() { _ = conn.Close() }()
		sessionCache = conn
	}
	api := authapi.New(authapi.Config{
		Supabase:   supabase.New(config.Env("SUPABASE_URL", ""), config.Env("SUPABASE_SERVICE_ROLE_KEY", "")),
		JWTSecret:  config.Env("SUPABASE_JWT_SECRET", ""),
		SessionTTL: session.TTLDefault,
		Cache:      sessionCache,
	})
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("auth"))
	addr := ":" + config.Env("PORT", "8081")
	log.Printf("auth escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}
