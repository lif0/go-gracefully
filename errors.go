package gogracefully

import "errors"

var ErrAllInstanceShutdownAlready = errors.New("after calling Registerer.Shutdown, you cannot register anymore")

// AlreadyRegisteredError is returned by the Register method if the GracefulShutdownObject to
// be registered has already been registered before, or a different GracefulShutdownObject
// that shares the same pointer has been registered before. Registration fails
// in that case, but you can detect from the kind of error what has
// happened. The error contains fields for the existing GracefulShutdownObject and the
// (rejected) new GracefulShutdownObject that equals the existing one. This can be used to
// find out if an equal GracefulShutdownObject has been registered before and switch over to
// using the old one, as demonstrated in the example.
type AlreadyRegisteredError struct {
	ExistingInstance, NewInstance GracefulShutdownObject
}

func (e *AlreadyRegisteredError) Error() string {
	return "duplicate graceful shutdown instance registration attempted"
}
