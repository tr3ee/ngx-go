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

		// if len(tc.Expected) != len(got) {
		// 	t.Fatalf("Corrupted data in unmarshal: expecting len(%d), got len(%d)", len(tc.Expected), len(got))
		// }

		// for k, v := range tc.Expected {
		// 	if got[k] != v {
		// 		t.Fatalf("Corrupted data in unmarshal: expecting %s:%s, got %s:%s", k, v, k, got[k])
		// 	}
		// }

		if !reflect.DeepEqual(got, tc.Expected) {
			t.Fatalf("corrupted data in unmarshal: expecting %v, got %v", tc.Expected, got)
		}
	}
}
