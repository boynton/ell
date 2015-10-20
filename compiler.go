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

// Compile - compile the source into a code object.
func Compile(expr *Object) (*Object, error) {
	target := MakeCode(0, nil, nil, "")
	err := compileExpr(target, EmptyList, expr, false, false, "")
	if err != nil {
		return nil, err
	}
	target.code.emitReturn()
	return target, nil
}

func calculateLocation(sym *Object, env *Object) (int, int, bool) {
	i := 0
	for env != EmptyList {
		j := 0
		ee := Car(env)
		for ee != EmptyList {
			if Car(ee) == sym {
				return i, j, true
			}
			j++
			ee = Cdr(ee)
		}
		i++
		env = Cdr(env)
	}
	return -1, -1, false
}

func compileExpr(target *Object, env *Object, expr *Object, isTail bool, ignoreResult bool, context string) error {
	if IsKeyword(expr) || IsType(expr) {
		if !ignoreResult {
			target.code.emitLiteral(expr)
			if isTail {
				target.code.emitReturn()
			}
		}
		return nil
	} else if IsSymbol(expr) {
		if GetMacro(expr) != nil {
			return Error(Intern("macro-error"), "Cannot use macro as a value: ", expr)
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
	} else if IsList(expr) {
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
		lstlen := ListLength(lst)
		if lstlen == 0 {
			return Error(SyntaxErrorKey, lst)
		}
		fn := Car(lst)
		switch fn {
		case Intern("quote"):
			// (quote <datum>)
			if lstlen != 2 {
				return Error(SyntaxErrorKey, expr)
			}
			if !ignoreResult {
				target.code.emitLiteral(Cadr(lst))
				if isTail {
					target.code.emitReturn()
				}
			}
			return nil
		case Intern("do"): // a sequence of expressions, for side-effect only
			// (do <expr> ...)
			return compileSequence(target, env, Cdr(lst), isTail, ignoreResult, context)
		case Intern("if"):
			// (if <pred> <Consequent>)
			// (if <pred> <Consequent> <antecedent>)
			if lstlen == 3 || lstlen == 4 {
				return compileIfElse(target, env, Cadr(expr), Caddr(expr), Cdddr(expr), isTail, ignoreResult, context)
			}
			return Error(SyntaxErrorKey, expr)
		case Intern("def"):
			// (def <name> <val>)
			if lstlen < 3 {
				return Error(SyntaxErrorKey, expr)
			}
			sym := Cadr(lst)
			val := Caddr(lst)
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
		case Intern("undef"):
			if lstlen != 2 {
				return Error(SyntaxErrorKey, expr)
			}
			sym := Cadr(lst)
			if !IsSymbol(sym) {
				return Error(SyntaxErrorKey, expr)
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
		case Intern("defmacro"):
			// (defmacro <name> (fn args & body))
			if lstlen != 3 {
				return Error(SyntaxErrorKey, expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				return Error(SyntaxErrorKey, expr)
			}
			err := compileExpr(target, env, Caddr(expr), false, false, sym.String())
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
				return Error(SyntaxErrorKey, expr)
			}
			body := Cddr(lst)
			args := Cadr(lst)
			return compileFn(target, env, args, body, isTail, ignoreResult, context)
		case Intern("set!"):
			// (set! <sym> <val>)
			if lstlen != 3 {
				return Error(SyntaxErrorKey, expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				return Error(SyntaxErrorKey, expr)
			}
			err := compileExpr(target, env, Caddr(lst), false, false, context)
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
		case Intern("code"):
			// (code <instruction> ...)
			return target.code.loadOps(Cdr(expr))
		case Intern("use"):
			// (use module_name)
			return compileUse(target, Cdr(lst))
		default: // a funcall
			// (<fn>)
			// (<fn> <arg> ...)
			fn, args := optimizeFuncall(target, env, fn, Cdr(lst), isTail, ignoreResult, context)
			return compileFuncall(target, env, fn, args, isTail, ignoreResult, context)
		}
	} else if IsVector(expr) {
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
	} else if IsStruct(expr) {
		//struct literal: the elements are evaluated
		vlen := len(expr.bindings) * 2
		vals := make([]*Object, 0, vlen)
		for k, v := range expr.bindings {
			vals = append(vals, k.toObject())
			vals = append(vals, v)
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

func compileFn(target *Object, env *Object, args *Object, body *Object, isTail bool, ignoreResult bool, context string) error {
	argc := 0
	var syms []*Object
	var defaults []*Object
	var keys []*Object
	tmp := args
	rest := false
	if !IsSymbol(args) {
		for tmp != EmptyList {
			a := Car(tmp)
			if IsVector(a) {
				//i.e. (x [y (z 23)]) is for optional y and z, but bound, z with default 23
				if Cdr(tmp) != EmptyList {
					return Error(SyntaxErrorKey, tmp)
				}
				defaults = make([]*Object, 0, len(a.elements))
				for _, sym := range a.elements {
					def := Null
					if IsList(sym) {
						def = Cadr(sym)
						sym = Car(sym)
					}
					if !IsSymbol(sym) {
						return Error(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					defaults = append(defaults, def)
				}
				tmp = EmptyList
				break
			} else if IsStruct(a) {
				//i.e. (x {y: 23, z: 57}]) is for optional y and z, keyword args, with defaults
				if Cdr(tmp) != EmptyList {
					return Error(SyntaxErrorKey, tmp)
				}
				slen := len(a.bindings)
				defaults = make([]*Object, 0, slen)
				keys = make([]*Object, 0, slen)
				for k, defValue := range a.bindings {
					sym := k.toObject()
					if IsList(sym) && Car(sym) == Intern("quote") && Cdr(sym) != EmptyList {
						sym = Cadr(sym)
					} else {
						var err error
						sym, err = unkeyworded(sym) //returns sym itself if not a keyword, otherwise strips the colon
						if err != nil {             //not a symbol or keyword
							return Error(SyntaxErrorKey, tmp)
						}
					}
					if !IsSymbol(sym) {
						return Error(SyntaxErrorKey, tmp)
					}
					syms = append(syms, sym)
					keys = append(keys, sym)
					defaults = append(defaults, defValue)
				}
				tmp = EmptyList
				break
			} else if !IsSymbol(a) {
				return Error(SyntaxErrorKey, tmp)
			}
			if a == Intern("&") { //the rest of the arglist is bound to a single variable
				//note that the & annotation is optional if  what follows is a struct or vector
				rest = true
			} else {
				if rest {
					syms = append(syms, a) //note: added, but argv not incremented
					defaults = make([]*Object, 0)
					tmp = EmptyList
					break
				}
				argc++
				syms = append(syms, a)
			}
			tmp = Cdr(tmp)
		}
	}
	if tmp != EmptyList { //entire arglist bound to a single variable
		if IsSymbol(tmp) {
			syms = append(syms, tmp) //note: added, but argv not incremented
			defaults = make([]*Object, 0)
		} else {
			return Error(SyntaxErrorKey, tmp)
		}
	}
	args = ListFromValues(syms) //why not just use the vector format in general?
	newEnv := Cons(args, env)
	fnCode := MakeCode(argc, defaults, keys, context)
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

func compileSequence(target *Object, env *Object, exprs *Object, isTail bool, ignoreResult bool, context string) error {
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
	return Error(SyntaxErrorKey, Cons(Intern("do"), exprs))
}

func optimizeFuncall(target *Object, env *Object, fn *Object, args *Object, isTail bool, ignoreResult bool, context string) (*Object, *Object) {
	size := ListLength(args)
	if size == 2 {
		switch fn {
		case Intern("+"):
			if Equal(One, Car(args)) { // (+ 1 x) ->  inc x)
				return Intern("inc"), Cdr(args)
			} else if Equal(One, Cadr(args)) { // (+ x 1) -> (inc x)
				return Intern("inc"), List(Car(args))
			}
		case Intern("-"):
			if Equal(One, Cadr(args)) { // (- x 1) -> (dec x)
				return Intern("dec"), List(Car(args))
			}
		}
	}
	return fn, args
}

func compileFuncall(target *Object, env *Object, fn *Object, args *Object, isTail bool, ignoreResult bool, context string) error {
	argc := ListLength(args)
	if argc < 0 {
		return Error(SyntaxErrorKey, Cons(fn, args))
	}
	//fprim := global(fn)
	//if fprim != nil && fprim.primitive != nil && !isTail { ... something more optimized. We know the function signature. }
	err := compileArgs(target, env, args, context)
	if err != nil {
		return err
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

func compileArgs(target *Object, env *Object, args *Object, context string) error {
	if args != EmptyList {
		err := compileArgs(target, env, Cdr(args), context)
		if err != nil {
			return err
		}
		return compileExpr(target, env, Car(args), false, false, context)
	}
	return nil
}

func compileIfElse(target *Object, env *Object, predicate *Object, Consequent *Object, antecedentOptional *Object, isTail bool, ignoreResult bool, context string) error {
	antecedent := Null
	if antecedentOptional != EmptyList {
		antecedent = Car(antecedentOptional)
	}
	err := compileExpr(target, env, predicate, false, false, context)
	if err != nil {
		return err
	}
	loc1 := target.code.emitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = compileExpr(target, env, Consequent, isTail, ignoreResult, context)
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

func compileUse(target *Object, rest *Object) error {
	lstlen := ListLength(rest)
	if lstlen != 1 {
		//to do: other options for use.
		return Error(SyntaxErrorKey, Cons(Intern("use"), rest))
	}
	sym := Car(rest)
	if !IsSymbol(sym) {
		return Error(SyntaxErrorKey, rest)
	}
	target.code.emitUse(sym)
	return nil
}
