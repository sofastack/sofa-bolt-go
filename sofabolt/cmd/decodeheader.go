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
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/sofastack/sofa-bolt-go/sofabolt"
	"github.com/sofastack/sofa-common-go/helper/easyreader"
)

var decodeheaderCmd = &cobra.Command{
	Use:   "decodeheader <input>",
	Short: "decodeheader to pretty stdout",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		reader, err := easyreader.EasyRead(easyreader.NewOption(), args[0])
		if err != nil {
			log.Fatal(err)
		}

		data, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		}

		var h sofabolt.SimpleMap
		if err := h.Decode(data); err != nil {
			log.Fatal(err)
		}

		fmt.Println(h.String())
	},
}
