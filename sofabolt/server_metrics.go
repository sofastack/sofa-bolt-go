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

import "sync/atomic"

type ServerMetrics struct {
	numwrite           int64
	numread            int64
	commands           int64
	pendingcommands    int64
	connections        int64
	pendingconnections int64
}

func (sm *ServerMetrics) GetBytesRead() int64 {
	return atomic.LoadInt64(&sm.numread)
}

func (sm *ServerMetrics) GetBytesWrite() int64 {
	return atomic.LoadInt64(&sm.numwrite)
}

func (sm *ServerMetrics) GetCommands() int64 {
	return atomic.LoadInt64(&sm.commands)
}

func (sm *ServerMetrics) GetPendingCommands() int64 {
	return atomic.LoadInt64(&sm.pendingcommands)
}

func (sm *ServerMetrics) GetConnections() int64 {
	return atomic.LoadInt64(&sm.connections)
}

func (sm *ServerMetrics) GetPendingConnections() int64 {
	return atomic.LoadInt64(&sm.pendingconnections)
}

func (sm *ServerMetrics) addConnections(n int64) {
	atomic.AddInt64(&sm.connections, n)
}

func (sm *ServerMetrics) addPendingConnections(n int64) {
	atomic.AddInt64(&sm.pendingconnections, n)
}

func (sm *ServerMetrics) addBytesRead(n int64) {
	atomic.AddInt64(&sm.numread, n)
}

func (sm *ServerMetrics) addBytesWrite(n int64) {
	atomic.AddInt64(&sm.numwrite, n)
}

func (sm *ServerMetrics) addCommands(n int64) {
	atomic.AddInt64(&sm.commands, n)
}

func (sm *ServerMetrics) addPendingCommands(n int64) {
	atomic.AddInt64(&sm.pendingcommands, n)
}
