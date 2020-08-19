package ngx

import (
	"bytes"
	"errors"
	"reflect"
	"sync"
	"unsafe"

	"github.com/modern-go/reflect2"
)

var (
	ErrNilPointer     = errors.New("cannot unmarshal into nil pointer")
	ErrNonPointer     = errors.New("cannot unmarshal into a non-pointer type data structure")
	ErrNotImplemented = errors.New("This feature is not implemented")
)

var (
	CombinedFmt = "$remote_addr - $remote_user [$time_local] \"$request\" $status $body_bytes_sent \"$http_referer\" \"$http_user_agent\""
	ngx, _      = Compile(CombinedFmt)
)

type Access struct {
	RemoteAddr    string `ngx:"remote_addr"`
	RemoteUser    string `ngx:"remote_user"`
	TimeLocal     string `ngx:"time_local"`
	Request       string `ngx:"request"`
	Status        int    `ngx:"status"`
	BytesSent     int    `ngx:"bytes_sent"`
	BodyBytesSent int    `ngx:"body_bytes_sent"`
	HTTPReferer   string `ngx:"http_referer"`
	HTTPUserAgent string `ngx:"http_user_agent"`
	HTTPCookie    string `ngx:"http_cookie"`
	RequestBody   string `ngx:"request_body"`
}

func Marshal(v interface{}) ([]byte, error) {
	return ngx.Marshal(v)
}

func MarshalToString(v interface{}) (string, error) {
	return ngx.MarshalToString(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return ngx.Unmarshal(data, v)
}

func UnmarshalFromString(str string, v interface{}) error {
	return ngx.UnmarshalFromString(str, v)
}

// see http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format for more details.

type NGX struct {
	cache     sync.Map
	ops       []baseOp
	esc       Esc
	supported map[string]int
}

func (ngx *NGX) MarshalToString(itf interface{}) (string, error) {
	if len(ngx.ops) <= 0 {
		return "", nil
	}

	ptr := reflect2.PtrOf(itf)

	rtyp := reflect2.RTypeOf(itf)

	buf := bytes.NewBuffer(nil)

	if codec, _ := ngx.cache.Load(rtyp); codec != nil {
		if err := codec.(Codec).Encode(ptr, buf); err != nil {
			return "", err
		}
		bytes := buf.Bytes()
		return *(*string)(unsafe.Pointer(&bytes)), nil
	}

	// create codec

	typ := reflect2.TypeOf(itf)

	d, err := codecOf(ngx, typ)
	if err != nil {
		return "", err
	}

	if typ.LikePtr() {
		d = &RefCodec{d}
	}

	ngx.cache.Store(rtyp, d)

	if err := d.Encode(ptr, buf); err != nil {
		return "", err
	}
	bytes := buf.Bytes()
	return *(*string)(unsafe.Pointer(&bytes)), nil
}

func (ngx *NGX) Marshal(itf interface{}) ([]byte, error) {
	if len(ngx.ops) <= 0 {
		return nil, nil
	}

	ptr := reflect2.PtrOf(itf)

	rtyp := reflect2.RTypeOf(itf)

	buf := bytes.NewBuffer(nil)

	if codec, _ := ngx.cache.Load(rtyp); codec != nil {
		if err := codec.(Codec).Encode(ptr, buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	// create codec

	d, err := codecOf(ngx, reflect2.TypeOf(itf))
	if err != nil {
		return nil, err
	}

	ngx.cache.Store(rtyp, d)

	if err := d.Encode(ptr, buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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

	if codec, _ := ngx.cache.Load(rtyp); codec != nil {
		return codec.(Codec).Decode(ptr, NewStringBuffer(data))
	}

	// create codec

	typ := reflect2.TypeOf(itf)
	if typ.Kind() != reflect.Ptr {
		return ErrNonPointer
	}

	d, err := codecOf(ngx, typ.(*reflect2.UnsafePtrType).Elem())
	if err != nil {
		return err
	}

	ngx.cache.Store(rtyp, d)

	return d.Decode(ptr, NewStringBuffer(data))
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

	if codec, _ := ngx.cache.Load(rtyp); codec != nil {
		return codec.(Codec).Decode(ptr, NewBytesBuffer(data))
	}

	// create codec

	typ := reflect2.TypeOf(itf)
	if typ.Kind() != reflect.Ptr {
		return ErrNonPointer
	}

	d, err := codecOf(ngx, typ.(*reflect2.UnsafePtrType).Elem())
	if err != nil {
		return err
	}

	ngx.cache.Store(rtyp, d)

	return d.Decode(ptr, NewBytesBuffer(data))
}
