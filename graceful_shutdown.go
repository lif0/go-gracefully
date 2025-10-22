package gracefully

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/lif0/go-gracefully/internal"
	"github.com/lif0/pkg/concurrency/chanx"
	"github.com/lif0/pkg/utils/errx"
)

var status = internal.NewSyncObject(StatusRunning)

// GetStatus returns the current service status during graceful shutdown.
//
// It is safe for concurrent use and reflects the latest recorded state.
func GetStatus() Status {
	return *status.GetObject()
}

func setStatus(nS Status) {
	status.Mutate(func(v *Status) {
		*v = nS
	})
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
							globalErrors.Mutate(func(v *errx.MultiError) {
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
