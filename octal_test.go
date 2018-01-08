package yaml

import (
	"testing"
)

func TestOctal(t *testing.T) {
	o := New("0660")
	n := o.Int64()
	if n != 432 {
		t.Error("The octal conversion is wrong, the result was: ", n)
	}
}

func TestOctalString(t *testing.T) {
	o := New("0775")
	s := o.String()
	if s != "0775" {
		t.Error("The octal string conversion is wrong, the result was: ", s)
	}
}
