package ngx

import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/modern-go/reflect2"
)

type structOp struct {
	baseOp
	Offset uintptr
	Codec  Codec
}

func codecOfStruct(ngx *NGX, typ *reflect2.UnsafeStructType) (Codec, error) {
	ops := make([]structOp, len(ngx.ops))
	for i := 0; i < len(ops); i++ {
		ops[i].baseOp = ngx.ops[i]
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name := field.Name()
		tag := field.Tag().Get("ngx")
		if name == "_" || tag == "_" {
			continue
		}

		if len(tag) > 0 {
			name = tag
		}
		if ind, ok := ngx.supported[name]; ok {
			ops[ind].Type = ngxBind
			ops[ind].Offset = field.Offset()
			dec, err := codecOf(ngx, field.Type())
			if err != nil {
				return nil, err
			}
			ops[ind].Codec = dec
		}
	}
	return &structCodec{ops, ngx.esc}, nil
}

type structCodec struct {
	ops []structOp
	esc Esc
}

func (d *structCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	length := len(d.ops)
	for i := 0; i < length; i++ {
		op := d.ops[i]
		switch op.Type {
		case ngxString, ngxEscString:
			text.Write(op.Extra)
		case ngxVariable:
			text.WriteString(d.esc.Nil())
		case ngxBind:
			bindPtr := unsafe.Pointer(uintptr(ptr) + op.Offset)
			if err := op.Codec.Encode(bindPtr, text); err != nil {
				return fmt.Errorf("field %q %v", op.Extra, err)
			}
		}
	}
	return nil
}

func (d *structCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	p := 0
	data := text.Bytes()
	length := len(d.ops)
	for i := 0; i < length; i++ {
		op := d.ops[i]
		switch op.Type {
		case ngxString, ngxEscString:
			if !bytes.HasPrefix(data[p:], op.Extra) {
				got := data[p:]
				if len(got) > len(op.Extra) {
					got = got[:len(op.Extra)]
				}
				return fmt.Errorf("got unexpected string %q, expecting %q", got, op.Extra)
			}
			p += len(op.Extra)
		case ngxVariable:
			if i+1 >= length {
				return nil
			}
			next := d.ops[i+1]
			switch next.Type {
			case ngxString:
				off := bytes.Index(data[p:], next.Extra)
				if off < 0 {
					return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
				}
				i++
				p += off + len(next.Extra)
			case ngxEscString:
			ngx_var_retry:
				off := bytes.Index(data[p:], next.Extra)
				if off < 0 {
					return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
				} else if off > 0 && data[p+off-1] == '\\' {
					if d.esc == EscJson {
						if _, err := d.esc.Unescape(data[p : p+off]); err != nil {
							p += off + len(next.Extra)
							goto ngx_var_retry
						}
					} else {
						p += off + len(next.Extra)
						goto ngx_var_retry
					}
				}
				i++
				p += off + len(next.Extra)
			default:
				return fmt.Errorf("ngx-go does not support '$%s$%s' style format", op.Extra, next.Extra)
			}
		case ngxBind:
			var (
				raw []byte
				err error
			)
			if i+1 >= length {
				raw = data[p:]
			} else {
				next := d.ops[i+1]
				switch next.Type {
				case ngxString:
					off := bytes.Index(data[p:], next.Extra)
					if off < 0 {
						return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
					}
					raw = data[p : p+off]
					i++
					p += off + len(next.Extra)
				case ngxEscString:
					oldp := p
				ngx_bind_retry:
					off := bytes.Index(data[p:], next.Extra)
					if off < 0 {
						return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
					} else if off > 0 && data[p+off-1] == '\\' {
						if d.esc == EscJson {
							if raw, err = d.esc.Unescape(data[oldp : p+off]); err == nil {
								i++
								p += off + len(next.Extra)
								goto afterUnescape
							}
						}
						p += off + len(next.Extra)
						goto ngx_bind_retry
					}
					raw = data[oldp : p+off]
					i++
					p += off + len(next.Extra)
				default:
					return fmt.Errorf("ngx-go does not support '$%s$%s' style format", op.Extra, next.Extra)
				}
			}

			raw, err = d.esc.Unescape(raw)
			if err != nil {
				return err
			}
		afterUnescape:
			bindPtr := unsafe.Pointer(uintptr(ptr) + op.Offset)

			if err := op.Codec.Decode(bindPtr, NewBytesReader(raw)); err != nil {
				return fmt.Errorf("field %q %v", op.Extra, err)
			}

		default:
			return fmt.Errorf("Unsupported operator type(%d)", op.Type)
		}
	}

	return nil
}
