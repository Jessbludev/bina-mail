// Package authapi implementa el servicio de sesiones opacas de corta vida.
package authapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/jwt"
	"github.com/jessbludev/bina-mail/services/internal/session"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
)

// sessionCacheTTL acota la propagacion de una revocacion a 30 segundos.
const sessionCacheTTL = 30 * time.Second

// SessionCache es el contrato del cache opcional (Dragonfly, ADR-002/008).
// Ante error el handler sigue sin cache: es una optimizacion.
type SessionCache interface {
	Get(key string) (value string, found bool, err error)
	SetEx(key, value string, ttl time.Duration) error
	Del(key string) error
}

// Config parametriza el handler. Now se inyecta para determinismo.
type Config struct {
	Supabase   *supabase.Client
	JWTSecret  string
	SessionTTL time.Duration
	Now        func() time.Time
	Cache      SessionCache
}

// Handler expone la API interna de sesiones.
type Handler struct {
	cfg Config
	mux *http.ServeMux
}

// New construye el handler con la configuracion dada.
func New(cfg Config) *Handler {
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = session.TTLDefault
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	h := &Handler{cfg: cfg, mux: http.NewServeMux()}
	h.mux.HandleFunc("POST /internal/sessions", h.handleIssue)
	h.mux.HandleFunc("GET /internal/sessions/{hash}", h.handleLookup)
	h.mux.HandleFunc("DELETE /internal/sessions/{hash}", h.handleRevoke)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleIssue(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SupabaseJWT string `json:"supabase_jwt"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "payload_invalido")
		return
	}
	claims, err := jwt.VerifyHS256(body.SupabaseJWT, h.cfg.JWTSecret, h.cfg.Now())
	if err != nil {
		if errors.Is(err, jwt.ErrExpired) {
			writeError(w, http.StatusUnauthorized, "jwt_expirado")
			return
		}
		writeError(w, http.StatusUnauthorized, "jwt_invalido")
		return
	}
	rec, token, err := session.Issue(claims.Subject, h.cfg.SessionTTL, h.cfg.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "entropia")
		return
	}
	row := map[string]string{
		"token_hash": rec.TokenHash,
		"user_id":    rec.UserID,
		"expires_at": rec.ExpiresAt.UTC().Format(time.RFC3339),
	}
	if err := h.cfg.Supabase.Insert(r.Context(), "sessions", row); err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_sesiones")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"token":      token,
		"expires_at": rec.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *Handler) handleLookup(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	if len(hash) != 64 {
		writeError(w, http.StatusBadRequest, "hash_invalido")
		return
	}
	if h.cfg.Cache != nil {
		if v, found, err := h.cfg.Cache.Get("sess:" + hash); err == nil && found {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(v))
			return
		}
	}
	var rows []session.Record
	err := h.cfg.Supabase.Select(r.Context(), "sessions", map[string]string{"token_hash": hash}, "*", "", 1, &rows)
	if err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_sesiones")
		return
	}
	if len(rows) == 0 {
		writeError(w, http.StatusUnauthorized, "sesion_inexistente")
		return
	}
	rec := rows[0]
	if !rec.Valid(h.cfg.Now()) {
		writeError(w, http.StatusUnauthorized, "sesion_expirada")
		return
	}
	body, err := json.Marshal(map[string]string{
		"user_id":    rec.UserID,
		"expires_at": rec.ExpiresAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "interno")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
	if h.cfg.Cache != nil {
		ttl := rec.ExpiresAt.Sub(h.cfg.Now())
		if ttl > sessionCacheTTL {
			ttl = sessionCacheTTL
		}
		_ = h.cfg.Cache.SetEx("sess:"+hash, string(body), ttl)
	}
}

func (h *Handler) handleRevoke(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")
	if len(hash) != 64 {
		writeError(w, http.StatusBadRequest, "hash_invalido")
		return
	}
	row := map[string]string{
		"revoked_at": h.cfg.Now().UTC().Format(time.RFC3339),
	}
	if err := h.cfg.Supabase.Update(r.Context(), "sessions", map[string]string{"token_hash": hash}, row); err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_sesiones")
		return
	}
	if h.cfg.Cache != nil {
		_ = h.cfg.Cache.Del("sess:" + hash)
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
