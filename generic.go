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

func methodSignature(formalArgs *LList) (LOB, error) {
	sig := ""
	for formalArgs != EmptyList {
		s := formalArgs.car //might be a symbol, might be a list
		tname := ""
		if lst, ok := s.(*LList); ok { //specialized
			t := cadr(lst)
			if ty, ok := t.(*LType); ok {
				tname = ty.text
			} else {
				return nil, Error(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
			}
		} else if _, ok := s.(*LSymbol); ok { //unspecialized
			tname = "<any>"
		} else {
			return nil, Error(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
		}
		sig += tname
		formalArgs = formalArgs.cdr
	}
	return intern(sig), nil
}

func arglistSignature(args []LOB) LOB {
	sig := ""
	for _, arg := range args {
		sig += arg.Type().String()
	}
	return intern(sig)
}

var symGenfns = internSymbol("*genfns*")
var keyMethods = intern("methods:")

func getfn(sym *LSymbol, sig LOB) (LOB, error) {
	gfs := global(symGenfns)
	if gfs != nil {
		if strct, ok := gfs.(*LStruct); ok {
			gf := strct.get(sym)
			if gf == Null {
				return nil, Error(ErrorKey, "Not a generic function: ", sym)
			}
			if strct2, ok := value(gf).(*LStruct); ok {
				methods := strct2.get(keyMethods)
				if strct3, ok := methods.(*LStruct); ok {
					fun := strct3.get(sig)
					if fun != Null {
						return fun, nil
					}
				}
			}
		}
	}
	return nil, Error(ErrorKey, "Generic function ", sym, ", has no matching method for: ", sig)
}

func isEmpty(obj LOB) bool {
	seq := value(obj)
	switch t := seq.(type) {
	case *LList:
		return t == EmptyList
	case *LString:
		return len(t.value) == 0
	case *LVector:
		return len(t.elements) == 0
	case *LStruct:
		return len(t.elements) == 0
	case *LNull: //?
		return true
	default:
		return false
	}
}

func length(obj LOB) int {
	seq := value(obj)
	switch t := seq.(type) {
	case *LString:
		return stringLength(t.value)
	case *LVector:
		return len(t.elements)
	case *LList:
		return listLength(t)
	case *LStruct:
		return len(t.elements) / 2
	default:
		return -1
	}
}

func reverse(lst *LList) *LList {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func flatten(lst *LList) *LList {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		var item *LList
		switch t := lst.car.(type) {
		case *LList:
			item = flatten(t)
		case *LVector:
			litem, _ := toList(t)
			item = flatten(litem)
		default:
			item = list(item)
		}
		if tail == EmptyList {
			result = item
			tail = result
		} else {
			tail.cdr = item
		}
		for tail.cdr != EmptyList {
			tail = tail.cdr
		}
		lst = lst.cdr
	}
	return result
}

func concat(seq1 *LList, seq2 *LList) (*LList, error) {
	rev := reverse(seq1)
	if rev == EmptyList {
		return seq2, nil
	}
	lst := seq2
	for rev != EmptyList {
		lst = cons(rev.car, lst)
		rev = rev.cdr
	}
	return lst, nil
}

func nextRangeValue(val int, end int, step int) (int, bool) {
	val += step
	return val, isValidRange(val, end, step)
}

func isValidRange(val int, end int, step int) bool {
	if step >= 0 {
		if val < end {
			return true
		}
	} else {
		if val > end {
			return true
		}
	}
	return false
}

func assoc(obj LOB, rest ...LOB) (LOB, error) {
	s := value(obj)
	switch t := s.(type) {
	case *LStruct:
		return assocStruct(t, rest)
	case *LVector:
		return assocVector(t, rest)
	default:
		return nil, Error(ArgumentErrorKey, "Expected a <struct>|<vector>, got a ", obj.Type())
	}
}

func dissoc(obj LOB, rest ...LOB) (LOB, error) {
	s := value(obj)
	switch t := s.(type) {
	case *LStruct:
		return dissocStruct(t, rest)
	default:
		return nil, Error(ArgumentErrorKey, "Expected a <struct>|<vector>, got a ", obj.Type())
	}
}

func dissocBang(obj LOB, rest ...LOB) (LOB, error) {
	s := value(obj)
	switch t := s.(type) {
	case *LStruct:
		return dissocBangStruct(t, rest)
	default:
		return nil, Error(ArgumentErrorKey, "Expected a <struct>|<vector>, got a ", obj.Type())
	}
}
