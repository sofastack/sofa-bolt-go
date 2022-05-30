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
	"net"
	"time"
)

// ClientConnOptionSetter configures a ClientConn.
type ClientConnOptionSetter interface {
	Set(*ClientConn)
}

type ClientConnOptionSetterFunc func(*ClientConn)

func (f ClientConnOptionSetterFunc) Set(c *ClientConn) {
	f(c)
}

func WithClientConnMetrics(cm *ClientMetrics) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.metrics = cm
	})
}

func WithClientConnTimeout(readtimeout,
	writetimeout,
	idletimeout time.Duration,
	flushInterval time.Duration) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.options.readTimeout = readtimeout
		c.options.writeTimeout = writetimeout
		c.options.idleTimeout = idletimeout
	})
}

func WithClientConnConn(conn net.Conn) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.conn = conn
	})
}

func WithClientConnMaxPendingCommands(m int) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.options.maxPendingCommands = m
	})
}

func WithClientConnRedial(dialer Dialer) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.options.dialer = dialer
	})
}

func WithClientConnStatusChanger(changer ClientConnStatusChanger) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.options.statusChanger = changer
	})
}

func WithClientConnDispatcher(dispatcher ClientConnDispatcher) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.options.dispatcher = dispatcher
	})
}

func WithClientConnProtocolDecoder(dec ClientConnProtocolDecoder) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.dec = dec
	})
}

func WithClientConnProtocolEncoder(enc ClientConnProtocolEncoder) ClientConnOptionSetterFunc {
	return ClientConnOptionSetterFunc(func(c *ClientConn) {
		c.enc = enc
	})
}
