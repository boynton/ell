/*
Copyright 2014 Lee Boynton

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
	"bytes"
)

func VectorEqual(v1 *LOB, v2 *LOB) bool {
	el1 := v1.elements
	el2 := v2.elements
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

func vectorToString(vec *LOB) string {
	el := vec.elements
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

func MakeVector(size int, init *LOB) *LOB {
	elements := make([]*LOB, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return VectorFromElementsNoCopy(elements)
}

func Vector(elements ...*LOB) *LOB {
	return VectorFromElements(elements, len(elements))
}

func VectorFromElements(elements []*LOB, count int) *LOB {
	el := make([]*LOB, count)
	copy(el, elements[0:count])
	return VectorFromElementsNoCopy(el)
}

func VectorFromElementsNoCopy(elements []*LOB) *LOB {
	vec := new(LOB)
	vec.Type = VectorType
	vec.elements = elements
	return vec
}

func CopyVector(vec *LOB) *LOB {
	return VectorFromElements(vec.elements, len(vec.elements))
}

func ToVector(obj *LOB) (*LOB, error) {
	switch obj.Type {
	case VectorType:
		return obj, nil
	case ListType:
		return listToVector(obj), nil
	case StructType:
		return structToVector(obj), nil
	case StringType:
		return stringToVector(obj), nil
	}
	return nil, Error(ArgumentErrorKey, "to-vector expected <vector>, <list>, <struct>, or <string>, got a ", obj.Type)
}
