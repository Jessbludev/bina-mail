// Package supabase es un cliente REST (PostgREST) minimo sin dependencias
// externas, alineado con ADR-007 (compilacion determinista con stdlib).
package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client accede a la API REST de Supabase con el rol de servicio.
type Client struct {
	BaseURL    string
	ServiceKey string
	HTTP       *http.Client
}

// New crea el cliente para una URL base y una clave de rol de servicio.
func New(baseURL, serviceKey string) *Client {
	return &Client{BaseURL: baseURL, ServiceKey: serviceKey, HTTP: http.DefaultClient}
}

func (c *Client) request(ctx context.Context, method, path string, body any, prefer string) (*http.Response, error) {
	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		payload = b
	}
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if prefer != "" {
		req.Header.Set("Prefer", prefer)
	}
	return c.HTTP.Do(req)
}

// Insert crea filas en la tabla sin devolver representacion.
func (c *Client) Insert(ctx context.Context, table string, row any) error {
	resp, err := c.request(ctx, http.MethodPost, "/rest/v1/"+table, row, "return=minimal")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("supabase: insert en %s devolvio %d", table, resp.StatusCode)
	}
	return nil
}

// InsertReturning crea una fila y decodifica su representacion en out
// (puntero a slice).
func (c *Client) InsertReturning(ctx context.Context, table string, row any, out any) error {
	resp, err := c.request(ctx, http.MethodPost, "/rest/v1/"+table, row, "return=representation")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("supabase: insert en %s devolvio %d", table, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Upsert inserta o actualiza por la columna de conflicto indicada.
func (c *Client) Upsert(ctx context.Context, table, onConflict string, row any) error {
	path := "/rest/v1/" + table + "?on_conflict=" + url.QueryEscape(onConflict)
	resp, err := c.request(ctx, http.MethodPost, path, row, "resolution=merge-duplicates,return=minimal")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("supabase: upsert en %s devolvio %d", table, resp.StatusCode)
	}
	return nil
}

// Update modifica las filas que cumplen los filtros (columna -> valor).
func (c *Client) Update(ctx context.Context, table string, filters map[string]string, row any) error {
	q := url.Values{}
	for k, v := range filters {
		q.Set(k, "eq."+v)
	}
	path := "/rest/v1/" + table + "?" + q.Encode()
	resp, err := c.request(ctx, http.MethodPatch, path, row, "return=minimal")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("supabase: update en %s devolvio %d", table, resp.StatusCode)
	}
	return nil
}

// Select consulta filas. order es opcional, p. ej. "created_at.desc".
func (c *Client) Select(ctx context.Context, table string, filters map[string]string, columns, order string, limit int, out any) error {
	q := url.Values{}
	q.Set("select", columns)
	for k, v := range filters {
		q.Set(k, "eq."+v)
	}
	if order != "" {
		q.Set("order", order)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := "/rest/v1/" + table + "?" + q.Encode()
	resp, err := c.request(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("supabase: select en %s devolvio %d", table, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
