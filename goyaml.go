package goyaml

import (
    "reflect"
    "os"
)


func Unmarshal(in []byte, out interface{}) os.Error {
    d := newDecoder(in)
    defer d.destroy()
    d.unmarshal(reflect.NewValue(out))
    return nil
}
