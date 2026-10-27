// Package config lee configuracion de entorno con valores por defecto.
package config

import "os"

// Env devuelve la variable de entorno o el valor por defecto.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
