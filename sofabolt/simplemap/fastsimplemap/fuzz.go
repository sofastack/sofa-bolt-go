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

package fastsimplemap

import (
	"encoding/hex"
	"fmt"
)

func Fuzz(data []byte) int {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(hex.EncodeToString(data))
			panic(r)
		}
	}()

	var (
		om FastSimpleMap
		nm FastSimpleMap
	)

	err := om.Decode(data)
	if err != nil {
		return 0
	}

	n := om.GetEncodeSize()
	e := make([]byte, n)
	om.Encode(e)

	if err := nm.Decode(e); err != nil {
		panic("failed to decode")
	}

	if !nm.Equal(&om) {
		panic("failed to equal")
	}

	return 1
}
