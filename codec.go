package ngx

import (
	"bytes"
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
		return &BoolCodec{}, nil
	case reflect.Int:
		return &IntCodec{}, nil
	case reflect.Uint:
		return &UintCodec{}, nil
	case reflect.Int8:
		return &Int8Codec{}, nil
	case reflect.Uint8:
		return &ByteCodec{}, nil
	case reflect.Int16:
		return &Int16Codec{}, nil
	case reflect.Uint16:
		return &Uint16Codec{}, nil
	case reflect.Int32:
		return &Int32Codec{}, nil
	case reflect.Uint32:
		return &Uint32Codec{}, nil
	case reflect.Int64:
		return &Int64Codec{}, nil
	case reflect.Uint64:
		return &Uint64Codec{}, nil
	case reflect.Slice:
		if typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
			return &BytesCodec{ngx.esc}, nil
		}
		return nil, ErrNotImplemented
	case reflect.String:
		return &StringCodec{ngx.esc}, nil
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
		return &PtrCodec{ngx.esc.Nil(), codec, elem}, nil
	default:
		return nil, fmt.Errorf("Unsupported decoding type %q", typ.Kind().String())
	}
}

type ByteCodec struct {
}

func (d *ByteCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*uint8)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *ByteCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	if text.Len() != 1 {
		return fmt.Errorf("expected byte, got %q", text.String())
	}
	*(*byte)(ptr) = byte(text.Bytes()[0])
	return nil
}

type Int8Codec struct {
}

func (d *Int8Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*int8)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *Int8Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type Int16Codec struct {
}

func (d *Int16Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*int16)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *Int16Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type Uint16Codec struct {
}

func (d *Uint16Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*uint16)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *Uint16Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type Int32Codec struct {
}

func (d *Int32Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*int32)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *Int32Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type Uint32Codec struct {
}

func (d *Uint32Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*uint32)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *Uint32Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type Int64Codec struct {
}

func (d *Int64Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*int64)(ptr)
	text.WriteString(strconv.FormatInt(v, 10))
	return nil
}

func (d *Int64Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type Uint64Codec struct {
}

func (d *Uint64Codec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*uint64)(ptr)
	text.WriteString(strconv.FormatUint(v, 10))
	return nil
}

func (d *Uint64Codec) Decode(ptr unsafe.Pointer, text Buffer) error {
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

type IntCodec struct {
}

func (d *IntCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*int)(ptr)
	text.WriteString(strconv.FormatInt(int64(v), 10))
	return nil
}

func (d *IntCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	v, err := strconv.ParseInt(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*int)(ptr) = int(v)
	return nil
}

type UintCodec struct {
}

func (d *UintCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	v := *(*uint)(ptr)
	text.WriteString(strconv.FormatUint(uint64(v), 10))
	return nil
}

func (d *UintCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	v, err := strconv.ParseUint(text.String(), 10, 0)
	if err != nil {
		return err
	}
	*(*uint)(ptr) = uint(v)
	return nil
}

type BoolCodec struct {
}

func (d *BoolCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	if *(*bool)(ptr) {
		text.WriteString("true")
	} else {
		text.WriteString("false")
	}
	return nil
}

func (d *BoolCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	if strings.ToLower(text.String()) == "true" {
		*(*bool)(ptr) = true
	} else {
		*(*bool)(ptr) = false
	}
	return nil
}

type PtrCodec struct {
	esc   string
	codec Codec
	typ   reflect2.Type
}

func (d *PtrCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	if *((*unsafe.Pointer)(ptr)) == nil {
		text.WriteString(d.esc)
		return nil
	}
	return d.codec.Encode(*(*unsafe.Pointer)(ptr), text)
}

func (d *PtrCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	if *((*unsafe.Pointer)(ptr)) == nil {
		*((*unsafe.Pointer)(ptr)) = d.typ.UnsafeNew()
	}
	return d.codec.Decode(*((*unsafe.Pointer)(ptr)), text)
}

type RefCodec struct {
	codec Codec
}

func (d *RefCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	return d.codec.Encode(unsafe.Pointer(&ptr), text)
}

func (d *RefCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	return d.codec.Decode(unsafe.Pointer(&ptr), text)
}

type BytesCodec struct {
	esc Esc
}

func (d *BytesCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	buf := NewBytesBuffer(*(*[]byte)(unsafe.Pointer(&ptr)))
	_, err := text.Write(d.esc.Escape(buf).Bytes())
	return err
}

func (d *BytesCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	b := text.NewBytes()
	*(*reflect.SliceHeader)(ptr) = *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	return nil
}

type StringCodec struct {
	esc Esc
}

func (d *StringCodec) Encode(ptr unsafe.Pointer, text *bytes.Buffer) error {
	buf := NewStringBuffer(*(*string)(ptr))
	_, err := text.Write(d.esc.Escape(buf).Bytes())
	return err
}

func (d *StringCodec) Decode(ptr unsafe.Pointer, text Buffer) error {
	*((*string)(ptr)) = text.String()
	return nil
}
