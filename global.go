package gogracefully

var globalReg = NewGracefullyRegister()

// SetGlobal sets a custom GracefullyRegister as the global registry.
// This allows replacing the default registry with a user-provided one.
func SetGlobal(gr *GracefullyRegister) {
	globalReg = gr
}
