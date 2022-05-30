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

// Package delayconn implements many smart net.Conn.
package delayconn

import (
	"io"
	"net"
	"time"

	"testing/iotest"
)

// OneByteWriteConn guarantees write a one bytes every time.
type OneByteWriteConn struct {
	w    io.Writer
	conn net.Conn
}

// NewOneByteWriteConn creates a new OneByteWriteConn.
func NewOneByteWriteConn(conn net.Conn) *OneByteWriteConn {
	return &OneByteWriteConn{
		conn: conn,
		w:    OneByteWriter(conn),
	}
}

// Write Writes data from the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *OneByteWriteConn) Read(b []byte) (n int, err error) {
	return rc.conn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *OneByteWriteConn) Write(b []byte) (n int, err error) {
	return rc.w.Write(b)
}

// Close closes the connection.
// Any blocked Write or Write operations will be unblocked and return errors.
func (rc *OneByteWriteConn) Close() error {
	return rc.conn.Close()
}

// LocalAddr returns the local network address.
func (rc *OneByteWriteConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (rc *OneByteWriteConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

// SetDeadline sets the Write and write deadlines associated
// with the connection. It is equivalent to calling both
// SetWriteDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Write or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Write or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (rc *OneByteWriteConn) SetDeadline(t time.Time) error {
	return rc.conn.SetDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// A zero value for t means Write will not time out.
func (rc *OneByteWriteConn) SetWriteDeadline(t time.Time) error {
	return rc.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
func (rc *OneByteWriteConn) SetReadDeadline(t time.Time) error {
	return rc.conn.SetReadDeadline(t)
}

// OneByteReadConn sets the delay for read operations.
type OneByteReadConn struct {
	reader io.Reader
	conn   net.Conn
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (rc *OneByteReadConn) Read(b []byte) (n int, err error) {
	return rc.reader.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *OneByteReadConn) Write(b []byte) (n int, err error) {
	return rc.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (rc *OneByteReadConn) Close() error {
	return rc.conn.Close()
}

// LocalAddr returns the local network address.
func (rc *OneByteReadConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (rc *OneByteReadConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (rc *OneByteReadConn) SetDeadline(t time.Time) error {
	return rc.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (rc *OneByteReadConn) SetReadDeadline(t time.Time) error {
	return rc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (rc *OneByteReadConn) SetWriteDeadline(t time.Time) error {
	return rc.conn.SetWriteDeadline(t)
}

func NewOneByteReadConn(conn net.Conn) *OneByteReadConn {
	return &OneByteReadConn{
		reader: iotest.OneByteReader(conn),
	}
}

// ReadDelayConn sets the read delay operations.
type ReadDelayConn struct {
	delay time.Duration
	conn  net.Conn
}

// NewReadDelayConn creates a new ReadDelayConn.
func NewReadDelayConn(delay time.Duration, conn net.Conn) *ReadDelayConn {
	return &ReadDelayConn{
		delay: delay,
		conn:  conn,
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (rc *ReadDelayConn) Read(b []byte) (n int, err error) {
	if rc.delay > 0 {
		time.Sleep(rc.delay)
	}
	return rc.conn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *ReadDelayConn) Write(b []byte) (n int, err error) {
	return rc.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (rc *ReadDelayConn) Close() error {
	return rc.conn.Close()
}

// LocalAddr returns the local network address.
func (rc *ReadDelayConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (rc *ReadDelayConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (rc *ReadDelayConn) SetDeadline(t time.Time) error {
	return rc.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (rc *ReadDelayConn) SetReadDeadline(t time.Time) error {
	return rc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (rc *ReadDelayConn) SetWriteDeadline(t time.Time) error {
	return rc.conn.SetWriteDeadline(t)
}

// WriteDelayConn implements the delay before Write.
type WriteDelayConn struct {
	delay time.Duration
	conn  net.Conn
}

// NewWriteDelayConn returns a new WriteDelayConn.
func NewWriteDelayConn(delay time.Duration, conn net.Conn) *WriteDelayConn {
	return &WriteDelayConn{
		delay: delay,
		conn:  conn,
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (rc *WriteDelayConn) Read(b []byte) (n int, err error) {
	return rc.conn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *WriteDelayConn) Write(b []byte) (n int, err error) {
	if rc.delay > 0 {
		time.Sleep(rc.delay)
	}
	return rc.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (rc *WriteDelayConn) Close() error {
	return rc.conn.Close()
}

// LocalAddr returns the local network address.
func (rc *WriteDelayConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (rc *WriteDelayConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (rc *WriteDelayConn) SetDeadline(t time.Time) error {
	return rc.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (rc *WriteDelayConn) SetReadDeadline(t time.Time) error {
	return rc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (rc *WriteDelayConn) SetWriteDeadline(t time.Time) error {
	return rc.conn.SetWriteDeadline(t)
}

// DelayConn delay the read and write operations.
type DelayConn struct {
	delay time.Duration
	conn  net.Conn
}

// NewDelayConn creates a new DelayConn.
func NewDelayConn(delay time.Duration, conn net.Conn) *DelayConn {
	return &DelayConn{
		delay: delay,
		conn:  conn,
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (rc *DelayConn) Read(b []byte) (n int, err error) {
	if rc.delay > 0 {
		time.Sleep(rc.delay)
	}
	return rc.conn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *DelayConn) Write(b []byte) (n int, err error) {
	if rc.delay > 0 {
		time.Sleep(rc.delay)
	}
	return rc.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (rc *DelayConn) Close() error {
	return rc.conn.Close()
}

// LocalAddr returns the local network address.
func (rc *DelayConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (rc *DelayConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (rc *DelayConn) SetDeadline(t time.Time) error {
	return rc.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (rc *DelayConn) SetReadDeadline(t time.Time) error {
	return rc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (rc *DelayConn) SetWriteDeadline(t time.Time) error {
	return rc.conn.SetWriteDeadline(t)
}

// PerWriteDelayConn implements the delay before Write.
type PerWriteDelayConn struct {
	delay time.Duration
	conn  net.Conn
}

// NewPerWriteDelayConn returns a new PerWriteDelayConn.
func NewPerWriteDelayConn(delay time.Duration, conn net.Conn) *PerWriteDelayConn {
	return &PerWriteDelayConn{
		delay: delay,
		conn:  conn,
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (rc *PerWriteDelayConn) Read(b []byte) (n int, err error) {
	return rc.conn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (rc *PerWriteDelayConn) Write(b []byte) (n int, err error) {
	if rc.delay > 0 {
		var (
			nw  int
			err error
		)
		for i := range b {
			time.Sleep(rc.delay)
			nw, err = rc.conn.Write(b[i : i+1])
			n += nw
			if err != nil {
				return nw, err
			}
		}
		return n, nil
	}
	return rc.conn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (rc *PerWriteDelayConn) Close() error {
	return rc.conn.Close()
}

// LocalAddr returns the local network address.
func (rc *PerWriteDelayConn) LocalAddr() net.Addr {
	return rc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (rc *PerWriteDelayConn) RemoteAddr() net.Addr {
	return rc.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (rc *PerWriteDelayConn) SetDeadline(t time.Time) error {
	return rc.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (rc *PerWriteDelayConn) SetReadDeadline(t time.Time) error {
	return rc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (rc *PerWriteDelayConn) SetWriteDeadline(t time.Time) error {
	return rc.conn.SetWriteDeadline(t)
}

type oneByteWriter struct {
	w io.Writer
}

// OneByteWriter returns a writer
func OneByteWriter(w io.Writer) io.Writer {
	return &oneByteWriter{w: w}
}

func (w *oneByteWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return w.w.Write(p[0:1])
}
