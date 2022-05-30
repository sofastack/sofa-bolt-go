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

package errorconn

import (
	"net"
	"time"
)

type Conn struct {
	err error
}

func New(err error) net.Conn {
	return &Conn{err: err}
}

func (dc *Conn) Read(b []byte) (n int, err error) {
	return 0, dc.err
}

func (dc *Conn) Write(b []byte) (n int, err error) {
	return 0, dc.err
}

func (dc *Conn) Close() error {
	return dc.err
}

func (dc *Conn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (dc *Conn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (dc *Conn) SetDeadline(t time.Time) error {
	return dc.err
}

func (dc *Conn) SetReadDeadline(t time.Time) error {
	return dc.err
}

func (dc *Conn) SetWriteDeadline(t time.Time) error {
	return dc.err
}
