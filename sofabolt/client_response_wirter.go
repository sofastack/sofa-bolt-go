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

var crwPool = sync.Pool{
	New: func() interface{} {
		return &clientResponseWriter{}
	},
}

func acquireClientResponseWriter(c *Client) *clientResponseWriter {
	crw, ok := crwPool.Get().(*clientResponseWriter)
	if !ok {
		panic("failed to type casting")
	}

	crw.c = c
	crw.pool = bytesPool

	return crw
}

func releaseClientResponseWriter(crw *clientResponseWriter) {
	crw.reset(nil)
	crwPool.Put(crw)
}

type clientResponseWriter struct {
	numwrite int64
	err      uatomic.Error
	c        *Client
	res      Response
	pool     *bytespool.Pool
}

func (c *clientResponseWriter) Derive(req *Request) {
	c.res.
		SetProto(req.GetProto()).
		SetRequestID(req.GetRequestID()).
		SetCMDCode(CMDCodeBOLTResponse)

	if req.GetCMDCode() == CMDCodeBOLTHeartbeat {
		c.res.SetCMDCode(CMDCodeBOLTHeartbeat)
	} else if req.GetCMDCode() == CMDCodeTRemotingRequest {
		c.res.SetCMDCode(CMDCodeTRemotingResponse)
	}
}

func (c *clientResponseWriter) GetID() uint64 {
	return 0
}

func (c *clientResponseWriter) GetWriter() io.Writer {
	return c.c.GetConn()
}

func (c *clientResponseWriter) GetResponse() *Response {
	return &c.res
}

func (c *clientResponseWriter) GetConn() net.Conn {
	return c.c.GetConn()
}

func (c *clientResponseWriter) getNumWrite() int64 { return atomic.LoadInt64(&c.numwrite) }
func (c *clientResponseWriter) setNumWrite(nw int) { atomic.StoreInt64(&c.numwrite, int64(nw)) }

func (c *clientResponseWriter) Write() (int, error) {
	if c.getNumWrite() > 0 {
		return 0, errors.New("sofabolt: duplicated write")
	}

	var (
		err error
		dp  = c.pool.Acquire()
		nw  int
	)

	*dp, err = c.res.Write(&WriteOption{}, (*dp)[:0])
	if err != nil {
		c.pool.Release(dp)
		return 0, err
	}

	nw, err = c.c.write(*dp)
	c.pool.Release(dp)
	c.setNumWrite(nw)

	if err != nil {
		c.err.Store(err)
		return nw, err
	}

	return nw, err
}

func (c *clientResponseWriter) GetWriteError() error {
	return c.err.Load()
}

func (c *clientResponseWriter) reset(client *Client) *clientResponseWriter {
	c.c = client
	c.numwrite = 0
	c.res.Reset()
	return c
}

func (c *clientResponseWriter) Hijack() (net.Conn, bool) {
	return nil, false
}
