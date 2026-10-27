package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPPublisher entrega el evento por POST interno (ADR-007/ADR-008).
// La entrega es de mejor esfuerzo: la senal es un aviso, no una transaccion.
type HTTPPublisher struct {
	EventsURL string
	HTTP      *http.Client
}

// NewHTTPPublisher construye el publicador hacia la URL de eventos de realtime.
func NewHTTPPublisher(eventsURL string) *HTTPPublisher {
	return &HTTPPublisher{EventsURL: eventsURL, HTTP: http.DefaultClient}
}

// PublishEnvelopeStored envia la senal y exige 2xx del consumidor.
func (p *HTTPPublisher) PublishEnvelopeStored(ctx context.Context, e EnvelopeStored) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.EventsURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("events: entrega devolvio %d", resp.StatusCode)
	}
	return nil
}
