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
	"math"
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

	defineFunction("version", ellVersion, StringType)
	defineFunction("boolean?", ellBooleanP, BooleanType, AnyType)
	defineFunction("not", ellNot, BooleanType, AnyType)
	defineFunction("equal?", ellEqualP, BooleanType, AnyType, AnyType)
	defineFunction("identical?", ellIdenticalP, BooleanType, AnyType, AnyType)
	defineFunction("null?", ellNullP, BooleanType, AnyType)
	defineFunction("def?", ellDefinedP, BooleanType, SymbolType)

	defineFunction("type", ellType, TypeType, AnyType)
	defineFunction("value", ellValue, AnyType, AnyType)
	defineFunction("instance", ellInstance, AnyType, TypeType, AnyType)

	defineFunction("type?", ellTypeP, BooleanType, AnyType)
	defineFunction("type-name", ellTypeName, SymbolType, TypeType)
	defineFunction("keyword?", ellKeywordP, BooleanType, AnyType)
	defineFunction("keyword-name", ellKeywordName, SymbolType, KeywordType)
	defineFunction("symbol?", ellSymbolP, BooleanType, AnyType)
	defineFunctionRestArgs("symbol", ellSymbol, SymbolType, AnyType, AnyType) //"(<any> <any>*) <symbol>")

	defineFunctionRestArgs("string?", ellStringP, BooleanType, AnyType)
	defineFunctionRestArgs("string", ellString, StringType, AnyType) //"(<any>*) <string>")
	defineFunction("to-string", ellToString, StringType, AnyType)
	defineFunction("string-length", ellStringLength, NumberType, StringType)
	defineFunction("split", ellSplit, ListType, StringType, StringType)
	defineFunction("join", ellJoin, ListType, ListType, StringType) // <list|vector> for both arg1 and result could work
	defineFunction("character?", ellCharacterP, BooleanType, AnyType)
	defineFunction("to-character", ellToCharacter, CharacterType, AnyType)

	defineFunction("number?", ellNumberP, BooleanType, AnyType)
	defineFunction("int?", ellIntP, BooleanType, AnyType)
	defineFunction("float?", ellFloatP, BooleanType, AnyType)
	defineFunction("int", ellInt, NumberType, AnyType)
	defineFunction("floor", ellFloor, NumberType, NumberType)
	defineFunction("ceil", ellCeil, NumberType, NumberType)
	defineFunction("inc", ellInc, NumberType, NumberType)
	defineFunction("dec", ellDec, NumberType, NumberType)
	defineFunction("+", ellAdd, NumberType, NumberType, NumberType)
	defineFunction("-", ellSub, NumberType, NumberType, NumberType)
	defineFunction("*", ellMul, NumberType, NumberType, NumberType)
	defineFunction("/", ellDiv, NumberType, NumberType, NumberType)
	defineFunction("quotient", ellQuotient, NumberType, NumberType, NumberType)
	defineFunction("remainder", ellRemainder, NumberType, NumberType, NumberType)
	defineFunction("modulo", ellRemainder, NumberType, NumberType, NumberType) //fix
	defineFunction("=", ellNumEqual, BooleanType, NumberType, NumberType)
	defineFunction("<=", ellNumLessEqual, BooleanType, NumberType, NumberType)
	defineFunction(">=", ellNumGreaterEqual, BooleanType, NumberType, NumberType)
	defineFunction(">", ellNumGreater, BooleanType, NumberType, NumberType)
	defineFunction("<", ellNumLess, BooleanType, NumberType, NumberType)
	defineFunction("zero?", ellZeroP, BooleanType, NumberType)

	defineFunction("seal!", ellSeal, AnyType, AnyType) //actually only list, vector, and struct for now

	defineFunction("list?", ellListP, BooleanType, AnyType)
	defineFunction("empty?", ellEmptyP, BooleanType, ListType)
	defineFunction("to-list", ellToList, ListType, AnyType)
	defineFunction("cons", ellCons, ListType, AnyType, ListType)
	defineFunction("car", ellCar, AnyType, ListType)
	defineFunction("cdr", ellCdr, ListType, ListType)
	defineFunction("set-car!", ellSetCarBang, NullType, ListType, AnyType)
	defineFunction("set-cdr!", ellSetCdrBang, NullType, ListType, ListType)
	defineFunction("list-length", ellListLength, NumberType, ListType)
	defineFunction("reverse", ellReverse, ListType, ListType)
	defineFunctionRestArgs("list", ellList, ListType, AnyType)
	defineFunctionRestArgs("concat", ellConcat, ListType, ListType)
	defineFunctionRestArgs("flatten", ellFlatten, ListType, ListType)

	defineFunction("vector?", ellVectorP, BooleanType, AnyType)
	defineFunction("to-vector", ellToVector, VectorType, AnyType)
	defineFunctionRestArgs("vector", ellVector, VectorType, AnyType)
	defineFunctionOptionalArgs("make-vector", ellMakeVector, VectorType, []*LOB{NumberType, AnyType}, Null)
	defineFunction("vector-length", ellVectorLength, NumberType, VectorType)
	defineFunction("vector-ref", ellVectorRef, AnyType, VectorType, NumberType)
	defineFunction("vector-set!", ellVectorSetBang, NullType, VectorType, NumberType, AnyType)

	defineFunction("struct?", ellStructP, BooleanType, AnyType)
	defineFunction("to-struct", ellToStruct, StructType, AnyType)
	defineFunctionRestArgs("struct", ellStruct, StructType, AnyType)
	defineFunction("make-struct", ellMakeStruct, StructType, NumberType)
	defineFunction("struct-length", ellStructLength, NumberType, StructType)
	defineFunction("has?", ellHasP, BooleanType, StructType, AnyType) // key is <symbol|keyword|type|string>
	defineFunction("get", ellGet, AnyType, StructType, AnyType)
	defineFunction("put!", ellPutBang, NullType, StructType, AnyType, AnyType)
	defineFunction("unput!", ellUnputBang, NullType, StructType, AnyType)
	defineFunction("keys", ellKeys, ListType, AnyType)     // <struct|instance>
	defineFunction("values", ellValues, ListType, AnyType) // <struct|instance>

	defineFunction("function?", ellFunctionP, BooleanType, AnyType)
	defineFunction("function-signature", ellFunctionSignature, StringType, FunctionType)
	defineFunctionRestArgs("validate-keyword-arg-list", ellValidateKeywordArgList, ListType, KeywordType, ListType)
	defineFunction("slurp", ellSlurp, StringType, StringType)
	defineFunctionKeyArgs("read", ellRead, AnyType, []*LOB{StringType, TypeType}, []*LOB{AnyType}, []*LOB{intern("keys:")})
	defineFunction("spit", ellSpit, NullType, StringType, AnyType)
	defineFunctionKeyArgs("write", ellWrite, NullType, []*LOB{AnyType, StringType}, []*LOB{EmptyString}, []*LOB{intern("indent:")})
	defineFunctionRestArgs("print", ellPrint, NullType, AnyType)
	defineFunctionRestArgs("println", ellPrintln, NullType, AnyType)
	defineFunction("macroexpand", ellMacroexpand, AnyType, AnyType)
	defineFunction("compile", ellCompile, CodeType, AnyType)

	defineFunctionRestArgs("make-error", ellMakeError, ErrorType, AnyType)
	defineFunction("error?", ellErrorP, BooleanType, AnyType)
	defineFunction("uncaught-error", ellUncaughtError, NullType, ErrorType) //doesn't return

	defineFunctionKeyArgs("json", ellJSON, StringType, []*LOB{AnyType, StringType}, []*LOB{EmptyString}, []*LOB{intern("indent:")})

	defineFunctionRestArgs("getfn", ellGetFn, FunctionType, AnyType, SymbolType)
	defineFunction("method-signature", ellMethodSignature, TypeType, ListType)

	defineFunctionKeyArgs("channel", ellChannel, ChannelType, []*LOB{StringType, NumberType}, []*LOB{EmptyString, Zero}, []*LOB{intern("name:"), intern("bufsize:")})
	defineFunctionOptionalArgs("send", ellSend, NullType, []*LOB{ChannelType, AnyType, NumberType}, MinusOne)
	defineFunctionOptionalArgs("recv", ellReceive, AnyType, []*LOB{ChannelType, NumberType}, MinusOne)
	defineFunction("close", ellClose, NullType, ChannelType)

	defineFunction("set-random-seed!", ellSetRandomSeedBang, NullType, NumberType)
	defineFunctionRestArgs("random", ellRandom, NumberType, NumberType)
	defineFunctionRestArgs("random-list", ellRandomList, ListType, NumberType)

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

func ellInt(argv []*LOB) (*LOB, error) {
	return toInt(argv[0])
}

func ellFloor(argv []*LOB) (*LOB, error) {
	return newFloat64(math.Floor(argv[0].fval)), nil
}

func ellCeil(argv []*LOB) (*LOB, error) {
	return newFloat64(math.Ceil(argv[0].fval)), nil
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
	case StructType, VectorType, ListType:
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
	if sym.variant != SymbolType {
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


func ellSetRandomSeedBang(argv []*LOB) (*LOB, error) {
	randomSeed(int64(argv[0].fval))
	return Null, nil
}

func ellRandom(argv []*LOB) (*LOB, error) {
	min := 0.0
	max := 1.0
	argc := len(argv)
	switch argc {
	case 0:
	case 1:
		max = argv[0].fval
	case 2:
		min = argv[0].fval
		max = argv[1].fval
	default:
      return nil, Error(ArgumentErrorKey, "random expected 0 to 2 arguments, got ", argc)
	}
	return random(min, max), nil
}

func ellRandomList(argv []*LOB) (*LOB, error) {
	count := int(argv[0].fval)
	min := 0.0
	max := 1.0
	argc := len(argv)
	switch argc {
	case 1:
	case 2:
		max = argv[1].fval
	case 3:
		min = argv[1].fval
		max = argv[2].fval
	default:
      return nil, Error(ArgumentErrorKey, "random-list expected 1 to 3 arguments, got ", argc)
	}
	return randomList(count, min, max), nil
}


