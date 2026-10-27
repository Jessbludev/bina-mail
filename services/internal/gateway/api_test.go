package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/jessbludev/bina-mail/services/internal/session"
)

const (
	validToken     = "token-valido"
	allowedOrigin  = "https://app.bina.example"
	wsReadyPayload = `{"type":"ready"}`
)

var validHash = session.Hash(validToken)

type stubLimiter struct {
	mu     sync.Mutex
	counts map[string]int
	limit  int
}

func (s *stubLimiter) Allow(_ context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.counts == nil {
		s.counts = map[string]int{}
	}
	s.counts[key]++
	return s.counts[key] <= s.limit, nil
}

type errorLimiter struct{}

func (errorLimiter) Allow(context.Context, string) (bool, error) {
	return false, errors.New("contador no disponible")
}

func newServices(t *testing.T) (authURL, directoryURL, mailboxURL string) {
	t.Helper()
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/"+validHash) {
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
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

func envelopeBody() string {
	return `{"recipient_fingerprint":"` + strings.Repeat("a", 64) + `","nonce":"AA==","ciphertext":"BB==","wraps":[{"recipient_key_id":"x"}],"blake3_fingerprint":"` + strings.Repeat("a", 64) + `"}`
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
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader(envelopeBody()))
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

func TestCerrarSesionRevocaEnAuth(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("estado = %d cuerpo=%s, se esperaba 204", rec.Code, rec.Body.String())
	}
}

func TestCerrarSesionSinSesion(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/session", nil)
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
		Limiter:      &stubLimiter{limit: 2},
	})
	for i := 1; i <= 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader(envelopeBody()))
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

func TestLimiteFallaAbierto(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := New(Config{
		AuthURL:      authURL,
		DirectoryURL: directoryURL,
		MailboxURL:   mailboxURL,
		Limiter:      errorLimiter{},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/envelopes", strings.NewReader(envelopeBody()))
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s, se esperaba 200 (fail-open)", rec.Code, rec.Body.String())
	}
}

func TestPreflightConOrigenPermitido(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := New(Config{
		AuthURL:       authURL,
		DirectoryURL:  directoryURL,
		MailboxURL:    mailboxURL,
		AllowedOrigin: allowedOrigin,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/envelopes", nil)
	req.Header.Set("Origin", allowedOrigin)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("estado = %d, se esperaba 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != allowedOrigin {
		t.Fatal("falta la cabecera de origen permitido")
	}
}

func TestOrigenAjenoRechazado(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := New(Config{
		AuthURL:       authURL,
		DirectoryURL:  directoryURL,
		MailboxURL:    mailboxURL,
		AllowedOrigin: allowedOrigin,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/envelopes", nil)
	req.Header.Set("Origin", "https://ajeno.example")
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("estado = %d, se esperaba 403", rec.Code)
	}
}

func TestCabecerasDeSeguridadPresentes(t *testing.T) {
	authURL, directoryURL, mailboxURL := newServices(t)
	api := newAPI(authURL, directoryURL, mailboxURL)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/envelopes", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	api.ServeHTTP(rec, req)
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("falta X-Content-Type-Options")
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("falta Cache-Control no-store")
	}
}

var stubUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

func TestRealtimeProxyTunel(t *testing.T) {
	realtime := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := stubUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, auth, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("realtime: lectura del mensaje de autenticacion: %v", err)
			return
		}
		if !strings.Contains(string(auth), validToken) {
			t.Errorf("realtime: mensaje inesperado: %s", auth)
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(wsReadyPayload))
	}))
	t.Cleanup(realtime.Close)

	api := New(Config{RealtimeURL: realtime.URL, AllowedOrigin: allowedOrigin})
	gateway := httptest.NewServer(api)
	t.Cleanup(gateway.Close)

	header := http.Header{"Origin": {allowedOrigin}}
	conn, _, err := websocket.DefaultDialer.Dial(strings.Replace(gateway.URL+"/api/v1/realtime", "http", "ws", 1), header)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	authMessage := `{"type":"auth","token":"` + validToken + `"}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(authMessage)); err != nil {
		t.Fatalf("escritura: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, ready, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("lectura: %v", err)
	}
	if string(ready) != wsReadyPayload {
		t.Fatalf("respuesta inesperada: %s", ready)
	}
}
