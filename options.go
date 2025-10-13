package gracefully

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TriggerConfig represents the configuration for shutdown triggers.
type TriggerConfig struct {
	sysch <-chan os.Signal
	usrch []<-chan struct{}

	timeout time.Duration
}

type TriggerOption func(*TriggerConfig)

// WithCustomSystemSignal adds a custom OS signal channel
//
// Example:
//
//		ch := make(chan os.Signal, 1)
//		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, ...other signals)
//	 	gogracefully.SetShutdownTrigger(ctx, WithCustomSystemSignal(sysCh))
func WithCustomSystemSignal(ch chan os.Signal) TriggerOption {
	return func(c *TriggerConfig) {
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
	return func(c *TriggerConfig) {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		c.sysch = ch
	}
}

// WithUserChanSignal adds custom user channels that will trigger graceful shutdown
// when closed. Useful for custom shutdown conditions beyond OS signals.
func WithUserChanSignal(uch ...<-chan struct{}) TriggerOption {
	return func(c *TriggerConfig) {
		c.usrch = uch
	}
}

// WithTimeout sets the shutdown timeout for the graceful shutdown process.
// If not specified, the default timeout is 30 seconds.
//
// Example:
//
//	WithTimeout(45 * time.Second)
func WithTimeout(timeout time.Duration) TriggerOption {
	return func(c *TriggerConfig) {
		c.timeout = timeout
	}
}

// newDefaultTriggerConfig create default config
func newDefaultTriggerConfig() *TriggerConfig {
	config := &TriggerConfig{}
	WithSysSignal()(config)
	WithTimeout(15 * time.Minute)(config)

	return config
}
