package gracefully

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/lif0/pkg/concurrency/chanx"
	"github.com/lif0/pkg/utils/errx"
)

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

			firstSignal = !firstSignal

			if firstSignal {
				once.Do(func() {
					shutdownCtx, cancel := context.WithTimeout(ctx, c.timeout)
					defer cancel()

					// log.Printf("gogracefully: Starting graceful shutdown with timeout\n")
					if muErr := defaultRegistry.Shutdown(shutdownCtx); muErr != nil {
						GlobalErrors.Mutate(func(v *errx.MultiError) {
							v.Append(muErr)
						})
					}
					log.Printf("gogracefully: Graceful shutdown completed. Use gogracefully.GlobalErrors for checks errors\n")
				})
			} else {
				// Second or subsequent signal: Force exit
				log.Printf("gogracefully: Received additional signal - forcing exit\n")
				os.Exit(1) // Or os.Exit(130) for SIGINT, etc.
			}
		}
	}()
}
