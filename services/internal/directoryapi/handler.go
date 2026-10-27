// Package directoryapi implementa el directorio de claves publicas.
package directoryapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/supabase"
	"github.com/jessbludev/bina-mail/services/internal/validate"
)

// keyCacheTTL acota la propagacion de una rotacion de claves a 60 segundos.
const keyCacheTTL = 60 * time.Second

// KeyCache es el contrato del cache opcional (Dragonfly, ADR-002/008).
// Ante error el handler sigue sin cache: es una optimizacion.
type KeyCache interface {
	Get(key string) (value string, found bool, err error)
	SetEx(key, value string, ttl time.Duration) error
	Del(key string) error
}

// KeyRow es la fila del directorio de claves.
type KeyRow struct {
	UserID                   string `json:"user_id"`
	X25519PublicKey         string `json:"x25519_public_key"`
	Ed25519SigningPublicKey string `json:"ed25519_signing_public_key"`
	Fingerprint             string `json:"fingerprint"`
	Active                  bool   `json:"active"`
}

// Config parametriza el handler.
type Config struct {
	Supabase *supabase.Client
	Cache    KeyCache
}

// Handler expone la API interna del directorio.
type Handler struct {
	cfg Config
	mux *http.ServeMux
}

// New construye el handler del directorio.
func New(cfg Config) *Handler {
	h := &Handler{cfg: cfg, mux: http.NewServeMux()}
	h.mux.HandleFunc("POST /internal/keys", h.handlePublish)
	h.mux.HandleFunc("GET /internal/keys/{fingerprint}", h.handleLookup)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handlePublish(w http.ResponseWriter, r *http.Request) {
	var key struct {
		UserID                   string `json:"user_id"`
		X25519PublicKey         string `json:"x25519_public_key"`
		Ed25519SigningPublicKey string `json:"ed25519_signing_public_key"`
		Fingerprint             string `json:"fingerprint"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024)).Decode(&key); err != nil {
		writeError(w, http.StatusBadRequest, "payload_invalido")
		return
	}
	if key.UserID == "" || !validate.Fingerprint(key.Fingerprint) ||
		!validate.Base64N(key.X25519PublicKey, 32) || !validate.Base64N(key.Ed25519SigningPublicKey, 32) {
		writeError(w, http.StatusBadRequest, "clave_invalida")
		return
	}
	row := map[string]string{
		"user_id":                    key.UserID,
		"x25519_public_key":          key.X25519PublicKey,
		"ed25519_signing_public_key": key.Ed25519SigningPublicKey,
		"fingerprint":                key.Fingerprint,
	}
	if err := h.cfg.Supabase.Upsert(r.Context(), "key_directory", "user_id", row); err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_claves")
		return
	}
	if h.cfg.Cache != nil {
		_ = h.cfg.Cache.Del("fp:" + key.Fingerprint)
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) handleLookup(w http.ResponseWriter, r *http.Request) {
	fingerprint := r.PathValue("fingerprint")
	if !validate.Fingerprint(fingerprint) {
		writeError(w, http.StatusBadRequest, "fingerprint_invalido")
		return
	}
	if h.cfg.Cache != nil {
		if v, found, err := h.cfg.Cache.Get("fp:" + fingerprint); err == nil && found {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(v))
			return
		}
	}
	var rows []KeyRow
	err := h.cfg.Supabase.Select(r.Context(), "key_directory", map[string]string{
		"fingerprint": fingerprint,
		"active":      "true",
	}, "*", "", 1, &rows)
	if err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_claves")
		return
	}
	if len(rows) == 0 {
		writeError(w, http.StatusNotFound, "clave_no_encontrada")
		return
	}
	body, err := json.Marshal(rows[0])
	if err != nil {
		writeError(w, http.StatusInternalServerError, "interno")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
	if h.cfg.Cache != nil {
		_ = h.cfg.Cache.SetEx("fp:"+fingerprint, string(body), keyCacheTTL)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
