// Package mailboxapi implementa el buzon de sobres cifrados E2EE.
package mailboxapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jessbludev/bina-mail/services/internal/events"
	"github.com/jessbludev/bina-mail/services/internal/supabase"
	"github.com/jessbludev/bina-mail/services/internal/validate"
)

// Envelope es la representacion JSON del sobre, alineada con packages/contracts.
type Envelope struct {
	ID                   string          `json:"id"`
	SenderFingerprint    string          `json:"sender_fingerprint,omitempty"`
	RecipientFingerprint string          `json:"recipient_fingerprint"`
	Nonce                string          `json:"nonce"`
	Ciphertext           string          `json:"ciphertext"`
	Wraps                json.RawMessage `json:"wraps"`
	BLAKE3Fingerprint    string          `json:"blake3_fingerprint"`
	SizeBytes            int             `json:"size_bytes"`
	CreatedAt            time.Time       `json:"created_at"`
}

type envelopeRequest struct {
	SenderFingerprint    string         `json:"sender_fingerprint,omitempty"`
	RecipientFingerprint string         `json:"recipient_fingerprint"`
	Nonce                string         `json:"nonce"`
	Ciphertext           string         `json:"ciphertext"`
	Wraps                []envelopeWrap `json:"wraps"`
	BLAKE3Fingerprint    string         `json:"blake3_fingerprint"`
}

type envelopeWrap struct {
	RecipientKeyID     string `json:"recipient_key_id"`
	EphemeralPublicKey string `json:"ephemeral_public_key"`
	WrappedKey         string `json:"wrapped_key"`
}

type envelopeRow struct {
	ID                   string          `json:"id"`
	SenderFingerprint    *string         `json:"sender_fingerprint"`
	RecipientFingerprint string          `json:"recipient_fingerprint"`
	Nonce                []byte          `json:"nonce"`
	Ciphertext           []byte          `json:"ciphertext"`
	Wraps                json.RawMessage `json:"wraps"`
	BLAKE3Fingerprint    string          `json:"blake3_fingerprint"`
	SizeBytes            int             `json:"size_bytes"`
	CreatedAt            time.Time       `json:"created_at"`
}

// Publisher envia senales de sobre nuevo al bus (Redpanda via
// kafka.NewProducer); nil desactiva la publicacion.
type Publisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

// Config parametriza el handler. Publisher es opcional (mejor esfuerzo).
type Config struct {
	Supabase         *supabase.Client
	MaxEnvelopeBytes int64
	// Publisher es opcional (Redpanda); nil desactiva las senales.
	Publisher Publisher
}

// Handler expone la API interna del buzon.
type Handler struct {
	cfg Config
	mux *http.ServeMux
}

// New construye el handler del buzon.
func New(cfg Config) *Handler {
	if cfg.MaxEnvelopeBytes <= 0 {
		cfg.MaxEnvelopeBytes = 1 << 20
	}
	h := &Handler{cfg: cfg, mux: http.NewServeMux()}
	h.mux.HandleFunc("POST /internal/envelopes", h.handleStore)
	h.mux.HandleFunc("GET /internal/envelopes", h.handleList)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleStore(w http.ResponseWriter, r *http.Request) {
	var req envelopeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, h.cfg.MaxEnvelopeBytes+(64<<10))).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "payload_invalido")
		return
	}
	if !h.validRequest(&req) {
		writeError(w, http.StatusBadRequest, "sobre_invalido")
		return
	}
	nonce, errN := base64.StdEncoding.DecodeString(req.Nonce)
	ciphertext, errC := base64.StdEncoding.DecodeString(req.Ciphertext)
	if errN != nil || errC != nil || len(nonce) != 24 || len(ciphertext) == 0 || int64(len(ciphertext)) > h.cfg.MaxEnvelopeBytes {
		writeError(w, http.StatusBadRequest, "sobre_invalido")
		return
	}
	sender := req.SenderFingerprint
	row := map[string]any{
		"sender_fingerprint":    &sender,
		"recipient_fingerprint": req.RecipientFingerprint,
		"nonce":                 nonce,
		"ciphertext":            ciphertext,
		"wraps":                 req.Wraps,
		"blake3_fingerprint":    req.BLAKE3Fingerprint,
		"size_bytes":            len(ciphertext),
	}
	var rows []envelopeRow
	if err := h.cfg.Supabase.InsertReturning(r.Context(), "envelopes", row, &rows); err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_sobres")
		return
	}
	if len(rows) == 0 {
		writeError(w, http.StatusInternalServerError, "sin_representacion")
		return
	}
	if h.cfg.Publisher != nil {
		payload := events.NewEnvelopeStored(rows[0].ID, req.RecipientFingerprint, time.Now().UTC())
		// Mejor esfuerzo e idempotente (ADR-008): la senal es un aviso sin
		// contenido; si la entrega falla, el cliente refresca su buzon.
		_ = h.cfg.Publisher.Publish(r.Context(), []byte(req.RecipientFingerprint), payload)
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": rows[0].ID})
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
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
	var rows []envelopeRow
	err := h.cfg.Supabase.Select(r.Context(), "envelopes", map[string]string{
		"recipient_fingerprint": recipient,
	}, "*", "created_at.desc", limit, &rows)
	if err != nil {
		writeError(w, http.StatusBadGateway, "persistencia_sobres")
		return
	}
	out := make([]Envelope, 0, len(rows))
	for _, row := range rows {
		out = append(out, Envelope{
			ID:                   row.ID,
			SenderFingerprint:    deref(row.SenderFingerprint),
			RecipientFingerprint: row.RecipientFingerprint,
			Nonce:                base64.StdEncoding.EncodeToString(row.Nonce),
			Ciphertext:           base64.StdEncoding.EncodeToString(row.Ciphertext),
			Wraps:                row.Wraps,
			BLAKE3Fingerprint:    row.BLAKE3Fingerprint,
			SizeBytes:            row.SizeBytes,
			CreatedAt:            row.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) validRequest(req *envelopeRequest) bool {
	if !validate.Fingerprint(req.RecipientFingerprint) || !validate.Fingerprint(req.BLAKE3Fingerprint) {
		return false
	}
	if req.SenderFingerprint != "" && !validate.Fingerprint(req.SenderFingerprint) {
		return false
	}
	if len(req.Wraps) == 0 {
		return false
	}
	for _, wrap := range req.Wraps {
		if wrap.RecipientKeyID == "" || !validate.Fingerprint(wrap.RecipientKeyID) ||
			!validate.Base64N(wrap.EphemeralPublicKey, 32) || !validate.Base64Len(wrap.WrappedKey, 40, 128) {
			return false
		}
	}
	return true
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}
