package ngx

import (
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
	case reflect.Map:
		return decoderOfMap(ngx, typ.(*reflect2.UnsafeMapType))
	case reflect.Struct:
		return decoderOfStruct(ngx, typ.(*reflect2.UnsafeStructType))
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

func (d *ByteDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
	if text.Len() != 1 {
		return fmt.Errorf("expected byte, got %q", text.String())
	}
	*(*byte)(ptr) = byte(text.Bytes()[0])
	return nil
}

type Int8Decoder struct {
}

func (d *Int8Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *Int16Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *Uint16Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *Int32Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *Uint32Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *Int64Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *Uint64Decoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *IntDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *UintDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
	v, err := strconv.ParseUint(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*uint)(ptr) = uint(v)
	return nil
}

type BoolDecoder struct {
}

func (d *BoolDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
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

func (d *PtrDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
	if *((*unsafe.Pointer)(ptr)) == nil {
		*((*unsafe.Pointer)(ptr)) = d.typ.UnsafeNew()
	}
	return d.dec.Decode(*((*unsafe.Pointer)(ptr)), text)
}

type BytesDecoder struct {
}

func (d *BytesDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
	b := text.NewBytes()
	*(*reflect.SliceHeader)(ptr) = *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	return nil
}

type StringDecoder struct {
}

func (d *StringDecoder) Decode(ptr unsafe.Pointer, text Buffer) error {
	*((*string)(ptr)) = text.String()
	return nil
}
