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

func expandSequence(module LModule, seq LObject) (LObject, error) {
	result := make([]LObject, 0)
	for IsPair(seq) {
		expanded, err := Macroexpand(module, Car(seq))
		if err != nil {
			return nil, err
		}
		result = append(result, expanded)
		seq = Cdr(seq)
	}
	lst := ToList(result)
	if seq != NIL {
		tmp := Cons(seq, NIL)
		return Concat(lst, tmp)
	}
	return lst, nil
}

func expandIf(module LModule, expr LObject) (LObject, error) {
	i := Length(expr)
	if i == 4 {
		tmp, err := expandSequence(module, Cdr(expr))
		if err != nil {
			return nil, err
		}
		return Cons(Car(expr), tmp), nil
	} else if i == 3 {
		tmp := List(Cadr(expr), Caddr(expr), NIL)
		tmp, err := expandSequence(module, tmp)
		if err != nil {
			return nil, err
		}
		return Cons(Car(expr), tmp), nil
	} else {
		return nil, Error("syntax error: ", expr)
	}
}

func expandDefine(module LModule, expr LObject) (LObject, error) {
	exprLen := Length(expr)
	if exprLen < 3 {
		return nil, Error("syntax error: ", expr)
	}
	name := Cadr(expr)
	if IsSymbol(name) {
		if exprLen > 3 {
			return nil, Error("syntax error: ", expr)
		}
		val, err := Macroexpand(module, Caddr(expr))
		if err != nil {
			return nil, err
		}
		return List(Car(expr), name, val), nil
	} else if IsPair(name) {
		args := Cdr(name)
		name = Car(name)
		body, err := expandSequence(module, Cddr(expr))
		if err != nil {
			return nil, err
		}
		return List(Car(expr), name, Cons(Intern("lambda"), Cons(args, body))), nil
	} else {
		return nil, Error("syntax error: ", expr)
	}
}

func expandLambda(module LModule, expr LObject) (LObject, error) {
	exprLen := Length(expr)
	if exprLen < 3 {
		return nil, Error("syntax error: ", expr)
	}
	body, err := expandSequence(module, Cddr(expr))
	if err != nil {
		return nil, err
	}
	args := Cadr(expr)
	return Cons(Car(expr), Cons(args, body)), nil
}

func expandSet(module LModule, expr LObject) (LObject, error) {
	exprLen := Length(expr)
	if exprLen != 3 {
		return nil, Error("syntax error: ", expr)
	}
	val, err := Macroexpand(module, Caddr(expr))
	if err != nil {
		return nil, err
	}
	return List(Car(expr), Cadr(expr), val), nil
}

func Macroexpand(module LModule, expr LObject) (LObject, error) {
	if IsPair(expr) {
		var head LObject = NIL
		fn := Car(expr)
		if IsSymbol(fn) {
			switch fn {
			case Intern("quote"):
				return expr, nil
			case Intern("begin"):
				return expandSequence(module, expr)
			case Intern("if"):
				return expandIf(module, expr)
			case Intern("define"):
				return expandDefine(module, expr)
			case Intern("define-macro"):
				return expandDefine(module, expr)
				//return expandDefineMacro(module, expr)
			case Intern("lambda"):
				return expandLambda(module, expr)
			case Intern("set!"):
				return expandSet(module, expr)
			case Intern("lap"):
				return expr, nil
			case Intern("use"):
				return expr, nil
			default:
				macro := module.Macro(fn)
				if macro != nil {
					return (macro.(*lmacro)).Expand(module, expr)
				} else {
					head = fn
				}
			}
		} else {
			expanded, err := Macroexpand(module, fn)
			if err != nil {
				return nil, err
			}
			head = expanded
		}
		tail, err := expandSequence(module, Cdr(expr))
		if err != nil {
			return nil, err
		}
		return Cons(head, tail), nil
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
	module := currentModule
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
	code, err := Macroexpand(module, Cons(Intern("lambda"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	values := NewList(Length(names), NIL)
	return Cons(code, values), nil
}

func crackLetBindings(module LModule, bindings LObject) (LObject, LObject, bool) {
	var names []LObject
	var values []LObject
	for bindings != NIL {
		if IsPair(bindings) {
			tmp := Car(bindings)
			if IsPair(tmp) {
				name := Car(tmp)
				if IsSymbol(name) {
					names = append(names, name)
					tmp2 := Cdr(tmp)
					if IsPair(tmp2) {
						val, err := Macroexpand(module, Car(tmp2))
						if err == nil {
							values = append(values, val)
							bindings = Cdr(bindings)
							continue
						}
					}
				}
			}
		}
		return nil, nil, false
	}
	return ToList(names), ToList(values), true
}

func ExpandLet(expr LObject) (LObject, error) {
	module := currentModule
	// (let () expr ...) -> (begin expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((lambda (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (lambda (label) expr
	if IsSymbol(Cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return ExpandNamedLet(expr)
	}
	names, values, ok := crackLetBindings(module, Cadr(expr))
	if !ok {
		return nil, Error("bad syntax for let: ", expr)
	}
	body := Cddr(expr)
	if body == NIL {
		return nil, Error("bad syntax for let: ", expr)
	}
	code, err := Macroexpand(module, Cons(Intern("lambda"), Cons(names, body)))
	if err != nil {
		return nil, err
	}
	return Cons(code, values), nil
}

func ExpandNamedLet(expr LObject) (LObject, error) {
	module := currentModule
	name := Cadr(expr)
	names, values, ok := crackLetBindings(module, Caddr(expr))
	if !ok {
		return nil, Error("bad syntax for let: ", expr)
	}
	body := Cdddr(expr)
	tmp := List(Intern("letrec"), List(List(name, Cons(Intern("lambda"), Cons(names, body)))), Cons(name, values))
	return Macroexpand(module, tmp)
}

func crackDoBindings(module LModule, bindings LObject) (LObject, LObject, LObject, bool) {
	//return names, inits, and incrs
	var names LObject = NIL
	var inits LObject = NIL
	var steps LObject = NIL
	for bindings != NIL {
		if !IsPair(bindings) {
			return nil, nil, nil, false
		}
		tmp := Car(bindings)
		if !IsPair(tmp) {
			return nil, nil, nil, false
		}
		if !IsSymbol(Car(tmp)) {
			return nil, nil, nil, false
		}
		if !IsPair(Cdr(tmp)) {
			return nil, nil, nil, false
		}
		names = Cons(Car(tmp), names)
		inits = Cons(Cadr(tmp), inits)
		if IsPair(Cddr(tmp)) {
			steps = Cons(Caddr(tmp), steps)
		} else {
			steps = Cons(Car(tmp), steps)
		}
		bindings = Cdr(bindings)
	}
	var err error
	inits, err = Macroexpand(module, inits)
	if err != nil { return nil, nil, nil, false }
	steps, err = Macroexpand(module, steps)
	if err != nil { return nil, nil, nil, false }
	return names, inits, steps, true
}

func ExpandDo(expr LObject) (LObject, error) {
	// (do ((myvar init-val) ...) (mytest expr ...) body ...)
	// (do ((myvar init-val step) ...) (mytest expr ...) body ...)
	//assert length(expr) >= 3
	var tmp, tmp2 LObject
	if Length(expr) < 3 {
		return nil, Error("bad syntax for do: ", expr)
	}
	names, inits, steps, ok := crackDoBindings(currentModule, Cadr(expr))
	if !ok {
		return nil, Error("bad syntax for do: ", expr)
	}
	//var err error
	//tmp, err = Macroexpand(currentModule, Caddr(expr))
	//if err != nil {
	//	return nil, err
	//}
	tmp = Caddr(expr)
	exitPred := Car(tmp)
	var exitExprs LObject = NIL
	if IsPair(Cddr(tmp)) {
		exitExprs = Cons(Intern("begin"), Cdr(tmp))
	} else {
		exitExprs = Cadr(tmp)
	}
	loopSym := Intern("system_loop")
	if IsPair(Cdddr(expr)) {
		//tmp = MacroExpand(Cdddr(expr))
		tmp = Cdddr(expr)
		tmp = Cons(Intern("begin"), tmp)
		tmp2 = Cons(loopSym, steps)
		tmp2 = List(tmp2)
		tmp = Cons(Intern("begin"), Cons(tmp, tmp2))
	} else {
		tmp = Cons(loopSym, steps)
	}
	tmp = List(tmp)
	tmp = Cons(Intern("if"), Cons(exitPred, Cons(exitExprs, tmp)))
	tmp = List(Intern("lambda"), names, tmp)
	tmp = List(loopSym, tmp)
	tmp = List(tmp)
	tmp2 = Cons(loopSym, inits)
	tmp = List(Intern("letrec"), tmp, tmp2)
	return Macroexpand(currentModule, tmp)
}
