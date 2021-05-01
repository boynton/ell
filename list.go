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
	. "github.com/boynton/ell/data" // -> "github.com/boynton/data"
)

// Car - return the first object in a list
func Car(v Value) Value {
	if lst, ok := v.(*List); ok && lst != nil {
		if lst != EmptyList {
			return lst.Car
		}
	}
	return Null //in Scheme, this is an error
}

// Cdr - return the rest of the list
func Cdr(v Value) *List {
	if lst, ok := v.(*List); ok {
		if lst != EmptyList {
			return lst.Cdr
		}
		//in Scheme, this is an error
	}
	return EmptyList
}

// Caar - return the Car of the Car of the list
func Caar(lst Value) Value {
	return Car(Car(lst))
}

// Cadr - return the Car of the Cdr of the list
func Cadr(lst Value) Value {
	return Car(Cdr(lst))
}

// Cdar - return the Cdr of the Car of the list
func Cdar(lst Value) *List {
	return Cdr(Car(lst))
}

// Cddr - return the Cdr of the Cdr of the list
func Cddr(lst Value) *List {
	return Cdr(Cdr(lst))
}

// Cadar - return the Car of the Cdr of the Car of the list
func Cadar(lst Value) Value {
	return Car(Cdr(Car(lst)))
}

// Caddr - return the Car of the Cdr of the Cdr of the list
func Caddr(lst Value) Value {
	return Car(Cdr(Cdr(lst)))
}

// Cdddr - return the Cdr of the Cdr of the Cdr of the list
func Cdddr(lst Value) *List {
	return Cdr(Cdr(Cdr(lst)))
}

// Cadddr - return the Car of the Cdr of the Cdr of the Cdr of the list
func Cadddr(lst Value) Value {
	return Car(Cdr(Cdr(Cdr(lst))))
}

// Cddddr - return the Cdr of the Cdr of the Cdr of the Cdr of the list
func Cddddr(lst Value) *List {
	return Cdr(Cdr(Cdr(Cdr(lst))))
}

// ListEqual returns true if the object is equal to the argument
func ListEqual(l1 Value, l2 Value) bool {
	if lst, ok := l1.(*List); ok {
		return lst.Equals(l2)
	}
	return false
}

/*
   func listToString(lst *Object) string {
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
*/

func ListLength(o Value) int {
	if lst, ok := o.(*List); ok {
		return lst.Length()
	}
	return -1
}

func MakeList(count int, val Value) *List {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = Cons(val, result)
	}
	return result
}

// ToList - convert the argument to a List, if possible
func ToList(obj Value) (*List, error) {
	switch p := obj.(type) {
	case *List:
		return p, nil
	case *Vector:
		return ListFromValues(p.Elements), nil
	case *Struct:
		return StructToList(p)
	case *String:
		return StringToList(p), nil
	}
	return nil, NewError(ArgumentErrorKey, "to-list cannot accept ", obj.Type())
}

func Reverse(lst *List) *List {
	rev := EmptyList
	for lst != EmptyList {
		rev = Cons(lst.Car, rev)
		lst = lst.Cdr
	}
	return rev
}

func Flatten(lst *List) *List {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.Car
		var lstItem *List
		switch p := item.(type) {
		case *List:
			lstItem = Flatten(p)
		case *Vector:
			litem, _ := ToList(item)
			lstItem = Flatten(litem)
		default:
			lstItem = NewList(item)
		}
		if tail == EmptyList {
			result = lstItem
			tail = result
		} else {
			tail.Cdr = lstItem
		}
		for tail.Cdr != EmptyList {
			tail = tail.Cdr
		}
		lst = lst.Cdr
	}
	return result
}

func Concat(seq1 *List, seq2 *List) (*List, error) {
	rev := Reverse(seq1)
	if rev == EmptyList {
		return seq2, nil
	}
	lst := seq2
	for rev != EmptyList {
		lst = Cons(rev.Car, lst)
		rev = rev.Cdr
	}
	return lst, nil
}

func IsList(obj Value) bool {
	if _, ok := obj.(*List); ok {
		return true
	}
	return false
}
