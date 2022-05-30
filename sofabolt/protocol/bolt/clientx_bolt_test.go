package bolt

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/sofastack/sofa-bolt-go/sofabolt"
)

func TestBOLTClient(t *testing.T) {
	p0, p1 := net.Pipe()

	srv, err := sofabolt.NewServer(
		sofabolt.WithServerHandler(sofabolt.HandlerFunc(func(rw sofabolt.ResponseWriter,
			req *sofabolt.Request) {
			rw.GetResponse().SetContentString("hello world")
			rw.Write()
		})),
		sofabolt.WithServerTimeout(
			0*time.Second,
			10*time.Second,
			60*time.Second,
			0*time.Second,
		),
	)
	require.Nil(t, err)
	go func() {
		srv.ServeConn(p1)
	}()

	c0, err := NewClientConnBOLT(
		WithClientConnBOLTConn(p0),
		WithClientConnBOLTMaxPendingCommands(128),
		WithClientConnBOLTTimeout(
			0,
			30*time.Second,
			0,
			0,
		),
	)
	require.Nil(t, err)

	req := sofabolt.AcquireRequest()
	res := sofabolt.AcquireResponse()
	defer func() {
		sofabolt.ReleaseRequest(req)
		sofabolt.ReleaseResponse(res)
	}()

	for i := 0; i < 1024; i++ {
		req.Reset()
		res.Reset()
		err = c0.Do(req, res)
		require.Nil(t, err)
		require.Equal(t, string(res.GetContent()), "hello world")
	}
}
