package realtime

import "sync"

// Registry agrupa conexiones activas por fingerprint para el fan-out.
type Registry struct {
	mu    sync.RWMutex
	conns map[string]map[*client]struct{}
}

func newRegistry() *Registry {
	return &Registry{conns: make(map[string]map[*client]struct{})}
}

func (reg *Registry) add(fingerprint string, c *client) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	set, ok := reg.conns[fingerprint]
	if !ok {
		set = make(map[*client]struct{})
		reg.conns[fingerprint] = set
	}
	set[c] = struct{}{}
}

func (reg *Registry) remove(fingerprint string, c *client) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	set, ok := reg.conns[fingerprint]
	if !ok {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(reg.conns, fingerprint)
	}
}

// signal envia el payload a todas las conexiones del fingerprint.
// La copia bajo RLock evita que un cliente lento frene el fan-out.
func (reg *Registry) signal(fingerprint string, payload []byte) {
	reg.mu.RLock()
	targets := make([]*client, 0, len(reg.conns[fingerprint]))
	for c := range reg.conns[fingerprint] {
		targets = append(targets, c)
	}
	reg.mu.RUnlock()
	for _, c := range targets {
		c.sendSignal(payload)
	}
}

// client es una conexion WS registrada de un usuario.
type client struct {
	ws          *wsConn
	fingerprint string
}

func (c *client) sendSignal(payload []byte) {
	if err := c.ws.writeMessage(opText, payload); err != nil {
		_ = c.ws.Close()
	}
}
