// Command directory gestiona el registro de claves públicas y fingerprints.
package main

import "github.com/jessbludev/bina-mail/services/internal/server"

func main() {
	server.Run("directory", "8082")
}
