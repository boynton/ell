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

func isValidStructKey(o *LOB) bool {
	switch o.variant {
	case typeString, typeSymbol, typeKeyword, typeType:
		return true
	}
	return false
}

func newStruct(fieldvals []*LOB) (*LOB, error) {
	count := len(fieldvals)
	strct := newLOB(typeStruct)
	strct.elements = make([]*LOB, 0, count)
	return initStruct(strct, fieldvals, count)
}

func initStruct(strct *LOB, fieldvals []*LOB, count int) (*LOB, error) {
	i := 0
	for i < count {
		o := value(fieldvals[i])
		i++
		switch value(o).variant {
		case typeStruct: // not a valid key, just copy bindings from it
			elements := o.elements
			jmax := len(elements) / 2
			for j := 0; j < jmax; j += 2 {
				put(strct, elements[j], elements[j+1])
			}
		case typeString, typeSymbol, typeKeyword, typeType:
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
func get(obj *LOB, key *LOB) (*LOB, error) {
	s := value(obj)
	if s.variant != typeStruct {
		return nil, Error(ArgumentErrorKey, "get expected a <struct> argument, got a ", obj.variant)
	}
	return structGet(s, key), nil
}

func structGet(s *LOB, key *LOB) *LOB {
	bindings := s.elements
	slen := len(bindings)
	switch key.variant {
	case typeKeyword, typeSymbol, typeType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				return bindings[i+1]
			}
		}
	case typeString:
		for i := 0; i < slen; i += 2 {
			if equal(bindings[i], key) {
				return bindings[i+1]
			}
		}
	}
	return Null
}

func has(obj *LOB, key *LOB) (bool, error) {
	tmp, err := get(obj, key)
	if err != nil {
		return false, err
	}
	if tmp == Null {
		return false, nil
	}
	return true, nil
}

func assocStruct(s *LOB, rest []*LOB) (*LOB, error) {
	//optimize this
	return newStruct(append(rest, s))
}

func assocBangStruct(s *LOB, rest []*LOB) (*LOB, error) {
	//optimize this
	return initStruct(s, rest, len(rest))
}

func dissocStruct(s *LOB, rest []*LOB) (*LOB, error) {
	return nil, Error(ErrorKey, "dissocStruct: NYI")
}
func dissocBangStruct(s *LOB, rest []*LOB) (*LOB, error) {
	return nil, Error(ErrorKey, "dissocStruct: NYI")
}

func put(obj *LOB, key *LOB, val *LOB) (*LOB, error) {
	//danger! side effects!
	s := value(obj)
	if s.variant != typeStruct {
		return nil, Error(ArgumentErrorKey, "put expected a <struct> for argument 1, got a ", s.variant)
	}
	if !isValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	bindings := s.elements
	slen := len(bindings)
	switch key.variant {
	case typeKeyword, typeSymbol, typeType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				bindings[i+1] = val
				return obj, nil
			}
		}
	case typeString:
		for i := 0; i < slen; i += 2 {
			if equal(bindings[i], key) {
				bindings[i+1] = val
				return obj, nil
			}
		}
	}
	s.elements = append(append(bindings, key), val)
	return obj, nil
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
	return newStruct(tmp)
}

func validateKeywordArgBindings(args *LOB, keys []*LOB) ([]*LOB, error) {
	count := length(args)
	bindings := make([]*LOB, 0, count)
	for args != EmptyList {
		key := car(args)
		switch value(key).variant {
		case typeSymbol:
			key = intern(key.text + ":")
			fallthrough
		case typeKeyword:
			if !sliceContains(keys, key) {
				return nil, Error(key, " bad keyword parameter. Allowed keys: ", keys)
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(key, " mismatched keyword/value pair in parameter")
			}
			bindings = slicePut(bindings, key, car(args))
		case typeStruct:
			elements := key.elements
			jmax := len(elements)
			for j := 0; j < jmax; j += 2 {
				k := elements[j]
				if sliceContains(keys, k) {
					bindings = slicePut(bindings, k, elements[j+1])
				}
			}
		}
		args = args.cdr
	}
	return bindings, nil
}

func structAssoc(bindings []*LOB, key *LOB, val *LOB) []*LOB {
	slen := len(bindings)
	switch key.variant {
	case typeKeyword, typeSymbol, typeType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				bindings[i+1] = val
				return bindings
			}
		}
	case typeString:
		for i := 0; i < slen; i += 2 {
			if equal(bindings[i], key) {
				bindings[i+1] = val
				return bindings
			}
		}
	}
	return append(append(bindings, key), val)
}

// Equal returns true if the object is equal to the argument
func structEqual(s1 *LOB, s2 *LOB) bool {
	bindings1 := s1.elements
	size := len(bindings1)
	bindings2 := s2.elements
	if size == len(bindings2) {
		for i := 0; i < size; i += 2 {
			k := bindings1[i]
			v := bindings1[i+1]
			v2, err := get(s2, k)
			if err != nil {
				return false
			}
			if v2 != Null {
				if !equal(v, v2) {
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

func structToString(s *LOB) string {
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

func structToList(s *LOB) (*LOB, error) {
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

func structToVector(s *LOB) *LOB {
	bindings := s.elements
	size := len(bindings)
	el := make([]*LOB, size/2)
	var j int
	for i := 0; i < size; i += 2 {
		el[j] = vector(bindings[i], bindings[i+1])
		j++
	}
	return vectorFromElements(el, size)
}

func structKeyList(s *LOB) *LOB {
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

func structValueList(s *LOB) *LOB {
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

func listToStruct(lst *LOB) (*LOB, error) {
	strct := newLOB(typeStruct)
	strct.elements = make([]*LOB, 0)
	for lst != EmptyList {
		k := lst.car
		lst = lst.cdr
		switch k.variant {
		case typeList:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			put(strct, k.car, k.cdr.car)
		case typeVector:
			elements := k.elements
			n := len(elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", elements[0])
			}
			put(strct, elements[0], elements[1])
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

func vectorToStruct(vec *LOB) (*LOB, error) {
	count := len(vec.elements)
	strct := newLOB(typeStruct)
	strct.elements = make([]*LOB, 0, count)
	i := 0
	for i < count {
		k := vec.elements[i]
		i++
		switch k.variant {
		case typeList:
			if EmptyList == k || EmptyList == k.cdr || EmptyList != k.cdr.cdr {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(k.car) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.car)
			}
			put(strct, k.car, k.cdr.car)
		case typeVector:
			elements := k.elements
			n := len(elements)
			if n != 2 {
				return nil, Error(ArgumentErrorKey, "Bad struct binding: ", k)
			}
			if !isValidStructKey(elements[0]) {
				return nil, Error(ArgumentErrorKey, "Bad struct key: ", k.elements[0])
			}
			put(strct, elements[0], elements[1])
		case typeString, typeSymbol, typeKeyword, typeType:
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

func toStruct(obj *LOB) (*LOB, error) {
	val := value(obj)
	switch val.variant {
	case typeStruct:
		return val, nil
	case typeList:
		return listToStruct(val)
	case typeVector:
		return vectorToStruct(val)
	}
	return nil, Error(ArgumentErrorKey, "to-struct cannot accept argument of type ", obj.variant)
}
