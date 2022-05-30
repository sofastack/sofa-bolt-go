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

import "io"

type ClientConnProtocolEncoderOption struct {
}

func NewClientConnProtocolEncoderOption() *ClientConnProtocolEncoderOption {
	return &ClientConnProtocolEncoderOption{}
}

type ClientConnProtocolEncoder interface {
	Encode(eo *ClientConnProtocolEncoderOption, dst []byte, cmd interface{}) ([]byte, error)
}

type ClientConnProtocolEncoderFunc func(*ClientConnProtocolEncoderOption, []byte, interface{}) ([]byte, error)

func (ccp ClientConnProtocolEncoderFunc) Encode(eo *ClientConnProtocolEncoderOption,
	dst []byte, cmd interface{}) ([]byte, error) {
	return ccp(eo, dst, cmd)
}

type ClientConnProtocolDecoderOption struct {
}

func NewClientConnProtocolDecoderOption() *ClientConnProtocolDecoderOption {
	return &ClientConnProtocolDecoderOption{}
}

type ClientConnProtocolDecoder interface {
	Decode(do *ClientConnProtocolDecoderOption, r io.Reader) (cmd interface{}, err error)
}

type ClientConnProtocolDecoderFunc func(do *ClientConnProtocolDecoderOption, r io.Reader) (cmd interface{}, err error)

func (ccp ClientConnProtocolDecoderFunc) Decode(do *ClientConnProtocolDecoderOption,
	r io.Reader) (cmd interface{}, err error) {
	return ccp(do, r)
}

type ClientConnProtocolIDIncrementer interface {
	IncrementID() uint64
}
