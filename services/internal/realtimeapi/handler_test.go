package realtimeapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const fpBob = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func newBackend(t *testing.T, authStatus int) (authURL, dirURL string) {
	t.Helper()
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authStatus != http.StatusOK {
			w.WriteHeader(authStatus)
			return
		}
		_, _ = w.Write([]byte(`{"user_id":"11111111-1111-1111-1111-111111111111","expires_at":"2099-01-01T00:00:00Z"}`))
	}))
	t.Cleanup(auth.Close)
	dir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"fingerprint":"` + fpBob + `"}`))
	}))
	t.Cleanup(dir.Close)
	return auth.URL, dir.URL
}

func wsURL(server *httptest.Server) string {
	return "ws" + strings.TrimPrefix(server.URL, "http") + "/internal/signal"
}

func TestSenalSinContenidoAlDestinatario(t *testing.T) {
	authURL, dirURL := newBackend(t, http.StatusOK)
	handler := New(Config{AuthURL: authURL, DirectoryURL: dirURL})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server), nil)
	if err != nil {
		t.Fatalf("Dial fallo: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.WriteJSON(map[string]string{"type": "auth", "token": "token-valido"}); err != nil {
		t.Fatalf("auth fallo: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, ready, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ready fallo: %v", err)
	}
	if !strings.Contains(string(ready), "ready") {
		t.Fatalf("ready inesperado: %s", ready)
	}

	if err := handler.Signal(fpBob); err != nil {
		t.Fatalf("Signal fallo: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, signal, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("senal fallo: %v", err)
	}
	if strings.Contains(string(signal), fpBob) {
		t.Fatal("la senal filtraba el fingerprint")
	}
	if !strings.Contains(string(signal), "mailbox.envelope.stored") {
		t.Fatalf("senal inesperada: %s", signal)
	}
}

func TestAutenticacionInvalidaCierra4401(t *testing.T) {
	authURL, dirURL := newBackend(t, http.StatusUnauthorized)
	handler := New(Config{AuthURL: authURL, DirectoryURL: dirURL})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server), nil)
	if err != nil {
		t.Fatalf("Dial fallo: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.WriteJSON(map[string]string{"type": "auth", "token": "token-falso"}); err != nil {
		t.Fatalf("auth fallo: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		_, _, err := conn.ReadMessage()
		if err == nil {
			continue
		}
		if websocket.IsCloseError(err, closeUnauthorized) {
			return
		}
		t.Fatalf("se esperaba cierre 4401, error: %v", err)
	}
}

func TestSignalRechazaFingerprintInvalido(t *testing.T) {
	handler := New(Config{})
	if err := handler.Signal("xyz"); err == nil {
		t.Fatal("se esperaba rechazo del fingerprint invalido")
	}
}
