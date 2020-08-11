package ngx

import (
	"unsafe"
)

type API interface {
	Marshal(v interface{}) ([]byte, error)
	MarshalToString(v interface{}) (string, error)
	Unmarshal(data []byte, v interface{}) error
	UnmarshalFromString(str string, v interface{}) error
}

type Decoder interface {
	Decode(unsafe.Pointer, Buffer) error
}

type Encoder interface {
	Encoder(unsafe.Pointer, Buffer) error
}
