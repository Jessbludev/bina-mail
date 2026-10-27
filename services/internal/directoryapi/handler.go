// Package directoryapi implementa el directorio de claves publicas.
package directoryapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/resp"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
	"github.com/jessbludev/bina-mail/services/internal/validate"
)

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
	// Cache es opcional (Dragonfly); nil desactiva la cache de claves.
	Cache resp.Cache
}

// keyCacheTTL acota la vida de las entradas de cache; la invalidacion
// por upsert cubre el indice por usuario (DEBT-004 para el anterior).
const keyCacheTTL = 60 * time.Second

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
	h.mux.HandleFunc("GET /internal/keys/by-user/{user_id}", h.handleLookupByUser)
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
		// Invalida el indice por usuario; la entrada previa por
		// fingerprint caduca por TTL (DEBT-004).
		_ = h.cfg.Cache.Del(r.Context(), "bina:keyuser:"+key.UserID)
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *Handler) handleLookup(w http.ResponseWriter, r *http.Request) {
	fingerprint := r.PathValue("fingerprint")
	if !validate.Fingerprint(fingerprint) {
		writeError(w, http.StatusBadRequest, "fingerprint_invalido")
		return
	}
	cacheKey := "bina:keyfp:" + fingerprint
	if row, ok := h.cachedKey(r.Context(), cacheKey); ok {
		writeJSON(w, http.StatusOK, row)
		return
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
	h.storeKey(r.Context(), cacheKey, rows[0])
	writeJSON(w, http.StatusOK, rows[0])
}

func (h *Handler) handleLookupByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" || len(userID) > 128 {
		writeError(w, http.StatusBadRequest, "usuario_invalido")
		return
	}
	cacheKey := "bina:keyuser:" + userID
	if row, ok := h.cachedKey(r.Context(), cacheKey); ok {
		writeJSON(w, http.StatusOK, row)
		return
	}
	var rows []KeyRow
	err := h.cfg.Supabase.Select(r.Context(), "key_directory", map[string]string{
		"user_id": userID,
		"active":  "true",
	}, "*", "", 1, &rows)
	if err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_claves")
		return
	}
	if len(rows) == 0 {
		writeError(w, http.StatusNotFound, "clave_no_encontrada")
		return
	}
	h.storeKey(r.Context(), cacheKey, rows[0])
	writeJSON(w, http.StatusOK, rows[0])
}

func (h *Handler) cachedKey(ctx context.Context, key string) (KeyRow, bool) {
	if h.cfg.Cache == nil {
		return KeyRow{}, false
	}
	value, ok, err := h.cfg.Cache.Get(ctx, key)
	if err != nil || !ok {
		return KeyRow{}, false
	}
	var row KeyRow
	if err := json.Unmarshal([]byte(value), &row); err != nil {
		return KeyRow{}, false
	}
	return row, true
}

func (h *Handler) storeKey(ctx context.Context, key string, row KeyRow) {
	if h.cfg.Cache == nil {
		return
	}
	value, err := json.Marshal(row)
	if err != nil {
		return
	}
	_ = h.cfg.Cache.SetEx(ctx, key, string(value), keyCacheTTL)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
