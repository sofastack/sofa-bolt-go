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
	"fmt"
	"log"
)

func ExampleCommand() {
	req := AcquireRequest()
	d, err := req.Write(NewWriteOption(), nil)
	if err != nil {
		log.Fatal(err)
	}

	newreq := AcquireRequest()
	_, err = newreq.Read(NewReadOption(), bytes.NewReader(d))
	if err != nil {
		log.Fatal(err)
	}

	res := AcquireResponse()
	d, err = res.Write(NewWriteOption(), nil)
	if err != nil {
		log.Fatal(err)
	}

	newres := AcquireResponse()
	_, err = newres.Read(NewReadOption(), bytes.NewReader(d))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(req.String() == newreq.String())
	fmt.Println(res.String() == newres.String())
	// Output: true
	// true
}
