package gracefully

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/lif0/pkg/utils/errx"
)

var stubFunc = func(context.Context) error {
	return nil
}

// Registry is a thread-safe registry for instances which should be can graceful shutdown.
//
// Use NewRegister to create a new instance.
type Registry struct {
	mu sync.Mutex

	gsiHash map[unsafe.Pointer]int
	gsFunc  []func(context.Context) error

	// chan shutdown done
	chsd     chan struct{}
	disposed atomic.Bool
}

// NewRegistry creates and returns a new initialized Registerer.
//
// If you want to set new Registerer as Global, use gogracefully.SetGlobal()
// (e.g. for testing purposes).
func NewRegistry() *Registry {
	return &Registry{
		mu:       sync.Mutex{},
		gsiHash:  make(map[unsafe.Pointer]int),
		gsFunc:   make([]func(context.Context) error, 0),
		chsd:     make(chan struct{}),
		disposed: atomic.Bool{},
	}
}

// Register implements Registerer.
func (r *Registry) Register(igs GracefulShutdownObject) error {
	if err := r.isDisposed(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.isDisposed(); err != nil {
		return err
	}

	ptr := reflect.ValueOf(igs).UnsafePointer()
	if _, ok := r.gsiHash[ptr]; ok {
		return ErrAlreadyRegistered
	}

	r.gsFunc = append(r.gsFunc, igs.GracefulShutdown)
	r.gsiHash[ptr] = len(r.gsFunc) - 1
	return nil
}

// RegisterFunc implements Registerer.
func (r *Registry) RegisterFunc(f func(context.Context) error) error {
	if err := r.isDisposed(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.isDisposed(); err != nil {
		return err
	}

	r.gsFunc = append(r.gsFunc, f)
	return nil
}

// Unregister implements Registerer.
func (r *Registry) Unregister(igs GracefulShutdownObject) bool {
	if r.isDisposed() != nil {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isDisposed() != nil {
		return false
	}

	ptr := reflect.ValueOf(igs).UnsafePointer()
	if i, ok := r.gsiHash[ptr]; ok {
		r.gsFunc[i] = stubFunc
		delete(r.gsiHash, ptr)
		return true
	}

	return false
}

// MustRegister implements Registerer.
// It panics if any instance is already registered.
func (r *Registry) MustRegister(igss ...GracefulShutdownObject) {
	for i := 0; i < len(igss); i++ {
		err := r.Register(igss[i])
		if err != nil {
			panic(err)
		}
	}
}

// Shutdown implements Registerer.
func (r *Registry) Shutdown(ctx context.Context) errx.MultiError {
	if err := r.isDisposed(); err != nil {
		return errx.MultiError{err}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.disposed.CompareAndSwap(false, true) {
		return errx.MultiError{ErrShutdownCalled}
	}

	errs := errx.MultiError{}
	for _, f := range r.gsFunc {
		gsErr := f(ctx)
		if gsErr != nil {
			errs.Append(gsErr)
		}
	}

	// broadcast for all who call WaitShutdown
	close(r.chsd)
	return errs
}

// WaitShutdown implements Registerer.
func (r *Registry) WaitShutdown() {
	<-r.chsd
}

// isDisposed ...
func (r *Registry) isDisposed() error {
	if r.disposed.Load() {
		return ErrShutdownCalled
	}
	return nil
}
