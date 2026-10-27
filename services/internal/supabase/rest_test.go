package supabase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInsertCabecerasYEstado(t *testing.T) {
	var gotKey, gotAuth, gotPrefer, gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("apikey")
		gotAuth = r.Header.Get("Authorization")
		gotPrefer = r.Header.Get("Prefer")
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New(srv.URL, "clave-servicio")
	if err := c.Insert(context.Background(), "sessions", map[string]string{"token_hash": "h"}); err != nil {
		t.Fatalf("Insert fallo: %v", err)
	}
	if gotKey != "clave-servicio" || gotAuth != "Bearer clave-servicio" {
		t.Fatalf("cabeceras: apikey=%q authorization=%q", gotKey, gotAuth)
	}
	if gotPrefer != "return=minimal" {
		t.Fatalf("Prefer = %q", gotPrefer)
	}
	if gotMethod != http.MethodPost || gotPath != "/rest/v1/sessions" {
		t.Fatalf("peticion: %s %s", gotMethod, gotPath)
	}
}

func TestInsertFallaConEstadoDeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL, "clave")
	if err := c.Insert(context.Background(), "sessions", map[string]string{}); err == nil {
		t.Fatal("se esperaba error con estado 500")
	}
}

func TestUpsertUsaOnConflict(t *testing.T) {
	var gotQuery, gotPrefer string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotPrefer = r.Header.Get("Prefer")
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	c := New(srv.URL, "clave")
	if err := c.Upsert(context.Background(), "key_directory", "user_id", map[string]string{}); err != nil {
		t.Fatalf("Upsert fallo: %v", err)
	}
	if gotQuery != "on_conflict=user_id" {
		t.Fatalf("query = %q", gotQuery)
	}
	if gotPrefer != "resolution=merge-duplicates,return=minimal" {
		t.Fatalf("Prefer = %q", gotPrefer)
	}
}

func TestSelectAplicaFiltros(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"token_hash":"h","user_id":"u"}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "clave")
	var out []map[string]any
	err := c.Select(context.Background(), "sessions", map[string]string{"token_hash": "h"}, "*", "", 1, &out)
	if err != nil {
		t.Fatalf("Select fallo: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("filas = %d", len(out))
	}
	if gotQuery != "limit=1&select=*&token_hash=eq.h" {
		t.Fatalf("query = %q", gotQuery)
	}
}

func TestUpdateAplicaFiltrosPatch(t *testing.T) {
	var gotMethod, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, "clave")
	err := c.Update(context.Background(), "sessions", map[string]string{"token_hash": "h"}, map[string]string{"revoked_at": "2026-10-23T00:00:00Z"})
	if err != nil {
		t.Fatalf("Update fallo: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Fatalf("metodo = %s", gotMethod)
	}
	if gotQuery != "token_hash=eq.h" {
		t.Fatalf("query = %q", gotQuery)
	}
}
