package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	tests := []struct {
		service string
		want    string
	}{
		{"gateway", "{\"service\":\"gateway\",\"status\":\"ok\"}\n"},
		{"mailbox", "{\"service\":\"mailbox\",\"status\":\"ok\"}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.service, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			Handler(tt.service).ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("estado = %d, se esperaba %d", rec.Code, http.StatusOK)
			}
			if got := rec.Body.String(); got != tt.want {
				t.Fatalf("cuerpo = %q, se esperaba %q", got, tt.want)
			}
		})
	}
}
