package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/session"
)

const validToken = "token-valido"

var validHash = session.Hash(validToken)

type fakeLimiter struct {
	mu     sync.Mutex
	counts map[string]int64
}

func (f *fakeLimiter) Incr(key string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counts == nil {
		f.counts = map[string]int64{}
	}
	f.counts[key]++
	return f.counts[key], nil
}

func (f *fakeLimiter) Expire(key string, ttl time.Duration) error {
	return nil
}

func newServices(t *testing.T) (authURL, directoryURL, mailboxURL string) {
	t.Helper()
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/"+validHash) {
			_, _ = w.Write([]byte(`{"user_id":"11111111-1111-1111-1111-111111111111","expires_at":"2099-01-01T00:00:00Z"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(auth.Close)

	directory := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Bina-User") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(directory.Close)

	mailbox := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"id-sobre-1"}`))
	}))
	t.Cleanup(mailbox.Close)

	return auth.URL, directory.URL, mailbox.URL
}

func newAPI(authURL, directoryURL, mailboxURL string) *API {
	return New(Config{AuthURL: authURL, DirectoryURL: directoryURL, MailboxURL: mailboxURL})
}

func TestSobreSinSesion(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader("{}"))
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("estado = %d, se esperaba 401", rec.Code)
	}
}

func TestSobreConSesionValida(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	body := `{"recipient_fingerprint":"` + strings.Repeat("a", 64) + `","nonce":"AA==","ciphertext":"BB==","wraps":[{"recipient_key_id":"x"}],"blake3_fingerprint":"` + strings.Repeat("a", 64) + `"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
}

func TestSobreConSesionInvalida(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer token-falso")
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("estado = %d, se esperaba 401", rec.Code)
	}
}

func TestPublicarClavePropagaUsuario(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/keys", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
}

func TestBuscarClaveConFingerprintInvalido(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/keys/xyz", nil)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestListarSobresValidaFingerprint(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/envelopes", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestLimiteDePeticionesPorUsuario(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := New(Config{
		AuthURL:      authURL,
		DirectoryURL: directoryURL,
		MailboxURL:   mailboxURL,
		Cache:        &fakeLimiter{},
		RateLimit:    2,
	})
	body := `{"recipient_fingerprint":"` + strings.Repeat("a", 64) + `","nonce":"AA==","ciphertext":"BB==","wraps":[{"recipient_key_id":"x"}],"blake3_fingerprint":"` + strings.Repeat("a", 64) + `"}`
	for i := 1; i <= 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+validToken)
		api.ServeHTTP(rec, req)
		if i <= 2 && rec.Code != http.StatusOK {
			t.Fatalf("peticion %d: estado = %d cuerpo=%s", i, rec.Code, rec.Body.String())
		}
		if i == 3 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("peticion %d: estado = %d, se esperaba 429", i, rec.Code)
		}
	}
}
