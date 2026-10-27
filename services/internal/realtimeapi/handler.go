// Package realtimeapi implementa el fan-out de senales WS sin contenido
// (ADR-005: la senal no lleva metadatos; el cliente consulta el buzon).
package realtimeapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/validate"
)

const (
	authDeadline      = 30 * time.Second
	readDeadline      = 90 * time.Second
	pingEvery         = 30 * time.Second
	writeDeadline     = 10 * time.Second
	signalPayload     = `{"type":"mailbox.envelope.stored"}`
	readyPayload      = `{"type":"ready"}`
	closeUnauthorized = 4401
)

// Config parametriza el servicio.
type Config struct {
	AuthURL      string
	DirectoryURL string
	HTTP         *http.Client
}

type authResponse struct {
	UserID string `json:"user_id"`
}

type directoryResponse struct {
	Fingerprint string `json:"fingerprint"`
}

// Handler sirve el WS de senales y expone el fan-out.
type Handler struct {
	cfg Config
	mux *http.ServeMux
	reg *Registry
}

// New construye el handler de realtime.
func New(cfg Config) *Handler {
	if cfg.HTTP == nil {
		cfg.HTTP = http.DefaultClient
	}
	h := &Handler{cfg: cfg, mux: http.NewServeMux(), reg: NewRegistry()}
	h.mux.HandleFunc("GET /internal/signal", h.handleSignal)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// Signal envia la senal a las conexiones del fingerprint; lo invoca el
// consumidor de eventos del bus.
func (h *Handler) Signal(fingerprint string) error {
	return h.reg.Signal(fingerprint)
}

// Registry agrupa las conexiones activas por fingerprint.
type Registry struct {
	mu    sync.RWMutex
	conns map[string]map[*clientConn]struct{}
}

// clientConn envuelve la escritura: gorilla exige un unico escritor.
type clientConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// NewRegistry crea el registro de conexiones.
func NewRegistry() *Registry {
	return &Registry{conns: map[string]map[*clientConn]struct{}{}}
}

func (r *Registry) add(fingerprint string, c *clientConn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	set, ok := r.conns[fingerprint]
	if !ok {
		set = map[*clientConn]struct{}{}
		r.conns[fingerprint] = set
	}
	set[c] = struct{}{}
}

func (r *Registry) remove(fingerprint string, c *clientConn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	set, ok := r.conns[fingerprint]
	if !ok {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(r.conns, fingerprint)
	}
}

// Signal envia la senal sin contenido a las conexiones del fingerprint.
func (r *Registry) Signal(fingerprint string) error {
	if !validate.Fingerprint(fingerprint) {
		return errors.New("realtime: fingerprint invalido")
	}
	r.mu.RLock()
	set := r.conns[fingerprint]
	targets := make([]*clientConn, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	r.mu.RUnlock()
	var firstErr error
	for _, c := range targets {
		c.mu.Lock()
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(signalPayload))
		c.mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// El origen se valida en el borde publico (gateway); este servicio
	// es interno y confia en la malla (mTLS en Fase 5).
	CheckOrigin: func(*http.Request) bool { return true },
}

func (h *Handler) handleSignal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &clientConn{conn: conn}

	fingerprint, err := h.authenticate(r.Context(), conn)
	if err != nil {
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeUnauthorized, ""), time.Now().Add(writeDeadline))
		_ = conn.Close()
		return
	}

	h.reg.add(fingerprint, c)
	defer h.reg.remove(fingerprint, c)
	defer func() { _ = conn.Close() }()

	c.mu.Lock()
	_ = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	err = conn.WriteMessage(websocket.TextMessage, []byte(readyPayload))
	c.mu.Unlock()
	if err != nil {
		return
	}

	stop := make(chan struct{})
	go pingLoop(conn, stop)
	defer close(stop)

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(readDeadline))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(readDeadline))
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		// Las tramas de datos del cliente se ignoran: el canal solo senaliza.
	}
}

func pingLoop(conn *websocket.Conn, stop chan struct{}) {
	ticker := time.NewTicker(pingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeDeadline)); err != nil {
				return
			}
		}
	}
}

// authenticate valida la sesion opaca contra auth y resuelve el
// fingerprint del usuario contra directory.
func (h *Handler) authenticate(ctx context.Context, conn *websocket.Conn) (string, error) {
	_ = conn.SetReadDeadline(time.Now().Add(authDeadline))
	messageType, raw, err := conn.ReadMessage()
	if err != nil || messageType != websocket.TextMessage {
		return "", errors.New("realtime: mensaje de autenticacion invalido")
	}
	var auth struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(raw, &auth); err != nil || auth.Type != "auth" || auth.Token == "" {
		return "", errors.New("realtime: autenticacion malformada")
	}
	var authResp authResponse
	if err := h.fetchJSON(ctx, h.cfg.AuthURL+"/internal/sessions/"+session.Hash(auth.Token), &authResp); err != nil {
		return "", err
	}
	if authResp.UserID == "" {
		return "", errors.New("realtime: sesion sin usuario")
	}
	var dirResp directoryResponse
	if err := h.fetchJSON(ctx, h.cfg.DirectoryURL+"/internal/keys/by-user/"+authResp.UserID, &dirResp); err != nil {
		return "", err
	}
	if !validate.Fingerprint(dirResp.Fingerprint) {
		return "", errors.New("realtime: fingerprint del usuario invalido")
	}
	return dirResp.Fingerprint, nil
}

func (h *Handler) fetchJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := h.cfg.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return errors.New("realtime: estado inesperado del servicio interno")
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
