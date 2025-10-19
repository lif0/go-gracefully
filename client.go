package gracefully

import (
	"context"

	"github.com/lif0/pkg/utils/errx"
)

// Registerer is the interface for the part of a registry in charge of
// registering and unregistering. Users of custom registries should use
// Registerer as type for registration purposes (rather than the Registry type
// directly). In that way, they are free to use custom Registerer implementation
// (e.g. for testing purposes).
type Registerer interface {
	// Register registers a new GracefulShutdownObject to be included in graceful
	// shutdown operations. It returns an error if the provided service is invalid
	// or if it — in combination with already registered services — violates uniqueness
	// criteria (e.g., duplicate pointers).
	//
	// Important:
	//   - instance.GracefulShutdown() are executed in the exact order they were registered.
	//
	// If the provided GracefulShutdownObject is equal to a service already registered
	// (which includes the case of re-registering the same service), the
	// returned error is ErrAlreadyRegistered, which
	// contains the previously registered service.
	Register(GracefulShutdownObject) error
	// RegisterFunc registers a shutdown callback (func(context.Context) error) to be
	// executed during graceful shutdown.
	//
	// Important:
	//   - Callbacks registered via RegisterFunc CANNOT be removed (there is no
	//     deregistration for functions). Deregistration is only supported for
	//     services added via Register.
	//   - Callbacks are executed in the exact order they were registered.
	//
	// If the provided function is nil, an error is returned. Registering the same
	// logical function multiple times is allowed; each registration is treated as a
	// separate callback and will be invoked separately.
	RegisterFunc(func(context.Context) error) error
	// MustRegister works like Register but registers any number of
	// GracefulShutdownObjects and panics upon the first registration that causes an
	// error.
	MustRegister(...GracefulShutdownObject)
	// Unregister unregisters the GracefulShutdownService that equals the service passed
	// in as an argument. (Two services are considered equal if their pointers match)
	// The function returns whether a service was unregistered. Note that an unchecked
	// service (with empty name) cannot be unregistered reliably.
	//
	// Note that even after unregistering, it will not be possible to
	// register a new service that conflicts with the unregistered
	// service (e.g., a service with the same name but different shutdown logic).
	// The rationale here is that the same registry instance must only manage
	// consistent services throughout its lifetime.
	Unregister(GracefulShutdownObject) bool
	// Shutdown shuts down all registered GracefulShutdownObject instances synchronously and in sequence.
	// It uses the provided context for shutdown operations and collects errors,
	// keyed by the instance names returned from GracefulShutdownName.
	// Returns a map of errors; empty if all shutdowns succeeded.
	Shutdown(ctx context.Context) errx.MultiError
	// WaitShutdown blocks the calling goroutine until Registry has finished
	// shutdown all registered instances.
	WaitShutdown()
}

// GracefulShutdownObject is an interface that defines the contract for objects
// capable of performing a graceful shutdown.
// Implementations must provide a shutdown method that respects
// the given context and returns an error if shutdown fails.
type GracefulShutdownObject interface {
	GracefulShutdown(context.Context) error
}
