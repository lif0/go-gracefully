package gracefully_test

import (
	"context"
	"testing"

	"github.com/lif0/go-gracefully"
	"github.com/stretchr/testify/assert"
)

type someStruct struct{}

func (ss *someStruct) GracefulShutdown(context.Context) error {
	return nil
}

func Test_NewInstance(t *testing.T) {
	obj := &someStruct{}

	assert.False(t, gracefully.Unregister(obj))

	gracefully.NewInstance(func() *someStruct { return obj })

	assert.True(t, gracefully.Unregister(obj))
}
