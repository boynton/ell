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

func newStruct(fieldvals []*LOB) (*LOB, error) {
	count := len(fieldvals)
	strct := newLOB(typeStruct)
	strct.elements = make([]*LOB, 0, count)
	i := 0
	for i < count {
		o := fieldvals[i]
		i++
		switch value(o).variant {
		case typeStruct: // not a valid key, just copy bindings from it
			jmax := len(o.elements) / 2
			for j := 0; j < jmax; j += 2 {
				put(strct, o.elements[j], o.elements[j+1])
			}
		case typeString, typeSymbol, typeKeyword, typeType:
			if i == count {
				return nil, Error("mismatched keyword/value in arglist: ", o)
			}
			put(strct, o, fieldvals[i])
			i++
		default:
			return nil, Error("bad parameter to struct: ", o)
		}
	}
	return strct, nil
}

func get(obj *LOB, key *LOB) (*LOB, error) {
	s := value(obj)
	if s.variant != typeStruct {
		return nil, TypeError(typeStruct, obj)
	}
	bindings := s.elements
	slen := len(bindings)
	switch key.variant {
	case typeKeyword, typeSymbol, typeType: //these are all intern'ed, so pointer equality works
		for i := 0; i < slen; i += 2 {
			if bindings[i] == key {
				return bindings[i+1], nil
			}
		}
	case typeString:
		for i := 0; i < slen; i += 2 {
			if bindings[i].variant == typeString && bindings[i].text == key.text {
				return bindings[i+1], nil
			}
		}
	}
	return Null, nil
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

func put(obj *LOB, key *LOB, val *LOB) (*LOB, error) {
	//danger! side effects!
	s := value(obj)
	if s.variant != typeStruct {
		return nil, TypeError(typeStruct, obj)
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
			if bindings[i].variant == typeString && bindings[i].text == key.text {
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

// (normalize-keyword-args '(x: 23) x: y:) => (x: 23)
// (normalize-keyword-args '(x: 23 z: 100) x: y:) =>  *** z: bad keyword parameter. Allowed keys: [x: y:] 
func normalizeKeywordArgList(args *LOB, keys []*LOB) (*LOB, error) {
	tmp, err := normalizeKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return listFromValues(tmp), nil
}

func normalizeKeywordArgs(args *LOB, keys []*LOB) (*LOB, error) {
	tmp, err := normalizeKeywordArgBindings(args, keys)
	if err != nil {
		return nil, err
	}
	return newStruct(tmp)
}

func normalizeKeywordArgBindings(args *LOB, keys []*LOB) ([]*LOB, error) {
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
			jmax := len(key.elements)
			for j := 0; j < jmax; j += 2 {
				k := key.elements[j]
				if sliceContains(keys, k) {
					bindings = slicePut(bindings, k, key.elements[j+1])
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
			if bindings[i].variant == typeString && bindings[i].text == key.text {
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
	if size == len(s2.elements) {
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
