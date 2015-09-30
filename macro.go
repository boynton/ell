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

type macro struct {
	name     *LOB
	expander *LOB //a function of one argument
}

func newMacro(name *LOB, expander *LOB) *macro {
	return &macro{name, expander}
}

func (mac *macro) String() string {
	return fmt.Sprintf("(macro %v %v)", mac.name, mac.expander)
}

func macroexpand(expr *LOB) (*LOB, error) {
	return macroexpandObject(expr)
}

func macroexpandObject(expr *LOB) (*LOB, error) {
	if isList(expr) {
		if expr != EmptyList {
			return macroexpandList(expr)
		}
	}
	return expr, nil
}

func macroexpandList(expr *LOB) (*LOB, error) {
	if expr == EmptyList {
		return expr, nil
	}
	lst := expr
	fn := car(lst)
	head := fn
	if isSymbol(fn) {
		result, err := expandPrimitive(fn, lst)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		head = fn
	} else if isList(fn) {
		expanded, err := macroexpandList(fn)
		if err != nil {
			return nil, err
		}
		head = expanded
	}
	tail, err := expandSequence(cdr(expr))
	if err != nil {
		return nil, err
	}
	return cons(head, tail), nil
}

func (mac *macro) expand(expr *LOB) (*LOB, error) {
	expander := mac.expander
	if expander.variant == typeFunction {
		if expander.function.code != nil {
			if expander.function.code.argc == 1 {
				expanded, err := execCompileTime(expander.function.code, expr)
				if err == nil {
					if isList(expanded) {
						return macroexpandObject(expanded)
					}
					return expanded, err
				}
				return nil, err
			}
		} else if expander.function.primitive != nil {
			args := []*LOB{expr}
			expanded, err := expander.function.primitive.fun(args)
			if err == nil {
				return macroexpandObject(expanded)
			}
			return nil, err
		}
	}
	return nil, Error(MacroErrorKey, "Bad macro expander function: ", expander)
}

func expandSequence(seq *LOB) (*LOB, error) {
	var result []*LOB
	if seq == nil {
		panic("Whoops: should be (), not nil!")
	}
	for seq != EmptyList {
		item := car(seq)
		if isList(item) {
			expanded, err := macroexpandList(item)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded)
		} else {
			result = append(result, item)
		}
		seq = cdr(seq)
	}
	lst := listFromValues(result)
	if seq != EmptyList {
		tmp := cons(seq, EmptyList)
		return concat(lst, tmp)
	}
	return lst, nil
}

func expandIf(expr *LOB) (*LOB, error) {
	i := length(expr)
	if i == 4 {
		tmp, err := expandSequence(cdr(expr))
		if err != nil {
			return nil, err
		}
		return cons(car(expr), tmp), nil
	} else if i == 3 {
		tmp := list(cadr(expr), caddr(expr), Null)
		tmp, err := expandSequence(tmp)
		if err != nil {
			return nil, err
		}
		return cons(car(expr), tmp), nil
	} else {
		return nil, Error(SyntaxErrorKey, expr)
	}
}

func expandUndef(expr *LOB) (*LOB, error) {
	if length(expr) != 2 || !isSymbol(cadr(expr)) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	return expr, nil
}

// (defn f (x) (+ 1 x))
//  ->
// (def f (fn (x) (+ 1 x)))
func expandDefn(expr *LOB) (*LOB, error) {
	exprLen := length(expr)
	if exprLen >= 4 {
		name := cadr(expr)
		if isSymbol(name) {
			args := caddr(expr)
			body, err := expandSequence(cdddr(expr))
			if err != nil {
				return nil, err
			}
			tmp, err := expandFn(cons(intern("fn"), cons(args, body)))
			if err != nil {
				return nil, err
			}
			return list(intern("def"), name, tmp), nil
		}
	}
	return nil, Error(SyntaxErrorKey, expr)
}

func expandDefmacro(expr *LOB) (*LOB, error) {
	exprLen := length(expr)
	if exprLen >= 4 {
		name := cadr(expr)
		if isSymbol(name) {
			args := caddr(expr)
			body, err := expandSequence(cdddr(expr))
			if err != nil {
				return nil, err
			}
			//(fn (expr) (apply xxx
			tmp, err := expandFn(cons(intern("fn"), cons(args, body))) //this is the expander with special args\
			if err != nil {
				return nil, err
			}
			sym := intern("expr")
			tmp, err = expandFn(list(intern("fn"), list(sym), list(intern("apply"), tmp, list(intern("cdr"), sym))))
			if err != nil {
				return nil, err
			}
			return list(intern("defmacro"), name, tmp), nil
		}
	}
	return nil, Error(SyntaxErrorKey, expr)
}

//(defmacro (defmacro expr)
//  `(defmacro ~(cadr expr) (fn (expr) (apply (fn ~(caddr expr) ~@(cdddr expr)) (cdr expr)))))

func expandDef(expr *LOB) (*LOB, error) {
	exprLen := length(expr)
	if exprLen != 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	name := cadr(expr)
	if !isSymbol(name) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	if exprLen > 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := caddr(expr)
	if !isList(body) {
		return expr, nil
	}
	val, err := macroexpandList(body)
	if err != nil {
		return nil, err
	}
	return list(car(expr), name, val), nil
}

func expandFn(expr *LOB) (*LOB, error) {
	exprLen := length(expr)
	if exprLen < 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body, err := expandSequence(cddr(expr))
	if err != nil {
		return nil, err
	}
	bodyLen := length(body)
	if bodyLen > 0 {
		tmp := body
		if isList(tmp) && caar(tmp) == intern("def") || caar(tmp) == intern("defmacro") {
			bindings := EmptyList
			for caar(tmp) == intern("def") || caar(tmp) == intern("defmacro") {
				if caar(tmp) == intern("defmacro") {
					return nil, Error(MacroErrorKey, "macros can only be defined at top level")
				}
				def, err := expandDef(car(tmp))
				if err != nil {
					return nil, err
				}
				bindings = cons(cdr(def), bindings)
				tmp = cdr(tmp)
			}
			bindings = reverse(bindings)
			tmp = cons(intern("letrec"), cons(bindings, tmp)) //scheme specifies letrec*
			tmp2, err := macroexpandList(tmp)
			return list(car(expr), cadr(expr), tmp2), err
		}
	}
	args := cadr(expr)
	return cons(car(expr), cons(args, body)), nil
}

func expandSetBang(expr *LOB) (*LOB, error) {
	exprLen := length(expr)
	if exprLen != 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	var val = caddr(expr)
	if isList(val) {
		v, err := macroexpandList(val)
		if err != nil {
			return nil, err
		}
		val = v
	}
	return list(car(expr), cadr(expr), val), nil
}

func expandPrimitive(fn *LOB, expr *LOB) (*LOB, error) {
	switch fn {
	case intern("quote"):
		return expr, nil
	case intern("do"):
		return expandSequence(expr)
	case intern("if"):
		return expandIf(expr)
	case intern("def"):
		return expandDef(expr)
	case intern("undef"):
		return expandUndef(expr)
	case intern("defn"):
		return expandDefn(expr)
	case intern("defmacro"):
		return expandDefmacro(expr)
	case intern("fn"):
		return expandFn(expr)
	case intern("set!"):
		return expandSetBang(expr)
	case intern("lap"):
		return expr, nil
	case intern("use"):
		return expr, nil
	default:
		macro := getMacro(fn)
		if macro != nil {
			tmp, err := macro.expand(expr)
			return tmp, err
		}
		return nil, nil
	}
}

func crackLetrecBindings(bindings *LOB, tail *LOB) (*LOB, *LOB, bool) {
	var names []*LOB
	inits := EmptyList
	for bindings != EmptyList {
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
					inits = cons(cons(intern("set!"), tmp), inits)
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
	inits = reverse(inits)
	head := inits
	for inits.cdr != EmptyList {
		inits = inits.cdr
	}
	inits.cdr = tail
	return listFromValues(names), head, true
}

func expandLetrec(expr *LOB) (*LOB, error) {
	// (letrec () expr ...) -> (do expr ...)
	// (letrec ((x 1) (y 2)) expr ...) -> ((fn (x y) (set! x 1) (set! y 2) expr ...) nil nil)
	body := cddr(expr)
	if body == EmptyList {
		return nil, Error(SyntaxErrorKey, expr)
	}
	bindings := cadr(expr)
	if !isList(bindings) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, body, ok := crackLetrecBindings(bindings, body)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	code, err := macroexpandList(cons(intern("fn"), cons(names, body)))
	if err != nil {
		return nil, err
	}
	values := newList(length(names), Null)
	return cons(code, values), nil
}

func crackLetBindings(bindings *LOB) (*LOB, *LOB, bool) {
	var names []*LOB
	var values []*LOB
	for bindings != EmptyList {
		tmp := car(bindings)
		if isList(tmp) {
			name := car(tmp)
			if isSymbol(name) {
				names = append(names, name)
				tmp2 := cdr(tmp)
				if tmp2 != EmptyList {
					val, err := macroexpandObject(car(tmp2))
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
	return listFromValues(names), listFromValues(values), true
}

func expandLet(expr *LOB) (*LOB, error) {
	// (let () expr ...) -> (do expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((fn (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (fn (label) expr
	if isSymbol(cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return expandNamedLet(expr)
	}
	bindings := cadr(expr)
	if !isList(bindings) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, values, ok := crackLetBindings(bindings)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := cddr(expr)
	if body == EmptyList {
		return nil, Error(SyntaxErrorKey, expr)
	}
	code, err := macroexpandList(cons(intern("fn"), cons(names, body)))
	if err != nil {
		return nil, err
	}
	return cons(code, values), nil
}

func expandNamedLet(expr *LOB) (*LOB, error) {
	name := cadr(expr)
	bindings := caddr(expr)
	if !isList(bindings) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, values, ok := crackLetBindings(bindings)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := cdddr(expr)
	tmp := list(intern("letrec"), list(list(name, cons(intern("fn"), cons(names, body)))), cons(name, values))
	return macroexpandList(tmp)
}

func crackDoBindings(bindings *LOB) (*LOB, *LOB, *LOB, bool) {
	names := EmptyList
	inits := EmptyList
	steps := EmptyList
	for bindings != EmptyList {
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
		if cddr(tmp) != EmptyList {
			steps = cons(caddr(tmp), steps)
		} else {
			steps = cons(car(tmp), steps)
		}
		bindings = cdr(bindings)
	}
	var err error
	inits, err = macroexpandList(inits)
	if err != nil {
		return nil, nil, nil, false
	}
	steps, err = macroexpandList(steps)
	if err != nil {
		return nil, nil, nil, false
	}
	return names, inits, steps, true
}

func nextCondClause(expr *LOB, clauses *LOB, count int) (*LOB, error) {
	var result *LOB
	var err error
	tmpsym := intern("__tmp__")
	ifsym := intern("if")
	elsesym := intern("else")
	letsym := intern("let")
	dosym := intern("do")

	clause0 := car(clauses)
	next := cdr(clauses)
	clause1 := car(next)

	if count == 2 {
		if !isList(clause1) {
			return nil, Error(SyntaxErrorKey, expr)
		}
		if elsesym == car(clause1) {
			if cadr(clause0) == intern("=>") {
				if length(clause0) != 3 {
					return nil, Error(SyntaxErrorKey, expr)
				}
				result = list(letsym, list(list(tmpsym, car(clause0))), list(ifsym, tmpsym, list(caddr(clause0), tmpsym), cons(dosym, cdr(clause1))))
			} else {
				result = list(ifsym, car(clause0), cons(dosym, cdr(clause0)), cons(dosym, cdr(clause1)))
			}
		} else {
			if cadr(clause1) == intern("=>") {
				if length(clause1) != 3 {
					return nil, Error(SyntaxErrorKey, expr)
				}
				result = list(letsym, list(list(tmpsym, car(clause1))), list(ifsym, tmpsym, list(caddr(clause1), tmpsym), clause1))
			} else {
				result = list(ifsym, car(clause1), cons(dosym, cdr(clause1)))
			}
			if cadr(clause0) == intern("=>") {
				if length(clause0) != 3 {
					return nil, Error(SyntaxErrorKey, expr)
				}
				result = list(letsym, list(list(tmpsym, car(clause0))), list(ifsym, tmpsym, list(caddr(clause0), tmpsym), result))
			} else {
				result = list(ifsym, car(clause0), cons(dosym, cdr(clause0)), result)
			}
		}
	} else {
		result, err = nextCondClause(expr, next, count-1)
		if err != nil {
			return nil, err
		}
		if cadr(clause0) == intern("=>") {
			if length(clause0) != 3 {
				return nil, Error(SyntaxErrorKey, expr)
			}
			result = list(letsym, list(list(tmpsym, car(clause0))), list(ifsym, tmpsym, list(caddr(clause0), tmpsym), result))
		} else {
			result = list(ifsym, car(clause0), cons(dosym, cdr(clause0)), result)
		}
	}
	return macroexpand(result)
}

func expandCond(expr *LOB) (*LOB, error) {
	i := length(expr)
	if i < 2 {
		return nil, Error(SyntaxErrorKey, expr)
	} else if i == 2 {
		tmp := cadr(expr)
		if car(tmp) == intern("else") {
			tmp = cons(intern("do"), cdr(tmp))
		} else {
			expr = cons(intern("do"), cdr(tmp))
			tmp = list(intern("if"), car(tmp), expr)
		}
		return macroexpand(tmp)
	} else {
		return nextCondClause(expr, cdr(expr), i-1)
	}
}

func expandQuasiquote(expr *LOB) (*LOB, error) {
	if length(expr) != 2 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	return expandQQ(cadr(expr))
}

func expandQQ(expr *LOB) (*LOB, error) {
	switch expr.variant {
	case typeList:
		if expr == EmptyList {
			return expr, nil
		}
		if expr.cdr != EmptyList {
			if expr.car == symUnquote {
				if expr.cdr.cdr != EmptyList {
					return nil, Error(SyntaxErrorKey, expr)
				}
				return macroexpand(expr.cdr.car)
			} else if expr.car == symUnquoteSplicing {
				return nil, Error(MacroErrorKey, "unquote-splicing can only occur in the context of a list ")
			}
		}
		tmp, err := expandQQList(expr)
		if err != nil {
			return nil, err
		}
		return macroexpand(tmp)
	case typeSymbol:
		return list(intern("quote"), expr), nil
	default: //all other objects evaluate to themselves
		return expr, nil
	}
}

func expandQQList(lst *LOB) (*LOB, error) {
	var tmp *LOB
	var err error
	result := list(intern("concat"))
	tail := result
	for lst != EmptyList {
		item := car(lst)
		if isList(item) && item != EmptyList {
			if car(item) == symQuasiquote {
				return nil, Error(MacroErrorKey, "nested quasiquote not supported")
			}
			if car(item) == symUnquote && length(item) == 2 {
				tmp, err = macroexpand(cadr(item))
				tmp = list(intern("list"), tmp)
				if err != nil {
					return nil, err
				}
				tail.cdr = list(tmp)
				tail = tail.cdr
			} else if car(item) == symUnquoteSplicing && length(item) == 2 {
				tmp, err = macroexpand(cadr(item))
				if err != nil {
					return nil, err
				}
				tail.cdr = list(tmp)
				tail = tail.cdr
			} else {
				tmp, err = expandQQList(item)
				if err != nil {
					return nil, err
				}
				tail.cdr = list(list(intern("list"), tmp))
				tail = tail.cdr
			}
		} else {
			tail.cdr = list(list(intern("quote"), list(item)))
			tail = tail.cdr
		}
		lst = cdr(lst)
	}
	return result, nil
}
