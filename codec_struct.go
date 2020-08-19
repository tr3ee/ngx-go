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
	// copy(ops, ngx.ops)
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
	return &StructCodec{ops, ngx.esc}, nil
}

type StructCodec struct {
	ops []structOp
	esc Esc
}

func (d *StructCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
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

func (d *StructCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
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
					p += off + len(next.Extra)
					goto ngx_var_retry
				}
				i++
				p += off + len(next.Extra)
			default:
				return fmt.Errorf("ngx-go does not support '$%s$%s' style format", op.Extra, next.Extra)
			}
		case ngxBind:
			var raw []byte
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

			text := Buffer(NewBytesBuffer(raw))
			if raw, err := d.esc.Unescape(text); err != nil {
				return err
			} else {
				text = raw
			}

			bindPtr := unsafe.Pointer(uintptr(ptr) + op.Offset)

			if err := op.Codec.Decode(bindPtr, text); err != nil {
				return fmt.Errorf("field %q %v", op.Extra, err)
			}

		default:
			return fmt.Errorf("Unsupported operator type(%d)", op.Type)
		}
	}

	return nil
}
