package ngx

import (
	"bytes"
	"errors"
	"fmt"
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

func (e Esc) Unescape(buf Buffer) (Buffer, error) {
	switch e {
	case EscDefault:
		return unescape(buf)
	case EscJson:
		return junescape(buf)
	default:
		return buf, nil
	}
}

var heximal = [maxLatin1 + 1]int8{}

func init() {
	for i := 0; i <= maxLatin1; i++ {
		if i >= 'a' && i <= 'f' {
			heximal[i] = int8(i) - 'a'
		} else if i >= 'A' && i <= 'F' {
			heximal[i] = int8(i) - 'A'
		} else if i >= '0' && i <= '9' {
			heximal[i] = int8(i) - '0'
		} else {
			heximal[i] = -1
		}
	}
}

func escape(buf Buffer) Buffer {
	return buf
}

func unescape(buf Buffer) (Buffer, error) {
	if buf.Len() <= 0 {
		return buf, nil
	}
	raw := buf.Bytes()
	length := len(raw)
	esc := bytes.NewBuffer(make([]byte, 0, length))

	for i := 0; i < length; i++ {
		backslash := bytes.IndexByte(raw[i:], '\\')
		if backslash < 0 {
			esc.Write(raw[i:])
			break
		} else {
			backslash += i
			esc.Write(raw[i:backslash])
		}

		backslash++
		if backslash >= length {
			return nil, errors.New("found EOF while unescaping '\\' format")
		}
		switch ch := raw[backslash]; ch {
		case '\\', '"':
			esc.WriteByte(ch)
		case 'x':
			if backslash+2 < length {
				if heximal[raw[backslash+1]] >= 0 && heximal[raw[backslash+2]] >= 0 {
					esc.WriteByte(byte(heximal[raw[backslash+1]]<<4 | heximal[raw[backslash+1]]))
					backslash += 2
				} else {
					return nil, fmt.Errorf("found invalid hex escape format \\x%c%c", raw[backslash+1], raw[backslash+2])
				}
			} else {
				return nil, errors.New("found EOF while unescaping '\\x??' format")
			}
		default:
			return nil, fmt.Errorf("found unknown escape format '\\%c'", ch)
		}
		i = backslash
	}
	return NewBytesBuffer(esc.Bytes()), nil
}

func jescape(buf Buffer) Buffer {
	return buf
}

func junescape(buf Buffer) (Buffer, error) {
	if buf.Len() <= 0 {
		return buf, nil
	}
	raw := buf.Bytes()
	length := len(raw)
	esc := bytes.NewBuffer(make([]byte, 0, length))

	for i := 0; i < length; i++ {
		backslash := bytes.IndexByte(raw[i:], '\\')
		if backslash < 0 {
			esc.Write(raw[i:])
			break
		} else {
			backslash += i
			esc.Write(raw[i:backslash])
		}

		backslash++
		if backslash >= length {
			return nil, errors.New("found EOF while unescaping '\\' format")
		}
		switch ch := raw[backslash]; ch {
		case '\\', '"':
			esc.WriteByte(ch)
		case 'n':
			esc.WriteByte('\n')
		case 'r':
			esc.WriteByte('\r')
		case 't':
			esc.WriteByte('\t')
		case 'b':
			esc.WriteByte('\b')
		case 'f':
			esc.WriteByte('\f')
		case 'u':
			if backslash+4 < length {
				if heximal[raw[backslash+1]] >= 0 && heximal[raw[backslash+2]] >= 0 && heximal[raw[backslash+3]] >= 0 && heximal[raw[backslash+4]] >= 0 {
					var r rune
					for j := 1; j <= 4; j++ {
						r = r<<4 | rune(heximal[raw[backslash+j]])
					}
					esc.WriteRune(r)
					backslash += 4
				} else {
					return nil, fmt.Errorf("found invalid unicode escape format \\u%c%c%c%c", raw[backslash+1], raw[backslash+2], raw[backslash+3], raw[backslash+4])
				}
			} else {
				return nil, errors.New("found EOF while unescaping '\\u??' format")
			}
		default:
			return nil, fmt.Errorf("found unknown escape format '\\%c'", ch)
		}
		i = backslash
	}
	return NewBytesBuffer(esc.Bytes()), nil
}
