package ngx

import (
	"bytes"
	"errors"
	"fmt"
	"unicode/utf16"
)

const maxLatin1 = 255

const (
	EscDefault = Esc(iota)
	EscJson
	EscNone
)

type Esc int

func (e Esc) String() string {
	switch e {
	case EscDefault:
		return "default"
	case EscJson:
		return "json"
	case EscNone:
		return "none"
	default:
		return "unknown"
	}
}

func (e Esc) isEscapeChar(ch byte) bool {
	switch e {
	case EscDefault:
		switch ch {
		case '\\', '"', 'x':
			return true
		default:
			return false
		}
	case EscJson:
		switch ch {
		case '\\', '"', 'n', 'r', 't', 'b', 'f', 'u':
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (e Esc) Escape(buf []byte) []byte {
	switch e {
	case EscDefault:
		return escape(buf)
	case EscJson:
		return jescape(buf)
	default:
		return buf
	}
}

func (e Esc) Unescape(buf []byte) ([]byte, error) {
	switch e {
	case EscDefault:
		return unescape(buf)
	case EscJson:
		return junescape(buf)
	default:
		return buf, nil
	}
}

func (e Esc) Nil() string {
	switch e {
	case EscDefault:
		return "-"
	case EscJson:
		return "null"
	default:
		return ""
	}
}

var heximal = [maxLatin1 + 1]int8{}

func init() {
	for i := 0; i <= maxLatin1; i++ {
		if i >= 'a' && i <= 'f' {
			heximal[i] = int8(i) - 'a' + 10
		} else if i >= 'A' && i <= 'F' {
			heximal[i] = int8(i) - 'A' + 10
		} else if i >= '0' && i <= '9' {
			heximal[i] = int8(i) - '0'
		} else {
			heximal[i] = -1
		}
	}
}

func escape(buf []byte) []byte {
	length := len(buf)
	if length <= 0 {
		return buf
	}
	w := AcquireWriter()

	for i := 0; i < length; i++ {
		ch := buf[i]
		if ch < 0x20 {
			w.WriteString(`\x`)
			w.WriteByte('0' + ch>>4)
			ch &= 0xF
			if ch < 10 {
				w.WriteByte('0' + ch)
			} else {
				w.WriteByte('A' + ch - 10)
			}
		} else {
			if ch == '\\' || ch == '"' {
				w.WriteByte('\\')
			}
			w.WriteByte(ch)
		}
	}

	esc := w.CopyBytes()
	ReleaseWriter(w)
	return esc
}

func unescape(buf []byte) ([]byte, error) {
	length := len(buf)
	if length <= 0 {
		return buf, nil
	}

	w := AcquireWriter()

	for i := 0; i < length; i++ {
		backslash := bytes.IndexByte(buf[i:], '\\')
		if backslash < 0 {
			w.Write(buf[i:])
			break
		} else {
			backslash += i
			w.Write(buf[i:backslash])
		}

		backslash++
		if backslash >= length {
			return nil, errors.New("found EOF while unescaping '\\' format")
		}
		switch ch := buf[backslash]; ch {
		case '\\', '"':
			w.WriteByte(ch)
		case 'x':
			if backslash+2 < length {
				if heximal[buf[backslash+1]] >= 0 && heximal[buf[backslash+2]] >= 0 {
					w.WriteByte(byte(heximal[buf[backslash+1]]<<4 | heximal[buf[backslash+1]]))
					backslash += 2
				} else {
					return nil, fmt.Errorf("found invalid hex escape format \\x%c%c", buf[backslash+1], buf[backslash+2])
				}
			} else {
				return nil, errors.New("found EOF while unescaping '\\x??' format")
			}
		default:
			return nil, fmt.Errorf("found unknown escape format '\\%c'", ch)
		}
		i = backslash
	}

	raw := w.CopyBytes()
	ReleaseWriter(w)
	return raw, nil
}

func jescape(buf []byte) []byte {
	length := len(buf)
	if length <= 0 {
		return buf
	}

	w := AcquireWriter()

	for i := 0; i < length; i++ {
		ch := buf[i]
		if ch < 0x20 {
			w.WriteByte('\\')
			switch ch {
			case '\n':
				w.WriteByte('n')
			case '\r':
				w.WriteByte('r')
			case '\t':
				w.WriteByte('t')
			case '\b':
				w.WriteByte('b')
			case '\f':
				w.WriteByte('f')
			default:
				w.WriteByte('0')
				w.WriteByte('0')
				w.WriteByte('u')
				w.WriteByte('0' + ch>>4)
				ch &= 0xF
				if ch < 10 {
					w.WriteByte('0' + ch)
				} else {
					w.WriteByte('A' + ch - 10)
				}
			}
		} else {
			if ch == '\\' || ch == '"' {
				w.WriteByte('\\')
			}
			w.WriteByte(ch)
		}
	}

	esc := w.CopyBytes()
	ReleaseWriter(w)
	return esc
}

func junescape(buf []byte) ([]byte, error) {
	length := len(buf)
	if length <= 0 {
		return buf, nil
	}

	w := AcquireWriter()

	for i := 0; i < length; i++ {
		backslash := bytes.IndexByte(buf[i:], '\\')
		if backslash < 0 {
			w.Write(buf[i:])
			break
		} else {
			backslash += i
			w.Write(buf[i:backslash])
		}

		backslash++
		if backslash >= length {
			return nil, errors.New("found EOF while unescaping '\\' format")
		}
		switch ch := buf[backslash]; ch {
		case '\\', '"':
			w.WriteByte(ch)
		case 'n':
			w.WriteByte('\n')
		case 'r':
			w.WriteByte('\r')
		case 't':
			w.WriteByte('\t')
		case 'b':
			w.WriteByte('\b')
		case 'f':
			w.WriteByte('\f')
		case 'u':
			if backslash+4 < length {
				if heximal[buf[backslash+1]] >= 0 && heximal[buf[backslash+2]] >= 0 && heximal[buf[backslash+3]] >= 0 && heximal[buf[backslash+4]] >= 0 {
					var r rune
					for j := 1; j <= 4; j++ {
						r = r<<4 | rune(heximal[buf[backslash+j]])
					}
					if utf16.IsSurrogate(r) {
						/*
								\ud800     \u????
								 ^         ^
							    backslash  next
						*/
						if next := backslash + 5; next+5 < length && buf[next] == '\\' && buf[next+1] == 'u' {
							if heximal[buf[next+2]] >= 0 && heximal[buf[next+3]] >= 0 && heximal[buf[next+4]] >= 0 && heximal[buf[next+5]] >= 0 {
								var r2 rune
								for j := 2; j <= 5; j++ {
									r2 = r2<<4 | rune(heximal[buf[next+j]])
								}
								combined := utf16.DecodeRune(r, r2)
								if combined == '\uFFFD' {
									appendRune(w, r)
									appendRune(w, r2)
								} else {
									appendRune(w, combined)
								}
								backslash = next + 1
							} else {
								return nil, fmt.Errorf("found invalid unicode escape format \\u%c%c%c%c", buf[next+2], buf[next+3], buf[next+4], buf[next+5])
							}
						} else {
							appendRune(w, r)
						}
					} else {
						appendRune(w, r)
					}
					backslash += 4
				} else {
					return nil, fmt.Errorf("found invalid unicode escape format \\u%c%c%c%c", buf[backslash+1], buf[backslash+2], buf[backslash+3], buf[backslash+4])
				}
			} else {
				return nil, errors.New("found EOF while unescaping '\\u??' format")
			}
		default:
			return nil, fmt.Errorf("found unknown escape format '\\%c'", ch)
		}
		i = backslash
	}

	raw := w.CopyBytes()
	ReleaseWriter(w)
	return raw, nil
}

const (
	t1 = 0x00 // 0000 0000
	tx = 0x80 // 1000 0000
	t2 = 0xC0 // 1100 0000
	t3 = 0xE0 // 1110 0000
	t4 = 0xF0 // 1111 0000
	t5 = 0xF8 // 1111 1000

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1

	surrogateMin = 0xD800
	surrogateMax = 0xDFFF

	maxRune   = '\U0010FFFF' // Maximum valid Unicode code point.
	runeError = '\uFFFD'     // the "error" Rune or "Unicode replacement character"
)

func appendRune(w Writer, r rune) {
	switch i := uint32(r); {
	case i <= rune1Max:
		w.WriteByte(byte(r))
	case i <= rune2Max:
		w.WriteByte(t2 | byte(r>>6))
		w.WriteByte(tx | byte(r)&maskx)
	case i > maxRune, surrogateMin <= i && i <= surrogateMax:
		r = runeError
		fallthrough
	case i <= rune3Max:
		w.WriteByte(t3 | byte(r>>12))
		w.WriteByte(tx | byte(r>>6)&maskx)
		w.WriteByte(tx | byte(r)&maskx)
	default:
		w.WriteByte(t4 | byte(r>>18))
		w.WriteByte(tx | byte(r>>12)&maskx)
		w.WriteByte(tx | byte(r>>6)&maskx)
		w.WriteByte(tx | byte(r)&maskx)
	}
}
