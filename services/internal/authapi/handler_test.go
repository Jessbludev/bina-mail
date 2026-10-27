package authapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/jwt"
	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

func newStub(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		expires := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		_, _ = w.Write([]byte(`[{"token_hash":"h","user_id":"11111111-1111-1111-1111-111111111111","expires_at":"` + expires + `"}]`))
	})
	mux.HandleFunc("PATCH /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newHandler(stub *httptest.Server) *Handler {
	return New(Config{
		Supabase:   supabase.New(stub.URL, "clave"),
		JWTSecret:  "secreto",
		SessionTTL: time.Minute,
		Now:        func() time.Time { return time.Unix(1700000000, 0).UTC() },
	})
}

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

func TestEmitirSesionValida(t *testing.T) {
	h := newHandler(newStub(t))
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
	h := newHandler(newStub(t))
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
	h := newHandler(newStub(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/sessions/"+strings.Repeat("a", 64), nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "11111111-1111-1111-1111-111111111111") {
		t.Fatal("la respuesta no contenia el user_id")
	}
}

func TestLookupHashInvalido(t *testing.T) {
	h := newHandler(newStub(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/sessions/xyz", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", rec.Code)
	}
}

func TestLookupUsaCache(t *testing.T) {
	getCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		getCalls++
		expires := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		_, _ = w.Write([]byte(`[{"token_hash":"h","user_id":"11111111-1111-1111-1111-111111111111","expires_at":"` + expires + `"}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	h := New(Config{
		Supabase:   supabase.New(srv.URL, "clave"),
		JWTSecret:  "secreto",
		SessionTTL: time.Minute,
		Now:        func() time.Time { return time.Unix(1700000000, 0).UTC() },
		Cache:      newFakeCache(),
	})
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/internal/sessions/"+strings.Repeat("a", 64), nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("llamada %d: estado = %d cuerpo=%s", i+1, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "11111111-1111-1111-1111-111111111111") {
			t.Fatalf("llamada %d: la respuesta no contenia el user_id", i+1)
		}
	}
	if getCalls != 1 {
		t.Fatalf("consultas a PostgREST = %d, se esperaba 1 (cache activo)", getCalls)
	}
}

func TestRevocarBorraCache(t *testing.T) {
	getCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		getCalls++
		expires := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		_, _ = w.Write([]byte(`[{"token_hash":"h","user_id":"11111111-1111-1111-1111-111111111111","expires_at":"` + expires + `"}]`))
	})
	mux.HandleFunc("PATCH /rest/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	h := New(Config{
		Supabase:   supabase.New(srv.URL, "clave"),
		JWTSecret:  "secreto",
		SessionTTL: time.Minute,
		Now:        func() time.Time { return time.Unix(1700000000, 0).UTC() },
		Cache:      newFakeCache(),
	})
	lookup := func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/internal/sessions/"+strings.Repeat("a", 64), nil)
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("estado = %d cuerpo=%s", rec.Code, rec.Body.String())
		}
	}
	lookup()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/internal/sessions/"+strings.Repeat("a", 64), nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("estado = %d, se esperaba 204", rec.Code)
	}
	lookup()
	if getCalls != 2 {
		t.Fatalf("consultas a PostgREST = %d, se esperaba 2 (revocacion borro el cache)", getCalls)
	}
}
