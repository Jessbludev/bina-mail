package resp

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// stubServer es un Dragonfly minimo en proceso para pruebas.
type stubServer struct {
	mu       sync.Mutex
	store    map[string]string
	counters map[string]int64
	commands []string
	ln       net.Listener
}

func newStubServer(t *testing.T) *stubServer {
	t.Helper()
	s := &stubServer{store: map[string]string{}, counters: map[string]int64{}}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen fallo: %v", err)
	}
	s.ln = ln
	go s.acceptLoop()
	t.Cleanup(func() { _ = ln.Close() })
	return s
}

func (s *stubServer) addr() string { return s.ln.Addr().String() }

func (s *stubServer) recorded() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.commands...)
}

func (s *stubServer) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *stubServer) handleConn(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	for {
		args, err := readCommand(reader)
		if err != nil {
			return
		}
		s.mu.Lock()
		s.commands = append(s.commands, strings.Join(args, " "))
		reply := s.execute(args)
		s.mu.Unlock()
		if _, err := conn.Write([]byte(reply)); err != nil {
			return
		}
	}
}

func readCommand(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")
	if len(line) < 2 || line[0] != '*' {
		return nil, errors.New("stub: comando invalido")
	}
	n, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		header, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		header = strings.TrimRight(header, "\r\n")
		if len(header) < 2 || header[0] != '$' {
			return nil, errors.New("stub: encabezado invalido")
		}
		size, err := strconv.Atoi(header[1:])
		if err != nil {
			return nil, err
		}
		buf := make([]byte, size+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:size]))
	}
	return args, nil
}

func (s *stubServer) execute(args []string) string {
	switch args[0] {
	case "PING":
		return "+PONG\r\n"
	case "GET":
		if value, ok := s.store[args[1]]; ok {
			return "$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"
		}
		return "$-1\r\n"
	case "SET":
		s.store[args[1]] = args[2]
		return "+OK\r\n"
	case "DEL":
		delete(s.store, args[1])
		return ":1\r\n"
	case "INCR":
		s.counters[args[1]]++
		return ":" + strconv.FormatInt(s.counters[args[1]], 10) + "\r\n"
	case "EXPIRE":
		return ":1\r\n"
	}
	return "-ERR stub: comando desconocido\r\n"
}

func TestDialGetSetDel(t *testing.T) {
	stub := newStubServer(t)
	client, err := Dial(stub.addr(), time.Second)
	if err != nil {
		t.Fatalf("Dial fallo: %v", err)
	}
	if err := client.SetEx(context.Background(), "k1", "v1", time.Minute); err != nil {
		t.Fatalf("SetEx fallo: %v", err)
	}
	value, ok, err := client.Get(context.Background(), "k1")
	if err != nil || !ok || value != "v1" {
		t.Fatalf("Get = (%q, %v, %v)", value, ok, err)
	}
	_, ok, err = client.Get(context.Background(), "k2")
	if err != nil || ok {
		t.Fatalf("Get de clave inexistente = (%v, %v)", ok, err)
	}
	if err := client.Del(context.Background(), "k1"); err != nil {
		t.Fatalf("Del fallo: %v", err)
	}
	_, ok, _ = client.Get(context.Background(), "k1")
	if ok {
		t.Fatal("la clave debia eliminarse")
	}
}

func TestIncrWindowFijaExpiracion(t *testing.T) {
	stub := newStubServer(t)
	client, err := Dial(stub.addr(), time.Second)
	if err != nil {
		t.Fatalf("Dial fallo: %v", err)
	}
	first, err := client.IncrWindow(context.Background(), "rl:u1", time.Minute)
	if err != nil || first != 1 {
		t.Fatalf("IncrWindow = (%d, %v)", first, err)
	}
	second, err := client.IncrWindow(context.Background(), "rl:u1", time.Minute)
	if err != nil || second != 2 {
		t.Fatalf("IncrWindow = (%d, %v)", second, err)
	}
	expireSeen := false
	for _, command := range stub.recorded() {
		if command == "EXPIRE rl:u1 60" {
			expireSeen = true
		}
	}
	if !expireSeen {
		t.Fatal("no se fijo la ventana de expiracion")
	}
}

func TestErrorDelServadorSeDevuelve(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen fallo: %v", err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		_, _ = conn.Write([]byte("-ERR fallo simulado\r\n"))
		_ = conn.Close()
	}()
	client, err := Dial(ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatalf("Dial fallo: %v", err)
	}
	if _, err := client.Do("GET", "k1"); err == nil {
		t.Fatal("se esperaba error del servidor")
	}
}
