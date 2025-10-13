# go-gracefully

[![build](https://github.com/lif0/go-gracefully/workflows/build/badge.svg)](https://github.com/lif0/go-gracefully/workflows/build/badge.svg)
[![go reference](https://pkg.go.dev/badge/github.com/lif0/go-gracefully.svg)](https://pkg.go.dev/github.com/lif0/go-gracefully)
![last version](https://img.shields.io/github/v/tag/lif0/go-gracefully?label=latest)
[![coverage](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Flif0%2Fgo-gracefully%2Frefs%2Fheads%2Fmain%2F.github%2Fassets%2Fbadges%2Fcoverage.json)](https://img.shields.io/endpoint?url=https%3A%2F%2Fraw.githubusercontent.com%2Flif0%2Fgo-gracefully%2Frefs%2Fheads%2Fmain%2F.github%2Fassets%2Fbadges%2Fcoverage.json)
[![report card](https://goreportcard.com/badge/github.com/lif0/go-gracefully)](https://goreportcard.com/report/github.com/lif0/go-gracefully)

Graceful shutdown utility for Golang.

## Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Installation](#installation)
- [Features](#features)
  - [GracefulShutdownObject Interface](#gracefulshutdownobject-interface)
  - [Registry and NewRegistry](#registry-and-newregistry)
  - [Register, Unregister, MustRegister](#register-unregister-mustregister)
  - [Shutdown and WaitShutdown](#shutdown-and-waitshutdown)
  - [Global Functions](#global-functions)
  - [SetShutdownTrigger and Trigger Options](#setshutdowntrigger-and-trigger-options)
  - [NewInstance](#newinstance)
  - [Errors](#errors)
- [Roadmap](#roadmap)
- [License](#license)

---

## 📋 Overview

Package `gracefully` provides a thread-safe registry for managing graceful shutdown instances in Go applications. It allows registering objects that implement the `GracefulShutdownObject` interface, performing synchronous shutdowns with context support, and handling errors. The package includes a global registry, registration functions, and utilities for safe instance management in concurrent environments.

This utility is designed for universal use, such as in web servers, CLI tools, or any long-running Go programs that need to handle shutdown signals (e.g., SIGINT, SIGTERM) gracefully, ensuring resources like databases, servers, or connections are closed properly.

## ⚙️ Requirements

- Go 1.19 or higher

## 📦 Installation

To install the package, run:

```bash
go get github.com/lif0/go-gracefully@latest
```

Import it in your code:

```go
import "github.com/lif0/go-gracefully"
```

## ✨ Features

### GracefulShutdownObject Interface

The core interface that any object must implement to be registered for graceful shutdown.

```go
type GracefulShutdownObject interface {
    GracefulShutdown(context.Context) error
}
```

**Description:**  
Implement this interface for any struct that needs to perform cleanup during shutdown. The method should respect the provided context (e.g., for timeouts) and return an error if shutdown fails.

**Example:**

```go
type Batcher struct {
    // server fields
}

func (s *Batcher) GracefulShutdown(ctx context.Context) error {
    // Flush data to disk,db, etc.
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        s.StopRecv()
        return s.FlushData()
    }
}
```

### Registry and NewRegistry

`Registry` is a thread-safe registry for instances that support graceful shutdown.

```go
func NewRegistry() *Registry
```

**Description:**  
Creates a new `Registry` instance. Use this for custom registries (e.g., in tests). The registry manages registrations, unregistrations, and shutdowns.

**Example:**

```go
reg := gracefully.NewRegistry()
```

### Register, Unregister, MustRegister

Methods on `Registry` (and `Registerer` interface) for managing instances.

- `Register(igs GracefulShutdownObject) error`
- `Unregister(igs GracefulShutdownObject) bool`
- `MustRegister(igss ...GracefulShutdownObject)`

**Description:**  

- `Register`: Adds an instance to the registry. Returns an error if already registered (e.g., `AlreadyRegisteredError`).
- `Unregister`: Removes an instance if it exists.
- `MustRegister`: Registers multiple instances and panics on error.

**Example:**

```go
dataBatcher := &Batcher{}
err := reg.Register(dataBatcher)
if err != nil {
    // Handle error
}

reg.MustRegister(anotherInstance)

unregistered := reg.Unregister(dataBatcher) // true if removed
```

### Shutdown and WaitShutdown

Methods to initiate and wait for shutdown.

- `Shutdown(ctx context.Context) errx.MultiError`
- `WaitShutdown()`

**Description:**  

- `Shutdown`: Shuts down all registered instances in sequence, collecting errors in a `MultiError`. Marks the registry as disposed.
- `WaitShutdown`: Blocks until shutdown is complete (via a channel).

**Example:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

errs := reg.Shutdown(ctx)
if errs != nil {
    // Handle errors
}

reg.WaitShutdown() // Block until done
```

### Global Functions

Convenience wrappers using the default global registry.

- `Register(igs GracefulShutdownObject) error`
- `Unregister(igs GracefulShutdownObject) bool`
- `MustRegister(igss ...GracefulShutdownObject)`
- `WaitShutdown()`
- `SetGlobal(gr *Registry)`

**Description:**  
These operate on the default registry. `SetGlobal` allows replacing the default for custom use (e.g., testing).

**Example:**

```go
gracefully.MustRegister(server)

gracefully.WaitShutdown()
```

### SetShutdownTrigger and Trigger Options

Sets up triggers for automatic shutdown.

```go
func SetShutdownTrigger(ctx context.Context, opts ...TriggerOption)
```

Options:

- `WithSysSignal()`: Default signals (SIGINT, SIGTERM).
- `WithCustomSystemSignal(ch chan os.Signal)`: Custom signal channel.
- `WithUserChanSignal(uch ...<-chan struct{})`: User-defined channels.
- `WithTimeout(timeout time.Duration)`: Shutdown timeout (default 30s).

**Description:**  
Runs in a goroutine, listening for signals or channels to trigger `Shutdown`. On second signal, forces exit with `os.Exit(1)`.

**Example:**

```go
userCh := make(chan struct{})

gracefully.SetShutdownTrigger(context.Background(),
    gracefully.WithSysSignal(),
    gracefully.WithUserChanSignal(userCh),
    gracefully.WithTimeout(45*time.Second),
)

// Later, close userCh to trigger shutdown
close(userCh)
```

### NewInstance

Creates and registers an instance generically.

```go
func NewInstance[T GracefulShutdownObject](fNew func() T) T
```

**Description:**  
Uses a factory function to create a new instance of type `T` (must implement `GracefulShutdownObject`) and registers it globally via `MustRegister`.

**Example:**

```go
server := gracefully.NewInstance(func() *MyServer {
    return &MyServer{}
})
```

### Errors

- `ErrAllInstanceShutdownAlready`: Cannot register after shutdown.
- `AlreadyRegisteredError`: Duplicate registration attempted.

**Description:**  
Errors are handled via `errx.MultiError` for multiple shutdown failures. Global errors are stored in `GlobalErrors`.

**Example:**

```go
if err, ok := err.(*gracefully.AlreadyRegisteredError); ok {
    // Use err.ExistingInstance
}
```

## 🗺️ Roadmap

- [ ] Add priority-based shutdown ordering.
- [ ] Support for asynchronous shutdowns.
- [ ] Integration with popular libraries (e.g., net/http.Server wrappers).
- [ ] Enhanced logging and metrics.
- [ ] Comprehensive unit and benchmarks.

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.