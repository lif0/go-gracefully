// Package gogracefully provides a thread-safe registry for managing graceful shutdown instances.
// It allows registering objects that implement the GracefulShutdownObject interface,
// performing synchronous shutdowns with context support, and handling errors.
// The package includes a global registry, registration functions, and utilities
// for safe instance management in concurrent environments.
package gracefully
