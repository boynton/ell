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

func cons(car *LOB, cdr *LOB) *LOB {
	if true { //for dev. No external code should call into this, internal code should always be correct
		if car == nil {
			panic("Assertion failure: don't call cons with nil as car")
		}
		if cdr == nil {
			panic("Assertion failure: don't call cons with nil as cdr")
		}
		if !isList(cdr) {
			panic("Assertion failure: don't call cons with non-list as cdr")
		}
	}
	if inExec {
		conses++
	}
	result := newLOB(typeList)
	result.car = car
	result.cdr = cdr
	return result
}

func safeCar(lst *LOB) (*LOB, error) {
	if !isList(lst) {
		return nil, ArgTypeError("list", 1, lst)
	}
	return car(lst), nil
}

func car(lst *LOB) *LOB {
	if lst == EmptyList {
		return Null
	}
	return lst.car
}

func setCar(lst *LOB, obj *LOB) error {
	if !isList(lst) {
		return ArgTypeError("list", 1, lst)
	}
	if isEmpty(lst) {
		return Error("argument 1 to set-car! must be a non-empty list: ", lst)
	}
	lst.car = obj
	return nil
}

func safeCdr(lst *LOB) (*LOB, error) {
	if !isList(lst) {
		return nil, ArgTypeError("list", 1, lst)
	}
	return car(lst), nil
}

func cdr(lst *LOB) *LOB {
	if lst == EmptyList {
		return lst
	}
	return lst.cdr
}

func setCdr(lst *LOB, obj *LOB) error {
	if !isList(lst) {
		return ArgTypeError("list", 1, lst)
	}
	if isEmpty(lst) {
		return Error("argument 1 to set-cdr! must be a non-empty list: ", lst)
	}
	if !isList(obj) {
		return Error("argument 2 to set-cdr! must be a list: ", lst)
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
func cddr(lst *LOB) *LOB {
	return cdr(cdr(lst))
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
	return &LOB{variant: typeList} //car and cdr are both nil
}

// Equal returns true if the object is equal to the argument
func listEqual(lst *LOB, a *LOB) bool {
	for lst != EmptyList {
		if a == EmptyList {
			return false
		}
		if !equal(lst.car, a.car) {
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
	switch obj.variant {
	case typeList:
		return obj, nil
	case typeVector:
		return listFromValues(obj.elements), nil
	case typeStruct:
		return structToList(obj)
	case typeString:
		return stringToList(obj), nil
	}
	return nil, Error("Cannot convert ", obj.variant, " to <list>")
}
