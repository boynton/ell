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

// LVector - the concrete type for vectors
type LVector struct {
	elements []LOB
}

// VectorType - the Type object for vectors
var VectorType = intern("<vector>")

// Type returns the type of the object
func (*LVector) Type() LOB {
	return VectorType
}

// Value returns the object itself for primitive types
func (vec *LVector) Value() LOB {
	return vec
}

// Equal returns true if the object is equal to the argument
func (vec *LVector) Equal(another LOB) bool {
	vec2, ok := another.(*LVector)
	if ok {
		return vectorEqual(vec, vec2)
	}
	return false
}

// String returns the string representation of the object
func (vec *LVector) String() string {
	return vectorToString(vec)
}

func isVector(obj LOB) bool {
	return obj.Type() == VectorType
}

func vectorEqual(v1 *LVector, v2 *LVector) bool {
	count := len(v1.elements)
	if count != len(v2.elements) {
		return false
	}
	for i := 0; i < count; i++ {
		if !isEqual(v1.elements[i], v2.elements[i]) {
			return false
		}
	}
	return true
}

func vectorToString(vec *LVector) string {
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

func newVector(size int, init LOB) *LVector {
	elements := make([]LOB, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return vectorFromElementsNoCopy(elements)
}

func vector(elements ...LOB) *LVector {
	return vectorFromElements(elements, len(elements))
}

func vectorFromElements(elements []LOB, count int) *LVector {
	el := make([]LOB, count)
	copy(el, elements[0:count])
	return vectorFromElementsNoCopy(el)
}

func vectorFromElementsNoCopy(elements []LOB) *LVector {
	return &LVector{elements}
}

func copyVector(vec *LVector) *LVector {
	return vectorFromElements(vec.elements, len(vec.elements))
}

func vectorSet(vec *LVector, idx int, obj LOB) error {
	if idx < 0 || idx >= len(vec.elements) {
		return Error(ArgumentErrorKey, "vector-set! index out of range: ", idx)
	}
	vec.elements[idx] = obj
	return nil
}

func vectorRef(vec *LVector, idx int) LOB {
	if idx < 0 || idx >= len(vec.elements) {
		return Null
	}
	return vec.elements[idx]
}

func assocBangVector(vec *LVector, fieldvals []LOB) (*LVector, error) {
	//danger! mutation!
	max := len(vec.elements)
	count := len(fieldvals)
	i := 0
	for i < count {
		key := value(fieldvals[i])
		i++
		switch k := key.(type) {
		case *LNumber:
			idx := int(k.value)
			if idx < 0 || idx > max {
				return nil, Error(ArgumentErrorKey, "assoc! index out of range: ", idx)
			}
			if i == count {
				return nil, Error(ArgumentErrorKey, "assoc! mismatched index/value: ", fieldvals)
			}
			vec.elements[idx] = fieldvals[i]
			i++
		default:
			return nil, Error(ArgumentErrorKey, "assoc! bad vector index: ", key)
		}
	}
	return vec, nil
}

func assocVector(vec *LVector, fieldvals []LOB) (*LVector, error) {
	//optimize this
	return assocBangVector(copyVector(vec), fieldvals)
}

func toVector(obj LOB) (*LVector, error) {
	switch t := obj.(type) {
	case *LVector:
		return t, nil
	case *LList:
		return listToVector(t), nil
	case *LStruct:
		return structToVector(t), nil
	case *LString:
		return stringToVector(t), nil
	}
	return nil, Error(ArgumentErrorKey, "to-vector expected <vector>, <list>, <struct>, or <string>, got a ", obj.Type())
}
