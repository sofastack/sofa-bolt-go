package sofabolt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testContextKey string

func TestRequestCopyTo(t *testing.T) {
	assert := assert.New(t)

	ctx := context.WithValue(context.TODO(), testContextKey("foo"), "bar")
	r1 := new(Request).SetContext(ctx)
	r2 := new(Request)
	r1.CopyTo(r2)
	assert.NotEqual(r1.GetContext(), r2.GetContext())
}
