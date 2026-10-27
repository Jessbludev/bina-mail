package events

import (
	"strings"
	"testing"
	"time"
)

func TestEventoIdaYVuelta(t *testing.T) {
	fingerprint := strings.Repeat("a", 64)
	payload := NewEnvelopeStored("sobre-1", fingerprint, time.Unix(1700000000, 0).UTC())
	event, err := ParseEnvelopeStored(payload)
	if err != nil {
		t.Fatalf("ParseEnvelopeStored fallo: %v", err)
	}
	if event.EnvelopeID != "sobre-1" || event.RecipientFingerprint != fingerprint {
		t.Fatalf("evento decodificado incorrecto: %+v", event)
	}
	if event.CreatedAtUnix != 1700000000 {
		t.Fatalf("created_at_unix = %d", event.CreatedAtUnix)
	}
}

func TestEventoConVersionIncorrecta(t *testing.T) {
	payload := []byte(`{"version":99,"type":"mailbox.envelope.stored","envelope_id":"s1","recipient_fingerprint":"` + strings.Repeat("a", 64) + `","created_at_unix":1}`)
	if _, err := ParseEnvelopeStored(payload); err == nil {
		t.Fatal("se esperaba rechazo por version")
	}
}

func TestEventoConFingerprintInvalido(t *testing.T) {
	payload := []byte(`{"version":1,"type":"mailbox.envelope.stored","envelope_id":"s1","recipient_fingerprint":"no-hex","created_at_unix":1}`)
	if _, err := ParseEnvelopeStored(payload); err == nil {
		t.Fatal("se esperaba rechazo por fingerprint")
	}
}
