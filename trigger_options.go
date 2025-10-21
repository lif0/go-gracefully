package gracefully

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TriggerConfig represents the configuration for shutdown triggers.
type triggerConfig struct {
	sysch <-chan os.Signal
	usrch []<-chan struct{}

	timeout time.Duration
}

type TriggerOption func(*triggerConfig)

// WithCustomSystemSignal adds a custom OS signal channel
//
// Example:
//
//		ch := make(chan os.Signal, 1)
//		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, ...other signals)
//	 	gogracefully.SetShutdownTrigger(ctx, WithCustomSystemSignal(sysCh))
func WithCustomSystemSignal(ch chan os.Signal) TriggerOption {
	return func(c *triggerConfig) {
		c.sysch = ch
	}
}

// WithSysSignal adds default OS signal handling for graceful shutdown
//
// SIGINT (Signal Interrupt) - Typically sent when user presses Ctrl+C
// SIGTERM (Signal Terminate) - Polite request to terminate the program (e.g., from Docker or Kubernetes).
//
// Example:
//
//	gogracefully.SetShutdownTrigger(ctx, WithSysSignal())
func WithSysSignal() TriggerOption {
	return func(c *triggerConfig) {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		c.sysch = ch
	}
}

// WithUserChanSignal adds custom user channels that will trigger graceful shutdown
// when closed. Useful for custom shutdown conditions beyond OS signals.
func WithUserChanSignal(uch ...<-chan struct{}) TriggerOption {
	return func(c *triggerConfig) {
		c.usrch = uch
	}
}

// WithTimeout sets the maximum duration for the graceful shutdown.
// By default, no timeout is applied - the service waits for all tasks to finish.
// A non-positive timeout disables the shutdown deadline.
//
// Example:
//
//	WithTimeout(5 * time.Minute)
func WithTimeout(timeout time.Duration) TriggerOption {
	return func(c *triggerConfig) {
		c.timeout = timeout
	}
}

// newDefaultTriggerConfig create default config
func newDefaultTriggerConfig() *triggerConfig {
	config := &triggerConfig{}
	WithSysSignal()(config)
	WithTimeout(0)(config)

	return config
}
