//
// Copyright (c) 2011-2019 Canonical Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yaml

import (
	"reflect"
	"sort"
	"unicode"
)

type reflectSorter []reflect.Value

func NewSortedKeys(mv reflect.Value) []reflect.Value {
	keys := mv.MapKeys()
	sort.Slice(keys, reflectSorter(keys).Less)
	return keys
}

func (l reflectSorter) Less(i, j int) bool {
	return lessByValues(l[i], l[j])
}

func lessByValues(a, b reflect.Value) bool {
	ak := a.Kind()
	bk := b.Kind()
	for (ak == reflect.Interface || ak == reflect.Ptr) && !a.IsNil() {
		a = a.Elem()
		ak = a.Kind()
	}
	for (bk == reflect.Interface || bk == reflect.Ptr) && !b.IsNil() {
		b = b.Elem()
		bk = b.Kind()
	}
	af, aok := keyFloat(a)
	bf, bok := keyFloat(b)
	if aok && bok {
		if af != bf {
			return af < bf
		}
		if ak != bk {
			return ak < bk
		}
		return numLess(a, b)
	}
	if ak == reflect.Struct && bk == ak {
		tp := a.Type()
		isEqual := false
		for fi, fc := 0, tp.NumField(); fi < fc; fi++ { //compare struct fields in declaration order
			if !tp.Field(fi).IsExported() {
				continue
			}
			if lessByValues(a.Field(fi), b.Field(fi)) {
				return true
			}
			if lessByValues(b.Field(fi), a.Field(fi)) {
				return false
			}
			isEqual = true
		}
		if isEqual {
			return false
		}
	}
	if ak != reflect.String || bk != reflect.String {
		return ak < bk
	}
	ar, br := []rune(a.String()), []rune(b.String())
	digits := false
	for i := 0; i < len(ar) && i < len(br); i++ {
		if ar[i] == br[i] {
			digits = unicode.IsDigit(ar[i])
			continue
		}
		al := unicode.IsLetter(ar[i])
		bl := unicode.IsLetter(br[i])
		if al && bl {
			return ar[i] < br[i]
		}
		if al || bl {
			if digits {
				return al
			} else {
				return bl
			}
		}
		var ai, bi int
		var an, bn int64
		if ar[i] == '0' || br[i] == '0' {
			for j := i - 1; j >= 0 && unicode.IsDigit(ar[j]); j-- {
				if ar[j] != '0' {
					an = 1
					bn = 1
					break
				}
			}
		}
		for ai = i; ai < len(ar) && unicode.IsDigit(ar[ai]); ai++ {
			an = an*10 + int64(ar[ai]-'0')
		}
		for bi = i; bi < len(br) && unicode.IsDigit(br[bi]); bi++ {
			bn = bn*10 + int64(br[bi]-'0')
		}
		if an != bn {
			return an < bn
		}
		if ai != bi {
			return ai < bi
		}
		return ar[i] < br[i]
	}
	return len(ar) < len(br)
}

// keyFloat returns a float value for v if it is a number/bool
// and whether it is a number/bool or not.
func keyFloat(v reflect.Value) (f float64, ok bool) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Bool:
		if v.Bool() {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

// numLess returns whether a < b.
// a and b must necessarily have the same kind.
func numLess(a, b reflect.Value) bool {
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() < b.Int()
	case reflect.Float32, reflect.Float64:
		return a.Float() < b.Float()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return a.Uint() < b.Uint()
	case reflect.Bool:
		return !a.Bool() && b.Bool()
	}
	panic("not a number")
}
