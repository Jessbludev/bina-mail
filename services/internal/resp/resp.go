// Package resp implementa un cliente minimo del protocolo Redis (RESP2)
// para Dragonfly, sin dependencias externas (ADR-008 confina las
// dependencias a los clientes de infraestructura: franz-go y gorilla).
package resp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Cache es la interfaz de cache usada por los servicios; nil la desactiva.
type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	SetEx(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Counter incrementa contadores de ventana fija para rate limiting.
type Counter interface {
	IncrWindow(ctx context.Context, key string, window time.Duration) (int64, error)
}

// maxBulk limita el tamano de respuesta bulk aceptada (defensa de memoria).
const maxBulk = 1 << 20

// Client es un cliente Dragonfly con una conexion persistente protegida
// por mutex y una reconexion por comando como maximo.
type Client struct {
	mu      sync.Mutex
	addr    string
	timeout time.Duration
	conn    net.Conn
	reader  *bufio.Reader
}

// Dial crea el cliente y verifica la conexion con PING.
func Dial(addr string, timeout time.Duration) (*Client, error) {
	c := &Client{addr: addr, timeout: timeout}
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.connectLocked(); err != nil {
		return nil, err
	}
	if _, err := c.attemptLocked([]string{"PING"}); err != nil {
		_ = c.conn.Close()
		c.conn = nil
		return nil, err
	}
	return c, nil
}

// Close cierra la conexion subyacente.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Do ejecuta un comando con un unico reintento ante errores de red.
func (c *Client) Do(args ...string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(args) == 0 {
		return nil, errors.New("resp: comando vacio")
	}
	reply, err := c.attemptLocked(args)
	if err != nil && isNetError(err) {
		reply, err = c.attemptLocked(args)
	}
	return reply, err
}

func (c *Client) attemptLocked(args []string) (any, error) {
	if c.conn == nil {
		if err := c.connectLocked(); err != nil {
			return nil, err
		}
	}
	_ = c.conn.SetDeadline(time.Now().Add(c.timeout))
	if _, err := c.conn.Write(encodeCommand(args)); err != nil {
		_ = c.conn.Close()
		c.conn = nil
		return nil, err
	}
	reply, err := readReply(c.reader)
	if err != nil {
		_ = c.conn.Close()
		c.conn = nil
		return nil, err
	}
	return reply, nil
}

func (c *Client) connectLocked() error {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	c.reader = bufio.NewReader(conn)
	return nil
}

func isNetError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, io.EOF)
}

// Get devuelve el valor de una clave; ok informa si existia.
func (c *Client) Get(ctx context.Context, key string) (string, bool, error) {
	if err := checkCtx(ctx); err != nil {
		return "", false, err
	}
	reply, err := c.Do("GET", key)
	if err != nil {
		return "", false, err
	}
	value, ok := reply.(string)
	if !ok {
		return "", false, nil
	}
	return value, true, nil
}

// SetEx escribe una clave con expiracion en segundos.
func (c *Client) SetEx(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	_, err := c.Do("SET", key, value, "EX", seconds(ttl))
	return err
}

// Del elimina claves.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	if err := checkCtx(ctx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	_, err := c.Do(append([]string{"DEL"}, keys...)...)
	return err
}

// IncrWindow incrementa un contador y le fija la ventana de expiracion
// cuando es la primera cuenta de la ventana.
func (c *Client) IncrWindow(ctx context.Context, key string, window time.Duration) (int64, error) {
	if err := checkCtx(ctx); err != nil {
		return 0, err
	}
	reply, err := c.Do("INCR", key)
	if err != nil {
		return 0, err
	}
	n, ok := reply.(int64)
	if !ok {
		return 0, errors.New("resp: respuesta INCR inesperada")
	}
	if n == 1 {
		_, _ = c.Do("EXPIRE", key, seconds(window))
	}
	return n, nil
}

func checkCtx(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func seconds(d time.Duration) string {
	s := int64(d / time.Second)
	if s < 1 {
		s = 1
	}
	return strconv.FormatInt(s, 10)
}

func encodeCommand(args []string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(arg), arg)
	}
	return []byte(b.String())
}

func readReply(r *bufio.Reader) (any, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return nil, errors.New("resp: respuesta vacia")
	}
	switch line[0] {
	case '+':
		return line[1:], nil
	case '-':
		return nil, errors.New("resp: " + line[1:])
	case ':':
		return strconv.ParseInt(line[1:], 10, 64)
	case '$':
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if n == -1 {
			return nil, nil
		}
		if n < 0 || n > maxBulk {
			return nil, errors.New("resp: tamano bulk fuera de rango")
		}
		buf := make([]byte, n+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		return string(buf[:n]), nil
	case '*':
		n, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			return nil, nil
		}
		items := make([]any, 0, n)
		for i := 0; i < n; i++ {
			item, err := readReply(r)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("resp: tipo de respuesta desconocido: %q", line[0])
	}
}
