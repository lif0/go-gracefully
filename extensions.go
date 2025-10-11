package gogracefully

import "unsafe"

// NewInstance creates a new instance of type T using the provided factory function,
// registers it in the global registry via MustRegister, and returns it.
// T must implement the GracefulShutdownObject interface.
func NewInstance[T GracefulShutdownObject](fNew func() T) T {
	instance := fNew()
	MustRegister(instance)
	return instance
}

// NewRegistry creates and returns a new initialized Registerer.
//
// If you want to set new Registerer as Global, use gogracefully.SetGlobal()
// (e.g. for testing purposes).
func NewRegistry() *Registry {
	return &Registry{
		instances: make(map[unsafe.Pointer]GracefulShutdownObject),
	}
}
