// Package events define el contrato de eventos internos de Bina Mail (ADR-008).
package events

import "context"

// Topics y nombres de evento con version logica. El transporte actual
// entrega por HTTP interno; Redpanda asume el mismo contrato en Fase 5.
const (
	TopicSignals         = "bina.mail.signals.v1"
	EventEnvelopeStored  = "envelope.stored"
)

// EnvelopeStored es la senal sin contenido que avisa un sobre nuevo.
// No transporta claves ni ciphertext: el cliente refresca su buzon.
type EnvelopeStored struct {
	Event                string `json:"event"`
	ID                   string `json:"id"`
	RecipientFingerprint string `json:"recipient_fingerprint"`
}

// Publisher entrega senales al servicio realtime.
type Publisher interface {
	PublishEnvelopeStored(ctx context.Context, e EnvelopeStored) error
}
