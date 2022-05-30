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
	"sync"
	"time"

	"github.com/sofastack/sofa-common-go/syncpool/bytespool"
	bufiorw "github.com/sofastack/sofa-common-go/writer/bufiorw"
)

var (
	b32Pool = sync.Pool{
		New: func() interface{} {
			var p [32]byte
			return &p
		},
	}

	bpool = bytespool.NewPool()

	requestPool = sync.Pool{
		New: func() interface{} {
			return &Request{}
		},
	}

	responsePool = sync.Pool{
		New: func() interface{} {
			return &Response{}
		},
	}

	brPool    = sync.Pool{}
	bwPool    = sync.Pool{}
	timerPool sync.Pool
	ictxPool  = sync.Pool{
		New: func() interface{} {
			return &InvokeContext{
				doneCh: make(chan struct{}),
			}
		},
	}
)

func AcquireInvokeContext(req *Request, res *Response, timeout time.Duration) *InvokeContext {
	ictx, ok := ictxPool.Get().(*InvokeContext)
	if !ok {
		panic("failed to type casting")
	}

	if len(ictx.errCh) == 1 || ictx.errCh == nil {
		ictx.errCh = make(chan error, 1) // sanity: allocate new error channel
	}

	ictx.req = req
	ictx.res = res
	ictx.created = time.Now()
	ictx.timeout = timeout

	return ictx
}

func ReleaseInvokeContext(ictx *InvokeContext) {
	ictxPool.Put(ictx)
}

func acquireB32() *[32]byte {
	return b32Pool.Get().(*[32]byte)
}

func releaseB32(p *[32]byte) {
	b32Pool.Put(p)
}

func AcquireRequest() *Request {
	req, ok := requestPool.Get().(*Request)
	if !ok {
		panic("failed to type casting")
	}
	req.SetProto(ProtoBOLTV1)
	req.SetCMDCode(CMDCodeBOLTRequest)
	req.SetType(TypeBOLTRequest)
	req.SetCodec(CodecHessian2)
	return req
}

func ReleaseRequest(di *Request) {
	di.Reset()
	requestPool.Put(di)
}

func AcquireResponse() *Response {
	res, ok := responsePool.Get().(*Response)
	if !ok {
		panic("failed to type casting")
	}
	res.SetProto(ProtoBOLTV1)
	res.SetCMDCode(CMDCodeBOLTResponse)
	res.SetType(TypeBOLTResponse)
	res.SetCodec(CodecHessian2)
	return res
}

func ReleaseResponse(di *Response) {
	di.Reset()
	responsePool.Put(di)
}

func acquireBufioWriter(w io.Writer) *bufiorw.Writer {
	i := bwPool.Get()
	if i == nil {
		return bufiorw.NewWriterSize(w, 8192)
	}

	bw, ok := i.(*bufiorw.Writer)
	if !ok {
		panic("failed to type casting")
	}
	bw.Reset(w)

	return bw
}

func releaseBufioWriter(bw *bufiorw.Writer) {
	bw.Reset(nil)
	bwPool.Put(bw)
}

func acquireBufioReader(r io.Reader) *bufiorw.Reader {
	i := brPool.Get()
	if i == nil {
		return bufiorw.NewReaderSize(r, 8192)
	}

	br, ok := i.(*bufiorw.Reader)
	if !ok {
		panic("failed to type casting")
	}
	br.Reset(r)

	return br
}

func releaseBufioReader(br *bufiorw.Reader) {
	br.Reset(nil)
	brPool.Put(br)
}

func initTimer(t *time.Timer, timeout time.Duration) *time.Timer {
	if t == nil {
		return time.NewTimer(timeout)
	}
	if t.Reset(timeout) {
		panic("BUG: active timer trapped into initTimer()")
	}
	return t
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		// Collect possibly added time from the channel
		// if timer has been stopped and nobody collected its' value.
		select {
		case <-t.C:
		default:
		}
	}
}

func AcquireTimer(timeout time.Duration) *time.Timer {
	v := timerPool.Get()
	if v == nil {
		return time.NewTimer(timeout)
	}
	t := v.(*time.Timer)
	initTimer(t, timeout)
	return t
}

func ReleaseTimer(t *time.Timer) {
	stopTimer(t)
	timerPool.Put(t)
}

func acquireBytes() *[]byte  { return bpool.Acquire() }
func releaseBytes(d *[]byte) { bpool.Release(d) }
