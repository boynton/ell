/*
Copyright 2021 Lee Boynton

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
package data

import(
	"bytes"
)

type List struct {
	Car Value
	Cdr *List
}

var EmptyList *List = &List{}

func NewList(values ...Value) *List {
	return ListFromValues(values)
}


func (lst *List) Type() Value {
	return ListType
}

func (lst *List) String() string {
	var buf bytes.Buffer
	/* Ell stuff
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
	*/
	buf.WriteString("(")
	delim := ""
	for lst != EmptyList {
		buf.WriteString(delim)
		delim = " "
		buf.WriteString(lst.Car.String())
		lst = lst.Cdr
	}
	buf.WriteString(")")
	return buf.String()
}

func (lst1 *List) Equals(another Value) bool {
	if lst2, ok := another.(*List); ok {
		for lst1 != EmptyList {
			if lst2 == EmptyList {
				return false
			}
			if !Equal(lst1.Car, lst2.Car) {
				return false
			}
			lst1 = lst1.Cdr
			lst2 = lst2.Cdr
		}
		if lst1 == lst2 {
			return true
		}
	}
	return false
}

func (lst *List) Length() int {
	if lst == EmptyList {
		return 0
	}
	count := 1
	o := lst.Cdr
	for o != EmptyList {
		count++
		o = o.Cdr
	}
	return count
}

// Cons - create a new list consisting of the first object and the rest of the list
func Cons(car Value, cdr *List) *List {
	return &List{
		Car: car,
		Cdr: cdr,
	}
}

func ListFromValues(values []Value) *List {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = Cons(v, p)
	}
	return p
}

func ListToVector(lst *List) *Vector {
	var elems []Value
	for lst != EmptyList {
		elems = append(elems, lst.Car)
		lst = lst.Cdr
	}
	return VectorFromElementsNoCopy(elems)
}

