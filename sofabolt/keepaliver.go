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
	"fmt"
    "net/http"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	sofalogger "github.com/sofastack/sofa-common-go/logger"
)

type KeepAliverOptions struct {
	Context           context.Context `json:"-"`
	MaClientConnUsed  int             `json:"max_client_used"`
	MinClientInPool   int             `json:"min_clinet_in_pool"`
	HeartbeatInterval time.Duration   `json:"heartbeat_interval"`
	HeartbeatTimeout  time.Duration   `json:"heartbeat_timeout"`
	CleanupInterval   time.Duration   `json:"cleanup_interval"`
	CleanupMaxChecks  int             `json:"cleanup_max_checks"`
}

type KeepAliver struct {
	logger  sofalogger.Logger
	options *KeepAliverOptions
	raw     PoolMap
	tls     PoolMap
	// dying holds the clients will dies
	dying sync.Map
}

func NewKeepAliver(o *KeepAliverOptions, logger sofalogger.Logger) (*KeepAliver, error) {
	ka := &KeepAliver{
		logger:  logger,
		options: o,
	}

	if err := ka.polyfill(); err != nil {
		return nil, err
	}

	go ka.doCleanup(o.Context)
	go ka.doHeartbeat(o.Context)

	return ka, nil
}

// nolint
func (ca *KeepAliver) polyfill() error {
	if ca.options.CleanupInterval == 0 {
		ca.options.CleanupInterval = 10 * time.Second
	}

	if ca.options.CleanupMaxChecks == 0 {
		ca.options.CleanupMaxChecks = 15
	}

	if ca.options.HeartbeatInterval == 0 {
		ca.options.HeartbeatInterval = 30 * time.Second
	}

	if ca.options.HeartbeatTimeout == 0 {
		ca.options.HeartbeatTimeout = 5 * time.Second
	}

	if ca.options.Context == nil {
		ca.options.Context = context.TODO()
	}

	return nil
}

func (ca *KeepAliver) doCleanup(ctx context.Context) {
	cleanupInterval := ca.options.CleanupInterval

	for {
		time.Sleep(cleanupInterval)
		select {
		case <-ctx.Done():
			ca.logger.Infof("shutdown bolt keepaliver cleanup")
			return
		default:
		}

		ca.dying.Range(func(key, value interface{}) bool {
			client, ok := key.(*Client)
			if !ok {
				panic("failed to type casting")
			}

			n, ok := value.(int)
			if !ok {
				panic("failed to type casting")
			}

			if n >= ca.options.CleanupMaxChecks {
				err := client.Close()
				ca.logger.Infof("close dying client (>= max checks) conn=%+v err=%=v",
					client.GetConn(), err)

			} else {
				if ref := client.GetMetrics().GetReferences(); ref > 0 {
					ca.logger.Infof("Skip close client ref=%d conn=%s", ref, client.GetConn())
					ca.dying.Store(client, n+1)

				} else {
					err := client.Close()
					ca.logger.Infof("Skip close client ref=%d conn=%s error=%+v", ref, client.GetConn(), err)
				}
			}

			return true
		})
	}
}

func (ca *KeepAliver) doHeartbeat(ctx context.Context) {
	heartbeatInterval := ca.options.HeartbeatInterval
	heartbeatTimeout := ca.options.HeartbeatTimeout

	req := AcquireRequest()
	res := AcquireResponse()
	req.SetProto(ProtoBOLTV1)
	req.SetCMDCode(CMDCodeBOLTHeartbeat)
	defer func() {
		ReleaseRequest(req)
		ReleaseResponse(res)
	}()

	clientChecker := func(scheme string, address string, p *Pool, t time.Time) bool {
		// send heartbeat and recive broken client
		clients := p.copyClients()
		// broken client list
		brokens := make([]*Client, 0)
		for _, client := range clients {
			if t.Unix()-client.GetMetrics().GetLasted() >= int64(heartbeatInterval.Seconds()) {
				// send heartbeat
				if err := client.DoTimeout(req, res, heartbeatTimeout); err != nil {
					// send heartbeat error, or server return error
					// this means client was broken
					brokens = append(brokens, client)
					ca.logger.Errorf("bolt heartbeat failed scheme=%s address=%s error=%+v",
						scheme, address, err)
				}
			}
		}

		// delete broken clients
		p.DeleteClients(brokens)

		// close broken clients
		for _, client := range brokens {
			_ = client.Close()
		}
		return true
	}

	timer := time.NewTicker(heartbeatInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			ca.logger.Infof("shutdown bolt clientaliver heartbeat")
			return
		case t := <-timer.C:
			ca.logger.Infof("try send bolt heartbeat")
			ca.raw.Range(func(addr string, pool *Pool) bool {
				return clientChecker("raw", addr, pool, t)
			})

			ca.raw.Range(func(addr string, pool *Pool) bool {
				return clientChecker("tls", addr, pool, t)
			})
		}
	}
}

func (ca *KeepAliver) Put(tls bool, force bool, addr string, client *Client) bool {
	if tls {
		return ca.put(&ca.tls, force, addr, client)
	}
	return ca.put(&ca.raw, force, addr, client)
}

func (ca *KeepAliver) put(m *PoolMap, force bool, addr string, client *Client) bool {
	var (
		loaded bool
		actual *Pool
		p      = NewPool()
	)

	p.Push(client)

	actual, loaded = m.LoadOrStore(addr, p)
	if loaded { // One guy win so release the loser
	}

	if actual.Size() >= 2 && ca.options.MaClientConnUsed > 0 &&
		client.GetMetrics().GetUsed() >= int64(ca.options.MaClientConnUsed) {
		return false
	}

	if force || (ca.options.MinClientInPool > 0 && actual.Size() <= ca.options.MinClientInPool) {
		if loaded {
			actual.Push(client)
		}
		return true
	}

	return false
}

func (t *KeepAliver) Get(tls bool, addr string) (*Client, bool) {
	if tls {
		return t.get(&t.tls, addr)
	}
	return t.get(&t.raw, addr)
}

func (t *KeepAliver) get(m *PoolMap, addr string) (*Client, bool) {
	p, ok := m.Load(addr)
	if !ok {
		return nil, false
	}

	var c *Client

	for {
		c, ok = p.Get()
		if !ok {
			return nil, false
		}

		if p.Size() >= 2 &&
			t.options.MaClientConnUsed > 0 &&
			c.GetMetrics().GetUsed() >= int64(t.options.MaClientConnUsed) {
			t.del(m, addr, c)
			t.GracefullyClose(c)
			continue
		}
		break
	}

	return c, true
}

func (ca *KeepAliver) Del(tls bool, address string, client *Client) bool {
	if tls {
		return ca.del(&ca.tls, address, client)
	}
	return ca.del(&ca.raw, address, client)
}

func (k *KeepAliver) del(m *PoolMap, addr string, client *Client) bool {
	p, ok := m.Load(addr)
	if !ok {
		return false
	}

	p.Delete(client)

	return true
}

func (k *KeepAliver) GracefullyClose(client *Client) {
	k.logger.Infof("try to gracefully close client used=%d lasted=%d ref=%d conn=%+v",
		client.GetMetrics().GetUsed(),
		client.GetMetrics().GetLasted(),
		client.GetMetrics().GetReferences(),
		client.GetConn())

	if ref := client.GetMetrics().GetReferences(); ref > 0 {
		k.dying.Store(client, 1)
	} else {
		err := client.Close()
		if err == nil {
			k.logger.Infof("direct close refless client")
		} else {
			k.logger.Infof("direct close refless client: %s", err.Error())
		}
	}
}

func (k *KeepAliver) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	type clientStatus struct {
		Connection string `json:"connection"`
		Used       int64  `json:"used"`
		References int64  `json:"references"`
		Lasted     int64  `json:"lasted"`
		Pending    int64  `json:"pending"`
		Check      int    `json:"check"`
	}

	type status struct {
		Elapsed string                   `json:"elapsed"`
		Options *KeepAliverOptions       `json:"options"`
		Raw     map[string]*Pool         `json:"raw"`
		TLS     map[string]*Pool         `json:"tls"`
		Dying   map[string]*clientStatus `json:"dying"`
	}

	started := time.Now()

	s := status{
		Options: k.options,
		Raw:     make(map[string]*Pool, 4096),
		TLS:     make(map[string]*Pool, 4096),
		Dying:   make(map[string]*clientStatus, 4096),
	}

	k.dying.Range(func(key, value interface{}) bool {
		client, ok := key.(*Client)
		if !ok {
			panic("failed to type casting")
		}

		n, ok := value.(int)
		if !ok {
			panic("failed to type casting")
		}

		conn := client.GetConn()
		s.Dying[conn.LocalAddr().String()] = &clientStatus{
			Connection: fmt.Sprintf("%s -> %s",
				conn.LocalAddr().String(),
				conn.RemoteAddr().String()),
			Used:       client.GetMetrics().GetUsed(),
			References: client.GetMetrics().GetReferences(),
			Lasted:     client.GetMetrics().GetLasted(),
			Pending:    client.GetMetrics().GetPendingCommands(),
			Check:      n,
		}

		return true
	})

	k.raw.Range(func(address string, pool *Pool) bool {
		s.Raw[address] = pool
		return true
	})

	k.tls.Range(func(address string, pool *Pool) bool {
		s.TLS[address] = pool
		return true
	})

	s.Elapsed = time.Since(started).String()

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = jsoniter.NewEncoder(rw).Encode(s)
}
