// Command mailbox persiste y consulta sobres cifrados E2EE (ADR-004).
package main

import (
	"log"
	"net/http"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/mailboxapi"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

func main() {
	api := mailboxapi.New(mailboxapi.Config{
		Supabase:         supabase.New(config.Env("SUPABASE_URL", ""), config.Env("SUPABASE_SERVICE_ROLE_KEY", "")),
		MaxEnvelopeBytes: 1 << 20,
	})
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("mailbox"))
	addr := ":" + config.Env("PORT", "8083")
	log.Printf("mailbox escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}
