package ngx

import (
	"bytes"
	"reflect"
	"testing"
)

var positiveStruct = []struct {
	Fmt       string
	Data      string
	Expected  Access
	Marshaled string
}{
	{CombinedFmt, "$remote_addr - $remote_user [$time_local] \"$request\" 200 0 \"$http_referer\" \"$http_user_agent\"", Access{RemoteAddr: "$remote_addr", RemoteUser: "$remote_user", TimeLocal: "$time_local", Request: "$request", Status: 200, BodyBytesSent: 0, HTTPReferer: "$http_referer", HTTPUserAgent: "$http_user_agent"}, "$remote_addr - $remote_user [$time_local] \"$request\" 200 0 \"$http_referer\" \"$http_user_agent\""},
	{`escape=json;{"$request":"$request_body"}`, `{"$request\\":"$request_body\""}`, Access{Request: "$request\\", RequestBody: "$request_body\""}, `{"$request\\":"$request_body\""}`},
	{`escape=json;{"$request":"$request_body"}`, `{"$request\\\"":"$request_body\"\\"}`, Access{Request: "$request\\\"", RequestBody: "$request_body\"\\"}, `{"$request\\\"":"$request_body\"\\"}`},
}

var positiveMap = []struct {
	Fmt       string
	Data      string
	Expected  map[string]string
	Marshaled string
}{
	{CombinedFmt, CombinedFmt, map[string]string{"remote_addr": "${remote_addr}", "remote_user": "${remote_user}", "time_local": "$time_local", "request": "${request}", "status": "${status}", "body_bytes_sent": "${body_bytes_sent}", "http_referer": "${http_referer}", "http_user_agent": "${http_user_agent}"}, CombinedFmt},
	{`\$request\$request_body\$header_cookie\`, `\request\request_body\header_cookie\`, map[string]string{"request": "request", "request_body": "request_body", "header_cookie": "header_cookie"}, `\request\request_body\header_cookie\`},
	{`\$request\"$request_body\"\"$header_cookie\"`, `\request\"request_body\"\"header_cookie\"`, map[string]string{"request": "request", "request_body": "request_body", "header_cookie": "header_cookie"}, `\request\"request_body\"\"header_cookie\"`},
	{`\$request\"$request_body\"\"$header_cookie\"`, `\requ\\\"est\"request_body\"\"header_cookie\"`, map[string]string{"request": "requ\\\"est", "request_body": "request_body", "header_cookie": "header_cookie"}, `\requ\\\"est\"request_body\"\"header_cookie\"`},
	{`\$request\"${request_body}a\"\"$header_cookie\"`, `\requ\\\"est\"request_bodya\"\"header_cookie\"`, map[string]string{"request": "requ\\\"est", "request_body": "request_body", "header_cookie": "header_cookie"}, `\requ\\\"est\"request_bodya\"\"header_cookie\"`},
	{`escape=json;{"$key":"$value"}`, `{"$key":"$value"}`, map[string]string{"key": "$key", "value": "$value"}, `{"$key":"$value"}`},
	{`escape=json;{"$key":"$_"}`, `{"$key":"$value"}`, map[string]string{"key": "$key"}, `{"$key":""}`},
	{`escape=json;{"$key":$_"$value"}$_`, `{"$key":    "$value"}`, map[string]string{"key": "$key", "value": "$value"}, `{"$key":"$value"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065y":"\r\f\t\uf755\n"}`, map[string]string{"key": "$key", "value": "\r\f\t\xef\x9d\x95\n"}, "{\"$key\":\"\\r\\f\\t\uf755\\n\"}"},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"\ud83c\udf09"}`, map[string]string{"key": "$key", "value": "🌉"}, `{"$key":"🌉"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"surrogate pair : \ud83c\udf09"}`, map[string]string{"key": "$key", "value": "surrogate pair : 🌉"}, `{"$key":"surrogate pair : 🌉"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"\ud83c\udf09\ud83c\udf09is\u0020surrogate\u0020pair"}`, map[string]string{"key": "$key", "value": "🌉🌉is surrogate pair"}, `{"$key":"🌉🌉is surrogate pair"}`},
	{`escape=json;{"$key":"$value"}`, `{"\u0024k\u0065\u0079":"\ud83c\udf09\ud83c\udf09\ud83c\udf09\ud83c\udf09\""}`, map[string]string{"key": "$key", "value": "🌉🌉🌉🌉\""}, `{"$key":"🌉🌉🌉🌉\""}`},
	{`escape=json;{"$$$key":"$$$value"}`, `{"$key":"$value"}`, map[string]string{"key": "key", "value": "value"}, `{"$key":"$value"}`},
	{`escape=json;{"$$${key}":"$$${value}"}`, `{"$key":"$value"}`, map[string]string{"key": "key", "value": "value"}, `{"$key":"$value"}`},
	{`$$key=$key, $$value=$value`, `$key=hello, $value=world`, map[string]string{"key": "hello", "value": "world"}, `$key=hello, $value=world`},
	{`$$$$key=$key, $$value=$value`, `$$key=hello, $value=world`, map[string]string{"key": "hello", "value": "world"}, `$$key=hello, $value=world`},
	{`$$ $$$$key=$key, $$value=$value`, `$ $$key=hello, $value=world`, map[string]string{"key": "hello", "value": "world"}, `$ $$key=hello, $value=world`},
	{`$$ $$$$key=$key, $$value=$value`, `$ $$key=\x68\x65\x6c\x6c\x6f, $value=\x77\x6f\x72\x6c\x64`, map[string]string{"key": "hello", "value": "world"}, `$ $$key=hello, $value=world`},
	{`escape=json;{"$key":"$value"}`, `{"$key\\":"$value\""}`, map[string]string{"key": "$key\\", "value": "$value\""}, `{"$key\\":"$value\""}`},
	{`escape=json;{"$key":"$value"}`, `{"$key\\\"":"$value\"\\"}`, map[string]string{"key": "$key\\\"", "value": "$value\"\\"}, `{"$key\\\"":"$value\"\\"}`},
	{`escape=json;{"${key}":"${value}"}`, `{"$key\\\"":"$value\"\\"}`, map[string]string{"key": "$key\\\"", "value": "$value\"\\"}, `{"$key\\\"":"$value\"\\"}`},
}

func TestStructCodec(t *testing.T) {
	for _, tc := range positiveStruct {
		ngx, err := Compile(tc.Fmt)
		if err != nil {
			t.Fatalf("failed to Compile() format %q: %v", tc.Fmt, err)
		}

		var got Access
		if err := ngx.Unmarshal([]byte(tc.Data), &got); err != nil {
			t.Fatalf("failed to Unmarshal() data %q: %v", tc.Data, err)
		}
		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in Unmarshal(): expecting %q, got %q", tc.Expected, got)
		}

		marshaledBytes, err := ngx.Marshal(got)
		if err != nil {
			t.Fatalf("failed to Marshal() data %q: %v", got, err)
		}
		if bytes.Compare(marshaledBytes, []byte(tc.Marshaled)) != 0 {
			t.Fatalf("corrupted data in Marshal(): expecting %q, got %q", tc.Marshaled, marshaledBytes)
		}

		got = Access{}
		if err := ngx.UnmarshalFromString(tc.Data, &got); err != nil {
			t.Fatalf("failed to UnmarshalFromString() data %q: %v", tc.Data, err)
		}
		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in UnmarshalFromString(): expecting %q, got %q", tc.Expected, got)
		}

		marshaled, err := ngx.MarshalToString(got)
		if err != nil {
			t.Fatalf("failed to MarshalToString() data %q: %v", got, err)
		}
		if marshaled != tc.Marshaled {
			t.Fatalf("corrupted data in MarshalToString(): expecting %q, got %q", tc.Marshaled, marshaled)
		}
	}
}

func TestMapCodec(t *testing.T) {
	for _, tc := range positiveMap {
		ngx, err := Compile(tc.Fmt)
		if err != nil {
			t.Fatalf("failed to Compile() format %q: %v", tc.Fmt, err)
		}

		got := make(map[string]string)
		if err := ngx.Unmarshal([]byte(tc.Data), &got); err != nil {
			t.Fatalf("failed to Unmarshal() data %q: %v", tc.Data, err)
		}
		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in Unmarshal(): expecting %q, got %q", tc.Expected, got)
		}

		marshaledBytes, err := ngx.Marshal(got)
		if err != nil {
			t.Fatalf("failed to Marshal() data %q: %v", got, err)
		}
		if bytes.Compare(marshaledBytes, []byte(tc.Marshaled)) != 0 {
			t.Fatalf("corrupted data in Marshal(): expecting %q, got %q", tc.Marshaled, marshaledBytes)
		}

		got = make(map[string]string)
		if err := ngx.UnmarshalFromString(tc.Data, &got); err != nil {
			t.Fatalf("failed to UnmarshalFromString() data %q: %v", tc.Data, err)
		}
		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in UnmarshalFromString(): expecting %q, got %q", tc.Expected, got)
		}

		marshaled, err := ngx.MarshalToString(got)
		if err != nil {
			t.Fatalf("failed to MarshalToString() data %q: %v", got, err)
		}
		if marshaled != tc.Marshaled {
			t.Fatalf("corrupted data in MarshalToString(): expecting %q, got %q", tc.Marshaled, marshaled)
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
