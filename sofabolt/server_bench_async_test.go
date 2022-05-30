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

import "testing"

func BenchmarkServerAsync1(b *testing.B) {
	benchmarkServer(b, 1, b.N, true)
}

// func BenchmarkServerAsync128(b *testing.B) {
// 	benchmarkServer(b, 128, b.N/128+1, true)
// }

// func BenchmarkServerAsync512(b *testing.B) {
// 	benchmarkServer(b, 512, b.N/512+1, true)
// }

func BenchmarkServerAsync1024(b *testing.B) {
	benchmarkServer(b, 1024, b.N/1024+1, true)
}
