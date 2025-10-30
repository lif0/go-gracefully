package gracefully

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/lif0/pkg/utils/errx"
	"github.com/lif0/pkg/utils/structx"
)

type anchor struct{ _ byte } // 1 byte size

// Registry is a thread-safe registry for instances which should be can graceful shutdown.
//
// Use NewRegister to create a new instance.
type Registry struct {
	mu sync.Mutex

	gsiHash     *structx.OrderedMap[unsafe.Pointer, func(context.Context) error]
	gsiFuncAnch []*anchor // wee should save pointer, because GC can remove it

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
		mu: sync.Mutex{},

		gsiHash:     structx.NewOrderedMap[unsafe.Pointer, func(context.Context) error](),
		gsiFuncAnch: make([]*anchor, 0),

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
	if _, ok := r.gsiHash.Get(ptr); ok {
		return ErrAlreadyRegistered
	}

	r.gsiHash.Put(ptr, igs.GracefulShutdown)
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

	anchor := &anchor{}
	ptr := unsafe.Pointer(anchor)

	r.gsiHash.Put(ptr, f)
	r.gsiFuncAnch = append(r.gsiFuncAnch, anchor)

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
	if _, ok := r.gsiHash.Get(ptr); ok {
		structx.Delete(r.gsiHash, ptr)
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
	for _, f := range r.gsiHash.Iter() {
		gsErr := f(ctx)
		if gsErr != nil {
			errs.Append(gsErr)
		}
	}

	// broadcast for all who call WaitShutdown()
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
