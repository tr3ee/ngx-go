package ngx

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidLogFormat         = errors.New("Invalid log format")
	ErrUnknownLogFormatEscaping = errors.New("Unknown log format escaping")
)

func Compile(logfmt string) (*NGX, error) {
	q, p := 0, 0
	ngx := &NGX{
		ops:       make([]operator, 0, 8),
		supported: make(map[string]int),
	}

	if strings.HasPrefix(logfmt, "escape=") {
		p += 7
		if strings.HasPrefix(logfmt[p:], "json") {
			p += 4
			ngx.jescape = true
		} else if strings.HasPrefix(logfmt[p:], "default") {
			p += 7
			ngx.jescape = false
		} else {
			return nil, ErrUnknownLogFormatEscaping
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
			index := len(ngx.ops)
			if index > 0 && ngx.ops[index-1].Type == ngxVariable {
				// skip
			} else {
				varname := logfmt[q:p]
				ngx.supported[varname] = index
				ngx.ops = append(ngx.ops, operator{
					Type:       ngxVariable,
					Extra:      varname,
					ExtraBytes: []byte(varname),
				})
			}
			q = p
		} else {
			typ := ngxString
			next := strings.IndexByte(logfmt[q:], '$')
			if ngx.jescape {
				if isJEscapeChar(logfmt[q]) {
					typ = ngxEscString
				}
			} else if isEscapeChar(logfmt[q]) {
				typ = ngxEscString
			}

			if next < 0 {
				ngx.ops = append(ngx.ops, operator{
					Type:       typ,
					Extra:      logfmt[q:],
					ExtraBytes: []byte(logfmt[q:]),
				})
				break
			} else {
				ngx.ops = append(ngx.ops, operator{
					Type:       typ,
					Extra:      logfmt[q : q+next],
					ExtraBytes: []byte(logfmt[q : q+next]),
				})
				q += next
				p = q
			}
		}
	}
	return ngx, nil
}
