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

import "time"

// serverOptionSetter configures a Server.
type serverOptionSetter interface {
	set(*Server)
}

type serverOptionSetterFunc func(*Server)

func (f serverOptionSetterFunc) set(srv *Server) {
	f(srv)
}

func WithServerMetrics(sm *ServerMetrics) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.metrics = sm
	})
}

func WithServerHandler(fn Handler) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.handler = fn
	})
}

func WithServerAsync(t bool) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.options.async = t
	})
}

func WithServerTimeout(readTimeout, writeTimeout, idleTimeout, flushInterval time.Duration) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.options.readTimeout = readTimeout
		srv.options.writeTimeout = writeTimeout
		srv.options.idleTimeout = idleTimeout
		srv.options.flushInterval = flushInterval
	})
}

func WithServerMaxConnctions(m int) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.options.maxConnections = m
	})
}

func WithServerMaxPendingCommands(m int) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.options.maxPendingCommand = m
	})
}

func WithServerOnEventHandler(e ServerOnEventHandler) serverOptionSetter {
	return serverOptionSetterFunc(func(srv *Server) {
		srv.onhandler = e
	})
}
