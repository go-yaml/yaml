package goyaml_test


import (
    . "gocheck"
    "testing"
    "goyaml"
    "reflect"
    "math"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

type testItem struct {
    data string
    value interface{}
}


var unmarshalTests = []testItem{
    // It will encode either value as a string if asked for.
    {"hello: world", map[string]string{"hello": "world"}},
    {"hello: true", map[string]string{"hello": "true"}},

    // And when given the option, will preserve the YAML type.
    {"hello: world", map[string]interface{}{"hello": "world"}},
    {"hello: true", map[string]interface{}{"hello": true}},
    {"hello: 10", map[string]interface{}{"hello": 10}},
    {"hello: 0b10", map[string]interface{}{"hello": 2}},
    {"hello: 0xA", map[string]interface{}{"hello": 10}},
    {"hello: 4294967296", map[string]interface{}{"hello": int64(4294967296)}},
    {"hello: 4294967296", map[string]int64{"hello": int64(4294967296)}},
    {"hello: 0.1", map[string]interface{}{"hello": 0.1}},
    {"hello: .1", map[string]interface{}{"hello": 0.1}},
    {"hello: .Inf", map[string]interface{}{"hello": math.Inf(+1)}},
    {"hello: -.Inf", map[string]interface{}{"hello": math.Inf(-1)}},
    {"hello: -10", map[string]interface{}{"hello": -10}},
    {"hello: -.1", map[string]interface{}{"hello": -0.1}},

    // Floats from spec
    {"canonical: 6.8523e+5", map[string]interface{}{"canonical": 6.8523e+5}},
    {"expo: 685.230_15e+03", map[string]interface{}{"expo": 685.23015e+03}},
    {"fixed: 685_230.15", map[string]interface{}{"fixed": 685230.15}},
    //{"sexa: 190:20:30.15", map[string]interface{}{"sexa": 0}}, // Unsupported
    {"neginf: -.inf", map[string]interface{}{"neginf": math.Inf(-1)}},
    {"notanum: .NaN", map[string]interface{}{"notanum": math.NaN}},
    {"fixed: 685_230.15", map[string]float{"fixed": 685230.15}},

    // Bools from spec
    {"canonical: y", map[string]interface{}{"canonical": true}},
    {"answer: NO", map[string]interface{}{"answer": false}},
    {"logical: True", map[string]interface{}{"logical": true}},
    {"option: on", map[string]interface{}{"option": true}},
    {"option: on", map[string]bool{"option": true}},

    // Ints from spec
    {"canonical: 685230", map[string]interface{}{"canonical": 685230}},
    {"decimal: +685_230", map[string]interface{}{"decimal": 685230}},
    {"octal: 02472256", map[string]interface{}{"octal": 685230}},
    {"hexa: 0x_0A_74_AE", map[string]interface{}{"hexa": 685230}},
    {"bin: 0b1010_0111_0100_1010_1110", map[string]interface{}{"bin": 685230}},
    //{"sexa: 190:20:30", map[string]interface{}{"sexa": 0}}, // Unsupported
    {"decimal: +685_230", map[string]int{"decimal": 685230}},

    // Sequence
    {"seq: [A,B]", map[string]interface{}{"seq": []interface{}{"A", "B"}}},
    {"seq: [A,B,C]", map[string][]string{"seq": []string{"A", "B", "C"}}},
    {"seq: [A,1,C]", map[string][]string{"seq": []string{"A", "1", "C"}}},
    {"seq: [A,1,C]", map[string][]int{"seq": []int{1}}},
    {"seq: [A,1,C]", map[string]interface{}{"seq": []interface{}{"A", 1, "C"}}},

    // Map inside interface with no type hints.
    {"a: {b: c}",
     map[string]interface{}{"a": map[interface{}]interface{}{"b": "c"}}},

    // Structs and type conversions.
    {"hello: world", &struct{Hello string}{"world"}},
    {"a: {b: c}", &struct{A struct{B string}}{struct{B string}{"c"}}},
    {"a: {b: c}", &struct{A *struct{B string}}{&struct{B string}{"c"}}},
    {"a: 1", &struct{A int}{1}},
    {"a: [1, 2]", &struct{A []int}{[]int{1, 2}}},
    {"a: 1", &struct{B int}{0}},
    {"a: 1", &struct{B int "a"}{1}},
    {"a: y", &struct{A bool}{true}},
}


func (s *S) TestUnmarshal(c *C) {
    for _, item := range unmarshalTests {
        t := reflect.NewValue(item.value).Type()
        var value interface{}
        if t, ok := t.(*reflect.MapType); ok {
            value = reflect.MakeMap(t).Interface()
        } else {
            pt := reflect.NewValue(item.value).Type().(*reflect.PtrType)
            pv := reflect.MakeZero(pt).(*reflect.PtrValue)
            pv.PointTo(reflect.MakeZero(pt.Elem()))
            value = pv.Interface()
        }
        err := goyaml.Unmarshal([]byte(item.data), value)
        c.Assert(err, IsNil)
        c.Assert(value, Equals, item.value)
    }
}
