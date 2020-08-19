package ngx

import "unsafe"

type Reader interface {
	Len() int
	Bytes() []byte
	String() string
	NewBytes() []byte
	NewString() string
}

type BytesReader struct {
	buf []byte
}

func NewBytesReader(b []byte) *BytesReader {
	return &BytesReader{b}
}

func (r *BytesReader) Len() int {
	return len(r.buf)
}

func (r *BytesReader) Bytes() []byte {
	return r.buf
}

func (r *BytesReader) String() string {
	return *(*string)(unsafe.Pointer(&r.buf))
}

func (r *BytesReader) NewBytes() []byte {
	buf := make([]byte, r.Len())
	copy(buf, r.buf)
	return buf
}

func (r *BytesReader) NewString() string {
	return string(r.buf)
}

type StringReader struct {
	buf string
	cap int
}

func NewStringReader(s string) *StringReader {
	return &StringReader{s, len(s)}
}

func (r *StringReader) Len() int {
	return len(r.buf)
}

func (r *StringReader) Bytes() []byte {
	return *(*[]byte)(unsafe.Pointer(r))
}

func (r *StringReader) String() string {
	return r.buf
}

func (r *StringReader) NewBytes() []byte {
	return []byte(r.buf)
}

func (r *StringReader) NewString() string {
	return string(r.buf)
}
