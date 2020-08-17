package ngx

import (
	"bytes"
	"fmt"
	"unsafe"

	"github.com/modern-go/reflect2"
)

func decoderOfMap(ngx *NGX, typ *reflect2.UnsafeMapType) (Decoder, error) {
	keyDecoder, err := decoderOf(ngx, typ.Key())
	if err != nil {
		return nil, err
	}

	elemDecoder, err := decoderOf(ngx, typ.Elem())
	if err != nil {
		return nil, err
	}

	return &MapDecoder{
		ops:         ngx.ops,
		esc:         ngx.esc,
		mapType:     typ,
		keyType:     typ.Key(),
		elemType:    typ.Elem(),
		keyDecoder:  keyDecoder,
		elemDecoder: elemDecoder,
	}, nil
}

type MapDecoder struct {
	ops []baseOp
	esc Esc

	mapType     *reflect2.UnsafeMapType
	keyType     reflect2.Type
	elemType    reflect2.Type
	keyDecoder  Decoder
	elemDecoder Decoder
}

func (d *MapDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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
			if err := d.keyDecoder.Decode(key, NewBytesBuffer(op.Extra)); err != nil {
				return err
			}

			text := Buffer(NewBytesBuffer(raw))
			if raw, err := d.esc.Unescape(text); err != nil {
				return err
			} else {
				text = raw
			}
			elem := d.elemType.UnsafeNew()
			d.elemDecoder.Decode(elem, text)

			d.mapType.UnsafeSetIndex(ptr, key, elem)

		default:
			return fmt.Errorf("Unsupported operator type(%d)", op.Type)
		}
	}

	return nil
}
