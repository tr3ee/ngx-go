// +build gofuzz

package ngx

func FuzzCompile(p []byte) int {
	if _, err := Compile(string(p)); err != nil {
		return 0
	}
	return 1
}

func FuzzUnmarshal(p []byte) int {
	if err := Unmarshal(p, new(Access)); err != nil {
		return 0
	}
	return 1
}
