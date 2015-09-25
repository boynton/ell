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

package main

import (
	"bytes"
)

func vectorEqual(v1 *LAny, v2 *LAny) bool {
	count := len(v1.elements)
	if count != len(v2.elements) {
		return false
	}
	for i := 0; i < count; i++ {
		if !equal(v1.elements[i], v2.elements[i]) {
			return false
		}
	}
	return true
}

func vectorToString(vec *LAny) string {
	var buf bytes.Buffer
	buf.WriteString("[")
	count := len(vec.elements)
	if count > 0 {
		buf.WriteString(vec.elements[0].String())
		for i := 1; i < count; i++ {
			buf.WriteString(" ")
			buf.WriteString(vec.elements[i].String())
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func newVector(size int, init *LAny) *LAny {
	elements := make([]*LAny, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	vec := new(LAny)
	vec.ltype = typeVector
	vec.elements = elements
	return vec
}

func vector(elements ...*LAny) *LAny {
	return vectorFromElements(elements, len(elements))
}

func vectorFromElements(elements []*LAny, count int) *LAny {
	el := make([]*LAny, count)
	copy(el, elements[0:count])
	vec := new(LAny)
	vec.ltype = typeVector
	vec.elements = el
	return vec
}

func copyVector(vec *LAny) *LAny {
	size := len(vec.elements)
	elements := make([]*LAny, size)
	copy(elements, vec.elements)
	return vectorFromElements(elements, size)
}

func vectorSet(vec *LAny, idx int, obj *LAny) error {
	if isVector(vec) {
		if idx < 0 || idx >= len(vec.elements) {
			return Error("Vector index out of range")
		}
		vec.elements[idx] = obj
		return nil
	}
	return TypeError(typeVector, vec)
}

func vectorRef(vec *LAny, idx int) (*LAny, error) {
	if isVector(vec) {
		if idx < 0 || idx >= len(vec.elements) {
			return nil, Error("Vector index out of range")
		}
		return vec.elements[idx], nil
	}
	return nil, TypeError(typeVector, vec)
}

func toVector(obj *LAny) (*LAny, error) {
	switch obj.ltype {
	case typeVector:
		return obj, nil
	case typeList:
		return listToVector(obj), nil
	case typeStruct:
		return structToVector(obj), nil
	case typeString:
		return stringToVector(obj), nil
	}
	return nil, Error("Cannot convert ", obj.ltype, " to <vector>")
}

/*
func flattenVector(vec *LVector) *LVector {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.car
		var fitem *LList
		lstItem, ok := item.(*LList)
		if ok {
			fitem = flatten(lstItem)
		} else {
			fitem = list(item)
		}
		if tail == EmptyList {
			result = fitem
			tail = result
		} else {
			tail.cdr = fitem
		}
		for tail.cdr != EmptyList {
			tail = tail.cdr
		}
		lst = lst.cdr
	}
	return result
}
*/
