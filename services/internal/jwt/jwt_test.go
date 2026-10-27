package jwt

import (
	"errors"
	"testing"
	"time"
)

var fixedNow = time.Unix(1700000000, 0).UTC()

func TestVerifyTokenValido(t *testing.T) {
	token, err := SignHS256("secreto", Claims{Subject: "usuario-1", Expiry: fixedNow.Unix() + 3600})
	if err != nil {
		t.Fatalf("SignHS256 fallo: %v", err)
	}
	claims, err := VerifyHS256(token, "secreto", fixedNow)
	if err != nil {
		t.Fatalf("VerifyHS256 fallo: %v", err)
	}
	if claims.Subject != "usuario-1" {
		t.Fatalf("subject = %q", claims.Subject)
	}
}

func TestVerifyFirmaInvalida(t *testing.T) {
	token, _ := SignHS256("secreto", Claims{Subject: "usuario-1", Expiry: fixedNow.Unix() + 3600})
	if _, err := VerifyHS256(token, "otro-secreto", fixedNow); !errors.Is(err, ErrBadSig) {
		t.Fatalf("error = %v, se esperaba ErrBadSig", err)
	}
}

func TestVerifyTokenExpirado(t *testing.T) {
	token, _ := SignHS256("secreto", Claims{Subject: "usuario-1", Expiry: fixedNow.Unix() - 10})
	if _, err := VerifyHS256(token, "secreto", fixedNow); !errors.Is(err, ErrExpired) {
		t.Fatalf("error = %v, se esperaba ErrExpired", err)
	}
}

func TestVerifyTokenMalformado(t *testing.T) {
	if _, err := VerifyHS256("no-es-un-jwt", "secreto", fixedNow); !errors.Is(err, ErrMalformed) {
		t.Fatalf("error = %v, se esperaba ErrMalformed", err)
	}
}
