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
	"testing"
)

func BenchmarkReadTBRemotingRequest(b *testing.B) {
	// nolint
	x := []byte{0x0d, 0x00, 0x04, 0x02, 0x00, 0x00, 0x00, 0x00, 0x81, 0x2c, 0x00, 0x00, 0x01, 0xdb, 0x4f, 0xba, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x61, 0x6f, 0x62, 0x61, 0x6f, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x69, 0x6d, 0x70, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x91, 0x03, 0x63, 0x74, 0x78, 0x6f, 0x90, 0x4f, 0xc8, 0x39, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x61, 0x6f, 0x62, 0x61, 0x6f, 0x2e, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x69, 0x6d, 0x70, 0x6c, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x24, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x92, 0x02, 0x69, 0x64, 0x06, 0x74, 0x68, 0x69, 0x73, 0x24, 0x30, 0x6f, 0x91, 0x3c, 0x09, 0xf4, 0x4a, 0x00, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x6c, 0x69, 0x70, 0x61, 0x79, 0x2e, 0x73, 0x6f, 0x66, 0x61, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x6f, 0x66, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4f, 0xbc, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x6c, 0x69, 0x70, 0x61, 0x79, 0x2e, 0x73, 0x6f, 0x66, 0x61, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x6f, 0x66, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x95, 0x0d, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x41, 0x70, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x0a, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x17, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x71, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x70, 0x73, 0x0d, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x41, 0x72, 0x67, 0x53, 0x69, 0x67, 0x73, 0x6f, 0x90, 0x04, 0x62, 0x61, 0x72, 0x31, 0x07, 0x65, 0x63, 0x68, 0x6f, 0x53, 0x74, 0x72, 0x53, 0x00, 0x36, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x6c, 0x69, 0x70, 0x61, 0x79, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x66, 0x61, 0x63, 0x61, 0x64, 0x65, 0x2e, 0x53, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x31, 0x2e, 0x30, 0x4d, 0x11, 0x72, 0x70, 0x63, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x4d, 0x09, 0x73, 0x6f, 0x66, 0x61, 0x52, 0x70, 0x63, 0x49, 0x64, 0x01, 0x30, 0x07, 0x45, 0x6c, 0x61, 0x73, 0x74, 0x69, 0x63, 0x01, 0x46, 0x0b, 0x73, 0x79, 0x73, 0x50, 0x65, 0x6e, 0x41, 0x74, 0x74, 0x72, 0x73, 0x00, 0x0d, 0x73, 0x6f, 0x66, 0x61, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x49, 0x64, 0x63, 0x03, 0x64, 0x65, 0x76, 0x09, 0x7a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x55, 0x49, 0x44, 0x00, 0x10, 0x7a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5a, 0x6f, 0x6e, 0x65, 0x00, 0x0c, 0x73, 0x6f, 0x66, 0x61, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x49, 0x70, 0x0d, 0x31, 0x31, 0x2e, 0x31, 0x36, 0x36, 0x2e, 0x32, 0x32, 0x2e, 0x31, 0x36, 0x31, 0x0b, 0x73, 0x6f, 0x66, 0x61, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64, 0x1e, 0x30, 0x62, 0x61, 0x36, 0x31, 0x36, 0x61, 0x31, 0x31, 0x35, 0x33, 0x35, 0x39, 0x38, 0x35, 0x34, 0x39, 0x31, 0x38, 0x37, 0x38, 0x31, 0x34, 0x38, 0x32, 0x32, 0x39, 0x31, 0x30, 0x36, 0x0c, 0x73, 0x6f, 0x66, 0x61, 0x50, 0x65, 0x6e, 0x41, 0x74, 0x74, 0x72, 0x73, 0x00, 0x0e, 0x73, 0x6f, 0x66, 0x61, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x5a, 0x6f, 0x6e, 0x65, 0x05, 0x47, 0x5a, 0x30, 0x30, 0x42, 0x0d, 0x73, 0x6f, 0x66, 0x61, 0x43, 0x61, 0x6c, 0x6c, 0x65, 0x72, 0x41, 0x70, 0x70, 0x03, 0x66, 0x6f, 0x6f, 0x0d, 0x7a, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x04, 0x33, 0x30, 0x30, 0x30, 0x7a, 0x7a, 0x56, 0x74, 0x00, 0x07, 0x5b, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x6e, 0x01, 0x10, 0x6a, 0x61, 0x76, 0x61, 0x2e, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x7a, 0x02, 0x74, 0x72}

	b.Run("request", func(b *testing.B) {
		var req Request
		br := bytes.NewReader(x)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			br.Reset(x)
			_, err := req.Read(NewReadOption(), br)
			if err != nil {
				b.Fatal(err)
			}
			req.Reset()
		}
	})

	b.Run("command", func(b *testing.B) {
		var cmd Command
		br := bytes.NewReader(x)
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			br.Reset(x)
			_, err := cmd.Read(NewReadOption(), br)
			if err != nil {
				b.Fatal(err)
			}
			cmd.Reset()
		}
	})
}

func BenchmarkWriteRequest(b *testing.B) {
	var r Request
	r.SetClassString("ccc")
	r.SetRequestID(123)
	r.SetType(TypeBOLTRequest)
	d := make([]byte, 64)
	var err error

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d, err = r.Write(&WriteOption{}, d[:0])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteResponse(b *testing.B) {
	var r Response
	r.SetClassString("ccc")
	r.GetHeaders().Set("Content-Type", "text/plain")
	r.SetType(TypeBOLTResponse)
	d := make([]byte, 64)
	var err error

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		d, err = r.Write(&WriteOption{}, d[:0])
		if err != nil {
			b.Fatal(err)
		}
	}
}
