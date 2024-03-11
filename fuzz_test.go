//go:build go1.18
// +build go1.18

package yaml_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

// FuzzEncodeFromJSON checks that any JSON encoded value can also be encoded as YAML... and decoded.
func FuzzEncodeFromJSON(f *testing.F) {
	f.Add(`null`)
	f.Add(`""`)
	f.Add(`0`)
	f.Add(`true`)
	f.Add(`false`)
	f.Add(`{}`)
	f.Add(`[]`)
	f.Add(`[[]]`)
	f.Add(`{"a":[]}`)

	f.Fuzz(func(t *testing.T, s string) {

		var v interface{}
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Skip("not valid JSON")
		}

		t.Logf("JSON %q", s)
		t.Logf("Go   %q <%[1]x>", v)

		// Encode as YAML
		b, err := yaml.Marshal(v)
		if err != nil {
			t.Error(err)
		}
		t.Logf("YAML %q <%[1]x>", b)

		// Decode as YAML
		var v2 interface{}
		if err := yaml.Unmarshal(b, &v2); err != nil {
			t.Error(err)
		}

		t.Logf("Go   %q <%[1]x>", v2)

		/*
			// Handling of number is different, so we can't have universal exact matching
			if !reflect.DeepEqual(v2, v) {
				t.Errorf("mismatch:\n-      got: %#v\n- expected: %#v", v2, v)
			}
		*/

		b2, err := yaml.Marshal(v2)
		if err != nil {
			t.Error(err)
		}
		t.Logf("YAML %q <%[1]x>", b2)

		if !bytes.Equal(b, b2) {
			t.Errorf("Marshal->Unmarshal->Marshal mismatch:\n- expected: %q\n- got:      %q", b, b2)
		}

	})
}

func TestEncodeString(t *testing.T) {
	b, _ := yaml.Marshal(`\n`)
	t.Logf("%q <%[1]x>", string(b))
}
