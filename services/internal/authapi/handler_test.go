package authapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRevocarSesion(t *testing.T) {
	h := newHandler(newStub(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/internal/sessions/"+strings.Repeat("a", 64), nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("estado = %d, se esperaba 204", rec.Code)
	}
}
