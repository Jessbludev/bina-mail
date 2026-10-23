// Command mailbox persiste y consulta sobres cifrados E2EE (ADR-004).
package main

import "github.com/jessbludev/bina-mail/services/internal/server"

func main() {
	server.Run("mailbox", "8083")
}
