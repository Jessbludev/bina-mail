// Package config lee configuracion de entorno con valores por defecto.
package config

import (
	"os"
	"strings"
)

// Env devuelve la variable de entorno o el valor por defecto.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// CSV devuelve una lista separada por comas de la variable de entorno,
// sin entradas vacias, o el valor por defecto.
func CSV(key, fallback string) []string {
	raw := Env(key, fallback)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
