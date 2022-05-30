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
    "bufio"
    "fmt"
    "log"

    "github.com/sofastack/sofa-bolt-go/sofabolt"
    "github.com/sofastack/sofa-common-go/helper/easyreader"
    "github.com/spf13/cobra"
)

var (
	format     string
	buffersize int
)

func init() {
	fs := decodeCmd.Flags()
	fs.StringVarP(&format, "format", "f", "hex", "Set the input format")
	fs.IntVarP(&buffersize, "buffersize", "s", 1024*8, "Set the decode buffer size")
}

var decodeCmd = &cobra.Command{
	Use:   "decode <input>",
	Short: "decode input to pretty stdout",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		o := easyreader.NewOption()
		if format == "hex" {
			o.SetDefaultFormat(easyreader.HexFormat)
		} else {
			o.SetDefaultFormat(easyreader.BinFormat)
		}
		reader, err := easyreader.EasyRead(o, args[0])
		if err != nil {
			log.Fatal(err)
		}

		var dcmd sofabolt.Command
		_, err = sofabolt.ReadCommand(sofabolt.NewReadOption(),
			bufio.NewReaderSize(reader, buffersize), &dcmd)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(dcmd.String())
	},
}
