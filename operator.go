package ngx

const (
	ngxUnknown = iota
	ngxString
	ngxVariable
	ngxBind
)

type operator struct {
	Type       int
	Index      int
	Offset     uintptr
	Extra      string
	ExtraBytes []byte
	Dec        Decoder
}
