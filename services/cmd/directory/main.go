// Command directory gestiona el registro de claves publicas y fingerprints.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/directoryapi"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/resp"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

func main() {
	var keyCache resp.Cache
	if addr := config.Env("DRAGONFLY_ADDR", ""); addr != "" {
		client, err := resp.Dial(addr, 2*time.Second)
		if err != nil {
			log.Fatalf("directory: cache no disponible en %s: %v", addr, err)
		}
		defer func() { _ = client.Close() }()
		keyCache = client
	}
	api := directoryapi.New(directoryapi.Config{
		Supabase: supabase.New(config.Env("SUPABASE_URL", ""), config.Env("SUPABASE_SERVICE_ROLE_KEY", "")),
		Cache:    keyCache,
	})
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("directory"))
	addr := ":" + config.Env("PORT", "8082")
	log.Printf("directory escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}
