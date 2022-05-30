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

// ClientOptionSetter configures a client.
type ClientOptionSetter interface {
	Set(*Client)
}

type ClientOptionSetterFunc func(*Client)

func (f ClientOptionSetterFunc) Set(c *Client) {
	f(c)
}

func WithClientMetrics(cm *ClientMetrics) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.metrics = cm
	})
}

func WithClientDisableAutoIncrementRequestID(b bool) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.options.disableAutoIncrementRequestID = b
	})
}

func WithClientTimeout(readtimeout,
	writetimeout,
	idletimeout time.Duration,
	flushInterval time.Duration) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.options.readTimeout = readtimeout
		c.options.writeTimeout = writetimeout
		c.options.idleTimeout = idletimeout
	})
}

func WithClientConn(conn net.Conn) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.conn = conn
	})
}

func WithClientHeartbeat(heartbeatinterval, heartbeattimeout time.Duration,
	heartbeatprobes int, onheartbeat func(success bool)) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.options.heartbeatTimeout = heartbeattimeout
		c.options.heartbeatInterval = heartbeatinterval
		c.options.heartbeatProbes = heartbeatprobes
		c.options.onHeartbeat = onheartbeat
	})
}

func WithClientMaxPendingCommands(m int) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.options.maxPendingCommands = m
	})
}

func WithClientRedial(dialer Dialer) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.options.dialer = dialer
	})
}

func WithClientHandler(handler Handler) ClientOptionSetterFunc {
	return ClientOptionSetterFunc(func(c *Client) {
		c.options.handler = handler
	})
}
