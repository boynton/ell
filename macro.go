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

func macroexpandObject(mod module, expr lob) (lob, error) {
	if lst, ok := expr.(*llist); ok {
		if lst != EMPTY_LIST {
			return macroexpand(mod, lst)
		}
	}
	return expr, nil
}

func macroexpand(module module, expr *llist) (*llist, error) {
	lst := expr
	fn := car(lst)
	var head lob = fn
	if isSymbol(fn) {
		result, err := expandPrimitive(module, fn, lst)
		if err != nil {
			println("cannot expand primitive: ", fn)
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		head = fn
	} else if isList(fn) {
		expanded, err := macroexpand(module, fn.(*llist))
		if err != nil {
			println("cannot macroexpand list")
			return nil, err
		}
		head = expanded
	}
	tail, err := expandSequence(module, cdr(expr))
	if err != nil {
			println("cannot macroexpand sequence")
		return nil, err
	}
	return cons(head, tail), nil
}

var currentModule module

func (mac *lmacro) expand(module module, expr *llist) (*llist, error) {
	expander := mac.expander
	switch fun := expander.(type) {
	case *lclosure:
		if fun.code.argc == 1 {
			expanded, err := exec(fun.code, expr)
			if err == nil {
				if result, ok := expanded.(*llist); ok {
					return result, nil
				}
			}
			return nil, newError("macro error in '", mac.name, "': ", err)
		}
	case *lprimitive:
		args := []lob{expr}
		currentModule = module //not ideal
		expanded, err := fun.fun(args, 1)
		if err == nil {
			if result, ok := expanded.(*llist); ok {
				return result, nil
			}
			err = newError("macro error in '", mac.name, "': ", err)
		}
		return nil, err
	}
	return nil, newError("Bad macro expander function")
}

func expandSequence(module module, seq *llist) (*llist, error) {
	var result []lob
	if seq == nil {
		panic("Whoops: should be (), not nil!")
	}
	for seq != EMPTY_LIST {
		switch item := car(seq).(type) {
		case *llist:
			expanded, err := macroexpand(module, item)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded)
		default:
			result = append(result, item)
		}
		seq = cdr(seq)
	}
	lst := toList(result)
	if seq != EMPTY_LIST {
		tmp := cons(seq, EMPTY_LIST)
		return concat(lst, tmp)
	}
	return lst, nil
}

func expandIf(module module, expr lob) (*llist, error) {
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

func expandDefine(module module, expr *llist) (*llist, error) {
	exprLen := length(expr)
	if exprLen < 3 {
		return nil, syntaxError(expr)
	}
	name := cadr(expr)
	if isSymbol(name) {
		if exprLen > 3 {
			return nil, syntaxError(expr)
		}
		body, ok := caddr(expr).(*llist)
		if !ok {
			return nil, syntaxError(expr)
		}
		val, err := macroexpand(module, body)
		if err != nil {
			return nil, err
		}
		return list(car(expr), name, val), nil
	} else if isList(name) {
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

func expandLambda(module module, expr *llist) (*llist, error) {
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

func expandSet(module module, expr *llist) (*llist, error) {
	exprLen := length(expr)
	if exprLen != 3 {
		return nil, syntaxError(expr)
	}
	var val = caddr(expr)
	switch vv := val.(type) {
	case *llist:
		v, err := macroexpand(module, vv)
		if err != nil {
			return nil, err
		}
		val = v
	}
	return list(car(expr), cadr(expr), val), nil
}

func expandPrimitive(module module, fn lob, expr *llist) (*llist, error) {
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

func crackLetrecBindings(bindings *llist, tail *llist) (*llist, *llist, bool) {
	var names []lob
	for bindings != EMPTY_LIST {
		if isList(bindings) {
			tmp := car(bindings)
			if isList(tmp) {
				name := car(tmp)
				if isSymbol(name) {
					names = append(names, name)
				} else {
					return nil, nil, false
				}
				if isList(cdr(tmp)) {
					tail = cons(cons(intern("set!"), tmp.(*llist)), tail)
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

func expandLetrec(expr *llist) (lob, error) {
	module := currentModule
	// (letrec () expr ...) -> (begin expr ...)
	// (letrec ((x 1) (y 2)) expr ...) -> ((lambda (x y) (set! x 1) (set! y 2) expr ...) nil nil)
	body := cddr(expr)
	if body == EMPTY_LIST {
		return nil, syntaxError(expr)
	}
	bindings := cadr(expr)
	lstBindings, ok := bindings.(*llist)
	if !ok {
		return nil, syntaxError(expr)
	}
	names, body, ok := crackLetrecBindings(lstBindings, body)
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

func crackLetBindings(module module, bindings *llist) (*llist, *llist, bool) {
	var names []lob
	var values []lob
	for bindings != EMPTY_LIST {
		tmp := car(bindings)
		if isList(tmp) {
			name := car(tmp)
			if isSymbol(name) {
				names = append(names, name)
				tmp2 := cdr(tmp)
				if tmp2 != EMPTY_LIST {
					val, err := macroexpandObject(module, car(tmp2))
					if err == nil {
						values = append(values, val)
						bindings = cdr(bindings)
						continue
					}
				}
			}
		}
		return nil, nil, false
	}
	return toList(names), toList(values), true
}

func expandLet(expr *llist) (*llist, error) {
	module := currentModule
	// (let () expr ...) -> (begin expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((lambda (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (lambda (label) expr
	if isSymbol(cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return expandNamedLet(expr)
	}
	bindings, ok := cadr(expr).(*llist)
	if !ok {
		return nil, syntaxError(expr)
	}
	names, values, ok := crackLetBindings(module, bindings)
	if !ok {
		return nil, syntaxError(expr)
	}
	body := cddr(expr)
	if body == EMPTY_LIST {
		return nil, syntaxError(expr)
	}
	code, err := macroexpand(module, cons(intern("lambda"), cons(names, body)))
	if err != nil {
		return nil, err
	}
	return cons(code, values), nil
}

func expandNamedLet(expr *llist) (*llist, error) {
	module := currentModule
	name := cadr(expr)
	bindings, ok := caddr(expr).(*llist)
	if !ok {
		return nil, syntaxError(expr)
	}
	names, values, ok := crackLetBindings(module, bindings)
	if !ok {
		println("cannot expandNamedLet")
		return nil, syntaxError(expr)
	}
	body := cdddr(expr)
	tmp := list(intern("letrec"), list(list(name, cons(intern("lambda"), cons(names, body)))), cons(name, values))
	return macroexpand(module, tmp)
}

func crackDoBindings(module module, bindings *llist) (*llist, *llist, *llist, bool) {
	names := EMPTY_LIST
	inits := EMPTY_LIST
	steps := EMPTY_LIST
	for bindings != EMPTY_LIST {
		tmp := car(bindings)
		if !isList(tmp) {
			return nil, nil, nil, false
		}
		if !isSymbol(car(tmp)) {
			return nil, nil, nil, false
		}
		if !isList(cdr(tmp)) {
			return nil, nil, nil, false
		}
		names = cons(car(tmp), names)
		inits = cons(cadr(tmp), inits)
		if cddr(tmp) != EMPTY_LIST {
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
	var tmp lob
	var tmpl, tmpl2 *llist
	if length(expr) < 3 {
		return nil, syntaxError(expr)
	}
	
	bindings, ok := cadr(expr).(*llist)
	if !ok {
		return nil, syntaxError(expr)
	}
	names, inits, steps, ok := crackDoBindings(currentModule, bindings)
	if !ok {
		return nil, syntaxError(expr)
	}
	tmp = caddr(expr)
	if !isList(tmp) {
		return nil, syntaxError(expr)
	}
	tmpl = tmp.(*llist)
	exitPred := car(tmpl)
	var exitExprs lob = NIL
	if cddr(tmpl) != EMPTY_LIST {
		exitExprs = cons(intern("begin"), cdr(tmpl))
	} else {
		exitExprs = cadr(tmpl)
	}
	loopSym := intern("system_loop")
	if cdddr(expr) != EMPTY_LIST {
		tmpl = cdddr(expr)
		tmpl = cons(intern("begin"), tmpl)
		tmpl2 = cons(loopSym, steps)
		tmpl2 = list(tmpl2)
		tmpl = cons(intern("begin"), cons(tmpl, tmpl2))
	} else {
		tmpl = cons(loopSym, steps)
	}
	tmpl = list(tmpl)
	tmpl = cons(intern("if"), cons(exitPred, cons(exitExprs, tmpl)))
	tmpl = list(intern("lambda"), names, tmpl)
	tmpl = list(loopSym, tmpl)
	tmpl = list(tmpl)
	tmpl2 = cons(loopSym, inits)
	tmpl = list(intern("letrec"), tmpl, tmpl2)
	return macroexpand(currentModule, tmpl)
}

func expandCond(expr lob) (lob, error) {
	return nil, newError("expandCond NYI")
}
