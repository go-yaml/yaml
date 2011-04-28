//
// goyaml - YAML support for the Go language
//
//   https://wiki.ubuntu.com/goyaml
//
// Copyright (c) 2011 Canonical Ltd.
//
// Written by Gustavo Niemeyer <gustavo.niemeyer@canonical.com>
//
package goyaml

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
	"os"
)

func handleErr(err *os.Error) {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		} else if _, ok := r.(*reflect.ValueError); ok {
			panic(r)
		} else if s, ok := r.(string); ok {
			*err = os.ErrorString("YAML error: " + s)
		} else if e, ok := r.(os.Error); ok {
			*err = e
		} else {
			panic(r)
		}
	}
}

// Objects implementing the goyaml.Setter interface will receive the YAML
// tag and value via the SetYAML method during unmarshaling, rather than
// being implicitly assigned by the goyaml machinery.  If setting the value
// works, the method should return true.  If it returns false, the given
// value will be omitted from maps and slices.
type Setter interface {
	SetYAML(tag string, value interface{}) bool
}

// Objects implementing the goyaml.Getter interface will get the GetYAML()
// method called when goyaml is requested to marshal the given value, and
// the result of this method will be marshaled in place of the actual object.
type Getter interface {
	GetYAML() (tag string, value interface{})
}

// Unmarshal decodes the first document found within the in byte slice
// and assigns decoded values into the object pointed by out.
//
// Maps, pointers to structs and ints, etc, may all be used as out values.
// If an internal pointer within a struct is not initialized, goyaml
// will initialize it if necessary for unmarshalling the provided data,
// but the struct provided as out must not be a nil pointer.
// 
// The type of the decoded values and the type of out will be considered,
// and Unmarshal() will do the best possible job to unmarshal values
// appropriately.  It is NOT considered an error, though, to skip values
// because they are not available in the decoded YAML, or if they are not
// compatible with the out value. To ensure something was properly
// unmarshaled use a map or compare against the previous value for the
// field (usually the zero value).
//
// Struct fields are only unmarshalled if they are exported (have an
// upper case first letter), and will be unmarshalled using the field
// name lowercased by default. When custom field names are desired, the
// tag value may be used to tweak the name. Everything before the last
// slash in the field tag will be used as the name. The value following
// the slash is used to tweak the marshalling process (see Marshal).
//
// For example:
//
//     type T struct {
//         F int "a/c"
//         B int
//     }
//     var T t
//     goyaml.Unmarshal([]byte("a: 1\nb: 2"), &t)
// 
func Unmarshal(in []byte, out interface{}) (err os.Error) {
	defer handleErr(&err)
	d := newDecoder()
	p := newParser(in)
	defer p.destroy()
	node := p.parse()
	if node != nil {
		d.unmarshal(node, reflect.ValueOf(out))
	}
	return nil
}

// Marshal writes the object provided into a YAML document. The structure
// of the generated document will reflect the structure of the value itself.
// Maps, pointers to structs and ints, etc, may all be used as the in value.
//
// Struct fields are only marshalled if they are exported (have an
// upper case first letter), and will be marshalled using the field
// name lowercased by default. When custom field names are desired, the
// tag value may be used to tweak the name. Everything before the last
// slash in the field tag will be used as the name. The characters
// following the slash are used as flags for the marshalling process,
// with 'c' meaning conditional (only marshal if non-zero), and 'f'
// meaning use flow style (useful for structs, sequences and maps).
//
// For example:
//
//     type T struct {
//         F int "a/c"
//         B int
//     }
//     goyaml.Marshal(&T{B: 2}) // Returns "b: 2\n"
//     goyaml.Marshal(&T{F: 1}} // Returns "a: 1\nb: 0\n"
//
func Marshal(in interface{}) (out []byte, err os.Error) {
	defer handleErr(&err)
	e := newEncoder()
	defer e.destroy()
	e.marshal("", reflect.ValueOf(in))
	e.finish()
	out = e.out
	return
}


// --------------------------------------------------------------------------
// Maintain a mapping of keys to structure field indexes

// The code in this section was copied from gobson.

type structFields struct {
	Map  map[string]fieldInfo
	List []fieldInfo
}

type fieldInfo struct {
	Key         string
	Num         int
	Conditional bool
	Flow        bool
}

var fieldMap = make(map[string]*structFields)
var fieldMapMutex sync.RWMutex

func getStructFields(st reflect.Type) (*structFields, os.Error) {
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
				case int('f'):
					info.Flow = true
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
	switch v.Kind() {
	case reflect.String:
		return len(v.String()) == 0
	case reflect.Interface:
		return v.IsNil()
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Bool:
		return !v.Bool()
	}
	return false
}
