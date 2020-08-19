package ngx

import (
	"sync"
	"unicode/utf8"
	"unsafe"
)

type Writer interface {
	Len() int
	Cap() int
	Grow(n int)
	Bytes() []byte
	String() string
	Write(p []byte) (int, error)
	WriteByte(c byte) error
	WriteRune(r rune) (int, error)
	WriteString(s string) (int, error)
}

var writerPool = &sync.Pool{
	New: func() interface{} {
		return &writer{
			buf: make([]byte, 0, 0x200),
		}
	},
}

func AcquireWriter() *writer {
	w := writerPool.Get().(*writer)
	w.Reset()
	return w
}

func ReleaseWriter(w *writer) {
	if cap(w.buf) < 1<<16 { // 64k
		writerPool.Put(w)
	}
}

// A writer is used to efficiently build a string using Write methods.
// It minimizes memory copying. The zero value is ready to use.
// Do not copy a non-zero writer.
//
// Copied from https://golang.org/src/strings/builder.go
type writer struct {
	addr *writer // of receiver, to detect copies by value
	buf  []byte
}

// noescape hides a pointer from escape analysis.  noescape is
// the identity function but escape analysis doesn't think the
// output depends on the input. noescape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
// This was copied from the runtime; see issues 23382 and 7921.
//go:nosplit
//go:nocheckptr
func noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func (w *writer) copyCheck() {
	if w.addr == nil {
		// This hack works around a failing of Go's escape analysis
		// that was causing b to escape and be heap allocated.
		// See issue 23382.
		// TODO: once issue 7921 is fixed, this should be reverted to
		// just "w.addr = b".
		w.addr = (*writer)(noescape(unsafe.Pointer(w)))
	} else if w.addr != w {
		panic("strings: illegal use of non-zero writer copied by value")
	}
}

// Bytes returns the accumulated bytes.
func (w *writer) Bytes() []byte {
	return w.buf
}

// String returns the accumulated string.
func (w *writer) String() string {
	return *(*string)(unsafe.Pointer(&w.buf))
}

// Len returns the number of accumulated bytes; w.Len() == len(w.String()).
func (w *writer) Len() int { return len(w.buf) }

// Cap returns the capacity of the builder's underlying byte slice. It is the
// total space allocated for the string being built and includes any bytes
// already written.
func (w *writer) Cap() int { return cap(w.buf) }

// grow copies the buffer to a new, larger buffer so that there are at least n
// bytes of capacity beyond len(w.buf).
func (w *writer) grow(n int) {
	buf := make([]byte, len(w.buf), 2*cap(w.buf)+n)
	copy(buf, w.buf)
	w.buf = buf
}

// Grow grows b's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to b
// without another allocation. If n is negative, Grow panics.
func (w *writer) Grow(n int) {
	w.copyCheck()
	if n < 0 {
		panic("strings.writer.Grow: negative count")
	}
	if cap(w.buf)-len(w.buf) < n {
		w.grow(n)
	}
}

// Write appends the contents of p to b's buffer.
// Write always returns len(p), nil.
func (w *writer) Write(p []byte) (int, error) {
	w.copyCheck()
	w.buf = append(w.buf, p...)
	return len(p), nil
}

// WriteByte appends the byte c to b's buffer.
// The returned error is always nil.
func (w *writer) WriteByte(c byte) error {
	w.copyCheck()
	w.buf = append(w.buf, c)
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to b's buffer.
// It returns the length of r and a nil error.
func (w *writer) WriteRune(r rune) (int, error) {
	w.copyCheck()
	if r < utf8.RuneSelf {
		w.buf = append(w.buf, byte(r))
		return 1, nil
	}
	l := len(w.buf)
	if cap(w.buf)-l < utf8.UTFMax {
		w.grow(utf8.UTFMax)
	}
	n := utf8.EncodeRune(w.buf[l:l+utf8.UTFMax], r)
	w.buf = w.buf[:l+n]
	return n, nil
}

// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (w *writer) WriteString(s string) (int, error) {
	w.copyCheck()
	w.buf = append(w.buf, s...)
	return len(s), nil
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
func (w *writer) Reset() {
	w.buf = w.buf[:0]
}
