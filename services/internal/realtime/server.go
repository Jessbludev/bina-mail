// Package realtime emite senales WS sin contenido ni metadatos (ADR-005/ADR-008).
package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/events"
	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/validate"
)

// Config parametriza el servicio. Las dependencias HTTP se inyectan.
type Config struct {
	AuthURL      string
	DirectoryURL string
	HTTP         *http.Client
}

// Handler expone GET /ws (cliente) y POST /internal/events (mailbox).
type Handler struct {
	cfg Config
	reg *Registry
	mux *http.ServeMux
}

// New construye el handler del servicio.
func New(cfg Config) *Handler {
	if cfg.HTTP == nil {
		cfg.HTTP = http.DefaultClient
	}
	h := &Handler{cfg: cfg, reg: newRegistry(), mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /ws", h.handleWS)
	h.mux.HandleFunc("POST /internal/events", h.handleEvent)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleEvent recibe senales de mailbox (entrega interna, ADR-008).
// La senal es un aviso sin contenido: reenviarla es idempotente.
func (h *Handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	var evt events.EnvelopeStored
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4*1024)).Decode(&evt); err != nil ||
		evt.Event != events.EventEnvelopeStored || evt.ID == "" ||
		!validate.Fingerprint(evt.RecipientFingerprint) {
		writeError(w, http.StatusBadRequest, "evento_invalido")
		return
	}
	payload, err := json.Marshal(map[string]string{
		"type":  "signal",
		"event": evt.Event,
		"id":    evt.ID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "interno")
		return
	}
	h.reg.signal(evt.RecipientFingerprint, payload)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade(w, r)
	if err != nil {
		return
	}
	c := &client{ws: ws}
	defer func() {
		_ = ws.Close()
		if c.fingerprint != "" {
			h.reg.remove(c.fingerprint, c)
		}
	}()
	userID, err := h.authenticate(ws)
	if err != nil {
		writeErrorFrame(ws, "sesion_invalida")
		return
	}
	fingerprint, err := h.bind(ws, userID)
	if err != nil {
		return
	}
	c.fingerprint = fingerprint
	h.reg.add(fingerprint, c)
	for {
		op, payload, err := ws.readMessage()
		if err != nil {
			return
		}
		if op != opText {
			writeErrorFrame(ws, "trama_invalida")
			return
		}
		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(payload, &msg); err != nil || msg.Type != "ping" {
			writeErrorFrame(ws, "mensaje_invalido")
			return
		}
		if err := writeJSONFrame(ws, map[string]string{"type": "pong"}); err != nil {
			return
		}
	}
}

// authenticate espera {"type":"auth","token":...} y valida contra auth.
// Ante cualquier fallo se rechaza la conexion (fail-closed, zero-trust).
func (h *Handler) authenticate(ws *wsConn) (string, error) {
	op, payload, err := ws.readMessage()
	if err != nil {
		return "", err
	}
	if op != opText {
		return "", errProtocol
	}
	var msg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(payload, &msg); err != nil || msg.Type != "auth" || msg.Token == "" {
		return "", errProtocol
	}
	body, err := getJSON(h.cfg.HTTP, h.cfg.AuthURL+"/internal/sessions/"+session.Hash(msg.Token))
	if err != nil {
		return "", err
	}
	var sess struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(body, &sess); err != nil || sess.UserID == "" {
		return "", errProtocol
	}
	if err := writeJSONFrame(ws, map[string]string{"type": "auth_ok"}); err != nil {
		return "", err
	}
	return sess.UserID, nil
}

// bind espera {"type":"bind","fingerprint":...} y verifica que la clave
// pertenezca al usuario autenticado y siga activa.
func (h *Handler) bind(ws *wsConn, userID string) (string, error) {
	op, payload, err := ws.readMessage()
	if err != nil {
		return "", err
	}
	if op != opText {
		writeErrorFrame(ws, "fingerprint_invalido")
		return "", errProtocol
	}
	var msg struct {
		Type        string `json:"type"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.Unmarshal(payload, &msg); err != nil || msg.Type != "bind" ||
		!validate.Fingerprint(msg.Fingerprint) {
		writeErrorFrame(ws, "fingerprint_invalido")
		return "", errProtocol
	}
	body, err := getJSON(h.cfg.HTTP, h.cfg.DirectoryURL+"/internal/keys/"+msg.Fingerprint)
	if err != nil {
		if statusOf(err) == http.StatusNotFound {
			writeErrorFrame(ws, "fingerprint_invalido")
		} else {
			writeErrorFrame(ws, "directorio_no_disponible")
		}
		return "", err
	}
	var key struct {
		UserID string `json:"user_id"`
		Active bool   `json:"active"`
	}
	if err := json.Unmarshal(body, &key); err != nil || !key.Active || key.UserID != userID {
		writeErrorFrame(ws, "fingerprint_invalido")
		return "", errProtocol
	}
	if err := writeJSONFrame(ws, map[string]string{"type": "bind_ok"}); err != nil {
		return "", err
	}
	return msg.Fingerprint, nil
}

func getJSON(client *http.Client, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errStatus{code: resp.StatusCode}
	}
	return io.ReadAll(io.LimitReader(resp.Body, 16*1024))
}

// errStatus transporta el codigo HTTP de una dependencia fallida.
type errStatus struct{ code int }

func (e errStatus) Error() string {
	return "realtime: estado http " + strconv.Itoa(e.code)
}

func statusOf(err error) int {
	var status errStatus
	if errors.As(err, &status) {
		return status.code
	}
	return 0
}

func writeJSONFrame(ws *wsConn, v any) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return ws.writeMessage(opText, payload)
}

func writeErrorFrame(ws *wsConn, code string) {
	_ = writeJSONFrame(ws, map[string]string{"type": "error", "code": code})
}

func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
