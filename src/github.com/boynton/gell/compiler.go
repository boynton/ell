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

func Compile(module LModule, expr LObject) (LCode, error) {
	code := NewCode(module, 0, NIL)
	env := NIL
	if IsList(expr) {
		if Car(expr) == Intern("lap") {
			err := code.LoadOps(Cdr(expr))
			if err != nil {
				return nil, err
			}
			return code, nil
		}
	}
	err := compileExpr(code, env, expr, false, false)
	if err != nil {
		return nil, err
	}
	code.EmitReturn()
	return code, nil
}

func calculateLocation(sym LObject, env LObject) (int, int, bool) {
	i := 0
	for env != NIL {
		j := 0
		e := Car(env)
		for e != NIL {
			if Car(e) == sym {
				return i, j, true
			}
			j++
			e = Cdr(e)
		}
		if e == sym { //a "rest" argument
			return i, j, true
		}
		i++
		env = Cdr(env)
	}
	return -1, -1, false
}

func compileExpr(code LCode, env LObject, expr LObject, isTail bool, ignoreResult bool) LError {
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
	} else if IsList(expr) {
		lst := expr
		lstlen := Length(lst)
		if lstlen == 0 {
			return Error("Cannot compile empty list:", lst)
		}
		fn := Car(lst)
		switch fn {
		case Intern("quote"):
			if lstlen != 2 {
				return Error("Syntax error:", expr)
			}
			if !ignoreResult {
				code.EmitLiteral(Cadr(lst))
				if isTail {
					code.EmitReturn()
				}
			}
			return nil
		case Intern("begin"):
			return compileSequence(code, env, Cdr(lst), isTail, ignoreResult)
		case Intern("if"):
			if lstlen == 3 || lstlen == 4 {
				return compileIfElse(code, env, Cadr(expr), Caddr(expr), Cdddr(expr), isTail, ignoreResult)
			} else {
				return Error("Syntax error:", expr)
			}
		case Intern("define"):
			if lstlen != 3 {
				return Error("Syntax error:", expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				Error("Syntax error:", expr)
			}
			//scope so we pick up "context" from current global def for syntax errors? sModuleName = lst.second()
			err := compileExpr(code, env, Caddr(lst), false, false)
			//sModuleName = sOldModuleName;
			if err == nil {
				code.EmitDefGlobal(sym)
				if ignoreResult {
					code.EmitPop()
				} else if isTail {
					code.EmitReturn()
				}
			}
			return err
		case Intern("lambda"):
			if lstlen < 3 {
				return Error("Syntax error:", expr)
			}
			body := Cddr(lst)
			args := Cadr(lst)
			if args != NIL && !IsList(args) {
				return Error("Invalid function formal argument list:", args)
			}
			return compileLambda(code, env, args, body, isTail, ignoreResult)
		case Intern("set!"):
			if lstlen != 3 {
				return Error("Wrong number of arguments to set!:", expr)
			}
			var sym = Cadr(lst)
			if !IsSymbol(sym) {
				Error("Non-symbol first argument to set!:", expr)
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
			return code.LoadOps(Cdr(expr))
		default: // a funcall
			return compileFuncall(code, env, fn, Cdr(lst), isTail, ignoreResult)
		}
		return Error("WTF")
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

func compileLambda(code LCode, env LObject, args LObject, body LObject, isTail bool, ignoreResult bool) LError {
	argc := 0
	var rest LObject = NIL
	tmp := args
	//to do: deal with rest, optional, and keywords arguments
	for tmp != NIL {
		if !IsSymbol(Car(tmp)) {
			return Error("Formal argument is not a symbol:", Car(tmp))
		}
		argc++
		tmp = Cdr(tmp)
	}
	newEnv := Cons(args, env)
	mod := (code.(*lcode)).module
	lambdaCode := NewCode(mod, argc, rest)
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

func compileSequence(code LCode, env LObject, exprs LObject, isTail bool, ignoreResult bool) LError {
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

func compileFuncall(code LCode, env LObject, fn LObject, args LObject, isTail bool, ignoreResult bool) LError {
	argc := Length(args)
	if argc < 0 {
		return Error("bad funcall:", Cons(fn, args))
	}
	err := compileArgs(code, env, args)
	if err != nil {
		return err
	}
	ok, err := compilePrimopCall(code, fn, argc, isTail, ignoreResult)
	if err != nil {
		return err
	}
	if ok {
		return nil
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

func compileArgs(code LCode, env LObject, args LObject) LError {
	if args != NIL {
		err := compileArgs(code, env, Cdr(args))
		if err != nil {
			return err
		}
		return compileExpr(code, env, Car(args), false, false)
	}
	return nil
}

func compilePrimopCall(code LCode, fn LObject, argc int, isTail bool, ignoreResult bool) (bool, LError) {
	b := false
	if Intern("car") == fn && argc == 1 {
		code.EmitCar()
		b = true
	} else if Intern("cdr") == fn && argc == 1 {
		code.EmitCdr()
		b = true
	} else if Intern("null?") == fn && argc == 1 {
		code.EmitNullP()
		b = true
	} else if Intern("+") == fn && argc == 2 {
		code.EmitAdd()
		b = true
	} else if true && Intern("*") == fn && argc == 2 {
		code.EmitMul()
		b = true
	}
	if b {
		if isTail {
			code.EmitReturn()
		} else if ignoreResult {
			code.EmitPop()
		}
	}
	return b, nil
}

func compileIfElse(code LCode, env LObject, predicate LObject, consequent LObject, antecedentOptional LObject, isTail bool, ignoreResult bool) LError {
	var antecedent LObject = NIL
	if antecedentOptional != NIL {
		antecedent = Car(antecedentOptional)
	}
	err := compileExpr(code, env, predicate, false, false)
	if err != nil {
		return nil
	}
	loc1 := code.EmitJumpFalse(0) //returns the location just *after* the jump. setJumpLocation knows this.
	err = compileExpr(code, env, consequent, isTail, ignoreResult)
	if err != nil {
		return nil
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
