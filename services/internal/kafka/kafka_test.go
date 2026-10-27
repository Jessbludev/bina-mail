package kafka

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/events"
	"github.com/twmb/franz-go/pkg/kfake"
)

func TestProducirYConsumir(t *testing.T) {
	cluster, err := kfake.NewCluster(kfake.SeedTopics(1, events.TopicMailboxEnvelopeStored))
	if err != nil {
		t.Fatalf("kfake.NewCluster fallo: %v", err)
	}
	defer cluster.Close()

	producer, err := NewProducer(cluster.ListenAddrs(), events.TopicMailboxEnvelopeStored)
	if err != nil {
		t.Fatalf("NewProducer fallo: %v", err)
	}
	defer producer.Close()

	fingerprint := strings.Repeat("a", 64)
	payload := events.NewEnvelopeStored("sobre-1", fingerprint, time.Unix(1700000000, 0).UTC())
	if err := producer.Publish(context.Background(), []byte(fingerprint), payload); err != nil {
		t.Fatalf("Publish fallo: %v", err)
	}

	consumer, err := NewConsumer(cluster.ListenAddrs(), "grupo-test", events.TopicMailboxEnvelopeStored)
	if err != nil {
		t.Fatalf("NewConsumer fallo: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(chan []byte, 1)
	go consumer.Run(ctx, func(_ context.Context, value []byte) {
		select {
		case received <- value:
		default:
		}
	})

	select {
	case value := <-received:
		event, err := events.ParseEnvelopeStored(value)
		if err != nil {
			t.Fatalf("ParseEnvelopeStored fallo: %v", err)
		}
		if event.EnvelopeID != "sobre-1" || event.RecipientFingerprint != fingerprint {
			t.Fatalf("evento inesperado: %+v", event)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("no se recibio el evento del cluster de pruebas")
	}
}
