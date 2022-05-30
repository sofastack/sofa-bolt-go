package bolt

import (
	"errors"
	"math"
	"net"
	"sync"
	"time"

	"github.com/sofastack/sofa-bolt-go/sofabolt"
)

var (
	zeroTimer    time.Timer
	zeroDuration time.Duration
)

type ClientConnBOLT struct {
	sync.RWMutex
	x        *sofabolt.ClientConn
	requests map[uint32]*sofabolt.InvokeContext
	conn     net.Conn
	options  struct {
		idleTimeout        time.Duration
		readTimeout        time.Duration
		writeTimeout       time.Duration
		flushInterval      time.Duration
		maxPendingCommands int
		onHeartbeat        func(success bool)
		heartbeatInterval  time.Duration
		heartbeatTimeout   time.Duration
		heartbeatProbes    int
		dialer             sofabolt.Dialer
		handler            sofabolt.Handler
		metrics            *sofabolt.ClientMetrics
	}
}

func NewClientConnBOLT(options ...ClientConnBOLTOptionSetter) (*ClientConnBOLT, error) {
	xb := &ClientConnBOLT{}

	for _, option := range options {
		option.Set(xb)
	}

	if err := xb.polyfill(); err != nil {
		return nil, err
	}

	return xb, nil
}

func (xb *ClientConnBOLT) polyfill() error {
	xb.requests = make(map[uint32]*sofabolt.InvokeContext, 256)

	if xb.options.metrics == nil {
		xb.options.metrics = &sofabolt.ClientMetrics{}
	}

	opts := make([]sofabolt.ClientConnOptionSetter, 0, 64)

	if xb.conn != nil {
		opts = append(opts, sofabolt.WithClientConnConn(xb.conn))
	} else {
		opts = append(opts, sofabolt.WithClientConnRedial(xb.options.dialer))
	}

	opts = append(opts, sofabolt.WithClientConnMaxPendingCommands(xb.options.maxPendingCommands))

	opts = append(opts, sofabolt.WithClientConnTimeout(
		xb.options.readTimeout,
		xb.options.writeTimeout,
		xb.options.idleTimeout,
		xb.options.flushInterval,
	))

	opts = append(opts, sofabolt.WithClientConnMaxPendingCommands(
		xb.options.maxPendingCommands,
	))

	opts = append(opts, sofabolt.WithClientConnMetrics(
		xb.options.metrics,
	))

	opts = append(opts, sofabolt.WithClientConnDispatcher(
		xb,
	))

	opts = append(opts,
		sofabolt.WithClientConnProtocolDecoder(&BOLTProtocolDecoder{}),
		sofabolt.WithClientConnProtocolEncoder(&BOLTProtocolEncoder{}),
	)

	x, err := sofabolt.NewClientConn(opts...)
	if err != nil {
		return err
	}
	xb.x = x
	return nil
}

func (xb *ClientConnBOLT) Dispatch(err error, cmd interface{}) {
	if err != nil {
		xb.cleanupPendingRequests(err)
		return
	}

	switch x := cmd.(type) {
	case *sofabolt.Command:
		if x.IsRequest() {
			var req sofabolt.Request
			// nolint
			req.ShallowCopyCommand(x)
			xb.dispatchRequest(&req)
		} else {
			var res sofabolt.Response
			res.ShallowCopyCommand(x)
			xb.dispatchResponse(&res)
		}
	default:
		// unknown command
	}
}

func (xb *ClientConnBOLT) dispatchRequest(req *sofabolt.Request) {
	if xb.options.handler == nil {
		return
	}

	crw := sofabolt.AcquireSofaResponseWriter(xb.x.GetConn(), xb.x)
	crw.Derive(req)

	xb.options.handler.ServeSofaBOLT(crw, req)

	sofabolt.ReleaseSofaResponseWriter(crw)
}

func (xb *ClientConnBOLT) dispatchResponse(res *sofabolt.Response) {
	ictx, ok := xb.getAndDelRequestContext(res.GetRequestID())
	if !ok {
		// We've not got any pending request. It usually means that
		// Write partially failed (timeout or request oneway),
		// and request was already removed;
	} else {
		ictx.Invoke(nil, res)
	}
}

func (xb *ClientConnBOLT) DoCallback(req *sofabolt.Request, cb sofabolt.ClientCallbacker) error {
	return xb.DoCallbackTimeout(req, cb, zeroDuration)
}

func (xb *ClientConnBOLT) DoCallbackTimeout(req *sofabolt.Request,
	cb sofabolt.ClientCallbacker, timeout time.Duration) error {
	xb.x.GetMetrics().AddReferences(1)
	xb.x.GetMetrics().AddUsed(1)

	// Do not allocate from sync.pool: it will be easily GC
	ictx := sofabolt.NewInvokeContext(req).SetCallback(cb).SetTimeout(timeout)

	err := xb.invoke(ictx, 0)

	xb.x.GetMetrics().AddReferences(-1)
	xb.x.GetMetrics().SetLasted()

	return err
}

func (xb *ClientConnBOLT) Do(req *sofabolt.Request, res *sofabolt.Response) error {
	return xb.DoTimeout(req, res, 0)
}

func (xb *ClientConnBOLT) DoTimeout(req *sofabolt.Request, res *sofabolt.Response, timeout time.Duration) error {
	xb.x.GetMetrics().AddReferences(1)
	xb.x.GetMetrics().AddUsed(1)

	ictx := sofabolt.AcquireInvokeContext(req, res, timeout)
	err := xb.invoke(ictx, timeout)
	if err != sofabolt.ErrClientTimeout { // let gc handle it if it it's not timeoutd.
		sofabolt.ReleaseInvokeContext(ictx)
	}

	xb.x.GetMetrics().SetLasted()
	xb.x.GetMetrics().AddReferences(-1)

	return err
}

func (xb *ClientConnBOLT) invoke(ctx *sofabolt.InvokeContext, timeout time.Duration) error {
	if xb.x.Closed() {
		return sofabolt.ErrClientWasClosed
	}

	id := xb.x.IncrementID()
	if id > math.MaxUint32 {
		return errors.New("client: request id overflow")
	}
	rid := uint32(id)
	ctx.GetRequest().SetRequestID(rid)
	xb.addRequestContext(rid, ctx)

	err := xb.x.Send(ctx.GetRequest())
	if err != nil {
		xb.delRequestContext(rid)
		return err
	}

	if ctx.GetRequest().GetType() == sofabolt.TypeBOLTRequestOneWay { // one way
		xb.delRequestContext(rid)
		return nil
	}

	if ch := ctx.GetErrorCh(); ch != nil {
		timer := &zeroTimer
		if timeout != 0 {
			timer = sofabolt.AcquireTimer(timeout)
			defer sofabolt.ReleaseTimer(timer)

		} else {
			timer = &zeroTimer
		}

		// wait a response
		select {
		// errCh is is guaranteed to see any ctx write after receiving on errCh completes
		case err = <-ch:
			if err != nil {
				return err
			}

			// Copy context response to client response
			ctx.CopyResponse(ctx.GetResponse())

			return nil

		case <-timer.C:
			xb.delRequestContext(rid)

			return sofabolt.ErrClientTimeout
		}
	}

	return nil
}

func (xb *ClientConnBOLT) cleanupPendingRequests(err error) {
	// cleanup pending requests at first
	xb.Lock()
	for i := range xb.requests {
		ictx := xb.requests[i]
		if ictx.GetCallback() != nil {
			ictx.GetCallback().Invoke(err, ictx)
		} else {
			select { // sanity send
			case ictx.GetErrorCh() <- err:
			default:
			}
		}
	}
	// clear the pending requests
	for id := range xb.requests {
		delete(xb.requests, id)
	}
	xb.Unlock()
}

func (xb *ClientConnBOLT) getAndDelRequestContext(rid uint32) (*sofabolt.InvokeContext, bool) {
	xb.Lock()
	ictx, ok := xb.requests[rid]
	if ok {
		delete(xb.requests, rid)
	}
	xb.Unlock()
	return ictx, ok
}

func (xb *ClientConnBOLT) addRequestContext(rid uint32, ictx *sofabolt.InvokeContext) {
	xb.Lock()
	xb.requests[rid] = ictx
	xb.Unlock()
}

func (xb *ClientConnBOLT) delRequestContext(rid uint32) {
	xb.Lock()
	delete(xb.requests, rid)
	xb.Unlock()
}
