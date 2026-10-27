// Package session emite y verifica tokens de sesion opacos de corta vida.
package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// TTLDefault es la vida por defecto de una sesion.
const TTLDefault = 30 * time.Minute

// ErrEntropy indica fallo del generador de entropia.
var ErrEntropy = errors.New("session: fallo de entropia")

// Record es la fila persistida: solo el hash del token, nunca el token.
type Record struct {
	TokenHash string     `json:"token_hash"`
	UserID    string     `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// Hash devuelve el hash SHA-256 en hex del token opaco.
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Issue genera un token opaco y su registro para persistir.
func Issue(userID string, ttl time.Duration, now time.Time) (Record, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Record{}, "", fmt.Errorf("%w: %v", ErrEntropy, err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return Record{
		TokenHash: Hash(token),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}, token, nil
}

// Valid informa si el registro sigue activo en el instante dado.
func (r Record) Valid(now time.Time) bool {
	return r.RevokedAt == nil && now.Before(r.ExpiresAt)
}
