// +build gofuzz

package ngx

func FuzzUnmarshal(p []byte) int {
	if err := Unmarshal(p, new(Access)); err != nil {
		return 0
	}
	return 1
}
