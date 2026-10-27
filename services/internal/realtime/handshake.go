package realtime

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

var errHandshake = errors.New("realtime: handshake invalido")

// upgrade valida la peticion y toma el control de la conexion (RFC 6455).
func upgrade(w http.ResponseWriter, r *http.Request) (*wsConn, error) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") ||
		r.Header.Get("Sec-WebSocket-Key") == "" ||
		r.Header.Get("Sec-WebSocket-Version") != "13" {
		http.Error(w, "handshake invalido", http.StatusBadRequest)
		return nil, errHandshake
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack no soportado", http.StatusInternalServerError)
		return nil, errHandshake
	}
	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}
	sum := sha1.Sum([]byte(r.Header.Get("Sec-WebSocket-Key") + wsGUID))
	accept := base64.StdEncoding.EncodeToString(sum[:])
	head := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"
	if _, err := conn.Write([]byte(head)); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &wsConn{Conn: conn, br: buf.Reader, requireMask: true}, nil
}
