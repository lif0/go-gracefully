package gogracefully

import "reflect"

// Register adds a GracefulShutdown instance to the global registry.
// It uses the instance's unsafe pointer as the key to check for duplicates.
// Returns true if the instance was added successfully, false if it already exists.
func Register(igs GracefulShutdown) bool {
	globalReg.mu.Lock()
	defer globalReg.mu.Unlock()

	ptr := reflect.ValueOf(igs).UnsafePointer()
	if _, ok := globalReg.instances[ptr]; ok {
		return false
	}

	globalReg.instances[ptr] = igs
	return true
}

// Unregister removes a GracefulShutdown instance from the global registry.
// It uses the instance's unsafe pointer to locate and delete it.
// Returns true if the instance was found and removed, false otherwise.
func Unregister(igs GracefulShutdown) bool {
	globalReg.mu.Lock()
	defer globalReg.mu.Unlock()

	ptr := reflect.ValueOf(igs).UnsafePointer()
	if _, ok := globalReg.instances[ptr]; !ok {
		return false
	}

	delete(globalReg.instances, ptr)
	return true
}

// MustRegister registers one or more GracefulShutdown instances to the global registry.
// It panics if any instance is already registered, ensuring no duplicates.
// This is useful for mandatory registrations during initialization.
func MustRegister(igss ...GracefulShutdown) {
	for i := 0; i < len(igss); i++ {
		res := Register(igss[i])
		if !res {
			panic("GracefulShutdown found duplicate graceful shutdown instance")
		}
	}
}

// New creates a new instance of type T using the provided factory function,
// registers it in the global registry via MustRegister, and returns it.
// T must implement the GracefulShutdown interface.
func New[T GracefulShutdown](fNew func() T) T {
	instance := fNew()
	MustRegister(instance)
	return instance
}
