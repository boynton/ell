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

type LStruct struct {
	elements []LOB
}

var StructType = intern("<struct>")

// Type returns the type of the object
func (*LStruct) Type() LOB {
	return StructType
}

// Value returns the object itself for primitive types
func (strct *LStruct) Value() LOB {
	return strct
}

// Equal returns true if the object is equal to the argument
func (strct *LStruct) Equal(another LOB) bool {
	strct2, ok := another.(*LStruct)
	if ok {
		return structEqual(strct, strct2)
	}
	return false
}

// String returns the string representation of the object
func (strct *LStruct) String() string {
	return structToString(strct)
}

func isStruct(obj LOB) bool {
	_, ok := obj.(*LStruct)
	return ok
}

func isValidStructKey(o LOB) bool {
	switch o.Type() {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

func newStruct(fieldvals []LOB) (*LStruct, error) {
	count := len(fieldvals)
	strct := &LStruct{make([]LOB, 0, count)}
	return initStruct(strct, fieldvals, count)
}

func initStruct(strct *LStruct, fieldvals []LOB, count int) (*LStruct, error) {
	i := 0
	for i < count {
		o := value(fieldvals[i])
		i++
		switch t := value(o).(type) {
		case *LStruct: // not a valid key, just copy bindings from it
			jmax := len(t.elements) / 2
			for j := 0; j < jmax; j += 2 {
				put(strct, t.elements[j], t.elements[j+1])
			}
		case *LString, *LSymbol, *LKeyword, *LType:
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in arglist: ", o)
			}
			put(strct, o, fieldvals[i])
			i++
		default:
			return nil, Error(ArgumentErrorKey, "Bad struct key: ", o)
		}
	}
	return strct, nil
}

//called by the VM, when a keyword is used as a function. Optimize!
func get(obj LOB, key LOB) (LOB, error) {
	s := value(obj)
	strct, ok := s.(*LStruct)
	if !ok {
		return nil, Error(ArgumentErrorKey, "get expected a <struct> argument, got a ", obj.Type())
	}
	return strct.get(key), nil
}

func (strct *LStruct) get(key LOB) LOB {
	bindings := strct.elements
	slen := len(bindings)
	switch key.(type) {
	case *LKeyword, *LSymbol, *LType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				return bindings[i+1]
			}
		}
	case *LString:
		for i := 0; i < slen; i += 2 {
			if isEqual(key, bindings[i]) {
				return bindings[i+1]
			}
		}
	}
	return Null
}

func (strct *LStruct) has(key LOB) bool {
	tmp := strct.get(key)
	if tmp == Null {
		return false
	}
	return true
}

func assocStruct(s *LStruct, rest []LOB) (*LStruct, error) {
	//optimize this
	return newStruct(append(rest, s))
}

func assocBangStruct(s *LStruct, rest []LOB) (*LStruct, error) {
	//optimize this
	return initStruct(s, rest, len(rest))
}

func dissocStruct(s *LStruct, rest []LOB) (*LStruct, error) {
	return nil, Error(ErrorKey, "dissocStruct: NYI")
}
func dissocBangStruct(s *LStruct, rest []LOB) (*LStruct, error) {
	return nil, Error(ErrorKey, "dissocStruct: NYI")
}

func put(obj LOB, key LOB, val LOB) (LOB, error) {
	//danger! side effects!
	s := value(obj)
	strct, ok := s.(*LStruct)
	if !ok {
		return nil, Error(ArgumentErrorKey, "put expected a <struct> for argument 1, got a ", obj.Type())
	}
	if !isValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	return strct.put(key, val), nil
}

func (strct *LStruct) put(key LOB, val LOB) LOB {
	bindings := strct.elements
	slen := len(bindings)
	switch key.(type) {
	case *LKeyword, *LSymbol, *LType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				bindings[i+1] = val
				return strct
			}
		}
	case *LString:
		for i := 0; i < slen; i += 2 {
			if isEqual(bindings[i], key) {
				bindings[i+1] = val
				return strct
			}
		}
	}
	strct.elements = append(append(bindings, key), val)
	return strct
}

func sliceContains(slice []LOB, obj LOB) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func sliceGet(bindings []LOB, key LOB) LOB {
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if key == bindings[i] {
			return bindings[i+1]
		}
	}
	return Null
}

func slicePut(bindings []LOB, key LOB, val LOB) []LOB {
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if key == bindings[i] {
			bindings[i+1] = val
			return bindings
		}
	}
	return append(append(bindings, key), val)
}

func validateKeywordArgList(args *LList, keys []LOB) (LOB, error) {
	tmp, err := validateKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return listFromValues(tmp), nil
}

func validateKeywordArgs(args *LList, keys []LOB) (LOB, error) {
	tmp, err := validateKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return newStruct(tmp)
}

func validateKeywordArgBindings(args *LList, keys []LOB) ([]LOB, error) {
	count := length(args)
	bindings := make([]LOB, 0, count)
	for args != EmptyList {
		switch t := car(args).(type) {
		case *LSymbol:
			key := internKeyword(t.text + ":")
			if !sliceContains(keys, key) {
				return nil, Error(key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, car(args))
		case *LKeyword:
			key := t
			if !sliceContains(keys, key) {
				return nil, Error(key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, car(args))
		case *LStruct:
			jmax := len(t.elements)
			for j := 0; j < jmax; j += 2 {
				k := t.elements[j]
				if sliceContains(keys, k) {
					bindings = slicePut(bindings, k, t.elements[j+1])
				}
			}
		}
		args = args.cdr
	}
	return bindings, nil
}

func structAssoc(bindings []LOB, key LOB, val LOB) []LOB {
	slen := len(bindings)
	switch k := key.(type) {
	case *LKeyword, *LSymbol, *LType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				bindings[i+1] = val
				return bindings
			}
		}
	case *LString:
		for i := 0; i < slen; i += 2 {
			if isEqual(bindings[i], k) {
				bindings[i+1] = val
				return bindings
			}
		}
	}
	return append(append(bindings, key), val)
}

// Equal returns true if the object is equal to the argument
func structEqual(s1 *LStruct, s2 *LStruct) bool {
	bindings1 := s1.elements
	size := len(bindings1)
	if size == len(s2.elements) {
		for i := 0; i < size; i += 2 {
			k := bindings1[i]
			v := bindings1[i+1]
			v2, err := get(s2, k)
			if err != nil {
				return false
			}
			if v2 != Null {
				if !isEqual(v, v2) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}

func structToString(s *LStruct) string {
	var buf bytes.Buffer
	buf.WriteString("{")
	bindings := s.elements
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(bindings[i].String())
		buf.WriteString(" ")
		buf.WriteString(bindings[i+1].String())
	}
	buf.WriteString("}")
	return buf.String()
}

func structToList(s *LStruct) (*LList, error) {
	result := EmptyList
	tail := EmptyList
	bindings := s.elements
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		tmp := list(bindings[i], bindings[i+1])
		if result == EmptyList {
			result = list(tmp)
			tail = result
		} else {
			tail.cdr = list(tmp)
			tail = tail.cdr
		}
	}
	return result, nil
}

func structToVector(s *LStruct) *LVector {
	bindings := s.elements
	size := len(bindings)
	el := make([]LOB, size/2)
	var j int
	for i := 0; i < size; i += 2 {
		el[j] = vector(bindings[i], bindings[i+1])
		j++
	}
	return vectorFromElements(el, size)
}

func structKeyList(s *LStruct) *LList {
	result := EmptyList
	tail := EmptyList
	bindings := s.elements
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		tmp := bindings[i]
		if result == EmptyList {
			result = list(tmp)
			tail = result
		} else {
			tail.cdr = list(tmp)
			tail = tail.cdr
		}
	}
	return result
}

func structValueList(s *LStruct) *LList {
	result := EmptyList
	tail := EmptyList
	bindings := s.elements
	size := len(bindings)
	for i := 0; i < size; i += 2 {
		tmp := bindings[i+1]
		if result == EmptyList {
			result = list(tmp)
			tail = result
		} else {
			tail.cdr = list(tmp)
			tail = tail.cdr
		}
	}
	return result
}

func listToStruct(lst *LList) (*LStruct, error) {
	strct := &LStruct{make([]LOB, 0)}
	for lst != EmptyList {
		lst = lst.cdr
		switch k := lst.car.(type) {
		case *LList:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			put(strct, k.car, k.cdr.car)
		case *LVector:
			n := len(k.elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(k.elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.elements[0])
			}
			put(strct, k.elements[0], k.elements[1])
		default:
			if !isValidStructKey(k) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if lst == EmptyList {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in list: ", k)
			}
			put(strct, k, lst.car)
			lst = lst.cdr
		}
	}
	return strct, nil
}

func vectorToStruct(vec *LVector) (*LStruct, error) {
	count := len(vec.elements)
	strct := &LStruct{make([]LOB, 0, count)}
	i := 0
	for i < count {
		i++
		switch k := vec.elements[i].(type) {
		case *LList:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			put(strct, k.car, k.cdr.car)
		case *LVector:
			n := len(k.elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(k.elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.elements[0])
			}
			put(strct, k.elements[0], k.elements[1])
		case *LString, *LSymbol, *LKeyword, *LType:
			//do nothing
		default:
			if !isValidStructKey(k) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k)
			}
			if i == count {
				return nil, Error(ArgumentErrorKey, "Mismatched keyword/value in vector: ", k)
			}
			put(strct, k, vec.elements[i])
			i++
		}
	}
	return strct, nil
}

func toStruct(obj LOB) (*LStruct, error) {
	val := value(obj)
	switch t := val.(type) {
	case *LStruct:
		return t, nil
	case *LList:
		return listToStruct(t)
	case *LVector:
		return vectorToStruct(t)
	}
	return nil, Error(ArgumentErrorKey, "to-struct cannot accept argument of type ", obj.Type())
}
