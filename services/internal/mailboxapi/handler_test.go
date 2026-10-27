package mailboxapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/jessbludev/bina-mail/services/internal/events"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

var (
	nonceB64   = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 24))
	cipherB64  = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{2}, 64))
	wrapKeyB64 = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{3}, 56))
	fpHex      = strings.Repeat("a", 64)
)

type fakePublisher struct {
	mu        sync.Mutex
	published [][]byte
}

func (f *fakePublisher) Publish(_ context.Context, _ []byte, value []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.published = append(f.published, value)
	return nil
}

func (f *fakePublisher) last() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.published) == 0 {
		return nil
	}
	return f.published[len(f.published)-1]
}

func newHandlerWithPublisher(t *testing.T, getResponse string, pub Publisher) *Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rest/v1/envelopes", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"id-sobre-1"}]`))
	})
	mux.HandleFunc("GET /rest/v1/envelopes", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(getResponse))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return New(Config{Supabase: supabase.New(srv.URL, "clave"), MaxEnvelopeBytes: 1024, Publisher: pub})
}

func newHandler(t *testing.T, getResponse string) *Handler {
	return newHandlerWithPublisher(t, getResponse, nil)
}

func storeBody() []byte {
	body, _ := json.Marshal(map[string]any{
		"recipient_fingerprint": fpHex,
		"nonce":                 nonceB64,
		"ciphertext":            cipherB64,
		"wraps": []map[string]string{{
			"recipient_key_id":     fpHex,
			"ephemeral_public_key": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32)),
			"wrapped_key":          wrapKeyB64,
		}},
		"blake3_fingerprint": fpHex,
	})
	return body
}

func TestAlmacenarSobreValido(t *testing.T) {
	h := newHandler(t, "[]")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/envelopes", bytes.NewReader(storeBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "id-sobre-1") {
		t.Fatal("la respuesta no contenia el id")
	}
}

func TestAlmacenarPublicaEvento(t *testing.T) {
	pub := &fakePublisher{}
	h := newHandlerWithPublisher(t, "[]", pub)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/envelopes", bytes.NewReader(storeBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	event, err := events.ParseEnvelopeStored(pub.last())
	if err != nil {
		t.Fatalf("evento invalido: %v", err)
	}
	if event.EnvelopeID != "id-sobre-1" {
		t.Fatalf("envelope_id = %q, se esperaba id-sobre-1", event.EnvelopeID)
	}
	if event.RecipientFingerprint != fpHex {
		t.Fatalf("recipient_fingerprint = %q", event.RecipientFingerprint)
	}
	if event.Type != events.TypeMailboxEnvelopeStored || event.Version != events.VersionEvents {
		t.Fatalf("evento inesperado: %+v", event)
	}
}

func TestAlmacenarSinPublicadorNoFallayGuarda(t *testing.T) {
	h := newHandlerWithPublisher(t, "[]", nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/envelopes", bytes.NewReader(storeBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d", rec.Code)
	}
}

func TestAlmacenarSobreConFingerprintInvalido(t *testing.T) {
	h := newHandler(t, "[]")
	body, _ := json.Marshal(map[string]any{
		"recipient_fingerprint": "no-hex",
		"nonce":                 nonceB64,
		"ciphertext":            cipherB64,
		"wraps":                 []map[string]string{{"recipient_key_id": fpHex}},
		"blake3_fingerprint":    fpHex,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/envelopes", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestAlmacenarSobreSinWraps(t *testing.T) {
	h := newHandler(t, "[]")
	body, _ := json.Marshal(map[string]any{
		"recipient_fingerprint": fpHex,
		"nonce":                 nonceB64,
		"ciphertext":            cipherB64,
		"wraps":                 []map[string]string{},
		"blake3_fingerprint":    fpHex,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/envelopes", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestListarSobres(t *testing.T) {
	nonceJSON, _ := json.Marshal(nonceB64)
	cipherJSON, _ := json.Marshal(cipherB64)
	response := `[{"id":"e1","recipient_fingerprint":"` + fpHex + `","nonce":` + string(nonceJSON) + `,"ciphertext":` + string(cipherJSON) + `,"wraps":[],"blake3_fingerprint":"` + fpHex + `","size_bytes":64,"created_at":"2026-10-23T00:00:00Z"}]`
	h := newHandler(t, response)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/envelopes?recipient_fingerprint="+fpHex+"&limit=10", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), nonceB64) {
		t.Fatal("la respuesta no contenia el nonce en base64")
	}
}

func TestListarSinFingerprint(t *testing.T) {
	h := newHandler(t, "[]")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/envelopes", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}
