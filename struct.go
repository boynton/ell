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
	. "github.com/boynton/ell/data" // -> "github.com/boynton/data"
)
/*
// Key - the key type for Structs. The string value and Ell type string are combined, so we can extract
// the type later when enumerating keys. Map keys cannot Objects, they are not "comparable" in golang.
type structKey struct {
	keyValue string
	keyType  string
}

func newStructKey(v *Object) structKey {
	return structKey{v.text, v.Type.text}
}

func (k structKey) toObject() *Object {
	if k.keyType == "<string>" {
		return String(k.keyValue)
	}
	return Intern(k.keyValue)
}

// IsValidStructKey - return true of the object is a valid <struct> key.
func IsValidStructKey(o *Object) bool {
	switch o.Type {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

// EmptyStruct - a <struct> with no bindings
var EmptyStruct = MakeStruct(0)

// MakeStruct - create an empty <struct> object with the specified capacity
func MakeStruct(capacity int) *Object {
	strct := new(Object)
	strct.Type = StructType
	strct.bindings = make(map[structKey]*Object, capacity)
	return strct
}

// Struct - create a new <struct> object from the arguments, which can be other structs, or key/value pairs
func Struct(fieldvals []*Object) (*Object, error) {
	strct := new(Object)
	strct.Type = StructType
	strct.bindings = make(map[structKey]*Object)
	count := len(fieldvals)
	i := 0
	var bindings map[structKey]*Object
	for i < count {
		o := Value(fieldvals[i])
		i++
		switch o.Type {
		case StructType: // not a valid key, just copy bindings from it
			if bindings == nil {
				bindings = make(map[structKey]*Object, len(o.bindings))
			}
			for k, v := range o.bindings {
				bindings[k] = v
			}
		case StringType, SymbolType, KeywordType, TypeType:
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in arglist: ", o)
			}
			if bindings == nil {
				bindings = make(map[structKey]*Object)
			}
			bindings[newStructKey(o)] = fieldvals[i]
			i++
		default:
			return nil, Error(ArgumentErrorKey, "Bad struct key: ", o)
		}
	}
	if bindings == nil {
		strct.bindings = make(map[structKey]*Object)
	} else {
		strct.bindings = bindings
	}
	return strct, nil
}
*/

// StructLength - return the length (field count) of the <struct> object
func StructLength(strct *Struct) int {
	return len(strct.Bindings)
}

// Get - return the value for the key of the object. The Value() function is first called to
// handle typed instances of <struct>.
// This is called by the VM, when a keyword is used as a function.
func Get(obj Value, key Value) (Value, error) {
	if pi, ok := obj.(*Instance); ok {
		obj = pi.Value
	}
	if p, ok := obj.(*Struct); ok {
		return p.Get(key), nil
	}
	return nil, NewError(ArgumentErrorKey, "Expected a <struct> argument, got a ", obj.Type())
}

func Has(obj Value, key Value) (bool, error) {
	tmp, err := Get(obj, key)
	if err != nil || tmp == Null {
		return false, err
	}
	return true, nil
}

func Put(obj Value, key Value, val Value) error {
	if pi, ok := obj.(*Instance); ok {
		obj = pi.Value
	}
	if p, ok := obj.(*Struct); ok {
		p.Put(key, val)
		return nil
	}
    return NewError(ArgumentErrorKey, "Expected a <struct> argument, got a ", obj.Type())
}

func Unput(obj Value, key Value) error {
	if pi, ok := obj.(*Instance); ok {
		obj = pi.Value
	}
	if p, ok := obj.(*Struct); ok {
		p.Unput(key)
		return nil
	}
    return NewError(ArgumentErrorKey, "Expected a <struct> argument, got a ", obj.Type())
}

func sliceContains(slice []Value, obj Value) bool {
	for _, o := range slice {
		if o == obj { //FIX: Equal() ?!
			return true
		}
	}
	return false
}

func slicePut(bindings []Value, key Value, val Value) []Value {
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if key == bindings[i] {
			bindings[i+1] = val
			return bindings
		}
	}
	return append(append(bindings, key), val)
}

func validateKeywordArgList(args *List, keys []Value) (Value, error) {
	tmp, err := validateKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return ListFromValues(tmp), nil
}

func validateKeywordArgBindings(args *List, keys []Value) ([]Value, error) {
	count := ListLength(args)
	bindings := make([]Value, 0, count)
	for args != EmptyList {
		key := Car(args)
		switch p := key.(type) {
		case *Symbol:
			key = Intern(p.Text + ":")
			if !sliceContains(keys, key) {
				return nil, NewError(ArgumentErrorKey, key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.Cdr
			if args == EmptyList {
				return nil, NewError(ArgumentErrorKey, key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, Car(args))
		case *Keyword:
			if !sliceContains(keys, key) {
				return nil, NewError(ArgumentErrorKey, key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.Cdr
			if args == EmptyList {
				return nil, NewError(ArgumentErrorKey, key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, Car(args))
		case *Struct:
			for k, v := range p.Bindings {
				sym := Intern(k.Value)
				if sliceContains(keys, sym) {
					bindings = slicePut(bindings, sym, v)
				}
			}
		default:
			return nil, NewError(ArgumentErrorKey, "Not a keyword: ", key)
		}
		args = args.Cdr
	}
	return bindings, nil
}

// Equal returns true if the object is equal to the argument
func StructEqual(s1 *Struct, s2 *Struct) bool {
	bindings1 := s1.Bindings
	size := len(bindings1)
	bindings2 := s2.Bindings
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

func structToString(s *Struct) string {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	for k, v := range s.Bindings {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(k.Value)
		buf.WriteString(" ")
		buf.WriteString(v.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func StructToList(s *Struct) (*List, error) {
	result := EmptyList
	tail := EmptyList
	for k, v := range s.Bindings {
		tmp := NewList(k.ToValue(), v)
		if result == EmptyList {
			result = NewList(tmp)
			tail = result
		} else {
			tail.Cdr = NewList(tmp)
			tail = tail.Cdr
		}
	}
	return result, nil
}

func StructToVector(s *Struct) *Vector {
	size := len(s.Bindings)
	el := make([]Value, size)
	j := 0
	for k, v := range s.Bindings {
		el[j] = NewVector(k.ToValue(), v)
		j++
	}
	return VectorFromElements(el, size)
}

func StructKeys(s Value) Value {
	if ss, ok := s.(*Struct); ok {
		return structKeyList(ss)
	}
	return EmptyList
}

func StructValues(s Value) Value {
	if ss, ok := s.(*Struct); ok {
		return structValueList(ss)
	}
	return EmptyList
}

func structKeyList(s *Struct) *List {
	result := EmptyList
	tail := EmptyList
	for k := range s.Bindings {
		key := k.ToValue()
		if result == EmptyList {
			result = NewList(key)
			tail = result
		} else {
			tail.Cdr = NewList(key)
			tail = tail.Cdr
		}
	}
	return result
}

func structValueList(s *Struct) *List {
	result := EmptyList
	tail := EmptyList
	for _, v := range s.Bindings {
		if result == EmptyList {
			result = NewList(v)
			tail = result
		} else {
			tail.Cdr = NewList(v)
			tail = tail.Cdr
		}
	}
	return result
}

func listToStruct(lst *List) (Value, error) {
	strct := NewStruct()
	//	strct.bindings = make(map[structKey]Value)
	for lst != EmptyList {
		k := lst.Car
		lst = lst.Cdr
		switch p := k.(type) {
		case *List:
			if EmptyList == p || EmptyList == p.Cdr || EmptyList != p.Cdr.Cdr {
				return nil, NewError(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(p.Car) {
				return nil, NewError(ArgumentErrorKey, "Bad struct key: ", p.Car)
			}
			strct.Put(p.Car, p.Cdr.Car)
		case *Vector:
			elements := p.Elements
			n := len(elements)
			if n != 2 {
				return nil, NewError(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(elements[0]) {
				return nil, NewError(ArgumentErrorKey, "Bad struct key: ", elements[0])
			}
			strct.Put(elements[0], elements[1])
		default:
			if !IsValidStructKey(k) {
				return nil, NewError(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if lst == EmptyList {
				return nil, NewError(ArgumentErrorKey, "Mismatched keyword/value in list: ", k)
			}
			Put(strct, k, lst.Car)
			lst = lst.Cdr
		}
	}
	return strct, nil
}

func vectorToStruct(vec *Vector) (Value, error) {
	count := len(vec.Elements)
	strct := NewStruct()
	i := 0
	for i < count {
		k := vec.Elements[i]
		i++
		switch p := k.(type) {
		case *List:
			if EmptyList == p || EmptyList == p.Cdr || EmptyList != p.Cdr.Cdr {
				return nil, NewError(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(p.Car) {
				return nil, NewError(ArgumentErrorKey, "Bad struct key: ", p.Car)
			}
			strct.Put(p.Car, p.Cdr.Car)
		case *Vector:
			elements := p.Elements
			n := len(elements)
			if n != 2 {
				return nil, NewError(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !IsValidStructKey(elements[0]) {
				return nil, NewError(ArgumentErrorKey, "Bad struct key: ", elements[0])
			}
			strct.Put(elements[0], elements[1])
		case *String, *Symbol, *Keyword, *Type:
		default:
			if !IsValidStructKey(k) {
				return nil, NewError(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if i == count {
				return nil, NewError(ArgumentErrorKey, "Mismatched keyword/value in vector: ", k)
			}
			Put(strct, k, vec.Elements[i])
			i++
		}
	}
	return strct, nil
}

func ToStruct(obj Value) (Value, error) {
	val := Value(obj)
	switch p := val.(type) {
	case *Struct:
		return p, nil
	case *List:
		return listToStruct(p)
	case *Vector:
		return vectorToStruct(p)
	}
	return nil, NewError(ArgumentErrorKey, "to-struct cannot accept argument of type ", obj.Type())
}

func IsStruct(obj Value) bool {
	if _, ok := obj.(*Struct); ok {
		return true
	}
	return false
}
