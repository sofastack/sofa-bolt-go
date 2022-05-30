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
	"errors"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jpillora/backoff"
	"github.com/sofastack/sofa-bolt-go/sofabolt/conn/asyncwriteconn"
	"github.com/sofastack/sofa-bolt-go/sofabolt/conn/errorconn"
)

type ClientConnStatusChanger interface {
	OnStatusChange(cc *ClientConn, from, to ClientConnStatus)
}

type ClientConnStatusChangerFunc func(cc *ClientConn, from, to ClientConnStatus)

func (cs ClientConnStatusChangerFunc) OnStatusChange(cc *ClientConn, from, to ClientConnStatus) {
	cs(cc, from, to)
}

type ClientConnDispatcher interface {
	Dispatch(err error, cmd interface{})
}

type ClientConnDispatcherFunc func(err error, cmd interface{})

func (cd ClientConnDispatcherFunc) Dispatch(err error, cmd interface{}) {
	cd(err, cmd)
}

type ClientConnOptions struct {
	idleTimeout        time.Duration
	readTimeout        time.Duration
	writeTimeout       time.Duration
	flushInterval      time.Duration
	blockmode          bool
	dialer             Dialer
	statusChanger      ClientConnStatusChanger
	dispatcher         ClientConnDispatcher
	maxPendingCommands int
}

type ClientConnDoer interface {
	Send(o interface{}) error
	Dispatch(err error, cmd interface{})
	GetStatus() ClientConnStatus
	IncrementID() uint64
	Close() error
}

var _ ClientConnDoer = (*ClientConn)(nil)

type ClientConn struct {
	enc      ClientConnProtocolEncoder
	dec      ClientConnProtocolDecoder
	options  ClientConnOptions
	connLock sync.RWMutex
	conn     net.Conn
	closed   int32
	id       uint64
	status   ClientConnStatus
	metrics  *ClientMetrics
}

func NewClientConn(options ...ClientConnOptionSetter) (*ClientConn, error) {
	c := &ClientConn{}

	for _, option := range options {
		option.Set(c)
	}

	if err := c.polyfill(); err != nil {
		return nil, err
	}

	go c.doread()

	return c, nil
}

func (x *ClientConn) polyfill() error {
	if x.options.dispatcher == nil {
		return errors.New("sofabolt: dispatcher cannot be nil")
	}

	if x.options.maxPendingCommands == 0 {
		x.options.maxPendingCommands = runtime.NumCPU() * 16
	}

	if x.metrics == nil {
		x.metrics = &ClientMetrics{
			created: time.Now().Unix(),
		}
	}

	x.OnStatusChange(IdleClientConnStatus)

	if x.conn == nil {
		if x.options.dialer == nil {
			return errors.New("sofabolt: client connection and dialer is nil")
		}

		x.OnStatusChange(ConnectingClientConnStatus)

		conn, err := x.options.dialer.Dial()
		if err != nil {
			x.setConn(errorconn.New(
				err,
			))
		} else {
			x.setConn(conn)
		}

	} else {
		conn, err := x.buildAsyncWriteConn(x.conn)
		if err != nil {
			return err
		}
		x.setConn(conn)
	}

	return nil
}

func (x *ClientConn) OnStatusChange(to ClientConnStatus) {
	from := ClientConnStatus(atomic.LoadUint32((*uint32)(&x.status)))
	if from == to { // skip if do not change
		return
	}
	atomic.StoreUint32((*uint32)(&x.status), uint32(to))
	if x.options.statusChanger != nil {
		x.options.statusChanger.OnStatusChange(x, from, to)
	}
}

func (x *ClientConn) GetMetrics() *ClientMetrics { return x.metrics }

func (x *ClientConn) Send(o interface{}) error {
	var (
		err error
		eo  = NewClientConnProtocolEncoderOption()
		b   = acquireBytes()
	)
	defer releaseBytes(b)

	*b, err = x.enc.Encode(eo, (*b)[:0], o)
	if err != nil {
		return err
	}
	_, err = x.Write(*b)
	return err
}

func (x *ClientConn) Write(b []byte) (int, error) {
	return x.GetConn().Write(b)
}

func (x *ClientConn) GetStatus() ClientConnStatus {
	return ClientConnStatus(atomic.LoadUint32((*uint32)(&x.status)))
}

func (x *ClientConn) IncrementID() uint64 { return atomic.AddUint64(&x.id, 1) }

func (x *ClientConn) Closed() bool {
	return atomic.LoadInt32(&x.closed) == 1
}

func (x *ClientConn) GetConn() net.Conn {
	return x.getConn()
}

func (x *ClientConn) Close() error {
	if !atomic.CompareAndSwapInt32(&x.closed, 0, 1) {
		return ErrClientWasClosed
	}
	return x.getConn().Close()
}

func (x *ClientConn) doread() {
	defer x.OnStatusChange(ShutdownClientConnStatus)

	var (
		newconn net.Conn
		dialerr error
		conn    = x.GetConn()
		br      = acquireBufioReader(conn)
		wg      sync.WaitGroup
		err     error
		ok      bool
		cmd     interface{}
		do      = NewClientConnProtocolDecoderOption()
	)
	beforeread := func() error {
		if err = x.setConnReadTimeout(conn); err != nil {
			return err
		}
		x.OnStatusChange(ActiveClientConnStatus)
		return nil
	}
	br.InstallBeforeReadHook(beforeread)

	x.OnStatusChange(ActiveClientConnStatus)

READLOOP:
	for {
		cmd, err = x.dec.Decode(do, br)
		if err != nil {
			// only process deadline error
			// if conn is error conn, we need redial
			if _, match := conn.(*errorconn.Conn); match {
				// dispatch the conn error
				x.DoDispatch(&wg, err, cmd)
				break
			}

			// if error is timeout error, continue read
			if IsDeadlineError(err) {
				x.OnStatusChange(ReadTimeoutClientConnStatus)
				continue
			}
		}

		// continue dispatch
		x.DoDispatch(&wg, err, cmd)
		if err != nil { // See any error then break read loop
			break
		}
	}

	x.OnStatusChange(TransientFailureClientConnStatus)

	newconn, dialerr, ok = x.mayRedial()
	if ok {
		if dialerr == nil {
			// close the old connection and ignore the error
			// nolint
			conn.Close()
			conn = newconn
			x.setConn(newconn)
			br.Reset(newconn)
			x.OnStatusChange(ActiveClientConnStatus)

			goto READLOOP
		}
	}

	wg.Wait() // wait all pending goroutines done
	releaseBufioReader(br)
	_ = x.GetConn().Close()
}

func (x *ClientConn) DoDispatch(wg *sync.WaitGroup, err error, cmd interface{}) {
	if !x.options.blockmode {
		x.Dispatch(err, cmd)
	} else {
		wg.Add(1)
		go x.dispatchAsync(wg, err, cmd)
	}
}

func (x *ClientConn) Dispatch(err error, cmd interface{}) {
	x.options.dispatcher.Dispatch(err, cmd)
}

func (x *ClientConn) dispatchAsync(wg *sync.WaitGroup, err error, cmd interface{}) {
	x.options.dispatcher.Dispatch(err, cmd)
	wg.Done()
}

func (x *ClientConn) redial() (net.Conn, error) {
	if x.options.dialer == nil {
		return nil, errors.New("sofabolt: disable redial")
	}

	retry := &backoff.Backoff{
		Min:    1 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	x.OnStatusChange(ConnectingClientConnStatus)
	// retry unit see success or client closed
	for {
		if x.Closed() {
			return nil, ErrClientWasClosed
		}

		time.Sleep(retry.Duration())
		conn, err := x.options.dialer.Dial()
		if err != nil {
			continue
		}

		return conn, nil
	}
}

func (x *ClientConn) mayRedial() (net.Conn, error, bool) {
	if x.Closed() {
		return nil, nil, false
	}

	if x.options.dialer == nil {
		return nil, nil, false
	}

	conn, err := x.redial()
	if err != nil { // restart to read
		return nil, nil, false
	}

	conn, err = x.buildAsyncWriteConn(conn)
	return conn, err, true
}

func (x *ClientConn) buildAsyncWriteConn(conn net.Conn) (net.Conn, error) {
	option := asyncwriteconn.NewOption()
	option.SetTimeout(x.options.writeTimeout)
	option.SetFlushInterval(x.options.flushInterval)
	option.SetBatch(x.options.maxPendingCommands)
	metrics := asyncwriteconn.NewMetrics()
	metrics.SetCommands(&x.metrics.commands)
	x.metrics.ResetPendingCommands()
	metrics.SetPendingCommands(&x.metrics.pendingCommands)
	metrics.SetBytes(&x.metrics.nwrite)
	return asyncwriteconn.New(conn,
		asyncwriteconn.WithOption(option),
		asyncwriteconn.WithMetrics(metrics),
	)
}

func (x *ClientConn) setConnReadTimeout(conn net.Conn) error {
	if x.options.readTimeout > 0 {
		return conn.SetReadDeadline(time.Now().Add(x.options.readTimeout))
	}
	return nil
}

func (x *ClientConn) getConn() net.Conn {
	x.connLock.RLock()
	conn := x.conn
	x.connLock.RUnlock()
	return conn
}

func (x *ClientConn) setConn(conn net.Conn) {
	x.connLock.Lock()
	x.conn = conn
	x.connLock.Unlock()
}
