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
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type MyCommand struct {
	A string
	B uint32
}

func TestClientConnReadTimeout(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	require.Nil(t, err)

	go func() {
		// nolint
		conn, err := ln.Accept()
		_ = err
		var b [1024]byte
		conn.Read(b[:]) // do not write data to trigger read timeout
	}()

	timeout := uint32(0)

	conn, err := NewClientConn(
		WithClientConnTimeout(
			100*time.Millisecond,
			0*time.Second,
			0*time.Second,
			0*time.Second,
		),

		WithClientConnStatusChanger(ClientConnStatusChangerFunc(func(cc *ClientConn,
			from, to ClientConnStatus) {
			if to == ReadTimeoutClientConnStatus {
				atomic.AddUint32(&timeout, 1)
			}
		})),

		WithClientConnRedial(DialerFunc(func() (net.Conn, error) {
			return net.Dial("tcp", ln.Addr().String())
		})),

		WithClientConnDispatcher(ClientConnDispatcherFunc(func(err error, cmd interface{}) {
		})),

		WithClientConnProtocolDecoder(ClientConnProtocolDecoderFunc(func(do *ClientConnProtocolDecoderOption,
			r io.Reader) (cmd interface{}, err error) {
			var b [1024]byte
			n, err := io.ReadFull(r, b[:]) // should be read timeout
			_ = n
			return nil, err
		})),
		WithClientConnProtocolEncoder(ClientConnProtocolEncoderFunc(func(*ClientConnProtocolEncoderOption,
			[]byte, interface{}) ([]byte, error) {
			return nil, nil
		})),
	)
	require.Nil(t, err)

	_ = conn
	time.Sleep(1 * time.Second)
	x := atomic.LoadUint32(&timeout)
	t.Logf("timeouts: %d", x)
	require.True(t, x > 0)
}

func TestClientConn(t *testing.T) {
	c0, c1 := net.Pipe()
	count := 2048

	go func() {
		var cmd MyCommand
		for i := 0; i < count; i++ {
			cmd.B = uint32(i)
			if err := gob.NewEncoder(c1).Encode(&cmd); err != nil {
				return
			}
		}
	}()

	dialer := DialerFunc(func() (net.Conn, error) {
		return c0, nil
	})

	statusLock := sync.RWMutex{}
	status := make([][2]ClientConnStatus, 0, 16)
	changer := ClientConnStatusChangerFunc(func(cc *ClientConn, from, to ClientConnStatus) {
		statusLock.Lock()
		defer statusLock.Unlock()
		status = append(status, [2]ClientConnStatus{from, to})
		t.Log(from, "=>", to)
	})

	dispatch := uint32(0)
	doneCh := make(chan struct{})
	dispatcher := ClientConnDispatcherFunc(func(err error, cmd interface{}) {
		c := atomic.AddUint32(&dispatch, 1)
		switch cmd.(type) {
		case *MyCommand:
		default:
			t.Fatalf("expect decode MyCommand  but not")
		}

		if c == uint32(count) {
			doneCh <- struct{}{}
		}
	})

	decoder := ClientConnProtocolDecoderFunc(func(do *ClientConnProtocolDecoderOption,
		r io.Reader) (interface{}, error) {
		var cmd MyCommand
		err := gob.NewDecoder(r).Decode(&cmd)
		return &cmd, err
	})

	encoder := ClientConnProtocolEncoderFunc(func(eo *ClientConnProtocolEncoderOption,
		dst []byte, cmd interface{}) ([]byte, error) {
		switch x := cmd.(type) {
		case *MyCommand:
			bw := bytes.NewBuffer(dst)
			err := gob.NewEncoder(bw).Encode(x)
			return bw.Bytes(), err
		default:
			return nil, fmt.Errorf("unknown command")
		}
	})

	conn, err := NewClientConn(
		WithClientConnRedial(dialer),
		WithClientConnStatusChanger(changer),
		WithClientConnDispatcher(dispatcher),
		WithClientConnProtocolDecoder(decoder),
		WithClientConnProtocolEncoder(encoder),
	)
	require.Nil(t, err)

	_ = conn
	<-doneCh

	conn.Close()

	// wait close
	time.Sleep(2 * time.Second)

	statusLock.Lock()
	defer statusLock.Unlock()
	for i, cs := range [][2]ClientConnStatus{
		{IdleClientConnStatus, ConnectingClientConnStatus},
		{ConnectingClientConnStatus, ActiveClientConnStatus},
		{ActiveClientConnStatus, TransientFailureClientConnStatus},
		{TransientFailureClientConnStatus, ShutdownClientConnStatus},
	} {
		require.Equal(t, cs[0], status[i][0])
		require.Equal(t, cs[1], status[i][1])
	}
}

func TestClinetConnStatus(t *testing.T) {
	ss := []string{
		"Idle",
		"Connecting",
		"ReadTimeout",
		"Active",
		"TransientFailure",
		"Shutdown",
	}
	for i, s := range []ClientConnStatus{
		IdleClientConnStatus,
		ConnectingClientConnStatus,
		ReadTimeoutClientConnStatus,
		ActiveClientConnStatus,
		TransientFailureClientConnStatus,
		ShutdownClientConnStatus,
	} {
		require.Equal(t, ss[i], s.String())
	}
}
