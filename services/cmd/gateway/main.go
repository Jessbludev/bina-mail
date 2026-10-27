// Command gateway es el borde publico HTTP/WS de Bina Mail (BFF, ADR-001).
package main

import (
	"log"
	"net/http"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/gateway"
	"github.com/jessbludev/bina-mail/services/internal/health"
)

func main() {
	api := gateway.New(gateway.Config{
		AuthURL:      config.Env("AUTH_URL", "http://localhost:8081"),
		DirectoryURL: config.Env("DIRECTORY_URL", "http://localhost:8082"),
		MailboxURL:   config.Env("MAILBOX_URL", "http://localhost:8083"),
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
