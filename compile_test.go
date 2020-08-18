package ngx

import "testing"

var positiveFormats = []string{
	`$request "$request_body""$header_cookie"`,
	`$request "$request_body" "$header_cookie"`,
	`\$request "$request_body" "$header_cookie"`,
	`\$request "$request_body" "$header_cookie"`,
	`\$request\"$request_body\"\"$header_cookie\"`,
	`escape=json ; $request "$request_body""$header_cookie"`,
	`escape=none ; $request "$request_body""$header_cookie"`,
	`escape=default           		; $request "$request_body" "$header_cookie"`,
	`escape=json;$request "$request_body""$header.cookie"`,
}

var negativeFormats = []string{
	`escape=json$request "$request_body""$header_cookie"`,
	`escape=json;${request "$request_body""$header_cookie"`,
	`escape=json $request "$request_body""$header_cookie"`,
	`escape=unknown ;$request "$request_body""$header_cookie"`,
	`escape=json;$request "$request_body""$.cookie"`,
	`escape=json;$request "$request_body.""$cookie"`,
	`escape=json;$request "$request_body""$header..cookie"`,
	`escape=json;$request "$request_body""$header....cookie"`,
}

func TestCompile(t *testing.T) {
	for _, fmt := range positiveFormats {
		_, err := Compile(fmt)
		if err != nil {
			t.Fatalf("failed to compile %q: %v", fmt, err)
		}
	}

	for _, fmt := range negativeFormats {
		_, err := Compile(fmt)
		if err == nil {
			t.Fatalf("expecting compile error on %q", fmt)
		}
	}
}

func BenchmarkCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Compile(CombinedFmt); err != nil {
			b.Fatal(err)
		}
	}
}
