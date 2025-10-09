package main

import (
	"context"
	"sync"
	"time"
)

type db interface {
	InsertRows(context.Context, []string) error
}

type eventBatcher struct {
	name      string
	ctx       context.Context
	ctxCancel func()

	mu   sync.Mutex
	data []string
	db   db
}

func NewBatcher(name string) *eventBatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &eventBatcher{
		ctx:       ctx,
		ctxCancel: cancel,
	}
}

func (b *eventBatcher) Run() {
	t := time.NewTimer(time.Second)

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-t.C:
			b.flushEvents(b.ctx)
		}
	}
}

func (b *eventBatcher) Stop() {
	b.ctxCancel()
}

func (b *eventBatcher) flushEvents(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := b.db.InsertRows(ctx, b.data)
	if err != nil {
		return err
	}

	b.data = []string{}
	return nil
}

func (b *eventBatcher) GracefulShutdownName() string {
	return b.name
}

func (b *eventBatcher) GracefulShutdown(ctx context.Context) error {
	err := b.flushEvents(ctx)
	return err
}
