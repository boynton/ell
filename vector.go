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

type LVector struct { // <vector>
	elements []LAny
}

var typeVector = intern("<vector>")

func isVector(obj LAny) bool {
	_, ok := obj.(*LVector)
	return ok
}

// Type returns the type of the object
func (*LVector) Type() LAny {
	return typeVector
}

// Value returns the object itself for primitive types
func (ary *LVector) Value() LAny {
	return ary
}

// Equal returns true if the object is equal to the argument
func (ary *LVector) Equal(another LAny) bool {
	if a, ok := another.(*LVector); ok {
		alen := len(ary.elements)
		if alen == len(a.elements) {
			for i := 0; i < alen; i++ {
				if !equal(ary.elements[i], a.elements[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (ary *LVector) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	count := len(ary.elements)
	if count > 0 {
		buf.WriteString(ary.elements[0].String())
		for i := 1; i < count; i++ {
			buf.WriteString(" ")
			buf.WriteString(ary.elements[i].String())
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func (vec *LVector) Copy() LAny {
	size := len(vec.elements)
	elements := make([]LAny, size)
	for i := 0; i < size; i++ {
		elements[i] = vec.elements[i].Copy()
	}
	return &LVector{elements}
}

func newVector(size int, init LAny) *LVector {
	elements := make([]LAny, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return &LVector{elements}
}

func vector(elements ...LAny) LAny {
	return toVector(elements, len(elements))
}

func toVector(elements []LAny, count int) LAny {
	el := make([]LAny, count)
	copy(el, elements[0:count])
	return &LVector{el}
}

func copyVector(a *LVector) *LVector {
	elements := make([]LAny, len(a.elements))
	copy(elements, a.elements)
	return &LVector{elements}
}

func vectorLength(ary LAny) (int, error) {
	if a, ok := ary.(*LVector); ok {
		return len(a.elements), nil
	}
	return 1, TypeError(typeVector, ary)
}

func vectorSet(ary LAny, idx int, obj LAny) error {
	if a, ok := ary.(*LVector); ok {
		if idx < 0 || idx >= len(a.elements) {
			return Error("Vector index out of range")
		}
		a.elements[idx] = obj
		return nil
	}
	return TypeError(typeVector, ary)
}

func vectorRef(ary LAny, idx int) (LAny, error) {
	if a, ok := ary.(*LVector); ok {
		if idx < 0 || idx >= len(a.elements) {
			return nil, Error("Vector index out of range")
		}
		return a.elements[idx], nil
	}
	return nil, TypeError(typeVector, ary)
}

