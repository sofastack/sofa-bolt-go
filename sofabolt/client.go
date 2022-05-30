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
	uatomic "go.uber.org/atomic"
)

var (
	zeroTime     time.Time
	zeroTimer    time.Timer
	zeroDuration time.Duration
)

type Dialer interface {
	Dial() (net.Conn, error)
}

type DialerFunc func() (net.Conn, error)

func (d DialerFunc) Dial() (net.Conn, error) {
	return d()
}

type ClientCallbacker interface {
	Invoke(error, *InvokeContext)
}

type ClientCallbackerFunc func(error, *InvokeContext)

func (c ClientCallbackerFunc) Invoke(err error, cctx *InvokeContext) {
	c(err, cctx)
}

type Client struct {
	sync.RWMutex
	requests map[uint32]*InvokeContext

	options struct {
		disableAutoIncrementRequestID bool
		idleTimeout                   time.Duration
		readTimeout                   time.Duration
		writeTimeout                  time.Duration
		flushInterval                 time.Duration
		maxPendingCommands            int
		onHeartbeat                   func(success bool)
		heartbeatInterval             time.Duration
		heartbeatTimeout              time.Duration
		heartbeatProbes               int
		handler                       Handler
		dialer                        Dialer
	}

	rid      uint32
	connLock sync.RWMutex
	conn     net.Conn
	metrics  *ClientMetrics
	closed   int32
	rerr     uatomic.Error
	rerrCh   chan error
}

func NewClient(options ...ClientOptionSetter) (*Client, error) {
	c := &Client{
		rerrCh: make(chan error, 1),
	}

	for _, option := range options {
		option.Set(c)
	}

	if err := c.polyfill(); err != nil {
		return nil, err
	}

	// nolint
	go c.doread()

	if c.options.heartbeatTimeout > 0 || c.options.heartbeatInterval > 0 ||
		c.options.heartbeatProbes > 0 || c.options.onHeartbeat != nil {
		go c.doheartbeat()
	}

	return c, nil
}

func (c *Client) polyfill() error {
	if c.options.maxPendingCommands == 0 {
		c.options.maxPendingCommands = runtime.NumCPU() * 16
	}

	if c.metrics == nil {
		c.metrics = &ClientMetrics{
			created: time.Now().Unix(),
		}
	}

	if c.conn == nil {
		if c.options.dialer == nil {
			return errors.New("sofabolt: client connection and dialer is nil")
		}

		conn, err := c.options.dialer.Dial()
		if err != nil {
			c.setConn(errorconn.New(
				err,
			))
		} else {
			c.setConn(conn)
		}

	} else {
		conn, err := c.buildAsyncWriteConn(c.conn)
		if err != nil {
			return err
		}
		c.setConn(conn)
	}

	c.requests = make(map[uint32]*InvokeContext, c.options.maxPendingCommands)

	return nil
}

func (c *Client) GetConn() net.Conn {
	return c.getConn()
}

func (c *Client) Closed() bool {
	return atomic.LoadInt32(&c.closed) == 1
}

func (c *Client) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return ErrClientWasClosed
	}
	return c.getConn().Close()
}

func (c *Client) GetMetrics() *ClientMetrics {
	return c.metrics
}

func (c *Client) DoCallback(req *Request, cb ClientCallbacker) error {
	return c.DoCallbackTimeout(req, cb, zeroDuration)
}

func (c *Client) DoCallbackTimeout(req *Request,
	cb ClientCallbacker, timeout time.Duration) error {
	atomic.AddInt64(&c.metrics.references, 1)

	// Do not allocate from sync.pool: it will be easily GC
	ictx := &InvokeContext{
		req:      req,
		callback: cb,
		created:  time.Now(),
		timeout:  timeout,
	}

	err := c.invoke(ictx, 0)

	atomic.StoreInt64(&c.metrics.lasted, time.Now().Unix())
	atomic.AddInt64(&c.metrics.used, 1)
	atomic.AddInt64(&c.metrics.references, -1)

	return err
}

func (c *Client) Do(req *Request, res *Response) error {
	return c.DoTimeout(req, res, 0)
}

func (c *Client) DoTimeout(req *Request, res *Response, timeout time.Duration) error {
	atomic.AddInt64(&c.metrics.references, 1)
	atomic.AddInt64(&c.metrics.used, 1)

	ictx := c.AcquireInvokeContext(req, res, timeout)
	err := c.invoke(ictx, timeout)
	if err != ErrClientTimeout { // let gc handle it if it it's not timeoutd.
		c.ReleaseInvokeContext(ictx)
	}

	atomic.StoreInt64(&c.metrics.lasted, time.Now().Unix())
	atomic.AddInt64(&c.metrics.references, -1)

	return err
}

func (c *Client) doheartbeat() {
	timer := time.NewTimer(c.options.heartbeatInterval)
	defer timer.Stop()

	var probes int

	req := AcquireRequest()
	res := AcquireResponse()
	defer func() {
		ReleaseRequest(req)
		ReleaseResponse(res)
	}()
	req.SetCMDCode(CMDCodeBOLTHeartbeat)

	for {
		time.Sleep(c.options.heartbeatInterval)
		err := c.DoTimeout(req, res, c.options.heartbeatTimeout)
		if err != nil {
			probes++
			if probes > c.options.heartbeatProbes {
				if c.options.onHeartbeat != nil {
					c.options.onHeartbeat(false)
				}

				if c.options.heartbeatProbes > 0 {
					goto DONE
				}
			}
			continue
		}
		probes = 0
		if c.options.onHeartbeat != nil {
			c.options.onHeartbeat(true)
		}
	}
DONE:
}

func (c *Client) doread() error {
	var (
		br       = acquireBufioReader(c.GetConn())
		conn     = c.GetConn()
		res      Response
		req      Request
		cmd      Command
		crw      = acquireClientResponseWriter(c)
		ictx     *InvokeContext
		err      error
		ok       bool
		nr       int
		commands uint64
	)
	beforeread := func() error {
		if cmd.GetProto() > 0 {
			// Partial command, expect the rest of message arrives in certain timeout.
			return c.setConnReadTimeout(conn)
		} else {
			// Wait for next command message.
			return c.setConnIdleTimeout(conn)
		}
	}
	br.InstallBeforeReadHook(beforeread)

READLOOP:
	for {
		cmd.Reset()
		if nr, err = cmd.Read(&ReadOption{}, br); err != nil {
			break
		}
		atomic.AddInt64(&c.metrics.nread, int64(nr))
		commands++

		if cmd.IsRequest() {
			req.Reset()
			// nolint
			req.command = cmd
			crw.reset(c).Derive(&req)
			c.handleRequest(crw, &req)

		} else {
			res.Reset()
			// nolint
			res.command = cmd
			c.handleResponse(&res)
		}
	}

	// cleanup pending requests at first
	c.Lock()
	for i := range c.requests {
		ictx = c.requests[i]
		if ictx.callback != nil {
			ictx.callback.Invoke(err, ictx)
		} else {
			select { // sanity send
			case ictx.errCh <- err:
			default:
			}
		}
	}
	// clear the pending requests
	for id := range c.requests {
		delete(c.requests, id)
	}
	c.Unlock()

	var (
		newconn net.Conn
		dialerr error
	)

	newconn, dialerr, ok = c.mayRedial()
	if ok {
		if dialerr == nil {
			// close the old connection and ignore the error
			// nolint
			conn.Close()
			conn = newconn
			c.setConn(newconn)
			br.Reset(newconn)

			goto READLOOP
		}
	}

	releaseClientResponseWriter(crw)
	releaseBufioReader(br)

	_ = c.Close()
	// store and notify the read error
	c.rerr.Store(err)
	c.rerrCh <- err

	return err
}

func (c *Client) GetReadError() chan error {
	return c.rerrCh
}

func (c *Client) handleRequest(crw ResponseWriter, req *Request) {
	if c.options.handler == nil {
		return
	}

	c.options.handler.ServeSofaBOLT(crw, req)
}

func (c *Client) handleResponse(res *Response) {
	ictx, ok := c.getAndDelRequestContext(res.GetRequestID())
	if !ok {
		// We've not got any pending request. It usually means that
		// Write partially failed (timeout or request oneway),
		// and request was already removed;
	} else {
		ictx.Invoke(nil, res)
	}

	// TODO(detailyang): cleanup stale requests via deadline
}

func (c *Client) mayRedial() (net.Conn, error, bool) {
	if c.Closed() {
		return nil, nil, false
	}

	if c.options.dialer == nil {
		return nil, nil, false
	}

	conn, err := c.redial()
	if err != nil { // restart to read
		return nil, nil, false
	}

	conn, err = c.buildAsyncWriteConn(conn)
	return conn, err, true
}

func (c *Client) redial() (net.Conn, error) {
	if c.options.dialer == nil {
		return nil, errors.New("sofabolt: disable redial")
	}

	retry := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	// retry unit see success or client closed
	for {
		if c.Closed() {
			return nil, ErrClientWasClosed
		}

		time.Sleep(retry.Duration())
		conn, err := c.options.dialer.Dial()
		if err != nil {
			continue
		}

		return conn, nil
	}
}

func (c *Client) invoke(ctx *InvokeContext, timeout time.Duration) error {
	if atomic.LoadInt32(&c.closed) == 1 {
		return ErrClientWasClosed
	}

	if err := c.rerr.Load(); err != nil {
		return err
	}

	if ctx.req.GetType() != TypeBOLTRequest && ctx.req.GetType() != TypeBOLTRequestOneWay {
		return ErrClientNotARequest
	}

	rid := ctx.req.GetRequestID()
	if !c.options.disableAutoIncrementRequestID {
		rid = atomic.AddUint32(&c.rid, 1)
		ctx.req.SetRequestID(rid)
	}

	var (
		err error
		dst = acquireBytes()
	)

	*dst, err = ctx.req.Write(&WriteOption{}, (*dst)[:0])
	if err != nil {
		releaseBytes(dst)
		return err
	}

	c.addRequestContext(rid, ctx)
	_, err = c.write(*dst)
	releaseBytes(dst)
	if err != nil {
		c.delRequestContext(rid)
		return err
	}

	if ctx.req.GetType() == TypeBOLTRequestOneWay { // one way
		c.delRequestContext(rid)
		return nil
	}

	if ctx.errCh != nil {
		timer := &zeroTimer
		if timeout != 0 {
			timer = AcquireTimer(timeout)
			defer ReleaseTimer(timer)

		} else {
			timer = &zeroTimer
		}

		// wait a response
		select {
		// errCh is is guaranteed to see any ctx write after receiving on errCh completes
		case err = <-ctx.errCh:
			if err != nil {
				return err
			}

			// Copy context response to client response
			ctx.ireslock.Lock()
			ctx.ires.CopyTo(ctx.res)
			ctx.ireslock.Unlock()

			return nil

		case <-timer.C:
			c.delRequestContext(rid)

			return ErrClientTimeout
		}
	}

	return nil
}

func (c *Client) getAndDelRequestContext(rid uint32) (*InvokeContext, bool) {
	c.Lock()
	ictx, ok := c.requests[rid]
	if ok {
		delete(c.requests, rid)
	}
	c.Unlock()
	return ictx, ok
}

func (c *Client) addRequestContext(rid uint32, ictx *InvokeContext) {
	c.Lock()
	c.requests[rid] = ictx
	c.Unlock()
}

func (c *Client) delRequestContext(rid uint32) {
	c.Lock()
	delete(c.requests, rid)
	c.Unlock()
}

func (c *Client) write(d []byte) (int, error) {
	return c.GetConn().Write(d)
}

func (c *Client) setConnReadTimeout(conn net.Conn) error {
	if c.options.readTimeout > 0 {
		return conn.SetReadDeadline(time.Now().Add(c.options.readTimeout))
	}
	return nil
}

func (c *Client) setConnIdleTimeout(conn net.Conn) error {
	if c.options.idleTimeout > 0 {
		return conn.SetReadDeadline(time.Now().Add(c.options.idleTimeout))
	}
	return conn.SetReadDeadline(zeroTime)
}

func (c *Client) buildAsyncWriteConn(conn net.Conn) (net.Conn, error) {
	option := asyncwriteconn.NewOption()
	option.SetTimeout(c.options.writeTimeout)
	option.SetFlushInterval(c.options.flushInterval)
	option.SetBatch(c.options.maxPendingCommands)
	metrics := asyncwriteconn.NewMetrics()
	metrics.SetCommands(&c.metrics.commands)
	// must cleanup the pending commands metrics
	c.metrics.ResetPendingCommands()
	metrics.SetPendingCommands(&c.metrics.pendingCommands)
	metrics.SetBytes(&c.metrics.nwrite)
	return asyncwriteconn.New(conn,
		asyncwriteconn.WithOption(option),
		asyncwriteconn.WithMetrics(metrics),
	)
}

func (c *Client) getConn() net.Conn {
	c.connLock.RLock()
	conn := c.conn
	c.connLock.RUnlock()
	return conn
}

func (c *Client) setConn(conn net.Conn) {
	c.connLock.Lock()
	c.conn = conn
	c.connLock.Unlock()
}

func (c *Client) ReleaseInvokeContext(ictx *InvokeContext) {
	ReleaseInvokeContext(ictx)
}

func (c *Client) AcquireInvokeContext(req *Request, res *Response, timeout time.Duration) *InvokeContext {
	return AcquireInvokeContext(req, res, timeout)
}
