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
	"strings"
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

	defineFunction("boolean?", ellBooleanP, "(<any>) <boolean>")
	defineFunction("not", ellNot, "(<any>) <boolean>")

	defineFunction("equal?", ellEq, "(<any> <any>) <boolean>")
	defineFunction("identical?", ellIdenticalP, "(<any> <any>) <boolean>")

	defineFunction("null?", ellNullP, "(<any>) <boolean>")

	defineFunction("def?", ellDefinedP, "(<any>)")

	defineFunction("type", ellType, "(<any>) <type>")
	defineFunction("value", ellValue, "<any>) <any>")

	defineFunction("instance", ellInstance, "(<type> <any>) <any>")
	defineFunction("validate-keyword-arg-list", ellValidateKeywordArgList, "(<list> <keyword>+) <list>") // used by defstruct

	defineFunction("type?", ellTypeP, "(<any>) <boolean>")
	defineFunction("type-name", ellTypeName, "(<type>) <symbol>")

	defineFunction("keyword?", ellKeywordP, "(<any>) <boolean>")
	defineFunction("keyword-name", ellKeywordName, "(<keyword>) <symbol>")

	defineFunction("symbol?", ellSymbolP, "(<any>) <boolean>")
	defineFunction("symbol", ellSymbol, "(<any>+) <boolean>")

	defineFunction("string?", ellStringP, "(<any>) <boolean>")
	defineFunction("string", ellString, "(<any>*) <string>")
	defineFunction("to-string", ellToString, "(<any>) <string>")
	//	defineFunction("format", ellFormat, "(<string> <any>*) <string>")
	defineFunction("split", ellSplit, "(<any>) <sequence>")
	defineFunction("join", ellJoin, "(<sequence>) <any>")

	defineFunction("character?", ellCharacterP, "(<any>) <boolean>")
	defineFunction("to-character", ellToCharacter, "(<any>) <character>")

	defineFunction("number?", ellNumberP, "(<any>) <boolean>") // either float or int
	defineFunction("int?", ellIntP, "(<any>) <boolean>")       //int only
	defineFunction("float?", ellFloatP, "(<any>) <boolean>")   //float only
	defineFunction("inc", ellInc, "(<number>) <number>")
	defineFunction("dec", ellDec, "(<number>) <number>")
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

	defineFunction("struct?", ellStructP, "(<any>) <boolean>")
	defineFunction("to-struct", ellToStruct, "(<any>) <struct>")
	defineFunction("struct", ellStruct, "(<any>+) <struct>")
	defineFunction("has?", ellHasP, "(<struct> <any>) <boolean>")
	defineFunction("keys", ellKeys, "(<struct>) <list>")
	defineFunction("values", ellValues, "(<struct>) <list>")

	defineFunction("function?", ellFunctionP, "(<any>) <boolean>")
	defineFunction("function-signature", ellFunctionSignature, "(<function>) <string>")

	defineFunction("slurp", ellSlurp, "(<string>) <string>")
	defineFunction("read", ellRead, "(<string> keys: <type>) <list>")

	defineFunction("spit", ellSpit, "(<string> <any>) <null>")
	defineFunction("write", ellWrite, "(<any> indent: <string>) <string>")
	defineFunction("print", ellPrint, "(<any>*) <null>")
	defineFunction("println", ellPrintln, "(<any>*) <null>")

	defineFunction("macroexpand", ellMacroexpand, "(<any>) <any>")
	defineFunction("compile", ellCompile, "(<any>) <code>")

	defineFunction("empty?", ellEmptyP, "(<list|vector|struct|string>) <boolean>")
	defineFunction("length", ellLength, "(<list|vector|struct|string>) <number>")
	defineFunction("get", ellGet, "(<any> <any>) <any>")
	defineFunction("assoc", ellAssoc, "(<any> <any>) <any>")
	defineFunction("dissoc", ellDissoc, "(<any> <any>) <any>")
	defineFunction("assoc!", ellAssocBang, "(<list|struct> <any>) <struct")    //mutate!
	defineFunction("dissoc!", ellDissocBang, "(<list|struct> <any>) <struct>") //mutate!

	defineFunction("error", ellFatal, "(<any>+) <null>")
	defineFunction("json", ellJSON, "(<any>) <string>")

	err := loadModule("ell")
	if err != nil {
		fatal("*** ", err)
	}
}

func getOptions(rest []*LOB, keys ...string) (*LOB, error) {
	var validOptions []*LOB
	for _, key := range keys {
		validOptions = append(validOptions, intern(key))
	}
	return validateKeywordArgs(list(rest...), validOptions)
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *LList
//

func ellLetrec(argv []*LOB) (*LOB, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []*LOB) (*LOB, error) {
	return expandLet(argv[0])
}

func ellCond(argv []*LOB) (*LOB, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []*LOB) (*LOB, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(argv []*LOB) (*LOB, error) {
	return newString(Version), nil
}

func ellDefinedP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellSlurp(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("slurp", "1+", argc)
	}
	url, err := asString(argv[0])
	if err != nil {
		return nil, ArgTypeError("string", 1, argv[0])
	}
	options, err := getOptions(argv[1:argc], "headers:")
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		//do an http GET
		return nil, Error("slurp on URL NYI: ", url, options)
	}
	return slurpFile(url)
}

func ellSpit(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 2 {
		return nil, ArgcError("slurp", "2+", argc)
	}
	url, err := asString(argv[0])
	if err != nil {
		return nil, ArgTypeError("string", 1, argv[0])
	}
	data, err := asString(argv[1])
	if err != nil {
		return nil, ArgTypeError("string", 2, argv[1])
	}
	options, err := getOptions(argv[2:argc], "headers:")
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		//do an http GET
		return nil, Error("slurp on URL NYI: ", url, options, data)
	}
	err = spitFile(url, data)
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellRead(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("read", "1+", argc)
	}
	input := argv[0]
	options, err := getOptions(argv[1:argc], "keys:")
	if err != nil {
		return nil, err
	}
	return readAll(input, options)
}

func ellMacroexpand(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("macroexpand", "1", argc)
	}
	return macroexpand(argv[0])
}

func ellCompile(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("compile", "1", argc)
	}
	expanded, err := macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return compile(expanded)
}

func ellType(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("type", "1", argc)
	}
	return argv[0].variant, nil
}

func ellValue(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("value", "1", argc)
	}
	return value(argv[0]), nil
}

func ellInstance(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, ArgcError("instance", "2", argc)
	}
	return instance(argv[0], argv[1])
}

func ellValidateKeywordArgList(argv []*LOB) (*LOB, error) {
	//(validate-keyword-arg-list '(x: 23) x: y:) -> (x:)
	//(validate-keyword-arg-list '(x: 23 z: 100) x: y:) -> error("bad keyword z: in argument list")
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("validate-keyword-arg-list", "1+", argc)
	}
	if isList(argv[0]) {
		return validateKeywordArgList(argv[0], argv[1:argc])
	}
	return nil, ArgTypeError("list", 1, argv[0])
}

func ellKeys(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("keys", "1", argc)
	}
	strct := value(argv[0])
	if !isStruct(strct) {
		return nil, ArgTypeError("struct", 1, argv[0])
	}
	return structKeyList(strct), nil
}

func ellValues(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("values", "1", argc)
	}
	strct := value(argv[0])
	if !isStruct(strct) {
		return nil, ArgTypeError("struct", 1, argv[0])
	}
	return structValueList(strct), nil
}

func ellStruct(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	return newStruct(argv[:argc])
}

func ellToStruct(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("to-struct", "1", argc)
	}
	//how about a keys: keyword argument to force a key type?
	return toStruct(argv[0])
}

func ellIdenticalP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("identical?", "2", argc)
}

func ellEq(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellNumeq(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellDisplay(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		//todo: add the optional output argument like scheme
		return nil, ArgcError("display", "1", argc)
	}
	fmt.Printf("%v", argv[0])
	return Null, nil
}

func ellWrite(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("write", "1+", argc)
	}
	args := argv[:argc]
	lastNonkey := 0
	for _, a := range args {
		if isKeyword(a) {
			break
		}
		lastNonkey++
	}
	args = args[:lastNonkey]
	options, err := getOptions(argv[lastNonkey:argc], "indent:")
	if err != nil {
		return nil, err
	}
	indent := "" //not indented
	if err == nil && options != nil {
		s, _ := get(options, intern("indent:"))
		if isString(s) {
			indent = s.text
		}
	}
	result := ""
	first := true
	for _, data := range args {
		s := writeIndent(data, indent)
		if first {
			first = false
		} else {
			result += "\n"
		}
		result += s
	}
	return newString(result), nil
}

func ellFatal(argv []*LOB) (*LOB, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}

func ellToString(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("to-string", "1", argc)
	}
	return toString(argv[0])
}

func ellPrint(argv []*LOB) (*LOB, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []*LOB) (*LOB, error) {
	ellPrint(argv)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellReverse(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("reverse", "1", argc)
	}
	lst := argv[0]
	if !isList(lst) {
		return nil, ArgTypeError("list", 1, lst)
	}
	return reverse(lst), nil
}

func ellFlatten(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellList(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func ellNumberP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("number?", "1", argc)
}

func ellIntP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isInt(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("int?", "1", argc)
}

func ellFloatP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isFloat(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("float?", "1", argc)
}

func ellQuotient(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellRemainder(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellInc(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("inc", "1", argc)
	}
	return inc(argv[0])
}

func ellDec(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("dec", "1", argc)
	}
	return dec(argv[0])
}

func ellPlus(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	return minus(argv, argc)
}

func ellTimes(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellDiv(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	return div(argv, argc)
}

func ellVector(argv []*LOB) (*LOB, error) {
	return vector(argv...), nil
}

func ellToVector(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("to-vector", "1", argc)
	}
	return toVector(argv[0])
}

func ellMakeVector(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellVectorP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isVector(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("vector?", "1", argc)
}

func ellGe(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError(">=", "2", argc)
}

func ellLe(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError("<=", "2", argc)
}

func ellGt(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError(">", "2", argc)
}

func ellLt(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, ArgcError("<", "2", argc)
}

func ellZeroP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		n := argv[0]
		if n.variant == typeNumber {
			if numberEqual(n.fval, 0.0) {
				return True, nil
			}
			return False, nil
		}
		return nil, ArgTypeError("number", 1, n)
	}
	return nil, ArgcError("zero?", "1", argc)
}

func ellNumberToString(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("number->string", "1", argc)
	}
	if !isNumber(argv[0]) {
		return nil, ArgTypeError("number", 1, argv[0])
	}
	return newString(argv[0].String()), nil
}

func ellLength(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		n := length(argv[0])
		if n < 0 {
			return nil, Error("Cannot take length of ", argv[0].variant)
		}
		return newInt(n), nil
	}
	return nil, ArgcError("length", "1", argc)
}

func ellNot(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("not", "1", argc)
}

func ellNullP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if argv[0] == Null {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("nil?", "1", argc)
}

func ellBooleanP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("boolean?", "1", argc)
}

func ellSymbolP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("symbol?", "1", argc)
}

func ellSymbol(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("symbol", "1+", argc)
	}
	return symbol(argv[:argc])
}

func ellKeywordP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("keyword?", "1", argc)
}

func ellKeywordName(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isType(argv[0]) {
			return keywordName(argv[0])
		}
		return False, nil
	}
	return nil, ArgcError("type-name", "1", argc)
}

func ellTypeP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isType(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("type?", "1", argc)
}

func ellTypeName(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isType(argv[0]) {
			return typeName(argv[0])
		}
		return False, nil
	}
	return nil, ArgcError("type-name", "1", argc)
}

func ellStringP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("string?", "1", argc)
}

func ellCharacterP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isCharacter(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("character?", "1", argc)
}

func ellToCharacter(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("to-character", "1", argc)
	}
	return toCharacter(argv[0])
}

func ellFunctionP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isFunction(argv[0]) || isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("function?", "1", argc)
}

func ellFunctionSignature(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isFunction(argv[0]) {
			return newString(argv[0].function.signature()), nil
		}
		return nil, ArgTypeError("function", 1, argv[0])
	}
	return nil, ArgcError("function?", "1", argc)
}

func ellListP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("list?", "1", argc)
}

func ellEmptyP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("empty?", "1", argc)
	}
	if isEmpty(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellString(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	s := ""
	for i := 0; i < argc; i++ {
		s += argv[i].String()
	}
	return newString(s), nil
}

func ellCar(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if lst.variant == typeList {
			if lst == EmptyList {
				return Null, nil
			}
			return lst.car, nil
		}
		return nil, ArgTypeError("cdr", 1, lst)
	}
	return nil, ArgcError("car", "1", argc)
}

func ellCdr(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if lst.variant == typeList {
			if lst == EmptyList {
				return lst, nil
			}
			return lst.cdr, nil
		}
		return nil, ArgTypeError("cdr", 1, lst)
	}
	return nil, ArgcError("cdr", "1", argc)
}

func ellSetCarBang(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		err := setCar(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return Null, nil
	}
	return nil, ArgcError("set-car!", "2", argc)
}

func ellSetCdrBang(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		err := setCdr(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return Null, nil
	}
	return nil, ArgcError("set-cdr!", "2", argc)
}

func ellCaar(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caar(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("caar", "1", argc)
}

func ellCadr(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cadr", "1", argc)
}

func ellCddr(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cddr", "1", argc)
}

func ellCadar(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(cdr(car(lst))), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cadar", "1", argc)
}

func ellCaddr(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caddr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("caddr", "1", argc)
}
func ellCdddr(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdddr(lst), nil
		}
		return nil, ArgTypeError("list", 1, lst)
	}
	return nil, ArgcError("cdddr", "1", argc)
}

func ellCons(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		lst := argv[1]
		if !isList(lst) {
			return nil, ArgTypeError("list", 2, lst)
		}
		return cons(argv[0], lst), nil
	}
	return nil, ArgcError("cons", "2", argc)
}

func ellStructP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isStruct(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, ArgcError("struct?", "1", argc)
}

func ellGet(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, ArgcError("get", "2", argc)
	}
	v := value(argv[0])
	switch v.variant {
	case typeStruct:
		return structGet(v, argv[1]), nil
	case typeVector:
		idx, err := intValue(argv[1])
		if err != nil {
			return Null, nil
		}
		return vectorRef(v, idx), nil
	case typeList:
		lst := v
		idx, err := intValue(argv[1])
		if err != nil {
			return Null, nil
		}
		i := 0
		for lst != EmptyList {
			if i == idx {
				return lst.car, nil
			}
			i++
			lst = lst.cdr
		}
		return Null, nil
	case typeString:
		idx, err := intValue(argv[1])
		if err != nil {
			return Null, nil
		}
		return stringRef(v, idx), nil
	default:
		return nil, Error("get cannot work with type ", argv[0].variant)
	}
}

func ellHasP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
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

func ellAssoc(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 3 {
		return nil, ArgcError("assoc", "3+", argc)
	}
	return assoc(argv[0], argv[1:]...)
}

func ellDissoc(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 2 {
		return nil, ArgcError("assoc", "2+", argc) //clojure takes 1+, i.e. (disassoc s) is valid (and just returns a copy)
	}
	return dissoc(argv[0], argv[1:]...)
}

func ellAssocBang(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 3 {
		return nil, ArgcError("assoc!", "3+", argc)
	}
	s := value(argv[0])
	switch s.variant {
	case typeStruct:
		if argc == 3 {
			return put(s, argv[1], argv[2])
		}
		return assocBangStruct(s, argv[1:argc])
	case typeVector:
		if argc == 3 {
			idx, err := intValue(argv[1])
			if err != nil {
				return nil, err
			}
			err = vectorSet(s, idx, argv[2])
			return s, err
		}
		return assocBangVector(s, argv[1:argc])
	default:
		return nil, Error("assoc! cannot work with type ", argv[0].variant)
	}
}

func ellDissocBang(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 2 {
		return nil, ArgcError("dissoc!", "2+", argc)
	}
	return dissocBang(argv[0], argv[1:]...)
}

func ellToList(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, ArgcError("to-list", "1", argc)
	}
	return toList(argv[0])
}

func ellSplit(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, ArgcError("split", "2", argc)
	}
	return stringSplit(argv[0], argv[1])
}

func ellJoin(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, ArgcError("join", "2", argc)
	}
	return stringJoin(argv[0], argv[1])
}

func ellJSON(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, ArgcError("json", "1+", argc)
	}
	data := argv[0]
	options, err := getOptions(argv[1:argc], "indent:")
	if err != nil {
		return nil, err
	}
	indent := ""
	if err == nil && options != nil {
		s, _ := get(options, intern("indent:"))
		if isString(s) {
			indent = s.text
		}
	}
	s, err := writeToString(data, true, indent)
	if err != nil {
		return nil, err
	}
	return newString(s), nil
}
