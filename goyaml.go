package goyaml

import (
    "reflect"
    "strings"
    "sync"
    "os"
)


func Unmarshal(in []byte, out interface{}) os.Error {
    d := newDecoder(in)
    defer d.destroy()
    d.unmarshal(reflect.NewValue(out))
    return nil
}


// --------------------------------------------------------------------------
// Maintain a mapping of keys to structure field indexes

// The code in this section was copied from gobson.

type structFields struct {
    Map map[string]fieldInfo
    List []fieldInfo
}

type fieldInfo struct {
    Key string
    Num int
    Conditional bool
}

var fieldMap = make(map[string]*structFields)
var fieldMapMutex sync.RWMutex

func getStructFields(st *reflect.StructType) (*structFields, os.Error) {
    path := st.PkgPath()
    name := st.Name()

    fullName := path + "." + name
    fieldMapMutex.RLock()
    fields, found := fieldMap[fullName]
    fieldMapMutex.RUnlock()
    if found {
        return fields, nil
    }

    n := st.NumField()
    fieldsMap := make(map[string]fieldInfo)
    fieldsList := make([]fieldInfo, n)
    for i := 0; i != n; i++ {
        field := st.Field(i)
        if field.PkgPath != "" {
            continue // Private field
        }

        info := fieldInfo{Num: i}

        if s := strings.LastIndex(field.Tag, "/"); s != -1 {
            for _, c := range field.Tag[s+1:] {
                switch c {
                case int('c'):
                    info.Conditional = true
                default:
                    panic("Unsupported field flag: " + string([]int{c}))
                }
            }
            field.Tag = field.Tag[:s]
        }

        if field.Tag != "" {
            info.Key = field.Tag
        } else {
            info.Key = strings.ToLower(field.Name)
        }

        if _, found = fieldsMap[info.Key]; found {
            msg := "Duplicated key '" + info.Key + "' in struct " + st.String()
            return nil, os.NewError(msg)
        }

        fieldsList[len(fieldsMap)] = info
        fieldsMap[info.Key] = info
    }

    fields = &structFields{fieldsMap, fieldsList[:len(fieldsMap)]}

    if fullName != "." {
        fieldMapMutex.Lock()
        fieldMap[fullName] = fields
        fieldMapMutex.Unlock()
    }

    return fields, nil
}

func isZero(v reflect.Value) bool {
    switch v := v.(type) {
    case *reflect.StringValue:
        return len(v.Get()) == 0
    case *reflect.InterfaceValue:
        return v.IsNil()
    case *reflect.SliceValue:
        return v.Len() == 0
    case *reflect.MapValue:
        return v.Len() == 0
    case *reflect.IntValue:
        return v.Get() == 0
    case *reflect.UintValue:
        return v.Get() == 0
    case *reflect.BoolValue:
        return !v.Get()
    }
    return false
}
