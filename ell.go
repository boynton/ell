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

// the primitive functions for the languages
import (
	"fmt"
)

// Ell defines the global functions for the top level environment
func Ell(module module) {
	module.defineFunction("version", ellVersion)

	module.define("nil", Nil)
	module.define("null", Nil)
	module.define("true", True)
	module.define("false", False)
	module.define("apply", APPLY)

	module.defineMacro("let", ellLet)
	module.defineMacro("letrec", ellLetrec)
	module.defineMacro("do", ellDo)
	module.defineMacro("cond", ellCond)
	module.defineMacro("and", ellAnd)

	module.defineFunction("macroexpand", ellMacroexpand)
	module.defineFunction("type", ellType)
	module.defineFunction("equal?", ellEq)
	module.defineFunction("identical?", ellIdenticalP)
	module.defineFunction("not", ellNot)

	module.defineFunction("boolean?", ellBooleanP)
	module.defineFunction("null?", ellNullP)
	module.defineFunction("symbol?", ellSymbolP)
	module.defineFunction("keyword?", ellKeywordP)
	module.defineFunction("string?", ellStringP)
	module.defineFunction("character?", ellCharacterP)
	module.defineFunction("function?", ellFunctionP)
	module.defineFunction("eof?", ellFunctionP)

	module.defineFunction("list?", ellListP)
	module.defineFunction("cons", ellCons)
	module.defineFunction("car", ellCar)
	module.defineFunction("cdr", ellCdr)
	module.defineFunction("cadr", ellCadr)
	module.defineFunction("cddr", ellCddr)
	module.defineFunction("list", ellList)

	module.defineFunction("string", ellString)
	module.defineFunction("display", ellDisplay)
	module.defineFunction("write", ellWrite)
	module.defineFunction("newline", ellNewline)
	module.defineFunction("print", ellPrint)
	module.defineFunction("println", ellPrintln)

	module.defineFunction("number?", ellNumberP)   // either real or integer
	module.defineFunction("integer?", ellIntegerP) //integer only
	module.defineFunction("+", ellPlus)
	module.defineFunction("-", ellMinus)
	module.defineFunction("*", ellTimes)
	module.defineFunction("/", ellDiv)
	module.defineFunction("quotient", ellQuotient)
	module.defineFunction("remainder", ellRemainder)
	module.defineFunction("modulo", ellRemainder) //fix!

	module.defineFunction("vector?", ellVectorP)
	module.defineFunction("make-vector", ellMakeVector)
	module.defineFunction("vector-set!", ellVectorSetBang)
	module.defineFunction("vector-ref", ellVectorRef)

	module.defineFunction("=", ellNumeq)
	module.defineFunction("<=", ellLe)
	module.defineFunction(">=", ellGe)
	module.defineFunction(">", ellGt)
	module.defineFunction("<", ellLt)
	module.defineFunction("zero?", ellZeroP)
	module.defineFunction("number->string", ellNumberToString)
	module.defineFunction("string-length", ellStringLength)

	module.defineFunction("map?", ellMapP)
	module.defineFunction("has?", ellHasP)
	module.defineFunction("get", ellGet)
	module.defineFunction("put!", ellPutBang)

	module.defineFunction("error", ellFatal)
	module.defineFunction("length", ellLength)
	module.defineFunction("json", ellJSON)
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *llist
//

func ellLetrec(argv []lob, argc int) (lob, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []lob, argc int) (lob, error) {
	return expandLet(argv[0])
}

func ellDo(argv []lob, argc int) (lob, error) {
	return expandDo(argv[0])
}

func ellCond(argv []lob, argc int) (lob, error) {
	return expandCond(argv[0])
}

func ellAnd(argv []lob, argc int) (lob, error) {
	return expandAnd(argv[0])
}

// functions

func ellVersion(argv []lob, argc int) (lob, error) {
	return newString(Version), nil
}

func ellMacroexpand(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError("macroexpand", "1", argc)
	}
	return macroexpand(argv[0])
}

func ellType(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError("type", "1", argc)
	}
	return argv[0].typeSymbol(), nil
}

func ellIdenticalP(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return argcError("identical?", "2", argc)
}

func ellEq(argv []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError("eq?", "1+", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if !equal(obj, argv[1]) {
			return False, nil
		}
	}
	return True, nil
}

func ellNumeq(argv []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError("=", "1+", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if b, err := numericallyEqual(obj, argv[1]); err != nil || !b {
			return False, err
		}
	}
	return True, nil
}

func ellDisplay(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		//todo: add the optional port argument like scheme
		return argcError("display", "1", argc)
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func ellWrite(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		//todo: add the optional port argument like scheme
		return argcError("write", "1", argc)
	}
	fmt.Printf("%v", write(argv[0]))
	return nil, nil
}

func ellNewline(argv []lob, argc int) (lob, error) {
	if argc != 0 {
		//todo: add the optional port argument like scheme
		return argcError("newline", "0", argc)
	}
	fmt.Printf("\n")
	return nil, nil
}

func ellFatal(argv []lob, argc int) (lob, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, newError(s)
}

func ellPrint(argv []lob, argc int) (lob, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return nil, nil
}

func ellPrintln(argv []lob, argc int) (lob, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return nil, nil
}

func ellList(argv []lob, argc int) (lob, error) {
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func ellNumberP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("number?", "1", argc)
}

func ellIntegerP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isInteger(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("integer?", "1", argc)
}

func ellQuotient(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		n1, err := integerValue(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := integerValue(argv[1])
		if err != nil {
			return nil, err
		}
		if n2 == 0 {
			return nil, newError("Quotient: divide by zero")
		}
		return newInteger(n1 / n2), nil
	}
	return argcError("quotient", "2", argc)
}

func ellRemainder(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		n1, err := integerValue(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := integerValue(argv[1])
		if n2 == 0 {
			return nil, newError("remainder: divide by zero")
		}
		if err != nil {
			return nil, err
		}
		return newInteger(n1 % n2), nil
	}
	return argcError("remainder", "2", argc)
}

func ellPlus(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []lob, argc int) (lob, error) {
	return minus(argv, argc)
}

func ellTimes(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellDiv(argv []lob, argc int) (lob, error) {
	return div(argv, argc)
}

func ellMakeVector(argv []lob, argc int) (lob, error) {
	if argc > 0 {
		var initVal lob = Nil
		vlen, err := integerValue(argv[0])
		if err != nil {
			return nil, err
		}
		if argc > 1 {
			if argc != 2 {
				return argcError("make-vector", "1-2", argc)
			}
			initVal = argv[1]
		}
		return newVector(int(vlen), initVal), nil
	}
	return argcError("make-vector", "1-2", argc)
}

func ellVectorP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isVector(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("list?", "1", argc)
}

func ellVectorSetBang(argv []lob, argc int) (lob, error) {
	if argc == 3 {
		v := argv[0]
		idx, err := integerValue(argv[1])
		if err != nil {
			return nil, err
		}
		err = vectorSet(v, int(idx), argv[2])
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return argcError("vector-set!", "3", argc)
}

func ellVectorRef(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		v := argv[0]
		idx, err := integerValue(argv[1])
		if err != nil {
			return nil, err
		}
		val, err := vectorRef(v, int(idx))
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return argcError("vector-ref", "2", argc)
}

func ellGe(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError(">=", "2", argc)
}

func ellLe(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError("<=", "2", argc)
}

func ellGt(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError(">", "2", argc)
}

func ellLt(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError("<", "2", argc)
}

func ellZeroP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		f, err := realValue(argv[0])
		if err != nil {
			return nil, err
		}
		if f == 0 {
			return True, nil
		}
		return False, nil
	}
	return argcError("zero?", "1", argc)
}

func ellNumberToString(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError("number->string", "1", argc)
	}
	if !isNumber(argv[0]) {
		return argTypeError("number", 1, argv[0])
	}
	return newString(argv[0].String()), nil
}

func ellStringLength(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError("string-length", "1", argc)
	}
	if !isString(argv[0]) {
		return argTypeError("string", 1, argv[0])
	}
	i := length(argv[0])
	return newInteger(int64(i)), nil
}

func ellLength(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		return newInteger(int64(length(argv[0]))), nil
	}
	return argcError("length", "1", argc)
}

func ellNot(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return argcError("not", "1", argc)
}

func ellNullP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if argv[0] == Nil {
			return True, nil
		}
		return False, nil
	}
	return argcError("null?", "1", argc)
}

func ellBooleanP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("boolean?", "1", argc)
}

func ellSymbolP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("symbol?", "1", argc)
}

func ellKeywordP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("keyword?", "1", argc)
}

func ellStringP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("string?", "1", argc)
}

func ellCharacterP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isCharacter(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("character?", "1", argc)
}

func ellFunctionP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isFunction(argv[0]) || isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("function?", "1", argc)
}

func ellListP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("list?", "1", argc)
}

func ellString(argv []lob, argc int) (lob, error) {
	s := ""
	for i := 0; i < argc; i++ {
		s += argv[i].String()
	}
	return newString(s), nil
}

func ellCar(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(lst), nil
		}
		return argTypeError("list", 1, lst)
	}
	return argcError("car", "1", argc)
}

func ellCdr(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdr(lst), nil
		}
		return argTypeError("list", 1, lst)
	}
	return argcError("cdr", "1", argc)
}

func ellCadr(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return argTypeError("list", 1, lst)
	}
	return argcError("cadr", "1", argc)
}

func ellCddr(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return argTypeError("list", 1, lst)
	}
	return argcError("cddr", "1", argc)
}

func ellCons(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		switch lst := argv[1].(type) {
		case *llist:
			return cons(argv[0], lst), nil
		default:
			return argTypeError("list", 2, lst)
		}
	}
	return argcError("cons", "2", argc)
}

func ellMapP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if isMap(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return argcError("list?", "1", argc)
}

func ellGet(argv []lob, argc int) (lob, error) {
	if argc != 2 {
		return argcError("get", "2", argc)
	}
	return get(argv[0], argv[1])
}

func ellHasP(argv []lob, argc int) (lob, error) {
	if argc != 2 {
		return argcError("has?", "2", argc)
	}
	b, err := has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return True, nil
	}
	return False, nil
}

func ellPutBang(argv []lob, argc int) (lob, error) {
	if argc != 3 {
		return argcError("put!", "3", argc)
	}
	return put(argv[0], argv[1], argv[2])
}

func ellJSON(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError("json", "1", argc)
	}
	s, err := toJSON(argv[0])
	if err != nil {
		return nil, err
	}
	return newString(s), nil
}
