package gracefully

import "errors"

// ErrShutdownCalled is returned when Register is invoked after Shutdown.
// Use errors.Is(err, ErrShutdownCalled) to detect this case.
var ErrShutdownCalled = errors.New("registration disabled after Shutdown")

// ErrAlreadyRegistered is returned when trying to register the same instance twice
// (or another instance with the same identity). Use errors.Is(err, ErrAlreadyRegistered).
var ErrAlreadyRegistered = errors.New("instance already registered")
