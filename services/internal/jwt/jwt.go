// Package jwt verifica tokens JWT HS256 emitidos por Supabase Auth.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMalformed = errors.New("jwt: token malformado")
	ErrBadAlg    = errors.New("jwt: algoritmo no soportado")
	ErrBadSig    = errors.New("jwt: firma invalida")
	ErrExpired   = errors.New("jwt: token expirado")
)

// Claims son los reclamos relevantes. Sin metadatos de mensajes (ADR-005).
type Claims struct {
	Subject string `json:"sub"`
	Expiry  int64  `json:"exp"`
	Issued  int64  `json:"iat"`
}

// VerifyHS256 valida firma y expiracion. El reloj se inyecta para determinismo.
func VerifyHS256(token, secret string, now time.Time) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return Claims{}, ErrMalformed
	}
	rawHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Claims{}, fmt.Errorf("%w: header", ErrMalformed)
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(rawHeader, &header); err != nil {
		return Claims{}, fmt.Errorf("%w: header json", ErrMalformed)
	}
	if header.Alg != "HS256" {
		return Claims{}, ErrBadAlg
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, ErrMalformed
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return Claims{}, ErrBadSig
	}
	rawClaims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrMalformed
	}
	var claims Claims
	if err := json.Unmarshal(rawClaims, &claims); err != nil {
		return Claims{}, fmt.Errorf("%w: claims json", ErrMalformed)
	}
	if claims.Subject == "" {
		return Claims{}, fmt.Errorf("%w: subject vacio", ErrMalformed)
	}
	if claims.Expiry != 0 && now.Unix() >= claims.Expiry {
		return Claims{}, ErrExpired
	}
	return claims, nil
}

// SignHS256 existe solo para pruebas: el emisor real es Supabase Auth.
func SignHS256(secret string, claims Claims) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
