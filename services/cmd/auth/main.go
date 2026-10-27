// Command auth gestiona sesiones opacas de corta vida (MAP.md).
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/authapi"
	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/resp"
	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

func main() {
	var sessionCache resp.Cache
	if addr := config.Env("DRAGONFLY_ADDR", ""); addr != "" {
		client, err := resp.Dial(addr, 2*time.Second)
		if err != nil {
			log.Fatalf("auth: cache no disponible en %s: %v", addr, err)
		}
		defer func() { _ = client.Close() }()
		sessionCache = client
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
