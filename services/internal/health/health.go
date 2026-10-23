// Package health expone el endpoint de verificación de estado de cada servicio.
package health

import (
	"encoding/json"
	"net/http"
)

// Handler devuelve un handler con /healthz para el servicio indicado.
// ADR-005: la respuesta no contiene metadatos de mensajes ni de usuarios.
func Handler(service string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"service": service,
			"status":  "ok",
		})
	})
	return mux
}
