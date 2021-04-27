package yaml

import (
	"reflect"
	"strings"
	"unicode"
)

type keyList []reflect.Value

func (l keyList) Len() int      { return len(l) }
func (l keyList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l keyList) Less(i, j int) bool {
	a := l[i]
	b := l[j]
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
	if ak != reflect.String || bk != reflect.String {
		return ak < bk
	}
	ar, br := []rune(a.String()), []rune(b.String())
	for i := 0; i < len(ar) && i < len(br); i++ {
		var adigits, bdigits int
		for j := i; j < len(ar) && unicode.IsDigit(ar[j]); j++ {
			adigits++
		}
		for j := i; j < len(br) && unicode.IsDigit(br[j]); j++ {
			bdigits++
		}
		if adigits > 0 && bdigits > 0 {
			// if both a and b have a sequence of digits
			// starting here, and they're not identical,
			// sort them by the expressed number ("9" <
			// "077") then by length ("01" < "001").
			var azeroes, bzeroes int
			for j := i; j < len(ar) && ar[j] == '0'; j++ {
				azeroes++
			}
			for j := i; j < len(br) && br[j] == '0'; j++ {
				bzeroes++
			}
			if cmp := (adigits - azeroes) - (bdigits - bzeroes); cmp != 0 {
				// with leading zeroes removed,
				// shorter numbers are always smaller
				return cmp < 0
			}
			if cmp := strings.Compare(string(ar[i+azeroes:i+adigits]), string(br[i+bzeroes:i+bdigits])); cmp != 0 {
				// with leading zeroes removed,
				// equal-length numbers sort in the
				// same order as their string
				// representation
				return cmp < 0
			}
			if cmp := azeroes - bzeroes; cmp != 0 {
				return cmp < 0
			}
			// the next adigits==bdigits runes are equal
			i += adigits - 1
			continue
		}
		if al, bl := unicode.IsLetter(ar[i]), unicode.IsLetter(br[i]); al != bl {
			return bl
		}
		if ar[i] != br[i] {
			return ar[i] < br[i]
		}
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
