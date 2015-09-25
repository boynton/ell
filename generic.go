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

func isEmpty(obj *LAny) bool {
	seq := value(obj)
	switch seq.ltype {
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

func length(obj *LAny) int {
	seq := value(obj)
	switch seq.ltype {
	case typeString:
		return len(seq.text)
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

/*
func assoc(seq *LAny, key *LAny, val *LAny) (*LAny, error) {
	switch seq.ltype {
	case typeStruct:
		s := copyStruct(seq)
		s.bindings[key] = val
		return s, nil
	case typeVector:
		if isNumber(key) {
			a := copyVector(seq)
			a.elements[int(seq.ival)] = val
			return a, nil
		}
		return nil, TypeError(typeNumber, key)
	default:
		return nil, Error("Cannot assoc with this value: ", seq)
	}
}

func dissoc(seq *LAny, key *LAny) (*LAny, error) {
	switch seq.ltype {
	case typeStruct:
		s := copyStruct(seq)
		delete(s.bindings, key)
		return s, nil
	default:
		return nil, Error("Cannot dissoc with this value: ", seq)
	}
}
*/

func reverse(lst *LAny) *LAny {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func flatten(lst *LAny) *LAny {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.car
		switch item.ltype {
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

func concat(seq1 *LAny, seq2 *LAny) (*LAny, error) {
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
