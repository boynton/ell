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

	defineFunction("make-error", ellMakeError, "(<any>+) <error>")
	defineFunction("error?", ellIsError, "(<any>) <boolean>")
	defineFunction("uncaught-error", ellUncaughtError, "(<error>) <null>")
	defineFunction("json", ellJSON, "(<any>) <string>")
	defineFunction("getfn", ellGetFn, "(<symbol> <list>) <any>")
	defineFunction("method-signature", ellMethodSignature, "(<list>) <type>")

	defineFunction("spawn", ellSpawn, "(<function>)")
	defineFunction("channel", ellChannel, "(<channel>)")
	defineFunction("send", ellSend, "(<channel> <any>) <boolean>")
	defineFunction("recv", ellReceive, "(<channel>) <any>")
	defineFunction("close", ellClose, "(<channel>)")

	if midi {
		initMidi()
	}
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
		return nil, Error(ArgumentErrorKey, "defined? expected 1 argument, got ", argc)
	}
	if !isSymbol(argv[0]) {
		return nil, Error(ArgumentErrorKey, "defined? expected a <symbol>, got a ", argv[0].variant)
	}
	if isDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSlurp(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "slurp expected at least 1 argument, got none")
	}
	url, err := asString(argv[0])
	if err != nil {
		return nil, Error(ArgumentErrorKey, "slurp expected a <string>, got a ", argv[0].variant)
	}
	options, err := getOptions(argv[1:argc], "headers:")
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		//do an http GET
		return nil, Error(ArgumentErrorKey, "slurp on URL NYI: ", url, options)
	}
	return slurpFile(url)
}

func ellSpit(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 2 {
		return nil, Error(ArgumentErrorKey, "spit expected at least 2 arguments, got ", argc)
	}
	url, err := asString(argv[0])
	if err != nil {
		return nil, Error(ArgumentErrorKey, "spit expected a <string> for argument 1, got a ", argv[0].variant)
	}
	data, err := asString(argv[1])
	if err != nil {
		return nil, Error(ArgumentErrorKey, "spit expected a <string> for argument 2, got a ", argv[1].variant)
	}
	options, err := getOptions(argv[2:argc], "headers:")
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		//do an http GET
		return nil, Error(ArgumentErrorKey, "slurp on URL NYI: ", url, options, data)
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
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "macroexpand expected 1 argument, got ", argc)
	}
	return macroexpand(argv[0])
}

func ellCompile(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "compile expected 1 argument, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "type expected 1 argument, got ", argc)
	}
	return argv[0].variant, nil
}

func ellValue(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "value expected 1 argument, got ", argc)
	}
	return value(argv[0]), nil
}

func ellInstance(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, Error(ArgumentErrorKey, "instance expected 2 arguments, got ", argc)
	}
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
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "keys expected 1 argument, got ", argc)
	}
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
	argc := len(argv)
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "identical? expected 2 arguments, got ", argc)
}

func ellEq(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "equal? expected at least 1 argument, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "= expected at least 1 argument, got ", argc)
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if b, err := numericallyEqual(obj, argv[1]); err != nil || b == False {
			return b, err
		}
	}
	return True, nil
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

func ellIsError(argv []*LOB) (*LOB, error) {
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
	argc := len(argv)
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "number? expected 1 argument, got ", argc)
}

func ellIntP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isInt(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "int? expected 1 argument, got ", argc)
}

func ellFloatP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isFloat(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "float? expected 1 argument, got ", argc)
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
			return nil, Error(ArgumentErrorKey, "Quotient: divide by zero")
		}
		return newInt64(n1 / n2), nil
	}
	return nil, Error(ArgumentErrorKey, "quotient expected 2 arguments, got ", argc)
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
			return nil, Error(ArgumentErrorKey, "remainder: divide by zero")
		}
		if err != nil {
			return nil, err
		}
		return newInt64(n1 % n2), nil
	}
	return nil, Error(ArgumentErrorKey, "remainder expected 2 arguments, got ", argc)
}

func ellInc(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "inc expected 1 argument, got ", argc)
	}
	return inc(argv[0])
}

func ellDec(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "dec expected 1 argument, got ", argc)
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

func ellGe(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, Error(ArgumentErrorKey, ">= expected 2 arguments, got ", argc)
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
	return nil, Error(ArgumentErrorKey, "<= expected 2 arguments, got ", argc)
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
	return nil, Error(ArgumentErrorKey, "> expected 2 arguments, got ", argc)
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
	return nil, Error(ArgumentErrorKey, "< expected 2 arguments, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "zero? expected a <number>, got a ", n.variant)
	}
	return nil, Error(ArgumentErrorKey, "zero? expected 1 argument, got ", argc)
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
	argc := len(argv)
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "not expected 1 argument, got ", argc)
}

func ellNullP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if argv[0] == Null {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "null? expected 1 argument, got ", argc)
}

func ellBooleanP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "boolean? expected 1 argument, got ", argc)
}

func ellSymbolP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "symbol? expected 1 argument, got ", argc)
}

func ellSymbol(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
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
	return nil, Error(ArgumentErrorKey, "keyword? expected 1 argument, got ", argc)
}

func ellKeywordName(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isType(argv[0]) {
			return keywordName(argv[0])
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "keyword-name expected 1 argument, got ", argc)
}

func ellTypeP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isType(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "type? expected 1 argument, got ", argc)
}

func ellTypeName(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isType(argv[0]) {
			return typeName(argv[0])
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "type-name expected 1 argument, got ", argc)
}

func ellStringP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "string? expected 1 argument, got ", argc)
}

func ellCharacterP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isCharacter(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "character? expected 1 argument, got ", argc)
}

func ellToCharacter(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "to-character expected 1 argument, got ", argc)
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
	return nil, Error(ArgumentErrorKey, "function? expected 1 argument, got ", argc)
}

func ellFunctionSignature(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isFunction(argv[0]) {
			return newString(functionSignature(argv[0])), nil
		}
		return nil, Error(ArgumentErrorKey, "function-signature expected a <function>, got a ", argv[0].variant)
	}
	return nil, Error(ArgumentErrorKey, "function-signature expected 1 argument, got ", argc)
}

func ellListP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "list? expected 1 argument, got ", argc)
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
	argc := len(argv)
	if argc == 1 {
		lst := argv[0]
		if lst.variant == typeList {
			if lst == EmptyList {
				return Null, nil
			}
			return lst.car, nil
		}
		return nil, Error(ArgumentErrorKey, "car expected a <list>, got a ", argv[0].variant)
	}
	return nil, Error(ArgumentErrorKey, "car expected 1 argument, got ", argc)
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
		return nil, Error(ArgumentErrorKey, "cdr expected a <list>, got a ", argv[0].variant)
	}
	return nil, Error(ArgumentErrorKey, "cdr expected 1 argument, got ", argc)
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
	return nil, Error(ArgumentErrorKey, "set-car! expected 2 arguments, got ", argc)
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
	return nil, Error(ArgumentErrorKey, "set-cdr! expected 2 arguments, got ", argc)
}

func ellCons(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 2 {
		lst := argv[1]
		if !isList(lst) {
			return nil, Error(ArgumentErrorKey, "cons expected a <list> for argument 2, got a ", argv[1].variant)
		}
		return cons(argv[0], lst), nil
	}
	return nil, Error(ArgumentErrorKey, "cons expected 2 arguments, got ", argc)
}

func ellStructP(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		if isStruct(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return nil, Error(ArgumentErrorKey, "struct? expected 1 argument, got ", argc)
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
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "to-list expected 1 argument, got ", argc)
	}
	return toList(argv[0])
}

func ellSplit(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc != 2 {
		return nil, Error(ArgumentErrorKey, "split expected 2 arguments, got ", argc)
	}
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
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "method-signature expected 1 argument, got ", argc)
	}
	formalArgs := argv[0]
	if formalArgs.variant != typeList {
		return nil, Error(ArgumentErrorKey, "method-signature expected a <list>, got ", formalArgs)
	}
	return methodSignature(formalArgs)
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
	argc := len(argv)
	if argc != 1 {
		return nil, Error(ArgumentErrorKey, "close expected 1 argument, got ", argc)
	}
	ch := argv[0]
	if ch.variant != typeChannel {
		return nil, Error(ArgumentErrorKey, "close expected a <channel> for argument 1, got ", ch)
	}
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
