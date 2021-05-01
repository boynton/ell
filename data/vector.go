/*
Copyright 2021 Lee Boynton

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
package data

import(
	"bytes"
)

type Vector struct {
	Elements []Value
}

var EmptyVector *Vector = VectorFromElementsNoCopy(nil) //NewVector()

func NewVector(elements ...Value) *Vector {
	return VectorFromElements(elements, len(elements))
}

func MakeVector(size int, init Value) *Vector {
	elements := make([]Value, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
    return VectorFromElementsNoCopy(elements)
}

func VectorFromElementsNoCopy(elements []Value) *Vector {
	return &Vector{Elements: elements}
}

// VectorFromElements - return a new <vector> object from the given slice of elements. The slice is copied.
func VectorFromElements(elements []Value, count int) *Vector {
	el := make([]Value, count)
	copy(el, elements[0:count])
	return VectorFromElementsNoCopy(el)
}

func (v *Vector) Type() Value {
	return VectorType
}

func (v *Vector) String() string {
	el := v.Elements
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

func (v1 *Vector) Equals(another Value) bool {
	if v2, ok := another.(*Vector); ok {
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
	return false
}
