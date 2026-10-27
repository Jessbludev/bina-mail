// Package cache es un cliente RESP (Dragonfly) minimo en stdlib (ADR-008).
package cache

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ErrProtocol indica una respuesta del servidor fuera del protocolo.
var ErrProtocol = errors.New("cache: respuesta inesperada")

// Conn es una conexion RESP serializada: cada comando escribe y lee
// bajo mutex, suficiente para el patron de uso del gateway y directory.
// El cache es una optimizacion: ante error, el llamador sigue sin cache.
type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

// Dial abre una conexion TCP al servidor de cache.
func Dial(addr string) (*Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn, reader: bufio.NewReader(conn)}, nil
}

// Close cierra la conexion.
func (c *Conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}

func (c *Conn) do(args ...string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err := c.writeCommand(args); err != nil {
		return nil, err
	}
	return c.readReply()
}

func (c *Conn) writeCommand(args []string) error {
	var b strings.Builder
	b.WriteString("*")
	b.WriteString(strconv.Itoa(len(args)))
	b.WriteString("\r\n")
	for _, arg := range args {
		b.WriteString("$")
		b.WriteString(strconv.Itoa(len(arg)))
		b.WriteString("\r\n")
		b.WriteString(arg)
		b.WriteString("\r\n")
	}
	_, err := c.conn.Write([]byte(b.String()))
	return err
}

func (c *Conn) readReply() (any, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return nil, ErrProtocol
	}
	switch line[0] {
	case '+':
		return line[1:], nil
	case '-':
		return nil, errors.New("cache: " + line[1:])
	case ':':
		value, err := strconv.ParseInt(line[1:], 10, 64)
		if err != nil {
			return nil, ErrProtocol
		}
		return value, nil
	case '$':
		size, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, ErrProtocol
		}
		if size < 0 {
			return nil, nil
		}
		payload := make([]byte, size+2)
		if _, err := io.ReadFull(c.reader, payload); err != nil {
			return nil, err
		}
		return string(payload[:size]), nil
	default:
		return nil, ErrProtocol
	}
}
