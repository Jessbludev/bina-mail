// Command mailbox persiste y consulta sobres cifrados E2EE (ADR-004) y
// publica senales de sobre nuevo en Redpanda (ADR-008).
package main

import (
	"log"
	"net/http"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/events"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/kafka"
	"github.com/jessbludev/bina-mail/services/internal/mailboxapi"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

func main() {
	cfg := mailboxapi.Config{
		Supabase:         supabase.New(config.Env("SUPABASE_URL", ""), config.Env("SUPABASE_SERVICE_ROLE_KEY", "")),
		MaxEnvelopeBytes: 1 << 20,
	}
	if brokers := config.CSV("KAFKA_BROKERS"); len(brokers) > 0 {
		producer, err := kafka.NewProducer(brokers, config.Env("KAFKA_TOPIC", events.TopicMailboxEnvelopeStored))
		if err != nil {
			log.Fatal(err)
		}
		defer producer.Close()
		cfg.Publisher = producer
	}
	api := mailboxapi.New(cfg)
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("mailbox"))
	addr := ":" + config.Env("PORT", "8083")
	log.Printf("mailbox escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}
