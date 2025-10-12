package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

const filename = "counter.txt"

type counter struct {
	val int
}

func NewCounter() *counter {
	return &counter{
		val: tryReadFile(),
	}
}

func (c *counter) Inc() {
	c.val++

	if c.val%10 == 0 {
		go c.flush()
	}
}

func (c *counter) flush() error {
	if err := os.WriteFile(filename, []byte(strconv.Itoa(c.val)), 0o644); err != nil {
		return fmt.Errorf("error saving counter: %v", err)
	}

	return nil
}

// github.com/lif0/go-gracefully/gogracefully.GracefulShutdownObject
func (c *counter) GracefulShutdown(ctx context.Context) error {
	return c.flush()
}

func tryReadFile() int {
	data, err := os.ReadFile(filename)
	if err == nil {
		counter, err := strconv.Atoi(string(data))
		if err != nil {
			return 0
		}

		return counter
	}

	return 0
}
