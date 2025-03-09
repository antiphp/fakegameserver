package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFifo(t *testing.T) {
	f := NewFifo[int]()
	f.Add(1)
	f.Add(2)
	f.Add(3)

	v, stopped := f.Get()

	assert.Equal(t, 1, v)
	assert.False(t, stopped)

	v, stopped = f.Get()

	assert.Equal(t, 2, v)
	assert.False(t, stopped)

	f.Add(4)

	v, stopped = f.Get()

	assert.Equal(t, 3, v)
	assert.False(t, stopped)

	v, stopped = f.Get()

	assert.Equal(t, 4, v)
	assert.False(t, stopped)

	f.Shutdown()
	v, stopped = f.Get()

	assert.Equal(t, 0, v)
	assert.True(t, stopped)
}
