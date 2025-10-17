package gracefully_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lif0/go-gracefully"
	"github.com/stretchr/testify/assert"
)

func Test_ErrAllInstanceShutdownAlready(t *testing.T) {
	t.Parallel()

	t.Run("ok/message", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, gracefully.ErrAllInstanceShutdownAlready, "sentinel must be non-nil")
		assert.Equal(t,
			"after calling Registerer.Shutdown, you cannot register anymore",
			gracefully.ErrAllInstanceShutdownAlready.Error(),
		)
	})

	t.Run("ok/wrap_is", func(t *testing.T) {
		t.Parallel()
		wrapped := fmt.Errorf("wrap: %w", gracefully.ErrAllInstanceShutdownAlready)
		assert.True(t, errors.Is(wrapped, gracefully.ErrAllInstanceShutdownAlready))
	})

	t.Run("edge/as_no_match", func(t *testing.T) {
		t.Parallel()
		var are *gracefully.AlreadyRegisteredError
		wrapped := fmt.Errorf("wrap: %w", gracefully.ErrAllInstanceShutdownAlready)
		var out *gracefully.AlreadyRegisteredError
		assert.False(t, errors.As(wrapped, &out), "should not match different error type")
		assert.Nil(t, out)
		// also check that comparing directly is false
		assert.NotEqual(t, error(are), gracefully.ErrAllInstanceShutdownAlready)
	})
}

func Test_AlreadyRegisteredError_Error(t *testing.T) {
	t.Parallel()

	t.Run("ok/basic", func(t *testing.T) {
		t.Parallel()
		e := &gracefully.AlreadyRegisteredError{}
		assert.Equal(t,
			"duplicate graceful shutdown instance registration attempted",
			e.Error(),
		)

		assert.Equal(t, e.Error(), fmt.Sprintf("%s", e))
		assert.Equal(t, e.Error(), fmt.Sprintf("%v", e))
	})

	t.Run("ok/errors_as", func(t *testing.T) {
		t.Parallel()
		e := &gracefully.AlreadyRegisteredError{}
		wrapped := fmt.Errorf("wrap: %w", e)

		var out *gracefully.AlreadyRegisteredError
		ok := errors.As(wrapped, &out)
		assert.True(t, ok)
		assert.Same(t, e, out, "As should retrieve the original pointer")
	})
}
