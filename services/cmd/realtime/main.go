// Command realtime emite señales WS sin contenido ni metadatos (ADR-005).
package main

import "github.com/jessbludev/bina-mail/services/internal/server"

func main() {
	server.Run("realtime", "8084")
}
