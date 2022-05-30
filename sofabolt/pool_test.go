package sofabolt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRangeDelete(t *testing.T) {
	var (
		assert = assert.New(t)
		pool   = NewPool()
	)

	pool.Push(&Client{})
	assert.Equal(1, pool.Size())

	pool.Iterate(func(c *Client) {
		pool.DeleteLocked(c)
	})
	assert.Equal(0, pool.Size())
}
