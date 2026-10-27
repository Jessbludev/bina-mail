package directoryapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/resp"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

var key32 = base64.StdEncoding.EncodeToString(make([]byte, 32))
var fp = strings.Repeat("a", 64)

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

func (m *memCache) len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.data)
}

type stub struct {
	getCalls int32
	response string
}

func newHandler(t *testing.T, s *stub, cache resp.Cache) *Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&s.getCalls, 1)
		_, _ = w.Write([]byte(s.response))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return New(Config{Supabase: supabase.New(srv.URL, "clave"), Cache: cache})
}

func publishBody() []byte {
	body, _ := json.Marshal(map[string]string{
		"user_id":                    "11111111-1111-1111-1111-111111111111",
		"x25519_public_key":          key32,
		"ed25519_signing_public_key": key32,
		"fingerprint":                fp,
	})
	return body
}

func keyRowJSON() string {
	keyJSON, _ := json.Marshal(key32)
	return `[{"user_id":"11111111-1111-1111-1111-111111111111","x25519_public_key":` + string(keyJSON) + `,"ed25519_signing_public_key":` + string(keyJSON) + `,"fingerprint":"` + fp + `","active":true}]`
}

func TestPublicarClaveValida(t *testing.T) {
	h := newHandler(t, &stub{response: "[]"}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/keys", bytes.NewReader(publishBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
}

func TestPublicarClaveInvalida(t *testing.T) {
	h := newHandler(t, &stub{response: "[]"}, nil)
	body, _ := json.Marshal(map[string]string{
		"user_id":           "u",
		"x25519_public_key": "no-es-base64-de-32",
		"fingerprint":       fp,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/keys", bytes.NewReader(body))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestBuscarClaveExistente(t *testing.T) {
	h := newHandler(t, &stub{response: keyRowJSON()}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "11111111-1111-1111-1111-111111111111") {
		t.Fatal("la respuesta no contenia el user_id")
	}
}

func TestBuscarClaveInexistente(t *testing.T) {
	h := newHandler(t, &stub{response: "[]"}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("estado = %d, se esperaba 404", rec.Code)
	}
}

func TestBuscarPorUsuario(t *testing.T) {
	h := newHandler(t, &stub{response: keyRowJSON()}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/by-user/11111111-1111-1111-1111-111111111111", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), fp) {
		t.Fatal("la respuesta no contenia el fingerprint")
	}
}

func TestBuscarPorUsuarioInvalido(t *testing.T) {
	h := newHandler(t, &stub{response: keyRowJSON()}, nil)
	body := strings.Repeat("x", 200)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/by-user/"+body, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestLaCacheEvitaLaSegundaConsulta(t *testing.T) {
	s := &stub{response: keyRowJSON()}
	h := newHandler(t, s, newMemCache())
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("estado = %d", rec.Code)
		}
	}
	if atomic.LoadInt32(&s.getCalls) != 1 {
		t.Fatalf("llamadas a la base de datos = %d, se esperaba 1", s.getCalls)
	}
}

func TestPublicarInvalidaLaCache(t *testing.T) {
	s := &stub{response: keyRowJSON()}
	cache := newMemCache()
	h := newHandler(t, s, cache)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/by-user/11111111-1111-1111-1111-111111111111", nil)
	h.ServeHTTP(rec, req)
	if cache.len() == 0 {
		t.Fatal("la cache debia poblarse tras la consulta")
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/keys", bytes.NewReader(publishBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d", rec.Code)
	}
	if cache.len() != 0 {
		t.Fatal("la cache debia quedar vacia tras publicar")
	}
}
