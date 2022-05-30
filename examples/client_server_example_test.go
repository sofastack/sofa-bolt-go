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

package examples

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/sofastack/sofa-bolt-go/sofabolt"
)

func ExampleClientAndServer() {
	srv, err := sofabolt.NewServer(
		sofabolt.WithServerTimeout(
			5*time.Second,
			5*time.Second,
			5*time.Second,
			0*time.Second,
		),
		sofabolt.WithServerHandler(sofabolt.HandlerFunc(func(rw sofabolt.ResponseWriter, req *sofabolt.Request) {
			fmt.Println(string(req.GetContent()))
			rw.GetResponse().SetContent(req.GetContent())
			rw.Write()
		})),
	)
	if err != nil {
		log.Fatal(err)
	}

	p0, p1 := net.Pipe()

	go func() {
		srv.ServeConn(p0)
	}()

	c, err := sofabolt.NewClient(
		sofabolt.WithClientTimeout(
			5*time.Second,
			5*time.Second,
			5*time.Second,
			0*time.Second,
		),
		sofabolt.WithClientConn(p1),
	)
	if err != nil {
		log.Fatal(err)
	}

	var (
		req = sofabolt.AcquireRequest()
		res = sofabolt.AcquireResponse()
	)
	defer func() {
		sofabolt.ReleaseRequest(req)
		sofabolt.ReleaseResponse(res)
	}()

	req.SetContentString("hello world")

	err = c.DoTimeout(req, res, 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	// Output: hello world
}

func ExampleClientAndServerDialer() {
	ln, err := net.Listen("tcp4", ":0")
	if err != nil {
		log.Fatal(err)
	}

	go func() { // start a temporary server
		srv, err := sofabolt.NewServer(
			sofabolt.WithServerTimeout(
				5*time.Second,
				5*time.Second,
				5*time.Second,
				0*time.Second,
			),
			sofabolt.WithServerHandler(sofabolt.HandlerFunc(func(rw sofabolt.ResponseWriter, req *sofabolt.Request) {
				fmt.Println(string(req.GetContent()))
				rw.GetResponse().SetContent(req.GetContent())
				rw.Write()
			})),
		)
		if err != nil {
			log.Fatal(err)
		}

		srv.Serve(ln)
	}()

	c, err := sofabolt.NewClient(
		sofabolt.WithClientRedial(sofabolt.DialerFunc(func() (net.Conn, error) {
			return net.DialTimeout("tcp4", ln.Addr().String(), 5*time.Second)
		})), // always redial
		sofabolt.WithClientTimeout(
			0,             // no read timeout
			5*time.Second, // write timeout
			0,             // no idle timeout
			0,             // no flush timeout
		),
		sofabolt.WithClientHandler(sofabolt.HandlerFunc(func(rw sofabolt.ResponseWriter, req *sofabolt.Request) {
			rw.Write()
		})),
		sofabolt.WithClientMaxPendingCommands( // set max pending commands
			512),
		sofabolt.WithClientHeartbeat(
			30*time.Second,
			10*time.Second,
			0,
			func(success bool) {
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	var (
		req = sofabolt.AcquireRequest()
		res = sofabolt.AcquireResponse()
	)
	defer func() {
		sofabolt.ReleaseRequest(req)
		sofabolt.ReleaseResponse(res)
	}()
	req.SetContentString("hello world")

	err = c.DoTimeout(req, res, 1*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	// Output: hello world
}
