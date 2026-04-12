package object

import "sync"

type Mutex struct {
	mu sync.Mutex
}

func (m *Mutex) Type() ObjectType {
	return "MUTEX"
}

func (m *Mutex) Inspect() string {
	return "<mutex>"
}

func (m *Mutex) Lock() {
	m.mu.Lock()
}

func (m *Mutex) Unlock() {
	m.mu.Unlock()
}
