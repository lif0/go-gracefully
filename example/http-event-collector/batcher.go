// This is just example
package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type eventBatcher struct {
	ctx       context.Context
	ctxCancel func()

	store *memStore
	fw    *fileWriter

	ebmu sync.Mutex
}

func NewBatcher(filePath string) *eventBatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &eventBatcher{
		ctx:       ctx,
		ctxCancel: cancel,
		fw:        &fileWriter{filePath: filePath},
		// Assuming a simple in-memory store implementation
		store: newSimpleMemStore(),
	}
}

func (b *eventBatcher) Store(data []string) {
	b.store.Store(data)
}

func (b *eventBatcher) Run() {
	t := time.NewTimer(time.Minute)

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-t.C:
			b.flushEvents()
			t.Reset(time.Second)
		}
	}
}

func (b *eventBatcher) Stop() {
	b.ctxCancel()
}

func (b *eventBatcher) flushEvents() error {
	b.ebmu.Lock()
	defer b.ebmu.Unlock()

	data := b.store.GetAndClear()
	if len(data) == 0 {
		return nil
	}

	err := b.fw.WriteToDisk(data)
	if err != nil {
		return err
	}

	log.Printf("eventBatcher(): flushed %d events to disk\n", len(data))
	return nil
}

func (b *eventBatcher) GracefulShutdown(ctx context.Context) error {
	defer log.Print("eventBatcher: call GracefulShutdown()")

	b.Stop()

	err := b.flushEvents()
	return err
}
