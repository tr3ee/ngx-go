package ngx

import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/modern-go/reflect2"
)

func codecOfMap(ngx *NGX, typ *reflect2.UnsafeMapType) (Codec, error) {
	keyCodec, err := codecOf(ngx, typ.Key())
	if err != nil {
		return nil, err
	}

	elemCodec, err := codecOf(ngx, typ.Elem())
	if err != nil {
		return nil, err
	}

	return &MapCodec{
		ops:       ngx.ops,
		esc:       ngx.esc,
		mapType:   typ,
		keyType:   typ.Key(),
		elemType:  typ.Elem(),
		keyCodec:  keyCodec,
		elemCodec: elemCodec,
	}, nil
}

type MapCodec struct {
	ops []baseOp
	esc Esc

	mapType   *reflect2.UnsafeMapType
	keyType   reflect2.Type
	elemType  reflect2.Type
	keyCodec  Codec
	elemCodec Codec
}

func (d *MapCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
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
		case ngxBind, ngxVariable:

			key := d.keyType.UnsafeNew()
			if err := d.keyCodec.Decode(key, NewBytesBuffer(op.Extra)); err != nil {
				return err
			}

			val := d.mapType.UnsafeGetIndex(ptr, key)

			if err := d.elemCodec.Encode(val, text); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *MapCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
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
		case ngxBind, ngxVariable:
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

			key := d.keyType.UnsafeNew()
			if err := d.keyCodec.Decode(key, NewBytesBuffer(op.Extra)); err != nil {
				return err
			}

			text := Buffer(NewBytesBuffer(raw))
			if raw, err := d.esc.Unescape(text); err != nil {
				return err
			} else {
				text = raw
			}
			elem := d.elemType.UnsafeNew()
			if err := d.elemCodec.Decode(elem, text); err != nil {
				return err
			}

			d.mapType.UnsafeSetIndex(ptr, key, elem)

		default:
			return fmt.Errorf("Unsupported operator type(%d)", op.Type)
		}
	}

	return nil
}
