package ngx

import (
	"bytes"
	"errors"
	"reflect"
	"sync"

	"github.com/modern-go/reflect2"
)

var (
	ErrNilPointer = errors.New("cannot unmarshal into nil pointer")
	ErrNonPointer = errors.New("cannot unmarshal into a non-pointer type data structure")
)

// see http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format for more details.

type NGX struct {
	cache     sync.Map
	ops       []operator
	jescape   bool
	supported map[string]int
}

func (ngx *NGX) MarshalToString(v interface{}) (string, error) {
	return "", nil
}

func (ngx *NGX) Marshal(v interface{}) ([]byte, error) {
	return nil, nil
}

func (ngx *NGX) UnmarshalFromString(data string, itf interface{}) error {
	if len(ngx.ops) <= 0 {
		return nil
	}

	ptr := reflect2.PtrOf(itf)
	if ptr == nil {
		return ErrNilPointer
	}

	rtyp := reflect2.RTypeOf(itf)

	if decoder, _ := ngx.cache.Load(rtyp); decoder != nil {
		return decoder.(Decoder).Decode(ptr, bytes.NewBufferString(data))
	}

	// create decoder

	typ := reflect2.TypeOf(itf)
	if typ.Kind() != reflect.Ptr {
		return ErrNonPointer
	}

	d, err := decoderOf(ngx, typ.(*reflect2.UnsafePtrType).Elem())
	if err != nil {
		return err
	}

	ngx.cache.Store(rtyp, d)

	return d.Decode(ptr, bytes.NewBufferString(data))
}

func (ngx *NGX) Unmarshal(data []byte, itf interface{}) error {
	if len(ngx.ops) <= 0 {
		return nil
	}

	ptr := reflect2.PtrOf(itf)
	if ptr == nil {
		return ErrNilPointer
	}

	rtyp := reflect2.RTypeOf(itf)

	if decoder, _ := ngx.cache.Load(rtyp); decoder != nil {
		return decoder.(Decoder).Decode(ptr, bytes.NewBuffer(data))
	}

	// create decoder

	typ := reflect2.TypeOf(itf)
	if typ.Kind() != reflect.Ptr {
		return ErrNonPointer
	}

	d, err := decoderOf(ngx, typ.(*reflect2.UnsafePtrType).Elem())
	if err != nil {
		return err
	}

	ngx.cache.Store(rtyp, d)

	return d.Decode(ptr, bytes.NewBuffer(data))
}
