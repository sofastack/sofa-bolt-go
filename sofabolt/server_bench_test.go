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
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/sofastack/sofa-common-go/helper/testnet"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func BenchmarkServer(b *testing.B) {
	b.Run("1 connection", func(b *testing.B) {
		benchmarkServer(b, 1, b.N, false)
	})

	b.Run("128 connection", func(b *testing.B) {
		benchmarkServer(b, 128, b.N/128+1, false)
	})

	b.Run("512 connection", func(b *testing.B) {
		benchmarkServer(b, 512, b.N/512+1, false)
	})

	b.Run("1024 connection", func(b *testing.B) {
		benchmarkServer(b, 1024, b.N/1024+1, false)
	})
}

func benchmarkServer(b *testing.B, connections int, countperconn int, async bool) {
	doneCh := make(chan struct{})
	count := uint64(0)
	srv, err := NewServer(
		WithServerTimeout(5*time.Second, 5*time.Second, 10*time.Second, 10*time.Millisecond),
		WithServerMaxConnctions(connections*2),
		WithServerMaxPendingCommands(10240),
		WithServerAsync(async),
		WithServerHandler(
			HandlerFunc(func(
				rw ResponseWriter,
				req *Request) {
				res := rw.GetResponse()
				res.SetContent(req.GetContent())
				_, err := rw.Write()
				if err != nil {
					b.Fatal("failed to write", err)
				}
				if atomic.AddUint64(&count, 1) == uint64(connections*countperconn) {
					doneCh <- struct{}{}
				}
			},
			)))
	require.Nil(b, err)

	bc := testnet.NewBufioConn(nil)
	c, err := NewClient(WithClientConn(bc), WithClientMaxPendingCommands(
		countperconn,
	))
	if err != nil {
		log.Fatal(err)
	}

	var (
		req = AcquireRequest()
		res = AcquireResponse()
	)
	req.SetType(TypeBOLTRequestOneWay)
	defer func() {
		ReleaseRequest(req)
		ReleaseResponse(res)
	}()

	for i := 0; i < countperconn; i++ {
		if err = c.Do(req, res); err != nil {
			b.Fatal(err)
		}
	}

	// wait write to dst.
	var payload []byte
	for {
		time.Sleep(100 * time.Millisecond)
		payload = bc.Bytes()
		if len(payload) > 0 {
			break
		}
	}

	ml := testnet.NewDataListener()
	for i := 0; i < connections; i++ {
		ml.AddConn(testnet.NewDataConn(
			payload,
		))
	}

	b.ReportAllocs()
	b.ResetTimer()

	go srv.Serve(ml)
	<-doneCh
}
