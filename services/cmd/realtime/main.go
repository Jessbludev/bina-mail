// Command realtime emite senales WS sin contenido ni metadatos,
// alimentado por los eventos de Redpanda (ADR-005/ADR-008).
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jessbludev/bina-mail/services/internal/config"
	"github.com/jessbludev/bina-mail/services/internal/events"
	"github.com/jessbludev/bina-mail/services/internal/health"
	"github.com/jessbludev/bina-mail/services/internal/kafka"
	"github.com/jessbludev/bina-mail/services/internal/realtimeapi"
)

func main() {
	api := realtimeapi.New(realtimeapi.Config{
		AuthURL:      config.Env("AUTH_URL", "http://localhost:8081"),
		DirectoryURL: config.Env("DIRECTORY_URL", "http://localhost:8082"),
	})
	if brokers := config.CSV("KAFKA_BROKERS"); len(brokers) > 0 {
		consumer, err := kafka.NewConsumer(
			brokers,
			config.Env("KAFKA_GROUP", "bina-realtime"),
			config.Env("KAFKA_TOPIC", events.TopicMailboxEnvelopeStored),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Close()
		go consume(consumer, api)
	}
	root := http.NewServeMux()
	root.Handle("/", api)
	root.Handle("/healthz", health.Handler("realtime"))
	addr := ":" + config.Env("PORT", "8084")
	log.Printf("realtime escuchando en %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}

// consume transforma cada evento del bus en una senal WS sin contenido.
func consume(consumer *kafka.Consumer, api *realtimeapi.Handler) {
	consumer.Run(context.Background(), func(_ context.Context, value []byte) {
		event, err := events.ParseEnvelopeStored(value)
		if err != nil {
			return
		}
		_ = api.Signal(event.RecipientFingerprint)
	})
}
