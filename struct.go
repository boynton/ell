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
	"bytes"
)

// Key - the key type for Structs. The string value and Ell type string are combined, so we can extract
// the type later when enumerating keys. Map keys cannot LOBs, they are not "comparable" in golang.
type structKey struct {
	keyValue string
	keyType  string
}

func newStructKey(v *LOB) structKey {
	return structKey{v.text, v.Type.text}
}

func (k structKey) toLOB() *LOB {
	if k.keyType == "<string>" {
		return String(k.keyValue)
	}
	return Intern(k.keyValue)
}

// IsValidStructKey - return true of the object is a valid <struct> key.
func IsValidStructKey(o *LOB) bool {
	switch o.Type {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

// EmptyStruct - a <struct> with no bindings
var EmptyStruct = MakeStruct(0)

// MakeStruct - create an empty <struct> object with the specified capacity
func MakeStruct(capacity int) *LOB {
	strct := new(LOB)
	strct.Type = StructType
	strct.bindings = make(map[structKey]*LOB, capacity)
	return strct
}

// Struct - create a new <struct> object from the arguments, which can be other structs, or key/value pairs
func Struct(fieldvals []*LOB) (*LOB, error) {
	strct := new(LOB)
	strct.Type = StructType
	strct.bindings = make(map[structKey]*LOB)
	count := len(fieldvals)
	i := 0
	var bindings map[structKey]*LOB
	for i < count {
		o := Value(fieldvals[i])
		i++
		switch o.Type {
		case StructType: // not a valid key, just copy bindings from it
			if bindings == nil {
				bindings = make(map[structKey]*LOB, len(o.bindings))
			}
			for k, v := range o.bindings {
				bindings[k] = v
			}
		case StringType, SymbolType, KeywordType, TypeType:
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in arglist: ", o)
			}
			if bindings == nil {
				bindings = make(map[structKey]*LOB)
			}
			bindings[newStructKey(o)] = fieldvals[i]
			i++
		default:
			return nil, Error(ArgumentErrorKey, "Bad struct key: ", o)
		}
	}
	if bindings == nil {
		strct.bindings = make(map[structKey]*LOB)
	} else {
		strct.bindings = bindings
	}
	return strct, nil
}

// StructLength - return the length (field count) of the <struct> object
func StructLength(strct *LOB) int {
	return len(strct.bindings)
}

// Get - return the value for the key of the object. The Value() function is first called to
// handle typed instances of <struct>.
// This is called by the VM, when a keyword is used as a function.
func Get(obj *LOB, key *LOB) (*LOB, error) {
	s := Value(obj)
	if s.Type != StructType {
		return nil, Error(ArgumentErrorKey, "get expected a <struct> argument, got a ", obj.Type)
	}
	return structGet(s, key), nil
}

func structGet(s *LOB, key *LOB) *LOB {
	switch key.Type {
	case KeywordType, SymbolType, TypeType, StringType:
		k := newStructKey(key)
		result, ok := s.bindings[k]
		if ok {
			return result
		}
	}
	return Null
}

func Has(obj *LOB, key *LOB) (bool, error) {
	tmp, err := Get(obj, key)
	if err != nil || tmp == Null {
		return false, err
	}
	return true, nil
}

func Put(obj *LOB, key *LOB, val *LOB) {
	k := newStructKey(key)
	obj.bindings[k] = val
}

func Unput(obj *LOB, key *LOB) {
	k := newStructKey(key)
	delete(obj.bindings, k)
}

func sliceContains(slice []*LOB, obj *LOB) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func sliceGet(bindings []*LOB, key *LOB) *LOB {
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if key == bindings[i] {
			return bindings[i+1]
		}
	}
	return Null
}

func slicePut(bindings []*LOB, key *LOB, val *LOB) []*LOB {
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if key == bindings[i] {
			bindings[i+1] = val
			return bindings
		}
	}
	return append(append(bindings, key), val)
}

func validateKeywordArgList(args *LOB, keys []*LOB) (*LOB, error) {
	tmp, err := validateKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return listFromValues(tmp), nil
}

func validateKeywordArgs(args *LOB, keys []*LOB) (*LOB, error) {
	tmp, err := validateKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return Struct(tmp)
}

func validateKeywordArgBindings(args *LOB, keys []*LOB) ([]*LOB, error) {
	count := ListLength(args)
	bindings := make([]*LOB, 0, count)
	for args != EmptyList {
		key := Car(args)
		switch key.Type {
		case SymbolType:
			key = Intern(key.text + ":")
			fallthrough
		case KeywordType:
			if !sliceContains(keys, key) {
				return nil, Error(ArgumentErrorKey, key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(ArgumentErrorKey, key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, Car(args))
		case StructType:
			for k, v := range key.bindings {
				sym := Intern(k.keyValue)
				if sliceContains(keys, sym) {
					bindings = slicePut(bindings, sym, v)
				}
			}
		default:
			return nil, Error(ArgumentErrorKey, "Not a keyword: ", key)
		}
		args = args.cdr
	}
	return bindings, nil
}

// Equal returns true if the object is equal to the argument
func StructEqual(s1 *LOB, s2 *LOB) bool {
	bindings1 := s1.bindings
	size := len(bindings1)
	bindings2 := s2.bindings
	if size == len(bindings2) {
		for k, v := range bindings1 {
			v2, ok := bindings2[k]
			if !ok {
				return false
			}
			if !Equal(v, v2) {
				return false
			}
		}
		return true
	}
	return false
}

func structToString(s *LOB) string {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	for k, v := range s.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(k.keyValue)
		buf.WriteString(" ")
		buf.WriteString(v.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func structToList(s *LOB) (*LOB, error) {
	result := EmptyList
	tail := EmptyList
	for k, v := range s.bindings {
		tmp := List(k.toLOB(), v)
		if result == EmptyList {
			result = List(tmp)
			tail = result
		} else {
			tail.cdr = List(tmp)
			tail = tail.cdr
		}
	}
	return result, nil
}

func structToVector(s *LOB) *LOB {
	size := len(s.bindings)
	el := make([]*LOB, size)
	j := 0
	for k, v := range s.bindings {
		el[j] = Vector(k.toLOB(), v)
		j++
	}
	return VectorFromElements(el, size)
}

func structKeyList(s *LOB) *LOB {
	result := EmptyList
	tail := EmptyList
	for k := range s.bindings {
		key := k.toLOB()
		if result == EmptyList {
			result = List(key)
			tail = result
		} else {
			tail.cdr = List(key)
			tail = tail.cdr
		}
	}
	return result
}

func structValueList(s *LOB) *LOB {
	result := EmptyList
	tail := EmptyList
	for _, v := range s.bindings {
		if result == EmptyList {
			result = List(v)
			tail = result
		} else {
			tail.cdr = List(v)
			tail = tail.cdr
		}
	}
	return result
}

func listToStruct(lst *LOB) (*LOB, error) {
	strct := new(LOB)
	strct.Type = StructType
	strct.bindings = make(map[structKey]*LOB)
	for lst != EmptyList {
		k := lst.car
		lst = lst.cdr
		switch k.Type {
		case ListType:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			Put(strct, k.car, k.cdr.car)
		case VectorType:
			elements := k.elements
			n := len(elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", elements[0])
			}
			Put(strct, elements[0], elements[1])
		default:
			if !IsValidStructKey(k) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if lst == EmptyList {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in list: ", k)
			}
			Put(strct, k, lst.car)
			lst = lst.cdr
		}
	}
	return strct, nil
}

func vectorToStruct(vec *LOB) (*LOB, error) {
	count := len(vec.elements)
	strct := new(LOB)
	strct.Type = StructType
	strct.bindings = make(map[structKey]*LOB, count)
	i := 0
	for i < count {
		k := vec.elements[i]
		i++
		switch k.Type {
		case ListType:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			Put(strct, k.car, k.cdr.car)
		case VectorType:
			elements := k.elements
			n := len(elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.elements[0])
			}
			Put(strct, elements[0], elements[1])
		case StringType, SymbolType, KeywordType, TypeType:
		default:
			if !IsValidStructKey(k) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in vector: ", k)
			}
			Put(strct, k, vec.elements[i])
			i++
		}
	}
	return strct, nil
}

func ToStruct(obj *LOB) (*LOB, error) {
	val := Value(obj)
	switch val.Type {
	case StructType:
		return val, nil
	case ListType:
		return listToStruct(val)
	case VectorType:
		return vectorToStruct(val)
	}
	return nil, Error(ArgumentErrorKey, "to-struct cannot accept argument of type ", obj.Type)
}
