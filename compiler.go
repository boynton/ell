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

func compile(expr LOB) (*LCode, error) {
	code := newCode(0, nil, nil, "")
	err := compileExpr(code, EmptyList, expr, false, false, "")
	if err != nil {
		return nil, err
	}
	code.emitReturn()
	return code, nil
}

func calculateLocation(sym LOB, env *LList) (int, int, bool) {
	i := 0
	for env != EmptyList {
		j := 0
		ee := car(env).(*LList)
		for ee != EmptyList {
			if car(ee) == sym {
				return i, j, true
			}
			j++
			ee = cdr(ee)
		}
		i++
		env = cdr(env)
	}
	return -1, -1, false
}

func compileExpr(code *LCode, env *LList, expr LOB, isTail bool, ignoreResult bool, context string) error {
	switch t := expr.(type) {
	case *LSymbol:
		if getMacro(expr) != nil {
			return Error(intern("macro-error"), "Cannot use macro as a value: ", expr)
		}
		if i, j, ok := calculateLocation(expr, env); ok {
			code.emitLocal(i, j)
		} else {
			code.emitGlobal(expr)
		}
		if ignoreResult {
			code.emitPop()
		} else if isTail {
			code.emitReturn()
		}
		return nil
	case *LList:
		lst := t
		if lst == EmptyList {
			if !ignoreResult {
				code.emitLiteral(expr)
				if isTail {
					code.emitReturn()
				}
			}
			return nil
		}
		lstlen := listLength(lst)
		if lstlen == 0 {
			return Error(SyntaxErrorKey, lst)
		}
		fn := car(lst)
		switch fn {
		case intern("quote"):
			// (quote <datum>)
			if lstlen != 2 {
				return Error(SyntaxErrorKey, expr)
			}
			if !ignoreResult {
				code.emitLiteral(cadr(lst))
				if isTail {
					code.emitReturn()
				}
			}
			return nil
		case intern("do"): // a sequence of expressions, for side-effect only
			// (do <expr> ...)
			return compileSequence(code, env, cdr(lst), isTail, ignoreResult, context)
		case intern("if"):
			// (if <pred> <consequent>)
			// (if <pred> <consequent> <antecedent>)
			if lstlen == 3 || lstlen == 4 {
				return compileIfElse(code, env, cadr(lst), caddr(lst), cdddr(lst), isTail, ignoreResult, context)
			}
			return Error(SyntaxErrorKey, expr)
		case intern("def"):
			// (def <name> <val>)
			if lstlen < 3 {
				return Error(SyntaxErrorKey, expr)
			}
			sym := cadr(lst)
			val := caddr(lst)
			err := compileExpr(code, env, val, false, false, sym.String())
			if err == nil {
				code.emitDefGlobal(sym)
				if ignoreResult {
					code.emitPop()
				} else if isTail {
					code.emitReturn()
				}
			}
			return err
		case intern("undef"):
			if lstlen != 2 {
				return Error(SyntaxErrorKey, expr)
			}
			sym, ok := cadr(lst).(*LSymbol)
			if !ok {
				return Error(SyntaxErrorKey, expr)
			}
			code.emitUndefGlobal(sym)
			if ignoreResult {
			} else {
				code.emitLiteral(sym)
				if isTail {
					code.emitReturn()
				}
			}
			return nil
		case intern("defmacro"):
			// (defmacro <name> (fn args & body))
			if lstlen != 3 {
				return Error(SyntaxErrorKey, expr)
			}
			sym, ok := cadr(lst).(*LSymbol)
			if !ok {
				return Error(SyntaxErrorKey, expr)
			}
			err := compileExpr(code, env, caddr(lst), false, false, sym.String())
			if err != nil {
				return err
			}
			if err == nil {
				code.emitDefMacro(sym)
				if ignoreResult {
					code.emitPop()
				} else if isTail {
					code.emitReturn()
				}
			}
			return err
		case intern("fn"):
			// (fn ()  <expr> ...)
			// (fn (sym ...)  <expr> ...) ;; binds arguments to successive syms
			// (fn (sym ... & rsym)  <expr> ...) ;; all args after the & are collected and bound to rsym
			// (fn (sym ... [sym sym])  <expr> ...) ;; all args up to the vector are required, the rest are optional
			// (fn (sym ... [(sym val) sym])  <expr> ...) ;; default values can be provided to optional args
			// (fn (sym ... {sym: def sym: def})  <expr> ...) ;; required args, then keyword args
			// (fn (& sym)  <expr> ...) ;; all args in a list, bound to sym. Same as the following form.
			// (fn sym <expr> ...) ;; all args in a list, bound to sym
			if lstlen < 3 {
				return Error(SyntaxErrorKey, expr)
			}
			body := cddr(lst)
			args := cadr(lst)
			return compileFn(code, env, args, body, isTail, ignoreResult, context)
		case intern("set!"):
			// (set! <sym> <val>)
			if lstlen != 3 {
				return Error(SyntaxErrorKey, expr)
			}
			sym, ok := cadr(lst).(*LSymbol)
			if !ok {
				return Error(SyntaxErrorKey, expr)
			}
			err := compileExpr(code, env, caddr(lst), false, false, context)
			if err != nil {
				return err
			}
			if i, j, ok := calculateLocation(sym, env); ok {
				code.emitSetLocal(i, j)
			} else {
				code.emitDefGlobal(sym) //fix, should be SetGlobal
			}
			if ignoreResult {
				code.emitPop()
			} else if isTail {
				code.emitReturn()
			}
			return nil
		case intern("code"):
			// (code <instruction> ...)
			return code.loadOps(cdr(lst))
		case intern("use"):
			// (use module_name)
			return compileUse(code, cdr(lst))
		default: // a funcall
			// (<fn>)
			// (<fn> <arg> ...)
			fn, args := optimizeFuncall(code, env, fn, cdr(lst), isTail, ignoreResult, context)
			return compileFuncall(code, env, fn, args, isTail, ignoreResult, context)
		}
	case *LVector:
		//vector literal: the elements are evaluated
		vlen := len(t.elements)
		for i := vlen - 1; i >= 0; i-- {
			obj := t.elements[i]
			err := compileExpr(code, env, obj, false, false, context)
			if err != nil {
				return err
			}
		}
		code.emitVector(vlen)
		if isTail {
			code.emitReturn()
		}
		return nil
	case *LStruct:
		//struct literal: the elements are evaluated
		vlen := len(t.elements)
		vals := make([]LOB, 0, vlen)
		for _, b := range t.elements {
			vals = append(vals, b)
		}
		for i := vlen - 1; i >= 0; i-- {
			obj := vals[i]
			err := compileExpr(code, env, obj, false, false, context)
			if err != nil {
				return err
			}
		}
		code.emitStruct(vlen)
		if isTail {
			code.emitReturn()
		}
		return nil
	default:
		if !ignoreResult {
			code.emitLiteral(expr)
			if isTail {
				code.emitReturn()
			}
		}
	}
	return nil
}

func compileFn(code *LCode, env *LList, args LOB, body *LList, isTail bool, ignoreResult bool, context string) error {
	argc := 0
	var syms []LOB
	var defaults []LOB
	var keys []LOB
	rest := false
	tmp2 := LOB(EmptyList)
	if tmp, ok := args.(*LList); ok {
		for tmp != EmptyList {
			a := car(tmp)
			switch t := a.(type) {
			case *LVector:
				//i.e. (x [y (z 23)]) is for optional y and z, but bound, z with default 23
				if cdr(tmp) != EmptyList {
					return Error(SyntaxErrorKey, tmp)
				}
				defaults = make([]LOB, 0, len(t.elements))
				for _, sym := range t.elements {
					def := LOB(Null)
					if lst, ok := sym.(*LList); ok {
						def = cadr(lst)
						sym = car(lst)
					}
					if !isSymbol(sym) {
						return Error(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					defaults = append(defaults, def)
				}
				tmp = EmptyList
				break //continue?
			case *LStruct:
				//i.e. (x {y: 23, z: 57}]) is for optional y and z, keyword args, with defaults
				if cdr(tmp) != EmptyList {
					return Error(SyntaxErrorKey, tmp)
				}
				elen := len(t.elements)
				slen := elen / 2
				defaults = make([]LOB, 0, slen)
				keys = make([]LOB, 0, slen)
				for i := 0; i < elen; i += 2 {
					sym := t.elements[i]
					defValue := t.elements[i+1]
					if lst, ok := sym.(*LList); ok && car(lst) == intern("quote") && cdr(lst) != EmptyList {
						sym = cadr(lst)
					} else {
						var err error
						sym, err = unkeyworded(sym) //returns sym itself if not a keyword, otherwise strips the colon
						if err != nil {             //not a symbol or keyword
							return Error(SyntaxErrorKey, tmp)
						}
					}
					if !isSymbol(sym) {
						return Error(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					keys = append(keys, sym)
					defaults = append(defaults, defValue)
				}
				tmp = EmptyList
				break
			case *LSymbol:
				//nothing
			default:
				return Error(SyntaxErrorKey, tmp)
			}
			if tmp == EmptyList {
				break
			}
			if a == intern("&") { //the rest of the arglist is bound to a single variable
				//note that the & annotation is optional if  what follows is a struct or vector
				rest = true
			} else {
				if rest {
					syms = append(syms, a) //note: added, but argv not incremented
					defaults = make([]LOB, 0)
					tmp = EmptyList
					break
				}
				argc++
				syms = append(syms, a)
			}
			tmp = cdr(tmp)
		}
		tmp2 = tmp
	} else {
		tmp2 = args
	}
	if tmp2 != EmptyList { //entire arglist bound to a single variable
		if isSymbol(tmp2) {
			syms = append(syms, tmp2) //note: added, but argv not incremented
			defaults = make([]LOB, 0)
		} else {
			return Error(SyntaxErrorKey, tmp2)
		}
	}
	args = listFromValues(syms) //why not just use the vector format in general?
	newEnv := cons(args, env)
	fnCode := newCode(argc, defaults, keys, context)
	err := compileSequence(fnCode, newEnv, body, true, false, context)
	if err == nil {
		if !ignoreResult {
			code.emitClosure(fnCode)
			if isTail {
				code.emitReturn()
			}
		}
	}
	return err
}

func compileSequence(code *LCode, env *LList, exprs *LList, isTail bool, ignoreResult bool, context string) error {
	if exprs != EmptyList {
		for cdr(exprs) != EmptyList {
			err := compileExpr(code, env, car(exprs), false, true, context)
			if err != nil {
				return err
			}
			exprs = cdr(exprs)
		}
		return compileExpr(code, env, car(exprs), isTail, ignoreResult, context)
	}
	return Error(SyntaxErrorKey, cons(intern("do"), exprs))
}

func optimizeFuncall(code *LCode, env *LList, fn LOB, args *LList, isTail bool, ignoreResult bool, context string) (LOB, *LList) {
	size := length(args)
	if size == 2 {
		if fn == intern("+") {
			//(+ 1 x) == (+ x 1) == (1+ x)
			if isEqual(One, car(args)) {
				return intern("inc"), cdr(args)
			} else if isEqual(One, cadr(args)) {
				return intern("inc"), list(car(args))
			}
		}
		//other things to collapse?
	}
	return fn, args
}

func compileFuncall(code *LCode, env *LList, fn LOB, args *LList, isTail bool, ignoreResult bool, context string) error {
	argc := length(args)
	if argc < 0 {
		return Error(SyntaxErrorKey, cons(fn, args))
	}
	err := compileArgs(code, env, args, context)
	if err != nil {
		return err
	}
	//fval := global(fn)
	//if the function is defined, we can do some additional compile-time argc check. In a dynamic env, though, not a win
	err = compileExpr(code, env, fn, false, false, context)
	if err != nil {
		return err
	}
	if isTail {
		code.emitTailCall(argc)
	} else {
		code.emitCall(argc)
		if ignoreResult {
			code.emitPop()
		}
	}
	return nil
}

func compileArgs(code *LCode, env *LList, args *LList, context string) error {
	if args != EmptyList {
		err := compileArgs(code, env, cdr(args), context)
		if err != nil {
			return err
		}
		return compileExpr(code, env, car(args), false, false, context)
	}
	return nil
}

func compileIfElse(code *LCode, env *LList, predicate LOB, consequent LOB, antecedentOptional *LList, isTail bool, ignoreResult bool, context string) error {
	antecedent := LOB(Null)
	if antecedentOptional != EmptyList {
		antecedent = car(antecedentOptional)
	}
	err := compileExpr(code, env, predicate, false, false, context)
	if err != nil {
		return err
	}
	loc1 := code.emitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = compileExpr(code, env, consequent, isTail, ignoreResult, context)
	if err != nil {
		return err
	}
	loc2 := 0
	if !isTail {
		loc2 = code.emitJump(0)
	}
	code.setJumpLocation(loc1)
	err = compileExpr(code, env, antecedent, isTail, ignoreResult, context)
	if err == nil {
		if !isTail {
			code.setJumpLocation(loc2)
		}
	}
	return err
}

func compileUse(code *LCode, rest *LList) error {
	lstlen := listLength(rest)
	if lstlen != 1 {
		//to do: other options for use.
		return Error(SyntaxErrorKey, cons(intern("use"), rest))
	}
	var name *LString
	switch t := car(rest).(type) {
	case *LSymbol:
		name = newString(t.text)
	case *LString:
		name = t
	default:
		return Error(SyntaxErrorKey, rest)
	}
	code.emitUse(name)
	return nil
}
