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

import "net"

//go:generate stringer -type=ServerEvent

type ServerEvent uint16

const (
	ServerTemporaryAcceptEvent    ServerEvent = 0
	ServerWorkerPoolOverflowEvent ServerEvent = 1
	ServerConnErrorEvent          ServerEvent = 2
	ServerConnHijackedEvent       ServerEvent = 3
)

type ServerEventContext struct {
	req   *Request
	res   *Response
	conn  net.Conn
	event ServerEvent
}

func NewServerEventContext(event ServerEvent) *ServerEventContext {
	return &ServerEventContext{event: event}
}

func (s ServerEventContext) GetType() ServerEvent { return s.event }

func (sec *ServerEventContext) SetConn(conn net.Conn) *ServerEventContext {
	sec.conn = conn
	return sec
}

func (sec *ServerEventContext) SetReq(req *Request) *ServerEventContext {
	sec.req = req
	return sec
}

func (sec *ServerEventContext) SetRes(res *Response) *ServerEventContext {
	sec.res = res
	return sec
}

type ServerOnEventHandler func(*Server, error, *ServerEventContext)

var DummyServerOnEventHandler = ServerOnEventHandler(func(*Server, error, *ServerEventContext) {
})
