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

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/sofastack/sofa-bolt-go/sofabolt"
	"github.com/sofastack/sofa-common-go/helper/easyreader"
)

var requestBody string

func init() {
	fs := requestCmd.Flags()
	fs.StringVarP(&requestBody, "request-body", "d", "", "Set the request body")
}

var requestCmd = &cobra.Command{
	Use:   "request <address>",
	Short: "request to bolt server",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			reader io.Reader
			err    error
		)

		if len(requestBody) > 0 {
			reader, err = easyreader.EasyRead(easyreader.NewOption(), requestBody)
			if err != nil {
				log.Fatal(err)
			}
		}

		conn, err := net.Dial("tcp", args[0])
		if err != nil {
			log.Fatal(err)
		}

		c, err := sofabolt.NewClient(sofabolt.WithClientConn(conn))
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

		if reader != nil {
			var data []byte
			data, err = ioutil.ReadAll(reader)
			if err != nil {
				log.Fatal(err)
			}
			req.SetContent(data)
		}

		err = c.DoTimeout(req, res, 5*time.Second)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(res.String())
	},
}
