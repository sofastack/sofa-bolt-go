// nolint
// Copyright 20xx The Alipay Authors.
//
// @authors[0]: bingwu.ybw(bingwu.ybw@antfin.com|detailyang@gmail.com)
// @authors[1]: robotx(robotx@antfin.com)
//
// *Legal Disclaimer*
// Within this source code, the comments in Chinese shall be the original, governing version. Any comment in other languages are for reference only. In the event of any conflict between the Chinese language version comments and other language version comments, the Chinese language version shall prevail.
// *法律免责声明*
// 关于代码注释部分，中文注释为官方版本，其它语言注释仅做参考。中文注释可能与其它语言注释存在不一致，当中文注释与其它语言注释存在不一致时，请以中文注释为准。
//
//

package sofabolt

import (
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientHeartbeatFailed(t *testing.T) {
	p0, p1 := net.Pipe()
	_ = p1
	successCh := make(chan bool)
	c, err := NewClient(
		WithClientConn(p0),
		WithClientMaxPendingCommands(1),
		WithClientHeartbeat(100*time.Millisecond, 100*time.Millisecond,
			1, func(success bool) {
				successCh <- success
			}),
	)
	require.Nil(t, err)
	_ = c
	require.False(t, <-successCh)
}

func TestClientClose(t *testing.T) {
	p0, p1 := net.Pipe()
	_ = p1
	successCh := make(chan bool)
	c, err := NewClient(
		WithClientConn(p0),
		WithClientMaxPendingCommands(1),
		WithClientHeartbeat(100*time.Millisecond, 100*time.Millisecond,
			1, func(success bool) {
				successCh <- success
			}),
	)
	_ = c
	require.Nil(t, err)
	p1.Close()
	time.Sleep(1 * time.Second)
	err = p0.SetWriteDeadline(time.Time{})
	require.Contains(t, err.Error(), "read/write on closed pipe")
}

func TestClientHeartbeatSuccess(t *testing.T) {
	p0, p1 := net.Pipe()
	srv, _ := NewServer(WithServerHandler(HandlerFunc(func(rw ResponseWriter, r *Request) {
		rw.Write()
	})))

	go func() {
		srv.ServeConn(p1)
	}()

	successCh := make(chan bool)
	c, err := NewClient(
		WithClientConn(p0),
		WithClientMaxPendingCommands(1),
		WithClientHeartbeat(100*time.Millisecond, 100*time.Millisecond,
			1, func(success bool) {
				successCh <- success
			}),
	)
	require.Nil(t, err)
	_ = c
	require.True(t, <-successCh)
}

func TestClientBidirection(t *testing.T) {
	p0, p1 := net.Pipe()

	client1, err := NewClient(
		WithClientConn(p0),
		WithClientHandler(HandlerFunc(func(rw ResponseWriter, req *Request) {
			require.Equal(t, "client2", string(req.GetContent()))
			rw.GetResponse().SetContentString("client1")
			rw.Write()
		})),
	)
	require.Nil(t, err)

	client2, err := NewClient(
		WithClientConn(p1),
		WithClientHandler(HandlerFunc(func(rw ResponseWriter, req *Request) {
			require.Equal(t, "client1", string(req.GetContent()))
			rw.GetResponse().SetContentString("client2")
			rw.Write()
		})),
	)
	require.Nil(t, err)

	var (
		req = AcquireRequest()
		res = AcquireResponse()
	)
	req.SetContentString("client2")

	for i := 0; i < 32; i++ {
		err = client2.DoTimeout(req, res, 1*time.Second)
		require.Nil(t, err)
		require.Equal(t, "client1", string(res.GetContent()))
	}

	for i := 0; i < 32; i++ {
		req.SetContentString("client1")
		err = client1.DoTimeout(req, res, 1*time.Second)
		require.Nil(t, err)
		require.Equal(t, "client2", string(res.GetContent()))
	}
}

func TestClientRedial(t *testing.T) {
	req := AcquireRequest()
	res := AcquireResponse()
	defer func() {
		ReleaseRequest(req)
		ReleaseResponse(res)
	}()

	ln, err := net.Listen("tcp4", ":0")
	require.Nil(t, err)

	srv, err := NewServer(WithServerHandler(HandlerFunc(func(rw ResponseWriter, r *Request) {
		rw.Write()
	})))
	require.Nil(t, err)

	go func() {
		srv.Serve(ln)
	}()

	c, err := NewClient(
		WithClientRedial(DialerFunc(func() (net.Conn, error) {
			time.Sleep(100 * time.Millisecond)
			return net.Dial("tcp4", ln.Addr().String())
		})),
	)
	require.Nil(t, err)

	err = c.Do(req, res)
	require.Nil(t, err)

	time.Sleep(1 * time.Second)

	err = c.Do(req, res)
	require.Nil(t, err)
}

func TestClientClosedAfterConnClosed(t *testing.T) {
	assert := assert.New(t)
	clientConn, serverConn := net.Pipe()
	client, err := NewClient(
		WithClientConn(clientConn),
		WithClientTimeout(
			time.Second,
			time.Second,
			0,
			0,
		),
	)
	assert.Nil(err)

	time.Sleep(10 * time.Millisecond)
	serverConn.Close()

	time.Sleep(10 * time.Millisecond)
	assert.True(client.Closed())

	req := AcquireRequest()
	res := AcquireResponse()
	err = client.Do(req, res)
	assert.Equal(ErrClientWasClosed, err)
}

//nolint
func TestClientSeeIOEOF(t *testing.T) {
	assert := assert.New(t)
	p0, p1 := net.Pipe()
	c, err := NewClient(
		WithClientConn(p0),
		WithClientTimeout(
			10*time.Second,
			1*time.Second,
			0,
			1*time.Second,
		),
		WithClientMaxPendingCommands(10240),
	)
	require.Nil(t, err)

	var wg sync.WaitGroup
	count := 4
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			var (
				req = AcquireRequest()
				res = AcquireResponse()
			)
			// Block until see the error
			err := c.Do(req, res)
			require.NotNil(t, err)
			isEofErr := (err == io.EOF || strings.Contains(err.Error(), "on closed pipe"))
			assert.True(isEofErr, "expect eof error, got: %v", err)
		}()
	}

	time.Sleep(time.Second)
	p1.Close()
	wg.Wait()
}

func TestClientReadAndSendTimeout(t *testing.T) {
	p0, _ := net.Pipe()
	c, err := NewClient(
		WithClientConn(p0),
		WithClientTimeout(
			1*time.Second,
			1*time.Second,
			1*time.Second,
			1*time.Second,
		),
	)
	require.Nil(t, err)
	var (
		req = AcquireRequest()
		res = AcquireResponse()
	)

	err = c.DoTimeout(req, res, 100*time.Millisecond)
	require.Equal(t, ErrClientTimeout, err)

	errCh := make(chan error)
	err = c.DoCallbackTimeout(req, ClientCallbackerFunc(func(cerr error, ctx *InvokeContext) {
		errCh <- cerr
	}), 100*time.Millisecond)
	require.Nil(t, err)
	cerr := <-errCh
	ne, ok := cerr.(net.Error)
	require.True(t, ok)
	require.True(t, ne.Timeout())
}

func TestClientPolyfill(t *testing.T) {
	_, err := NewClient()
	require.NotNil(t, err)

	p0, _ := net.Pipe()
	_, err = NewClient(WithClientConn(p0))
	require.Nil(t, err)
}

func TestClientMetrics(t *testing.T) {
	p0, p1 := net.Pipe()
	srv, err := NewServer(WithServerHandler(HandlerFunc(func(rw ResponseWriter, r *Request) {
		rw.GetResponse().SetContentString("Hello World")
		rw.Write()
	})))
	require.Nil(t, err)

	go func() {
		srv.ServeConn(p1)
	}()

	c, err := NewClient(WithClientConn(p0))
	require.Nil(t, err)

	count := 100
	for i := 0; i < count; i++ {
		req := AcquireRequest()
		res := AcquireResponse()
		if err := c.DoTimeout(req, res, 1*time.Second); err != nil {
			t.Fatal(err)
		}
	}

	c.Close()
	require.Equal(t, int64(7700), c.GetMetrics().GetBytesRead())
	require.Equal(t, int64(2200), c.GetMetrics().GetBytesWrite())
	require.Equal(t, int64(count), c.GetMetrics().GetCommands())
	require.Equal(t, int64(0), c.GetMetrics().GetPendingCommands())
	require.Equal(t, int64(0), c.GetMetrics().GetReferences())
	require.Equal(t, int64(100), c.GetMetrics().GetUsed())
	require.Equal(t, c.metrics.lasted, c.GetMetrics().GetLasted())
	require.Equal(t, c.metrics.created, c.GetMetrics().GetCreated())
	require.True(t, c.Closed())
}

func TestClientWithMaxConcurrentRequests(t *testing.T) {
	var sleep uint32
	sleep = 5
	p0, p1 := net.Pipe()
	srv, _ := NewServer(WithServerHandler(HandlerFunc(func(rw ResponseWriter, r *Request) {
		n := atomic.LoadUint32(&sleep)
		time.Sleep(time.Duration(n) * time.Second)
		rw.GetResponse().SetContentString("Hello World")
		rw.Write()
	})))

	go func() {
		srv.ServeConn(p0)
	}()

	c, err := NewClient(WithClientConn(p1), WithClientMaxPendingCommands(1))
	if err != nil {
		t.Fatal(err)
	}

	var failed uint32
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := AcquireRequest()
			res := AcquireResponse()
			defer ReleaseRequest(req)
			defer ReleaseResponse(res)
			err := c.DoTimeout(req, res, 1*time.Second)
			if err != ErrClientTimeout {
				atomic.AddUint32(&failed, 1)
			}
		}()
	}

	wg.Wait()
	if atomic.LoadUint32(&failed) == 1 {
		t.Error("expect timeout but not")
	}
}
