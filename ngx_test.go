package ngx

import "testing"

type logmsg struct {
	Str   string `ngx:"str"`
	Int   int    `ngx:"int"`
	Uint  uint   `ngx:"uint"`
	Byte  byte   `ngx:"byte"`
	P2Int *int   `ngx:"p2int"`
	Empty string
	_     *logmsg
}

var dead = 0xdead

var testcases = []struct {
	LogFormat  string
	CompileErr error
	RawMessage string
	Expected   logmsg
}{
	{"Str = $str, Int = $int, Uint = $uint, Byte = $byte, P2Int = $p2int, and Empty = $empty", nil, `Str = tr\\3e, Int = 57005, Uint = 1000, Byte = T, P2Int = 57005, and Empty = not empty at all`, logmsg{"tr\\3e", dead, 1000, 'T', &dead, "", nil}},
	{"Str = \"$str\", Int = $int, Uint = $uint, Byte = $byte, P2Int = $p2int, and Empty = $empty", nil, `Str = "tr\", Int = 3e", Int = 57005, Uint = 1000, Byte = T, P2Int = 57005, and Empty = not empty at all`, logmsg{"tr\", Int = 3e", dead, 1000, 'T', &dead, "", nil}},
}

func TestCompile(t *testing.T) {
	for _, tc := range testcases {
		_, err := Compile(tc.LogFormat)
		if tc.CompileErr != err {
			t.Fatalf("Failed on %q: expected err=%v, got %v", tc.LogFormat, tc.CompileErr, err)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range testcases {
		if tc.CompileErr != nil {
			continue
		}
		ngx, err := Compile(tc.LogFormat)
		if err != nil {
			t.Fatal(err)
		}
		msg := new(logmsg)
		if err := ngx.Unmarshal([]byte(tc.RawMessage), msg); err != nil {
			t.Fatal(err)
		}
		if msg.Str != tc.Expected.Str {
			t.Fatalf("Failed on %q: expected Str = %v, got %v", tc.RawMessage, tc.Expected.Str, msg.Str)
		}
		if msg.Int != tc.Expected.Int {
			t.Fatalf("Failed on %q: expected Int = %v, got %v", tc.RawMessage, tc.Expected.Int, msg.Int)
		}
		if msg.Uint != tc.Expected.Uint {
			t.Fatalf("Failed on %q: expected Uint = %v, got %v", tc.RawMessage, tc.Expected.Uint, msg.Uint)
		}
		if msg.Byte != tc.Expected.Byte {
			t.Fatalf("Failed on %q: expected Byte = %v, got %v", tc.RawMessage, tc.Expected.Byte, msg.Byte)
		}
		if *msg.P2Int != *tc.Expected.P2Int {
			t.Fatalf("Failed on %q: expected P2Int = %v, got %v", tc.RawMessage, *tc.Expected.P2Int, *msg.P2Int)
		}
		if msg.Empty != tc.Expected.Empty {
			t.Fatalf("Failed on %q: expected Empty = %v, got %v", tc.RawMessage, tc.Expected.Empty, msg.Empty)
		}
	}
}
