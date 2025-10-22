package internal

import "sync"

type SyncObject[T any] struct {
	mu sync.RWMutex
	v  T
}

func NewSyncObject[T any](v T) *SyncObject[T] {
	return &SyncObject[T]{
		mu: sync.RWMutex{},
		v:  v,
	}
}

func (so *SyncObject[T]) Mutate(f func(v *T)) {
	so.mu.Lock()
	defer so.mu.Unlock()

	f(&so.v)
}

func (so *SyncObject[T]) GetObject() *T {
	so.mu.Lock()
	defer so.mu.Unlock()

	return &so.v
}
