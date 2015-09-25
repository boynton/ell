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

func compile(expr *LAny) (*LAny, error) {
	target := newCode(0, nil, nil, "")
	err := compileExpr(target, EmptyList, expr, false, false, "")
	if err != nil {
		return nil, err
	}
	target.code.emitReturn()
	return target, nil
}

func calculateLocation(sym *LAny, env *LAny) (int, int, bool) {
	i := 0
	for env != EmptyList {
		j := 0
		ee := car(env)
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

func compileExpr(target *LAny, env *LAny, expr *LAny, isTail bool, ignoreResult bool, context string) error {
	if isKeyword(expr) || isType(expr) {
		if !ignoreResult {
			target.code.emitLiteral(expr)
			if isTail {
				target.code.emitReturn()
			}
		}
		return nil
	} else if isSymbol(expr) {
		if getMacro(expr) != nil {
			return Error("Cannot use macro as a value: ", expr)
		}
		if i, j, ok := calculateLocation(expr, env); ok {
			target.code.emitLocal(i, j)
		} else {
			target.code.emitGlobal(expr)
		}
		if ignoreResult {
			target.code.emitPop()
		} else if isTail {
			target.code.emitReturn()
		}
		return nil
	} else if isList(expr) {
		if expr == EmptyList {
			if !ignoreResult {
				target.code.emitLiteral(expr)
				if isTail {
					target.code.emitReturn()
				}
			}
			return nil
		}
		lst := expr
		lstlen := length(lst)
		if lstlen == 0 {
			return SyntaxError(lst)
		}
		fn := car(lst)
		switch fn {
		case intern("quote"):
			// (quote <datum>)
			if lstlen != 2 {
				return SyntaxError(expr)
			}
			if !ignoreResult {
				target.code.emitLiteral(cadr(lst))
				if isTail {
					target.code.emitReturn()
				}
			}
			return nil
		case intern("do"): // a sequence of expressions, for side-effect only
			// (do <expr> ...)
			return compileSequence(target, env, cdr(lst), isTail, ignoreResult, context)
		case intern("if"):
			// (if <pred> <consequent>)
			// (if <pred> <consequent> <antecedent>)
			if lstlen == 3 || lstlen == 4 {
				return compileIfElse(target, env, cadr(expr), caddr(expr), cdddr(expr), isTail, ignoreResult, context)
			}
			return SyntaxError(expr)
		case intern("def"):
			// (def <name> <val>)
			if lstlen < 3 {
				return SyntaxError(expr)
			}
			sym := cadr(lst)
			val := caddr(lst)
			err := compileExpr(target, env, val, false, false, sym.String())
			if err == nil {
				target.code.emitDefGlobal(sym)
				if ignoreResult {
					target.code.emitPop()
				} else if isTail {
					target.code.emitReturn()
				}
			}
			return err
		case intern("undef"):
			if lstlen != 2 {
				return SyntaxError(expr)
			}
			sym := cadr(lst)
			if !isSymbol(sym) {
				return SyntaxError(expr)
			}
			target.code.emitUndefGlobal(sym)
			if ignoreResult {
			} else {
				target.code.emitLiteral(sym)
				if isTail {
					target.code.emitReturn()
				}
			}
			return nil
		case intern("defmacro"):
			// (defmacro <name> (fn args & body))
			if lstlen != 3 {
				return SyntaxError(expr)
			}
			var sym = cadr(lst)
			if !isSymbol(sym) {
				return SyntaxError(expr)
			}
			err := compileExpr(target, env, caddr(expr), false, false, sym.String())
			if err != nil {
				return err
			}
			if err == nil {
				target.code.emitDefMacro(sym)
				if ignoreResult {
					target.code.emitPop()
				} else if isTail {
					target.code.emitReturn()
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
				return SyntaxError(expr)
			}
			body := cddr(lst)
			args := cadr(lst)
			return compileFn(target, env, args, body, isTail, ignoreResult, context)
		case intern("set!"):
			// (set! <sym> <val>)
			if lstlen != 3 {
				return SyntaxError(expr)
			}
			var sym = cadr(lst)
			if !isSymbol(sym) {
				return SyntaxError(expr)
			}
			err := compileExpr(target, env, caddr(lst), false, false, context)
			if err != nil {
				return err
			}
			if i, j, ok := calculateLocation(sym, env); ok {
				target.code.emitSetLocal(i, j)
			} else {
				target.code.emitDefGlobal(sym) //fix, should be SetGlobal
			}
			if ignoreResult {
				target.code.emitPop()
			} else if isTail {
				target.code.emitReturn()
			}
			return nil
		case intern("code"):
			// (code <instruction> ...)
			return target.code.loadOps(cdr(expr))
		case intern("use"):
			// (use module_name)
			return compileUse(target, cdr(lst))
		default: // a funcall
			// (<fn>)
			// (<fn> <arg> ...)
			return compileFuncall(target, env, fn, cdr(lst), isTail, ignoreResult, context)
		}
	} else if isVector(expr) {
		//vector literal: the elements are evaluated
		vlen := len(expr.elements)
		for i := vlen - 1; i >= 0; i-- {
			obj := expr.elements[i]
			err := compileExpr(target, env, obj, false, false, context)
			if err != nil {
				return err
			}
		}
		target.code.emitVector(vlen)
		if isTail {
			target.code.emitReturn()
		}
		return nil
	} else if isStruct(expr) {
		//struct literal: the elements are evaluated
		vlen := len(expr.elements)
		vals := make([]*LAny, 0, vlen)
		for _, b := range expr.elements {
			vals = append(vals, b)
		}
		for i := vlen - 1; i >= 0; i-- {
			obj := vals[i]
			err := compileExpr(target, env, obj, false, false, context)
			if err != nil {
				return err
			}
		}
		target.code.emitStruct(vlen)
		if isTail {
			target.code.emitReturn()
		}
		return nil
	}
	if !ignoreResult {
		target.code.emitLiteral(expr)
		if isTail {
			target.code.emitReturn()
		}
	}
	return nil
}

func compileFn(target *LAny, env *LAny, args *LAny, body *LAny, isTail bool, ignoreResult bool, context string) error {
	argc := 0
	var syms []*LAny
	var defaults []*LAny
	var keys []*LAny
	tmp := args
	rest := false
	if !isSymbol(args) {
		for tmp != EmptyList {
			a := car(tmp)
			if isVector(a) {
				//i.e. (x [y (z 23)]) is for optional y and z, but bound, z with default 23
				if cdr(tmp) != EmptyList {
					return SyntaxError(tmp)
				}
				defaults = make([]*LAny, 0, len(a.elements))
				for _, sym := range a.elements {
					def := Null
					if isList(sym) {
						def = cadr(sym)
						sym = car(sym)
					}
					if !isSymbol(sym) {
						return SyntaxError(tmp)
					}
					syms = append(syms, sym)
					defaults = append(defaults, def)
				}
				tmp = EmptyList
				break
			} else if isStruct(a) {
				//i.e. (x {y: 23, z: 57}]) is for optional y and z, keyword args, with defaults
				if cdr(tmp) != EmptyList {
					return SyntaxError(tmp)
				}
				elen := len(a.elements)
				slen := elen / 2
				defaults = make([]*LAny, 0, slen)
				keys = make([]*LAny, 0, slen)
				for i := 0; i < elen; i += 2 {
					sym := a.elements[i]
					defValue := a.elements[i+1]
					if isList(sym) && car(sym) == intern("quote") && cdr(sym) != EmptyList {
						sym = cadr(sym)
					} else {
						var err error
						sym, err = unkeyworded(sym) //returns sym itself if not a keyword, otherwise strips the colon
						if err != nil {             //not a symbol or keyword
							return SyntaxError(tmp)
						}
					}
					if !isSymbol(sym) {
						return SyntaxError(tmp)
					}
					syms = append(syms, sym)
					keys = append(keys, sym)
					defaults = append(defaults, defValue)
				}
				tmp = EmptyList
				break
			} else if !isSymbol(a) {
				return SyntaxError(tmp)
			}
			if a == intern("&") { //the rest of the arglist is bound to a single variable
				//note that the & annotation is optional if  what follows is a struct or vector
				rest = true
			} else {
				if rest {
					syms = append(syms, a) //note: added, but argv not incremented
					defaults = make([]*LAny, 0)
					tmp = EmptyList
					break
				}
				argc++
				syms = append(syms, a)
			}
			tmp = cdr(tmp)
		}
	}
	if tmp != EmptyList { //entire arglist bound to a single variable
		if isSymbol(tmp) {
			syms = append(syms, tmp) //note: added, but argv not incremented
			defaults = make([]*LAny, 0)
		} else {
			return SyntaxError(tmp)
		}
	}
	args = listFromValues(syms) //why not just use the vector format in general?
	newEnv := cons(args, env)
	fnCode := newCode(argc, defaults, keys, context)
	err := compileSequence(fnCode, newEnv, body, true, false, context)
	if err == nil {
		if !ignoreResult {
			target.code.emitClosure(fnCode)
			if isTail {
				target.code.emitReturn()
			}
		}
	}
	return err
}

func compileSequence(target *LAny, env *LAny, exprs *LAny, isTail bool, ignoreResult bool, context string) error {
	if exprs != EmptyList {
		for cdr(exprs) != EmptyList {
			err := compileExpr(target, env, car(exprs), false, true, context)
			if err != nil {
				return err
			}
			exprs = cdr(exprs)
		}
		return compileExpr(target, env, car(exprs), isTail, ignoreResult, context)
	}
	return SyntaxError(cons(intern("do"), exprs))
}

func compileFuncall(target *LAny, env *LAny, fn *LAny, args *LAny, isTail bool, ignoreResult bool, context string) error {
	argc := length(args)
	if argc < 0 {
		return SyntaxError(cons(fn, args))
	}
	err := compileArgs(target, env, args, context)
	if err != nil {
		return err
	}
	fval := global(fn)
	if fval != nil && fval.ltype == typeFunction && fval.function.primitive != nil {
		target.code.emitPrimCall(fval.function.primitive, argc)
		if ignoreResult {
			target.code.emitPop()
		} else if isTail {
			target.code.emitReturn()
		}
		return nil
	}
	err = compileExpr(target, env, fn, false, false, context)
	if err != nil {
		return err
	}
	if isTail {
		target.code.emitTailCall(argc)
	} else {
		target.code.emitCall(argc)
		if ignoreResult {
			target.code.emitPop()
		}
	}
	return nil
}

func compileArgs(target *LAny, env *LAny, args *LAny, context string) error {
	if args != EmptyList {
		err := compileArgs(target, env, cdr(args), context)
		if err != nil {
			return err
		}
		return compileExpr(target, env, car(args), false, false, context)
	}
	return nil
}

func compileIfElse(target *LAny, env *LAny, predicate *LAny, consequent *LAny, antecedentOptional *LAny, isTail bool, ignoreResult bool, context string) error {
	antecedent := Null
	if antecedentOptional != EmptyList {
		antecedent = car(antecedentOptional)
	}
	err := compileExpr(target, env, predicate, false, false, context)
	if err != nil {
		return err
	}
	loc1 := target.code.emitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = compileExpr(target, env, consequent, isTail, ignoreResult, context)
	if err != nil {
		return err
	}
	loc2 := 0
	if !isTail {
		loc2 = target.code.emitJump(0)
	}
	target.code.setJumpLocation(loc1)
	err = compileExpr(target, env, antecedent, isTail, ignoreResult, context)
	if err == nil {
		if !isTail {
			target.code.setJumpLocation(loc2)
		}
	}
	return err
}

func compileUse(target *LAny, rest *LAny) error {
	lstlen := length(rest)
	if lstlen != 1 {
		//to do: other options for use.
		return SyntaxError(cons(intern("use"), rest))
	}
	sym := car(rest)
	if !isSymbol(sym) {
		return SyntaxError(rest)
	}
	target.code.emitUse(sym)
	return nil
}

//SyntaxError returns an error indicating that the given expression has bad syntax
func SyntaxError(expr *LAny) error {
	return Error("Syntax error: ", expr)
}
