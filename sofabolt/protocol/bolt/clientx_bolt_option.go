package bolt

import (
	"net"
	"time"

	"github.com/sofastack/sofa-bolt-go/sofabolt"
)

// ClientConnBOLTOptionSetter configures a ClientConnBOLT.
type ClientConnBOLTOptionSetter interface {
	Set(*ClientConnBOLT)
}

type ClientConnBOLTOptionSetterFunc func(*ClientConnBOLT)

func (f ClientConnBOLTOptionSetterFunc) Set(c *ClientConnBOLT) {
	f(c)
}

func WithClientConnBOLTHandler(handler sofabolt.Handler) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.options.handler = handler
	})
}

func WithClientConnBOLTMetrics(cm *sofabolt.ClientMetrics) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.options.metrics = cm
	})
}

func WithClientConnBOLTTimeout(
	readtimeout,
	writetimeout,
	idletimeout,
	flushInterval time.Duration) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.options.readTimeout = readtimeout
		c.options.writeTimeout = writetimeout
		c.options.idleTimeout = idletimeout
	})
}

func WithClientConnBOLTConn(conn net.Conn) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.conn = conn
	})
}

func WithClientConnBOLTHeartbeat(
	heartbeatinterval,
	heartbeattimeout time.Duration,
	heartbeatprobes int, onheartbeat func(success bool)) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.options.heartbeatTimeout = heartbeattimeout
		c.options.heartbeatInterval = heartbeatinterval
		c.options.heartbeatProbes = heartbeatprobes
		c.options.onHeartbeat = onheartbeat
	})
}

func WithClientConnBOLTMaxPendingCommands(m int) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.options.maxPendingCommands = m
	})
}

func WithClientConnBOLTRedial(dialer sofabolt.Dialer) ClientConnBOLTOptionSetterFunc {
	return ClientConnBOLTOptionSetterFunc(func(c *ClientConnBOLT) {
		c.options.dialer = dialer
	})
}
