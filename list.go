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

type LList struct { // <list>
	car LAny
	cdr *LList
}

var typeList = intern("<list>")

var symList = intern("list")
var symQuote = intern("quote")
var symQuasiquote = intern("quasiquote")
var symUnquote = intern("unquote")
var symUnquoteSplicing = intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *LList {
	var lst LList
	return &lst
}

func isEmpty(col LAny) bool {
	switch v := col.(type) {
	case LNull: //Do I really want this?
		return true
	case LString:
		return len(v) == 0
	case *LVector:
		return len(v.elements) == 0
	case *LList:
		return v == EmptyList
	case *LStruct:
		return len(v.bindings) == 0
	default:
		return false
	}
}

func isList(obj LAny) bool {
	_, ok := obj.(*LList)
	return ok
}

// Type returns the type of the object
func (*LList) Type() LAny {
	return typeList
}

// Value returns the object itself for primitive types
func (lst *LList) Value() LAny {
	return lst
}

// Equal returns true if the object is equal to the argument
func (lst *LList) Equal(another LAny) bool {
	if a, ok := another.(*LList); ok {
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
	}
	return false
}

func (lst *LList) String() string {
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

func (lst *LList) Copy() LAny {
	//deep copy
	if lst == EmptyList {
		return lst
	}
	return cons(lst.car.Copy(), lst.cdr.Copy().(*LList))
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

func newList(count int, val LAny) *LList {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func cons(car LAny, cdr *LList) *LList {
	if car == nil {
		panic("Assertion failure: don't call cons with nil as car")
	}
	if cdr == nil {
		panic("Assertion failure: don't call cons with nil as cdr")
	}
	if inExec {
		conses++
	}
	return &LList{car, cdr}
}

func safeCar(lst LAny) (LAny, error) {
	switch p := lst.(type) {
	case *LList:
		if p == EmptyList {
			return Null, nil
		}
		return p.car, nil
	default:
		return nil, ArgTypeError("list", 1, lst)
	}
}

func car(lst LAny) LAny {
	if lst != EmptyList {
		if p, ok := lst.(*LList); ok {
			return p.car
		}
	}
	return Null
}

func setCar(lst LAny, obj LAny) error {
	switch p := lst.(type) {
	case *LList:
		if p == EmptyList {
			return Error("argument to set-car! must be a non-empty list: ", lst)
		}
		p.car = obj
		return nil
	default:
		return ArgTypeError("list", 1, lst)
	}
}

func safeCdr(lst LAny) (*LList, error) {
	switch p := lst.(type) {
	case *LList:
		if lst != EmptyList {
			return p.cdr, nil
		}
		return EmptyList, nil
	default:
		return nil, ArgTypeError("list", 1, lst)
	}
}

func cdr(lst LAny) *LList {
	if lst != EmptyList {
		if p, ok := lst.(*LList); ok {
			return p.cdr
		}
	}
	return EmptyList
}

func setCdr(lst LAny, obj LAny) error {
	if p, ok := lst.(*LList); ok {
		if p != EmptyList {
			if d, ok := obj.(*LList); ok {
				p.cdr = d
				return nil
			}
			return Error("argument 2 to set-cdr! must be a list: ", lst)
		}
	}
	return Error("argument 1 to set-cdr! must be a non-empty list: ", lst)
}

func caar(lst LAny) LAny {
	return car(car(lst))
}
func cadr(lst LAny) LAny {
	return car(cdr(lst))
}
func cddr(lst LAny) *LList {
	return cdr(cdr(lst))
}
func caddr(lst LAny) LAny {
	return car(cdr(cdr(lst)))
}
func cdddr(lst LAny) *LList {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst LAny) LAny {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst LAny) *LList {
	return cdr(cdr(cdr(cdr(lst))))
}

func listFromValues(values []LAny) *LList {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = cons(v, p)
	}
	return p
}

func list(values ...LAny) *LList {
	return listFromValues(values)
}

func listToVector(lst *LList) *LVector {
	var elems []LAny
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return &LVector{elems}
}

func vectorToList(ary LAny) (LAny, error) {
	v, ok := ary.(*LVector)
	if !ok {
		return nil, TypeError(typeVector, ary)
	}
	return listFromValues(v.elements), nil
}

func toList(obj LAny) (LAny, error) {
	switch t := obj.(type) {
	case *LList:
		return t, nil
	case *LVector:
		return listFromValues(t.elements), nil
	case *LStruct:
		return structToList(t)
	case LString:
		return stringToList(t), nil
	}
	return nil, Error("Cannot convert ", obj.Type(), " to <list>")
}

