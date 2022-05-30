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
	"syscall"
)

// nolint
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

func IsDeadlineError(err error) bool {
	if err == nil {
		return false
	}

	nerr, ok := err.(net.Error)
	if !ok {
		return false
	}

	if !nerr.Timeout() {
		return false
	}

	// Check the syscall.ETIMEDOUT to avoid positive check

	oerr, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	errno, ok := oerr.Err.(syscall.Errno)
	if ok {
		return false
	}
	_ = errno

	return true
}
