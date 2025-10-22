package gracefully_test

import (
	"testing"

	"github.com/lif0/go-gracefully"
	"github.com/stretchr/testify/assert"
)

func Test_Status_String(t *testing.T) {
	t.Parallel()

	t.Run("ok/running", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := gracefully.StatusRunning
		// act
		got := s.String()
		// assert
		assert.Equal(t, "Running", got)
	})

	t.Run("ok/draining", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := gracefully.StatusDraining
		// act
		got := s.String()
		// assert
		assert.Equal(t, "Draining", got)
	})

	t.Run("bug/stopped_string_mismatch", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := gracefully.StatusStopped
		// act
		got := s.String()
		// assert
		assert.Equal(t, "Stopped", got, "String() для StatusStopped должно быть 'Stopped'")
	})

	t.Run("edge/unknown_value", func(t *testing.T) {
		t.Parallel()
		// arrange
		s := gracefully.Status(99)
		// act
		got := s.String()
		// assert
		assert.Equal(t, "Status(99)", got)
	})

	t.Run("edge/zero_value_is_running", func(t *testing.T) {
		t.Parallel()
		// arrange
		var s gracefully.Status
		// act
		gotName := s.String()
		// assert
		assert.Equal(t, gracefully.StatusRunning, s)
		assert.Equal(t, "Running", gotName)
	})
}
