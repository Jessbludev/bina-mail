package validate

import (
	"strings"
	"testing"
)

func TestFingerprint(t *testing.T) {
	if !Fingerprint(strings.Repeat("a", 64)) {
		t.Fatal("fingerprint hex valido rechazado")
	}
	invalid := []string{"", "XYZ", strings.Repeat("A", 64), strings.Repeat("a", 63), strings.Repeat("a", 65)}
	for _, s := range invalid {
		if Fingerprint(s) {
			t.Fatalf("fingerprint invalido aceptado: %q", s)
		}
	}
}

func TestBase64N(t *testing.T) {
	if !Base64N("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", 32) {
		t.Fatal("base64 de 32 bytes rechazado")
	}
	if Base64N("AAAA", 32) {
		t.Fatal("base64 corto aceptado")
	}
}
