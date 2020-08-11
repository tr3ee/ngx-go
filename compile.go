package ngx

import (
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
				// skip
			case ';':
				break skip_semi
			default:
				return nil, fmt.Errorf("expecting ';' after escape=%s", ngx.esc)
			}
			p++
		}
	}

	for q = p; p < len(logfmt); {
		if logfmt[p] == '$' {
			p++
			bracket := false
			if p >= len(logfmt) {
				return nil, ErrInvalidLogFormat
			}
			if logfmt[p] == '{' {
				bracket = true
				p++
				if p >= len(logfmt) {
					return nil, ErrInvalidLogFormat
				}
			}
		loop:
			for q = p; p < len(logfmt); p++ {
				ch := logfmt[p]
				switch {
				case bracket && ch == '}':
					p++
					bracket = false
					break loop
				case (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_':
				default:
					break loop
				}
			}
			if bracket {
				return nil, fmt.Errorf("the closing bracket in %q variable is missing", logfmt[q:p])
			}
			if p-q <= 0 {
				return nil, ErrInvalidLogFormat
			}
			pos := len(ngx.ops)
			if pos > 0 && ngx.ops[pos-1].Type == ngxVariable {
				// skip
			} else {
				varname := logfmt[q:p]
				ngx.supported[varname] = pos
				ngx.ops = append(ngx.ops, baseOp{
					Type:  ngxVariable,
					Extra: []byte(varname),
				})
			}
			q = p
		} else {
			typ := ngxString
			next := strings.IndexByte(logfmt[q:], '$')
			if ngx.esc.isEscapeChar(logfmt[q]) {
				typ = ngxEscString
			}

			if next < 0 {
				ngx.ops = append(ngx.ops, baseOp{
					Type:  typ,
					Extra: []byte(logfmt[q:]),
				})
				break
			} else {
				ngx.ops = append(ngx.ops, baseOp{
					Type:  typ,
					Extra: []byte(logfmt[q : q+next]),
				})
				q += next
				p = q
			}
		}
	}
	return ngx, nil
}
