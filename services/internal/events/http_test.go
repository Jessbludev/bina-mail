package events

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPEntregaSenal(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = string(body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	publisher := NewHTTPPublisher(srv.URL)
	err := publisher.PublishEnvelopeStored(context.Background(), EnvelopeStored{
		Type:                 TypeMailboxEnvelopeStored,
		EnvelopeID:           "env-1",
		RecipientFingerprint: strings.Repeat("a", 64),
		CreatedAtUnix:        time.Now().Unix(),
	})
	if err != nil {
		t.Fatalf("PublishEnvelopeStored: %v", err)
	}
	if !strings.Contains(got, "envelope.stored") || !strings.Contains(got, "env-1") {
		t.Fatalf("evento recibido = %q", got)
	}
}

func TestHTTPFallaConEstadoDeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	publisher := NewHTTPPublisher(srv.URL)
	err := publisher.PublishEnvelopeStored(context.Background(), EnvelopeStored{
		Type:                 TypeMailboxEnvelopeStored,
		EnvelopeID:           "env-1",
		RecipientFingerprint: strings.Repeat("a", 64),
		CreatedAtUnix:        time.Now().Unix(),
	})
	if err == nil {
		t.Fatal("se esperaba error con estado 500")
	}
}
