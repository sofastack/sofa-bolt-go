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
	"bytes"
	"errors"
	"math"
	"strings"

	"github.com/sofastack/sofa-bolt-go/sofabolt/simplemap/buffer"
	sofalogger "github.com/sofastack/sofa-common-go/logger"
)

var (
	ErrSimleMapKeyFailed   = errors.New("simplemap: parse key failed")
	ErrSimleMapValueFailed = errors.New("simplemap: parse value failed")
)

var (
	StrHost          = "host"
	BHost            = []byte(StrHost)
	StrContentType   = "content-type"
	BContentType     = []byte(StrContentType)
	StrService       = "service"
	BService         = []byte(StrService)
	StrAuthorization = "authorization"
	BAuthorization   = []byte(StrAuthorization)
)

type kvHeader struct {
	// nolint
	noCopy noCopy

	key   []byte
	value []byte
}

func (k *kvHeader) Reset() {
	k.key = k.key[:0]
	k.value = k.value[:0]
}

func (k *kvHeader) Equal(dst *kvHeader) bool {
	if !bytes.Equal(k.key, dst.key) {
		return false
	}

	if !bytes.Equal(k.value, dst.value) {
		return false
	}

	return true
}

type FastSimpleMap struct {
	// nolint
	noCopy noCopy

	authorization []byte
	service       []byte
	host          []byte
	contentType   []byte
	kvs           []kvHeader

	raw []byte
}

func New() FastSimpleMap {
	return FastSimpleMap{}
}

func (m *FastSimpleMap) MarshalLogObject(enc sofalogger.ObjectEncoder) error {
	m.Range(func(k, v string) {
		enc.AddString(k, v)
	})
	return nil
}

func (m *FastSimpleMap) Reset() {
	m.authorization = m.authorization[:0]
	m.service = m.service[:0]
	m.host = m.host[:0]
	m.contentType = m.contentType[:0]
	m.kvs = m.kvs[:0]
}

func (m *FastSimpleMap) GetEncodeSize() int {
	l := 0

	if len(m.service) > 0 {
		l += (4 + len(StrService) + 4 + len(m.service))
	}

	if len(m.host) > 0 {
		l += (4 + len(StrHost) + 4 + len(m.host))
	}

	if len(m.contentType) > 0 {
		l += (4 + len(StrContentType) + 4 + len(m.contentType))
	}

	if len(m.authorization) > 0 {
		l += (4 + len(StrAuthorization)) + 4 + len(m.authorization)
	}

	for i := 0; i < len(m.kvs); i++ {
		l += (4 + len(m.kvs[i].key) + 4 + len(m.kvs[i].value))
	}

	return l
}

func (m *FastSimpleMap) CopyTo(dst *FastSimpleMap) {
	dst.host = append(dst.host[:0], m.host...)
	dst.service = append(dst.service[:0], m.service...)
	dst.contentType = append(dst.contentType[:0], m.contentType...)
	dst.authorization = append(dst.authorization[:0], m.authorization...)

	n := len(m.kvs)

	if cap(dst.kvs) < n {
		tmp := make([]kvHeader, n)
		copy(tmp, dst.kvs)
		dst.kvs = tmp
	}

	dst.kvs = dst.kvs[:n]
	for i := 0; i < n; i++ {
		dst.kvs[i].key = append(dst.kvs[i].key[:0], m.kvs[i].key...)
		dst.kvs[i].value = append(dst.kvs[i].value[:0], m.kvs[i].value...)
	}
}

func (m *FastSimpleMap) Encode(data []byte) int {
	b := buffer.New(data)

	if len(m.service) > 0 {
		b.MustPutUint32String(StrService).MustPutUint32String(b2s(m.service))
	}

	if len(m.host) > 0 {
		b.MustPutUint32String(StrHost).MustPutUint32String(b2s(m.host))
	}

	if len(m.contentType) > 0 {
		b.MustPutUint32String(StrContentType).MustPutUint32String(b2s(m.contentType))
	}

	if len(m.authorization) > 0 {
		b.MustPutUint32String(StrAuthorization).MustPutUint32String(b2s(m.authorization))
	}

	for i := 0; i < len(m.kvs); i++ {
		b.MustPutUint32String(b2s(m.kvs[i].key)).MustPutUint32String(b2s(m.kvs[i].value))
	}

	return b.Pos()
}

func (m *FastSimpleMap) Equal(dst *FastSimpleMap) bool {
	if len(m.kvs) != len(dst.kvs) {
		return false
	}

	if !bytes.Equal(m.host, dst.host) {
		return false
	}

	if !bytes.Equal(m.service, dst.service) {
		return false
	}

	if !bytes.Equal(m.contentType, dst.contentType) {
		return false
	}

	if !bytes.Equal(m.authorization, dst.authorization) {
		return false
	}

	for i := 0; i < len(m.kvs); i++ {
		if !m.kvs[i].Equal(&dst.kvs[i]) {
			return false
		}
	}

	return true
}

func (m *FastSimpleMap) Decode(data []byte) error {
	m.raw = data
	b := buffer.New(m.raw)
	l := len(m.raw)

	var (
		k   string
		v   string
		d   []byte
		err error
	)

	for l != 0 {
		k, err = b.Uint32String()
		if err != nil {
			return ErrSimleMapKeyFailed
		}

		l -= 4
		l -= len(k)

		u32, err := b.Uint32()
		if err != nil {
			return ErrSimleMapValueFailed
		}

		if u32 == math.MaxUint32 { // Real Null
			v = ""
		} else if u32 == 0 {
			v = ""
		} else {
			if b.Check(int(u32)) == false {
				return ErrSimleMapValueFailed
			}

			b.MustRef(int(u32), &d)
			v = b2s(d)
		}

		l -= 4
		l -= len(v)

		m.set(s2b(k), s2b(v))
	}

	return nil
}

func (m *FastSimpleMap) set(k, v []byte) {
	switch b2s(k) {
	case StrHost:
		m.setHost(v)
		return
	case StrContentType:
		m.setContentType(v)
		return
	case StrService:
		m.setService(v)
		return
	case StrAuthorization:
		m.setAuthorization(v)
		return
	default:
		n := len(m.kvs)
		var kv *kvHeader
		for i := 0; i < n; i++ {
			kv = &m.kvs[i]
			// Until the Go compiler is smart to fix bytes.Equal
			if b2s(kv.key) == b2s(k) {
				kv.value = append(kv.value[:0], v...)
				return
			}
		}

		if cap(m.kvs) <= n {
			m.kvs = append(m.kvs, make([]kvHeader, 4)...)
		}
		m.kvs = m.kvs[:n+1]

		m.kvs[n].key = append(m.kvs[n].key[:0], k...)
		m.kvs[n].value = append(m.kvs[n].value[:0], v...)
	}
}

func (m *FastSimpleMap) Get(k string) string {
	return b2s(m.get(s2b(k)))
}

func (m *FastSimpleMap) Set(k, v string) *FastSimpleMap {
	m.set(s2b(k), s2b(v))
	return m
}

func (m *FastSimpleMap) Del(k string) {
	m.del(s2b(k))
}

func (m *FastSimpleMap) Range(fn func(k, v string)) {
	if len(m.service) > 0 {
		fn(StrService, b2s(m.service))
	}

	if len(m.host) > 0 {
		fn(StrHost, b2s(m.host))
	}

	if len(m.contentType) > 0 {
		fn(StrContentType, b2s(m.contentType))
	}

	if len(m.authorization) > 0 {
		fn(StrAuthorization, b2s(m.authorization))
	}

	for i := 0; i < len(m.kvs); i++ {
		fn(b2s(m.kvs[i].key), b2s(m.kvs[i].value))
	}
}

func (m *FastSimpleMap) del(k []byte) {
	switch b2s(k) {
	case StrHost:
		m.host = m.host[:0]
		return
	case StrContentType:
		m.contentType = m.contentType[:0]
		return
	case StrService:
		m.service = m.service[:0]
		return
	case StrAuthorization:
		m.authorization = m.authorization[:0]
		return
	}

	for i := 0; i < len(m.kvs); i++ {
		if bytes.Equal(m.kvs[i].key, k) {
			m.kvs = append(m.kvs[:i], m.kvs[i+1:]...)
		}
	}
}

func (m *FastSimpleMap) get(k []byte) []byte {
	switch b2s(k) {
	case StrHost:
		return m.getHost()
	case StrContentType:
		return m.getContentType()
	case StrService:
		return m.getService()
	case StrAuthorization:
		return m.getAuthenticate()
	}

	for i := 0; i < len(m.kvs); i++ {
		if bytes.Equal(m.kvs[i].key, k) {
			return m.kvs[i].value
		}
	}

	return nil
}

func (m *FastSimpleMap) getAuthenticate() []byte {
	return m.authorization
}

func (m *FastSimpleMap) setAuthorization(authorization []byte) {
	m.authorization = append(m.authorization[:0], authorization...)
}

func (m *FastSimpleMap) getService() []byte {
	return m.service
}

func (m *FastSimpleMap) setService(service []byte) {
	m.service = append(m.service[:0], service...)
}

func (m *FastSimpleMap) getHost() []byte {
	return m.host
}

func (m *FastSimpleMap) setHost(host []byte) {
	m.host = append(m.host[:0], host...)
}

func (m *FastSimpleMap) getContentType() []byte {
	return m.contentType
}

func (m *FastSimpleMap) setContentType(contentType []byte) {
	m.contentType = append(m.contentType[:0], contentType...)
}

func (m *FastSimpleMap) String() string {
	return m.Dump([]byte("&"))
}

// Dump dumps the map to string.
//
// nolint
func (m *FastSimpleMap) Dump(sep []byte) string {
	var b strings.Builder

	if len(m.service) > 0 {
		b.WriteString(StrService)
		b.WriteByte(':')
		b.Write(m.service)
		b.Write(sep)
	}

	if len(m.host) > 0 {
		b.WriteString(StrHost)
		b.WriteByte(':')
		b.Write(m.host)
		b.Write(sep)
	}

	if len(m.contentType) > 0 {
		b.WriteString(StrContentType)
		b.WriteByte(':')
		b.Write(m.contentType)
		b.Write(sep)
	}

	if len(m.authorization) > 0 {
		b.WriteString(StrAuthorization)
		b.WriteByte(':')
		b.Write(m.authorization)
		b.Write(sep)
	}

	for i := 0; i < len(m.kvs); i++ {
		b.Write(m.kvs[i].key)
		b.WriteByte(':')
		b.Write(m.kvs[i].value)
		b.Write(sep)
	}

	return b.String()
}
