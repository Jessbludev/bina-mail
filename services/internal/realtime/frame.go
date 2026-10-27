// Tramas WebSocket (RFC 6455) en stdlib pura.
package realtime

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const (
	opContinuation byte = 0x0
	opText         byte = 0x1
	opBinary       byte = 0x2
	opClose        byte = 0x8
	opPing         byte = 0x9
	opPong         byte = 0xA
)

const (
	maxMessageBytes = 16 * 1024
	maxFrameBytes   = 1 << 20
	writeTimeout    = 5 * time.Second
)

var (
	errProtocol = errors.New("realtime: trama invalida")
	errTooLarge = errors.New("realtime: mensaje demasiado grande")
	errClosed   = errors.New("realtime: conexion cerrada")
)

// wsConn envuelve una conexion WebSocket ya negociada.
type wsConn struct {
	net.Conn
	br          *bufio.Reader
	writeMu     sync.Mutex
	requireMask bool
}

// writeFrame escribe una trama; el cliente debe enviar con mask=true
// y el servidor responde sin mascara (RFC 6455).
func (c *wsConn) writeFrame(opcode byte, payload []byte, mask bool) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	buf := make([]byte, 0, len(payload)+14)
	buf = append(buf, 0x80|opcode)
	length := len(payload)
	var maskBit byte
	if mask {
		maskBit = 0x80
	}
	switch {
	case length < 126:
		buf = append(buf, maskBit|byte(length))
	case length <= 0xFFFF:
		buf = append(buf, maskBit|126)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		buf = append(buf, ext[:]...)
	default:
		buf = append(buf, maskBit|127)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		buf = append(buf, ext[:]...)
	}
	if mask {
		var key [4]byte
		if _, err := rand.Read(key[:]); err != nil {
			return err
		}
		buf = append(buf, key[:]...)
		masked := make([]byte, length)
		for i := 0; i < length; i++ {
			masked[i] = payload[i] ^ key[i%4]
		}
		payload = masked
	}
	buf = append(buf, payload...)
	_ = c.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err := c.Conn.Write(buf)
	return err
}

// writeMessage escribe una trama de datos del servidor (sin mascara).
func (c *wsConn) writeMessage(opcode byte, payload []byte) error {
	return c.writeFrame(opcode, payload, false)
}

// readFrame lee una trama individual y des-enmascara el payload.
func (c *wsConn) readFrame() (opcode byte, fin bool, payload []byte, err error) {
	var head [2]byte
	if _, err = io.ReadFull(c.br, head[:]); err != nil {
		return
	}
	fin = head[0]&0x80 != 0
	opcode = head[0] & 0x0F
	masked := head[1]&0x80 != 0
	if c.requireMask && !masked {
		err = errProtocol
		return
	}
	length := int64(head[1] & 0x7F)
	switch length {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(c.br, ext[:]); err != nil {
			return
		}
		length = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(c.br, ext[:]); err != nil {
			return
		}
		length = int64(binary.BigEndian.Uint64(ext[:]))
	}
	if length > maxFrameBytes {
		err = errTooLarge
		return
	}
	payload = make([]byte, length)
	if _, err = io.ReadFull(c.br, payload); err != nil {
		return
	}
	if masked {
		var key [4]byte
		if _, err = io.ReadFull(c.br, key[:]); err != nil {
			return
		}
		for i := range payload {
			payload[i] ^= key[i%4]
		}
	}
	return
}

// readMessage ensambla mensajes completos y gestiona tramas de control:
// responde ping con pong y cierra ante una trama close del cliente.
func (c *wsConn) readMessage() (opcode byte, payload []byte, err error) {
	var msg []byte
	var msgOp byte
	for {
		op, fin, frame, ferr := c.readFrame()
		if ferr != nil {
			return 0, nil, ferr
		}
		switch op {
		case opPing:
			if ferr = c.writeMessage(opPong, frame); ferr != nil {
				return 0, nil, ferr
			}
		case opPong:
			// Latencia: se ignora en el MVP.
		case opClose:
			_ = c.writeMessage(opClose, frame)
			return 0, nil, errClosed
		case opText, opBinary:
			if msgOp != 0 {
				return 0, nil, errProtocol
			}
			msgOp = op
			msg = append(msg, frame...)
			if fin {
				return msgOp, msg, nil
			}
		case opContinuation:
			if msgOp == 0 {
				return 0, nil, errProtocol
			}
			msg = append(msg, frame...)
			if fin {
				return msgOp, msg, nil
			}
		default:
			return 0, nil, errProtocol
		}
		if len(msg) > maxMessageBytes {
			return 0, nil, errTooLarge
		}
	}
}
