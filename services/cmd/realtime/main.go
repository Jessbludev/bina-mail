// Command realtime emite senales WS sin contenido ni metadatos (ADR-005/ADR-008).
package main

import (
	"log"
	"net/http"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/realtime"
)

func main() {
	api := realtime.New(realtime.Config{
		AuthURL:      config.Env("AUTH_URL", "http://localhost:8081"),
		DirectoryURL: config.Env("DIRECTORY_URL", "http://localhost:8082"),
	})
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("realtime"))
	addr := ":" + config.Env("PORT", "8084")
	log.Printf("realtime escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}
