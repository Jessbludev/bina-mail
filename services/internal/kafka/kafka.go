// Package kafka encapsula los clientes Redpanda (franz-go) detras de una
// superficie minima. Es la unica dependencia externa de ejecucion junto a
// gorilla/websocket (ADR-008).
package kafka

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer publica registros de forma sincrona en un tema fijo.
type Producer struct {
	client *kgo.Client
	topic  string
}

// NewProducer crea el productor con confirmacion de todas las replicas
// en sincronia.
func NewProducer(brokers []string, topic string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID("bina-mail-producer"),
		kgo.RecordDeliveryTimeout(3*time.Second),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		return nil, err
	}
	return &Producer{client: client, topic: topic}, nil
}

// Publish escribe un registro y espera su confirmacion.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	return p.client.ProduceSync(ctx, &kgo.Record{Topic: p.topic, Key: key, Value: value})
}

// Close libera el cliente subyacente.
func (p *Producer) Close() {
	p.client.Close()
}

// Consumer consume un tema como grupo con entrega al menos una vez.
type Consumer struct {
	client *kgo.Client
}

// NewConsumer crea el consumidor del grupo indicado, leyendo desde el
// inicio del tema al crearse el grupo.
func NewConsumer(brokers []string, group, topic string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ClientID("bina-mail-consumer"),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{client: client}, nil
}

// Run invoca el manejador por registro hasta cancelar el contexto.
// La senal es idempotente: las reentregas duplicadas son inofensivas.
func (c *Consumer) Run(ctx context.Context, handler func(ctx context.Context, value []byte)) {
	for {
		records := c.client.PollRecords(ctx, 100)
		if ctx.Err() != nil {
			return
		}
		for _, record := range records {
			handler(ctx, record.Value)
		}
	}
}

// Close libera el cliente subyacente.
func (c *Consumer) Close() {
	c.client.Close()
}
