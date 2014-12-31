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

var extendedInstructions bool = false

func EnableExtendedInstructions(b bool) {
	extendedInstructions = b
}

func Compile(module LModule, expr LObject) (LCode, error) {
	code := NewCode(module, 0, nil, nil)
	err := compileExpr(code, NIL, expr, false, false)
	if err != nil {
		return nil, err
	}
	code.EmitReturn()
	return code, nil
}

func calculateLocation(sym LObject, env LObject) (int, int, bool) {
	i := 0
	for IsPair(env) {
		j := 0
		e := Car(env)
		for IsPair(e) {
			if Car(e) == sym {
				return i, j, true
			}
			j++
			e = Cdr(e)
		}
		i++
		env = Cdr(env)
	}
	return -1, -1, false
}

func compileExpr(code LCode, env LObject, expr LObject, isTail bool, ignoreResult bool) error {
	//Println("COMPILE: ", expr, " isTail: ", isTail, ", ignoreResult: ", ignoreResult)
	if IsSymbol(expr) {
		if i, j, ok := calculateLocation(expr, env); ok {
			code.EmitLocal(i, j)
		} else {
			code.EmitGlobal(expr)
		}
		if ignoreResult {
			code.EmitPop()
		} else if isTail {
			code.EmitReturn()
		}
		return nil
	} else if IsPair(expr) {
		lst := expr
		lstlen := Length(lst)
		if lstlen == 0 {
			return Error("Cannot compile empty list: ", lst)
		}
		fn := Car(lst)
		switch fn {
		case Intern("quote"):
			// (quote <datum>)
			if lstlen != 2 {
				return Error("Syntax error: ", expr)
			}
			if !ignoreResult {
				code.EmitLiteral(Cadr(lst))
				if isTail {
					code.EmitReturn()
				}
			}
			return nil
		case Intern("begin"):
			// (begin <expr> ...)
			return compileSequence(code, env, Cdr(lst), isTail, ignoreResult)
		case Intern("if"):
			// (if <pred> <consequent>)
			// (if <pred> <consequent> <antecedent>)
			if lstlen == 3 || lstlen == 4 {
				return compileIfElse(code, env, Cadr(expr), Caddr(expr), Cdddr(expr), isTail, ignoreResult)
			} else {
				return Error("Syntax error: ", expr)
			}
		case Intern("define"):
			// (define <name> <val>)
			if lstlen != 3 {
				return Error("Syntax error: ", expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				return Error("Syntax error: ", expr)
			}
			err := compileExpr(code, env, Caddr(lst), false, false)
			if err == nil {
				code.EmitDefGlobal(sym)
				if ignoreResult {
					code.EmitPop()
				} else if isTail {
					code.EmitReturn()
				}
			}
			return err
		case Intern("define-macro"):
			// (defmacro <name> (lambda (expr) '(the expanded value)))
			if lstlen != 3 {
				return Error("Syntax error: ", expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				return Error("Syntax error: ", expr)
			}
			err := compileExpr(code, env, Caddr(lst), false, false)
			if err == nil {
				code.EmitDefMacro(sym)
				if ignoreResult {
					code.EmitPop()
				} else if isTail {
					code.EmitReturn()
				}
			}
			return err
		case Intern("lambda"):
			// (lambda ()  <expr> ...)
			// (lambda (sym ...)  <expr> ...)
			// (lambda (sym ... . rest)  <expr> ...)
			// (lambda sym <expr> ...) ;; all args in a list, bound to sym
			if lstlen < 3 {
				return Error("Syntax error: ", expr)
			}
			body := Cddr(lst)
			args := Cadr(lst)
			return compileLambda(code, env, args, body, isTail, ignoreResult)
		case Intern("set!"):
			// (set! <sym> <val>)
			if lstlen != 3 {
				return Error("Wrong number of arguments to set!: ", expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				return Error("Non-symbol first argument to set!: ", expr)
			}
			err := compileExpr(code, env, Caddr(lst), false, false)
			if err != nil {
				return err
			}
			if i, j, ok := calculateLocation(sym, env); ok {
				code.EmitSetLocal(i, j)
			} else {
				code.EmitDefGlobal(sym) //fix, should be SetGlobal!!!
			}
			if ignoreResult {
				code.EmitPop()
			} else if isTail {
				code.EmitReturn()
			}
			return nil
		case Intern("lap"):
			// (lap <instruction> ...)
			return code.LoadOps(Cdr(expr))
		case Intern("use"):
			// (use module_name)
			return compileUse(code, Cdr(lst))
		default: // a funcall
			// (<fn>)
			// (<fn> <arg> ...)
			return compileFuncall(code, env, fn, Cdr(lst), isTail, ignoreResult)
		}
	} else if vec, ok := expr.(*lvector); ok {
		//vector literal: the elements are evaluated
		vlen := len(vec.elements)
		for i := vlen - 1; i >= 0; i-- {
			obj := vec.elements[i]
			err := compileExpr(code, env, obj, false, false)
			if err != nil {
				return err
			}
		}
		code.EmitVector(vlen)
		return nil
	} else if amap, ok := expr.(*lmap); ok {
		//vector literal: the elements are evaluated
		mlen := len(amap.bindings)
		vlen := mlen * 2
		vals := make([]LObject, 0, vlen)
		for k, v := range amap.bindings {
			vals = append(vals, k)
			vals = append(vals, v)
		}
		for i := vlen - 1; i >= 0; i-- {
			obj := vals[i]
			err := compileExpr(code, env, obj, false, false)
			if err != nil {
				return err
			}
		}
		code.EmitMap(vlen)
		return nil
	} else {
		if !ignoreResult {
			code.EmitLiteral(expr)
			if isTail {
				code.EmitReturn()
			}
		}
		return nil
	}
}

func compileLambda(code LCode, env LObject, args LObject, body LObject, isTail bool, ignoreResult bool) error {
	argc := 0
	syms := []LObject{}
	var defaults []LObject
	var keys []LObject
	tmp := args
	//to do: deal with rest, optional, and keywords arguments
	for IsPair(tmp) {
		a := Car(tmp)
		if vec, ok := a.(*lvector); ok {
			//i.e. (x [y (z 23)]) is for optional y and z, but bound, z with default 23
			if Cdr(tmp) != NIL {
				return Error("optional args must be the last in the list")
			}
			defaults = make([]LObject, 0, len(vec.elements))
			for _, sym := range vec.elements {
				var def LObject = NIL
				if lst, ok := sym.(*lpair); ok {
					next := lst.cdr
					sym = lst.car
					if lst2, ok := next.(*lpair); ok {
						def = lst2.car
					}
				}
				if !IsSymbol(sym) {
					return Error("Formal argument is not a symbol: ", sym)
				}
				syms = append(syms, sym)
				defaults = append(defaults, def)
			}
			tmp = NIL
			break
		} else if mp, ok := a.(*lmap); ok {
			//i.e. (x {y: 23, z: 57}]) is for optional y and z, keyword args, with defaults
			if Cdr(tmp) != NIL {
				return Error("keywords args must be the last in the list")
			}
			defaults = make([]LObject, 0, len(mp.bindings))
			keys = make([]LObject, 0, len(mp.bindings))
			for sym, defValue := range mp.bindings {
				if IsPair(sym) && Car(sym) == Intern("quote") && IsPair(Cdr(sym)) {
					sym = Cadr(sym)
				}
				if !IsSymbol(sym) {
					return Error("Formal argument is not a symbol: ", sym)
				}
				syms = append(syms, sym)
				keys = append(keys, sym)
				defaults = append(defaults, defValue)
			}
			tmp = NIL
			break
		} else if !IsSymbol(a) {
			return Error("Formal argument is not a symbol: ", a)
		} else {
			argc++
			syms = append(syms, a)
			tmp = Cdr(tmp)
		}
	}
	if tmp != NIL {
		//rest arg
		if IsSymbol(tmp) {
			syms = append(syms, tmp) //note: added, but argv not incremented
			defaults = make([]LObject, 0)
		} else {
			return Error("Formal argument list is malformed: ", tmp)
		}
	}
	args = ToList(syms) //why not just use the array format in general?
	newEnv := Cons(args, env)
	mod := (code.(*lcode)).module
	lambdaCode := NewCode(mod, argc, defaults, keys)
	err := compileSequence(lambdaCode, newEnv, body, true, false)
	if err == nil {
		if !ignoreResult {
			code.EmitClosure(lambdaCode)
			if isTail {
				code.EmitReturn()
			}
		}
	}
	return err
}

func compileSequence(code LCode, env LObject, exprs LObject, isTail bool, ignoreResult bool) error {
	if exprs != NIL {
		for Cdr(exprs) != NIL {
			err := compileExpr(code, env, Car(exprs), false, true)
			if err != nil {
				return err
			}
			exprs = Cdr(exprs)
		}
		return compileExpr(code, env, Car(exprs), isTail, ignoreResult)
	} else {
		return Error("Bad syntax: ", Cons(Intern("begin"), exprs))
	}
}

func compileFuncall(code LCode, env LObject, fn LObject, args LObject, isTail bool, ignoreResult bool) error {
	argc := Length(args)
	if argc < 0 {
		return Error("Bad funcall: ", Cons(fn, args))
	}
	err := compileArgs(code, env, args)
	if err != nil {
		return err
	}
	if extendedInstructions {
		ok, err := compilePrimopCall(code, fn, argc, isTail, ignoreResult)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	err = compileExpr(code, env, fn, false, false)
	if err != nil {
		return err
	}
	if isTail {
		code.EmitTailCall(argc)
	} else {
		code.EmitCall(argc)
		if ignoreResult {
			code.EmitPop()
		}
	}
	return nil
}

func compileArgs(code LCode, env LObject, args LObject) error {
	if args != NIL {
		err := compileArgs(code, env, Cdr(args))
		if err != nil {
			return err
		}
		return compileExpr(code, env, Car(args), false, false)
	}
	return nil
}

func compilePrimopCall(code LCode, fn LObject, argc int, isTail bool, ignoreResult bool) (bool, error) {
	switch fn {
	case Intern("car"):
		if argc != 1 {
			return false, nil
		}
		code.EmitCar()
	case Intern("cdr"):
		if argc != 1 {
			return false, nil
		}
		code.EmitCdr()
	case Intern("null?"):
		if argc != 1 {
			return false, nil
		}
		code.EmitNull()
	case Intern("+"):
		if argc != 2 {
			return false, nil
		}
		code.EmitAdd()
	case Intern("*"):
		if argc != 2 {
			return false, nil
		}
		code.EmitMul()
	default:
		return false, nil
	}
	if isTail {
		code.EmitReturn()
	} else if ignoreResult {
		code.EmitPop()
	}
	return true, nil
}

func compileIfElse(code LCode, env LObject, predicate LObject, consequent LObject, antecedentOptional LObject, isTail bool, ignoreResult bool) error {
	var antecedent LObject = NIL
	if antecedentOptional != NIL {
		antecedent = Car(antecedentOptional)
	}
	err := compileExpr(code, env, predicate, false, false)
	if err != nil {
		return err
	}
	loc1 := code.EmitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = compileExpr(code, env, consequent, isTail, ignoreResult)
	if err != nil {
		return err
	}
	loc2 := 0
	if !isTail {
		loc2 = code.EmitJump(0)
	}
	code.SetJumpLocation(loc1)
	err = compileExpr(code, env, antecedent, isTail, ignoreResult)
	if err == nil {
		if !isTail {
			code.SetJumpLocation(loc2)
		}
	}
	return err
}

func compileUse(code LCode, rest LObject) error {
	lstlen := Length(rest)
	if lstlen != 1 {
		//to do: other options for use.
		return Error("Wrong number of arguments to use: ", Cons(Intern("use"), rest))
	}
	sym := Car(rest)
	if !IsSymbol(sym) {
		return Error("Non-symbol first argument to use: ", sym)
	}
	code.EmitUse(sym)
	return nil
}
