package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lif0/go-gracefully"
)

var stopChan chan struct{}

func main() {
	// configure
	gracefully.SetShutdownTrigger(
		context.Background(),
		gracefully.WithSysSignal(),
		gracefully.WithUserChanSignal(stopChan),
	)
	// <========
	counter := NewCounter()

	gracefully.Register(counter)

	fmt.Printf("Last counter: %d\n", counter.val)
	fmt.Println("Press any key to increment (press 'q' to quit)")

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			counter.Inc()
			fmt.Printf("counter: %v\n", counter.val)
		}
	}()

	gracefully.WaitShutdown()
	fmt.Println("App finish")
}
