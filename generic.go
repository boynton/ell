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

func length(seq LAny) int {
	switch v := seq.Value().(type) {
	case LString:
		return len(v)
	case *LVector:
		return len(v.elements)
	case *LList:
		return listLength(v)
	case *LStruct:
		return len(v.bindings)
	default:
		return -1
	}
}


func assoc(seq LAny, key LAny, val LAny) (LAny, error) {
	switch s := seq.(type) {
	case *LStruct:
		s2 := copyStruct(s)
		s2.bindings[key] = val
		return s2, nil
	case *LVector:
		if idx, ok := key.(LNumber); ok {
			a := copyVector(s)
			a.elements[int(idx)] = val
			return a, nil
		}
		return nil, TypeError(typeNumber, key)
	default:
		return nil, Error("Cannot assoc with this value: ", seq)
	}
}

func dissoc(seq LAny, key LAny) (LAny, error) {
	switch s := seq.(type) {
	case *LStruct:
		s2 := copyStruct(s)
		delete(s2.bindings, key)
		return s2, nil
	default:
		return nil, Error("Cannot dissoc with this value: ", seq)
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
		item := lst.car
		var fitem *LList
		switch titem := item.(type) {
		case *LList:
			fitem = flatten(titem)
		case *LVector:
			litem, _ := toList(titem)
			fitem = flatten(litem.(*LList))
		default:
			fitem = list(item)
		}
		if tail == EmptyList {
			result = fitem
			tail = result
		} else {
			tail.cdr = fitem
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

