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
	"encoding/json"
	"sync"
)

//go:generate syncmap -pkg sofabolt -o pool_generated.go -name PoolMap map[string]*Pool

type Pool struct {
	sync.RWMutex
	clients []*Client
	next    int
}

func NewPool() *Pool {
	return &Pool{
		clients: make([]*Client, 0, 8),
	}
}

func (p *Pool) Size() int {
	var n int
	p.RLock()
	n = len(p.clients)
	p.RUnlock()
	return n
}

func (p *Pool) Iterate(fn func(client *Client)) {
	p.Lock()
	for i := range p.clients {
		fn(p.clients[i])
	}
	p.Unlock()
}

func (p *Pool) copyClients() []*Client {
	p.RLock()
	defer p.RUnlock()

	clientsCopy := make([]*Client, len(p.clients))
	copy(clientsCopy, p.clients)
	return clientsCopy
}

func (p *Pool) DeleteLocked(client *Client) {
	var i int
	for i = 0; i < len(p.clients); i++ {
		if client == p.clients[i] {
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
			break
		}
	}
}

func (p *Pool) Delete(client *Client) {
	var i int
	p.Lock()
	for i = 0; i < len(p.clients); i++ {
		if client == p.clients[i] {
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
			break
		}
	}
	p.Unlock()
}

func (p *Pool) DeleteClients(clients []*Client) {
	p.Lock()
	defer p.Unlock()

	for _, client := range clients {
		p.DeleteLocked(client)
	}
}

func (p *Pool) Push(client *Client) {
	p.Lock()
	p.clients = append(p.clients, client)
	p.Unlock()
}

func (p *Pool) Get() (*Client, bool) {
	var (
		client *Client
		n      int
	)

	p.Lock()
	n = len(p.clients)
	if n == 0 {
		p.Unlock()
		return nil, false
	}
	p.next = (p.next + 1) % n
	client = p.clients[p.next]
	p.Unlock()

	return client, true
}

func (p *Pool) MarshalJSON() ([]byte, error) {
	type clientStatus struct {
		Closed          bool  `json:"closed"`
		Ref             int64 `json:"ref"`
		Lasted          int64 `json:"lasted"`
		Used            int64 `json:"used"`
		Created         int64 `json:"created"`
		PendingRequests int64 `json:"pending_requests"`
	}

	type status struct {
		Next    int            `json:"next"`
		Clients []clientStatus `json:"clients"`
	}

	s := status{
		Clients: make([]clientStatus, 0, 8),
	}

	p.RLock()
	s.Next = p.next
	for i := 0; i < len(p.clients); i++ {
		s.Clients = append(s.Clients, clientStatus{
			Lasted:          p.clients[i].GetMetrics().GetLasted(),
			Closed:          p.clients[i].Closed(),
			Ref:             p.clients[i].GetMetrics().GetReferences(),
			Used:            p.clients[i].GetMetrics().GetUsed(),
			Created:         p.clients[i].GetMetrics().GetCreated(),
			PendingRequests: p.clients[i].GetMetrics().GetPendingCommands(),
		})
	}
	p.RUnlock()

	return json.Marshal(s)
}
