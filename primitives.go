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
	"time"
)

const midi = true

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
	define("callcc", CallCC)

	defineTypedFunction("version", ellVersion, typeString, []*LOB{}, nil, nil)

	defineTypedFunction("boolean?", ellBooleanP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("not", ellNot, typeBoolean, []*LOB{typeAny}, nil, nil)

	defineTypedFunction("equal?", ellEqualP, typeBoolean, []*LOB{typeAny, typeAny}, nil, nil)
	defineTypedFunction("identical?", ellIdenticalP, typeBoolean, []*LOB{typeAny, typeAny}, nil, nil)

	defineTypedFunction("null?", ellNullP, typeBoolean, []*LOB{typeAny}, nil, nil)

	defineTypedFunction("def?", ellDefinedP, typeBoolean, []*LOB{typeSymbol}, nil, nil)

	defineTypedFunction("type", ellType, typeType, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("value", ellValue, typeAny, []*LOB{typeAny}, nil, nil)

	defineTypedFunction("instance", ellInstance, typeAny, []*LOB{typeType, typeAny}, nil, nil)

	defineTypedFunction("type?", ellTypeP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("type-name", ellTypeName, typeSymbol, []*LOB{typeType}, nil, nil)

	defineTypedFunction("keyword?", ellKeywordP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("keyword-name", ellKeywordName, typeSymbol, []*LOB{typeKeyword}, nil, nil)

	defineTypedFunction("symbol?", ellSymbolP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineFunction("symbol", ellSymbol, "(<any>+) <boolean>") // 1..N args

	defineTypedFunction("string?", ellStringP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineFunction("string", ellString, "(<any>*) <string>") // 0..N args
	defineTypedFunction("to-string", ellToString, typeString, []*LOB{typeAny}, nil, nil)
	//	defineFunction("format", ellFormat, "(<string> <any>*) <string>")
	defineTypedFunction("split", ellSplit, typeList, []*LOB{typeString, typeString}, nil, nil)
	defineFunction("join", ellJoin, "(<sequence> <string>) <any>") // <list|vector>

	defineTypedFunction("character?", ellCharacterP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("to-character", ellToCharacter, typeCharacter, []*LOB{typeAny}, nil, nil)

	defineTypedFunction("number?", ellNumberP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("int?", ellIntP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("float?", ellFloatP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("inc", ellInc, typeNumber, []*LOB{typeNumber}, nil, nil)
	defineTypedFunction("dec", ellDec, typeNumber, []*LOB{typeNumber}, nil, nil)
	defineTypedFunction("+", ellAdd, typeNumber, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("-", ellSub, typeNumber, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("*", ellMul, typeNumber, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("/", ellDiv, typeNumber, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("quotient", ellQuotient, typeNumber, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("remainder", ellRemainder, typeNumber, []*LOB{typeNumber, typeNumber}, nil, nil)
	defineTypedFunction("modulo", ellRemainder, typeNumber, []*LOB{typeNumber, typeNumber}, nil, nil) //fix

//	defineFunction("+", ellSum, "(<number>*) <number>")
//	defineFunction("-", ellDifference, "(<number>+) <number>")
//	defineFunction("*", ellProduct, "(<number>*) <number>")
//	defineFunction("/", ellDivision, "(<number>+) <number>")

	defineTypedFunction("=", ellNumEqual, typeBoolean, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("<=", ellNumLessEqual, typeBoolean, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction(">=", ellNumGreaterEqual, typeBoolean, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction(">", ellNumGreater, typeBoolean, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("<", ellNumLess, typeBoolean, []*LOB{typeNumber,typeNumber}, nil, nil)
	defineTypedFunction("zero?", ellZeroP, typeBoolean, []*LOB{typeNumber}, nil, nil)

	defineTypedFunction("list?", ellListP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("to-list", ellToList, typeList, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("cons", ellCons, typeList, []*LOB{typeAny, typeList}, nil, nil)
	defineTypedFunction("car", ellCar, typeAny, []*LOB{typeList}, nil, nil)
	defineTypedFunction("cdr", ellCdr, typeList, []*LOB{typeList}, nil, nil)
//	defineTypedFunction("list", ellList, typeList, []*LOB{typeAny}, nil, nil) //0..n args
//	defineFunction("list", ellList, "(<any>*) <list>") //could be defined as (defn list args args) more simply!
	defineFunction("concat", ellConcat, "(<list>*) <list>") //0..n args
	defineTypedFunction("reverse", ellReverse, typeList, []*LOB{typeList}, nil, nil)
	defineFunction("flatten", ellFlatten, "(<list|vector>) <list>") //union type
	defineTypedFunction("set-car!", ellSetCarBang, typeNull, []*LOB{typeList, typeAny}, nil, nil)
	defineTypedFunction("set-cdr!", ellSetCdrBang, typeNull, []*LOB{typeList, typeList}, nil, nil)

	defineTypedFunction("vector?", ellVectorP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("to-vector", ellToVector, typeVector, []*LOB{typeAny}, nil, nil)
	defineFunction("vector", ellVector, "(<any>*) <vector>") // n args
	defineTypedFunction("make-vector", ellMakeVector, typeVector, []*LOB{typeNumber, typeAny}, nil, nil)

	defineTypedFunction("struct?", ellStructP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("to-struct", ellToStruct, typeStruct, []*LOB{typeAny}, nil, nil)
	defineFunction("struct", ellStruct, "(<any>+) <struct>") // n args
	defineTypedFunction("has?", ellHasP, typeBoolean, []*LOB{typeStruct, typeAny}, nil, nil) // key is <symbol|keyword|type|string>
//	defineTypedFunction("keys", ellKeys, typeList, []*LOB{typeStruct}, nil, nil)
	defineFunction("keys", ellKeys, "(<struct|instance>) <list>")
//	defineTypedFunction("values", ellValues, typeList, []*LOB{typeStruct}, nil, nil)
	defineFunction("values", ellValues, "(<struct|instance>) <list>")

	defineTypedFunction("function?", ellFunctionP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("function-signature", ellFunctionSignature, typeString, []*LOB{typeFunction}, nil, nil)
	defineFunction("validate-keyword-arg-list", ellValidateKeywordArgList, "(<list> <keyword>+) <list>") // n-args

	defineTypedFunction("slurp", ellSlurp, typeString, []*LOB{typeString}, nil, nil)
//	defineTypedFunction("read", ellRead, typeAny, []*LOB{typeString, typeType}, []*LOB{typeAny}, []*LOB{intern("keys:")})
	defineFunction("read", ellRead, "(<string> keys: <type>) <list>") //key args

	defineTypedFunction("spit", ellSpit, typeNull, []*LOB{typeString, typeAny}, nil, nil)
//NYI	defineTypedFunction("write", ellWrite, typeNull, []*LOB{typeAny, typeString}, []*LOB{EmptyString}, []*LOB{intern("indent:")})
	defineFunction("write", ellWrite, "(<any> indent: <string>) <string>")
	defineFunction("print", ellPrint, "(<any>*) <null>") // n args
	defineFunction("println", ellPrintln, "(<any>*) <null>") // n args

	defineTypedFunction("macroexpand", ellMacroexpand, typeAny, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("compile", ellCompile, typeCode, []*LOB{typeAny}, nil, nil)

	defineFunction("empty?", ellEmptyP, "(<list|vector|struct|string>) <boolean>") // union
	defineFunction("length", ellLength, "(<list|vector|struct|string>) <number>")  // union
	defineFunction("get", ellGet, "(<any> <any>) <any>") //union
	defineFunction("assoc", ellAssoc, "(<any> <any>) <any>") //union
	defineFunction("dissoc", ellDissoc, "(<any> <any>) <any>") //union
	defineFunction("assoc!", ellAssocBang, "(<list|struct> <any>) <struct")    //mutate!
	defineFunction("dissoc!", ellDissocBang, "(<list|struct> <any>) <struct>") //mutate!

//	defineTypedFunction("make-error", ellMakeError, typeError, []*LOB{typeAny}, nil, nil)
	defineFunction("make-error", ellMakeError, "(<any>+) <error>") //nargs
	defineTypedFunction("error?", ellErrorP, typeBoolean, []*LOB{typeAny}, nil, nil)
	defineTypedFunction("uncaught-error", ellUncaughtError, typeNull, []*LOB{typeError}, nil, nil) //doesn't return
//	defineTypedFunction("json", ellJSON, typeString, []*LOB{typeAny}, nil, nil)
	defineFunction("json", ellJSON, "(<any> indent: <string>) <string>") //keys

//	defineTypedFunction("getfn", ellGetFn, typeAny, []*LOB{typeSymbol, typeList}, nil, nil) //union <function|null>
	defineFunction("getfn", ellGetFn, "(<symbol> <any>*)") //nargs
	defineTypedFunction("method-signature", ellMethodSignature, typeType, []*LOB{typeList}, nil, nil)

//	defineTypedFunction("spawn", ellSpawn, typeNull, []*LOB{typeFunction}, nil, nil) //union <function|null>
	defineFunction("spawn", ellSpawn, "(<function> <any>*)") //nargs
	//	defineTypedFunction("channel", ellChannel, typeChannel, []*LOB{typeString, typeNumber}, []*LOB{EmptyString, Zero}, []*LOB{intern("name:"), intern("bufsize:")})
	defineFunction("channel", ellChannel, "(<channel>)") // keys
	defineFunction("send", ellSend, "(<channel> <any> <number>?) <boolean>") //optional 3rd arg for timeout
	defineFunction("recv", ellReceive, "(<channel> <number>?) <any>") //optional second arg for timeout
	defineTypedFunction("close", ellClose, typeNull, []*LOB{typeChannel}, nil, nil)

//	if midi {
//		initMidi()
//	}
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
	if isDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSlurp(argv []*LOB) (*LOB, error) {
	return slurpFile(argv[0].text)
}

func ellSpit(argv []*LOB) (*LOB, error) {
	url := argv[0].text
	data := argv[1].text
	err := spitFile(url, data)
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellRead(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "read expected at least 1 argument, got none")
	}
	input := argv[0]
	options, err := getOptions(argv[1:argc], "keys:")
	if err != nil {
		return nil, err
	}
	return readAll(input, options)
}

func ellMacroexpand(argv []*LOB) (*LOB, error) {
	return macroexpand(argv[0])
}

func ellCompile(argv []*LOB) (*LOB, error) {
	expanded, err := macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return compile(expanded)
}

func ellType(argv []*LOB) (*LOB, error) {
	return argv[0].variant, nil
}

func ellValue(argv []*LOB) (*LOB, error) {
	return value(argv[0]), nil
}

func ellInstance(argv []*LOB) (*LOB, error) {
	return instance(argv[0], argv[1])
}

func ellValidateKeywordArgList(argv []*LOB) (*LOB, error) {
	//(validate-keyword-arg-list '(x: 23) x: y:) -> (x:)
	//(validate-keyword-arg-list '(x: 23 z: 100) x: y:) -> error("bad keyword z: in argument list")
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "validate-keyword-arg-list expected at least 1 argument, got none")
	}
	if isList(argv[0]) {
		return validateKeywordArgList(argv[0], argv[1:argc])
	}
	return nil, Error(ArgumentErrorKey, "validate-keyword-arg-list expected a <list>, got a ", argv[0].variant)
}

func ellKeys(argv []*LOB) (*LOB, error) {
	strct := value(argv[0])
	if !isStruct(strct) {
		return nil, Error(ArgumentErrorKey, "keys expected a <struct>, got a ", argv[0].variant)
	}
	return structKeyList(strct), nil
}

func ellValues(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "values expected 1 argument, got ", argc)
	}
	strct := value(argv[0])
	if !isStruct(strct) {
		return nil, Error(ArgumentErrorKey, "values expected a <struct>, got a ", argv[0].variant)
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
		return nil, Error(ArgumentErrorKey, "to-struct expected 1 argument, got ", argc)
	}
	//how about a keys: keyword argument to force a key type?
	return toStruct(argv[0])
}

func ellIdenticalP(argv []*LOB) (*LOB, error) {
	if argv[0] == argv[1] {
		return True, nil
	}
	return False, nil
}

func ellEqualP(argv []*LOB) (*LOB, error) {
	if equal(argv[0], argv[1]) {
		return True, nil
	}
	return False, nil
}

func ellNumEqual(argv []*LOB) (*LOB, error) {
	if numberEqual(argv[0].fval, argv[1].fval) {
		return True, nil
	}
	return False, nil
}

func ellNumLess(argv []*LOB) (*LOB, error) {
	if argv[0].fval < argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellNumLessEqual(argv []*LOB) (*LOB, error) {
	if argv[0].fval <= argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellNumGreater(argv []*LOB) (*LOB, error) {
	if argv[0].fval > argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellNumGreaterEqual(argv []*LOB) (*LOB, error) {
	if argv[0].fval >= argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellWrite(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "write expected at least 1 argument, got ", argc)
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

func ellMakeError(argv []*LOB) (*LOB, error) {
	return newError(argv...), nil
}

func ellErrorP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "error? expected 1 argument, got ", argc)
	}
	if isError(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellUncaughtError(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "uncaught-error expected 1 argument, got ", argc)
	}
	if !isError(argv[0]) {
		return nil, Error(ArgumentErrorKey, "uncaught-error expected an <error>, got a ", argv[0].variant)
	}
	return nil, argv[0]
}

func ellToString(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "to-string expected 1 argument, got ", argc)
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
			return nil, Error(ArgumentErrorKey, "concat expected a <list>, got a ", argv[0].variant)
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
		return nil, Error(ArgumentErrorKey, "reverse expected 1 argument, got ", argc)
	}
	lst := argv[0]
	if !isList(lst) {
		return nil, Error(ArgumentErrorKey, "reverse expected a <list>, got a ", argv[0].variant)
	}
	return reverse(lst), nil
}

func ellFlatten(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "flatten expected 1 argument, got ", argc)
	}
	seq := argv[0]
	switch seq.variant {
	case typeList:
		return flatten(seq), nil
	case typeVector:
		lst, _ := toList(seq)
		return flatten(lst), nil
	default:
		return nil, Error(ArgumentErrorKey, "flatten expected a <list>, got a ", seq.variant)
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
	if isNumber(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellIntP(argv []*LOB) (*LOB, error) {
	if isInt(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFloatP(argv []*LOB) (*LOB, error) {
	if isFloat(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellInc(argv []*LOB) (*LOB, error) {
	return newFloat64(argv[0].fval + 1), nil
}

func ellDec(argv []*LOB) (*LOB, error) {
	return newFloat64(argv[0].fval - 1), nil
}

func ellAdd(argv []*LOB) (*LOB, error) {
	return newFloat64(argv[0].fval + argv[1].fval), nil
}

func ellSub(argv []*LOB) (*LOB, error) {
	return newFloat64(argv[0].fval - argv[1].fval), nil
}

func ellMul(argv []*LOB) (*LOB, error) {
	return newFloat64(argv[0].fval * argv[1].fval), nil
}

func ellDiv(argv []*LOB) (*LOB, error) {
	return newFloat64(argv[0].fval / argv[1].fval), nil
}

func ellQuotient(argv []*LOB) (*LOB, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "quotient: divide by zero")
	}
	return newInt64(int64(argv[0].fval) / denom), nil
}

func ellRemainder(argv []*LOB) (*LOB, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "remainder: divide by zero")
	}
	return newInt64(int64(argv[0].fval) % denom), nil
}

func ellVector(argv []*LOB) (*LOB, error) {
	return vector(argv...), nil
}

func ellToVector(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "to-vector expected 1 argument, got ", argc)
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
				return nil, Error(ArgumentErrorKey, "make-vector expected 1 or 2 arguments, got ", argc)
			}
			initVal = argv[1]
		}
		return newVector(int(vlen), initVal), nil
	}
	return nil, Error(ArgumentErrorKey, "make-vector expected 1 or 2 arguments, got ", argc)
}

func ellVectorP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isVector(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "vector? expected 1 argument, got ", argc)
}

func ellZeroP(argv []*LOB) (*LOB, error) {
	if numberEqual(argv[0].fval, 0.0) {
		return True, nil
	}
	return False, nil
}

func ellLength(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		n := length(argv[0])
		if n < 0 {
			return nil, Error(ArgumentErrorKey, "Cannot take length of ", argv[0].variant)
		}
		return newInt(n), nil
	}
	return nil, Error(ArgumentErrorKey, "length expected 1 argument, got ", argc)
}

func ellNot(argv []*LOB) (*LOB, error) {
	if argv[0] == False {
		return True, nil
	}
	return False, nil
}

func ellNullP(argv []*LOB) (*LOB, error) {
	if argv[0] == Null {
		return True, nil
	}
	return False, nil
}

func ellBooleanP(argv []*LOB) (*LOB, error) {
	if isBoolean(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbolP(argv []*LOB) (*LOB, error) {
	if isSymbol(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbol(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	return symbol(argv[:argc])
}

func ellKeywordP(argv []*LOB) (*LOB, error) {
	if isKeyword(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellKeywordName(argv []*LOB) (*LOB, error) {
	return keywordName(argv[0])
}

func ellTypeP(argv []*LOB) (*LOB, error) {
	if isType(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellTypeName(argv []*LOB) (*LOB, error) {
	return typeName(argv[0])
}

func ellStringP(argv []*LOB) (*LOB, error) {
	if isString(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellCharacterP(argv []*LOB) (*LOB, error) {
	if isCharacter(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellToCharacter(argv []*LOB) (*LOB, error) {
	return toCharacter(argv[0])
}

func ellFunctionP(argv []*LOB) (*LOB, error) {
//	if isFunction(argv[0]) || isKeyword(argv[0]) {
	if isFunction(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFunctionSignature(argv []*LOB) (*LOB, error) {
	return newString(functionSignature(argv[0])), nil
}

func ellListP(argv []*LOB) (*LOB, error) {
	if isList(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellEmptyP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "empty? expected 1 argument, got ", argc)
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
	lst := argv[0]
	if lst == EmptyList {
		return Null, nil
	}
	return lst.car, nil
}

func ellCdr(argv []*LOB) (*LOB, error) {
	lst := argv[0]
	if lst == EmptyList {
		return lst, nil
	}
	return lst.cdr, nil
}

func ellSetCarBang(argv []*LOB) (*LOB, error) {
	err := setCar(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellSetCdrBang(argv []*LOB) (*LOB, error) {
	err := setCdr(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellCons(argv []*LOB) (*LOB, error) {
	return cons(argv[0], argv[1]), nil
}

func ellStructP(argv []*LOB) (*LOB, error) {
	if isStruct(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellGet(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, Error(ArgumentErrorKey, "get expected 2 arguments, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "get cannot work with type ", argv[0].variant)
	}
}

func ellHasP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, Error(ArgumentErrorKey, "has? expected 2 arguments, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "assoc expected at least 3 arguments, got ", argc)
	}
	return assoc(argv[0], argv[1:]...)
}

func ellDissoc(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 2 {
		return nil, Error(ArgumentErrorKey, "assoc expected at least 2 arguments, got ", argc)
	}
	return dissoc(argv[0], argv[1:]...)
}

func ellAssocBang(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 3 {
		return nil, Error(ArgumentErrorKey, "assoc! expected at least 3 arguments, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "assoc! cannot work with type ", argv[0].variant)
	}
}

func ellDissocBang(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 2 {
		return nil, Error(ArgumentErrorKey, "dissoc! expected at least 2 arguments, got ", argc)
	}
	return dissocBang(argv[0], argv[1:]...)
}

func ellToList(argv []*LOB) (*LOB, error) {
	return toList(argv[0])
}

func ellSplit(argv []*LOB) (*LOB, error) {
	return stringSplit(argv[0], argv[1])
}

func ellJoin(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, Error(ArgumentErrorKey, "join expected 2 arguments, got ", argc)
	}
	return stringJoin(argv[0], argv[1])
}

func ellJSON(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "json expected at least 1 argument, got none")
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

func ellGetFn(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "getfn expected at least 1 argument, got none")
	}
	sym := argv[0]
	if sym.variant != typeSymbol {
		return nil, Error(ArgumentErrorKey, "getfn expected a <symbol> for argument 1, got ", sym)
	}
	return getfn(sym, argv[1:])
}

func ellMethodSignature(argv []*LOB) (*LOB, error) {
	return methodSignature(argv[0])
}

//spawn a new "go routine" for the thunk
func ellSpawn(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc >= 1 {
		fun := argv[0]
		if isFunction(fun) {
			if fun.code != nil {
				spawn(fun.code, argv[1:])
				return Null, nil
			}
		}
	}
	return nil, Error(ArgumentErrorKey, "spawn expects 1 function")
}

func ellChannel(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	bufsize := 0
	name := ""
	if argc > 0 {
		options, err := getOptions(argv, "name:", "bufsize:")
		if err != nil {
			return nil, err
		}
		opt := structGet(options, intern("name:"))
		if opt != Null {
			switch opt.variant {
			case typeString, typeSymbol:
				name = opt.text
			default:
				return nil, Error(ArgumentErrorKey, "channel expected a <string> or <symbol> as name: argument, got ", opt)
			}
		}
		opt = structGet(options, intern("bufsize:"))
		if opt != Null {
			bufsize, err = intValue(opt)
			if err != nil {
				return nil, Error(ArgumentErrorKey, "channel expected a <number> as bufsize: argument, got ", opt)
			}
		}
	}
	return newChannel(bufsize, name), nil
}

func ellClose(argv []*LOB) (*LOB, error) {
	ch := argv[0]
	if ch.channel != nil {
		close(ch.channel)
		ch.channel = nil
	}
	return Null, nil
}

func ellSend(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	timeout := 0
	if argc == 3 {
		f, err := floatValue(argv[2])
		if err != nil {
			return nil, Error(ArgumentErrorKey, "send expected a number for its optional third argument, got ", argv[1])
		}
		timeout = int(f * 1000)
	}
	if argc < 2 || argc > 3 {
		return nil, Error(ArgumentErrorKey, "send expected 2 or 3 arguments, got ", argc)
	}
	ch := argv[0]
	if ch.variant != typeChannel {
		return nil, Error(ArgumentErrorKey, "send expected a <channel> for argument 1, got ", ch)
	}
	val := argv[1]
	if ch.channel != nil {
		if timeout > 0 {
			dur := time.Millisecond * time.Duration(timeout)
			select {
			case ch.channel <- val:
				return True, nil
			case <-time.After(dur):
			}
		} else {
			select {
			case ch.channel <- val:
				return True, nil
			default:
			}
		}
	}
	return False, nil
}

func ellReceive(argv []*LOB) (*LOB, error) {
	timeout := 0
	argc := len(argv)
	if argc == 2 {
		f, err := floatValue(argv[1])
		if err != nil {
			return nil, Error(ArgumentErrorKey, "recv expected a number for its optional second argument, got ", argv[1])
		}
		timeout = int(f * 1000)
	}
	if argc < 1 || argc > 2 {
		return nil, Error(ArgumentErrorKey, "recv expected 1 or 2 arguments, got ", argc)
	}
	ch := argv[0]
	if ch.variant != typeChannel {
		return nil, Error(ArgumentErrorKey, "recv expected a <channel> for argument 1, got ", ch)
	}
	if ch.channel != nil {
		if timeout > 0 {
			dur := time.Millisecond * time.Duration(timeout)
			select {
			case val, ok := <-ch.channel:
				if ok {
					return val, nil
				}
			case <-time.After(dur):
			}
		} else {
			select {
			case val, ok := <-ch.channel:
				if ok {
					return val, nil
				}
			default:
			}
		}
	}
	return Null, nil
}
