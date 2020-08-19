package ngx

import (
	"bytes"
	"unsafe"
)

type API interface {
	Marshal(v interface{}) ([]byte, error)
	MarshalToString(v interface{}) (string, error)
	Unmarshal(data []byte, v interface{}) error
	UnmarshalFromString(str string, v interface{}) error
}

type Codec interface {
	Encode(unsafe.Pointer, *bytes.Buffer) error
	Decode(unsafe.Pointer, Buffer) error
}
