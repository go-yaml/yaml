// +build gofuzz

package yaml

func Fuzz(b []byte) int {
	var v interface{}
	if err := Unmarshal(b, &v); err != nil {
		return 0
	}
	return 1
}
