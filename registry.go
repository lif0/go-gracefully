package gracefully

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/lif0/pkg/utils/errx"
)

// Registry is a thread-safe registry for instances which should be can graceful shutdown.
//
// Use NewRegister to create a new instance.
type Registry struct {
	mu        sync.Mutex
	instances map[unsafe.Pointer]GracefulShutdownObject

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
		mu:        sync.Mutex{},
		instances: make(map[unsafe.Pointer]GracefulShutdownObject),
		disposed:  atomic.Bool{},
		chsd:      make(chan struct{}),
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
	if ei, ok := r.instances[ptr]; ok {
		return &AlreadyRegisteredError{
			ExistingInstance: ei,
			NewInstance:      igs,
		}
	}

	r.instances[ptr] = igs
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
	if _, ok := r.instances[ptr]; !ok {
		return false
	}

	delete(r.instances, ptr)
	return true
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
		return errx.MultiError{ErrAllInstanceShutdownAlready}
	}

	errs := errx.MultiError{}
	for _, v := range r.instances {
		gsErr := v.GracefulShutdown(ctx)
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
		return ErrAllInstanceShutdownAlready
	}
	return nil
}
