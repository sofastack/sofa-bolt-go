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

package stateconn

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateConn(t *testing.T) {
	p0, p1 := net.Pipe()
	go func() {
		var p [1024]byte
		for {
			p1.Read(p[:])
			p1.Write(p[:])
		}
	}()
	sc := AcquireConn(p0)
	_, st := sc.GetState()
	require.Equal(t, StateNew, st)
	sc.Write([]byte("abcd"))
	_, st = sc.GetState()
	require.Equal(t, StateActive, st)
	sc.SetState(StateIdle)
	_, st = sc.GetState()
	require.Equal(t, StateIdle, st)

	var p [1024]byte
	sc.Read(p[:])
	_, st = sc.GetState()
	require.Equal(t, StateActive, st)

	sc.Close()
	_, st = sc.GetState()
	require.Equal(t, StateClosed, st)

	ReleaseConn(sc)
}
