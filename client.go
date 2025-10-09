package gogracefully

import (
	"context"
	"sync"
	"unsafe"
)

// GracefulShutdown is an interface that defines the contract for objects
// capable of performing a graceful shutdown.
// Implementations must provide a name and a shutdown method that respects
// the given context and returns an error if shutdown fails.
type GracefulShutdown interface {
	GracefulShutdownName() string
	GracefulShutdown(context.Context) error
}

// GracefullyRegister is a thread-safe registry for GracefulShutdown instances.
// Use NewGracefullyRegister to create a new instance.
type GracefullyRegister struct {
	mu        sync.Mutex
	instances map[unsafe.Pointer]GracefulShutdown
}

// NewGracefullyRegister creates and returns a new initialized GracefullyRegister.
// If you want to set new GracefullyRegister as Global, use gogracefully.SetGlobal().
func NewGracefullyRegister() *GracefullyRegister {
	return &GracefullyRegister{
		instances: make(map[unsafe.Pointer]GracefulShutdown),
	}
}
