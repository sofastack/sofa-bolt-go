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
	"sync/atomic"
	"time"
)

type ClientMetrics struct {
	nread           int64
	nwrite          int64
	commands        int64
	pendingCommands int64
	references      int64
	used            int64
	lasted          int64
	created         int64
}

func (cm *ClientMetrics) GetBytesRead() int64         { return atomic.LoadInt64(&cm.nread) }
func (cm *ClientMetrics) GetBytesWrite() int64        { return atomic.LoadInt64(&cm.nwrite) }
func (cm *ClientMetrics) GetCommands() int64          { return atomic.LoadInt64(&cm.commands) }
func (cm *ClientMetrics) GetPendingCommands() int64   { return atomic.LoadInt64(&cm.pendingCommands) }
func (cm *ClientMetrics) ResetPendingCommands()       { atomic.StoreInt64(&cm.pendingCommands, 0) }
func (cm *ClientMetrics) GetReferences() int64        { return atomic.LoadInt64(&cm.references) }
func (cm *ClientMetrics) AddReferences(n int64) int64 { return atomic.AddInt64(&cm.references, n) }
func (cm *ClientMetrics) GetUsed() int64              { return atomic.LoadInt64(&cm.used) }
func (cm *ClientMetrics) AddUsed(n int64) int64       { return atomic.AddInt64(&cm.used, n) }
func (cm *ClientMetrics) GetLasted() int64            { return atomic.LoadInt64(&cm.lasted) }
func (cm *ClientMetrics) SetLasted() {
	atomic.StoreInt64(&cm.lasted, time.Now().Unix())
}
func (cm *ClientMetrics) GetCreated() int64 { return atomic.LoadInt64(&cm.created) }
