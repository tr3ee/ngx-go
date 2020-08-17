package ngx

import (
	"reflect"
	"testing"
)

var positiveUnmarshal = []struct {
	Fmt      string
	Data     string
	Expected map[string]string
}{
	{CombinedFmt, CombinedFmt, map[string]string{"remote_addr": "$remote_addr", "remote_user": "$remote_user", "time_local": "$time_local", "request": "$request", "status": "$status", "body_bytes_sent": "$body_bytes_sent", "http_referer": "$http_referer", "http_user_agent": "$http_user_agent"}},
	{`\$request\$request_body\$header_cookie\`, `\request\request_body\header_cookie\`, map[string]string{"request": "request", "request_body": "request_body", "header_cookie": "header_cookie"}},
	{`\$request\"$request_body\"\"$header_cookie\"`, `\request\"request_body\"\"header_cookie\"`, map[string]string{"request": "request", "request_body": "request_body", "header_cookie": "header_cookie"}},
	{`\$request\"$request_body\"\"$header_cookie\"`, `\requ\\\"est\"request_body\"\"header_cookie\"`, map[string]string{"request": "requ\\\"est", "request_body": "request_body", "header_cookie": "header_cookie"}},
	{`escape=json;{"$key":"$value"}`, `{"$key":"$value"}`, map[string]string{"key": "$key", "value": "$value"}},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065y":"\r\f\t\uf755\n"}`, map[string]string{"key": "$key", "value": "\r\f\t\xef\x9d\x95\n"}},
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range positiveUnmarshal {
		ngx, err := Compile(tc.Fmt)
		if err != nil {
			t.Fatalf("failed to compile format %q: %v", tc.Fmt, err)
		}

		got := make(map[string]string)
		if err := ngx.UnmarshalFromString(tc.Data, &got); err != nil {
			t.Fatalf("failed to unmarshal data %q: %v", tc.Data, err)
		}

		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in unmarshal: expecting %q, got %q", tc.Expected, got)
		}
	}
}

func BenchmarkUnmarshalFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[string]string)
		if err := UnmarshalFromString(CombinedFmt, &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(CombinedFmt)
	for i := 0; i < b.N; i++ {
		m := make(map[string]string)
		if err := Unmarshal(data, &m); err != nil {
			b.Fatal(err)
		}
	}
}
