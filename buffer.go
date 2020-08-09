package ngx

import "unsafe"

type Buffer interface {
	Len() int
	Bytes() []byte
	String() string
	NewBytes() []byte
}

type Bytes struct {
	buf []byte
}

func NewBytesBuffer(b []byte) *Bytes {
	return &Bytes{b}
}

func (b *Bytes) Len() int {
	return len(b.buf)
}

func (b *Bytes) Bytes() []byte {
	return b.buf
}

func (b *Bytes) String() string {
	return *(*string)(unsafe.Pointer(&b.buf))
}

func (b *Bytes) NewBytes() []byte {
	buf := make([]byte, b.Len())
	copy(buf, b.buf)
	return buf
}

type String struct {
	buf string
	cap int
}

func NewStringBuffer(s string) *String {
	return &String{s, len(s)}
}

func (s *String) Len() int {
	return len(s.buf)
}

func (s *String) Bytes() []byte {
	return *(*[]byte)(unsafe.Pointer(s))
}

func (s *String) String() string {
	return s.buf
}

func (s *String) NewBytes() []byte {
	return []byte(s.buf)
}
