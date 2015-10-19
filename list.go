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

package ell

import (
	"bytes"
)

func cons(car *LOB, cdr *LOB) *LOB {
	result := newLOB(ListType)
	result.car = car
	result.cdr = cdr
	return result
}

func car(lst *LOB) *LOB {
	if lst == EmptyList {
		return Null
	}
	return lst.car
}

func setCar(lst *LOB, obj *LOB) error {
	if lst == EmptyList {
		return Error(ArgumentErrorKey, "set-car! expected a non-empty <list>")
	}
	lst.car = obj
	return nil
}

func cdr(lst *LOB) *LOB {
	if lst == EmptyList {
		return lst
	}
	return lst.cdr
}

func setCdr(lst *LOB, obj *LOB) error {
	if lst == EmptyList {
		return Error(ArgumentErrorKey, "set-cdr! expected a non-empty <list>")
	}
	lst.cdr = obj
	return nil
}

func caar(lst *LOB) *LOB {
	return car(car(lst))
}
func cadr(lst *LOB) *LOB {
	return car(cdr(lst))
}
func cdar(lst *LOB) *LOB {
	return car(cdr(lst))
}
func cddr(lst *LOB) *LOB {
	return cdr(cdr(lst))
}
func cadar(lst *LOB) *LOB {
	return car(cdr(car(lst)))
}
func caddr(lst *LOB) *LOB {
	return car(cdr(cdr(lst)))
}
func cdddr(lst *LOB) *LOB {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst *LOB) *LOB {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst *LOB) *LOB {
	return cdr(cdr(cdr(cdr(lst))))
}

var symList = intern("list")
var symQuote = intern("quote")
var symQuasiquote = intern("quasiquote")
var symUnquote = intern("unquote")
var symUnquoteSplicing = intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *LOB {
	return &LOB{Type: ListType} //car and cdr are both nil
}

// Equal returns true if the object is equal to the argument
func listEqual(lst *LOB, a *LOB) bool {
	for lst != EmptyList {
		if a == EmptyList {
			return false
		}
		if !Equal(lst.car, a.car) {
			return false
		}
		lst = lst.cdr
		a = a.cdr
	}
	if lst == a {
		return true
	}
	return false
}

func listToString(lst *LOB) string {
	var buf bytes.Buffer
	if lst != EmptyList && lst.cdr != EmptyList && cddr(lst) == EmptyList {
		if lst.car == symQuote {
			buf.WriteString("'")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		} else if lst.car == symQuasiquote {
			buf.WriteString("`")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		} else if lst.car == symUnquote {
			buf.WriteString("~")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		} else if lst.car == symUnquoteSplicing {
			buf.WriteString("~")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		}
	}
	buf.WriteString("(")
	delim := ""
	for lst != EmptyList {
		buf.WriteString(delim)
		delim = " "
		buf.WriteString(lst.car.String())
		lst = lst.cdr
	}
	buf.WriteString(")")
	return buf.String()
}

func listLength(lst *LOB) int {
	if lst == EmptyList {
		return 0
	}
	count := 1
	o := lst.cdr
	for o != EmptyList {
		count++
		o = o.cdr
	}
	return count
}

func newList(count int, val *LOB) *LOB {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func listFromValues(values []*LOB) *LOB {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = cons(v, p)
	}
	return p
}

func List(values ...*LOB) *LOB {
	return listFromValues(values)
}
func list(values ...*LOB) *LOB {
	return listFromValues(values)
}

func listToVector(lst *LOB) *LOB {
	var elems []*LOB
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return vectorFromElementsNoCopy(elems)
}

func toList(obj *LOB) (*LOB, error) {
	switch obj.Type {
	case ListType:
		return obj, nil
	case VectorType:
		return listFromValues(obj.elements), nil
	case StructType:
		return structToList(obj)
	case StringType:
		return stringToList(obj), nil
	}
	return nil, Error(ArgumentErrorKey, "to-list cannot accept ", obj.Type)
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
		switch item.Type {
		case ListType:
			item = flatten(item)
		case VectorType:
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
