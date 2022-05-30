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
	"encoding/binary"
	"errors"
	"hash/crc32"
	"unsafe"
)

var ErrNotEnough = errors.New("buffer: not enough")

type Buffer struct {
	pos  int
	data []byte
}

func New(data []byte) *Buffer {
	return &Buffer{
		pos:  0,
		data: data,
	}
}

func (b *Buffer) Remain() int {
	return len(b.data) - b.pos
}

func (b *Buffer) Check(size int) bool {
	return b.pos+size <= len(b.data)
}

func (b *Buffer) Gap() int {
	return cap(b.data) - len(b.data)
}

func (b *Buffer) CapTo(size int) {
	b.data = append(b.data, make([]byte, size)...)
}

func (b *Buffer) Append(data []byte) *Buffer {
	b.data = append(b.data, data...)
	return b
}

func (b *Buffer) Peek(i int) byte {
	return b.data[b.pos+i]
}

func (b *Buffer) MustAppend(size int, data []byte) []byte {
	data = append(data, b.data[b.pos:b.pos+size]...)
	b.pos += size
	return data
}

func (b *Buffer) Uint8() (uint8, error) {
	if !b.Check(1) {
		return 0, ErrNotEnough
	}

	u8 := b.data[b.pos]
	b.pos += 1
	return u8, nil
}

func (b *Buffer) Uint32() (uint32, error) {
	if !b.Check(4) {
		return 0, ErrNotEnough
	}

	u32 := binary.BigEndian.Uint32(b.data[b.pos:])
	b.pos += 4

	return u32, nil
}

func (b *Buffer) Uint32String() (string, error) {
	if !b.Check(4) {
		return "", ErrNotEnough
	}

	var u32 uint32
	b.MustUint32(&u32)

	if !b.Check(int(u32)) {
		return "", ErrNotEnough
	}

	var data []byte

	b.MustRef(int(u32), &data)

	return b2s(data), nil
}

func (b *Buffer) MustUint8(u8 *uint8) *Buffer {
	*u8 = b.data[b.pos]
	b.pos += 1
	return b
}

func (b *Buffer) MustUint16(u16 *uint16) *Buffer {
	// Use builtin API: It does some optimzing for compiler
	*u16 = binary.BigEndian.Uint16(b.data[b.pos:])
	b.pos += 2
	return b
}

func (b *Buffer) MustUint32(u32 *uint32) *Buffer {
	*u32 = binary.BigEndian.Uint32(b.data[b.pos:])
	b.pos += 4
	return b
}

func (b *Buffer) MustUint64(u64 *uint64) *Buffer {
	*u64 = binary.BigEndian.Uint64(b.data[b.pos:])
	b.pos += 8
	return b
}

func (b *Buffer) MustTake(dst []byte) *Buffer {
	copy(dst, b.data[b.pos:b.pos+len(dst)])
	return b
}

func (b *Buffer) MustCopy(size int, bp []byte) *Buffer {
	copy(bp, b.data[b.pos:b.pos+size])
	b.pos += size
	return b
}

func (b *Buffer) MustRef(size int, bp *[]byte) *Buffer {
	*bp = b.data[b.pos : b.pos+size]
	b.pos += size
	return b
}

func (b *Buffer) MustPutUint8(u8 uint8) *Buffer {
	b.data[b.pos] = u8
	b.pos += 1
	return b
}

func (b *Buffer) MustPutUint16(u16 uint16) *Buffer {
	binary.BigEndian.PutUint16(b.data[b.pos:], u16)
	b.pos += 2
	return b
}

func (b *Buffer) MustPutUint32(u32 uint32) *Buffer {
	binary.BigEndian.PutUint32(b.data[b.pos:], u32)
	b.pos += 4
	return b
}

//
//
// Go is sutpid and not way to inline below
func (b *Buffer) MustPutUint64(u64 uint64) *Buffer {
	binary.BigEndian.PutUint64(b.data[b.pos:], u64)
	b.pos += 8
	return b
}

func (b *Buffer) MustPutBytes(data []byte) *Buffer {
	copy(b.data[b.pos:b.pos+len(data)], data)
	b.pos += len(data)
	return b
}

func (b *Buffer) MustPutCRC32() *Buffer {
	c32 := crc32.Checksum(b.data, crc32.IEEETable)
	return b.MustPutUint32(c32)
}

func (b *Buffer) MustPutUint32String(s string) *Buffer {
	b.MustPutUint32(uint32(len(s))).MustPutBytes(s2b(s))
	return b
}

func (b *Buffer) MustSkip(n int) *Buffer {
	b.pos += n
	return b
}

func (b *Buffer) Pos() int {
	return b.pos
}

func (b *Buffer) Bytes() []byte {
	return b.data
}

// nolint
func b2s(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// nolint
func s2b(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
