package authapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/jwt"
	"github.com/jessbludev/bina-mail/services/internal/resp"
	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

type memCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newMemCache() *memCache {
	return &memCache{data: map[string]string{}}
}

func (m *memCache) Get(_ context.Context, key string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.data[key]
	return value, ok, nil
}

func (m *memCache) SetEx(_ context.Context, key, value string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *memCache) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

type authStub struct {
	getCalls int32
	expires  string
}

func newStub(t *testing.T, s *authStub) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&s.getCalls, 1)
		_, _ = w.Write([]byte(`[{"token_hash":"h","user_id":"11111111-1111-1111-1111-111111111111","expires_at":"` + s.expires + `"}]`))
	})
	mux.HandleFunc("PATCH /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newHandler(srv *httptest.Server, cache resp.Cache) *Handler {
	return New(Config{
		Supabase:   supabase.New(srv.URL, "clave"),
		JWTSecret:  "secreto",
		SessionTTL: time.Minute,
		Now:        func() time.Time { return time.Unix(1700000000, 0).UTC() },
		Cache:      cache,
	})
}

func lookupRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/internal/sessions/"+strings.Repeat("a", 64), nil)
}

func TestEmitirSesionValida(t *testing.T) {
	srv := newStub(t, &authStub{expires: farExpiry()})
	h := newHandler(srv, nil)
	now := time.Unix(1700000000, 0).UTC()
	token, err := jwt.SignHS256("secreto", jwt.Claims{Subject: "11111111-1111-1111-1111-111111111111", Expiry: now.Add(time.Hour).Unix()})
	if err != nil {
		t.Fatalf("SignHS256: %v", err)
	}
	body, _ := json.Marshal(map[string]string{"supabase_jwt": token})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/sessions", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	var response struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&response)
	if response.Token == "" || len(session.Hash(response.Token)) != 64 {
		t.Fatal("token invalido en la respuesta")
	}
}

func TestEmitirSesionConJwtInvalido(t *testing.T) {
	srv := newStub(t, &authStub{expires: farExpiry()})
	h := newHandler(srv, nil)
	token, _ := jwt.SignHS256("secreto-equivocado", jwt.Claims{Subject: "u", Expiry: time.Now().Add(time.Hour).Unix()})
	body, _ := json.Marshal(map[string]string{"supabase_jwt": token})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/sessions", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("estado = %d, se esperaba 401", rec.Code)
	}
}

func TestLookupSesionActiva(t *testing.T) {
	srv := newStub(t, &authStub{expires: farExpiry()})
	h := newHandler(srv, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, lookupRequest())
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "11111111-1111-1111-1111-111111111111") {
		t.Fatal("la respuesta no contenia el user_id")
	}
}

func TestLookupHashInvalido(t *testing.T) {
	srv := newStub(t, &authStub{expires: farExpiry()})
	h := newHandler(srv, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/sessions/xyz", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestRevocarSesion(t *testing.T) {
	srv := newStub(t, &authStub{expires: farExpiry()})
	h := newHandler(srv, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/internal/sessions/"+strings.Repeat("a", 64), nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("estado = %d, se esperaba 204", rec.Code)
	}
}

func TestLaCacheAhorraConsultas(t *testing.T) {
	stub := &authStub{expires: farExpiry()}
	srv := newStub(t, stub)
	h := newHandler(srv, newMemCache())
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, lookupRequest())
		if rec.Code != http.StatusOK {
			t.Fatalf("estado = %d", rec.Code)
		}
	}
	if atomic.LoadInt32(&stub.getCalls) != 1 {
		t.Fatalf("llamadas a la base de datos = %d, se esperaba 1", stub.getCalls)
	}
}

func TestRevocarInvalidaLaCache(t *testing.T) {
	stub := &authStub{expires: farExpiry()}
	srv := newStub(t, stub)
	h := newHandler(srv, newMemCache())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, lookupRequest())
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/internal/sessions/"+strings.Repeat("a", 64), nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("estado = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, lookupRequest())
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d", rec.Code)
	}
	if atomic.LoadInt32(&stub.getCalls) != 2 {
		t.Fatalf("llamadas a la base de datos = %d, se esperaba 2 tras revocar", stub.getCalls)
	}
}

func farExpiry() string {
	return time.Unix(1900000000, 0).UTC().Format(time.RFC3339)
}
