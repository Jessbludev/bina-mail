// Package events define el esquema JSON de los eventos del bus (Redpanda),
// alineado con packages/contracts (ADR-007 y ADR-008).
package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/validate"
)

const (
	// VersionEvents es la version del esquema de eventos.
	VersionEvents = 1
	// TypeMailboxEnvelopeStored indica un sobre persistido en el buzon.
	TypeMailboxEnvelopeStored = "mailbox.envelope.stored"
	// TopicMailboxEnvelopeStored es el tema Redpanda del evento.
	TopicMailboxEnvelopeStored = "bina-mail.mailbox.envelope-stored.v1"
)

// EnvelopeStored es la senal de sobre nuevo. Solo contiene el minimo de
// enrutado; sin contenido ni metadatos de mensaje (politica zero-log).
type EnvelopeStored struct {
	Version              int    `json:"version"`
	Type                 string `json:"type"`
	EnvelopeID           string `json:"envelope_id"`
	RecipientFingerprint string `json:"recipient_fingerprint"`
	CreatedAtUnix        int64  `json:"created_at_unix"`
}

// NewEnvelopeStored serializa el evento para su publicacion.
func NewEnvelopeStored(envelopeID, recipientFingerprint string, now time.Time) []byte {
	payload, _ := json.Marshal(EnvelopeStored{
		Version:              VersionEvents,
		Type:                 TypeMailboxEnvelopeStored,
		EnvelopeID:           envelopeID,
		RecipientFingerprint: recipientFingerprint,
		CreatedAtUnix:        now.Unix(),
	})
	return payload
}

// ParseEnvelopeStored valida y decodifica un evento del bus.
func ParseEnvelopeStored(data []byte) (EnvelopeStored, error) {
	var event EnvelopeStored
	if err := json.Unmarshal(data, &event); err != nil {
		return EnvelopeStored{}, fmt.Errorf("events: json invalido: %w", err)
	}
	if event.Version != VersionEvents {
		return EnvelopeStored{}, fmt.Errorf("events: version no soportada: %d", event.Version)
	}
	if event.Type != TypeMailboxEnvelopeStored {
		return EnvelopeStored{}, fmt.Errorf("events: tipo desconocido: %s", event.Type)
	}
	if event.EnvelopeID == "" || !validate.Fingerprint(event.RecipientFingerprint) {
		return EnvelopeStored{}, fmt.Errorf("events: evento invalido")
	}
	return event, nil
}
