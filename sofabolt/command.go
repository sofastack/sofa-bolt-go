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
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"io"
	"strconv"

	"github.com/sofastack/sofa-common-go/helper/readhelper"
	"github.com/sofastack/sofa-hessian-go/javaobject"
	"github.com/sofastack/sofa-hessian-go/sofahessian"
)

var trregistry sofahessian.ClassRegistry

// nolint
func init() {
	trregistry.RegisterJavaClass(&javaobject.TBRemotingConnectionRequest{})
	trregistry.RegisterJavaClass(&javaobject.TBRemotingConnectionResponse{})
}

const (
	ClassRequest  = "com.alipay.sofa.rpc.core.request.SofaRequest"
	ClassResponse = "com.alipay.sofa.rpc.core.response.SofaResponse"
)

var (
	ClassRequestBytes  = []byte(ClassRequest)
	ClassResponseBytes = []byte(ClassResponse)
)

type Proto uint8

const (
	ProtoBOLTV1     Proto = 0x01
	ProtoBOLTV2     Proto = 0x02
	ProtoTBRemoting Proto = 0x0d
)

func (p Proto) String() string {
	switch p {
	case ProtoBOLTV1:
		return "boltv1"
	case ProtoBOLTV2:
		return "boltv2"
	case ProtoTBRemoting:
		return "tbremoting"
	default:
		return "unknown"
	}
}

type Version uint8

const (
	VersionBOLTV1 Version = 1
	VersionBOLTV2 Version = 2
)

func (c Version) String() string {
	switch c {
	case VersionBOLTV1:
		return "boltv1"
	case VersionBOLTV2:
		return "boltv2"
	default:
		return "unknown"
	}
}

type Type uint8

const (
	TypeBOLTResponse      Type = 0
	TypeBOLTRequest       Type = 1
	TypeBOLTRequestOneWay Type = 2
	TypeTBRemotingOneWay  Type = 1
	TypeTBRemotingTwoWay  Type = 2
)

func (t Type) String() string {
	switch t {
	case TypeBOLTResponse:
		return "response"
	case TypeBOLTRequest:
		return "request"
	case TypeBOLTRequestOneWay:
		return "oneway"
	default:
		return "unknown"
	}
}

type CMDCode uint16

const (
	CMDCodeBOLTHeartbeat CMDCode = 0
	CMDCodeBOLTRequest   CMDCode = 1
	CMDCodeBOLTResponse  CMDCode = 2

	CMDCodeTRemotingHeartbeat CMDCode = 0
	CMDCodeTRemotingRequest   CMDCode = 13
	CMDCodeTRemotingResponse  CMDCode = 14
)

func (c CMDCode) String() string {
	switch c {
	case CMDCodeBOLTHeartbeat:
		return "heartbeat"
	case CMDCodeBOLTRequest:
		return "bolt-req"
	case CMDCodeBOLTResponse:
		return "bolt-res"
	case CMDCodeTRemotingRequest:
		return "tr-req"
	case CMDCodeTRemotingResponse:
		return "tr-res"
	default:
		return "unknown cmdcode"
	}
}

type Codec uint8

const (
	CodecHessian            Codec = 0
	CodecHessian2           Codec = 1
	CodecProtobuf           Codec = 11
	CodecJSON               Codec = 12
	CodecTBRemotingHessian2 Codec = 4
	CodecTBRemotingHessian1 Codec = 1
)

func (c Codec) String() string {
	switch c {
	case CodecHessian:
		return "hessian"
	case CodecHessian2:
		return "hessian2"
	case CodecProtobuf:
		return "prtobuf"
	case CodecTBRemotingHessian2:
		return "tbhessian2"
	default:
		return "unknown"
	}
}

type Status uint16

const (
	StatusSuccess                Status = 0  // 0x00 response status
	StatusError                  Status = 1  // 0x01
	StatusServerException        Status = 2  // 0x02
	StatusUnknown                Status = 3  // 0x03
	StatusServerThreadPoolBusy   Status = 4  // 0x04
	StatusErrorComm              Status = 5  // 0x05
	StatusNoProcessor            Status = 6  // 0x06
	StatusTimeout                Status = 7  // 0x07
	StatusClientSendError        Status = 8  // 0x08
	StatusCodecException         Status = 9  // 0x09
	StatusConnectionClosed       Status = 16 // 0x10
	StatusServerSerialException  Status = 17 // 0x11
	StatusServerDeseralException Status = 18 // 0x12
)

type Command struct {
	proto    Proto
	ver1     Version
	typ      Type
	cmdcode  CMDCode
	ver2     uint8
	rid      uint32
	codec    Codec
	switc    uint8
	timeout  uint32
	status   Status
	reserved [1]byte

	connection []byte
	class      []byte
	header     []byte
	content    []byte
	crc32      uint32

	headers SimpleMap
}

func (c *Command) Reset() {
	c.proto = 0
	c.ver1 = 0
	c.typ = 0
	c.cmdcode = 0
	c.ver2 = 0
	c.rid = 0
	c.codec = 0
	c.switc = 0
	c.timeout = 0
	c.status = 0
	c.reserved[0] = 0

	c.connection = c.connection[:0]
	c.class = c.class[:0]
	c.header = c.header[:0]
	c.content = c.content[:0]
	c.crc32 = 0

	c.headers.Reset()
}

func (c *Command) IsRequest() bool {
	return c.typ == TypeBOLTRequest || c.typ == TypeBOLTRequestOneWay || c.typ == TypeTBRemotingOneWay
}

func (c *Command) SetProto(p Proto)       { c.proto = p }
func (c *Command) SetVer1(v Version)      { c.ver1 = v }
func (c *Command) SetType(t Type)         { c.typ = t }
func (c *Command) SetCMDCode(cmd CMDCode) { c.cmdcode = cmd }
func (c *Command) SetVer2(v uint8)        { c.ver2 = v }
func (c *Command) SetRequestID(id uint32) { c.rid = id }
func (c *Command) SetCodec(codec Codec)   { c.codec = codec }
func (c *Command) SetSwitc(s uint8)       { c.switc = s }
func (c *Command) SetTimeout(t uint32)    { c.timeout = t }
func (c *Command) SetStatus(s Status)     { c.status = s }
func (c *Command) SetConnection(connection []byte) {
	c.connection = append(c.connection[:0], connection...)
}

func (c *Command) SetClass(class []byte)           { c.class = append(c.class[:0], class...) }
func (c *Command) SetClassString(class string)     { c.class = append(c.class[:0], class...) }
func (c *Command) SetContent(content []byte)       { c.content = append(c.content[:0], content...) }
func (c *Command) SetContentString(content string) { c.content = append(c.content[:0], content...) }
func (c *Command) GetProto() Proto                 { return c.proto }
func (c *Command) GetVer1() Version                { return c.ver1 }
func (c *Command) GetType() Type                   { return c.typ }
func (c *Command) GetCMDCode() CMDCode             { return c.cmdcode }
func (c *Command) GetVer2() uint8                  { return c.ver2 }
func (c *Command) GetRequestID() uint32            { return c.rid }
func (c *Command) GetCodec() Codec                 { return c.codec }
func (c *Command) GetSwitc() uint8                 { return c.switc }
func (c *Command) GetTimeout() uint32              { return c.timeout }
func (c *Command) GetStatus() Status               { return c.status }
func (c *Command) GetClass() []byte                { return c.class }
func (c *Command) GetHeaders() *SimpleMap          { return &c.headers }
func (c *Command) GetContent() []byte              { return c.content }
func (c *Command) GetConnection() []byte           { return c.connection }

func (c *Command) ShallowCopyTo(d *Command) {
	d.proto = c.proto
	d.ver1 = c.ver1
	d.typ = c.typ
	d.cmdcode = c.cmdcode
	d.ver2 = c.ver2
	d.rid = c.rid
	d.codec = c.codec
	d.switc = c.switc
	d.timeout = c.timeout
	d.status = c.status

	d.class = c.class
	d.header = c.header
	d.content = c.content
	d.crc32 = c.crc32
	// nolint
	d.headers = d.headers
}

func (c *Command) CopyTo(d *Command) {
	d.proto = c.proto
	d.ver1 = c.ver1
	d.typ = c.typ
	d.cmdcode = c.cmdcode
	d.ver2 = c.ver2
	d.rid = c.rid
	d.codec = c.codec
	d.switc = c.switc
	d.timeout = c.timeout
	d.status = c.status

	d.class = append(d.class[:0], c.class...)
	d.header = append(d.header[:0], c.header...)
	d.content = append(d.content[:0], c.content...)
	d.crc32 = c.crc32

	c.headers.CopyTo(&d.headers)
}

// nolint
func (c *Command) String() string {
	sep := []byte(",")
	w := bytes.NewBuffer(make([]byte, 0, 64))

	w.WriteString("Proto:")
	w.WriteString(c.proto.String())
	w.Write(sep)

	w.WriteString("Ver1:")
	w.WriteString(c.ver1.String())
	w.Write(sep)

	w.WriteString("Type:")
	w.WriteString(c.typ.String())
	w.Write(sep)

	w.WriteString("Cmdcode:")
	w.WriteString(c.cmdcode.String())
	w.Write(sep)

	w.WriteString("Ver2:")
	w.WriteString(strconv.Itoa(int(c.ver2)))
	w.Write(sep)

	w.WriteString("Rid:")
	w.WriteString(strconv.Itoa(int(c.rid)))
	w.Write(sep)

	w.WriteString("Codec:")
	w.WriteString(c.codec.String())
	w.Write(sep)

	w.WriteString("Switch:")
	w.WriteString(strconv.Itoa(int(c.switc)))
	w.Write(sep)

	w.WriteString("Timeout:")
	w.WriteString(strconv.Itoa(int(c.timeout)))
	w.Write(sep)

	w.WriteString("Status:")
	w.WriteString(strconv.Itoa(int(c.status)))
	w.Write(sep)

	w.WriteString("ConnectionLen:")
	w.WriteString(strconv.Itoa(len(c.connection)))
	w.Write(sep)

	w.WriteString("ClassLen:")
	w.WriteString(strconv.Itoa(len(c.class)))
	w.Write(sep)

	w.WriteString("HeaderLen:")
	w.WriteString(strconv.Itoa(c.headers.GetEncodeSize()))
	w.Write(sep)

	w.WriteString("ContentLen:")
	w.WriteString(strconv.Itoa(len(c.content)))
	w.Write(sep)

	w.WriteString("Class:")
	w.Write(c.class)
	w.Write(sep)

	w.WriteString("Connection:")
	w.Write(c.connection)
	w.Write(sep)

	w.WriteString("Headers:")
	w.WriteString(c.headers.String())
	w.Write(sep)

	w.WriteString("Content:")
	w.WriteString(hex.EncodeToString(c.content))
	w.Write(sep)

	w.WriteString("CRC32:")
	w.WriteString(strconv.Itoa(int(c.crc32)))

	return w.String()
}

func (c *Command) Size() int {
	classLen := len(c.class)
	headerLen := c.headers.GetEncodeSize()
	contentLen := len(c.content)

	switch c.proto {
	case 0, ProtoBOLTV1:
		if c.typ == TypeBOLTRequest || c.typ == TypeBOLTRequestOneWay {
			return 20 + classLen + headerLen + contentLen
		}
		return 24 + classLen + headerLen + contentLen
	case ProtoBOLTV2:
		if c.typ == TypeBOLTRequest || c.typ == TypeBOLTRequestOneWay {
			return 24 + classLen + headerLen + contentLen
		}
		return 22 + classLen + headerLen + contentLen
	case ProtoTBRemoting:
		return 14 + classLen + headerLen + contentLen
	default:
		return 0
	}
}

func (c *Command) Write(wo *WriteOption, b []byte) ([]byte, error) {
	return WriteCommand(wo, b, c)
}

func (c *Command) Read(ro *ReadOption, br io.Reader) (int, error) {
	return ReadCommand(ro, br, c)
}

// WriteCommand writes the command to []byte.
func WriteCommand(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	switch cmd.proto {
	case 0, ProtoBOLTV1:
		return writeCommandBOLTV1(wo, b, cmd)
	case ProtoBOLTV2:
		return writeCommandBOLTV2(wo, b, cmd)
	case ProtoTBRemoting:
		return writeCommandTBRemoting(wo, b, cmd)
	default:
		return nil, ErrMalformedProto
	}
}

func writeCommandTBRemoting(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	/**
	 *   Header(1B): 报文版本
	 *   Header(1B): 请求/响应
	 *   Header(1B): 序列化协议(HESSIAN/JAVA)
	 *   Header(1B): 单向/双向(响应报文中不使用这个字段)
	 *   Header(1B): Reserved
	 *   Header(4B): 通信层对象长度
	 *   Header(1B): 应用层对象类名长度
	 *   Header(4B): 应用层对象长度
	 *   Body:       通信层对象
	 *   Body:       应用层对象类名
	 *   Body:       应用层对象
	 */

	connlen := len(cmd.connection)
	classlen := len(cmd.class)
	contentlen := len(cmd.content)

	b = readhelper.AllocToAtLeast(b, 14+connlen+classlen+contentlen)

	// Compiler hints
	_ = b[:14]
	b[0] = byte(cmd.proto)
	b[1] = byte(cmd.cmdcode)
	b[2] = byte(cmd.codec)
	b[3] = byte(cmd.typ)
	b[4] = 0
	binary.BigEndian.PutUint32(b[5:9], uint32(connlen))
	b[9] = byte(classlen)
	binary.BigEndian.PutUint32(b[10:14], uint32(contentlen))
	copy(b[14:14+connlen], cmd.connection)
	copy(b[14+connlen:14+connlen+classlen], cmd.class)
	copy(b[14+connlen+classlen:], cmd.content)

	return b, nil
}

func writeCommandBOLTV2(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	switch cmd.typ {
	case TypeBOLTRequest, TypeBOLTRequestOneWay:
		return writeRequestCommandBOLTV2(wo, b, cmd)
	case TypeBOLTResponse:
		return writeResponseCommandBOLTV2(wo, b, cmd)
	default:
		return nil, ErrMalformedType
	}
}

func writeResponseCommandBOLTV2(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	// 	Response command protocol for v2
	// 0     1     2     3     4           6           8          10     11    12          14          16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+-----+-----+-----+-----+
	// |proto| ver1| type| cmdcode   |ver2 |   requestId           |codec|switch|respstatus |  classLen |
	// +-----------+-----------+-----------+-----------+-----------+------------+-----------+-----------+
	// |headerLen  | contentLen            |                      ...                                   |
	// +-----------------------------------+                                                            +
	// |               className + header  + content  bytes                                             |
	// +                                                                                                +
	// |                               ... ...                                  | CRC32(optional)       |
	// +------------------------------------------------------------------------------------------------+

	classLen := len(cmd.class)
	headerLen := cmd.headers.GetEncodeSize()
	contentLen := len(cmd.content)

	b = readhelper.AllocToAtLeast(b, 22+classLen+headerLen+contentLen)

	// Compiler hints
	_ = b[:22]

	b[0] = byte(cmd.proto)
	b[1] = byte(cmd.ver1)
	b[2] = byte(cmd.typ)
	binary.BigEndian.PutUint16(b[3:], uint16(cmd.cmdcode))
	b[5] = cmd.ver2
	binary.BigEndian.PutUint32(b[6:10], cmd.rid)
	b[10] = byte(cmd.codec)
	b[11] = cmd.switc
	binary.BigEndian.PutUint16(b[12:], uint16(cmd.status))
	binary.BigEndian.PutUint16(b[14:], uint16(classLen))
	binary.BigEndian.PutUint16(b[16:], uint16(headerLen))
	binary.BigEndian.PutUint32(b[18:], uint32(contentLen))
	copy(b[22:22+classLen], cmd.class)
	cmd.headers.Encode(b[22+classLen:])
	copy(b[22+classLen+headerLen:], cmd.content)
	if cmd.switc > 0 {
		c32 := crc32.Checksum(b, crc32.IEEETable)
		b = append(b, "0123"...)
		binary.BigEndian.PutUint32(b[len(b)-4:], c32)
	}
	return b, nil
}

func writeRequestCommandBOLTV2(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	// Request command protocol for v2
	// 0     1     2           4           6           8          10     11     12          14         16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+-----+-----+-----+-----+
	// |proto| ver1|type | cmdcode   |ver2 |   requestId           |codec|switch|   timeout             |
	// +-----------+-----------+-----------+-----------+-----------+------------+-----------+-----------+
	// |classLen   |headerLen  |contentLen             |           ...                                  |
	// +-----------+-----------+-----------+-----------+                                                +
	// |               className + header  + content  bytes                                             |
	// +                                                                                                +
	// |                               ... ...                                  | CRC32(optional)       |
	// +------------------------------------------------------------------------------------------------+

	classLen := len(cmd.class)
	headerLen := cmd.headers.GetEncodeSize()
	contentLen := len(cmd.content)

	b = readhelper.AllocToAtLeast(b, 24+classLen+headerLen+contentLen)

	// Compiler hints
	_ = b[:24]

	b[0] = byte(cmd.proto)
	b[1] = byte(cmd.ver1)
	b[2] = byte(cmd.typ)
	binary.BigEndian.PutUint16(b[3:], uint16(cmd.cmdcode))
	b[5] = cmd.ver2
	binary.BigEndian.PutUint32(b[6:10], cmd.rid)
	b[10] = byte(cmd.codec)
	b[11] = cmd.switc
	binary.BigEndian.PutUint32(b[12:], cmd.timeout)
	binary.BigEndian.PutUint16(b[16:], uint16(classLen))
	binary.BigEndian.PutUint16(b[18:], uint16(headerLen))
	binary.BigEndian.PutUint32(b[20:], uint32(contentLen))
	copy(b[24:], cmd.class)
	cmd.headers.Encode(b[24+classLen:])
	copy(b[24+classLen+headerLen:], cmd.content)
	if cmd.switc > 0 {
		c32 := crc32.Checksum(b, crc32.IEEETable)
		b = append(b, "0123"...)
		binary.BigEndian.PutUint32(b[len(b)-4:], c32)
	}
	return b, nil
}

func writeCommandBOLTV1(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	switch cmd.typ {
	case TypeBOLTRequest, TypeBOLTRequestOneWay:
		return writeRequestCommandBOLTV1(wo, b, cmd)
	case TypeBOLTResponse:
		return writeResponseCommandBOLTV1(wo, b, cmd)
	default:
		return nil, ErrMalformedType
	}
}

func writeResponseCommandBOLTV1(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	// Response command protocol for v1
	// 0     1     2     3     4           6           8          10           12          14         16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
	// |proto| type| cmdcode   |ver2 |   requestId           |codec|respstatus |  classLen |headerLen  |
	// +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
	// | contentLen            |                  ... ...                                              |
	// +-----------------------+                                                                       +
	// |                          header  + content  bytes                                             |
	// +                                                                                               +
	// |                               ... ...                                                         |
	// +-----------------------------------------------------------------------------------------------+

	classLen := len(cmd.class)
	headerLen := cmd.headers.GetEncodeSize()
	contentLen := len(cmd.content)

	b = readhelper.AllocToAtLeast(b, 20+classLen+headerLen+contentLen)

	// Compiler hints
	_ = b[:20]
	b[0] = byte(cmd.proto)
	b[1] = byte(cmd.typ)
	binary.BigEndian.PutUint16(b[2:], uint16(cmd.cmdcode))
	b[4] = cmd.ver2
	binary.BigEndian.PutUint32(b[5:9], cmd.rid)
	b[9] = byte(cmd.codec)
	binary.BigEndian.PutUint16(b[10:], uint16(cmd.status))
	binary.BigEndian.PutUint16(b[12:], uint16(classLen))
	binary.BigEndian.PutUint16(b[14:], uint16(headerLen))
	binary.BigEndian.PutUint32(b[16:], uint32(contentLen))
	copy(b[20:], cmd.class)
	cmd.headers.Encode(b[20+classLen:])
	copy(b[20+classLen+headerLen:], cmd.content)

	return b, nil
}

func writeRequestCommandBOLTV1(wo *WriteOption, b []byte, cmd *Command) ([]byte, error) {
	_ = wo
	// nolint
	// Request command protocol for v1
	// 0     1     2           4           6           8          10           12          14         16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
	// |proto| type| cmdcode   |ver2 |   requestId           |codec|        timeout        |  classLen |
	// +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
	// |headerLen  | contentLen            |                             ... ...                       |
	// +-----------+-----------+-----------+                                                                                               +
	// |               className + header  + content  bytes                                            |
	// +                                                                                               +
	// |                               ... ...                                                         |
	// +-----------------------------------------------------------------------------------------------+

	classLen := len(cmd.class)
	headerLen := cmd.headers.GetEncodeSize()
	contentLen := len(cmd.content)

	b = readhelper.AllocToAtLeast(b, 22+classLen+headerLen+contentLen)

	// Compiler hints
	_ = b[:22]
	b[0] = byte(cmd.proto)
	b[1] = byte(cmd.typ)
	binary.BigEndian.PutUint16(b[2:], uint16(cmd.cmdcode))
	b[4] = cmd.ver2
	binary.BigEndian.PutUint32(b[5:9], cmd.rid)
	b[9] = byte(cmd.codec)
	binary.BigEndian.PutUint32(b[10:], cmd.timeout)
	binary.BigEndian.PutUint16(b[14:], uint16(classLen))
	binary.BigEndian.PutUint16(b[16:], uint16(headerLen))
	binary.BigEndian.PutUint32(b[18:], uint32(contentLen))
	copy(b[22:22+classLen], cmd.class)
	cmd.headers.Encode(b[22+classLen:])
	copy(b[22+classLen+headerLen:], cmd.content)

	return b, nil
}

// ReadCommand reads a command from io.Reader.
func ReadCommand(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	var (
		err error
		u8  uint8
		n   int
	)

	u8, err = readhelper.ReadUint8WithBytes(br, cmd.reserved[:])
	cmd.reserved[0] = 0
	if err != nil {
		return 0, err
	}
	cmd.proto = Proto(u8)

	switch cmd.proto {
	case 0, ProtoBOLTV1:
		n, err = readCommandBOLTV1(ro, br, cmd)
	case ProtoBOLTV2:
		n, err = readCommandBOLTV2(ro, br, cmd)
	case ProtoTBRemoting:
		n, err = readCommandTBRemoting(ro, br, cmd)
	default:
		err = ErrMalformedProto
	}

	return n, err
}

func readCommandTBRemoting(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	/**
	 *   Header(1B): 报文版本
	 *   Header(1B): 请求/响应
	 *   Header(1B): 序列化协议(HESSIAN/JAVA)
	 *   Header(1B): 单向/双向(响应报文中不使用这个字段)
	 *   Header(1B): Reserved
	 *   Header(4B): 通信层对象长度
	 *   Header(1B): 应用层对象类名长度
	 *   Header(4B): 应用层对象长度
	 *   Body:       通信层对象
	 *   Body:       应用层对象类名
	 *   Body:       应用层对象
	 */

	var (
		err        error
		connlen    uint32
		classlen   uint8
		contentlen uint32
		b32        = acquireB32()
	)

	// Compiler hints
	_ = (*b32)[:13]

	if err = readhelper.ReadToBytes(br, 13, (*b32)[:13]); err != nil {
		goto DONE
	}

	cmd.proto = ProtoTBRemoting
	cmd.cmdcode = CMDCode((*b32)[0])
	cmd.codec = Codec((*b32)[1])
	cmd.typ = Type((*b32)[2])

	cmd.reserved[0] = (*b32)[3]
	connlen = binary.BigEndian.Uint32((*b32)[4:])
	classlen = (*b32)[8]
	contentlen = binary.BigEndian.Uint32((*b32)[9:])

	cmd.connection = readhelper.AllocToAtLeast(cmd.connection, int(connlen))

	if err = readhelper.ReadToBytes(br, int(connlen), cmd.connection); err != nil {
		goto DONE
	}

	cmd.class = readhelper.AllocToAtLeast(cmd.class, int(classlen))
	if err = readhelper.ReadToBytes(br, int(classlen), cmd.class); err != nil {
		goto DONE
	}

	cmd.content = readhelper.AllocToAtLeast(cmd.content, int(contentlen))
	if err = readhelper.ReadToBytes(br, int(contentlen), cmd.content); err != nil {
		goto DONE
	}

DONE:
	releaseB32(b32)

	if err != nil {
		return 0, err
	}

	return 14 + int(connlen) + int(classlen) + int(contentlen), err
}

func readCommandBOLTV1(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	var (
		err error
		u8  uint8
	)

	b32 := acquireB32()
	u8, err = readhelper.ReadUint8WithBytes(br, b32[:1])
	releaseB32(b32)

	if err != nil {
		return 0, err
	}
	cmd.typ = Type(u8)

	switch cmd.typ {
	case TypeBOLTResponse:
		return readResponseCommandBOLTV1(ro, br, cmd)
	case TypeBOLTRequest, TypeBOLTRequestOneWay:
		return readRequestCommandBOLTV1(ro, br, cmd)
	default:
		return 0, ErrMalformedType
	}
}

func readResponseCommandBOLTV1(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	// Response command protocol for v1
	// 0     1     2     3     4           6           8          10           12          14         16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
	// |proto| type| cmdcode   |ver2 |   requestId           |codec|respstatus |  classLen |headerLen  |
	// +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
	// | contentLen            |                  ... ...                                              |
	// +-----------------------+                                                                       +
	// |                          header  + content  bytes                                             |
	// +                                                                                               +
	// |                               ... ...                                                         |
	// +-----------------------------------------------------------------------------------------------+
	var (
		err                 error
		classLen, headerLen uint16
		contentLen          uint32
		b32                 = acquireB32()
	)

	// Compiler hints
	_ = (*b32)[:18]

	if err = readhelper.ReadToBytes(br, 18, (*b32)[:18]); err != nil {
		goto DONE
	}

	cmd.cmdcode = CMDCode(binary.BigEndian.Uint16((*b32)[:2]))
	cmd.ver2 = (*b32)[3]
	cmd.rid = binary.BigEndian.Uint32((*b32)[3:7])
	cmd.codec = Codec((*b32)[7])
	cmd.status = Status(binary.BigEndian.Uint16((*b32)[8:10]))
	classLen = binary.BigEndian.Uint16((*b32)[10:12])
	headerLen = binary.BigEndian.Uint16((*b32)[12:14])
	contentLen = binary.BigEndian.Uint32((*b32)[14:18])

	cmd.class = readhelper.AllocToAtLeast(cmd.class, int(classLen))
	if err = readhelper.ReadToBytes(br, int(classLen), cmd.class); err != nil {
		goto DONE
	}

	cmd.header = readhelper.AllocToAtLeast(cmd.header, int(headerLen))
	if err = readhelper.ReadToBytes(br, int(headerLen), cmd.header); err != nil {
		goto DONE
	}

	if err = cmd.headers.Decode(cmd.header); err != nil {
		return 0, err
	}

	cmd.content = readhelper.AllocToAtLeast(cmd.content, int(contentLen))
	if err = readhelper.ReadToBytes(br, int(contentLen), cmd.content); err != nil {
		goto DONE
	}

DONE:
	releaseB32(b32)

	if err != nil {
		return 0, err
	}
	return 14 + 2 + 4 + int(classLen+headerLen) + int(contentLen), err
}

func readRequestCommandBOLTV1(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	// nolint
	// Request command protocol for v1
	// 0     1     2           4           6           8          10           12          14         16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
	// |proto| type| cmdcode   |ver2 |   requestId           |codec|        timeout        |  classLen |
	// +-----------+-----------+-----------+-----------+-----------+-----------+-----------+-----------+
	// |headerLen  | contentLen            |                             ... ...                       |
	// +-----------+-----------+-----------+                                                                                               +
	// |               className + header  + content  bytes                                            |
	// +                                                                                               +
	// |                               ... ...                                                         |
	// +-----------------------------------------------------------------------------------------------+
	var (
		err                 error
		classLen, headerLen uint16
		contentLen          uint32
		b32                 = acquireB32()
	)

	// Compiler hints
	_ = (*b32)[:20]

	if err = readhelper.ReadToBytes(br, 20, (*b32)[:]); err != nil {
		goto DONE
	}

	cmd.cmdcode = CMDCode(binary.BigEndian.Uint16((*b32)[:2]))
	cmd.ver2 = (*b32)[3]
	cmd.rid = binary.BigEndian.Uint32((*b32)[3:7])
	cmd.codec = Codec((*b32)[7])
	cmd.timeout = binary.BigEndian.Uint32((*b32)[8:12])
	classLen = binary.BigEndian.Uint16((*b32)[12:14])
	headerLen = binary.BigEndian.Uint16((*b32)[14:16])
	contentLen = binary.BigEndian.Uint32((*b32)[16:20])

	cmd.class = readhelper.AllocToAtLeast(cmd.class[:0], int(classLen))
	if err = readhelper.ReadToBytes(br, int(classLen), cmd.class); err != nil {
		goto DONE
	}

	cmd.header = readhelper.AllocToAtLeast(cmd.header[:0], int(headerLen))
	if err = readhelper.ReadToBytes(br, int(headerLen), cmd.header); err != nil {
		goto DONE
	}

	if err = cmd.headers.Decode(cmd.header); err != nil {
		return 0, err
	}

	cmd.content = readhelper.AllocToAtLeast(cmd.content[:0], int(contentLen))
	if err = readhelper.ReadToBytes(br, int(contentLen), cmd.content); err != nil {
		goto DONE
	}

DONE:
	releaseB32(b32)

	if err != nil {
		return 0, err
	}

	return 16 + 2 + 4 + int(classLen+headerLen) + int(contentLen), err
}

func readCommandBOLTV2(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	var (
		err error
		b2  [2]byte
	)

	if err = readhelper.ReadToBytes(br, 2, b2[:]); err != nil {
		return 0, err
	}

	cmd.ver1 = Version(b2[0])
	cmd.typ = Type(b2[1])

	switch cmd.typ {
	case TypeBOLTResponse:
		return readResponseCommandBOLTV2(ro, br, cmd)
	case TypeBOLTRequest, TypeBOLTRequestOneWay:
		return readRequestCommandBOLTV2(ro, br, cmd)
	default:
		return 0, ErrMalformedType
	}
}

// nolint:unparam
func readRequestCommandBOLTV2(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	// Request command protocol for v2
	// 0     1     2           4           6           8          10     11     12          14         16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+-----+-----+-----+-----+
	// |proto| ver1|type | cmdcode   |ver2 |   requestId           |codec|switch|   timeout             |
	// +-----------+-----------+-----------+-----------+-----------+------------+-----------+-----------+
	// |classLen   |headerLen  |contentLen             |           ...                                  |
	// +-----------+-----------+-----------+-----------+                                                +
	// |               className + header  + content  bytes                                             |
	// +                                                                                                +
	// |                               ... ...                                  | CRC32(optional)       |
	// +------------------------------------------------------------------------------------------------+

	var (
		err                 error
		classLen, headerLen uint16
		contentLen          uint32
		b32                 = acquireB32()
	)

	// Compiler hints
	_ = (*b32)[:21]

	if err = readhelper.ReadToBytes(br, 21, (*b32)[:]); err != nil {
		goto DONE
	}

	cmd.cmdcode = CMDCode(binary.BigEndian.Uint16((*b32)[:2]))
	cmd.ver2 = (*b32)[2]
	cmd.rid = binary.BigEndian.Uint32((*b32)[3:7])
	cmd.codec = Codec((*b32)[7])
	cmd.switc = (*b32)[8]
	cmd.timeout = binary.BigEndian.Uint32((*b32)[9:13])
	classLen = binary.BigEndian.Uint16((*b32)[13:15])
	headerLen = binary.BigEndian.Uint16((*b32)[15:17])
	contentLen = binary.BigEndian.Uint32((*b32)[17:21])

	cmd.class = readhelper.AllocToAtLeast(cmd.class, int(classLen))
	if err = readhelper.ReadToBytes(br, int(classLen), cmd.class); err != nil {
		goto DONE
	}

	cmd.header = readhelper.AllocToAtLeast(cmd.header, int(headerLen))
	if err = readhelper.ReadToBytes(br, int(headerLen), cmd.header); err != nil {
		goto DONE
	}

	if err = cmd.headers.Decode(cmd.header); err != nil {
		return 0, err
	}

	cmd.content = readhelper.AllocToAtLeast(cmd.content, int(contentLen))
	if err = readhelper.ReadToBytes(br, int(contentLen), cmd.content); err != nil {
		goto DONE
	}

	if cmd.switc > 0 {
		cmd.crc32, err = readhelper.ReadBigEndianUint32WithBytes(br, (*b32)[0:4])
		if err != nil {
			goto DONE
		}
		// TODO(detailyang): crc32 check
	}

DONE:
	releaseB32(b32)

	if err != nil {
		return 0, err
	}

	return 24 + int(classLen+headerLen) + int(contentLen), err
}

// nolint:unparam
func readResponseCommandBOLTV2(ro *ReadOption, br io.Reader, cmd *Command) (int, error) {
	_ = ro
	// nolint
	// Response command protocol for v2
	// 0     1     2     3     4           6           8          10     11    12          14          16
	// +-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+-----+-----+-----+-----+
	// |proto| ver1| type| cmdcode   |ver2 |   requestId           |codec|switch|respstatus |  classLen |
	// +-----------+-----------+-----------+-----------+-----------+------------+-----------+-----------+
	// |headerLen  | contentLen            |                      ...                                   |
	// +-----------------------------------+                                                            +
	// |               className + header  + content  bytes                                             |
	// +                                                                                                +
	// |                               ... ...                                  | CRC32(optional)       |
	// +------------------------------------------------------------------------------------------------+

	var (
		err                 error
		classLen, headerLen uint16
		contentLen          uint32
		b32                 = acquireB32()
	)

	// Compiler hints
	_ = (*b32)[:19]

	if err = readhelper.ReadToBytes(br, 19, (*b32)[:]); err != nil {
		goto DONE
	}

	cmd.cmdcode = CMDCode(binary.BigEndian.Uint16((*b32)[:2]))
	cmd.ver2 = (*b32)[2]
	cmd.rid = binary.BigEndian.Uint32((*b32)[3:7])
	cmd.codec = Codec((*b32)[7])
	cmd.switc = (*b32)[8]
	cmd.status = Status(binary.BigEndian.Uint16((*b32)[9:11]))
	classLen = binary.BigEndian.Uint16((*b32)[11:13])
	headerLen = binary.BigEndian.Uint16((*b32)[13:15])
	contentLen = binary.BigEndian.Uint32((*b32)[15:19])

	cmd.class = readhelper.AllocToAtLeast(cmd.class, int(classLen))
	if err = readhelper.ReadToBytes(br, int(classLen), cmd.class); err != nil {
		goto DONE
	}

	cmd.header = readhelper.AllocToAtLeast(cmd.header, int(headerLen))
	if err = readhelper.ReadToBytes(br, int(headerLen), cmd.header); err != nil {
		goto DONE
	}

	if err = cmd.headers.Decode(cmd.header); err != nil {
		return 0, err
	}

	cmd.content = readhelper.AllocToAtLeast(cmd.content, int(contentLen))
	if err = readhelper.ReadToBytes(br, int(contentLen), cmd.content); err != nil {
		goto DONE
	}

	if cmd.switc > 0 {
		cmd.crc32, err = readhelper.ReadBigEndianUint32WithBytes(br, (*b32)[0:4])
		if err != nil {
			goto DONE
		}
		// TODO(detailyang): crc32 check
	}

DONE:
	releaseB32(b32)

	if err != nil {
		return 0, err
	}

	return 22 + int(classLen+headerLen) + int(contentLen), err
}
