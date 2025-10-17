package gracefully

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_WithCustomSystemSignal(t *testing.T) {
	t.Parallel()

	t.Run("ok/assigns_same_channel", func(t *testing.T) {
		t.Parallel()
		// arrange
		cfg := &triggerConfig{}
		ch := make(chan os.Signal, 1)

		// act
		WithCustomSystemSignal(ch)(cfg)

		// assert
		assert.NotNil(t, cfg)
		assert.NotNil(t, cfg.sysch)
	})

	t.Run("edge/nil", func(t *testing.T) {
		t.Parallel()
		// arrange
		cfg := &triggerConfig{}

		// act
		WithCustomSystemSignal(nil)(cfg)

		// assert
		assert.Nil(t, cfg.sysch)
	})
}

func Test_WithSysSignal(t *testing.T) {
	t.Run("ok/creates_buffered_channel_and_registers", func(t *testing.T) {
		// arrange
		cfg := &triggerConfig{}

		// act
		WithSysSignal()(cfg)

		// assert
		assert.NotNil(t, cfg.sysch)
	})
}

func Test_WithUserChanSignal(t *testing.T) {
	t.Parallel()

	t.Run("ok/multiple_channels_preserved", func(t *testing.T) {
		t.Parallel()
		// arrange
		cfg := &triggerConfig{}
		a := make(chan struct{})
		b := make(chan struct{})
		c := make(chan struct{})

		// act
		WithUserChanSignal(a, b, c)(cfg)

		// assert
		assert.Len(t, cfg.usrch, 3)
	})

	t.Run("edge/zero_args_nil_slice", func(t *testing.T) {
		t.Parallel()
		// arrange
		cfg := &triggerConfig{}

		// act
		WithUserChanSignal()(cfg)

		// assert
		assert.Nil(t, cfg.usrch)
	})
}

func Test_WithTimeout(t *testing.T) {
	t.Parallel()

	t.Run("ok/assigns_value", func(t *testing.T) {
		t.Parallel()
		// arrange
		cfg := &triggerConfig{}
		d := 123 * time.Millisecond

		// act
		WithTimeout(d)(cfg)

		// assert
		assert.Equal(t, d, cfg.timeout)
	})

	t.Run("edge/zero_duration", func(t *testing.T) {
		t.Parallel()
		// arrange
		cfg := &triggerConfig{}

		// act
		WithTimeout(0)(cfg)

		// assert
		assert.Equal(t, time.Duration(0), cfg.timeout)
	})
}

func Test_newDefaultTriggerConfig(t *testing.T) {
	// signal.Notify touches global process state; avoid parallel here
	t.Run("ok/defaults", func(t *testing.T) {
		// arrange

		// act
		cfg := newDefaultTriggerConfig()

		// assert
		assert.NotNil(t, cfg.sysch)
		assert.Equal(t, 15*time.Minute, cfg.timeout)
	})

	t.Run("api/independent_instances", func(t *testing.T) {
		// arrange

		// act
		a := newDefaultTriggerConfig()
		b := newDefaultTriggerConfig()

		// assert
		assert.NotSame(t, a, b)
		assert.NotNil(t, a.sysch)
		assert.NotNil(t, b.sysch)
	})
}
