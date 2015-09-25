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
	defineMacro("cond", ellCond)
	defineMacro("quasiquote", ellQuasiquote)

	define("null", Null)
	define("true", True)
	define("false", False)

	define("apply", Apply)

	defineFunction("version", ellVersion, "()")

	defineFunction("def?", ellDefinedP, "(<any>)")

	defineFunction("file-contents", ellFileContents, "(<string>) <string>")
	defineFunction("open-input-string", ellOpenInputString, "(<string>) <input>")
	defineFunction("open-input-file", ellOpenInputFile, "(<string>) <input>")
	defineFunction("read", ellRead, "([<input>]) <any>")
	defineFunction("close-input", ellCloseInput, "(<input>) <null>")

	defineFunction("macroexpand", ellMacroexpand, "(<any>) <any>")
	defineFunction("compile", ellCompile, "(<any>) <code>")

	defineFunction("type", elvariant, "(<any>) <type>")
	defineFunction("value", ellValue, "<any>) <any>")
	defineFunction("instance", ellInstance, "(<type> <any>) <any>")
	defineFunction("normalize-keyword-arg-list", ellNormalizeKeywordArgList, "(<list> <keyword>+) <list>")
	defineFunction("normalize-keyword-args", ellNormalizeKeywordArgList, "(<list> <keyword>+) <struct>")

	defineFunction("type?", elvariantP, "(<any>) <boolean>")
	defineFunction("type-name", elvariantName, "(<type>) <symbol>")

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
	defineFunction("character?", ellCharacterP, "(<any>) <boolean>")
	defineFunction("function?", ellFunctionP, "(<any>) <boolean>")
	defineFunction("eof?", ellFunctionP, "(<any>) <boolean>")

	defineFunction("list?", ellListP, "(<any>) <boolean>")
	defineFunction("to-list", ellToList, "(<any>) <list>")
	defineFunction("cons", ellCons, "(<any> <list>) <list>")
	defineFunction("car", ellCar, "(<list>) <any>")
	defineFunction("cdr", ellCdr, "(<list>) <list>")
	defineFunction("list", ellList, "(<any>*) <list>")
	defineFunction("concat", ellConcat, "(<list>*) <list>")
	defineFunction("reverse", ellReverse, "(<list>) <list>")
	defineFunction("flatten", ellFlatten, "(<list>) <list>")
	defineFunction("set-car!", ellSetCarBang, "(<list> <any>) <null>")  //mutate!
	defineFunction("set-cdr!", ellSetCdrBang, "(<list> <list>) <null>") //mutate!

	defineFunction("vector?", ellVectorP, "(<any>) <boolean>")
	defineFunction("to-vector", ellToVector, "(<any>) <vector>")
	defineFunction("vector", ellVector, "(<any>*) <vector>")
	defineFunction("make-vector", ellMakeVector, "(<number> <any>) <vector>")
	defineFunction("vector-set!", ellVectorSetBang, "(<vector> <number> <any>) <null>") //mutate!
	defineFunction("vector-ref", ellVectorRef, "(<vector> <number>) <any>")

	defineFunction("struct?", ellStructP, "(<any>) <boolean>")
	defineFunction("has?", ellHasP, "(<struct> <any>) <boolean>")
	defineFunction("get", ellGet, "(<struct> <any>) <any>")
	defineFunction("put!", ellPutBang, "(<struct> <any> <any>) <null>") //mutate!
	//	defineFunction("assoc", ellAssoc, "(<struct> <any>) <struct")
	//	defineFunction("dissoc", ellDissoc, "(<struct> <any>) <struct>")

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

func ellLetrec(argv []*LOB, argc int) (*LOB, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []*LOB, argc int) (*LOB, error) {
	return expandLet(argv[0])
}

func ellCond(argv []*LOB, argc int) (*LOB, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []*LOB, argc int) (*LOB, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(argv []*LOB, argc int) (*LOB, error) {
	return newString(Version), nil
}

func ellDefinedP(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("defined?", "1", argc)
	}
	if !isSymbol(argv[0]) {
		return nil, ArgTypeError("symbol", 1, argv[0])
	}
	if isDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFileContents(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("file-contents", "1", argc)
	}
	if !isString(argv[0]) {
		return nil, ArgTypeError("string", 1, argv[0])
	}
	fname, err := asString(argv[0])
	if err != nil {
		return nil, err
	}
	s, err := fileContents(fname)
	if err != nil {
		return nil, err
	}
	return newString(s), nil
}

func ellOpenInputString(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("open-input-string", "1", argc)
	}
	s, err := asString(argv[0])
	if err != nil {
		return nil, err
	}
	return openInputString(s), nil
}

func ellOpenInputFile(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("open-input-file", "1", argc)
	}
	s, err := asString(argv[0])
	if err != nil {
		return nil, err
	}
	return openInputFile(s)
}

func ellRead(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("read", "1+", argc)
	}
	options, err := getOptions(argv[1:argc], "keys:")
	if err != nil {
		return nil, err
	}
	return readInputPort(argv[0], options)
}

func ellCloseInput(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("read", "1", argc)
	}
	return nil, closeInputPort(argv[0])
}

func ellMacroexpand(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("macroexpand", "1", argc)
	}
	return macroexpand(argv[0])
}

func ellCompile(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("compile", "1", argc)
	}
	expanded, err := macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return compile(expanded)
}

func elvariant(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("type", "1", argc)
	}
	return argv[0].variant, nil
}

func ellValue(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("value", "1", argc)
	}
	return value(argv[0]), nil
}

func ellInstance(argv []*LOB, argc int) (*LOB, error) {
	if argc != 2 {
		return nil, ArgcError("instance", "2", argc)
	}
	return instance(argv[0], argv[1])
}

func ellNormalizeKeywordArgList(argv []*LOB, argc int) (*LOB, error) {
	//(normalize-keyword-arglist '(x: 23) x: y:) -> (x:)
	//(normalize-keyword-arglist '(x: 23 z: 100) x: y:) -> error("bad keyword z: in argument list")
	if argc < 1 {
		return nil, ArgcError("normalizeKeywordArgList", "1+", argc)
	}
	if isList(argv[0]) {
		return normalizeKeywordArgList(argv[0], argv[1:argc])
	}
	return nil, ArgTypeError("list", 1, argv[0])
}

func ellNormalizeKeywordArgs(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("normalizeKeywordArgs", "1+", argc)
	}
	if !isList(argv[0]) {
		return nil, ArgTypeError("list", 1, argv[0])
	}
	return normalizeKeywordArgs(argv[0], argv[1:argc])
}

func ellStruct(argv []*LOB, argc int) (*LOB, error) {
	return newStruct(argv[:argc])
}

func ellIdenticalP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("identical?", "2", argc)
}

func ellEq(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("eq?", "1+", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if !equal(obj, argv[1]) {
			return False, nil
		}
	}
	return True, nil
}

func ellNumeq(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("=", "1+", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if b, err := numericallyEqual(obj, argv[1]); err != nil || b == False {
			return b, err
		}
	}
	return True, nil
}

func ellDisplay(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		//todo: add the optional output argument like scheme
		return nil, ArgcError("display", "1", argc)
	}
	fmt.Printf("%v", argv[0])
	return Null, nil
}

func ellWrite(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("write", "1+", argc)
	}
	validOptions := []*LOB{intern("pretty:")} //to do: a to: argument for an output destination (port, file, url)
	options, err := normalizeKeywordArgs(list(argv[1:argc]...), validOptions)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v", writeToString(argv[0], options))
	return Null, nil
}

func ellNewline(argv []*LOB, argc int) (*LOB, error) {
	if argc != 0 {
		//todo: add the optional output argument like scheme
		return nil, ArgcError("newline", "0", argc)
	}
	fmt.Println("")
	return Null, nil
}

func ellFatal(argv []*LOB, argc int) (*LOB, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}

func ellToString(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("to-string", "1", argc)
	}
	s := writeToString(argv[0], Null)
	return newString(s), nil
}

func ellPrint(argv []*LOB, argc int) (*LOB, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []*LOB, argc int) (*LOB, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []*LOB, argc int) (*LOB, error) {
	result := EmptyList
	tail := result
	for i := 0; i < argc; i++ {
		lst := argv[i]
		if !isList(lst) {
			return nil, ArgTypeError("list", i+1, lst)
		}
		for lst != EmptyList {
			if tail == EmptyList {
				result = list(lst.car)
				tail = result
			} else {
				tail.cdr = list(lst.car)
				tail = tail.cdr
			}
			lst = lst.cdr
		}
	}
	return result, nil
}

func ellReverse(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("reverse", "1", argc)
	}
	lst := argv[0]
	if !isList(lst) {
		return nil, ArgTypeError("list", 1, lst)
	}
	return reverse(lst), nil
}

func ellFlatten(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("flatten", "1", argc)
	}
	seq := argv[0]
	switch seq.variant {
	case typeList:
		return flatten(seq), nil
	case typeVector:
		lst, _ := toList(seq)
		return flatten(lst), nil
	default:
		return nil, ArgTypeError("list", 1, seq)
	}
}

func ellList(argv []*LOB, argc int) (*LOB, error) {
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func ellNumberP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("number?", "1", argc)
}

func ellIntP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isInt(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("int?", "1", argc)
}

func ellFloatP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isFloat(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("float?", "1", argc)
}

func ellQuotient(argv []*LOB, argc int) (*LOB, error) {
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
		return newInt64(n1 / n2), nil
	}
	return nil, ArgcError("quotient", "2", argc)
}

func ellRemainder(argv []*LOB, argc int) (*LOB, error) {
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
		return newInt64(n1 % n2), nil
	}
	return nil, ArgcError("remainder", "2", argc)
}

func ellPlus(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []*LOB, argc int) (*LOB, error) {
	return minus(argv, argc)
}

func ellTimes(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellDiv(argv []*LOB, argc int) (*LOB, error) {
	return div(argv, argc)
}

func ellVector(argv []*LOB, argc int) (*LOB, error) {
	return vector(argv...), nil
}

func ellToVector(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("to-vector", "1", argc)
	}
	return toVector(argv[0])
}

func ellMakeVector(argv []*LOB, argc int) (*LOB, error) {
	if argc > 0 {
		initVal := Null
		vlen, err := intValue(argv[0])
		if err != nil {
			return nil, err
		}
		if argc > 1 {
			if argc != 2 {
				return nil, ArgcError("make-vector", "1-2", argc)
			}
			initVal = argv[1]
		}
		return newVector(int(vlen), initVal), nil
	}
	return nil, ArgcError("make-vector", "1-2", argc)
}

func ellVectorP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isVector(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("vector?", "1", argc)
}

func ellVectorSetBang(argv []*LOB, argc int) (*LOB, error) {
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
	return nil, ArgcError("vector-set!", "3", argc)
}

func ellVectorRef(argv []*LOB, argc int) (*LOB, error) {
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
	return nil, ArgcError("vector-ref", "2", argc)
}

func ellGe(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError(">=", "2", argc)
}

func ellLe(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError("<=", "2", argc)
}

func ellGt(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError(">", "2", argc)
}

func ellLt(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError("<", "2", argc)
}

func ellZeroP(argv []*LOB, argc int) (*LOB, error) {
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
	return nil, ArgcError("zero?", "1", argc)
}

func ellNumberToString(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("number->string", "1", argc)
	}
	if !isNumber(argv[0]) {
		return nil, ArgTypeError("number", 1, argv[0])
	}
	return newString(argv[0].String()), nil
}

func ellStringLength(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("string-length", "1", argc)
	}
	if !isString(argv[0]) {
		return nil, ArgTypeError("string", 1, argv[0])
	}
	i := length(argv[0])
	return newInt(i), nil
}

func ellLength(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		return newInt(length(argv[0])), nil
	}
	return nil, ArgcError("length", "1", argc)
}

func ellNot(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("not", "1", argc)
}

func ellNullP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if argv[0] == Null {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("nil?", "1", argc)
}

func ellBooleanP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("boolean?", "1", argc)
}

func elLSymbolP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("symbol?", "1", argc)
}

func elLSymbol(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("symbol", "1+", argc)
	}
	return symbol(argv[:argc])
}

func ellKeywordP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("keyword?", "1", argc)
}

func elvariantP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("type?", "1", argc)
}

func elvariantName(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return typeName(argv[0])
		}
		return False, nil
	}
	return nil, ArgcError("type-name", "1", argc)
}

func ellStringP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("string?", "1", argc)
}

func ellCharacterP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isCharacter(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("character?", "1", argc)
}

func ellFunctionP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isFunction(argv[0]) || isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("function?", "1", argc)
}

func ellListP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("list?", "1", argc)
}

func ellEmptyP(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("empty?", "1", argc)
	}
	if isEmpty(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellString(argv []*LOB, argc int) (*LOB, error) {
	s := ""
	for i := 0; i < argc; i++ {
		s += argv[i].String()
	}
	return newString(s), nil
}

func ellCar(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		//		return safeCar(argv[0])
		return car(argv[0]), nil
	}
	return nil, ArgcError("car", "1", argc)
}

func ellCdr(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		//		return safeCdr(argv[0])
		return cdr(argv[0]), nil
	}
	return nil, ArgcError("cdr", "1", argc)
}

func ellSetCarBang(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		err := setCar(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return Null, nil
	}
	return nil, ArgcError("set-car!", "2", argc)
}

func ellSetCdrBang(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		err := setCdr(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return Null, nil
	}
	return nil, ArgcError("set-cdr!", "2", argc)
}

func ellCaar(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caar(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("caar", "1", argc)
}

func ellCadr(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cadr", "1", argc)
}

func ellCddr(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cddr", "1", argc)
}

func ellCadar(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(cdr(car(lst))), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cadar", "1", argc)
}

func ellCaddr(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caddr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("caddr", "1", argc)
}
func ellCdddr(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdddr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cdddr", "1", argc)
}

func ellCons(argv []*LOB, argc int) (*LOB, error) {
	if argc == 2 {
		lst := argv[1]
		if !isList(lst) {
			return nil, ArgTypeError("list", 2, lst)
		}
		return cons(argv[0], lst), nil
	}
	return nil, ArgcError("cons", "2", argc)
}

func ellStructP(argv []*LOB, argc int) (*LOB, error) {
	if argc == 1 {
		if isStruct(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("struct?", "1", argc)
}

func ellGet(argv []*LOB, argc int) (*LOB, error) {
	if argc != 2 {
		return nil, ArgcError("get", "2", argc)
	}
	return get(argv[0], argv[1])
}

func ellHasP(argv []*LOB, argc int) (*LOB, error) {
	if argc != 2 {
		return nil, ArgcError("has?", "2", argc)
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

func ellPutBang(argv []*LOB, argc int) (*LOB, error) {
	if argc != 3 {
		return nil, ArgcError("put!", "3", argc)
	}
	return put(argv[0], argv[1], argv[2])
}

/*
func ellAssoc(argv []*LOB, argc int) (*LOB, error) {
	if argc != 3 {
		return nil, ArgcError("assoc", "3", argc)
	}
	return assoc(argv[0], argv[1], argv[2])
}

func ellDissoc(argv []*LOB, argc int) (*LOB, error) {
	if argc != 3 {
		return nil, ArgcError("dissoc", "2", argc)
	}
	return dissoc(argv[0], argv[1])
}
*/

func ellToList(argv []*LOB, argc int) (*LOB, error) {
	if argc != 1 {
		return nil, ArgcError("to-list", "1", argc)
	}
	return toList(argv[0])
}

func ellJSON(argv []*LOB, argc int) (*LOB, error) {
	if argc < 1 {
		return nil, ArgcError("json", "1+", argc)
	}
	options, err := getOptions(argv[1:argc], "pretty:")
	if err != nil {
		return nil, err
	}
	s, err := toJSON(argv[0], options)
	if err != nil {
		return nil, err
	}
	return newString(s), nil
}

func getOptions(rest []*LOB, keys ...string) (*LOB, error) {
	var validOptions []*LOB
	for _, key := range keys {
		validOptions = append(validOptions, intern(key))
	}
	return normalizeKeywordArgs(list(rest...), validOptions)
}
