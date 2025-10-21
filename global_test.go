package gracefully_test

// THESE ARE FORMAL TESTS TO INCREASE THE PROJECT'S TEST COVERAGE PERCENTAGE.

import (
	"context"
	"testing"
	"time"

	"github.com/lif0/go-gracefully"
	"github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	// arrange
	old := gracefully.DefaultRegisterer
	t.Cleanup(func() { gracefully.DefaultRegisterer = old })

	// act
	r := gracefully.NewRegistry()
	gracefully.SetGlobal(r)

	// assert
	assert.Equal(t, r, gracefully.DefaultRegisterer)

	// arrange
	obj := &stubGSO{id: 1}
	userCh := make(chan struct{}, 1)
	gracefully.SetShutdownTrigger(
		context.Background(),
		gracefully.WithUserChanSignal(userCh),
		gracefully.WithTimeout(10),
	)

	// act
	err := gracefully.Register(obj)
	// assert
	assert.NoError(t, err)

	// act
	unreg1 := gracefully.Unregister(obj)
	// assert
	assert.True(t, unreg1, "first unregister should succeed after register")

	// act
	unreg2 := gracefully.Unregister(obj)
	// assert
	assert.False(t, unreg2, "second unregister should report false (idempotent)")

	// act
	err = gracefully.RegisterFunc(func(ctx context.Context) error { return nil })
	// assert
	assert.NoError(t, err)

	// act
	assert.NotPanics(t, func() { gracefully.MustRegister(obj) })
	// assert (nothing extra)

	// act
	go func() {
		time.Sleep(time.Microsecond * 50)
		userCh <- struct{}{}
	}()

	gracefully.WaitShutdown()
	gracefully.WaitShutdown()
	gracefully.WaitShutdown()

	// pkg have bug in empty err it have len(multi_err) == 1
	globalErr := gracefully.GlobalError()
	assert.Len(t, globalErr, 0)
	assert.True(t, globalErr.IsEmpty())

	assert.NoError(t, globalErr.MaybeUnwrap())

	// edge: direct nil global panics for all helpers
	// act
	gracefully.DefaultRegisterer = nil
	// assert
	assert.Panics(t, func() { gracefully.MustRegister(obj) })
	assert.Panics(t, func() { gracefully.WaitShutdown() })
	assert.Panics(t, func() { _ = gracefully.RegisterFunc(func(ctx context.Context) error { return nil }) })
	assert.Panics(t, func() { _ = gracefully.Unregister(obj) })
}
