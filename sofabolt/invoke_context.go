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
	"sync"
	"time"
)

type InvokeContext struct {
	// nolint
	noCopy   noCopy
	timeout  time.Duration
	created  time.Time
	req      *Request
	res      *Response
	ireslock sync.Mutex
	ires     Response
	errCh    chan error
	doneCh   chan struct{}
	callback ClientCallbacker
}

func NewInvokeContext(req *Request) *InvokeContext {
	return &InvokeContext{
		created: time.Now(),
		req:     req,
	}
}

func (i *InvokeContext) GetCallback() ClientCallbacker {
	return i.callback
}

func (i *InvokeContext) GetDeadline() time.Time {
	return i.created.Add(i.timeout)
}

func (i *InvokeContext) GetTimeout() time.Duration                      { return i.timeout }
func (i *InvokeContext) SetTimeout(t time.Duration) *InvokeContext      { i.timeout = t; return i }
func (i *InvokeContext) SetCallback(cb ClientCallbacker) *InvokeContext { i.callback = cb; return i }
func (i *InvokeContext) GetCreated() time.Time                          { return i.created }
func (i *InvokeContext) GetRequest() *Request                           { return i.req }
func (i *InvokeContext) GetErrorCh() chan error                         { return i.errCh }

func (i *InvokeContext) CopyResponse(res *Response) {
	i.ireslock.Lock()
	i.ires.CopyTo(res)
	i.ireslock.Unlock()
}

func (i *InvokeContext) AssignResponse(res *Response) {
	i.ireslock.Lock()
	i.ires.Reset()
	res.CopyTo(&i.ires)
	i.ireslock.Unlock()
}

func (i *InvokeContext) GetResponse() *Response {
	return i.res
}

func (i *InvokeContext) Invoke(err error, res *Response) {
	if i.callback != nil {
		i.res = res
		i.callback.Invoke(err, i)

	} else {
		i.AssignResponse(res)
		// Notify the sender
		i.errCh <- err
	}
}
