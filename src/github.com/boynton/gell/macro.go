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

package gell

import (
	"fmt"
)

type lmacro struct {
	name     LObject
	expander LObject //a function of one argument
}

func NewMacro(name LObject, expander LObject) LObject {
	mac := lmacro{name, expander}
	return &mac
}

func (*lmacro) Type() LObject {
	return Intern("macro")
}
func (m *lmacro) Equal(another LObject) bool {
	if a, ok := another.(*lmacro); ok {
		return m == a
	}
	return false
}

func (macro *lmacro) String() string {
	return fmt.Sprintf("(macro %v %v)", macro.name, macro.expander)
}

var currentModule LModule

func (macro *lmacro) Expand(module LModule, expr LObject) (LObject, error) {
	expander := macro.expander
	switch fun := expander.(type) {
	case *lclosure:
		if fun.code.argc == 1 {
			expanded, err := Exec(fun.code, expr)
			if err != nil {
				return nil, Error("macro error in '", macro.name, "': ", err)
			}
			return expanded, nil
		}
	case *lprimitive:
		args := []LObject{expr}
		currentModule = module //not ideal
		expanded, err := fun.fun(args, 1)
		if err != nil {
			return nil, Error("macro error in '", macro.name, "': ", err)
		}
		return expanded, nil
	}
	return nil, Error("Bad macro expander function")
}

func Macroexpand(module LModule, expr LObject) (LObject, error) {
	if IsPair(expr) {
		var head LObject = NIL
		fn := Car(expr)
		if IsSymbol(fn) {
			macro := module.Macro(fn)
			if macro != nil {
				return (macro.(*lmacro)).Expand(module, expr)
			} else {
				head = Cons(fn, NIL)
			}
		} else {
			expanded, err := Macroexpand(module, fn)
			if err != nil {
				return nil, err
			}
			head = Cons(expanded, NIL)
		}
		var tail LObject = head
		rest := Cdr(expr)
		for IsPair(rest) {
			expanded, err := Macroexpand(module, Car(rest))
			if err != nil {
				return nil, err
			}
			tmp := Cons(expanded, NIL)
			(tail.(*lpair)).cdr = tmp
			tail = tmp
			rest = Cdr(rest)
		}
		if rest != NIL {
			(tail.(*lpair)).cdr = rest
		}
		return head, nil
	}
	return expr, nil
}

func crackLetrecBindings(bindings LObject, tail LObject) (LObject, LObject, bool) {
	var names []LObject
	for bindings != NIL {
		if IsPair(bindings) {
			tmp := Car(bindings)
			if IsPair(tmp) {
				name := Car(tmp)
				if IsSymbol(name) {
					names = append(names, name)
				} else {
					return nil, nil, false
				}
				if IsPair(Cdr(tmp)) {
					tail = Cons(Cons(Intern("set!"), tmp), tail)
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
	return ToList(names), tail, true
}

func ExpandLetrec(expr LObject) (LObject, error) {
	// (letrec () expr ...) -> (begin expr ...)
	// (letrec ((x 1) (y 2)) expr ...) -> ((lambda (x y) (set! x 1) (set! y 2) expr ...) nil nil)
	body := Cddr(expr)
	if body == NIL {
                return nil, Error("no body in letrec: ", expr)
	}
	names, body, ok := crackLetrecBindings(Cadr(expr), body)
	if !ok {
		return nil, Error("bad bindings declaration syntax for letrec: ", expr)
	}
	code, err := Macroexpand(currentModule, Cons(Intern("lambda"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	values := NewList(Length(names), NIL)
	return Cons(code, values), nil
}

// some standard expanders
func crackLetBindings(bindings LObject) (LObject, LObject, bool) {
	var names []LObject
	var values []LObject
	for bindings != NIL {
		if IsPair(bindings) {
			tmp := Car(bindings)
			if IsPair(tmp) {
				name := Car(tmp)
				if IsSymbol(name) {
					names = append(names, name)
				} else {
					
					return nil, nil, false
				}
				tmp2 := Cdr(tmp)
				if IsPair(tmp2) {
					values = append(values, Car(tmp2))
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
	return ToList(names), ToList(values), true
}

func ExpandLet(expr LObject) (LObject, error) {
	// (let () expr ...) -> (begin expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((lambda (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (lambda (label) expr
	if IsSymbol(Cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return ExpandNamedLet(expr)
	}
	names, values, ok := crackLetBindings(Cadr(expr))
	if !ok {
		return nil, Error("bad syntax for let: ", expr)
	}
	body := Cddr(expr)
	if body == NIL {
                return nil, Error("bad syntax for let: ", expr)
	}
	code, err := Macroexpand(currentModule, Cons(Intern("lambda"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	return Cons(code, values), nil
}

func ExpandNamedLet(expr LObject) (LObject, error) {
	name := Cadr(expr)
        names, values, ok := crackLetBindings(Caddr(expr))
	if !ok {
                return nil, Error("bad syntax for let: ", expr)
	}
	body := Cdddr(expr)
	tmp := List(Intern("letrec"), List(List(name, Cons(Intern("lambda"), Cons(names, body)))), Cons(name, values))
	return Macroexpand(currentModule, tmp)
}

