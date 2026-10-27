package session

import (
	"testing"
	"time"
)

var issueNow = time.Unix(1700000000, 0).UTC()

func TestIssueGeneraTokenOpaco(t *testing.T) {
	rec, token, err := Issue("usuario-1", time.Minute, issueNow)
	if err != nil {
		t.Fatalf("Issue fallo: %v", err)
	}
	if len(token) != 43 {
		t.Fatalf("longitud del token = %d, se esperaba 43", len(token))
	}
	if rec.TokenHash != Hash(token) {
		t.Fatal("el hash del registro no corresponde al token")
	}
	if rec.UserID != "usuario-1" {
		t.Fatalf("user_id = %q", rec.UserID)
	}
}

func TestIssueEsAleatorio(t *testing.T) {
	_, a, err := Issue("usuario-1", time.Minute, issueNow)
	if err != nil {
		t.Fatalf("Issue fallo: %v", err)
	}
	_, b, err := Issue("usuario-1", time.Minute, issueNow)
	if err != nil {
		t.Fatalf("Issue fallo: %v", err)
	}
	if a == b {
		t.Fatal("dos emisiones produjeron el mismo token")
	}
}

func TestValid(t *testing.T) {
	rec, _, _ := Issue("usuario-1", time.Minute, issueNow)
	if !rec.Valid(issueNow.Add(30 * time.Second)) {
		t.Fatal("la sesion debia seguir activa")
	}
	if rec.Valid(issueNow.Add(2 * time.Minute)) {
		t.Fatal("la sesion debia estar expirada")
	}
	revoked := issueNow.Add(-time.Hour)
	rec.RevokedAt = &revoked
	if rec.Valid(issueNow) {
		t.Fatal("la sesion revocada debia ser invalida")
	}
}
