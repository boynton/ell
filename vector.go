/*
Copyright 2015 Lee Boynton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ell

import (
	. "github.com/boynton/ell/data" // -> "github.com/boynton/data"
)

/*
// VectorEqual - return true of the two vectors are equal, i.e. the same length and
// all the elements are also equal
func VectorEqual(v1 *Vector, v2 *Vector) bool {
	el1 := v1.Elements
	el2 := v2.Elements
	count := len(el1)
	if count != len(el2) {
		return false
	}
	for i := 0; i < count; i++ {
		if !Equal(el1[i], el2[i]) {
			return false
		}
	}
	return true
}

func VectorToString(vec *Vector) string {
	el := vec.Elements
	var buf bytes.Buffer
	buf.WriteString("[")
	count := len(el)
	if count > 0 {
		buf.WriteString(el[0].String())
		for i := 1; i < count; i++ {
			buf.WriteString(" ")
			buf.WriteString(el[i].String())
		}
	}
	buf.WriteString("]")
	return buf.String()
}

// MakeVector - create a new <vector> object of the specified size, with all elements initialized to
// the specified value
func MakeVector(size int, init Value) *Vector {
	elements := make([]Value, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return VectorFromElementsNoCopy(elements)
}

// Vector - create a new <vector> object from the given element objects.
func Vector(elements ...Value) *Vector {
	return VectorFromElements(elements, len(elements))
}

// VectorFromElements - return a new <vector> object from the given slice of elements. The slice is copied.
func VectorFromElements(elements []Value, count int) *Vector {
	el := make([]Value, count)
	copy(el, elements[0:count])
	return VectorFromElementsNoCopy(el)
}

// VectorFromElementsNoCopy - create a new <vector> object from the given slice of elements. The slice is NOT copied.
func VectorFromElementsNoCopy(elements []Value) *Vector {
	return &Vector{
		Elements: elements,
	}
}

// CopyVector - return a copy of the <vector>
func CopyVector(vec *Vector) *Vector {
	return VectorFromElements(vec.Elements, len(vec.Elements))
}
*/

// ToVector - convert the object to a <vector>, if possible
func ToVector(obj Value) (*Vector, error) {
	switch p := obj.(type) {
	case *Vector:
		return p, nil
	case *List:
		return ListToVector(p), nil
	case *Struct:
		return StructToVector(p), nil
	case *String:
		return StringToVector(p), nil
	}
	return nil, NewError(ArgumentErrorKey, "to-vector expected <vector>, <list>, <struct>, or <string>, got a ", obj.Type())
}

func IsVector(obj Value) bool {
	if _, ok := obj.(*Vector); ok {
		return true
	}
	return false
}
