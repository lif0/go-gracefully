package main

import (
	gracefully "github.com/lif0/go-gracefully"
)

func main() {
	serverEventCollector := NewBatcher("server-event-collector")
	userEventCollector := NewBatcher("user-event-collector")

	gracefully.MustRegister(serverEventCollector, userEventCollector)
}

func mainV2() {
	serverEventCollector := gracefully.New(func() *eventBatcher { return NewBatcher("server-event-collector") })
	userEventCollector := gracefully.New(func() *eventBatcher { return NewBatcher("user-event-collector") })

	gracefully.MustRegister(serverEventCollector, userEventCollector)
}
