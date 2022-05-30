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
    "context"
    "io"
    "net"
    "sync"
    "sync/atomic"
    "time"

    stateconn "github.com/sofastack/sofa-bolt-go/sofabolt/conn/stateconn"
    workerpool "github.com/sofastack/sofa-common-go/syncpool/fast-workerpool"
    bufiorw "github.com/sofastack/sofa-common-go/writer/bufiorw"
)

const (
    defaultShutdownPollInterval = 500 * time.Millisecond
    defaultConnIdleThreshold    = 5 * time.Second
)

type Server struct {
    sync.Mutex
    listeners map[net.Listener]struct{}
    conns     map[net.Conn]struct{}
    servepool *workerpool.WorkerPool

    handler   Handler
    onhandler ServerOnEventHandler

    options struct {
        async             bool
        readTimeout       time.Duration
        writeTimeout      time.Duration
        idleTimeout       time.Duration
        flushInterval     time.Duration
        maxPendingCommand int
        maxConnections    int
    }

    metrics *ServerMetrics
}

func NewServer(options ...serverOptionSetter) (*Server, error) {
    srv := &Server{
        listeners: make(map[net.Listener]struct{}, 4),
    }

    for _, op := range options {
        op.set(srv)
    }

    if err := srv.polyfill(); err != nil {
        return nil, err
    }

    // startup worker pool
    srv.servepool.Start()

    return srv, nil
}

func (srv *Server) GetMetrics() *ServerMetrics { return srv.metrics }

func (srv *Server) polyfill() error {
    if srv.handler == nil {
        return ErrServerHandler
    }

    if srv.onhandler == nil {
        srv.onhandler = DummyServerOnEventHandler
    }

    if srv.options.maxConnections == 0 {
        srv.options.maxConnections = 10240
    }

    if srv.metrics == nil {
        srv.metrics = &ServerMetrics{}
    }

    srv.conns = make(map[net.Conn]struct{}, srv.options.maxConnections)

    // worker pool
    var err error
    srv.servepool, err = workerpool.New(srv,
        workerpool.WithWorkerPoolMaxWorkersCount(srv.options.maxConnections))
    if err != nil {
        return err
    }

    return nil
}

func (srv *Server) ServeJob(v interface{}) {
    conn, ok := v.(net.Conn)
    if !ok {
        panic("failed to type casting")
    }
    // nolint
    srv.ServeConn(conn)
}

func (srv *Server) Serve(ln net.Listener) error {
    srv.addListener(ln)
    defer srv.delListener(ln)

    var (
        rawc  net.Conn
        delay time.Duration
        err   error
    )

    for {
        rawc, err = ln.Accept()
        if err != nil {
            if ne, ok := err.(net.Error); ok && ne.Temporary() { // temporary: try later
                if delay == 0 {
                    delay = 5 * time.Millisecond
                } else {
                    delay *= 2
                }

                if max := 1 * time.Second; delay > max {
                    delay = max
                }

                srv.onhandler(srv, err, NewServerEventContext(ServerTemporaryAcceptEvent).
                    SetConn(rawc))
                time.Sleep(delay)
                continue
            }
            break
        }

        if !srv.servepool.Serve(rawc) {
            srv.onhandler(srv, err, NewServerEventContext(ServerWorkerPoolOverflowEvent).
                SetConn(rawc))
        }
    }

    return err
}

// ServeConn serves a net.Conn
func (srv *Server) ServeConn(conn net.Conn) error {
    srv.metrics.addConnections(1)
    srv.metrics.addPendingConnections(1)
    sc := stateconn.AcquireConn(conn)
    srv.addConn(sc)

    hijacked, err := srv.serveConn(sc)
    if hijacked {
        srv.onhandler(srv, nil, NewServerEventContext(ServerConnHijackedEvent).SetConn(conn))
    } else {
        if err != nil && err != io.EOF {
            srv.onhandler(srv, err, NewServerEventContext(ServerConnErrorEvent).SetConn(conn))
        }
        // nolint
        sc.Close() // discard close error
        stateconn.ReleaseConn(sc)
    }

    srv.delConn(sc)
    atomic.AddInt64(&srv.metrics.pendingconnections, -1)

    return err
}

func (srv *Server) setConnReadTimeout(conn net.Conn) error {
    if srv.options.readTimeout > 0 {
        return conn.SetReadDeadline(time.Now().Add(srv.options.readTimeout))
    }
    return nil
}

func (srv *Server) setConnIdleTimeout(conn net.Conn) error {
    if srv.options.idleTimeout > 0 {
        return conn.SetReadDeadline(time.Now().Add(srv.options.idleTimeout))
    }
    return nil
}

func (srv *Server) serveConn(conn net.Conn) (hijacked bool, err error) {
    var (
        req           Request
        nr            int
        requests      uint64
        rw            *SofaResponseWriter
        wg            sync.WaitGroup
        lastFlushTime time.Time
    )

    br := acquireBufioReader(conn)
    bw := acquireBufioWriter(conn)
    rw = AcquireSofaResponseWriter(conn, bw)
    beforeread := func() error {
        if requests > 1 {
            if req.GetProto() > 0 {
                if err = srv.setConnReadTimeout(conn); err != nil {
                    return err
                }
            } else {
                if err = srv.setConnIdleTimeout(conn); err != nil {
                    return err
                }
            }
        } else {
            if err = srv.setConnReadTimeout(conn); err != nil {
                return err
            }
        }

        if bw.Buffered() > 0 {
            if lastFlushTime, err = srv.flushWrite(conn, bw, lastFlushTime); err != nil {
                return err
            }
        }
        return nil
    }
    br.InstallBeforeReadHook(beforeread)

READLOOP:
    for {
        req.Reset()
        if nr, err = req.Read(&ReadOption{}, br); err != nil {
            break READLOOP
        }

        srv.metrics.addBytesRead(int64(nr))
        requests++

        if req.GetType() != TypeBOLTRequest &&
            req.GetType() != TypeBOLTRequestOneWay &&
            req.GetType() != TypeTBRemotingOneWay {
            err = ErrServerNotARequest
            break READLOOP
        }

        hijacked = srv.HandleCommand(&wg, conn, bw, rw, &req)

        if lastFlushTime, err = srv.flushWrite(conn, bw, lastFlushTime); err != nil {
            break READLOOP
        }

        if hijacked {
            break READLOOP
        }
    }

    // flush the remaining buffer
    if bw.Buffered() > 0 {
        // nolint
        srv.flushWrite(conn, bw, lastFlushTime)
    }

    wg.Wait() // wait all pending goroutines done

    releaseBufioReader(br)
    releaseBufioWriter(bw)
    ReleaseSofaResponseWriter(rw)

    return hijacked, err
}

func (srv *Server) flushWrite(conn net.Conn, bw *bufiorw.Writer, lastFlushTime time.Time) (time.Time, error) {
    now := time.Now()
    if srv.options.flushInterval > 0 {
        if now.Sub(lastFlushTime) >= srv.options.flushInterval {
            return now, srv.flush(conn, bw)
        }
    } else { // Flush always
        return now, srv.flush(conn, bw)
    }

    return now, nil
}

func (srv *Server) flush(conn net.Conn, bw *bufiorw.Writer) error {
    if srv.options.writeTimeout > 0 {
        if err := conn.SetWriteDeadline(time.Now().Add(srv.options.writeTimeout)); err != nil {
            return err
        }
    }
    return bw.Flush()
}

func (srv *Server) HandleCommand(wg *sync.WaitGroup, conn net.Conn, bw *bufiorw.Writer,
    rw *SofaResponseWriter, req *Request) bool {
    if !srv.options.async {
        return srv.handleCommandSync(bw, rw, req)
    }

    srv.handleCommandAsync(wg, conn, rw, req)
    return false
}

func (srv *Server) handleCommandAsync(wg *sync.WaitGroup, conn net.Conn, rw *SofaResponseWriter, raw *Request) {
    req := AcquireRequest()
    req.CopyCommand(&raw.command)

    wg.Add(1)
    go func(id uint64) {
        srv.doHandleCommandAsync(conn, id, req)
        ReleaseRequest(req)
        wg.Done()
    }(rw.id)
}

func (srv *Server) doHandleCommandAsync(conn net.Conn, id uint64, req *Request) {
    rw := AcquireSofaResponseWriter(conn, conn)
    rw.id = id
    rw.Derive(req)

    srv.serveCommand(rw, req)

    if rw.numwrite == 0 && req.GetType() != TypeBOLTRequestOneWay &&
        req.GetType() != TypeTBRemotingOneWay {
        // write once to avoid nil response
        // nolint
        rw.Write()
    } else {
        srv.metrics.addBytesWrite(int64(rw.numwrite))
    }

    ReleaseSofaResponseWriter(rw)
}

// nolint
func (srv *Server) handleCommandSync(bw *bufiorw.Writer, rw *SofaResponseWriter, req *Request) bool {
    rw.Reset(bw).Derive(req)
    srv.serveCommand(rw, req)
    if rw.numwrite == 0 && req.GetType() != TypeBOLTRequestOneWay &&
        req.GetType() != TypeTBRemotingOneWay {
        // write once to avoid nil response
        // nolint
        rw.Write()
    } else {
        srv.metrics.addBytesWrite(int64(rw.numwrite))
    }

    return rw.IsHijacked()
}

func (srv *Server) serveCommand(rw ResponseWriter, req *Request) {
    srv.metrics.addCommands(1)
    srv.metrics.addPendingCommands(1)

    srv.handler.ServeSofaBOLT(rw, req)

    srv.metrics.addPendingCommands(-1)
}

func (srv *Server) addListener(ln net.Listener) {
    srv.Lock()
    srv.listeners[ln] = struct{}{}
    srv.Unlock()
}

func (srv *Server) delListener(ln net.Listener) {
    srv.Lock()
    delete(srv.listeners, ln)
    srv.Unlock()
}

func (srv *Server) addConn(conn net.Conn) {
    srv.Lock()
    srv.conns[conn] = struct{}{}
    srv.Unlock()
}

func (srv *Server) delConn(conn net.Conn) {
    srv.Lock()
    delete(srv.conns, conn)
    srv.Unlock()
}

func (srv *Server) closeConns(force bool) bool {
    now := time.Now().Unix()
    lived := false
    srv.Lock()
    for c := range srv.conns {
        if !force {
            if sg, ok := c.(stateconn.StateGetter); ok {
                lasted, st := sg.GetState()
                if st == stateconn.StateNew && lasted.Unix() < now-int64(defaultConnIdleThreshold.Seconds()) {
                    st = stateconn.StateIdle
                }

                if st != stateconn.StateIdle && lasted.Unix() < now-int64(defaultConnIdleThreshold.Seconds()) {
                    lived = true
                    continue
                }

                // close raw connection to avoid data race
                // nolint
                sg.GetConn().Close()
            }
        }

        // force close the connection
        // nolint
        c.Close()
        delete(srv.conns, c)
    }
    srv.Unlock()
    return lived
}

func (srv *Server) closeListeners() error {
    var err error

    srv.Lock()
    for ln := range srv.listeners { // Close accept: no new connection
        if cerr := ln.Close(); cerr != nil && err == nil {
            err = cerr
        } else {
            delete(srv.listeners, ln)
        }
    }
    srv.Unlock()

    return err
}

func (srv *Server) Shutdown(ctx context.Context) error {
    err := srv.closeListeners()
    if err != nil {
        return err
    }

    ticker := time.NewTicker(defaultShutdownPollInterval)
    defer ticker.Stop()
    for {
        if srv.closeConns(false) {
            return err
        }
        select {
        case <-ctx.Done():
            if srv.closeConns(true) {
                return err
            }
            return ctx.Err()
        case <-ticker.C:
        }
    }
}
