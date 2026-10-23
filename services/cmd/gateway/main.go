// Command gateway es el borde público HTTP/WS de Bina Mail (BFF, ADR-001).
// Fase 1: healthcheck. Chi y las rutas reales llegan en Fase 2.
package main

import "github.com/jessbludev/bina-mail/services/internal/server"

func main() {
	server.Run("gateway", "8080")
}
