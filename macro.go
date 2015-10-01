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
	name     LOB
	expander LOB //a function of one argument
}

func newMacro(name LOB, expander LOB) *macro {
	return &macro{name, expander}
}

func (mac *macro) String() string {
	return fmt.Sprintf("(macro %v %v)", mac.name, mac.expander)
}

func macroexpand(expr LOB) (LOB, error) {
	return macroexpandObject(expr)
}

func macroexpandObject(expr LOB) (LOB, error) {
	if lst, ok := expr.(*LList); ok {
		if lst != EmptyList {
			return macroexpandList(lst)
		}
	}
	return expr, nil
}

func macroexpandList(expr *LList) (LOB, error) {
	if expr == EmptyList {
		return expr, nil
	}
	lst := expr
	fn := car(lst)
	head := fn
	switch t := fn.(type) {
	case *LSymbol:
		result, err := expandPrimitive(t, lst)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}
		head = fn
	case *LList:
		expanded, err := macroexpandList(t)
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

func (mac *macro) expand(expr LOB) (LOB, error) {
	if expander, ok := mac.expander.(*LFunction); ok {
		if expander.code != nil {
			if expander.code.argc == 1 {
				expanded, err := execCompileTime(expander.code, expr)
				if err == nil {
					if isList(expanded) {
						return macroexpandObject(expanded)
					}
					return expanded, err
				}
				return nil, err
			}
		} else if expander.primitive != nil {
			args := []LOB{expr}
			expanded, err := expander.primitive.fun(args)
			if err == nil {
				return macroexpandObject(expanded)
			}
			return nil, err
		}
	}
	return nil, Error(MacroErrorKey, "Bad macro expander function: ", mac.expander)
}

func expandSequence(seq *LList) (*LList, error) {
	var result []LOB
	if seq == nil {
		panic("Whoops: should be (), not nil!")
	}
	for seq != EmptyList {
		item := car(seq)
		if lstItem, ok := item.(*LList); ok {
			expanded, err := macroexpandList(lstItem)
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

func expandIf(expr *LList) (LOB, error) {
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

func expandUndef(expr *LList) (LOB, error) {
	if length(expr) != 2 || !isSymbol(cadr(expr)) {
		return nil, Error(SyntaxErrorKey, expr)
	}
	return expr, nil
}

// (defn f (x) (+ 1 x))
//  ->
// (def f (fn (x) (+ 1 x)))
func expandDefn(expr *LList) (*LList, error) {
	exprLen := listLength(expr)
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

func expandDefmacro(expr *LList) (LOB, error) {
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

func expandDef(expr *LList) (*LList, error) {
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
	if lstBody, ok := body.(*LList); ok {
		val, err := macroexpandList(lstBody)
		if err != nil {
			return nil, err
		}
		return list(car(expr), name, val), nil
	}
	return expr, nil
}

func expandFn(expr *LList) (LOB, error) {
	exprLen := listLength(expr)
	if exprLen < 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body, err := expandSequence(cddr(expr))
	if err != nil {
		return nil, err
	}
	bodyLen := listLength(body)
	if bodyLen > 0 {
		if firstExpr, ok := car(body).(*LList); ok {
			if car(firstExpr) == intern("def") || car(firstExpr) == intern("defmacro") {
				tmp := body
				bindings := EmptyList
				for car(firstExpr) == intern("def") || car(firstExpr) == intern("defmacro") {
					if car(firstExpr) == intern("defmacro") {
						return nil, Error(MacroErrorKey, "macros can only be defined at top level")
					}
					def, err := expandDef(firstExpr)
					if err != nil {
						return nil, err
					}
					bindings = cons(cdr(def), bindings)
					tmp = cdr(tmp)
					firstExpr, ok = car(tmp).(*LList)
				}
				bindings = reverse(bindings)
				tmp = cons(intern("letrec"), cons(bindings, tmp)) //scheme specifies letrec*
				tmp2, err := macroexpandList(tmp)
				return list(car(expr), cadr(expr), tmp2), err
			}
		}
	}
	args := cadr(expr)
	return cons(car(expr), cons(args, body)), nil
}

func expandSetBang(expr *LList) (*LList, error) {
	exprLen := length(expr)
	if exprLen != 3 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	var val = caddr(expr)
	if lstVal, ok := val.(*LList); ok {
		v, err := macroexpandList(lstVal)
		if err != nil {
			return nil, err
		}
		val = v
	}
	return list(car(expr), cadr(expr), val), nil
}

func expandPrimitive(fn *LSymbol, expr *LList) (LOB, error) {
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

func crackLetrecBindings(bindings *LList, tail *LList) (*LList, *LList, bool) {
	var names []LOB
	inits := EmptyList
	for bindings != EmptyList {
		if isList(bindings) {
			tmp := car(bindings)
			if lstTmp, ok := tmp.(*LList); ok && lstTmp != EmptyList {
				name := car(lstTmp)
				if isSymbol(name) {
					names = append(names, name)
				} else {
					return nil, nil, false
				}
				if cdr(lstTmp) != EmptyList {
					inits = cons(cons(intern("set!"), lstTmp), inits)
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
	for cdr(inits) != EmptyList {
		inits = cdr(inits)
	}
	inits.cdr = tail
	return listFromValues(names), head, true
}

func expandLetrec(expr *LList) (*LList, error) {
	// (letrec () expr ...) -> (do expr ...)
	// (letrec ((x 1) (y 2)) expr ...) -> ((fn (x y) (set! x 1) (set! y 2) expr ...) nil nil)
	body := cddr(expr)
	if body == EmptyList {
		return nil, Error(SyntaxErrorKey, expr)
	}
	bindings, ok := cadr(expr).(*LList)
	if !ok {
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

func crackLetBindings(bindings *LList) (*LList, *LList, bool) {
	var names []LOB
	var values []LOB
	for bindings != EmptyList {
		if tmp, ok := car(bindings).(*LList); ok {
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

func expandLet(expr *LList) (*LList, error) {
	// (let () expr ...) -> (do expr ...)
	// (let ((x 1) (y 2)) expr ...) -> ((fn (x y) expr ...) 1 2)
	// (let label ((x 1) (y 2)) expr ...) -> (fn (label) expr
	if isSymbol(cadr(expr)) {
		//return ell_expand_named_let(argv, argc)
		return expandNamedLet(expr)
	}
	bindings, ok := cadr(expr).(*LList)
	if !ok {
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

func expandNamedLet(expr *LList) (*LList, error) {
	name := cadr(expr)
	bindings, ok := caddr(expr).(*LList)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	names, values, ok := crackLetBindings(bindings)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	body := cdddr(expr)
	tmp := list(intern("letrec"), list(list(name, cons(intern("fn"), cons(names, body)))), cons(name, values))
	tmp2, err := macroexpandList(tmp)
	if err != nil {
		return nil, err
	}
	tmp, _ = tmp2.(*LList) //we know it always expands to a list
	return tmp, nil
}

func nextCondClause(expr *LList, clauses *LList, count int) (*LList, error) {
	var result *LList
	var err error
	tmpsym := intern("__tmp__")
	ifsym := intern("if")
	elsesym := intern("else")
	letsym := intern("let")
	dosym := intern("do")

	clause0, ok := car(clauses).(*LList)
	if !ok {
		return nil, Error(SyntaxErrorKey, expr)
	}
	next := cdr(clauses)

	if count == 2 {
		clause1, ok := car(next).(*LList)
		if !ok {
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
	res, err := macroexpand(result)
	if err != nil {
		return nil, err
	}
	result, _ = res.(*LList)
	return result, nil
}

func expandCond(expr *LList) (*LList, error) {
	i := listLength(expr)
	if i < 2 {
		return nil, Error(SyntaxErrorKey, expr)
	} else if i == 2 {
		tmp, ok := cadr(expr).(*LList)
		if !ok {
			return nil, Error(SyntaxErrorKey, expr)
		}
		if car(tmp) == intern("else") {
			tmp = cons(intern("do"), cdr(tmp))
		} else {
			expr = cons(intern("do"), cdr(tmp))
			tmp = list(intern("if"), car(tmp), expr)
		}
		res, err := macroexpand(tmp)
		if err != nil {
			return nil, err
		}
		result, _ := res.(*LList)
		return result, nil
	} else {
		return nextCondClause(expr, cdr(expr), i-1)
	}
}

func expandQuasiquote(expr *LList) (LOB, error) {
	if listLength(expr) != 2 {
		return nil, Error(SyntaxErrorKey, expr)
	}
	return expandQQ(cadr(expr))
}

func expandQQ(expr LOB) (LOB, error) {
	switch t := expr.(type) {
	case *LList:
		if t == EmptyList {
			return expr, nil
		}
		if cdr(t) != EmptyList {
			if car(t) == symUnquote {
				if cddr(t) != EmptyList {
					return nil, Error(SyntaxErrorKey, expr)
				}
				return macroexpand(cadr(t))
			} else if car(t) == symUnquoteSplicing {
				return nil, Error(MacroErrorKey, "unquote-splicing can only occur in the context of a list ")
			}
		}
		tmp, err := expandQQList(t)
		if err != nil {
			return nil, err
		}
		return macroexpand(tmp)
	case *LSymbol:
		return list(intern("quote"), expr), nil
	default: //all other objects evaluate to themselves
		return expr, nil
	}
}

func expandQQList(lst *LList) (LOB, error) {
	var tmp LOB
	var err error
	result := list(intern("concat"))
	tail := result
	for lst != EmptyList {
		item := car(lst)
		if lstItem, ok := item.(*LList); ok && lstItem != EmptyList {
			if car(lstItem) == symQuasiquote {
				return nil, Error(MacroErrorKey, "nested quasiquote not supported")
			}
			if car(lstItem) == symUnquote && listLength(lstItem) == 2 {
				tmp, err = macroexpand(cadr(lstItem))
				tmp = list(intern("list"), tmp)
				if err != nil {
					return nil, err
				}
				tail.cdr = list(tmp)
				tail = tail.cdr
			} else if car(lstItem) == symUnquoteSplicing && listLength(lstItem) == 2 {
				tmp, err = macroexpand(cadr(lstItem))
				if err != nil {
					return nil, err
				}
				tail.cdr = list(tmp)
				tail = tail.cdr
			} else {
				tmp, err = expandQQList(lstItem)
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
