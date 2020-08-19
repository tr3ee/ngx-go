package ngx

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/modern-go/reflect2"
)

func codecOf(ngx *NGX, typ reflect2.Type) (Codec, error) {
	switch typ.Kind() {
	case reflect.Bool:
		return &boolCodec{}, nil
	case reflect.Int:
		return &intCodec{}, nil
	case reflect.Uint:
		return &uintCodec{}, nil
	case reflect.Int8:
		return &int8Codec{}, nil
	case reflect.Uint8:
		return &byteCodec{}, nil
	case reflect.Int16:
		return &int16Codec{}, nil
	case reflect.Uint16:
		return &uint16Codec{}, nil
	case reflect.Int32:
		return &int32Codec{}, nil
	case reflect.Uint32:
		return &uint32Codec{}, nil
	case reflect.Int64:
		return &int64Codec{}, nil
	case reflect.Uint64:
		return &uint64Codec{}, nil
	case reflect.Slice:
		if typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
			return &bytesCodec{ngx.esc}, nil
		}
		return nil, ErrNotImplemented
	case reflect.String:
		return &stringCodec{ngx.esc}, nil
	case reflect.Map:
		return codecOfMap(ngx, typ.(*reflect2.UnsafeMapType))
	case reflect.Struct:
		return codecOfStruct(ngx, typ.(*reflect2.UnsafeStructType))
	case reflect.Ptr:
		elem := typ.(*reflect2.UnsafePtrType).Elem()
		codec, err := codecOf(ngx, elem)
		if err != nil {
			return nil, err
		}
		return &ptrCodec{ngx.esc.Nil(), codec, elem}, nil
	default:
		return nil, fmt.Errorf("Unsupported decoding type %q", typ.Kind().String())
	}
}

type byteCodec struct {
}

func (d *byteCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*uint8)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *byteCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	if text.Len() != 1 {
		return fmt.Errorf("expected byte, got %q", text.String())
	}
	*(*byte)(ptr) = byte(text.Bytes()[0])
	return nil
}

type int8Codec struct {
}

func (d *int8Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*int8)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *int8Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type int16Codec struct {
}

func (d *int16Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*int16)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *int16Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type uint16Codec struct {
}

func (d *uint16Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*uint16)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *uint16Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type int32Codec struct {
}

func (d *int32Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*int32)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *int32Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type uint32Codec struct {
}

func (d *uint32Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*uint32)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *uint32Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type int64Codec struct {
}

func (d *int64Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*int64)(ptr)
	text.WriteString(strconv.FormatInt(v, 10))
	return nil
}

func (d *int64Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type uint64Codec struct {
}

func (d *uint64Codec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*uint64)(ptr)
	text.WriteString(strconv.FormatUint(v, 10))
	return nil
}

func (d *uint64Codec) Decode(ptr unsafe.Pointer, text Reader) error {
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

type intCodec struct {
}

func (d *intCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*int)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *intCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	v, err := strconv.ParseInt(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*int)(ptr) = int(v)
	return nil
}

type uintCodec struct {
}

func (d *uintCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	v := *(*uint)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *uintCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	v, err := strconv.ParseUint(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*uint)(ptr) = uint(v)
	return nil
}

type boolCodec struct {
}

func (d *boolCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	if *(*bool)(ptr) {
		text.WriteString("true")
	} else {
		text.WriteString("false")
	}
	return nil
}

func (d *boolCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	if strings.ToLower(text.String()) == "true" {
		*(*bool)(ptr) = true
	} else {
		*(*bool)(ptr) = false
	}
	return nil
}

type ptrCodec struct {
	esc   string
	codec Codec
	typ   reflect2.Type
}

func (d *ptrCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	if *((*unsafe.Pointer)(ptr)) == nil {
		text.WriteString(d.esc)
		return nil
	}
	return d.codec.Encode(*(*unsafe.Pointer)(ptr), text)
}

func (d *ptrCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	if *((*unsafe.Pointer)(ptr)) == nil {
		*((*unsafe.Pointer)(ptr)) = d.typ.UnsafeNew()
	}
	return d.codec.Decode(*((*unsafe.Pointer)(ptr)), text)
}

type refCodec struct {
	codec Codec
}

func (d *refCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	return d.codec.Encode(unsafe.Pointer(&ptr), text)
}

func (d *refCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	return d.codec.Decode(ptr, text)
}

type bytesCodec struct {
	esc Esc
}

func (d *bytesCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	if ptr == nil {
		return nil
	}
	_, err := text.Write(d.esc.Escape(*(*[]byte)(unsafe.Pointer(&ptr))))
	return err
}

func (d *bytesCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	b := text.NewBytes()
	*(*reflect.SliceHeader)(ptr) = *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	return nil
}

type stringCodec struct {
	esc Esc
}

func (d *stringCodec) Encode(ptr unsafe.Pointer, text Writer) error {
	buf := NewStringReader(*(*string)(ptr))
	_, err := text.Write(d.esc.Escape(buf.Bytes()))
	return err
}

func (d *stringCodec) Decode(ptr unsafe.Pointer, text Reader) error {
	*((*string)(ptr)) = text.String()
	return nil
}
