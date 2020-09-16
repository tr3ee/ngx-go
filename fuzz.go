// +build gofuzz

package ngx

func FuzzCompile(p []byte) int {
	if _, err := Compile(string(p)); err != nil {
		return 0
	}
	return 1
}

func FuzzUnmarshalStruct(p []byte) int {
	if err := Unmarshal(p, new(Access)); err != nil {
		return 0
	}
	return 1
}

func FuzzUnmarshalMap(p []byte) int {
	m := make(map[string]string)
	if err := Unmarshal(p, &m); err != nil {
		return 0
	}
	return 1
}
