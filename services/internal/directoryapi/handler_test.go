package directoryapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

var key32 = base64.StdEncoding.EncodeToString(make([]byte, 32))
var fp = strings.Repeat("a", 64)

type fakeCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeCache() *fakeCache { return &fakeCache{data: map[string]string{}} }

func (f *fakeCache) Get(key string) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[key]
	return v, ok, nil
}

func (f *fakeCache) SetEx(key, value string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = value
	return nil
}

func (f *fakeCache) Del(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	return nil
}

func newHandler(t *testing.T, respondGet string) *Handler {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(respondGet))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return New(Config{Supabase: supabase.New(srv.URL, "clave")})
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

func TestPublicarClaveValida(t *testing.T) {
	h := newHandler(t, "[]")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/keys", bytes.NewReader(publishBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
}

func TestPublicarClaveInvalida(t *testing.T) {
	h := newHandler(t, "[]")
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
	keyJSON, _ := json.Marshal(key32)
	response := `[{"user_id":"u1","x25519_public_key":` + string(keyJSON) + `,"ed25519_signing_public_key":` + string(keyJSON) + `,"fingerprint":"` + fp + `","active":true}]`
	h := newHandler(t, response)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "u1") {
		t.Fatal("la respuesta no contenia el user_id")
	}
}

func TestBuscarClaveInexistente(t *testing.T) {
	h := newHandler(t, "[]")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("estado = %d, se esperaba 404", rec.Code)
	}
}

func TestLookupCachaClave(t *testing.T) {
	getCalls := 0
	keyJSON, _ := json.Marshal(key32)
	response := `[{"user_id":"u1","x25519_public_key":` + string(keyJSON) + `,"ed25519_signing_public_key":` + string(keyJSON) + `,"fingerprint":"` + fp + `","active":true}]`
	mux := http.NewServeMux()
	mux.HandleFunc("GET /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		getCalls++
		_, _ = w.Write([]byte(response))
	})
	mux.HandleFunc("POST /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	h := New(Config{Supabase: supabase.New(srv.URL, "clave"), Cache: newFakeCache()})
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("llamada %d: estado = %d cuerpo=%s", i+1, rec.Code, rec.Body.String())
		}
	}
	if getCalls != 1 {
		t.Fatalf("consultas a PostgREST = %d, se esperaba 1 (cache activo)", getCalls)
	}
}

func TestPublicarInvalidaCache(t *testing.T) {
	keyJSON, _ := json.Marshal(key32)
	response := `[{"user_id":"11111111-1111-1111-1111-111111111111","x25519_public_key":` + string(keyJSON) + `,"ed25519_signing_public_key":` + string(keyJSON) + `,"fingerprint":"` + fp + `","active":true}]`
	mux := http.NewServeMux()
	mux.HandleFunc("GET /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(response))
	})
	mux.HandleFunc("POST /rest/v1/key_directory", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := newFakeCache()
	h := New(Config{Supabase: supabase.New(srv.URL, "clave"), Cache: c})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/keys/"+fp, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/keys", bytes.NewReader(publishBody()))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	c.mu.Lock()
	_, cached := c.data["fp:"+fp]
	c.mu.Unlock()
	if cached {
		t.Fatal("la publicacion debio invalidar el cache del fingerprint")
	}
}
