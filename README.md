<div align="center">
  <!-- Logo -->
  <img src=".github/assets/logo_round.png" width="256" height="256" alt="go-gracefully logo">

  <!-- Badges -->
  <p align="center">
    <a href="https://github.com/lif0/go-gracefully/workflows/build/badge.svg">
      <img src="https://github.com/lif0/go-gracefully/workflows/build/badge.svg" alt="build">
    </a>
    <a href="https://pkg.go.dev/github.com/lif0/go-gracefully">
      <img src="https://pkg.go.dev/badge/github.com/lif0/go-gracefully.svg" alt="go reference">
    </a>
    <img src="https://img.shields.io/github/v/tag/lif0/go-gracefully?label=latest" alt="latest version">
    <a href="https://coveralls.io/github/lif0/go-gracefully?branch=main">
      <img src="https://coveralls.io/repos/github/lif0/go-gracefully/badge.svg?branch=main" alt="coverage">
    </a>
    <a href="https://goreportcard.com/report/github.com/lif0/go-gracefully">
      <img src="https://goreportcard.com/badge/github.com/lif0/go-gracefully" alt="go report card">
    </a>
  </p>

  <!-- Description -->
  <p align="center">
    Graceful shutdown for Go services: register cleanup hooks and shut down safely on OS signals or custom events - without losing state in your stateful apps/services.
  </p>
</div>

---

## Contents

- [Overview](#-overview)
- [Requirements](#-requirements)
- [Installation](#-installation)
- [Usage Guide](#-usage-guide)
    - [Step 1: GracefulShutdownObject](#step-1-implement-the-gracefulshutdownobject-interface)
    - [Step 2: Register objects](#step-2-register-objects)
    - [Step 3: Set triggers](#step-3-set-up-shutdown-triggers)
    - [Step 4: Handle shutdown](#step-4-handle-shutdown)
    - [Step 5: Unregister](#step-5-unregister-if-needed)
- [Features](#feature)
    - [Create&Register instances](#advanced-create-and-register-instances) 
    - [Error handling](#error-handling)
- [Examples](#-examples)
    - [app-counter](#app-counter)
    - [more](#more)
- [Roadmap](#-roadmap)
- [License](#-license)

---

## 📋 Overview

The `gracefully` package is designed to simplify graceful shutdowns in Go applications. The core concept revolves around a thread-safe registry that manages objects implementing a simple interface for shutdown logic. This allows you to register components like servers, databases, or any resources that require proper cleanup before the program exits.

Key ideas:

- **Registry-based management**: A central registry (global by default) holds references to shutdownable objects. It's safe for concurrent use and prevents duplicate registrations.
- **Trigger-based shutdown**: Shutdown can be triggered by OS signals (e.g., SIGINT, SIGTERM), custom channels, or manually. It respects contexts for timeouts and collects errors from failed shutdowns.
- **Error handling**: Uses a multi-error type to aggregate issues, with global access for post-shutdown checks.
- **Flexibility**: Supports custom registries, priorities (future), and extensions for universal use in web apps, CLI tools, or services.

This approach ensures your application handles interruptions politely, avoiding data corruption or abrupt terminations, especially in production environments like Docker or Kubernetes.

## 📦 Installation

To install the package, run:

```bash
go get github.com/lif0/go-gracefully@latest
```

Import it in your code:

```go
import "github.com/lif0/go-gracefully"
```

## ✨ Usage Guide

### Step 1: Implement the GracefulShutdownObject Interface

Any object that needs graceful shutdown must implement this interface:

```go
type GracefulShutdownObject interface {
    GracefulShutdown(ctx context.Context) error
}
```

Example implementation for a custom batcher:

```go
type MyBatcher struct {
    // Your batcher fields, e.g.
}

func (s *MyBatcher) GracefulShutdown(ctx context.Context) error {
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

### Step 2: Register Objects/Register Func

Use the global registry:

```go
import "github.com/lif0/go-gracefully"

type MyBatcher struct {
  // some fields
}

func (mb *MyBatcher) GracefulShutdown() {...}

func (mb *MyBatcher) Closer() {...}

// order is important
// at first will be called GracefulShutdown()
// at second will be called Closer()

myBatcher := &MyBatcher{}
gracefully.MustRegister(myBatcher) // will be register myBatcher.GracefulShutdown()
gracefully.RegisterFunc(myBatcher.Closer)
```

or

```go
import "github.com/lif0/go-gracefully"

myBatcher := &MyBatcher{}
gracefully.MustRegister(myBatcher)
```

### Step 3: Set Up Shutdown Triggers

Launch a goroutine to listen for triggers.

⚠️ Important: If you do not call this function, no triggers will be registered, and graceful shutdown will not occur automatically (e.g., on signals). You would need to call Shutdown manually in that case. (see [Step 4](#step-4-handle-shutdown))

```go

gracefully.SetShutdownTrigger(context.Background())
```

or

```go
gracefully.SetShutdownTrigger(
    context.Background(),
    gracefully.WithSysSignal(),
    gracefully.WithTimeout(time.Hour)
    )
```

### Custom Options

#### WithSysSignal()

Registers `signal.Notify` for SIGINT and SIGTERM signals on the signal channel. This option is enabled by default.

A repeated signal will invoke `os.Exit(130)`, which immediately terminates the application without waiting for any ongoing processes.

#### WithCustomSystemSignal(ch chan os.Signal)

Provides your own custom signal channel for handling OS signals.

⚠️ Don't close os.Signal chan.

```go
ch := make(chan os.Signal, 1)
signal.Notify(ch, syscall.SIGUSR1 /* or any other signals */)
gracefully.SetShutdownTrigger(ctx, gracefully.WithCustomSystemSignal(ch))
```

#### WithUserChanSignal(uch ...<-chan struct{})

Allows you to pass one or more custom channels. When any of these channels is closed or receives a value, the graceful shutdown process will be triggered.

A repeated signal will invoke `os.Exit(130)`, which immediately terminates the application without waiting for any ongoing processes.

⚠️ Don't close custom chan.

```go
chShutdown := make(chan struct{})
gracefully.SetShutdownTrigger(ctx, gracefully.WithUserChanSignal(chShutdown))

chShutdown <- struct{}{} // to trigger the shutdown
chShutdown <- struct{}{} // to trigger the os.Exit
```

#### WithTimeout(d time.Duration)

Sets the maximum duration for the graceful shutdown. By default, no timeout is applied - the service waits for all tasks to finish. A non-positive timeout disables the shutdown deadline.

### Step 4: Handle Shutdown

The trigger will call `Shutdown` automatically. Manually:

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
errs := gracefully.DefaultRegisterer.Shutdown(shutdownCtx)
if errs != nil {
    // Check gracefully.GlobalErrors for details
}
```

Wait for completion:

```go
gracefully.WaitShutdown()
```

### Step 5: Unregister if Needed

```go
unregistered := gracefully.Unregister(server) // Returns true if removed
```

## Advanced

### Check current Status

Use `gracefully.GetStatus()` for check status.

Statuses:

| Name             | Description                                                                                                        |
| ---------------- | ------------------------------------------------------------------------------------------------------------------ |
| `StatusRunning`  | The service is running and accepting new requests.                                                                 |
| `StatusDraining` | The service is shutting down gracefully; it no longer accepts new requests but continues processing existing ones. |
| `StatusStopped`  | Graceful shutdown has fully finished; all resources are released and the process can safely exit.                  |


```go
import (
	"fmt"

	"github.com/lif0/go-gracefully"
)

func handleServiceStatus() {
	switch gracefully.GetStatus() {
	case gracefully.StatusRunning:
		// ignore
	case gracefully.StatusDraining:
		// For example: return error for any request
	case gracefully.StatusStopped:
		// For Example: log and finish app
	}
}
```

### Create and Register Instances

Use generics for quick creation:

```go
batcher := gracefully.NewInstance(func() *MyBatcher {
    return &MyBatcher{}
})
```

### Error Handling

- Check `gracefully.GlobalErrors` after shutdown.

For full details, see the GoDoc: [pkg.go.dev/github.com/lif0/go-gracefully](https://pkg.go.dev/github.com/lif0/go-gracefully).

## 👩🏻‍🏫 Examples

### App counter

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lif0/go-gracefully"
)

var stopChan chan struct{}

func main() {
	// configure
    // It will be triggered:
    // close(stopChan) OR
    // stopChan<-struct{}{} OR
    // kill {PID} (Ctrl+C in console)

	gracefully.SetShutdownTrigger(
		context.Background(),
		gracefully.WithSysSignal(),
		gracefully.WithUserChanSignal(stopChan),
	)

	counter := NewCounter()

	gracefully.MustRegister(counter)

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			counter.Inc()
			fmt.Printf("counter: %v\n", counter.val)
		}
	}()

    go func() {
        time.Sleep(time.Hour)
        close(stopChan)
    }()

	gracefully.WaitShutdown() // Wait finish stop all registered objects
	fmt.Println("App finish")
}
```

### More
Check out the [examples directory](https://github.com/lif0/go-gracefully/tree/main/example) for complete, runnable demos, including HTTP server shutdown and custom triggers.

## 🗺️ Roadmap

- [x] Add an interface for validating the structure being registered
- [x] Add object registration/deregistration
- [x] Add options for the GracefulShutdown trigger (WithUserChanSignal, WithCustomSystemSignal, WithSysSignal, WithTimeout)
- [x] Add the ability to register functions
- [x] Reach >90% test coverage
- [ ] Write benchmarks
- [ ] (Internal) Improve the deduplication algorithm (add an OrderedMap)
- [x] Add func: gracefully.Status.
- [ ] Add func: gracefully.WatchStatus().


## 📄 License

MIT License. See [LICENSE](LICENSE) for details.