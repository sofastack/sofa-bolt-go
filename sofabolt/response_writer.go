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
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/sofastack/sofa-common-go/syncpool/bytespool"
	uatomic "go.uber.org/atomic"
)

var (
	bytesPool = bytespool.NewPool()
	id        uint64
)

type ResponseWriter interface {
	GetID() uint64
	GetConn() net.Conn
	GetWriter() io.Writer
	GetResponse() *Response
	Hijack() (net.Conn, bool)
	Write() (int, error)
	GetWriteError() error
}

type TestResponseWriter struct {
	sync.Mutex
	ID       uint64
	Conn     net.Conn
	Writer   io.Writer
	Response Response
	Error    uatomic.Error
	NumWrite int
	Hijacked uint32
}

func (rw *TestResponseWriter) Hijack() (net.Conn, bool) {
	return nil, false
}

func (rw *TestResponseWriter) GetConn() net.Conn {
	rw.Lock()
	defer rw.Unlock()
	return rw.Conn
}

func (rw *TestResponseWriter) GetID() uint64 {
	rw.Lock()
	defer rw.Unlock()
	return rw.ID
}

func (rw *TestResponseWriter) GetResponse() *Response {
	rw.Lock()
	defer rw.Unlock()
	return &rw.Response
}

func (rw *TestResponseWriter) GetWriteError() error {
	return rw.Error.Load()
}

func (rw *TestResponseWriter) Write() (int, error) {
	rw.Lock()
	defer rw.Unlock()
	if rw.NumWrite > 0 {
		return 0, errors.New("sofabolt: duplicated write")
	}

	var (
		err error
		dp  []byte
	)

	dp, err = rw.Response.Write(&WriteOption{}, dp)
	if err != nil {
		return 0, err
	}

	rw.NumWrite, err = rw.Writer.Write(dp)

	if err != nil {
		// store the error
		rw.Error.Store(err)
		return 0, err
	}

	return rw.NumWrite, err
}

func (rw *TestResponseWriter) GetWriter() io.Writer {
	return rw.Writer
}

type SofaResponseWriter struct {
	id       uint64
	pool     *bytespool.Pool
	conn     net.Conn
	writer   io.Writer
	numwrite int
	res      Response
	err      uatomic.Error
	hijacked uint32
}

var ioResponseWriterPool = sync.Pool{
	New: func() interface{} {
		return &SofaResponseWriter{
			pool: bytesPool,
		}
	},
}

func AcquireSofaResponseWriter(conn net.Conn, w io.Writer) *SofaResponseWriter {
	rw, ok := ioResponseWriterPool.Get().(*SofaResponseWriter)
	if !ok {
		panic("failed to type casting")
	}

	rw.id = atomic.AddUint64(&id, 1)
	rw.writer = w
	rw.conn = conn
	rw.pool = bytesPool
	return rw
}

func ReleaseSofaResponseWriter(crw *SofaResponseWriter) {
	crw.Reset(nil)
	ioResponseWriterPool.Put(crw)
}

func (rw *SofaResponseWriter) GetConn() net.Conn { return rw.conn }

func (rw *SofaResponseWriter) GetID() uint64 { return atomic.LoadUint64(&rw.id) }

func (rw *SofaResponseWriter) Reset(w io.Writer) *SofaResponseWriter {
	rw.writer = w
	rw.numwrite = 0
	rw.res.Reset()
	atomic.StoreUint32(&rw.hijacked, 0)
	return rw
}

func (rw *SofaResponseWriter) Derive(req *Request) {
	rw.res.
		SetProto(req.GetProto()).
		SetRequestID(req.GetRequestID()).
		SetCMDCode(CMDCodeBOLTResponse).
		SetClassString(ClassResponse)

	if req.GetCMDCode() == CMDCodeBOLTHeartbeat {
		rw.res.SetCMDCode(CMDCodeBOLTHeartbeat)
	} else if req.GetCMDCode() == CMDCodeTRemotingRequest {
		rw.res.SetCMDCode(CMDCodeTRemotingResponse)
	}
}

func (rw *SofaResponseWriter) GetResponse() *Response {
	return &rw.res
}

func (rw *SofaResponseWriter) GetWriteError() error {
	return rw.err.Load()
}

func (rw *SofaResponseWriter) Write() (int, error) {
	if rw.numwrite > 0 {
		return 0, errors.New("sofabolt: duplicated write")
	}

	var err error
	dp := rw.pool.Acquire()

	*dp, err = rw.res.Write(&WriteOption{}, (*dp)[:0])
	if err != nil {
		rw.pool.Release(dp)
		return 0, err
	}

	rw.numwrite, err = rw.writer.Write(*dp)
	rw.pool.Release(dp)

	if err != nil {
		// store the error
		rw.err.Store(err)
		return rw.numwrite, err
	}

	return rw.numwrite, err
}

func (rw *SofaResponseWriter) GetWriter() io.Writer {
	return rw.writer
}

func (rw *SofaResponseWriter) Hijack() (net.Conn, bool) {
	if atomic.CompareAndSwapUint32(&rw.hijacked, 0, 1) {
		return rw.conn, true
	}
	return nil, false
}

func (rw *SofaResponseWriter) IsHijacked() bool {
	return atomic.LoadUint32(&rw.hijacked) == 1
}
