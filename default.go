package ngx

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
