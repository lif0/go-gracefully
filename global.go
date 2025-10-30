package gracefully

import (
	"context"

	"github.com/lif0/pkg/concurrency"
	"github.com/lif0/pkg/utils/errx"
)

var (
	defaultRegistry = NewRegistry()
	globalErrors    = concurrency.NewSyncValue(errx.MultiError{})

	DefaultRegisterer Registerer = defaultRegistry
)

func GlobalError() errx.MultiError {
	var errs errx.MultiError

	globalErrors.ReadValue(func(v *errx.MultiError) {
		errs = append([]error(nil), *v...) // make copy
	})

	return errs
}

// SetGlobal sets a custom GracefullyRegister as the global registry.
// This allows replacing the default registry with a user-provided one
// (e.g. for testing purposes).
func SetGlobal(gr *Registry) {
	defaultRegistry = gr
	DefaultRegisterer = defaultRegistry
}

// Register registers the provided GracefulShutdownObject with the DefaultRegisterer.
//
// Register is a shortcut for DefaultRegisterer.Register(c).
func Register(igs GracefulShutdownObject) error {
	return DefaultRegisterer.Register(igs)
}

// RegisterFunc registers the provided func with the DefaultRegisterer.
//
// Register is a shortcut for DefaultRegisterer.RegisterFunc(f).
func RegisterFunc(f func(context.Context) error) error {
	return DefaultRegisterer.RegisterFunc(f)
}

// Unregister removes the registration of the provided GracefulShutdownObject from the
// DefaultRegisterer.
//
// Unregister is a shortcut for DefaultRegisterer.Unregister(c).
func Unregister(igs GracefulShutdownObject) bool {
	return DefaultRegisterer.Unregister(igs)
}

// MustRegister registers the provided GracefulShutdownObject with the DefaultRegisterer and
// panics if any error occurs.
//
// MustRegister is a shortcut for DefaultRegisterer.MustRegister(cs...).
func MustRegister(igss ...GracefulShutdownObject) {
	DefaultRegisterer.MustRegister(igss...)
}

// WaitShutdown blocks the calling goroutine until with the DefaultRegisterer
// has finished shutdown all registered instances
//
// WaitShutdown is a shortcut for DefaultRegisterer.WaitShutdown().
func WaitShutdown() {
	DefaultRegisterer.WaitShutdown()
}
