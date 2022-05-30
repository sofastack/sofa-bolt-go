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
	"errors"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

type StateGetter interface {
	GetState() (time.Time, State)
	GetConn() net.Conn
}

type StateConn struct {
	state int64
	net.Conn
}

// filer describes an object that has ability to return os.File.
type filer interface {
	// File returns a copy of object's file descriptor.
	File() (*os.File, error)
}

// File returns a copy of object's file descriptor.
func (s *StateConn) File() (*os.File, error) {
	if sf, ok := s.Conn.(filer); ok {
		return sf.File()
	}

	return nil, errors.New("not implement filer interface")
}

func (s *StateConn) SyscallConn() (syscall.RawConn, error) {
	if sf, ok := s.Conn.(syscall.Conn); ok {
		return sf.SyscallConn()
	}

	return nil, errors.New("not implement syscall.Conn interface")
}

func (s *StateConn) Write(p []byte) (n int, err error) {
	s.SetState(StateActive)
	return s.Conn.Write(p)
}

func (s *StateConn) Read(p []byte) (n int, err error) {
	s.SetState(StateActive)
	return s.Conn.Read(p)
}

func (s *StateConn) Close() error {
	// close once
	if _, state := s.GetState(); state != StateClosed {
		s.SetState(StateClosed)
		return s.Conn.Close()
	}
	return nil
}

func (s *StateConn) GetConn() net.Conn {
	return s.Conn
}

func (s *StateConn) SetState(state State) {
	atomic.StoreInt64(&s.state, packState(state))
}

func (s *StateConn) GetState() (time.Time, State) {
	i64 := atomic.LoadInt64(&s.state)
	return time.Unix(i64>>8, 0), State(i64 & 0xFF)
}
