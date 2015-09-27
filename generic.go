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

func isEmpty(obj *LOB) bool {
	seq := value(obj)
	switch seq.variant {
	case typeList:
		return seq == EmptyList
	case typeString:
		return len(seq.text) == 0
	case typeVector:
		return len(seq.elements) == 0
	case typeStruct:
		return len(seq.elements) == 0
	case typeNull: //?
		return true
	default:
		return false
	}
}

func length(obj *LOB) int {
	seq := value(obj)
	switch seq.variant {
	case typeString:
		return stringLength(seq.text)
	case typeVector:
		return len(seq.elements)
	case typeList:
		return listLength(seq)
	case typeStruct:
		return len(seq.elements) / 2
	default:
		return -1
	}
}

func reverse(lst *LOB) *LOB {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func flatten(lst *LOB) *LOB {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.car
		switch item.variant {
		case typeList:
			item = flatten(item)
		case typeVector:
			litem, _ := toList(item)
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

func concat(seq1 *LOB, seq2 *LOB) (*LOB, error) {
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

func assoc(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return assocStruct(s, rest)
	case typeVector:
		return assocVector(s, rest)
	default:
		return nil, Error("assoc cannot work with type ", obj.variant)
	}
}

func assocBang(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return assocBangStruct(s, rest)
	case typeVector:
		return assocBangVector(s, rest)
	default:
		return nil, Error("assoc! cannot work with type ", obj.variant)
	}
}

func dissoc(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return dissocStruct(s, rest)
	default:
		return nil, Error("dissoc cannot work with type ", obj.variant)
	}
}

func dissocBang(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return dissocBangStruct(s, rest)
	default:
		return nil, Error("dissoc! cannot work with type ", obj.variant)
	}
}
