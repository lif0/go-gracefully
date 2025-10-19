package gracefully_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lif0/go-gracefully"
	"github.com/lif0/pkg/utils/errx"
	"github.com/stretchr/testify/assert"
)

type fakeGSO struct {
	calls int32
	ret   error
	hook  func()
}

func (f *fakeGSO) GracefulShutdown(ctx context.Context) error {
	if f.hook != nil {
		f.hook()
	}
	atomic.AddInt32(&f.calls, 1)
	return f.ret
}

type fakeService struct {
	calls int32
	ret   error
	hook  func()
}

func (f *fakeService) Close(ctx context.Context) error {
	if f.hook != nil {
		f.hook()
	}
	atomic.AddInt32(&f.calls, 1)
	return f.ret
}

func assertMultiErrorContains(t *testing.T, me errx.MultiError, target error) {
	t.Helper()

	found := false
	for _, e := range me {
		if errors.Is(e, target) {
			found = true
			break
		}
	}
	assert.Truef(t, found, "MultiError does not contain target error: %v", target)
}

func firstErr(me errx.MultiError) error {
	return me[0]
}

func Test_NewRegistry(t *testing.T) {
	t.Parallel()

	t.Run("ok/basicConstructs", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		// act
		// assert: WaitShutdown должен блокироваться пока не вызван Shutdown
		blocked := make(chan struct{})
		go func() {
			r.WaitShutdown()
			close(blocked)
		}()
		select {
		case <-blocked:
			t.Fatalf("WaitShutdown must block before Shutdown")
		case <-time.After(80 * time.Millisecond):
			// ok
		}
		_ = r.Shutdown(context.Background())
		select {
		case <-blocked:
			// ok, разблокировался
		case <-time.After(150 * time.Millisecond):
			t.Fatalf("WaitShutdown did not unblock after Shutdown")
		}
	})
}

func Test_Register(t *testing.T) {
	t.Parallel()

	t.Run("ok/basic", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeGSO{}
		// act
		err := r.Register(f)
		// assert
		assert.NoError(t, err)
	})

	t.Run("err/duplicate", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeGSO{}
		err := r.Register(f)
		assert.NoError(t, err)
		// act
		err = r.Register(f)
		// assert
		assert.ErrorIs(t, err, gracefully.ErrAlreadyRegistered)
	})

	t.Run("race/concurrentSameInstance", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeGSO{}
		const n = 32
		errs := make(chan error, n)
		var wg sync.WaitGroup
		wg.Add(n)

		// act
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				errs <- r.Register(f)
			}()
		}
		wg.Wait()
		close(errs)

		// assert
		var nils, dups int
		for e := range errs {
			if e == nil {
				nils++
				continue
			}
			assert.ErrorIs(t, e, gracefully.ErrAlreadyRegistered)
			dups++
		}
		assert.Equal(t, 1, nils)
		assert.Equal(t, n-1, dups)

		ok := r.Unregister(f)
		assert.True(t, ok)
		assert.False(t, r.Unregister(f))
	})

	t.Run("err/afterShutdown", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		_ = r.Shutdown(context.Background())
		// act
		err := r.Register(&fakeGSO{})
		// assert
		assert.ErrorIs(t, firstErr(r.Shutdown(context.Background())), gracefully.ErrShutdownCalled)
		assert.ErrorIs(t, err, gracefully.ErrShutdownCalled)
	})
}

func Test_RegisterFunc(t *testing.T) {
	t.Parallel()

	t.Run("ok/basic", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeService{}
		// act
		err := r.RegisterFunc(f.Close)
		// assert
		assert.NoError(t, err)
	})

	t.Run("err/duplicate", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeService{}
		err := r.RegisterFunc(f.Close)
		assert.NoError(t, err)
		// act
		err = r.RegisterFunc(f.Close)
		// assert
		assert.NoError(t, err)
	})

	t.Run("race/concurrentSameInstance", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeService{}
		const n = 32
		errs := make(chan error, n)
		var wg sync.WaitGroup
		wg.Add(n)

		// act
		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()
				errs <- r.RegisterFunc(f.Close)
			}()
		}
		wg.Wait()
		close(errs)

		// assert
		var nils, dups int
		for e := range errs {
			if e == nil {
				nils++
				continue
			}

			assert.ErrorIs(t, e, gracefully.ErrAlreadyRegistered)
			dups++
		}
		assert.Equal(t, n, nils)
	})

	t.Run("err/afterShutdown", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		_ = r.Shutdown(context.Background())
		// act
		f := &fakeService{}
		err := r.RegisterFunc(f.Close)
		// assert
		assert.ErrorIs(t, firstErr(r.Shutdown(context.Background())), gracefully.ErrShutdownCalled)
		assert.ErrorIs(t, err, gracefully.ErrShutdownCalled)
	})
}

func Test_Unregister(t *testing.T) {
	t.Parallel()

	t.Run("ok/idempotent", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeGSO{}
		assert.NoError(t, r.Register(f))
		// act
		ok1 := r.Unregister(f)
		ok2 := r.Unregister(f)
		// assert
		assert.True(t, ok1)
		assert.False(t, ok2)
	})

	t.Run("err/afterShutdown", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		f := &fakeGSO{}
		assert.NoError(t, r.Register(f))
		_ = r.Shutdown(context.Background())
		// act
		ok := r.Unregister(f)
		// assert
		assert.False(t, ok)
	})
}

// func Test_UnregisterFunc(t *testing.T) {
// 	t.Parallel()

// 	t.Run("ok/idempotent", func(t *testing.T) {
// 		t.Parallel()
// 		// arrange
// 		r := gracefully.NewRegistry()
// 		f := &fakeGSO{}
// 		assert.NoError(t, r.RegisterFunc(f.GracefulShutdown))
// 		// act
// 		ok1 := r.UnregisterFunc(f.GracefulShutdown)
// 		ok2 := r.UnregisterFunc(f.GracefulShutdown)
// 		// assert
// 		assert.True(t, ok1)
// 		assert.False(t, ok2)
// 	})

// 	t.Run("err/afterShutdown", func(t *testing.T) {
// 		t.Parallel()
// 		// arrange
// 		r := gracefully.NewRegistry()
// 		f := &fakeGSO{}
// 		assert.NoError(t, r.RegisterFunc(f.GracefulShutdown))
// 		_ = r.Shutdown(context.Background())
// 		// act
// 		ok := r.UnregisterFunc(f.GracefulShutdown)
// 		// assert
// 		assert.False(t, ok)
// 	})
// }

func Test_MustRegister(t *testing.T) {
	t.Parallel()

	t.Run("ok/multipleDistinct", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		a, b := &fakeGSO{}, &fakeGSO{}
		// act
		assert.NotPanics(t, func() {
			r.MustRegister(a, b)
		})
		// assert
		assert.True(t, r.Unregister(a))
		assert.True(t, r.Unregister(b))
	})

	t.Run("panic/alreadyRegistered", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		a := &fakeGSO{}
		// act + assert
		assert.Panics(t, func() {
			r.MustRegister(a, a)
		})
	})
}

func Test_Shutdown(t *testing.T) {
	t.Parallel()

	t.Run("ok/callsAll_and_collectsErrors", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		a := &fakeGSO{}
		b := &fakeGSO{ret: errors.New("boom")}
		c := &fakeGSO{}
		assert.NoError(t, r.Register(a))
		assert.NoError(t, r.Register(b))
		assert.NoError(t, r.Register(c))

		// act
		me := r.Shutdown(context.Background())

		// assert
		assert.Equal(t, int32(1), atomic.LoadInt32(&a.calls))
		assert.Equal(t, int32(1), atomic.LoadInt32(&b.calls))
		assert.Equal(t, int32(1), atomic.LoadInt32(&c.calls))

		// ожидаем, что ошибки собраны
		assertMultiErrorContains(t, me, b.ret)
	})

	t.Run("err/repeat", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		_ = r.Shutdown(context.Background())
		// act
		me := r.Shutdown(context.Background())
		// assert
		// MultiError должен содержать ErrAllInstanceShutdownAlready
		assertMultiErrorContains(t, me, gracefully.ErrShutdownCalled)
	})
}

func Test_WaitShutdown(t *testing.T) {
	t.Parallel()

	t.Run("ok/unblocksAfterShutdown", func(t *testing.T) {
		t.Parallel()
		// arrange
		r := gracefully.NewRegistry()
		done := make(chan struct{})
		go func() {
			// act
			r.WaitShutdown()
			close(done)
		}()

		select {
		case <-done:
			t.Fatalf("WaitShutdown returned before Shutdown")
		case <-time.After(80 * time.Millisecond):
			// ok, ещё блокируется
		}

		_ = r.Shutdown(context.Background())

		select {
		case <-done:
			// assert
		case <-time.After(150 * time.Millisecond):
			t.Fatalf("WaitShutdown did not unblock after Shutdown")
		}
	})
}
