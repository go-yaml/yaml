package yaml_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var UnmarshalWithEnvParametrize = []struct {
	name      string
	yaml_str  string
	env_vars  map[interface{}]interface{}
	expection map[interface{}]interface{}
}{
	{
		name:     "no default value, no env value",
		yaml_str: "FOO: ${BAR}",
		env_vars: make(map[interface{}]interface{}),
		expection: map[interface{}]interface{}{
			"FOO": (interface{})(nil),
		},
	},
	{
		name:     "use default value if env value not provided",
		yaml_str: "FOO: ${BAR:foo}",
		env_vars: make(map[interface{}]interface{}),
		expection: map[interface{}]interface{}{
			"FOO": "foo",
		},
	},
	{
		name:     "use env value",
		yaml_str: "FOO: ${BAR}",
		env_vars: map[interface{}]interface{}{
			"BAR": "bar",
		},
		expection: map[interface{}]interface{}{
			"FOO": "bar",
		},
	},
	{
		name:     "use env value instead of default",
		yaml_str: "FOO: ${BAR:foo}",
		env_vars: map[interface{}]interface{}{
			"BAR": "bar",
		},
		expection: map[interface{}]interface{}{
			"FOO": "bar",
		},
	},
	{
		name: "handles multi line",
		yaml_str: `
        FOO: ${BAR:foo}
        BAR: ${FOO:bar}
        `,
		env_vars: map[interface{}]interface{}{
			"BAR": "bar", "FOO": "foo",
		},
		expection: map[interface{}]interface{}{
			"FOO": "bar", "BAR": "foo",
		},
	},
	{
		name:     "multiple substitutions",
		yaml_str: "FOO: http://${BAR:foo}/${FOO:bar}",
		env_vars: map[interface{}]interface{}{
			"BAR": "bar", "FOO": "foo",
		},
		expection: map[interface{}]interface{}{
			"FOO": "http://bar/foo",
		},
	},
	{
		name: "can handle int, float and boolean",
		yaml_str: `
        FOO:
            - BAR: ${INT:1}
            - BAR: ${FLOAT:1.1}
            - BAR: ${BOOL:True}
        `,
		env_vars: map[interface{}]interface{}{},
		expection: map[interface{}]interface{}{
			"FOO": []interface{}{
				map[string]interface{}{
					"BAR": 1,
				},
				map[string]interface{}{
					"BAR": 1.1,
				},
				map[string]interface{}{
					"BAR": true,
				},
			},
		},
	},
	{
		name: "quoted default results in string",
		yaml_str: `
        FOO:
            - BAR: ${INT:"1"}
            - BAR: ${FLOAT:"1.1"}
            - BAR: ${BOOL:"True"}
        `,
		env_vars: map[interface{}]interface{}{},
		expection: map[interface{}]interface{}{
			"FOO": []interface{}{
				map[string]interface{}{
					"BAR": "1",
				},
				map[string]interface{}{
					"BAR": "1.1",
				},
				map[string]interface{}{
					"BAR": "True",
				},
			},
		},
	},
	{
		name: "inline list",
		yaml_str: `
        FOO: ${LIST}
        BAR:
            - 1
            - 2
            - 3
        `,
		env_vars: map[interface{}]interface{}{
			"LIST": "[1,2,3]",
		},
		expection: map[interface{}]interface{}{
			"FOO": []interface{}{1, 2, 3},
			"BAR": []interface{}{1, 2, 3},
		},
	},
	{
		name: "inline list containing list",
		yaml_str: `
        FOO: ${LIST}
        BAR:
            - 1
            - 2
            - 3
        `,
		env_vars: map[interface{}]interface{}{
			"LIST": "[1, 2, 3, ['a', 'b', 'c']]",
		},
		expection: map[interface{}]interface{}{
			"FOO": []interface{}{1, 2, 3, []interface{}{"a", "b", "c"}},
			"BAR": []interface{}{1, 2, 3},
		},
	},
	{
		name: "inline dict",
		yaml_str: `
        FOO: ${DICT}
        BAR:
            - 1
            - 2
            - 3
        `,
		env_vars: map[interface{}]interface{}{
			"DICT": "{'one': 1, 'two': 2}",
		},
		expection: map[interface{}]interface{}{
			"FOO": map[string]interface{}{
				"one": 1,
				"two": 2,
			},
			"BAR": []interface{}{1, 2, 3},
		},
	},
}

func TestUnmarshalWithEnv(t *testing.T) {
	for _, tc := range UnmarshalWithEnvParametrize {
		t.Run(tc.name, func(t *testing.T) {
			// In each for loop, merge tc.env_vars to os.environ
			for k, v := range tc.env_vars {
				os.Setenv(k.(string), v.(string))
			}
			var m map[interface{}]interface{}
			err := yaml.Unmarshal(tc.yaml_str, &m)
			assert.NoError(t, err)
			assert.Equal(t, tc.expection, m)
			// Clean up env vars
			for k := range tc.env_vars {
				os.Unsetenv(k.(string))
			}
		})
	}
}
