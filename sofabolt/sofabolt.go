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
	"errors"
)

var (
	ErrBufferNotEnough       = errors.New("sofabolt: buffer not enough")
	ErrMalformedProto        = errors.New("sofabolt: malformed proto")
	ErrMalformedType         = errors.New("sofabolt: malformed type")
	ErrServerHandler         = errors.New("sofabolt: server handler cannot be nil")
	ErrServerNotARequest     = errors.New("sofabolt: server received a response")
	ErrClientExpectResponse  = errors.New("sofabolt: receive a request")
	ErrClientTimeout         = errors.New("sofabolt: client do timeout")
	ErrClientNotARequest     = errors.New("sofabolt: client send a response")
	ErrClientWasClosed       = errors.New("sofabolt: client was closed")
	ErrClientTooManyRequests = errors.New("sofabolt: client too many requests")
	ErrClientServerTimeout   = errors.New("sofabolt: clientserver do timeout")
	ErrClientDisableRedial   = errors.New("sofabolt: disable redial")
	ErrClientNilConnection   = errors.New("sofabolt: client connection is nil")
)
