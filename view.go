package view

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type View interface {
	Execute(w io.Writer, data interface{}) error
}

type Manager struct {
	m       *sync.RWMutex
	buffers *bufferPool
	views   map[string]View
}

func New(bufferPoolSize int, views map[string]View) *Manager {
	m := &Manager{
		m:       &sync.RWMutex{},
		buffers: newBufferPool(bufferPoolSize),
		views:   make(map[string]View),
	}
	if views != nil {
		for name, v := range views {
			if v == nil {
				panic(fmt.Errorf("view: View \"%s\" is nil", name))
			}
			m.views[name] = v
		}
	}
	return m
}

func (m *Manager) MustRegister(name string, v View) *Manager {
	if v == nil {
		panic(fmt.Errorf("view: View \"%s\" is nil", name))
	}
	m.m.Lock()
	m.views[name] = v
	m.m.Unlock()
	return m
}

func (m *Manager) Render(name string, w io.Writer, data interface{}) error {
	m.m.RLock()
	v, ok := m.views[name]
	m.m.RUnlock()
	if !ok {
		return fmt.Errorf("view: View \"%s\" doesn't exist", name)
	}
	b := m.buffers.Get()
	// trade-off:
	// when Render causes a panic, the buffer will not be reused
	// but the runtime overhead of defer is avoided
	err := v.Execute(b, data)
	if err != nil {
		m.buffers.Put(b)
		return err
	}
	_, err = b.WriteTo(w)
	m.buffers.Put(b)
	return err
}

type bufferPool struct{ c chan *bytes.Buffer }

func newBufferPool(size int) (bp *bufferPool) { return &bufferPool{c: make(chan *bytes.Buffer, size)} }

func (bp *bufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.c: // reuse existing buffer
	default:
		b = bytes.NewBuffer([]byte{})
	}
	return
}

func (bp *bufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	select {
	case bp.c <- b:
	default: // Discard the buffer if the pool is full.
	}
}
