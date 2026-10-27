// Package gateway implementa el borde publico HTTP/WS de Bina Mail (BFF, ADR-001).
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/validate"
)

// Config parametriza la API.
type Config struct {
	AuthURL      string
	DirectoryURL string
	MailboxURL   string
	HTTP         *http.Client
}

type ctxKey int

const userIDKey ctxKey = 0

// API expone los endpoints publicos del BFF.
type API struct {
	cfg Config
	mux *http.ServeMux
}

// New construye la API del gateway.
func New(cfg Config) *API {
	if cfg.HTTP == nil {
		cfg.HTTP = http.DefaultClient
	}
	a := &API{cfg: cfg, mux: http.NewServeMux()}
	a.mux.HandleFunc("POST /api/v1/auth/session", a.handleSessionIssue)
	a.mux.HandleFunc("POST /api/v1/keys", a.requireSession(a.handlePublishKey))
	a.mux.HandleFunc("GET /api/v1/keys/{fingerprint}", a.handleLookupKey)
	a.mux.HandleFunc("POST /api/v1/envelopes", a.requireSession(a.handleStoreEnvelope))
	a.mux.HandleFunc("GET /api/v1/envelopes", a.requireSession(a.handleListEnvelopes))
	return a
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

// UserID devuelve el usuario autenticado por la sesion opaca.
func UserID(r *http.Request) string {
	userID, _ := r.Context().Value(userIDKey).(string)
	return userID
}

func (a *API) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			unauthorized(w)
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, a.cfg.AuthURL+"/internal/sessions/"+session.Hash(token), nil)
		if err != nil {
			unauthorized(w)
			return
		}
		resp, err := a.cfg.HTTP.Do(req)
		if err != nil {
			unauthorized(w)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			unauthorized(w)
			return
		}
		var sess struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil || sess.UserID == "" {
			unauthorized(w)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userIDKey, sess.UserID)))
	}
}

func (a *API) handleSessionIssue(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r, 8*1024)
	if !ok {
		return
	}
	a.proxy(w, r, http.MethodPost, a.cfg.AuthURL+"/internal/sessions", body, nil)
}

func (a *API) handlePublishKey(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r, 16*1024)
	if !ok {
		return
	}
	headers := map[string]string{"X-Bina-User": UserID(r)}
	a.proxy(w, r, http.MethodPost, a.cfg.DirectoryURL+"/internal/keys", body, headers)
}

func (a *API) handleLookupKey(w http.ResponseWriter, r *http.Request) {
	fingerprint := r.PathValue("fingerprint")
	if !validate.Fingerprint(fingerprint) {
		writeError(w, http.StatusBadRequest, "fingerprint_invalido")
		return
	}
	a.proxy(w, r, http.MethodGet, a.cfg.DirectoryURL+"/internal/keys/"+fingerprint, nil, nil)
}

func (a *API) handleStoreEnvelope(w http.ResponseWriter, r *http.Request) {
	body, ok := readBody(w, r, 2<<20)
	if !ok {
		return
	}
	var env struct {
		RecipientFingerprint string            `json:"recipient_fingerprint"`
		Wraps                []json.RawMessage `json:"wraps"`
	}
	if err := json.Unmarshal(body, &env); err != nil || !validate.Fingerprint(env.RecipientFingerprint) || len(env.Wraps) == 0 {
		writeError(w, http.StatusBadRequest, "sobre_invalido")
		return
	}
	a.proxy(w, r, http.MethodPost, a.cfg.MailboxURL+"/internal/envelopes", body, nil)
}

func (a *API) handleListEnvelopes(w http.ResponseWriter, r *http.Request) {
	recipient := r.URL.Query().Get("recipient_fingerprint")
	if !validate.Fingerprint(recipient) {
		writeError(w, http.StatusBadRequest, "fingerprint_invalido")
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "limite_invalido")
			return
		}
		limit = parsed
	}
	q := url.Values{}
	q.Set("recipient_fingerprint", recipient)
	q.Set("limit", strconv.Itoa(limit))
	a.proxy(w, r, http.MethodGet, a.cfg.MailboxURL+"/internal/envelopes?"+q.Encode(), nil, nil)
}

func (a *API) proxy(w http.ResponseWriter, r *http.Request, method, target string, body []byte, headers map[string]string) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(r.Context(), method, target, reader)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "interno")
		return
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := a.cfg.HTTP.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "servicio_no_disponible")
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func readBody(w http.ResponseWriter, r *http.Request, limit int64) ([]byte, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if err != nil {
		writeError(w, http.StatusBadRequest, "payload_invalido")
		return nil, false
	}
	return body, true
}

func unauthorized(w http.ResponseWriter) {
	writeError(w, http.StatusUnauthorized, "sesion_invalida")
}

func writeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
