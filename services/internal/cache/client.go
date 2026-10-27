package cache

import (
	"strconv"
	"time"
)

// Ping verifica la disponibilidad del servidor.
func (c *Conn) Ping() error {
	_, err := c.do("PING")
	return err
}

// Get devuelve el valor de la clave; found indica si existe.
func (c *Conn) Get(key string) (value string, found bool, err error) {
	reply, err := c.do("GET", key)
	if err != nil {
		return "", false, err
	}
	s, ok := reply.(string)
	return s, ok, nil
}

// SetEx escribe la clave con expiracion.
func (c *Conn) SetEx(key, value string, ttl time.Duration) error {
	_, err := c.do("SETEX", key, strconv.Itoa(ttlSeconds(ttl)), value)
	return err
}

// Del elimina la clave si existe.
func (c *Conn) Del(key string) error {
	_, err := c.do("DEL", key)
	return err
}

// Incr incrementa el contador de la clave y devuelve el nuevo valor.
func (c *Conn) Incr(key string) (int64, error) {
	reply, err := c.do("INCR", key)
	if err != nil {
		return 0, err
	}
	n, ok := reply.(int64)
	if !ok {
		return 0, ErrProtocol
	}
	return n, nil
}

// Expire fija la expiracion de una clave existente.
func (c *Conn) Expire(key string, ttl time.Duration) error {
	_, err := c.do("EXPIRE", key, strconv.Itoa(ttlSeconds(ttl)))
	return err
}

func ttlSeconds(ttl time.Duration) int {
	secs := int(ttl / time.Second)
	if secs < 1 {
		secs = 1
	}
	return secs
}
