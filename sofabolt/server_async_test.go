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
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerAsync(t *testing.T) {
	srv, err := NewServer(
		WithServerHandler(&MyHandler{}),
		WithServerAsync(true),
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
