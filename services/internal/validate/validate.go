// Package validate contiene validadores de borde compartidos.
package validate

import (
	"encoding/base64"
	"regexp"
)

var fingerprintRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Fingerprint acepta fingerprints BLAKE3 en hex minuscula de 64 caracteres.
func Fingerprint(s string) bool {
	return fingerprintRe.MatchString(s)
}

// Base64N acepta cadenas base64 (RFC 4648, con relleno) que decodifican a n bytes.
func Base64N(s string, n int) bool {
	b, err := base64.StdEncoding.DecodeString(s)
	return err == nil && len(b) == n
}

// Base64Len acepta base64 con longitud decodificada en [min, max].
func Base64Len(s string, min, max int) bool {
	b, err := base64.StdEncoding.DecodeString(s)
	return err == nil && len(b) >= min && len(b) <= max
}
