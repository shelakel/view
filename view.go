package view

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// View represents a template, such as html/template or view/template
type View interface {
	Execute(w io.Writer, data interface{}) error
}

// Manager is a simple view manager supporting sandboxed (pooled buffer) renders and runtime view registration/replacement
type Manager struct {
	m       *sync.RWMutex
	buffers *bufferPool
	views   map[string]View
}

var ErrBufferPoolSizeInvalid = fmt.Errorf("view: bufferPoolSize should be greater than zero")

// New creates a new view Manager
//
// - bufferPoolSize should be greater than zero, typically 50 or more.
// - views represent the initial views registered and can be nil
func New(bufferPoolSize int, views map[string]View) *Manager {
	if bufferPoolSize <= 0 {
		panic(ErrBufferPoolSizeInvalid)
	}
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

// MustRegister registers or replaces a view with the name
//
// - v can't be nil
func (m *Manager) MustRegister(name string, v View) *Manager {
	if v == nil {
		panic(fmt.Errorf("view: View \"%s\" is nil", name))
	}
	m.m.Lock()
	m.views[name] = v
	m.m.Unlock()
	return m
}

var ErrWriterRequired = fmt.Errorf("view: w can't be nil")

// Render renders the view to a pooled buffer instance and if no error occurred, writes to w
//
// - calling render with a name of a view that's not registered will return an error
// - w can't be nil
func (m *Manager) Render(name string, w io.Writer, data interface{}) error {
	if w == nil {
		return ErrWriterRequired
	}
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
