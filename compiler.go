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

import(
		"fmt"
	. "github.com/boynton/ell/data"
)

var _ = fmt.Println

// Compile - compile the source into a code object.
func Compile(expr Value) (*Code, error) {
	target := MakeCode(0, nil, nil, "")
	err := compileExpr(target, EmptyList, expr, false, false, "")
	if err != nil {
		return nil, err
	}
	target.emitReturn()
	return target, nil
}

func calculateLocation(sym Value, env *List) (int, int, bool) {
	i := 0
	for env != EmptyList {
		j := 0
		ee := env.Car
		for ee != EmptyList {
			if Car(ee) == sym {
				return i, j, true
			}
			j++
			ee = Cdr(ee)
		}
		i++
		env = env.Cdr
	}
	return -1, -1, false
}

func compileSelfEvalLiteral(target *Code, expr Value, isTail bool, ignoreResult bool) error {
	if !ignoreResult {
		target.emitLiteral(expr)
		if isTail {
			target.emitReturn()
		}
	}
	return nil
}

func compileSymbol(target *Code, env *List, expr Value, isTail bool, ignoreResult bool) error {
	if GetMacro(expr) != nil {
		return NewError(Intern("macro-error"), "Cannot use macro as a value: ", expr)
	}
	if i, j, ok := calculateLocation(expr, env); ok {
		target.emitLocal(i, j)
	} else {
		target.emitGlobal(expr)
	}
	if ignoreResult {
		target.emitPop()
	} else if isTail {
		target.emitReturn()
	}
	return nil
}

func compileQuote(target *Code, expr Value, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen != 2 {
		return NewError(SyntaxErrorKey, expr)
	}
	if !ignoreResult {
		target.emitLiteral(Cadr(expr))
		if isTail {
			target.emitReturn()
		}
	}
	return nil
}

func compileDef(target *Code, env *List, lst Value, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen < 3 {
		return NewError(SyntaxErrorKey, lst)
	}
	sym := Cadr(lst)
	val := Caddr(lst)
	err := compileExpr(target, env, val, false, false, sym.String())
	if err == nil {
		target.emitDefGlobal(sym)
		if ignoreResult {
			target.emitPop()
		} else if isTail {
			target.emitReturn()
		}
	}
	return err
}

func compileUndef(target *Code, lst Value, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen != 2 {
		return NewError(SyntaxErrorKey, lst)
	}
	sym := Cadr(lst)
	if !IsSymbol(sym) {
		return NewError(SyntaxErrorKey, lst)
	}
	target.emitUndefGlobal(sym)
	if ignoreResult {
	} else {
		target.emitLiteral(sym)
		if isTail {
			target.emitReturn()
		}
	}
	return nil
}

func compileMacro(target *Code, env *List, expr Value, isTail bool, ignoreResult bool, lstlen int) error {
	if lstlen != 3 {
		return NewError(SyntaxErrorKey, expr)
	}
	var sym = Cadr(expr)
	if !IsSymbol(sym) {
		return NewError(SyntaxErrorKey, expr)
	}
	err := compileExpr(target, env, Caddr(expr), false, false, sym.String())
	if err != nil {
		return err
	}
	if err == nil {
		target.emitDefMacro(sym)
		if ignoreResult {
			target.emitPop()
		} else if isTail {
			target.emitReturn()
		}
	}
	return err
}

func compileSet(target *Code, env *List, lst Value, isTail bool, ignoreResult bool, context string, lstlen int) error {
	if lstlen != 3 {
		return NewError(SyntaxErrorKey, lst)
	}
	var sym = Cadr(lst)
	if !IsSymbol(sym) {
		return NewError(SyntaxErrorKey, lst)
	}
	err := compileExpr(target, env, Caddr(lst), false, false, context)
	if err != nil {
		return err
	}
	if i, j, ok := calculateLocation(sym, env); ok {
		target.emitSetLocal(i, j)
	} else {
		target.emitDefGlobal(sym) //fix, should be SetGlobal
	}
	if ignoreResult {
		target.emitPop()
	} else if isTail {
		target.emitReturn()
	}
	return nil
}

func compileList(target *Code, env *List, expr Value, isTail bool, ignoreResult bool, context string) error {
	if expr == EmptyList {
		if !ignoreResult {
			target.emitLiteral(expr)
			if isTail {
				target.emitReturn()
			}
		}
		return nil
	}
	lst := expr
	lstlen := ListLength(lst)
	if lstlen == 0 {
		return NewError(SyntaxErrorKey, lst)
	}
	fn := Car(lst)
	switch fn {
	case Intern("quote"):
		// (quote <datum>)
		return compileQuote(target, expr, isTail, ignoreResult, lstlen)
	case Intern("do"): // a sequence of expressions, for side-effect only
		// (do <expr> ...)
		return compileSequence(target, env, Cdr(lst), isTail, ignoreResult, context)
	case Intern("if"):
		// (if pred consequent)
		// (if pred consequent antecedent)
		if lstlen == 3 || lstlen == 4 {
			return compileIfElse(target, env, Cadr(expr), Caddr(expr), Cdddr(expr), isTail, ignoreResult, context)
		}
		return NewError(SyntaxErrorKey, expr)
	case Intern("def"):
		// (def <name> <val>)
		return compileDef(target, env, expr, isTail, ignoreResult, lstlen)
	case Intern("undef"):
		// (undef <name>)
		return compileUndef(target, expr, isTail, ignoreResult, lstlen)
	case Intern("defmacro"):
		// (defmacro <name> (fn args & body))
		return compileMacro(target, env, expr, isTail, ignoreResult, lstlen)
	case Intern("fn"):
		// (fn ()  <expr> ...)
		// (fn (sym ...)  <expr> ...) ;; binds arguments to successive syms
		// (fn (sym ... & rsym)  <expr> ...) ;; all args after the & are collected and bound to rsym
		// (fn (sym ... [sym sym])  <expr> ...) ;; all args up to the vector are required, the rest are optional
		// (fn (sym ... [(sym val) sym])  <expr> ...) ;; default values can be provided to optional args
		// (fn (sym ... {sym: def sym: def})  <expr> ...) ;; required args, then keyword args
		// (fn (& sym)  <expr> ...) ;; all args in a list, bound to sym. Same as the following form.
		// (fn sym <expr> ...) ;; all args in a list, bound to sym
		if lstlen < 3 {
			return NewError(SyntaxErrorKey, expr)
		}
		body := Cddr(lst)
		args := Cadr(lst)
		return compileFn(target, env, args, body, isTail, ignoreResult, context)
	case Intern("set!"):
		// (set! <sym> <val>)
		return compileSet(target, env, expr, isTail, ignoreResult, context, lstlen)
	case Intern("code"):
		// (code <instruction> ...)
		return target.loadOps(Cdr(expr))
	case Intern("use"):
		// (use module_name)
		return compileUse(target, Cdr(lst))
	default: // a funcall
		// (<fn>)
		// (<fn> <arg> ...)
		fn, args := fn, Cdr(lst)
		if optimize {
			fn, args = optimizeFuncall(fn, args)
		}
		return compileFuncall(target, env, fn, args, isTail, ignoreResult, context)
	}
}

func compileVector(target *Code, env *List, vec *Vector, isTail bool, ignoreResult bool, context string) error {
	//vector literal: the elements are evaluated
	vlen := len(vec.Elements)
	for i := vlen - 1; i >= 0; i-- {
		obj := vec.Elements[i]
		err := compileExpr(target, env, obj, false, false, context)
		if err != nil {
			return err
		}
	}
	if !ignoreResult {
		target.emitVector(vlen)
		if isTail {
			target.emitReturn()
		}
	}
	return nil
}

func compileStruct(target *Code, env *List, strct *Struct, isTail bool, ignoreResult bool, context string) error {
	//struct literal: the elements are evaluated
	vlen := len(strct.Bindings) * 2
	vals := make([]Value, 0, vlen)
	for k, v := range strct.Bindings {
		vals = append(vals, k.ToValue())
		vals = append(vals, v)
	}
	for i := vlen - 1; i >= 0; i-- {
		obj := vals[i]
		err := compileExpr(target, env, obj, false, false, context)
		if err != nil {
			return err
		}
	}
	if !ignoreResult {
		target.emitStruct(vlen)
		if isTail {
			target.emitReturn()
		}
	}
	return nil
}

func compileExpr(target *Code, env *List, expr Value, isTail bool, ignoreResult bool, context string) error {
	switch p := expr.(type) {
	case *Keyword:
		return compileSelfEvalLiteral(target, expr, isTail, ignoreResult)
	case *Type:
		return compileSelfEvalLiteral(target, expr, isTail, ignoreResult)
	case *Symbol:
		return compileSymbol(target, env, p, isTail, ignoreResult)
	case *List:
		return compileList(target, env, p, isTail, ignoreResult, context)
	case *Vector:
		return compileVector(target, env, p, isTail, ignoreResult, context)
	case *Struct:
		return compileStruct(target, env, p, isTail, ignoreResult, context)
	}
	if !ignoreResult {
		target.emitLiteral(expr)
		if isTail {
			target.emitReturn()
		}
	}
	return nil
}

func compileFn(target *Code, env *List, args Value, body *List, isTail bool, ignoreResult bool, context string) error {
	argc := 0
	var syms []Value
	var defaults []Value
	var keys []Value
	tmp := args
	rest := false
	if !IsSymbol(args) {
		if IsVector(tmp) {
			//clojure style. Should this be an error?
			tmp, _ = ToList(tmp)
		}
		for tmp != EmptyList {
			a := Car(tmp)
			if vec, ok := a.(*Vector); ok {
				//i.e. (x [y (z 23)]) is for optional y and z, but bound, z with default 23
				if Cdr(tmp) != EmptyList {
					return NewError(SyntaxErrorKey, tmp)
				}
				defaults = make([]Value, 0, len(vec.Elements))
				for _, sym := range vec.Elements {
					def := Null
					if lst, ok := sym.(*List); ok {
						def = Cadr(lst)
						sym = lst.Car
					}
					if !IsSymbol(sym) {
						return NewError(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					defaults = append(defaults, def)
				}
				tmp = EmptyList
				break
			} else if strct, ok := a.(*Struct); ok {
				//i.e. (x {y: 23, z: 57}]) is for optional y and z, keyword args, with defaults
				if Cdr(tmp) != EmptyList {
					return NewError(SyntaxErrorKey, tmp)
				}
				slen := len(strct.Bindings)
				defaults = make([]Value, 0, slen)
				keys = make([]Value, 0, slen)
				for k, defValue := range strct.Bindings {
					sym := k.ToValue()
					if IsList(sym) && Car(sym) == Intern("quote") && Cdr(sym) != EmptyList {
						sym = Cadr(sym)
					} else {
						var err error
						sym, err = Unkeyworded(sym) //returns sym itself if not a keyword, otherwise strips the colon
						if err != nil {             //not a symbol or keyword
							return NewError(SyntaxErrorKey, tmp)
						}
					}
					if !IsSymbol(sym) {
						return NewError(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					keys = append(keys, sym)
					defaults = append(defaults, defValue)
				}
				tmp = EmptyList
				break
			} else if !IsSymbol(a) {
				return NewError(SyntaxErrorKey, tmp)
			}
			if a == Intern("&") { //the rest of the arglist is bound to a single variable
				//note that the & annotation is optional if  what follows is a struct or vector
				rest = true
			} else {
				if rest {
					syms = append(syms, a) //note: added, but argv not incremented
					defaults = make([]Value, 0)
					tmp = EmptyList
					break
				}
				argc++
				syms = append(syms, a)
			}
			tmp = Cdr(tmp)
		}
	}
	if tmp != EmptyList { //remainder of the arglist bound to a single variable
		if IsSymbol(tmp) {
			syms = append(syms, tmp) //note: added, but argv not incremented
			defaults = make([]Value, 0)
		} else {
			return NewError(SyntaxErrorKey, tmp)
		}
	}
	args = ListFromValues(syms) //why not just use the vector format in general?
	newEnv := Cons(args, env)
	fnCode := MakeCode(argc, defaults, keys, context)
	err := compileSequence(fnCode, newEnv, body, true, false, context)
	if err == nil {
		if !ignoreResult {
			target.emitClosure(fnCode)
			if isTail {
				target.emitReturn()
			}
		}
	}
	return err
}

func compileSequence(target *Code, env *List, exprs *List, isTail bool, ignoreResult bool, context string) error {
	if exprs != EmptyList {
		for Cdr(exprs) != EmptyList {
			err := compileExpr(target, env, Car(exprs), false, true, context)
			if err != nil {
				return err
			}
			exprs = Cdr(exprs)
		}
		return compileExpr(target, env, Car(exprs), isTail, ignoreResult, context)
	}
	return NewError(SyntaxErrorKey, Cons(Intern("do"), exprs))
}

func optimizeFuncall(fn Value, args *List) (Value, *List) {
	size := ListLength(args)
	if size == 2 {
		switch fn {
		case Intern("+"):
			if Equal(One, Car(args)) { // (+ 1 x) ->  inc x)
				return Intern("inc"), Cdr(args)
			} else if Equal(One, Cadr(args)) { // (+ x 1) -> (inc x)
				return Intern("inc"), NewList(Car(args))
			}
		case Intern("-"):
			if Equal(One, Cadr(args)) { // (- x 1) -> (dec x)
				return Intern("dec"), NewList(Car(args))
			}
		}
	}
	return fn, args
}

func compileFuncall(target *Code, env *List, fn Value, args *List, isTail bool, ignoreResult bool, context string) error {
	argc := ListLength(args)
	if argc < 0 {
		return NewError(SyntaxErrorKey, Cons(fn, args))
	}
	err := compileArgs(target, env, args, context)
	if err != nil {
		return err
	}
	err = compileExpr(target, env, fn, false, false, context)
	if err != nil {
		return err
	}
	if isTail {
		target.emitTailCall(argc)
	} else {
		target.emitCall(argc)
		if ignoreResult {
			target.emitPop()
		}
	}
	return nil
}

func compileArgs(target *Code, env *List, args Value, context string) error {
	if args != EmptyList {
		err := compileArgs(target, env, Cdr(args), context)
		if err != nil {
			return err
		}
		return compileExpr(target, env, Car(args), false, false, context)
	}
	return nil
}

func compileIfElse(target *Code, env *List, predicate Value, Consequent Value, antecedentOptional Value, isTail bool, ignoreResult bool, context string) error {
	antecedent := Null
	if antecedentOptional != EmptyList {
		antecedent = Car(antecedentOptional)
	}
	err := compileExpr(target, env, predicate, false, false, context)
	if err != nil {
		return err
	}
	loc1 := target.emitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = compileExpr(target, env, Consequent, isTail, ignoreResult, context)
	if err != nil {
		return err
	}
	loc2 := 0
	if !isTail {
		loc2 = target.emitJump(0)
	}
	target.setJumpLocation(loc1)
	err = compileExpr(target, env, antecedent, isTail, ignoreResult, context)
	if err == nil {
		if !isTail {
			target.setJumpLocation(loc2)
		}
	}
	return err
}

func compileUse(target *Code, rest *List) error {
	lstlen := ListLength(rest)
	if lstlen != 1 {
		//to do: other options for use.
		return NewError(SyntaxErrorKey, Cons(Intern("use"), rest))
	}
	sym := Car(rest)
	if !IsSymbol(sym) {
		return NewError(SyntaxErrorKey, rest)
	}
	target.emitUse(sym)
	return nil
}
