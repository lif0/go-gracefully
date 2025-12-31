package gracefully

import (
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/lif0/pkg/concurrency/chanx"
	"github.com/lif0/pkg/utils/errx"
)

var (
	status   atomic.Uint32
	statusCh atomic.Pointer[chan struct{}]
)

func init() { setStatus(StatusRunning) }

// GetStatus returns the current service status during graceful shutdown.
//
// It is safe for concurrent use and reflects the latest recorded state.
func GetStatus() Status { return Status(status.Load()) }

func setStatus(nS Status) {
	status.Store(uint32(nS))

	if statusCh.Load() != nil {
		select {
		case *statusCh.Load() <- struct{}{}:
		default:
		}
	}
}

// WatchStatus subscribes to status changes.
//
// When the status changes, all provided callback functions are invoked.
// Each callback receives the new status value as an argument.
func WatchStatus(ctx context.Context, callbacks ...func(newStatus Status)) {
	if statusCh.Load() == nil {
		ch := make(chan struct{}, 3)
		statusCh.CompareAndSwap(nil, &ch)
	}

	go func() {
		lastStatus := status.Load()
		e := *statusCh.Load()

		for {
			select {
			case <-ctx.Done():
				close(*statusCh.Load())
				statusCh.Store(nil)
				return
			case <-e:
				if status.Load() != lastStatus {
					lastStatus = status.Load()
					newStatus := GetStatus()
					for i := range callbacks {
						callbacks[i](newStatus)
					}
				}
			}
		}
	}()
}

// SetShutdownTrigger sets up a trigger for Registry.Shutdown.
//
// This global function takes a context for cancellation; if the context is canceled,
// the trigger will not activate, and the function will simply return.
// It accepts options to specify signals or channels that will trigger the shutdown.
func SetShutdownTrigger(ctx context.Context, opts ...TriggerOption) {
	c := newDefaultTriggerConfig()
	for _, opt := range opts {
		opt(c)
	}

	go func() {
		var once sync.Once // ensures graceful Shutdown is attempted only once
		var firstSignal bool = false
		singleUserChan := chanx.FanIn(ctx, c.usrch...)

		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-c.sysch:
				log.Printf("gogracefully: Received system signal - %s\n", sig.String())
			case <-singleUserChan:
				log.Printf("gogracefully: Received user trigger\n")
			}

			setStatus(StatusDraining)

			firstSignal = !firstSignal

			if firstSignal {
				go func() { // because we should be have can handle second signal.
					once.Do(func() {
						defer setStatus(StatusStopped)

						shutdownCtx := ctx
						if c.timeout > 0 {
							sctx, cancel := context.WithTimeout(ctx, c.timeout)
							shutdownCtx = sctx
							defer cancel()
						}

						// log.Printf("gogracefully: Starting graceful shutdown with timeout\n")
						if muErr := defaultRegistry.Shutdown(shutdownCtx); muErr != nil && !muErr.IsEmpty() {
							globalErrors.MutateValue(func(v *errx.MultiError) {
								v.Append(muErr)
							})
						}
						log.Printf("gogracefully: Graceful shutdown completed. Use gogracefully.GlobalErrors for checks errors\n")
					})
				}()
			} else {
				// Second or subsequent signal: Force exit
				log.Printf("gogracefully: Received additional signal - forcing exit\n")
				os.Exit(1) // Or os.Exit(130) for SIGINT, etc.
			}
		}
	}()
}
