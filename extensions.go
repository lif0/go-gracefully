package gogracefully

// NewInstance creates a new instance of type T using the provided factory function,
// registers it in the global registry via MustRegister, and returns it.
// T must implement the GracefulShutdownObject interface.
func NewInstance[T GracefulShutdownObject](fNew func() T) T {
	instance := fNew()
	MustRegister(instance)
	return instance
}
