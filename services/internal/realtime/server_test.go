package realtime

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jessbludev/bina-mail/services/internal/session"
)

const validToken = "token-realtime"

const userID = "11111111-1111-1111-1111-111111111111"

var (
	validHash       = session.Hash(validToken)
	testFingerprint = strings.Repeat("a", 64)
)

func newAuthStub(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /internal/sessions/", func(w http.ResponseWriter, r *http.Request) {
		hash := strings.TrimPrefix(r.URL.Path, "/internal/sessions/")
		if hash != validHash {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"user_id":"` + userID + `","expires_at":"2099-01-01T00:00:00Z"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newDirectoryStub(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /internal/keys/", func(w http.ResponseWriter, r *http.Request) {
		fingerprint := strings.TrimPrefix(r.URL.Path, "/internal/keys/")
		if fingerprint != testFingerprint {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"user_id":"` + userID + `","active":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newRealtime(t *testing.T) *httptest.Server {
	t.Helper()
	auth := newAuthStub(t)
	directory := newDirectoryStub(t)
	h := New(Config{AuthURL: auth.URL, DirectoryURL: directory.URL})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// dialWS abre una conexion WS de cliente contra el servidor de prueba.
func dialWS(t *testing.T, serverURL string) *wsConn {
	t.Helper()
	host := strings.TrimPrefix(serverURL, "http://")
	conn, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	request := "GET /ws HTTP/1.1\r\nHost: " + host +
		"\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: " + key +
		"\r\nSec-WebSocket-Version: 13\r\n\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		t.Fatalf("write: %v", err)
	}
	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	if err != nil || !strings.Contains(line, "101") {
		t.Fatalf("handshake fallo: %q %v", line, err)
	}
	for {
		headerLine, err := br.ReadString('\n')
		if err != nil {
			t.Fatalf("headers: %v", err)
		}
		if strings.TrimRight(headerLine, "\r\n") == "" {
			break
		}
	}
	return &wsConn{Conn: conn, br: br}
}

func sendText(t *testing.T, ws *wsConn, payload string) {
	t.Helper()
	if err := ws.writeFrame(opText, []byte(payload), true); err != nil {
		t.Fatalf("send: %v", err)
	}
}

func readText(t *testing.T, ws *wsConn) map[string]string {
	t.Helper()
	op, payload, err := ws.readMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if op != opText {
		t.Fatalf("opcode = %d", op)
	}
	var m map[string]string
	if err := json.Unmarshal(payload, &m); err != nil {
		t.Fatalf("json: %v", err)
	}
	return m
}

func TestSenalExtremoAExtremo(t *testing.T) {
	srv := newRealtime(t)
	ws := dialWS(t, srv.URL)
	defer func() { _ = ws.Close() }()

	sendText(t, ws, `{"type":"auth","token":"`+validToken+`"}`)
	if m := readText(t, ws); m["type"] != "auth_ok" {
		t.Fatalf("respuesta auth = %v", m)
	}
	sendText(t, ws, `{"type":"bind","fingerprint":"`+testFingerprint+`"}`)
	if m := readText(t, ws); m["type"] != "bind_ok" {
		t.Fatalf("respuesta bind = %v", m)
	}
	sendText(t, ws, `{"type":"ping"}`)
	if m := readText(t, ws); m["type"] != "pong" {
		t.Fatalf("respuesta ping = %v", m)
	}

	body := `{"event":"envelope.stored","id":"env-1","recipient_fingerprint":"` + testFingerprint + `"}`
	resp, err := http.Post(srv.URL+"/internal/events", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post evento: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("estado evento = %d", resp.StatusCode)
	}

	m := readText(t, ws)
	if m["type"] != "signal" || m["event"] != "envelope.stored" || m["id"] != "env-1" {
		t.Fatalf("senal = %v", m)
	}
}

func TestAuthInvalidaRechazaConexion(t *testing.T) {
	srv := newRealtime(t)
	ws := dialWS(t, srv.URL)
	defer func() { _ = ws.Close() }()

	sendText(t, ws, `{"type":"auth","token":"token-falso"}`)
	if m := readText(t, ws); m["type"] != "error" || m["code"] != "sesion_invalida" {
		t.Fatalf("respuesta = %v", m)
	}
	if _, _, err := ws.readMessage(); err == nil {
		t.Fatal("se esperaba cierre de la conexion")
	}
}

func TestBindFingerprintAjenoRechazado(t *testing.T) {
	srv := newRealtime(t)
	ws := dialWS(t, srv.URL)
	defer func() { _ = ws.Close() }()

	sendText(t, ws, `{"type":"auth","token":"`+validToken+`"}`)
	if m := readText(t, ws); m["type"] != "auth_ok" {
		t.Fatalf("respuesta auth = %v", m)
	}
	sendText(t, ws, `{"type":"bind","fingerprint":"`+strings.Repeat("b", 64)+`"}`)
	if m := readText(t, ws); m["type"] != "error" || m["code"] != "fingerprint_invalido" {
		t.Fatalf("respuesta bind = %v", m)
	}
}

func TestEventoInvalidoRechazado(t *testing.T) {
	srv := newRealtime(t)
	resp, err := http.Post(srv.URL+"/internal/events", "application/json", strings.NewReader(`{"event":"otro"}`))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("estado = %d, se esperaba 400", resp.StatusCode)
	}
}

func TestSenalAFingerprintSinConexionesNoFalla(t *testing.T) {
	srv := newRealtime(t)
	body := `{"event":"envelope.stored","id":"env-2","recipient_fingerprint":"` + strings.Repeat("c", 64) + `"}`
	resp, err := http.Post(srv.URL+"/internal/events", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("estado = %d", resp.StatusCode)
	}
}
