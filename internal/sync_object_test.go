package internal_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lif0/go-gracefully/internal"
	"github.com/stretchr/testify/assert"
)

type pair struct {
	X int
	Y int
}

func Test_NewSyncObject(t *testing.T) {
	t.Parallel()

	t.Run("ok/basic", func(t *testing.T) {
		t.Parallel()
		// arrange
		so := internal.NewSyncObject(pair{X: 1, Y: 2})
		// act
		var got pair
		so.Mutate(func(p *pair) {
			got = *p
		})
		// assert
		assert.Equal(t, pair{X: 1, Y: 2}, got)
	})

	t.Run("ok/zeroValueWorks", func(t *testing.T) {
		t.Parallel()
		// arrange
		var so internal.SyncObject[int]
		// act
		so.Mutate(func(v *int) {
			*v = 42
		})
		// assert
		so.Mutate(func(v *int) {
			assert.Equal(t, 42, *v)
		})
	})
}

func Test_Mutate(t *testing.T) {
	t.Parallel()

	t.Run("ok/basic", func(t *testing.T) {
		t.Parallel()
		// arrange
		so := internal.NewSyncObject(0)
		// act
		so.Mutate(func(v *int) {
			*v = 7
		})
		// assert
		so.Mutate(func(v *int) {
			assert.Equal(t, 7, *v)
		})
	})

	t.Run("race/concurrent", func(t *testing.T) {
		t.Parallel()
		// arrange
		const workers = 16
		const perWorker = 1000
		so := internal.NewSyncObject(0)

		var wg sync.WaitGroup
		wg.Add(workers)

		// act
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < perWorker; j++ {
					so.Mutate(func(v *int) {
						*v++
					})
				}
			}()
		}
		wg.Wait()

		// assert
		so.Mutate(func(v *int) {
			assert.Equal(t, workers*perWorker, *v)
		})
	})

	t.Run("edge/panicInCallback", func(t *testing.T) {
		t.Parallel()
		// arrange
		so := internal.NewSyncObject(0)

		// act
		func() {
			defer func() {
				r := recover()
				// assert
				assert.NotNil(t, r)
			}()
			so.Mutate(func(v *int) {
				panic("boom")
			})
		}()

		// assert: мьютекс должен быть разблокирован — следующая мутация проходит
		done := make(chan struct{})
		go func() {
			so.Mutate(func(v *int) { *v = 123 })
			close(done)
		}()

		select {
		case <-done:
			// ok
		case <-time.After(150 * time.Millisecond):
			t.Fatalf("mutex remained locked after panic")
		}
	})

	t.Run("edge/nilFunc_panics", func(t *testing.T) {
		t.Parallel()
		// arrange
		so := internal.NewSyncObject(0)
		// act
		defer func() {
			r := recover()
			// assert
			assert.NotNil(t, r)
			assert.Contains(t, r.(interface{}), "nil") // сообщение зависит от рантайма, фиксируем факт паники
		}()
		var fn func(*int)
		so.Mutate(fn)
	})
}

func Test_GetObject(t *testing.T) {
	t.Parallel()

	t.Run("api/stableAddress", func(t *testing.T) {
		t.Parallel()
		// arrange
		so := internal.NewSyncObject(pair{X: 1, Y: 2})
		// act
		p1 := so.GetObject()
		p2 := so.GetObject()
		// assert
		assert.Same(t, p1, p2)
	})

	t.Run("api/leaksUnsafePointer", func(t *testing.T) {
		t.Parallel()
		// arrange
		type obj struct {
			X int
			Y int
		}
		so := internal.NewSyncObject(obj{X: 0, Y: 0})
		ptr := so.GetObject()

		locked := make(chan struct{})
		proceed := make(chan struct{})

		// act
		done := make(chan struct{})
		go func() {
			defer close(done)
			so.Mutate(func(v *obj) {
				close(locked) // RWMutex: write-lock удержан
				<-proceed     // ждём внешней записи
				v.X = 1       // записываем X после внешнего изменения
				// Y намеренно не трогаем
			})
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		select {
		case <-locked:
			ptr.Y = 99 // внешняя модификация без синхронизации
			close(proceed)
		case <-ctx.Done():
			t.Fatalf("timeout waiting for lock acquisition")
		}

		select {
		case <-done:
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("mutate did not finish")
		}

		// assert
		// Y изменён, хотя запись произошла во время удержания мьютекса в Mutate — нарушение инварианты инкапсуляции.
		so.Mutate(func(v *obj) {
			assert.Equal(t, 1, v.X)
			assert.Equal(t, 99, v.Y)
		})
	})
}
