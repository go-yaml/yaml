package goyaml

import (
    "strconv"
    "strings"
    "math"
)


// TODO: Support merge, timestamps, and base 60 floats.


type stdTag int

var StrTag = stdTag(1)
var BoolTag = stdTag(2)
var IntTag = stdTag(3)
var FloatTag = stdTag(4)

func (t stdTag) String() string {
    switch t {
    case StrTag:
        return "tag:yaml.org,2002:str"
    case BoolTag:
        return "tag:yaml.org,2002:bool"
    case IntTag:
        return "tag:yaml.org,2002:int"
    case FloatTag:
        return "tag:yaml.org,2002:float"
    default:
        panic("Internal error: missing tag case")
    }
    return ""
}


type resolveMapItem struct {
    value interface{}
    tag stdTag
}

var resolveTable = make([]byte, 256)
var resolveMap = make(map[string]resolveMapItem)


func init() {
    t := resolveTable
    t[int('+')] = 'S' // Sign
    t[int('-')] = 'S'
    for _, c := range "0123456789" {
        t[int(c)] = 'D' // Digit
    }
    for _, c := range "yYnNtTfFoO" {
        t[int(c)] = 'M' // In map
    }
    t[int('.')] = '.' // Float (potentially in map)
    t[int('<')] = '<' // Merge

    var resolveMapList = []struct{v interface{}; tag stdTag; l []string} {
        {true, BoolTag, []string{"y", "Y", "yes", "Yes", "YES"}},
        {true, BoolTag, []string{"true", "True", "TRUE"}},
        {true, BoolTag, []string{"on", "On", "ON"}},
        {false, BoolTag, []string{"n", "N", "no", "No", "NO"}},
        {false, BoolTag, []string{"false", "False", "FALSE"}},
        {false, BoolTag, []string{"off", "Off", "OFF"}},
        {math.NaN, FloatTag, []string{".nan", ".NaN", ".NAN"}},
        {math.Inf(+1), FloatTag, []string{".inf", ".Inf", ".INF"}},
        {math.Inf(+1), FloatTag, []string{"+.inf", "+.Inf", "+.INF"}},
        {math.Inf(-1), FloatTag, []string{"-.inf", "-.Inf", "-.INF"}},
    }

    m := resolveMap
    for _, item := range resolveMapList {
        for _, s := range item.l {
            m[s] = resolveMapItem{item.v, item.tag}
        }
    }
}

func resolve(in string) (out interface{}, tag stdTag) {
    if in == "" {
        return in, tag
    }
    c := resolveTable[in[0]]
    if c == 0 {
        // It's a string for sure. Nothing to do.
        return in, StrTag
    }

    // Handle things we can lookup in a map.
    if item, ok := resolveMap[in]; ok {
        return item.value, item.tag
    }

    switch c {
    case 'M':
        // We've already checked the map above.

    case '.':
        // Not in the map, so maybe a normal float.
        floatv, err := strconv.Atof(in)
        if err == nil {
            return floatv, FloatTag
        }
        // XXX Handle base 60 floats here.

    case 'D', 'S':
        // Int, float, or timestamp.
        for i := 0; i != len(in); i++ {
            if in[i] == '_' {
                in = strings.Replace(in, "_", "", -1)
                break
            }
        }
        intv, err := strconv.Btoi64(in, 0)
        if err == nil {
            if intv == int64(int(intv)) {
                return int(intv), IntTag
            } else {
                return intv, IntTag
            }
        }
        floatv, err := strconv.Atof(in)
        if err == nil {
            return floatv, FloatTag
        }
        // XXX Handle timestamps here.

    case '<':
        // XXX Handle merge (<<) here.

    default:
        panic("resolveTable item not yet handled: " +
              string([]byte{c}) + " (with " + in +")")
    }
    return in, StrTag
}
