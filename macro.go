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
	"fmt"
)

type lmacro struct {
	name     lob
	expander lob //a function of one argument
}

func newMacro(name lob, expander lob) lob {
	mac := lmacro{name, expander}
	return &mac
}

func (*lmacro) typeSymbol() lob {
	return intern("macro")
}
func (mac *lmacro) equal(another lob) bool {
	if a, ok := another.(*lmacro); ok {
		return mac == a
	}
	return false
}

func (mac *lmacro) String() string {
	return fmt.Sprintf("(macro %v %v)", mac.name, mac.expander)
}

var currentModule module

func (mac *lmacro) expand(module module, expr lob) (lob, error) {
	expander := mac.expander
	switch fun := expander.(type) {
	case *lclosure:
		if fun.code.argc == 1 {
			expanded, err := exec(fun.code, expr)
			if err != nil {
				return nil, newError("macro error in '", mac.name, "': ", err)
			}
			return expanded, nil
		}
	case *lprimitive:
		args := []lob{expr}
		currentModule = module //not ideal
		expanded, err := fun.fun(args, 1)
		if err != nil {
			return nil, newError("macro error in '", mac.name, "': ", err)
		}
		return expanded, nil
	}
	return nil, newError("Bad macro expander function")
}

func expandSequence(module module, seq lob) (lob, error) {
	var result []lob
	for isPair(seq) {
		expanded, err := macroexpand(module, car(seq))
		if err != nil {
			return nil, err
		}
		result = append(result, expanded)
		seq = cdr(seq)
	}
	lst := toList(result)
	if seq != NIL {
		tmp := cons(seq, NIL)
		return concat(lst, tmp)
	}
	return lst, nil
}

func expandIf(module module, expr lob) (lob, error) {
	i := length(expr)
	if i == 4 {
		tmp, err := expandSequence(module, cdr(expr))
		if err != nil {
			return nil, err
		}
		return cons(car(expr), tmp), nil
	} else if i == 3 {
		tmp := list(cadr(expr), caddr(expr), NIL)
		tmp, err := expandSequence(module, tmp)
		if err != nil {
			return nil, err
		}
		return cons(car(expr), tmp), nil
	} else {
		return nil, syntaxError(expr)
	}
}

func expandDefine(module module, expr lob) (lob, error) {
	exprLen := length(expr)
	if exprLen < 3 {
		return nil, syntaxError(expr)
	}
	name := cadr(expr)
	if isSymbol(name) {
		if exprLen > 3 {
			return nil, syntaxError(expr)
		}
		val, err := macroexpand(module, caddr(expr))
		if err != nil {
			return nil, err
		}
		return list(car(expr), name, val), nil
	} else if isPair(name) {
		args := cdr(name)
		name = car(name)
		body, err := expandSequence(module, cddr(expr))
		if err != nil {
			return nil, err
		}
		return list(car(expr), name, cons(intern("lambda"), cons(args, body))), nil
	} else {
		return nil, syntaxError(expr)
	}
}

func expandLambda(module module, expr lob) (lob, error) {
	exprLen := length(expr)
	if exprLen < 3 {
		return nil, syntaxError(expr)
	}
	body, err := expandSequence(module, cddr(expr))
	if err != nil {
		return nil, err
	}
	args := cadr(expr)
	return cons(car(expr), cons(args, body)), nil
}

func expandSet(module module, expr lob) (lob, error) {
	exprLen := length(expr)
	if exprLen != 3 {
		return nil, syntaxError(expr)
	}
	val, err := macroexpand(module, caddr(expr))
	if err != nil {
		return nil, err
	}
	return list(car(expr), cadr(expr), val), nil
}

func expandPrimitive(module module, fn lob, expr lob) (lob, error) {
	switch fn {
	case intern("quote"):
		return expr, nil
	case intern("begin"):
		return expandSequence(module, expr)
	case intern("if"):
		return expandIf(module, expr)
	case intern("define"):
		return expandDefine(module, expr)
	case intern("define-macro"):
		return expandDefine(module, expr)
		//return expandDefineMacro(module, expr)
	case intern("lambda"):
		return expandLambda(module, expr)
	case intern("set!"):
		return expandSet(module, expr)
	case intern("lap"):
		return expr, nil
	case intern("use"):
		return expr, nil
	default:
		macro := module.macro(fn)
		if macro != nil {
			return (macro.(*lmacro)).expand(module, expr)
		}
		return nil, nil
	}
}

func macroexpand(module module, expr lob) (lob, error) {
	if isPair(expr) {
		var head lob = NIL
		fn := car(expr)
		if isSymbol(fn) {
			result, err := expandPrimitive(module, fn, expr)
			if err != nil {
				return nil, err
			}
			if result != nil {
				return result, nil
			}
			head = fn
		} else {
			expanded, err := macroexpand(module, fn)
			if err != nil {
				return nil, err
			}
			head = expanded
		}
		tail, err := expandSequence(module, cdr(expr))
		if err != nil {
			return nil, err
		}
		return cons(head, tail), nil
	}
	return expr, nil
}

func crackLetrecBindings(bindings lob, tail lob) (lob, lob, bool) {
	var names []lob
	for bindings != NIL {
		if isPair(bindings) {
			tmp := car(bindings)
			if isPair(tmp) {
				name := car(tmp)
				if isSymbol(name) {
					names = append(names, name)
				} else {
					return nil, nil, false
				}
				if isPair(cdr(tmp)) {
					tail = cons(cons(intern("set!"), tmp), tail)
				} else {
					return nil, nil, false
				}
			} else {
				return nil, nil, false
			}

		} else {
			return nil, nil, false
		}
		bindings = cdr(bindings)
	}
	return toList(names), tail, true
}

func expandLetrec(expr lob) (lob, error) {
	module := currentModule
	// (letrec () expr ...) -> (begin expr ...)
	// (letrec ((x 1) (y 2)) expr ...) -> ((lambda (x y) (set! x 1) (set! y 2) expr ...) nil nil)
	body := cddr(expr)
	if body == NIL {
		return nil, syntaxError(expr)
	}
	names, body, ok := crackLetrecBindings(cadr(expr), body)
	if !ok {
		return nil, syntaxError(expr)
	}
	code, err := macroexpand(module, cons(intern("lambda"), cons(names, body)))
	if err != nil {
		return nil, err
	}
	values := newList(length(names), NIL)
	return cons(code, values), nil
}

func crackLetBindings(module module, bindings lob) (lob, lob, bool) {
	var names []lob
	var values []lob
	for bindings != NIL {
		if isPair(bindings) {
			tmp := car(bindings)
			if isPair(tmp) {
				name := car(tmp)
				if isSymbol(name) {
					names = append(names, name)
					tmp2 := cdr(tmp)
					if isPair(tmp2) {
						val, err := macroexpand(module, car(tmp2))
						if err == nil {
							values = append(values, val)
							bindings = cdr(bindings)
							continue
						}
					}
				}
			}
		}
		return nil, nil, false
	}
	return toList(names), toList(values), true
}

func expandLet(expr lob) (lob, error) {
	module := currentModule
	// (let () expr ...) -> (begin expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((lambda (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (lambda (label) expr
	if isSymbol(cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return expandNamedLet(expr)
	}
	names, values, ok := crackLetBindings(module, cadr(expr))
	if !ok {
		return nil, syntaxError(expr)
	}
	body := cddr(expr)
	if body == NIL {
		return nil, syntaxError(expr)
	}
	code, err := macroexpand(module, cons(intern("lambda"), cons(names, body)))
	if err != nil {
		return nil, err
	}
	return cons(code, values), nil
}

func expandNamedLet(expr lob) (lob, error) {
	module := currentModule
	name := cadr(expr)
	names, values, ok := crackLetBindings(module, caddr(expr))
	if !ok {
		return nil, syntaxError(expr)
	}
	body := cdddr(expr)
	tmp := list(intern("letrec"), list(list(name, cons(intern("lambda"), cons(names, body)))), cons(name, values))
	return macroexpand(module, tmp)
}

func crackDoBindings(module module, bindings lob) (lob, lob, lob, bool) {
	//return names, inits, and incrs
	var names lob = NIL
	var inits lob = NIL
	var steps lob = NIL
	for bindings != NIL {
		if !isPair(bindings) {
			return nil, nil, nil, false
		}
		tmp := car(bindings)
		if !isPair(tmp) {
			return nil, nil, nil, false
		}
		if !isSymbol(car(tmp)) {
			return nil, nil, nil, false
		}
		if !isPair(cdr(tmp)) {
			return nil, nil, nil, false
		}
		names = cons(car(tmp), names)
		inits = cons(cadr(tmp), inits)
		if isPair(cddr(tmp)) {
			steps = cons(caddr(tmp), steps)
		} else {
			steps = cons(car(tmp), steps)
		}
		bindings = cdr(bindings)
	}
	var err error
	inits, err = macroexpand(module, inits)
	if err != nil {
		return nil, nil, nil, false
	}
	steps, err = macroexpand(module, steps)
	if err != nil {
		return nil, nil, nil, false
	}
	return names, inits, steps, true
}

func expandDo(expr lob) (lob, error) {
	// (do ((myvar init-val) ...) (mytest expr ...) body ...)
	// (do ((myvar init-val step) ...) (mytest expr ...) body ...)
	var tmp, tmp2 lob
	if length(expr) < 3 {
		return nil, syntaxError(expr)
	}
	names, inits, steps, ok := crackDoBindings(currentModule, cadr(expr))
	if !ok {
		return nil, syntaxError(expr)
	}
	tmp = caddr(expr)
	exitPred := car(tmp)
	var exitExprs lob = NIL
	if isPair(cddr(tmp)) {
		exitExprs = cons(intern("begin"), cdr(tmp))
	} else {
		exitExprs = cadr(tmp)
	}
	loopSym := intern("system_loop")
	if isPair(cdddr(expr)) {
		//tmp = MacroExpand(Cdddr(expr))
		tmp = cdddr(expr)
		tmp = cons(intern("begin"), tmp)
		tmp2 = cons(loopSym, steps)
		tmp2 = list(tmp2)
		tmp = cons(intern("begin"), cons(tmp, tmp2))
	} else {
		tmp = cons(loopSym, steps)
	}
	tmp = list(tmp)
	tmp = cons(intern("if"), cons(exitPred, cons(exitExprs, tmp)))
	tmp = list(intern("lambda"), names, tmp)
	tmp = list(loopSym, tmp)
	tmp = list(tmp)
	tmp2 = cons(loopSym, inits)
	tmp = list(intern("letrec"), tmp, tmp2)
	return macroexpand(currentModule, tmp)
}
