// Package server arranca un servicio HTTP con healthcheck.
package server

import (
	"log"
	"net/http"
	"os"

	"github.com/jessbludev/bina-mail/services/internal/health"
)

// Run inicia el servicio en el puerto indicado (variable de entorno PORT).
// Registra solo eventos operativos, sin metadatos de mensajes (ADR-005).
func Run(service, defaultPort string) {
	addr := ":" + env("PORT", defaultPort)
	mux := http.NewServeMux()
	mux.Handle("/", health.Handler(service))
	log.Printf("%s escuchando en %s", service, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
