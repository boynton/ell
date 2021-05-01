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
	"fmt"
	. "github.com/boynton/ell/data"
)

var MacroErrorKey = Intern("macro-error:")

type macro struct {
	name     Value
	expander *Function //a function of one argument
}

// Macro - create a new Macro
func NewMacro(name Value, expander *Function) *macro {
	return &macro{name, expander}
}

func (mac *macro) String() string {
	return fmt.Sprintf("(macro %v %v)", mac.name, mac.expander)
}

// Macroexpand - return the expansion of all macros in the object and return the result
func Macroexpand(expr Value) (Value, error) {
	return macroexpandObject(expr)
}

func macroexpandObject(expr Value) (Value, error) {
	if lst, ok := expr.(*List); ok {
		if lst != EmptyList {
			return macroexpandList(lst)
		}
	}
	return expr, nil
}

func macroexpandList(expr *List) (Value, error) {
	if expr == nil {
		panic("whoops")
	}
	if expr == EmptyList {
		return expr, nil
	}
	lst := expr
	fn := Car(lst)
	head := fn
	if IsSymbol(fn) {
		result, err := expandPrimitive(fn, lst)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		head = fn
	} else if lst, ok := fn.(*List); ok {
		//panic("non-primitive macro")
		expanded, err := macroexpandList(lst)
		if err != nil {
			return nil, err
		}
		head = expanded
	}
	tail, err := expandSequence(Cdr(expr))
	if err != nil {
		return nil, err
	}
	return Cons(head, tail), nil
}

func (mac *macro) expand(expr Value) (Value, error) {
	if mac.expander.code != nil {
		if mac.expander.code.argc == 1 {
			expanded, err := execCompileTime(mac.expander.code, expr)
			if err == nil {
				if IsList(expanded) {
					return macroexpandObject(expanded)
				}
				return expanded, err
			}
			return nil, err
		}
	} else if mac.expander.primitive != nil {
		args := []Value{expr}
		expanded, err := mac.expander.primitive.fun(args)
		if err == nil {
			return macroexpandObject(expanded)
		}
		return nil, err
	}
	return nil, NewError(MacroErrorKey, "Bad macro expander function: ", mac.expander)
}

func expandSequence(seq Value) (*List, error) {
	var result []Value
	if seq == nil {
		panic("Whoops: should be (), not nil!")
	}
	for seq != EmptyList {
		item := Car(seq)
		if lst, ok := item.(*List); ok {
			expanded, err := macroexpandList(lst)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded)
		} else {
			result = append(result, item)
		}
		seq = Cdr(seq)
	}
	lst := ListFromValues(result)
	if seq != EmptyList {
		tmp := Cons(seq, EmptyList)
		return Concat(lst, tmp)
	}
	return lst, nil
}

func expandIf(expr Value) (Value, error) {
	i := ListLength(expr)
	if i == 4 {
		tmp, err := expandSequence(Cdr(expr))
		if err != nil {
			return nil, err
		}
		return Cons(Car(expr), tmp), nil
	} else if i == 3 {
		tmp := NewList(Cadr(expr), Caddr(expr), Null)
		tmp, err := expandSequence(tmp)
		if err != nil {
			return nil, err
		}
		return Cons(Car(expr), tmp), nil
	} else {
		return nil, NewError(SyntaxErrorKey, expr)
	}
}

func expandUndef(expr Value) (Value, error) {
	if ListLength(expr) != 2 || !IsSymbol(Cadr(expr)) {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	return expr, nil
}

// (defn f (x) (+ 1 x))
//  ->
// (def f (fn (x) (+ 1 x)))
func expandDefn(expr Value) (Value, error) {
	exprLen := ListLength(expr)
	if exprLen >= 4 {
		name := Cadr(expr)
		if IsSymbol(name) {
			args := Caddr(expr)
			body, err := expandSequence(Cdddr(expr))
			if err != nil {
				return nil, err
			}
			tmp, err := expandFn(Cons(Intern("fn"), Cons(args, body)))
			if err != nil {
				return nil, err
			}
			return NewList(Intern("def"), name, tmp), nil
		}
	}
	return nil, NewError(SyntaxErrorKey, expr)
}

func expandDefmacro(expr Value) (Value, error) {
	exprLen := ListLength(expr)
	if exprLen >= 4 {
		name := Cadr(expr)
		if IsSymbol(name) {
			args := Caddr(expr)
			body, err := expandSequence(Cdddr(expr))
			if err != nil {
				return nil, err
			}
			//(fn (expr) (apply xxx
			tmp, err := expandFn(Cons(Intern("fn"), Cons(args, body))) //this is the expander with special args\
			if err != nil {
				return nil, err
			}
			sym := Intern("expr")
			tmp, err = expandFn(NewList(Intern("fn"), NewList(sym), NewList(Intern("apply"), tmp, NewList(Intern("cdr"), sym))))
			if err != nil {
				return nil, err
			}
			return NewList(Intern("defmacro"), name, tmp), nil
		}
	}
	return nil, NewError(SyntaxErrorKey, expr)
}

//(defmacro (defmacro expr)
//  `(defmacro ~(cadr expr) (fn (expr) (apply (fn ~(caddr expr) ~@(cdddr expr)) (cdr expr)))))

func expandDef(expr Value) (Value, error) {
	exprLen := ListLength(expr)
	if exprLen != 3 {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	name := Cadr(expr)
	if !IsSymbol(name) {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	if exprLen > 3 {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	body := Caddr(expr)
	if lst, ok := body.(*List); ok {
		val, err := macroexpandList(lst)
		if err != nil {
			return nil, err
		}
		return NewList(Car(expr), name, val), nil
	}
	return expr, nil
}

func expandFn(expr Value) (Value, error) {
	exprLen := ListLength(expr)
	if exprLen < 3 {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	body, err := expandSequence(Cddr(expr))
	if err != nil {
		return nil, err
	}
	bodyLen := ListLength(body)
	if bodyLen > 0 {
		tmp := body
		if IsList(tmp) && Caar(tmp) == Intern("def") || Caar(tmp) == Intern("defmacro") {
			bindings := EmptyList
			for Caar(tmp) == Intern("def") || Caar(tmp) == Intern("defmacro") {
				if Caar(tmp) == Intern("defmacro") {
					return nil, NewError(MacroErrorKey, "macros can only be defined at top level")
				}
				def, err := expandDef(Car(tmp))
				if err != nil {
					return nil, err
				}
				bindings = Cons(Cdr(def), bindings)
				tmp = Cdr(tmp)
			}
			bindings = Reverse(bindings)
			tmp = Cons(Intern("letrec"), Cons(bindings, tmp)) //scheme specifies letrec*
			tmp2, err := macroexpandList(tmp)
			return NewList(Car(expr), Cadr(expr), tmp2), err
		}
	}
	args := Cadr(expr)
	return Cons(Car(expr), Cons(args, body)), nil
}

func expandSetBang(expr Value) (Value, error) {
	exprLen := ListLength(expr)
	if exprLen != 3 {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	var val = Caddr(expr)
	if lst, ok := val.(*List); ok {
		v, err := macroexpandList(lst)
		if err != nil {
			return nil, err
		}
		val = v
	}
	return NewList(Car(expr), Cadr(expr), val), nil
}

func expandPrimitive(fn Value, expr Value) (Value, error) {
	switch fn {
	case Intern("quote"):
		return expr, nil
	case Intern("do"):
		return expandSequence(expr)
	case Intern("if"):
		return expandIf(expr)
	case Intern("def"):
		return expandDef(expr)
	case Intern("undef"):
		return expandUndef(expr)
	case Intern("defn"):
		return expandDefn(expr)
	case Intern("defmacro"):
		return expandDefmacro(expr)
	case Intern("fn"):
		return expandFn(expr)
	case Intern("set!"):
		return expandSetBang(expr)
	case Intern("lap"):
		return expr, nil
	case Intern("use"):
		return expr, nil
	default:
		macro := GetMacro(fn)
		if macro != nil {
			tmp, err := macro.expand(expr)
			return tmp, err
		}
		return nil, nil
	}
}

func crackLetrecBindings(bindings Value, tail *List) (*List, *List, bool) {
	var names []Value
	inits := EmptyList
	for bindings != EmptyList {
		if IsList(bindings) {
			tmp := Car(bindings)
			if lst, ok := tmp.(*List); ok {
				name := lst.Car
				if IsSymbol(name) {
					names = append(names, name)
				} else {
					return nil, nil, false
				}
				if IsList(lst.Cdr) {
					inits = Cons(Cons(Intern("set!"), lst), inits)
				} else {
					return nil, nil, false
				}
			} else {
				return nil, nil, false
			}

		} else {
			return nil, nil, false
		}
		bindings = Cdr(bindings)
	}
	inits = Reverse(inits)
	head := inits
	for inits.Cdr != EmptyList {
		inits = inits.Cdr
	}
	inits.Cdr = tail
	return ListFromValues(names), head, true
}

func expandLetrec(expr Value) (Value, error) {
	// (letrec () expr ...) -> (do expr ...)
	// (letrec ((x 1) (y 2)) expr ...) -> ((fn (x y) (set! x 1) (set! y 2) expr ...) nil nil)
	body := Cddr(expr)
	if body == EmptyList {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	bindings := Cadr(expr)
	if !IsList(bindings) {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	names, body, ok := crackLetrecBindings(bindings, body)
	if !ok {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	code, err := macroexpandList(Cons(Intern("fn"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	values := MakeList(ListLength(names), Null)
	return Cons(code, values), nil
}

func crackLetBindings(bindings Value) (*List, *List, bool) {
	var names []Value
	var values []Value
	for bindings != EmptyList {
		tmp := Car(bindings)
		if IsList(tmp) {
			name := Car(tmp)
			if IsSymbol(name) {
				names = append(names, name)
				tmp2 := Cdr(tmp)
				if tmp2 != EmptyList {
					val, err := macroexpandObject(Car(tmp2))
					if err == nil {
						values = append(values, val)
						bindings = Cdr(bindings)
						continue
					}
				}
			}
		}
		return nil, nil, false
	}
	return ListFromValues(names), ListFromValues(values), true
}

func expandLet(expr Value) (Value, error) {
	// (let () expr ...) -> (do expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((fn (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (fn (label) expr
	if IsSymbol(Cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return expandNamedLet(expr)
	}
	bindings := Cadr(expr)
	if !IsList(bindings) {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	names, values, ok := crackLetBindings(bindings)
	if !ok {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	body := Cddr(expr)
	if body == EmptyList {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	code, err := macroexpandList(Cons(Intern("fn"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	return Cons(code, values), nil
}

func expandNamedLet(expr Value) (Value, error) {
	name := Cadr(expr)
	bindings := Caddr(expr)
	if !IsList(bindings) {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	names, values, ok := crackLetBindings(bindings)
	if !ok {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	body := Cdddr(expr)
	tmp := NewList(Intern("letrec"), NewList(NewList(name, Cons(Intern("fn"), Cons(names, body)))), Cons(name, values))
	return macroexpandList(tmp)
}

func nextCondClause(expr Value, clauses Value, count int) (Value, error) {
	var result Value
	var err error
	tmpsym := Intern("__tmp__")
	ifsym := Intern("if")
	elsesym := Intern("else")
	letsym := Intern("let")
	dosym := Intern("do")

	clause0 := Car(clauses)
	next := Cdr(clauses)
	clause1 := Car(next)

	if count == 2 {
		if !IsList(clause1) {
			return nil, NewError(SyntaxErrorKey, expr)
		}
		if elsesym == Car(clause1) {
			if Cadr(clause0) == Intern("=>") {
				if ListLength(clause0) != 3 {
					return nil, NewError(SyntaxErrorKey, expr)
				}
				result = NewList(letsym, NewList(NewList(tmpsym, Car(clause0))), NewList(ifsym, tmpsym, NewList(Caddr(clause0), tmpsym), Cons(dosym, Cdr(clause1))))
			} else {
				result = NewList(ifsym, Car(clause0), Cons(dosym, Cdr(clause0)), Cons(dosym, Cdr(clause1)))
			}
		} else {
			if Cadr(clause1) == Intern("=>") {
				if ListLength(clause1) != 3 {
					return nil, NewError(SyntaxErrorKey, expr)
				}
				result = NewList(letsym, NewList(NewList(tmpsym, Car(clause1))), NewList(ifsym, tmpsym, NewList(Caddr(clause1), tmpsym), clause1))
			} else {
				result = NewList(ifsym, Car(clause1), Cons(dosym, Cdr(clause1)))
			}
			if Cadr(clause0) == Intern("=>") {
				if ListLength(clause0) != 3 {
					return nil, NewError(SyntaxErrorKey, expr)
				}
				result = NewList(letsym, NewList(NewList(tmpsym, Car(clause0))), NewList(ifsym, tmpsym, NewList(Caddr(clause0), tmpsym), result))
			} else {
				result = NewList(ifsym, Car(clause0), Cons(dosym, Cdr(clause0)), result)
			}
		}
	} else {
		result, err = nextCondClause(expr, next, count-1)
		if err != nil {
			return nil, err
		}
		if Cadr(clause0) == Intern("=>") {
			if ListLength(clause0) != 3 {
				return nil, NewError(SyntaxErrorKey, expr)
			}
			result = NewList(letsym, NewList(NewList(tmpsym, Car(clause0))), NewList(ifsym, tmpsym, NewList(Caddr(clause0), tmpsym), result))
		} else {
			result = NewList(ifsym, Car(clause0), Cons(dosym, Cdr(clause0)), result)
		}
	}
	return macroexpandObject(result)
}

func expandCond(expr Value) (Value, error) {
	i := ListLength(expr)
	if i < 2 {
		return nil, NewError(SyntaxErrorKey, expr)
	} else if i == 2 {
		tmp := Cadr(expr)
		if Car(tmp) == Intern("else") {
			tmp = Cons(Intern("do"), Cdr(tmp))
		} else {
			expr = Cons(Intern("do"), Cdr(tmp))
			tmp = NewList(Intern("if"), Car(tmp), expr)
		}
		return macroexpandObject(tmp)
	} else {
		return nextCondClause(expr, Cdr(expr), i-1)
	}
}

func expandQuasiquote(expr Value) (Value, error) {
	if ListLength(expr) != 2 {
		return nil, NewError(SyntaxErrorKey, expr)
	}
	return expandQQ(Cadr(expr))
}

func expandQQ(expr Value) (Value, error) {
	switch p := expr.(type) {
	case *List:
		if p == EmptyList {
			return expr, nil
		}
		if p.Cdr != EmptyList {
			if p.Car == UnquoteSymbol {
				if p.Cdr.Cdr != EmptyList {
					return nil, NewError(SyntaxErrorKey, expr)
				}
				return macroexpandObject(p.Cdr.Car)
			} else if p.Car == UnquoteSymbolSplicing {
				return nil, NewError(MacroErrorKey, "unquote-splicing can only occur in the context of a list ")
			}
		}
		tmp, err := expandQQList(p)
		if err != nil {
			return nil, err
		}
		return macroexpandObject(tmp)
	case *Symbol:
		return NewList(Intern("quote"), expr), nil
	default: //all other objects evaluate to themselves
		return expr, nil
	}
}

func expandQQList(lst *List) (*List, error) {
	var tmp Value
	var err error
	result := NewList(Intern("concat"))
	tail := result
	for lst != EmptyList {
		if item, ok := Car(lst).(*List); ok && item != EmptyList{
			if item.Car == QuasiquoteSymbol {
				return nil, NewError(MacroErrorKey, "nested quasiquote not supported")
			}
			if item.Car == UnquoteSymbol && item.Length() == 2 {
				tmp, err = macroexpandObject(Cadr(item))
				tmp = NewList(Intern("list"), tmp)
				if err != nil {
					return nil, err
				}
				tail.Cdr = NewList(tmp)
				tail = tail.Cdr
			} else if item.Car == UnquoteSymbolSplicing && item.Length() == 2 {
				tmp, err = macroexpandObject(Cadr(item))
				if err != nil {
					return nil, err
				}
				tail.Cdr = NewList(tmp)
				tail = tail.Cdr
			} else {
				tmp, err = expandQQList(item)
				if err != nil {
					return nil, err
				}
				tail.Cdr = NewList(NewList(Intern("list"), tmp))
				tail = tail.Cdr
			}
		} else {
			tail.Cdr = NewList(NewList(Intern("quote"), NewList(Car(lst))))
			tail = tail.Cdr
		}
		lst = lst.Cdr
	}
	return result, nil
}
