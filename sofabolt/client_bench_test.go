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
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

type BOLTConn struct {
	wb   bytes.Buffer
	req  Request
	res  Response
	idCh chan uint32
}

func (d *BOLTConn) Read(b []byte) (int, error) {
	id := <-d.idCh
	d.res.SetRequestID(id)
	var err error
	b, err = d.res.Write(&WriteOption{}, b[:0])
	return len(b), err
}

func (d *BOLTConn) Write(b []byte) (int, error) {
	d.wb.Write(b)
	for d.wb.Len() > 0 {
		n, err := d.req.Read(&ReadOption{}, &d.wb)
		if err != nil {
			return n, err
		}
		d.idCh <- d.req.GetRequestID()
	}
	return len(b), nil
}

func (do *BOLTConn) Close() error                       { return nil }
func (do *BOLTConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (do *BOLTConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (do *BOLTConn) SetDeadline(t time.Time) error      { return nil }
func (do *BOLTConn) SetReadDeadline(t time.Time) error  { return nil }
func (do *BOLTConn) SetWriteDeadline(t time.Time) error { return nil }

func BenchmarkClientConcurrent(b *testing.B) {
	dc := &BOLTConn{
		idCh: make(chan uint32, 16),
	}
	c, err := NewClient(WithClientConn(dc), WithClientMaxPendingCommands(
		b.N,
	))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		req := AcquireRequest()
		res := AcquireResponse()
		for pb.Next() {
			req.Reset()
			res.Reset()
			if err := c.Do(req, res); err != nil {
				continue
			}
			if req.GetRequestID() != res.GetRequestID() {
				b.Fail()
			}
		}
	})
}

func BenchmarkClient(b *testing.B) {
	var (
		req = AcquireRequest()
		res = AcquireResponse()
	)
	defer func() {
		ReleaseRequest(req)
		ReleaseResponse(res)
	}()

	dc := &BOLTConn{
		idCh: make(chan uint32, 16),
	}
	c, err := NewClient(WithClientConn(dc), WithClientMaxPendingCommands(
		b.N*2,
	))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req.Reset()
		res.Reset()
		if err = c.Do(req, res); err != nil {
			if err != io.EOF {
				b.Fatal(err)
			}
			break
		}
		if req.GetRequestID() != res.GetRequestID() {
			b.Fatalf("bug: id mismatch %d %d", req.GetRequestID(), res.GetRequestID())
		}
	}
}
