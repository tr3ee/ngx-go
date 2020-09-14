package ngx

import (
	"unsafe"
)

type API interface {
	Marshal(v interface{}) ([]byte, error)
	MarshalToString(v interface{}) (string, error)
	Unmarshal(data []byte, v interface{}) error
	UnmarshalFromString(str string, v interface{}) error
	Supported() map[string]int
}

type Codec interface {
	Encode(unsafe.Pointer, Writer) error
	Decode(unsafe.Pointer, Reader) error
}
