package cache

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// newFakeCache arranca un servidor RESP en memoria que responde con la
// funcion indicada; valida la codificacion real de nuestro cliente.
func newFakeCache(t *testing.T, respond func(args []string) string) (*Conn, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go serveFake(conn, respond)
		}
	}()
	client, err := Dial(listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return client, func() {
		_ = client.Close()
		_ = listener.Close()
	}
}

func serveFake(conn net.Conn, respond func(args []string) string) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	for {
		args, err := readFakeCommand(reader)
		if err != nil {
			return
		}
		if _, err := conn.Write([]byte(respond(args))); err != nil {
			return
		}
	}
}

func readFakeCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(strings.TrimRight(line, "\r\n")[1:])
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, count)
	for i := 0; i < count; i++ {
		sizeLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimRight(sizeLine, "\r\n")[1:])
		if err != nil {
			return nil, err
		}
		buf := make([]byte, size+2)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:size]))
	}
	return args, nil
}

func TestPing(t *testing.T) {
	client, stop := newFakeCache(t, func(args []string) string {
		if args[0] == "PING" {
			return "+PONG\r\n"
		}
		return "-ERR comando\r\n"
	})
	defer stop()
	if err := client.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestSetExGetDel(t *testing.T) {
	data := make(map[string]string)
	client, stop := newFakeCache(t, func(args []string) string {
		switch args[0] {
		case "SETEX":
			data[args[1]] = args[3]
			return "+OK\r\n"
		case "GET":
			value, ok := data[args[1]]
			if !ok {
				return "$-1\r\n"
			}
			return "$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n"
		case "DEL":
			delete(data, args[1])
			return ":1\r\n"
		}
		return "-ERR comando\r\n"
	})
	defer stop()
	if err := client.SetEx("k", "v", time.Minute); err != nil {
		t.Fatalf("SetEx: %v", err)
	}
	value, found, err := client.Get("k")
	if err != nil || !found || value != "v" {
		t.Fatalf("Get = (%q, %v, %v)", value, found, err)
	}
	if err := client.Del("k"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	_, found, err = client.Get("k")
	if err != nil || found {
		t.Fatalf("Get tras Del = (%v, %v)", found, err)
	}
}

func TestIncrYExpire(t *testing.T) {
	counters := make(map[string]int64)
	client, stop := newFakeCache(t, func(args []string) string {
		switch args[0] {
		case "INCR":
			counters[args[1]]++
			return ":" + strconv.FormatInt(counters[args[1]], 10) + "\r\n"
		case "EXPIRE":
			return ":1\r\n"
		}
		return "-ERR comando\r\n"
	})
	defer stop()
	first, err := client.Incr("rl")
	if err != nil || first != 1 {
		t.Fatalf("Incr = (%d, %v)", first, err)
	}
	second, err := client.Incr("rl")
	if err != nil || second != 2 {
		t.Fatalf("Incr = (%d, %v)", second, err)
	}
	if err := client.Expire("rl", time.Minute); err != nil {
		t.Fatalf("Expire: %v", err)
	}
}

func TestErrorDelServidor(t *testing.T) {
	client, stop := newFakeCache(t, func(args []string) string {
		return "-ERR no disponible\r\n"
	})
	defer stop()
	if err := client.Ping(); err == nil {
		t.Fatal("se esperaba error del servidor")
	}
}

func TestRespuestaFueraDeProtocolo(t *testing.T) {
	client, stop := newFakeCache(t, func(args []string) string {
		return "basura\r\n"
	})
	defer stop()
	err := client.Ping()
	if !errors.Is(err, ErrProtocol) {
		t.Fatalf("error = %v, se esperaba ErrProtocol", err)
	}
}
