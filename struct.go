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

//
// LStruct - Ell structs (objects). They are extensible, having a special type symbol in them.
//
type LStruct struct { // <struct>
	bindings map[LAny]LAny
}

var typeStruct = intern("<struct>")

// Type returns the type of the object
func (s *LStruct) Type() LAny {
	return typeStruct
}

// Value returns the object itself for primitive types
func (s *LStruct) Value() LAny {
	return s
}

func sliceContains(slice []LAny, obj LAny) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func normalizeKeywordArgs(args *LList, keys []LAny) (*LList, error) {
	count := length(args)
	bindings := make(map[LAny]LAny, count/2)
	for args != EmptyList {
		key := car(args)
		switch t := key.Value().(type) {
		case *LSymbol:
			if !isKeyword(key) {
				key = intern(t.String() + ":")
			}
			if !sliceContains(keys, key) {
				return nil, Error(key, " bad keyword parameter")
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(key, " mismatched keyword/value pair in parameter")
			}
			bindings[key] = car(args)
		case *LStruct:
			for k, v := range t.bindings {
				if sliceContains(keys, k) {
					bindings[k] = v
				}
			}
		}
		args = args.cdr
	}
	count = len(bindings)
	if count == 0 {
		return EmptyList, nil
	}
	lst := make([]LAny, 0, count*2)
	for k, v := range bindings {
		lst = append(lst, k)
		lst = append(lst, v)
	}
	return listFromValues(lst), nil
}

func copyStruct(s *LStruct) *LStruct {
	bindings := make(map[LAny]LAny, len(s.bindings))
	for k, v := range s.bindings {
		bindings[k] = v
	}
	return &LStruct{bindings}
}

func (s *LStruct) Copy() LAny {
	//deep copy
	bindings := make(map[LAny]LAny, len(s.bindings))
	for k, v := range s.bindings {
		bindings[k] = v.Copy()
	}
	return &LStruct{bindings}
}

func newStruct(fieldvals []LAny) (*LStruct, error) {
	count := len(fieldvals)
	i := 0
	bindings := make(map[LAny]LAny, count/2) //optimal if all key/value pairs
	for i < count {
		o := fieldvals[i]
		i++
		switch t := o.Value().(type) {
		case LNull:
			//ignore
		case LString:
			if i == count {
				return nil, Error("mismatched keyword/value in arglist: ", o)
			}
			bindings[o] = fieldvals[i]
			i++
		case *LSymbol:
			if i == count {
				return nil, Error("mismatched keyword/value in arglist: ", o)
			}
			bindings[o] = fieldvals[i]
			i++
		case *LStruct:
			for k, v := range t.bindings {
				bindings[k] = v
			}
		default:
			return nil, Error("bad parameter to struct: ", o)
		}
	}
	return &LStruct{bindings}, nil
}

func isStruct(obj LAny) bool {
	_, ok := obj.(*LStruct)
	return ok
}

// Equal returns true if the object is equal to the argument
func (s *LStruct) Equal(another LAny) bool {
	if a, ok := another.(*LStruct); ok {
		slen := len(s.bindings)
		if slen == len(a.bindings) {
			for k, v := range s.bindings {
				if v2, ok := a.bindings[k]; ok {
					if !equal(v, v2) {
						return false
					}
				} else {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (s *LStruct) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	for k, v := range s.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(k.String())
		buf.WriteString(" ")
		buf.WriteString(v.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func has(obj LAny, key LAny) (bool, error) {
	o := obj.Value()
	if s, ok := o.(*LStruct); ok {
		_, ok := s.bindings[key]
		return ok, nil
	}
	return false, TypeError(typeStruct, obj)
}

func get(obj LAny, key LAny) (LAny, error) {
	o := obj.Value()
	if s, ok := o.(*LStruct); ok {
		if val, ok := s.bindings[key]; ok {
			return val, nil
		}
		return Null, nil
	}
	return nil, TypeError(typeStruct, obj)
}

func put(obj LAny, key LAny, value LAny) (LAny, error) {
	if aStruct, ok := obj.(*LStruct); ok {
		aStruct.bindings[key] = value
		return aStruct, nil
	}
	return nil, TypeError(typeStruct, obj)
}

func structToList(aStruct *LStruct) (*LList, error) {
	result := EmptyList
	tail := EmptyList
	for k, v := range aStruct.bindings {
		tmp := list(k, v)
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

func structToVector(aStruct *LStruct) *LVector {
	size := len(aStruct.bindings)
	el := make([]LAny, size)
	var i int
	for k, v := range aStruct.bindings {
		el[i] = vector(k, v)
		i++
	}
	return &LVector{el}
}
