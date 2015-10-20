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
)

// Cons - create a new list consisting of the first object and the rest of the list
func Cons(car *LOB, cdr *LOB) *LOB {
	result := new(LOB)
	result.Type = ListType
	result.car = car
	result.cdr = cdr
	return result
}

// Car - return the first object in a list
func Car(lst *LOB) *LOB {
	if lst == EmptyList {
		return Null
	}
	return lst.car
}

// Cdr - return the rest of the list
func Cdr(lst *LOB) *LOB {
	if lst == EmptyList {
		return lst
	}
	return lst.cdr
}

// Caar - return the Car of the Car of the list
func Caar(lst *LOB) *LOB {
	return Car(Car(lst))
}

// Cadr - return the Car of the Cdr of the list
func Cadr(lst *LOB) *LOB {
	return Car(Cdr(lst))
}

// Cdar - return the Cdr of the Car of the list
func Cdar(lst *LOB) *LOB {
	return Car(Cdr(lst))
}

// Cddr - return the Cdr of the Cdr of the list
func Cddr(lst *LOB) *LOB {
	return Cdr(Cdr(lst))
}

// Cadar - return the Car of the Cdr of the Car of the list
func Cadar(lst *LOB) *LOB {
	return Car(Cdr(Car(lst)))
}

// Caddr - return the Car of the Cdr of the Cdr of the list
func Caddr(lst *LOB) *LOB {
	return Car(Cdr(Cdr(lst)))
}

// Cdddr - return the Cdr of the Cdr of the Cdr of the list
func Cdddr(lst *LOB) *LOB {
	return Cdr(Cdr(Cdr(lst)))
}

// Cadddr - return the Car of the Cdr of the Cdr of the Cdr of the list
func Cadddr(lst *LOB) *LOB {
	return Car(Cdr(Cdr(Cdr(lst))))
}

// Cddddr - return the Cdr of the Cdr of the Cdr of the Cdr of the list
func Cddddr(lst *LOB) *LOB {
	return Cdr(Cdr(Cdr(Cdr(lst))))
}

var ListSymbol = Intern("list")
var QuoteSymbol = Intern("quote")
var QuasiquoteSymbol = Intern("quasiquote")
var UnquoteSymbol = Intern("unquote")
var UnquoteSymbolSplicing = Intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *LOB {
	return &LOB{Type: ListType} //car and cdr are both nil
}

// ListEqual returns true if the object is equal to the argument
func ListEqual(lst *LOB, a *LOB) bool {
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
	if lst != EmptyList && lst.cdr != EmptyList && Cddr(lst) == EmptyList {
		if lst.car == QuoteSymbol {
			buf.WriteString("'")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		} else if lst.car == QuasiquoteSymbol {
			buf.WriteString("`")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		} else if lst.car == UnquoteSymbol {
			buf.WriteString("~")
			buf.WriteString(Cadr(lst).String())
			return buf.String()
		} else if lst.car == UnquoteSymbolSplicing {
			buf.WriteString("~")
			buf.WriteString(Cadr(lst).String())
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

func ListLength(lst *LOB) int {
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

func MakeList(count int, val *LOB) *LOB {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = Cons(val, result)
	}
	return result
}

func listFromValues(values []*LOB) *LOB {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = Cons(v, p)
	}
	return p
}

func List(values ...*LOB) *LOB {
	return listFromValues(values)
}

func listToVector(lst *LOB) *LOB {
	var elems []*LOB
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return VectorFromElementsNoCopy(elems)
}

// ToList - conver the argument to a List, if possible
func ToList(obj *LOB) (*LOB, error) {
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

func Reverse(lst *LOB) *LOB {
	rev := EmptyList
	for lst != EmptyList {
		rev = Cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func Flatten(lst *LOB) *LOB {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.car
		switch item.Type {
		case ListType:
			item = Flatten(item)
		case VectorType:
			litem, _ := ToList(item)
			item = Flatten(litem)
		default:
			item = List(item)
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

func Concat(seq1 *LOB, seq2 *LOB) (*LOB, error) {
	rev := Reverse(seq1)
	if rev == EmptyList {
		return seq2, nil
	}
	lst := seq2
	for rev != EmptyList {
		lst = Cons(rev.car, lst)
		rev = rev.cdr
	}
	return lst, nil
}
