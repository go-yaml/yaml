package yaml

import (
	"reflect"
	"testing"
)

// TestLessFullyTransitive checks that the Less function is fully transitive.
// This means that for every element in a known list, checking that the element
// is greater then all elements before it, and less than all elements after it.
func TestLessFullyTransitive(t *testing.T) {
	order := []interface{}{
		false,
		true,
		1,
		uint(1),
		1.0,
		1.1,
		1.2,
		2,
		uint(2),
		2.0,
		2.1,
		"",
		".1",
		".2",
		".a",
		"1",
		"2",
		"a!10",
		"a/0001",
		"a/002",
		"a/3",
		"a/10",
		"a/11",
		"a/0012",
		"a/100",
		"a~10",
		"ab/1",
		"b/1",
		"b/01",
		"b/2",
		"b/02",
		"b/3",
		"b/03",
		"b1",
		"b01",
		"b3",
		"c2.10",
		"c10.2",
		"d1",
		"d7",
		"d7abc",
		"d12",
		"d12a",
		"z1a",
		"z01",
		"z13",
	}

	var kl = make(keyList, len(order))
	for index, item := range order {
		kl[index] = reflect.ValueOf(item)
	}

	// Compare every element, against every other element.
	for indexFirst, itemFirst := range kl {
		for indexSecond, itemSecond := range kl {
			switch {
			case indexFirst == indexSecond:
				continue
			case indexFirst < indexSecond:
				if !kl.Less(indexFirst, indexSecond) {
					t.Fatalf("expected %v < %v", itemFirst, itemSecond)
				}
			case indexFirst > indexSecond:
				if kl.Less(indexFirst, indexSecond) {
					t.Fatalf("expected %v > %v", itemFirst, itemSecond)
				}
			}
		}
	}
}
