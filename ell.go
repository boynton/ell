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

// defines the global functions/variables/macros for the top level environment
func initEnvironment() {

	defineMacro("let", ellLet)
	defineMacro("letrec", ellLetrec)
	defineMacro("do", ellDo) //scheme's do. I don't like it, will replace with... "for" of some sort
	//note clojure uses "do" instead of "begin". I rather prefer that.
	defineMacro("cond", ellCond)
	defineMacro("quasiquote", ellQuasiquote)

	define("null", Null)
	define("true", True)
	define("false", False)

	define("apply", Apply)

	defineFunction("version", ellVersion, "()")

	defineFunction("defined?", ellDefinedP, "(<any>)")

	defineFunction("file-contents", ellFileContents, "(<string>) <string>")
	defineFunction("open-input-string", ellOpenInputString, "(<string>) <input>")
	defineFunction("open-input-file", ellOpenInputFile, "(<string>) <input>")
	defineFunction("read", ellRead, "([<input>]) <any>")
	defineFunction("close-input", ellCloseInput, "(<input>) <null>")

	defineFunction("macroexpand", ellMacroexpand, "(<any>) <any>")
	defineFunction("compile", ellCompile, "(<any>) <code>")

	defineFunction("type", ellType, "(<any>) <type>")
	defineFunction("value", ellValue, "<any>) <any>")
	defineFunction("instance", ellInstance, "(<type> <any>) <any>")
	defineFunction("normalize-keyword-args", ellNormalizeKeywordArgs, "(<list> <keyword>+) <list>")

	defineFunction("type?", ellTypeP, "(<any>) <boolean>")
	defineFunction("type-name", ellTypeName, "(<type>) <symbol>")

	defineFunction("struct", ellStruct, "(<any>+) <struct>")
	defineFunction("equal?", ellEq, "(<any> <any>) <boolean>")
	defineFunction("identical?", ellIdenticalP, "(<any> <any>) <boolean>")
	defineFunction("not", ellNot, "(<any>) <boolean>")

	defineFunction("boolean?", ellBooleanP, "(<any>) <boolean>")
	defineFunction("null?", ellNullP, "(<any>) <boolean>")
	defineFunction("symbol?", elLSymbolP, "(<any>) <boolean>")
	defineFunction("symbol", elLSymbol, "(<any>+) <boolean>")

	defineFunction("keyword?", ellKeywordP, "(<any>) <boolean>")
	defineFunction("string?", ellStringP, "(<any>) <boolean>")
	defineFunction("char?", ellCharP, "(<any>) <boolean>")
	defineFunction("function?", ellFunctionP, "(<any>) <boolean>")
	defineFunction("eof?", ellFunctionP, "(<any>) <boolean>")

	defineFunction("list?", ellListP, "(<any>) <boolean>")
	defineFunction("cons", ellCons, "(<any> <list>) <list>")
	defineFunction("car", ellCar, "(<list>) <any>")
	defineFunction("cdr", ellCdr, "(<list>) <list>")
	defineFunction("list", ellList, "(<any>*) <list>")
	defineFunction("concat", ellConcat, "(<list>*) <list>")
	defineFunction("reverse", ellReverse, "(<list>) <list>")
	defineFunction("set-car!", ellSetCarBang, "(<list> <any>) <null>") //mutate!
	defineFunction("set-cdr!", ellSetCdrBang, "(<list> <list>) <null>") //mutate!

	defineFunction("vector?", ellVectorP, "(<any>) <boolean>")
	defineFunction("vector", ellVector, "(<any>*) <vector>")
	defineFunction("make-vector", ellMakeVector, "(<number> <any>) <vector>")
	defineFunction("vector-set!", ellVectorSetBang, "(<vector> <number> <any>) <null>") //mutate!
	defineFunction("vector-ref", ellVectorRef, "(<vector> <number>) <any>")

	defineFunction("struct?", ellStructP, "(<any>) <boolean>")
	defineFunction("has?", ellHasP, "(<struct> <any>) <boolean>")
	defineFunction("get", ellGet, "(<struct> <any>) <any>")
	defineFunction("assoc", ellAssoc, "(<struct> <any>) <struct")
	defineFunction("dissoc", ellDissoc, "(<struct> <any>) <struct>")
	defineFunction("put!", ellPutBang, "(<struct> <any> <any>) <null>") //mutate!
	defineFunction("struct->list", ellStructToList, "(<struct>) <list>")

	defineFunction("empty?", ellEmptyP, "(<any>) <boolean>")

	defineFunction("string", ellString, "(<any>*) <string>")
	defineFunction("string-length", ellStringLength, "(<string>) <number>")

	defineFunction("display", ellDisplay, "(<any>) <null>")
	defineFunction("write", ellWrite, "(<any>) <null>")
	defineFunction("newline", ellNewline, "() <null>")
	defineFunction("print", ellPrint, "(<any>*) <null>")
	defineFunction("println", ellPrintln, "(<any>*) <null>")
	defineFunction("to-string", ellToString, "(<any>)  <string>")

	defineFunction("number?", ellNumberP, "(<any>) <boolean>") // either float or int
	defineFunction("int?", ellIntP, "(<any>) <boolean>")       //int only
	defineFunction("float?", ellFloatP, "(<any>) <boolean>")   //float only
	defineFunction("+", ellPlus, "(<number>*) <number>")
	defineFunction("-", ellMinus, "(<number>+) <number>")
	defineFunction("*", ellTimes, "(<number>*) <number>")
	defineFunction("/", ellDiv, "(<number>+) <number>")
	defineFunction("quotient", ellQuotient, "(<number> <number>) <number>")
	defineFunction("remainder", ellRemainder, "(<number> <number>) <number>")
	defineFunction("modulo", ellRemainder, "(<number> <number>) <number>") //fix!

	defineFunction("=", ellNumeq, "(<number>+) <boolean>")
	defineFunction("<=", ellLe, "(<number>+) <boolean>")
	defineFunction(">=", ellGe, "(<number>+) <boolean>")
	defineFunction(">", ellGt, "(<number>+) <boolean>")
	defineFunction("<", ellLt, "(<number>+) <boolean>")
	defineFunction("zero?", ellZeroP, "(<number>) <boolean>")

	defineFunction("error", ellFatal, "(<any>+) <null>")
	defineFunction("length", ellLength, "(<any>) <number>")
	defineFunction("json", ellJSON, "(<any>) <string>")

	err := loadModule("ell")
	if err != nil {
		fatal("*** ", err)
	}
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *LList
//

func ellLetrec(argv []LAny, argc int) (LAny, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []LAny, argc int) (LAny, error) {
	return expandLet(argv[0])
}

func ellDo(argv []LAny, argc int) (LAny, error) {
	return expandDo(argv[0])
}

func ellCond(argv []LAny, argc int) (LAny, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []LAny, argc int) (LAny, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(argv []LAny, argc int) (LAny, error) {
	return LString(Version), nil
}

func ellDefinedP(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("defined?", "1", argc)
	}
	if !isSymbol(argv[0]) {
		return ArgTypeError("symbol", 1, argv[0])
	}
	if isDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFileContents(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("file-contents", "1", argc)
	}
	if !isString(argv[0]) {
		return ArgTypeError("string", 1, argv[0])
	}
	fname, err := stringValue(argv[0])
	if err != nil {
		return nil, err
	}
	s, err := fileContents(fname)
	if err != nil {
		return nil, err
	}
	return LString(s), nil
}

func ellOpenInputString(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("open-input-string", "1", argc)
	}
	if !isString(argv[0]) {
		return ArgTypeError("string", 1, argv[0])
	}
	s, err := stringValue(argv[0])
	if err != nil {
		return nil, err
	}
	return openInputString(s), nil
}

func ellOpenInputFile(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("open-input-file", "1", argc)
	}
	if !isString(argv[0]) {
		return ArgTypeError("string", 1, argv[0])
	}
	s, err := stringValue(argv[0])
	if err != nil {
		return nil, err
	}
	return openInputFile(s)
}

func ellRead(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("read", "1", argc)
	}
	return readInput(argv[0])
}
func ellCloseInput(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("read", "1", argc)
	}
	return nil, closeInput(argv[0])
}

func ellMacroexpand(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("macroexpand", "1", argc)
	}
	return macroexpand(argv[0])
}

func ellCompile(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("compile", "1", argc)
	}
	expanded, err := macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return compile(expanded)
}

func ellType(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("type", "1", argc)
	}
	return argv[0].Type(), nil
}

func ellValue(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("value", "1", argc)
	}
	return argv[0].Value(), nil
}

func ellInstance(argv []LAny, argc int) (LAny, error) {
	if argc != 2 {
		return ArgcError("instance", "2", argc)
	}
	return instance(argv[0], argv[1])
}

func ellNormalizeKeywordArgs(argv []LAny, argc int) (LAny, error) {
	//(normalize-keyword-args '(x: 23) x: y:) -> (x:)
	//(normalize-keyword-args '(x: 23 z: 100) x: y:) -> error("bad keyword z: in argument list")
	if argc < 1 {
		return ArgcError("normalizeKeywordArgs", "1+", argc)
	}
	if args, ok := argv[0].(*LList); ok {
		return normalizeKeywordArgs(args, argv[1:argc])
	}
	return ArgTypeError("list", 1, argv[0])
}

func ellStruct(argv []LAny, argc int) (LAny, error) {
	return newStruct(argv[:argc])
}

func ellIdenticalP(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("identical?", "2", argc)
}

func ellEq(argv []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return ArgcError("eq?", "1+", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if !equal(obj, argv[1]) {
			return False, nil
		}
	}
	return True, nil
}

func ellNumeq(argv []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return ArgcError("=", "1+", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if b, err := numericallyEqual(obj, argv[1]); err != nil || !b {
			return False, err
		}
	}
	return True, nil
}

func ellDisplay(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		//todo: add the optional output argument like scheme
		return ArgcError("display", "1", argc)
	}
	fmt.Printf("%v", argv[0])
	return Null, nil
}

func ellWrite(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		//todo: add the optional output argument like scheme
		return ArgcError("write", "1", argc)
	}
	fmt.Printf("%v", write(argv[0]))
	return Null, nil
}

func ellNewline(argv []LAny, argc int) (LAny, error) {
	if argc != 0 {
		//todo: add the optional output argument like scheme
		return ArgcError("newline", "0", argc)
	}
	fmt.Printf("\n")
	return Null, nil
}

func ellFatal(argv []LAny, argc int) (LAny, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}

func ellToString(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("to-string", "1", argc)
	}
	s := write(argv[0])
	return LString(s), nil
}

func ellPrint(argv []LAny, argc int) (LAny, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []LAny, argc int) (LAny, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []LAny, argc int) (LAny, error) {
	result := EmptyList
	tail := result
	for i := 0; i < argc; i++ {
		o := argv[i]
		switch lst := o.(type) {
		case *LList:
			c := lst
			for c != EmptyList {
				if tail == EmptyList {
					result = list(lst.car)
					tail = result
				} else {
					tail.cdr = list(c.car)
					tail = tail.cdr
				}
				c = c.cdr
			}
		default:
			return ArgTypeError("list", i+1, o)
		}
	}
	return result, nil
}

func ellReverse(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("reverse", "1", argc)
	}
	o := argv[0]
	switch lst := o.(type) {
	case *LList:
		return reverse(lst), nil
	default:
		return ArgTypeError("list", 1, o)
	}
}

func ellList(argv []LAny, argc int) (LAny, error) {
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func ellNumberP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("number?", "1", argc)
}

func ellIntP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isInt(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("int?", "1", argc)
}

func ellFloatP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isFloat(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("float?", "1", argc)
}

func ellQuotient(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		n1, err := int64Value(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := int64Value(argv[1])
		if err != nil {
			return nil, err
		}
		if n2 == 0 {
			return nil, Error("Quotient: divide by zero")
		}
		return LNumber(n1 / n2), nil
	}
	return ArgcError("quotient", "2", argc)
}

func ellRemainder(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		n1, err := int64Value(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := int64Value(argv[1])
		if n2 == 0 {
			return nil, Error("remainder: divide by zero")
		}
		if err != nil {
			return nil, err
		}
		return LNumber(n1 % n2), nil
	}
	return ArgcError("remainder", "2", argc)
}

func ellPlus(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []LAny, argc int) (LAny, error) {
	return minus(argv, argc)
}

func ellTimes(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellDiv(argv []LAny, argc int) (LAny, error) {
	return div(argv, argc)
}

func ellVector(argv []LAny, argc int) (LAny, error) {
	return vector(argv...), nil
}

func ellMakeVector(argv []LAny, argc int) (LAny, error) {
	if argc > 0 {
		initVal := LAny(Null)
		vlen, err := intValue(argv[0])
		if err != nil {
			return nil, err
		}
		if argc > 1 {
			if argc != 2 {
				return ArgcError("make-vector", "1-2", argc)
			}
			initVal = argv[1]
		}
		return newVector(int(vlen), initVal), nil
	}
	return ArgcError("make-vector", "1-2", argc)
}

func ellVectorP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isVector(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("vector?", "1", argc)
}

func ellVectorSetBang(argv []LAny, argc int) (LAny, error) {
	if argc == 3 {
		a := argv[0]
		idx, err := intValue(argv[1])
		if err != nil {
			return nil, err
		}
		err = vectorSet(a, int(idx), argv[2])
		if err != nil {
			return nil, err
		}
		return a, nil
	}
	return ArgcError("vector-set!", "3", argc)
}

func ellVectorRef(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		a := argv[0]
		idx, err := intValue(argv[1])
		if err != nil {
			return nil, err
		}
		val, err := vectorRef(a, int(idx))
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return ArgcError("vector-ref", "2", argc)
}

func ellGe(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError(">=", "2", argc)
}

func ellLe(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError("<=", "2", argc)
}

func ellGt(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError(">", "2", argc)
}

func ellLt(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError("<", "2", argc)
}

func ellZeroP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		f, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		if f == 0 {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("zero?", "1", argc)
}

func ellNumberToString(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("number->string", "1", argc)
	}
	if !isNumber(argv[0]) {
		return ArgTypeError("number", 1, argv[0])
	}
	return LString(argv[0].String()), nil
}

func ellStringLength(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("string-length", "1", argc)
	}
	if !isString(argv[0]) {
		return ArgTypeError("string", 1, argv[0])
	}
	i := length(argv[0])
	return LNumber(int64(i)), nil
}

func ellLength(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		return LNumber(int64(length(argv[0]))), nil
	}
	return ArgcError("length", "1", argc)
}

func ellNot(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("not", "1", argc)
}

func ellNullP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if argv[0] == Null {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("nil?", "1", argc)
}

func ellBooleanP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("boolean?", "1", argc)
}

func elLSymbolP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("symbol?", "1", argc)
}

func elLSymbol(argv []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return ArgcError("symbol", "1+", argc)
	}
	return symbol(argv[:argc])
}

func ellKeywordP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("keyword?", "1", argc)
}

func ellTypeP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("type?", "1", argc)
}

func ellTypeName(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return typeName(argv[0])
		}
		return False, nil
	}
	return ArgcError("type-name", "1", argc)
}

func ellStringP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("string?", "1", argc)
}

func ellCharP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isChar(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("character?", "1", argc)
}

func ellFunctionP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isFunction(argv[0]) || isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("function?", "1", argc)
}

func ellListP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("list?", "1", argc)
}

func ellEmptyP(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("empty?", "1", argc)
	}
	if isEmpty(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellString(argv []LAny, argc int) (LAny, error) {
	s := ""
	for i := 0; i < argc; i++ {
		s += argv[i].String()
	}
	return LString(s), nil
}

func ellCar(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("car", "1", argc)
}

func ellCdr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cdr", "1", argc)
}

func ellSetCarBang(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		lst := argv[0]
		if isList(lst) {
			setCar(lst, argv[1])
			return Null, nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("set-car!", "2", argc)
}

func ellSetCdrBang(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		lst := argv[0]
		if isList(lst) {
			if !isList(argv[1]) {
				return ArgTypeError("list", 2, lst)
			}
			setCdr(lst, argv[1])
			return Null, nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("set-cdr!", "2", argc)
}

func ellCaar(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caar(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("caar", "1", argc)
}

func ellCadr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cadr", "1", argc)
}

func ellCddr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cddr", "1", argc)
}

func ellCadar(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(cdr(car(lst))), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cadar", "1", argc)
}

func ellCaddr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("caddr", "1", argc)
}
func ellCdddr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cdddr", "1", argc)
}

func ellCons(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		switch lst := argv[1].(type) {
		case *LList:
			return cons(argv[0], lst), nil
		default:
			return ArgTypeError("list", 2, lst)
		}
	}
	return ArgcError("cons", "2", argc)
}

func ellStructP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isStruct(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("struct?", "1", argc)
}

func ellGet(argv []LAny, argc int) (LAny, error) {
	if argc != 2 {
		return ArgcError("get", "2", argc)
	}
	return get(argv[0], argv[1])
}

func ellHasP(argv []LAny, argc int) (LAny, error) {
	if argc != 2 {
		return ArgcError("has?", "2", argc)
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

func ellPutBang(argv []LAny, argc int) (LAny, error) {
	if argc != 3 {
		return ArgcError("put!", "3", argc)
	}
	return put(argv[0], argv[1], argv[2])
}

func ellAssoc(argv []LAny, argc int) (LAny, error) {
	if argc != 3 {
		return ArgcError("assoc", "3", argc)
	}
	return assoc(argv[0], argv[1], argv[2])
}

func ellDissoc(argv []LAny, argc int) (LAny, error) {
	if argc != 3 {
		return ArgcError("dissoc", "2", argc)
	}
	return dissoc(argv[0], argv[1])
}

func ellStructToList(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("struct->list", "1", argc)
	}
	return structToList(argv[0])
}

func ellJSON(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("json", "1", argc)
	}
	s, err := toJSON(argv[0])
	if err != nil {
		return nil, err
	}
	return LString(s), nil
}
