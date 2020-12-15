package ngx

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const (
	ngxUnknown = iota
	ngxString
	ngxEscString
	ngxVariable
	ngxBind
)

var (
	ErrInvalidLogFormat         = errors.New("Invalid log format")
	ErrUnknownLogFormatEscaping = errors.New("Unknown log format escaping")
)

type baseOp struct {
	Type  int
	Extra []byte
}

func Compile(logfmt string) (*NGX, error) {
	q, p := 0, 0
	ngx := &NGX{
		ops:       make([]baseOp, 0, 8),
		supported: make(map[string]int),
	}

	if strings.HasPrefix(logfmt, "escape=") {
		p += 7
		if strings.HasPrefix(logfmt[p:], "json") {
			p += 4
			ngx.esc = EscJson
		} else if strings.HasPrefix(logfmt[p:], "default") {
			p += 7
			ngx.esc = EscDefault
		} else if strings.HasPrefix(logfmt[p:], "none") {
			p += 4
			ngx.esc = EscNone
		} else {
			return nil, ErrUnknownLogFormatEscaping
		}
	skip_semi:
		for p < len(logfmt) {
			switch logfmt[p] {
			case ' ', '\r', '\n', '\t', '\v', '\f':
				p++
			case ';':
				p++
				break skip_semi
			default:
				return nil, fmt.Errorf("expecting ';' after escape=%s", ngx.esc)
			}
		}
	}

	last := bytes.NewBuffer(nil)

	for q = p; p < len(logfmt); {
		if logfmt[p] == '$' {
			p++
			bracket := false
			if p >= len(logfmt) {
				return nil, ErrInvalidLogFormat
			}
			if logfmt[p] == '$' {
				last.WriteByte('$')
				p++
				q = p
				continue
			} else if logfmt[p] == '{' {
				bracket = true
				p++
				if p >= len(logfmt) {
					return nil, ErrInvalidLogFormat
				}
			}
			if last.Len() > 0 {
				typ := ngxString
				lastBuf := last.Bytes()
				if ngx.esc.isEscapeChar(lastBuf[0]) {
					typ = ngxEscString
				}
				ngx.ops = append(ngx.ops, baseOp{
					Type:  typ,
					Extra: lastBuf,
				})
				last = bytes.NewBuffer(nil)
			}
		loop:
			for q = p; p < len(logfmt); p++ {
				ch := logfmt[p]
				switch {
				case bracket && ch == '}':
					p++
					bracket = false
					break loop
				case (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '.':
				default:
					break loop
				}
			}
			if bracket {
				return nil, fmt.Errorf("the closing bracket of variable %q is missing", logfmt[q:p])
			}
			// if p-q <= 0 {
			// 	return nil, ErrInvalidLogFormat
			// }
			varname := logfmt[q:p]
			if len(varname) <= 0 || varname == "}" {
				return nil, ErrInvalidLogFormat
			}
			if varname[len(varname)-1] == '}' {
				varname = varname[:len(varname)-1]
			}
			varlen := len(varname)
			if varlen <= 0 {
				return nil, ErrInvalidLogFormat
			}
			if varname[0] == '.' {
				return nil, fmt.Errorf("variable %q cannot start with '.'", varname)
			}
			if varname[varlen-1] == '.' {
				return nil, fmt.Errorf("variable %q cannot end with '.'", varname)
			}
			if strings.Contains(varname, "..") {
				return nil, fmt.Errorf("variable %q cannot have consecutive dots", varname)
			}
			pos := len(ngx.ops)
			if pos > 0 && ngx.ops[pos-1].Type == ngxVariable {
				// skip
			} else {
				ngx.supported[varname] = pos
				ngx.ops = append(ngx.ops, baseOp{
					Type:  ngxVariable,
					Extra: []byte(varname),
				})
			}
			q = p
		} else {
			next := strings.IndexByte(logfmt[q:], '$')

			if next > 0 {
				last.WriteString(logfmt[q : q+next])
				q += next
				p = q
			} else {
				last.WriteString(logfmt[q:])
				break
			}

		}
	}

	if last.Len() > 0 {
		typ := ngxString
		lastBuf := last.Bytes()
		if ngx.esc.isEscapeChar(lastBuf[0]) {
			typ = ngxEscString
		}
		ngx.ops = append(ngx.ops, baseOp{
			Type:  typ,
			Extra: lastBuf,
		})
	}

	return ngx, nil
}
