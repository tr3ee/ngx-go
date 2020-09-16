package ngx

import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/modern-go/reflect2"
)

type mapOp struct {
	baseOp
	KeyV unsafe.Pointer
}

func codecOfMap(ngx *NGX, typ *reflect2.UnsafeMapType) (Codec, error) {
	keyCodec, err := codecOf(ngx, typ.Key())
	if err != nil {
		return nil, err
	}

	elemCodec, err := codecOf(ngx, typ.Elem())
	if err != nil {
		return nil, err
	}

	ops := make([]mapOp, len(ngx.ops))
	for i := 0; i < len(ngx.ops); i++ {
		ops[i].baseOp = ngx.ops[i]
		if ops[i].Type == ngxVariable {
			if string(ops[i].Extra) == "_" {
				continue
			}
			ops[i].Type = ngxBind
		}
		ops[i].KeyV = typ.Key().UnsafeNew()
		if err := keyCodec.Decode(ops[i].KeyV, NewBytesReader(ops[i].Extra)); err != nil {
			return nil, err
		}
	}

	return &mapCodec{
		ops:       ops,
		esc:       ngx.esc,
		mapType:   typ,
		keyType:   typ.Key(),
		elemType:  typ.Elem(),
		keyCodec:  keyCodec,
		elemCodec: elemCodec,
	}, nil
}

type mapCodec struct {
	ops []mapOp
	esc Esc

	mapType   *reflect2.UnsafeMapType
	keyType   reflect2.Type
	elemType  reflect2.Type
	keyCodec  Codec
	elemCodec Codec
}

func (d *mapCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	if *(*unsafe.Pointer)(ptr) == nil {
		text.WriteString(d.esc.Nil())
		return nil
	}
	length := len(d.ops)
	for i := 0; i < length; i++ {
		op := d.ops[i]
		switch op.Type {
		case ngxString, ngxEscString:
			text.Write(op.Extra)
		case ngxVariable:
			// skip
		case ngxBind:
			val := d.mapType.UnsafeGetIndex(ptr, op.KeyV)
			if err := d.elemCodec.Encode(val, text); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *mapCodec) Decode(ptr unsafe.Pointer, text Reader) error {
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

			raw, err := d.esc.Unescape(raw)
			if err != nil {
				return err
			}

			elem := d.elemType.UnsafeNew()
			if err := d.elemCodec.Decode(elem, NewBytesReader(raw)); err != nil {
				return err
			}

			d.mapType.UnsafeSetIndex(ptr, op.KeyV, elem)

		default:
			return fmt.Errorf("Unsupported operator type(%d)", op.Type)
		}
	}

	return nil
}
