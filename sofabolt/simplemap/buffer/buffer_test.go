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

package buffer

import (
	"testing"
)

func TestBufferEncode(t *testing.T) {
	en := New(make([]byte, 1+2+4+8+5))
	en.MustPutUint8(1)
	en.MustPutUint16(2)
	en.MustPutUint32(3)
	en.MustPutUint64(4)
	en.MustPutBytes([]byte("abcde"))

	var (
		a uint8
		b uint16
		c uint32
		d uint64
		e []byte
	)

	de := New(en.Bytes())
	de.MustUint8(&a).MustUint16(&b).MustUint32(&c).MustUint64(&d).MustRef(5, &e)
	if a != 1 || b != 2 || c != 3 || d != 4 || string(e) != "abcde" {
		t.Fatal("encode deocde failed")
	}

	if de.Remain() != 0 {
		t.Fatal("remain failed")
	}
}
