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

func vectorEqual(v1 *LOB, v2 *LOB) bool {
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

func vectorToString(vec *LOB) string {
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

func newVector(size int, init *LOB) *LOB {
	elements := make([]*LOB, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return vectorFromElementsNoCopy(elements)
}

func vector(elements ...*LOB) *LOB {
	return vectorFromElements(elements, len(elements))
}

func vectorFromElements(elements []*LOB, count int) *LOB {
	el := make([]*LOB, count)
	copy(el, elements[0:count])
	return vectorFromElementsNoCopy(el)
}

func vectorFromElementsNoCopy(elements []*LOB) *LOB {
	vec := newLOB(typeVector)
	vec.elements = elements
	return vec
}

func copyVector(vec *LOB) *LOB {
	return vectorFromElements(vec.elements, len(vec.elements))
}

func vectorSet(vec *LOB, idx int, obj *LOB) error {
	if isVector(vec) {
		if idx < 0 || idx >= len(vec.elements) {
			return Error("Vector index out of range")
		}
		vec.elements[idx] = obj
		return nil
	}
	return TypeError(typeVector, vec)
}

func vectorRef(vec *LOB, idx int) (*LOB, error) {
	if isVector(vec) {
		if idx < 0 || idx >= len(vec.elements) {
			return nil, Error("Vector index out of range")
		}
		return vec.elements[idx], nil
	}
	return nil, TypeError(typeVector, vec)
}

func toVector(obj *LOB) (*LOB, error) {
	switch obj.variant {
	case typeVector:
		return obj, nil
	case typeList:
		return listToVector(obj), nil
	case typeStruct:
		return structToVector(obj), nil
	case typeString:
		return stringToVector(obj), nil
	}
	return nil, Error("Cannot convert ", obj.variant, " to <vector>")
}
