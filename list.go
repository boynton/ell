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

type LList struct {
	car LOB
	cdr *LList
}

var ListType = intern("<list>")

// Type returns the type of the object
func (*LList) Type() LOB {
	return ListType
}

// Value returns the object itself for primitive types
func (lst *LList) Value() LOB {
	return lst
}

// Equal returns true if the object is equal to the argument
func (lst *LList) Equal(another LOB) bool {
	lst2, ok := another.(*LList)
	if ok {
		return listEqual(lst, lst2)
	}
	return false
}

// String returns the string representation of the object
func (lst *LList) String() string {
	return listToString(lst)
}


func isList(obj LOB) bool {
	return obj.Type() == ListType
}

func cons(car LOB, cdr *LList) *LList {
	if true { //for dev. No external code should call into this, internal code should always be correct
		if car == nil {
			panic("Assertion failure: don't call cons with nil as car")
		}
		if cdr == nil {
			panic("Assertion failure: don't call cons with nil as cdr")
		}
	}
	if inExec {
		conses++
	}
	return &LList{car, cdr}
}

func car(lst *LList) LOB {
	if lst == EmptyList {
		return Null
	}
	return lst.car
}

func setCar(lst *LList, obj LOB) error {
	if lst == EmptyList {
		return Error(ArgumentErrorKey, "set-car! expected a non-empty <list>")
	}
	lst.car = obj
	return nil
}

func cdr(lst *LList) *LList {
	if lst == EmptyList {
		return lst
	}
	return lst.cdr
}

func setCdr(lst *LList, obj *LList) error {
	if isEmpty(lst) {
		return Error(ArgumentErrorKey, "set-cdr! expected a non-empty <list>")
	}
	lst.cdr = obj
	return nil
}

func caar(lst *LList) LOB {
	if lst != EmptyList {
		if tmp, ok := lst.car.(*LList); ok {
			return tmp.car
		}
	}
	return Null
}
func cadr(lst *LList) LOB {
	return car(cdr(lst))
}
func cdar(lst *LList) *LList {
	if lst != EmptyList {
		if tmp, ok := lst.car.(*LList); ok {
			return tmp.cdr
		}
	}
	return EmptyList
}
func cddr(lst *LList) *LList {
	return cdr(cdr(lst))
}
//func cadar(lst *LList) *LList {
//	return car(cdr(car(lst)))
//}
func caddr(lst *LList) LOB {
	return car(cdr(cdr(lst)))
}
func cdddr(lst *LList) *LList {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst *LList) LOB {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst *LList) *LList {
	return cdr(cdr(cdr(cdr(lst))))
}

var symList = intern("list")
var symQuote = intern("quote")
var symQuasiquote = intern("quasiquote")
var symUnquote = intern("unquote")
var symUnquoteSplicing = intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *LList {
	return &LList{nil, nil}
}

// Equal returns true if the object is equal to the argument
func listEqual(lst *LList, a *LList) bool {
	for lst != EmptyList {
		if a == EmptyList {
			return false
		}
		if !isEqual(lst.car, a.car) {
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

func listToString(lst *LList) string {
	var buf bytes.Buffer
	if lst == nil {
		panic("nil list!")
	}
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

func listLength(lst *LList) int {
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

func newList(count int, val LOB) *LList {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func listFromValues(values []LOB) *LList {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = cons(v, p)
	}
	return p
}

func list(values ...LOB) *LList {
	return listFromValues(values)
}

func listToVector(lst *LList) *LVector {
	var elems []LOB
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return vectorFromElementsNoCopy(elems)
}

func toList(obj LOB) (*LList, error) {
	switch t := obj.(type) {
	case *LList:
		return t, nil
	case *LVector:
		return listFromValues(t.elements), nil
	case *LStruct:
		return structToList(t)
	case *LString:
		return stringToList(t), nil
	}
	return nil, Error(ArgumentErrorKey, "to-list cannot accept ", obj.Type())
}
