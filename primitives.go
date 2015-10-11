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

	defineGlobal("null", Null)
	defineGlobal("true", True)
	defineGlobal("false", False)

	defineGlobal("apply", Apply)
	defineGlobal("callcc", CallCC)
	defineGlobal("spawn", Spawn)

	defineFunction("version", ellVersion, typeString)
	defineFunction("boolean?", ellBooleanP, typeBoolean, typeAny)
	defineFunction("not", ellNot, typeBoolean, typeAny)
	defineFunction("equal?", ellEqualP, typeBoolean, typeAny, typeAny)
	defineFunction("identical?", ellIdenticalP, typeBoolean, typeAny, typeAny)
	defineFunction("null?", ellNullP, typeBoolean, typeAny)
	defineFunction("def?", ellDefinedP, typeBoolean, typeSymbol)

	defineFunction("type", ellType, typeType, typeAny)
	defineFunction("value", ellValue, typeAny, typeAny)
	defineFunction("instance", ellInstance, typeAny, typeType, typeAny)

	defineFunction("type?", ellTypeP, typeBoolean, typeAny)
	defineFunction("type-name", ellTypeName, typeSymbol, typeType)
	defineFunction("keyword?", ellKeywordP, typeBoolean, typeAny)
	defineFunction("keyword-name", ellKeywordName, typeSymbol, typeKeyword)
	defineFunction("symbol?", ellSymbolP, typeBoolean, typeAny)
	defineFunctionRestArgs("symbol", ellSymbol, typeSymbol, typeAny, typeAny) //"(<any> <any>*) <symbol>")

	defineFunctionRestArgs("string?", ellStringP, typeBoolean, typeAny)
	defineFunctionRestArgs("string", ellString, typeString, typeAny) //"(<any>*) <string>")
	defineFunction("to-string", ellToString, typeString, typeAny)
	defineFunction("string-length", ellStringLength, typeNumber, typeString)
	defineFunction("split", ellSplit, typeList, typeString, typeString)
	defineFunction("join", ellJoin, typeList, typeList, typeString) // <list|vector> for both arg1 and result could work
	defineFunction("character?", ellCharacterP, typeBoolean, typeAny)
	defineFunction("to-character", ellToCharacter, typeCharacter, typeAny)

	defineFunction("number?", ellNumberP, typeBoolean, typeAny)
	defineFunction("int?", ellIntP, typeBoolean, typeAny)
	defineFunction("float?", ellFloatP, typeBoolean, typeAny)
	defineFunction("inc", ellInc, typeNumber, typeNumber)
	defineFunction("dec", ellDec, typeNumber, typeNumber)
	defineFunction("+", ellAdd, typeNumber, typeNumber, typeNumber)
	defineFunction("-", ellSub, typeNumber, typeNumber, typeNumber)
	defineFunction("*", ellMul, typeNumber, typeNumber, typeNumber)
	defineFunction("/", ellDiv, typeNumber, typeNumber, typeNumber)
	defineFunction("quotient", ellQuotient, typeNumber, typeNumber, typeNumber)
	defineFunction("remainder", ellRemainder, typeNumber, typeNumber, typeNumber)
	defineFunction("modulo", ellRemainder, typeNumber, typeNumber, typeNumber) //fix
	defineFunction("=", ellNumEqual, typeBoolean, typeNumber, typeNumber)
	defineFunction("<=", ellNumLessEqual, typeBoolean, typeNumber, typeNumber)
	defineFunction(">=", ellNumGreaterEqual, typeBoolean, typeNumber, typeNumber)
	defineFunction(">", ellNumGreater, typeBoolean, typeNumber, typeNumber)
	defineFunction("<", ellNumLess, typeBoolean, typeNumber, typeNumber)
	defineFunction("zero?", ellZeroP, typeBoolean, typeNumber)

	defineFunction("seal!", ellSeal, typeAny, typeAny) //actually only list, vector, and struct for now

	defineFunction("list?", ellListP, typeBoolean, typeAny)
	defineFunction("empty?", ellEmptyP, typeBoolean, typeList)
	defineFunction("to-list", ellToList, typeList, typeAny)
	defineFunction("cons", ellCons, typeList, typeAny, typeList)
	defineFunction("car", ellCar, typeAny, typeList)
	defineFunction("cdr", ellCdr, typeList, typeList)
	defineFunction("set-car!", ellSetCarBang, typeNull, typeList, typeAny)
	defineFunction("set-cdr!", ellSetCdrBang, typeNull, typeList, typeList)
	defineFunction("list-length", ellListLength, typeNumber, typeList)
	defineFunction("reverse", ellReverse, typeList, typeList)
	defineFunctionRestArgs("list", ellList, typeList, typeAny)
	defineFunctionRestArgs("concat", ellConcat, typeList, typeList)
	defineFunctionRestArgs("flatten", ellFlatten, typeList, typeList)

	defineFunction("vector?", ellVectorP, typeBoolean, typeAny)
	defineFunction("to-vector", ellToVector, typeVector, typeAny)
	defineFunctionRestArgs("vector", ellVector, typeVector, typeAny)
	defineFunctionOptionalArgs("make-vector", ellMakeVector, typeVector, []*LOB{typeNumber, typeAny}, Null)
	defineFunction("vector-length", ellVectorLength, typeNumber, typeVector)
	defineFunction("vector-ref", ellVectorRef, typeAny, typeVector, typeNumber)
	defineFunction("vector-set!", ellVectorSetBang, typeNull, typeVector, typeNumber, typeAny)

	defineFunction("struct?", ellStructP, typeBoolean, typeAny)
	defineFunction("to-struct", ellToStruct, typeStruct, typeAny)
	defineFunctionRestArgs("struct", ellStruct, typeStruct, typeAny)
	defineFunction("make-struct", ellMakeStruct, typeStruct, typeNumber)
	defineFunction("struct-length", ellStructLength, typeNumber, typeStruct)
	defineFunction("has?", ellHasP, typeBoolean, typeStruct, typeAny) // key is <symbol|keyword|type|string>
	defineFunction("get", ellGet, typeAny, typeStruct, typeAny)
	defineFunction("put!", ellPutBang, typeNull, typeStruct, typeAny, typeAny)
	defineFunction("unput!", ellUnputBang, typeNull, typeStruct, typeAny)
	defineFunction("keys", ellKeys, typeList, typeAny)     // <struct|instance>
	defineFunction("values", ellValues, typeList, typeAny) // <struct|instance>

	defineFunction("function?", ellFunctionP, typeBoolean, typeAny)
	defineFunction("function-signature", ellFunctionSignature, typeString, typeFunction)
	defineFunctionRestArgs("validate-keyword-arg-list", ellValidateKeywordArgList, typeList, typeKeyword, typeList)
	defineFunction("slurp", ellSlurp, typeString, typeString)
	defineFunctionKeyArgs("read", ellRead, typeAny, []*LOB{typeString, typeType}, []*LOB{typeAny}, []*LOB{intern("keys:")})
	defineFunction("spit", ellSpit, typeNull, typeString, typeAny)
	defineFunctionKeyArgs("write", ellWrite, typeNull, []*LOB{typeAny, typeString}, []*LOB{EmptyString}, []*LOB{intern("indent:")})
	defineFunctionRestArgs("print", ellPrint, typeNull, typeAny)
	defineFunctionRestArgs("println", ellPrintln, typeNull, typeAny)
	defineFunction("macroexpand", ellMacroexpand, typeAny, typeAny)
	defineFunction("compile", ellCompile, typeCode, typeAny)

	defineFunctionRestArgs("make-error", ellMakeError, typeError, typeAny)
	defineFunction("error?", ellErrorP, typeBoolean, typeAny)
	defineFunction("uncaught-error", ellUncaughtError, typeNull, typeError) //doesn't return

	defineFunctionKeyArgs("json", ellJSON, typeString, []*LOB{typeAny, typeString}, []*LOB{EmptyString}, []*LOB{intern("indent:")})

	defineFunctionRestArgs("getfn", ellGetFn, typeFunction, typeAny, typeSymbol)
	defineFunction("method-signature", ellMethodSignature, typeType, typeList)

	defineFunctionKeyArgs("channel", ellChannel, typeChannel, []*LOB{typeString, typeNumber}, []*LOB{EmptyString, Zero}, []*LOB{intern("name:"), intern("bufsize:")})
	defineFunctionOptionalArgs("send", ellSend, typeNull, []*LOB{typeChannel, typeAny, typeNumber}, MinusOne)
	defineFunctionOptionalArgs("recv", ellReceive, typeAny, []*LOB{typeChannel, typeNumber}, MinusOne)
	defineFunction("close", ellClose, typeNull, typeChannel)

	if midi {
		initMidi()
	}

	err := loadModule("ell")
	if err != nil {
		fatal("*** ", err)
	}
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
	return readAll(argv[0], argv[1])
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
	return validateKeywordArgList(argv[0], argv[1:])
}

func ellKeys(argv []*LOB) (*LOB, error) {
	return structKeyList(argv[0]), nil
}

func ellValues(argv []*LOB) (*LOB, error) {
	return structValueList(argv[0]), nil
}

func ellStruct(argv []*LOB) (*LOB, error) {
	return newStruct(argv)
}
func ellMakeStruct(argv []*LOB) (*LOB, error) {
	return makeStruct(int(argv[0].fval)), nil
}

func ellToStruct(argv []*LOB) (*LOB, error) {
	//how about a keys: keyword argument to force a key type, like read does?
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
	return newString(writeIndent(argv[0], argv[1].text)), nil
}

func ellMakeError(argv []*LOB) (*LOB, error) {
	return newError(argv...), nil
}

func ellErrorP(argv []*LOB) (*LOB, error) {
	if isError(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellUncaughtError(argv []*LOB) (*LOB, error) {
	return nil, argv[0]
}

func ellToString(argv []*LOB) (*LOB, error) {
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
	result := EmptyList
	tail := result
	for _, lst := range argv {
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
	return reverse(argv[0]), nil
}

func ellFlatten(argv []*LOB) (*LOB, error) {
	return flatten(argv[0]), nil
}

func ellList(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func ellListLength(argv []*LOB) (*LOB, error) {
	return newInt(listLength(argv[0])), nil
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
	return toVector(argv[0])
}

func ellMakeVector(argv []*LOB) (*LOB, error) {
	vlen := int(argv[0].fval)
	init := argv[1]
	return newVector(vlen, init), nil
}

func ellVectorP(argv []*LOB) (*LOB, error) {
	if isVector(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellVectorLength(argv []*LOB) (*LOB, error) {
	return newInt(len(argv[0].elements)), nil
}

func ellVectorRef(argv []*LOB) (*LOB, error) {
	el := argv[0].elements
	idx := int(argv[1].fval)
	if idx < 0 || idx > len(el) {
		return nil, Error(ArgumentErrorKey, "Vector index out of range")
	}
	return el[idx], nil
}

func ellVectorSetBang(argv []*LOB) (*LOB, error) {
	if argv[0].ival != 0 {
		return nil, Error(ArgumentErrorKey, "vector-set! on sealed vector")
	}
	el := argv[0].elements
	idx := int(argv[1].fval)
	if idx < 0 || idx > len(el) {
		return nil, Error(ArgumentErrorKey, "Vector index out of range")
	}
	el[idx] = argv[2]
	return Null, nil
}

func ellZeroP(argv []*LOB) (*LOB, error) {
	if numberEqual(argv[0].fval, 0.0) {
		return True, nil
	}
	return False, nil
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
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	return symbol(argv)
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
	if argv[0] == EmptyList {
		return True, nil
	}
	return False, nil
}

func ellString(argv []*LOB) (*LOB, error) {
	s := ""
	for _, ss := range argv {
		s += ss.String()
	}
	return newString(s), nil
}

func ellStringLength(argv []*LOB) (*LOB, error) {
	return newInt(stringLength(argv[0].text)), nil
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
	if argv[0].ival != 0 {
		return nil, Error(ArgumentErrorKey, "set-car! on sealed list")
	}
	err := setCar(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellSetCdrBang(argv []*LOB) (*LOB, error) {
	if argv[0].ival != 0 {
		return nil, Error(ArgumentErrorKey, "set-cdr! on sealed list")
	}
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
	return structGet(argv[0], argv[1]), nil
}

func ellStructLength(argv []*LOB) (*LOB, error) {
	return newInt(structLength(argv[0])), nil
}

func ellHasP(argv []*LOB) (*LOB, error) {
	b, err := has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return True, nil
	}
	return False, nil
}

func ellSeal(argv []*LOB) (*LOB, error) {
	switch argv[0].variant {
	case typeStruct, typeVector, typeList:
		argv[0].ival = 1
		return argv[0], nil
	default:
		return nil, Error(ArgumentErrorKey, "cannot seal! ", argv[0])
	}
}

func ellPutBang(argv []*LOB) (*LOB, error) {
	key := argv[1]
	if !isValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	if argv[0].ival != 0 {
		return nil, Error(ArgumentErrorKey, "put! on sealed struct")
	}
	put(argv[0], key, argv[2])
	return Null, nil
}

func ellUnputBang(argv []*LOB) (*LOB, error) {
	key := argv[1]
	if !isValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	if argv[0].ival != 0 {
		return nil, Error(ArgumentErrorKey, "unput! on sealed struct")
	}
	unput(argv[0], key)
	return Null, nil
}

func ellToList(argv []*LOB) (*LOB, error) {
	return toList(argv[0])
}

func ellSplit(argv []*LOB) (*LOB, error) {
	return stringSplit(argv[0], argv[1])
}

func ellJoin(argv []*LOB) (*LOB, error) {
	return stringJoin(argv[0], argv[1])
}

func ellJSON(argv []*LOB) (*LOB, error) {
	s, err := writeToString(argv[0], true, argv[1].text)
	if err != nil {
		return nil, err
	}
	return newString(s), nil
}

func ellGetFn(argv []*LOB) (*LOB, error) {
	if len(argv) < 1 {
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

func ellChannel(argv []*LOB) (*LOB, error) {
	name := argv[0].text
	bufsize := int(argv[1].fval)
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
	ch := argv[0]
	if ch.channel != nil { //not closed
		val := argv[1]
		timeout := argv[2].fval        //FIX: timeouts in seconds, floating point
		if numberEqual(timeout, 0.0) { //non-blocking
			select {
			case ch.channel <- val:
				return True, nil
			default:
			}
		} else if timeout > 0 { //with timeout
			dur := time.Millisecond * time.Duration(timeout)
			select {
			case ch.channel <- val:
				return True, nil
			case <-time.After(dur):
			}
		} else { //block forever
			ch.channel <- val
			return True, nil
		}
	}
	return False, nil
}

func ellReceive(argv []*LOB) (*LOB, error) {
	ch := argv[0]
	if ch.channel != nil { //not closed
		timeout := argv[1].fval
		if numberEqual(timeout, 0.0) { //non-blocking
			select {
			case val, ok := <-ch.channel:
				if ok {
					return val, nil
				}
			default:
			}
		} else if timeout > 0 { //with timeout
			dur := time.Millisecond * time.Duration(timeout)
			select {
			case val, ok := <-ch.channel:
				if ok {
					return val, nil
				}
			case <-time.After(dur):
			}
		} else { //block forever
			return <-ch.channel, nil
		}
	}
	return Null, nil
}
