package ngx

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/modern-go/reflect2"
)

var (
	ErrNotImplemented = errors.New("This feature is not implemented")
)

func decoderOf(ngx *NGX, typ reflect2.Type) (Decoder, error) {
	switch typ.Kind() {
	case reflect.Bool:
		return &BoolDecoder{}, nil
	case reflect.Int:
		return &IntDecoder{}, nil
	case reflect.Uint:
		return &UintDecoder{}, nil
	case reflect.Int8:
		return &Int8Decoder{}, nil
	case reflect.Uint8:
		return &ByteDecoder{}, nil
	case reflect.Int16:
		return &Int16Decoder{}, nil
	case reflect.Uint16:
		return &Uint16Decoder{}, nil
	case reflect.Int32:
		return &Int32Decoder{}, nil
	case reflect.Uint32:
		return &Uint32Decoder{}, nil
	case reflect.Int64:
		return &Int64Decoder{}, nil
	case reflect.Uint64:
		return &Uint64Decoder{}, nil
	case reflect.Slice:
		if typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
			return &BytesDecoder{}, nil
		}
		return nil, ErrNotImplemented
	case reflect.String:
		return &StringDecoder{}, nil
	case reflect.Struct:
		ops := make([]operator, len(ngx.ops))
		copy(ops, ngx.ops)
		elemType := typ.(*reflect2.UnsafeStructType)
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			name := field.Name()
			tag := field.Tag().Get("ngx")
			if name == "_" || tag == "_" {
				continue
			}

			if len(tag) > 0 {
				name = tag
			}
			if ind, ok := ngx.supported[name]; ok {
				ops[ind].Index = i
				ops[ind].Type = ngxBind
				ops[ind].Offset = field.Offset()
				dec, err := decoderOf(ngx, field.Type())
				if err != nil {
					return nil, err
				}
				ops[ind].Dec = dec
			}
		}
		return &StructDecoder{ops, ngx.jescape}, nil
	case reflect.Map:
		mapType := typ.(*reflect2.UnsafeMapType)
		keyDecoder, err := decoderOf(ngx, mapType.Key())
		if err != nil {
			return nil, err
		}
		elemDecoder, err := decoderOf(ngx, mapType.Elem())
		if err != nil {
			return nil, err
		}
		return &MapDecoder{
			ops:         ngx.ops,
			jescape:     ngx.jescape,
			mapType:     mapType,
			keyType:     mapType.Key(),
			elemType:    mapType.Elem(),
			keyDecoder:  keyDecoder,
			elemDecoder: elemDecoder,
		}, nil
	case reflect.Ptr:
		elem := typ.(*reflect2.UnsafePtrType).Elem()
		decoder, err := decoderOf(ngx, elem)
		if err != nil {
			return nil, err
		}
		return &PtrDecoder{decoder, elem}, nil
	default:
		return nil, fmt.Errorf("Unsupported decoding type %q", typ.Kind().String())
	}
}

type ByteDecoder struct {
}

func (d *ByteDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	if text.Len() != 1 {
		return fmt.Errorf("expected byte, got %q", text.String())
	}
	*(*byte)(ptr) = byte(text.Bytes()[0])
	return nil
}

type Int8Decoder struct {
}

func (d *Int8Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseInt(text.String(), 10, 8)
	if err != nil {
		return fmt.Errorf("expected int8, got %q", text.String())
	}
	if v > math.MaxInt8 {
		return fmt.Errorf("%v overflows int8", v)
	}
	*(*int8)(ptr) = int8(v)
	return nil
}

type Int16Decoder struct {
}

func (d *Int16Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseInt(text.String(), 10, 16)
	if err != nil {
		return err
	}
	if v > math.MaxInt16 {
		return fmt.Errorf("%v overflows int16", v)
	}
	*(*int16)(ptr) = int16(v)
	return nil
}

type Uint16Decoder struct {
}

func (d *Uint16Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseUint(text.String(), 10, 16)
	if err != nil {
		return err
	}
	if v > math.MaxUint16 {
		return fmt.Errorf("%v overflows uint16", v)
	}
	*(*uint16)(ptr) = uint16(v)
	return nil
}

type Int32Decoder struct {
}

func (d *Int32Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseInt(text.String(), 10, 32)
	if err != nil {
		return err
	}
	if v > math.MaxInt32 {
		return fmt.Errorf("%v overflows int32", v)
	}
	*(*int32)(ptr) = int32(v)
	return nil
}

type Uint32Decoder struct {
}

func (d *Uint32Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseUint(text.String(), 10, 32)
	if err != nil {
		return err
	}
	if v > math.MaxUint32 {
		return fmt.Errorf("%v overflows uint32", v)
	}
	*(*uint32)(ptr) = uint32(v)
	return nil
}

type Int64Decoder struct {
}

func (d *Int64Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseInt(text.String(), 10, 64)
	if err != nil {
		return err
	}
	if v > math.MaxInt64 {
		return fmt.Errorf("%v overflows int64", v)
	}
	*(*int64)(ptr) = int64(v)
	return nil
}

type Uint64Decoder struct {
}

func (d *Uint64Decoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseUint(text.String(), 10, 64)
	if err != nil {
		return err
	}
	if v > math.MaxUint64 {
		return fmt.Errorf("%v overflows uint64", v)
	}
	*(*uint64)(ptr) = uint64(v)
	return nil
}

type IntDecoder struct {
}

func (d *IntDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseInt(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*int)(ptr) = int(v)
	return nil
}

type UintDecoder struct {
	Name string
}

func (d *UintDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v, err := strconv.ParseUint(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*uint)(ptr) = uint(v)
	return nil
}

type BoolDecoder struct {
}

func (d *BoolDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	if strings.ToLower(text.String()) == "true" {
		*(*bool)(ptr) = true
	} else {
		*(*bool)(ptr) = false
	}
	return nil
}

type PtrDecoder struct {
	dec Decoder
	typ reflect2.Type
}

func (d *PtrDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	if *((*unsafe.Pointer)(ptr)) == nil {
		*((*unsafe.Pointer)(ptr)) = d.typ.UnsafeNew()
	}
	return d.dec.Decode(*((*unsafe.Pointer)(ptr)), text)
}

type BytesDecoder struct {
}

func (d *BytesDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	b := text.Bytes()
	*(*reflect.SliceHeader)(ptr) = *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	return nil
}

type StringDecoder struct {
}

func (d *StringDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	*((*string)(ptr)) = text.String()
	return nil
}

type MapDecoder struct {
	ops     []operator
	jescape bool

	mapType     *reflect2.UnsafeMapType
	keyType     reflect2.Type
	elemType    reflect2.Type
	keyDecoder  Decoder
	elemDecoder Decoder
}

func (d *MapDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	p := 0
	data := text.Bytes()
	length := len(d.ops)
	for i := 0; i < length; i++ {
		op := d.ops[i]
		switch op.Type {
		case ngxString, ngxEscString:
			p += len(op.Extra)
		case ngxBind, ngxVariable:
			var raw []byte
			if i+1 >= length {
				raw = data[p:]
			} else {
				next := d.ops[i+1]
				switch next.Type {
				case ngxString:
					off := bytes.Index(data[p:], next.ExtraBytes)
					if off < 0 {
						return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
					}
					raw = data[p : p+off]
					i++
					p += off + len(next.ExtraBytes)
				case ngxEscString:
					oldp := p
				ngx_bind_retry:
					off := bytes.Index(data[p:], next.ExtraBytes)
					if off < 0 {
						return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
					} else if off > 0 && data[p+off-1] == '\\' {
						p += off + len(next.ExtraBytes)
						goto ngx_bind_retry
					}
					raw = data[oldp : p+off]
					i++
					p += off + len(next.ExtraBytes)
				default:
					return fmt.Errorf("ngx-go does not support '$%s$%s' style format", op.Extra, next.Extra)
				}
			}

			key := d.keyType.UnsafeNew()
			if err := d.keyDecoder.Decode(key, bytes.NewBuffer([]byte(op.Extra))); err != nil {
				return err
			}

			text := bytes.NewBuffer(raw)
			if d.jescape {
				raw, err := junescape(text)
				if err != nil {
					return err
				}
				text = raw
			} else {
				raw, err := unescape(text)
				if err != nil {
					return err
				}
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

type StructDecoder struct {
	ops     []operator
	jescape bool
}

func (d *StructDecoder) Decode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	p := 0
	data := text.Bytes()
	length := len(d.ops)
	for i := 0; i < length; i++ {
		op := d.ops[i]
		switch op.Type {
		case ngxString, ngxEscString:
			p += len(op.Extra)
		case ngxVariable:
			if i+1 >= length {
				return nil
			}
			next := d.ops[i+1]
			switch next.Type {
			case ngxString:
				off := bytes.Index(data[p:], next.ExtraBytes)
				if off < 0 {
					return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
				}
				i++
				p += off + len(next.ExtraBytes)
			case ngxEscString:
			ngx_var_retry:
				off := bytes.Index(data[p:], next.ExtraBytes)
				if off < 0 {
					return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
				} else if off > 0 && data[p+off-1] == '\\' {
					p += off + len(next.ExtraBytes)
					goto ngx_var_retry
				}
				i++
				p += off + len(next.ExtraBytes)
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
					off := bytes.Index(data[p:], next.ExtraBytes)
					if off < 0 {
						return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
					}
					raw = data[p : p+off]
					i++
					p += off + len(next.ExtraBytes)
				case ngxEscString:
					oldp := p
				ngx_bind_retry:
					off := bytes.Index(data[p:], next.ExtraBytes)
					if off < 0 {
						return fmt.Errorf("got unexpected EOF: expecting %q after $%s", next.Extra, op.Extra)
					} else if off > 0 && data[p+off-1] == '\\' {
						p += off + len(next.ExtraBytes)
						goto ngx_bind_retry
					}
					raw = data[oldp : p+off]
					i++
					p += off + len(next.ExtraBytes)
				default:
					return fmt.Errorf("ngx-go does not support '$%s$%s' style format", op.Extra, next.Extra)
				}
			}

			text := bytes.NewBuffer(raw)
			if d.jescape {
				raw, err := junescape(text)
				if err != nil {
					return err
				}
				text = raw
			} else {
				raw, err := unescape(text)
				if err != nil {
					return err
				}
				text = raw
			}

			bindPtr := unsafe.Pointer(uintptr(ptr) + op.Offset)

			if err := op.Dec.Decode(bindPtr, text); err != nil {
				return fmt.Errorf("field %q %v", op.Extra, err)
			}

		default:
			return fmt.Errorf("Unsupported operator type(%d)", op.Type)
		}
	}

	return nil
}
