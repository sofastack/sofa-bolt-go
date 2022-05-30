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

package metricconn

import (
	"crypto/tls"
	"errors"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	freeConnections   int64
	reusedConnections int64
)

var dummyTLSConnectionState tls.ConnectionState

func AddFreeConnections(n int64) int64 {
	return atomic.AddInt64(&freeConnections, n)
}

func LoadFreeConnections(n int64) int64 {
	return atomic.LoadInt64(&freeConnections)
}

func AddReusedConnections(n int64) int64 {
	return atomic.AddInt64(&reusedConnections, n)
}

func LoadReusedConnections(n int64) int64 {
	return atomic.LoadInt64(&reusedConnections)
}

type Conn struct {
	lasted  int64
	created int64
	used    int64
	conn    net.Conn
}

func New(conn net.Conn, now int64) *Conn {
	return &Conn{
		conn:    conn,
		created: now,
		lasted:  now,
		used:    1,
	}
}

// filer describes an object that has ability to return os.File.
type filer interface {
	// File returns a copy of object's file descriptor.
	File() (*os.File, error)
}

// File returns a copy of object's file descriptor.
func (s *Conn) File() (*os.File, error) {
	if sf, ok := s.conn.(filer); ok {
		return sf.File()
	}

	return nil, errors.New("not implement filer interface")
}

func (s *Conn) SyscallConn() (syscall.RawConn, error) {
	if sf, ok := s.conn.(syscall.Conn); ok {
		return sf.SyscallConn()
	}

	return nil, errors.New("not implement syscall.Conn interface")
}

func (s *Conn) IncrUsed() {
	AddReusedConnections(1)
	atomic.AddInt64(&s.used, 1)
	atomic.StoreInt64(&s.lasted, time.Now().Unix())
}

func (s *Conn) GetRawConn() net.Conn {
	return s.conn
}

func (s *Conn) GetUsed() int64 {
	return atomic.LoadInt64(&s.used)
}

func (s *Conn) GetCreated() int64 {
	return atomic.LoadInt64(&s.created)
}

func (s *Conn) GetLasted() int64 {
	return atomic.LoadInt64(&s.lasted)
}

func (s *Conn) Read(b []byte) (n int, err error) {
	return s.conn.Read(b)
}

func (s *Conn) Write(b []byte) (n int, err error) {
	return s.conn.Write(b)
}

func (s *Conn) Close() error {
	return s.conn.Close()
}

func (s *Conn) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Conn) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Conn) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

func (s *Conn) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *Conn) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

type TLSConnectionStater interface {
	ConnectionState() tls.ConnectionState
}

func (s *Conn) ConnectionState() tls.ConnectionState {
	if ts, ok := s.conn.(TLSConnectionStater); ok {
		return ts.ConnectionState()
	}

	return dummyTLSConnectionState
}
