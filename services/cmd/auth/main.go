// Command auth gestiona identidad, sesiones y rotación de tokens (MAP.md).
package main

import "github.com/jessbludev/bina-mail/services/internal/server"

func main() {
	server.Run("auth", "8081")
}
