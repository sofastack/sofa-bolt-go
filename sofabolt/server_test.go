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
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MyHandler struct {
	i uint32
}

func (mh *MyHandler) ServeSofaBOLT(rw ResponseWriter, req *Request) {
	atomic.AddUint32(&mh.i, 1)
	res := rw.GetResponse()
	res.SetContentString("hello world")
	h := res.GetHeaders()
	h.Set("count", req.GetHeaders().Get("count"))
	h.Set("service", req.GetHeaders().Get("service"))
	res.SetStatus(200)

	if _, err := rw.Write(); err != nil {
		log.Fatal("failed to write ", err, atomic.LoadUint32(&mh.i))
	}
}

func TestServer(t *testing.T) {
	srv, err := NewServer(
		WithServerHandler(&MyHandler{}),
	)
	require.Nil(t, err)

	errCh := make(chan error)
	p0, p1 := net.Pipe()
	go func() {
		errCh <- srv.ServeConn(p0)
	}()

	c, err := NewClient(WithClientConn(p1))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		cn := fmt.Sprintf("%03d", i)
		req := AcquireRequest()
		req.Reset()
		h := req.GetHeaders()
		h.Set("count", cn)
		res := AcquireResponse()
		err := c.DoTimeout(req, res, 100*time.Millisecond)
		require.Nil(t, err)
		if h.Get("count") != cn {
			t.Fatalf("expect get content: %s but got %s", cn, h.Get("count"))
		}
	}

	c.Close()
	serr := <-errCh
	require.Equal(t, io.EOF, serr)

	require.Equal(t, int64(3800), srv.GetMetrics().GetBytesRead())
	require.Equal(t, int64(9300), srv.GetMetrics().GetBytesWrite())
	require.Equal(t, int64(100), srv.GetMetrics().GetCommands())
	require.Equal(t, int64(1), srv.GetMetrics().GetConnections())
	require.Equal(t, int64(0), srv.GetMetrics().GetPendingCommands())
}

func TestServerParallelWrite(t *testing.T) {
	srv, _ := NewServer(WithServerHandler(&MyHandler{}))

	errCh := make(chan error)
	p0, p1 := net.Pipe()
	go func() {
		errCh <- srv.ServeConn(p0)
	}()

	c, err := NewClient(
		WithClientConn(p1),
		WithClientMaxPendingCommands(102400),
	)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 128; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := AcquireRequest()
			req.GetHeaders().Set("test", "abcd")
			res := AcquireResponse()
			defer func() {
				ReleaseRequest(req)
				ReleaseResponse(res)
			}()

			err := c.DoTimeout(req, res, 100*time.Millisecond)
			if err != nil {
				log.Fatal(err)
			}
			require.Nil(t, err)
			require.Equal(t, "hello world", string(res.GetContent()))
		}()
	}

	wg.Wait()

	c.Close()
	serr := <-errCh
	require.Equal(t, io.EOF, serr)
}

func TestServerListener(t *testing.T) {
	count := int64(0)
	handler := func(rw ResponseWriter, req *Request) {
		atomic.AddInt64(&count, 1)
		rw.GetResponse().SetContent([]byte("hello world"))
		rw.GetResponse().SetClassString("test")
		rw.Write()
	}

	ln, err := net.Listen("tcp4", ":0")
	require.Nil(t, err)

	srv, err := NewServer(
		WithServerHandler(HandlerFunc(handler)),
	)
	require.Nil(t, err)

	doneCh := make(chan struct{})
	go func() {
		srv.Serve(ln)
		doneCh <- struct{}{}
	}()

	conn, err := net.Dial("tcp4", ln.Addr().String())
	require.Nil(t, err)

	c, err := NewClient(
		WithClientConn(conn),
		WithClientMaxPendingCommands(1024),
	)
	if err != nil {
		t.Fatal(err)
	}

	var (
		req = AcquireRequest()
		res = AcquireResponse()
	)

	for i := 0; i < 128; i++ {
		err = c.DoTimeout(req, res, 500*time.Millisecond)
		require.Nil(t, err)
		require.Equal(t, "hello world", string(res.GetContent()))
		require.Equal(t, "test", string(res.GetClass()))
	}

	c.Close()
	ln.Close()
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	<-doneCh
}

func TestServerReadTimeout(t *testing.T) {
	srv, err := NewServer(
		WithServerHandler(&MyHandler{}),
		WithServerTimeout(
			1*time.Second,
			1*time.Second,
			1*time.Second,
			1*time.Second,
		),
	)
	require.Nil(t, err)

	p0, _ := net.Pipe()
	err = srv.ServeConn(p0)
	ne, ok := err.(net.Error)
	require.True(t, ok)
	require.True(t, ne.Timeout())
}

func TestServerWriteTimeout(t *testing.T) {
	srv, err := NewServer(
		WithServerHandler(&MyHandler{}),
		WithServerTimeout(
			0*time.Second,
			1*time.Second,
			1*time.Second,
			1*time.Second,
		),
	)
	require.Nil(t, err)

	p0, p1 := net.Pipe()
	go func() {
		req := AcquireRequest()
		d, derr := req.Write(&WriteOption{}, nil)
		require.Nil(t, derr)
		_, derr = p0.Write(d[:])
		require.Nil(t, derr)
	}()

	err = srv.ServeConn(p1)
	ne, ok := err.(net.Error)
	require.True(t, ok)
	require.True(t, ne.Timeout())
}

func TestServerSeeEOF(t *testing.T) {
	srv, err := NewServer(
		WithServerHandler(&MyHandler{}),
		WithServerTimeout(
			0*time.Second,
			1*time.Second,
			1*time.Second,
			1*time.Second,
		),
	)
	require.Nil(t, err)
	p0, p1 := net.Pipe()
	go func() {
		time.Sleep(1 * time.Second)
		p1.Close()
	}()

	err = srv.ServeConn(p0)
	require.Equal(t, io.EOF, err)
}
