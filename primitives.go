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

package ell

// the primitive functions for the languages
import (
	"bytes"
	"fmt"
	"github.com/pborman/uuid"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strings"
	"time"
)

type Extension interface {
	Init() error
	Cleanup()
}

var extension Extension

// InitEnvironment - defines the global functions/variables/macros for the top level environment
func Init(ext Extension) {
	extension = ext

	DefineMacro("let", ellLet)
	DefineMacro("letrec", ellLetrec)
	DefineMacro("cond", ellCond)
	DefineMacro("quasiquote", ellQuasiquote)

	DefineGlobal("null", Null)
	DefineGlobal("true", True)
	DefineGlobal("false", False)

	DefineGlobal("apply", Apply)
	DefineGlobal("callcc", CallCC)
	DefineGlobal("spawn", Spawn)

	DefineFunction("version", ellVersion, StringType)
	DefineFunction("boolean?", ellBooleanP, BooleanType, AnyType)
	DefineFunction("not", ellNot, BooleanType, AnyType)
	DefineFunction("equal?", ellEqualP, BooleanType, AnyType, AnyType)
	DefineFunction("identical?", ellIdenticalP, BooleanType, AnyType, AnyType)
	DefineFunction("null?", ellNullP, BooleanType, AnyType)
	DefineFunction("def?", ellDefinedP, BooleanType, SymbolType)

	DefineFunction("type", ellType, TypeType, AnyType)
	DefineFunction("value", ellValue, AnyType, AnyType)
	DefineFunction("instance", ellInstance, AnyType, TypeType, AnyType)

	DefineFunction("type?", ellTypeP, BooleanType, AnyType)
	DefineFunction("type-name", ellTypeName, SymbolType, TypeType)
	DefineFunction("keyword?", ellKeywordP, BooleanType, AnyType)
	DefineFunction("keyword-name", ellKeywordName, SymbolType, KeywordType)
	DefineFunction("symbol?", ellSymbolP, BooleanType, AnyType)
	DefineFunctionRestArgs("symbol", ellSymbol, SymbolType, AnyType, AnyType) //"(<any> <any>*) <symbol>")

	DefineFunctionRestArgs("string?", ellStringP, BooleanType, AnyType)
	DefineFunctionRestArgs("string", ellString, StringType, AnyType) //"(<any>*) <string>")
	DefineFunction("to-string", ellToString, StringType, AnyType)
	DefineFunction("string-length", ellStringLength, NumberType, StringType)
	DefineFunction("split", ellSplit, ListType, StringType, StringType)
	DefineFunction("join", ellJoin, ListType, ListType, StringType) // <list|vector> for both arg1 and result could work
	DefineFunction("character?", ellCharacterP, BooleanType, AnyType)
	DefineFunction("to-character", ellToCharacter, CharacterType, AnyType)

	DefineFunction("blob?", ellBlobP, BooleanType, AnyType)
	DefineFunction("to-blob", ellToBlob, BlobType, AnyType)
	DefineFunction("make-blob", ellMakeBlob, BlobType, NumberType)
	DefineFunction("blob-length", ellBlobLength, NumberType, BlobType)
	DefineFunction("blob-ref", ellBlobRef, NumberType, BlobType, NumberType)

	DefineFunction("number?", ellNumberP, BooleanType, AnyType)
	DefineFunction("int?", ellIntP, BooleanType, AnyType)
	DefineFunction("float?", ellFloatP, BooleanType, AnyType)
	DefineFunction("to-number", ellToNumber, NumberType, AnyType)
	DefineFunction("int", ellInt, NumberType, AnyType)
	DefineFunction("floor", ellFloor, NumberType, NumberType)
	DefineFunction("ceiling", ellCeiling, NumberType, NumberType)
	DefineFunction("inc", ellInc, NumberType, NumberType)
	DefineFunction("dec", ellDec, NumberType, NumberType)
	DefineFunction("+", ellAdd, NumberType, NumberType, NumberType)
	DefineFunction("-", ellSub, NumberType, NumberType, NumberType)
	DefineFunction("*", ellMul, NumberType, NumberType, NumberType)
	DefineFunction("/", ellDiv, NumberType, NumberType, NumberType)
	DefineFunction("quotient", ellQuotient, NumberType, NumberType, NumberType)
	DefineFunction("remainder", ellRemainder, NumberType, NumberType, NumberType)
	DefineFunction("modulo", ellRemainder, NumberType, NumberType, NumberType) //fix
	DefineFunction("=", ellNumEqual, BooleanType, NumberType, NumberType)
	DefineFunction("<=", ellNumLessEqual, BooleanType, NumberType, NumberType)
	DefineFunction(">=", ellNumGreaterEqual, BooleanType, NumberType, NumberType)
	DefineFunction(">", ellNumGreater, BooleanType, NumberType, NumberType)
	DefineFunction("<", ellNumLess, BooleanType, NumberType, NumberType)
	DefineFunction("zero?", ellZeroP, BooleanType, NumberType)
	DefineFunction("abs", ellAbs, NumberType, NumberType)
	DefineFunction("exp", ellExp, NumberType, NumberType)
	DefineFunction("log", ellLog, NumberType, NumberType)
	DefineFunction("sin", ellSin, NumberType, NumberType)
	DefineFunction("cos", ellCos, NumberType, NumberType)
	DefineFunction("tan", ellTan, NumberType, NumberType)
	DefineFunction("asin", ellAsin, NumberType, NumberType)
	DefineFunction("acos", ellAcos, NumberType, NumberType)
	DefineFunction("atan", ellAtan, NumberType, NumberType)
	DefineFunction("atan2", ellAtan2, NumberType, NumberType, NumberType)

	DefineFunction("seal!", ellSeal, AnyType, AnyType) //actually only list, vector, and struct for now

	DefineFunction("list?", ellListP, BooleanType, AnyType)
	DefineFunction("empty?", ellEmptyP, BooleanType, ListType)
	DefineFunction("to-list", ellToList, ListType, AnyType)
	DefineFunction("cons", ellCons, ListType, AnyType, ListType)
	DefineFunction("car", ellCar, AnyType, ListType)
	DefineFunction("cdr", ellCdr, ListType, ListType)
	DefineFunction("set-car!", ellSetCarBang, NullType, ListType, AnyType)
	DefineFunction("set-cdr!", ellSetCdrBang, NullType, ListType, ListType)
	DefineFunction("list-length", ellListLength, NumberType, ListType)
	DefineFunction("reverse", ellReverse, ListType, ListType)
	DefineFunctionRestArgs("list", ellList, ListType, AnyType)
	DefineFunctionRestArgs("concat", ellConcat, ListType, ListType)
	DefineFunctionRestArgs("flatten", ellFlatten, ListType, ListType)

	DefineFunction("vector?", ellVectorP, BooleanType, AnyType)
	DefineFunction("to-vector", ellToVector, VectorType, AnyType)
	DefineFunctionRestArgs("vector", ellVector, VectorType, AnyType)
	DefineFunctionOptionalArgs("make-vector", ellMakeVector, VectorType, []*LOB{NumberType, AnyType}, Null)
	DefineFunction("vector-length", ellVectorLength, NumberType, VectorType)
	DefineFunction("vector-ref", ellVectorRef, AnyType, VectorType, NumberType)
	DefineFunction("vector-set!", ellVectorSetBang, NullType, VectorType, NumberType, AnyType)

	DefineFunction("struct?", ellStructP, BooleanType, AnyType)
	DefineFunction("to-struct", ellToStruct, StructType, AnyType)
	DefineFunctionRestArgs("struct", ellStruct, StructType, AnyType)
	DefineFunction("make-struct", ellMakeStruct, StructType, NumberType)
	DefineFunction("struct-length", ellStructLength, NumberType, StructType)
	DefineFunction("has?", ellHasP, BooleanType, StructType, AnyType) // key is <symbol|keyword|type|string>
	DefineFunction("get", ellGet, AnyType, StructType, AnyType)
	DefineFunction("put!", ellPutBang, NullType, StructType, AnyType, AnyType)
	DefineFunction("unput!", ellUnputBang, NullType, StructType, AnyType)
	DefineFunction("keys", ellKeys, ListType, AnyType)     // <struct|instance>
	DefineFunction("values", ellValues, ListType, AnyType) // <struct|instance>

	DefineFunction("function?", ellFunctionP, BooleanType, AnyType)
	DefineFunction("function-signature", ellFunctionSignature, StringType, FunctionType)
	DefineFunctionRestArgs("validate-keyword-arg-list", ellValidateKeywordArgList, ListType, KeywordType, ListType)
	DefineFunction("slurp", ellSlurp, StringType, StringType)
	DefineFunctionKeyArgs("read", ellRead, AnyType, []*LOB{StringType, TypeType}, []*LOB{AnyType}, []*LOB{Intern("keys:")})
	DefineFunctionKeyArgs("read-all", ellReadAll, AnyType, []*LOB{StringType, TypeType}, []*LOB{AnyType}, []*LOB{Intern("keys:")})
	DefineFunction("spit", ellSpit, NullType, StringType, StringType)
	DefineFunctionKeyArgs("write", ellWrite, NullType, []*LOB{AnyType, StringType}, []*LOB{EmptyString}, []*LOB{Intern("indent:")})
	DefineFunctionKeyArgs("write-all", ellWriteAll, NullType, []*LOB{AnyType, StringType}, []*LOB{EmptyString}, []*LOB{Intern("indent:")})
	DefineFunctionRestArgs("print", ellPrint, NullType, AnyType)
	DefineFunctionRestArgs("println", ellPrintln, NullType, AnyType)
	DefineFunction("macroexpand", ellMacroexpand, AnyType, AnyType)
	DefineFunction("compile", ellCompile, CodeType, AnyType)

	DefineFunctionRestArgs("make-error", ellMakeError, ErrorType, AnyType)
	DefineFunction("error?", ellErrorP, BooleanType, AnyType)
	DefineFunction("uncaught-error", ellUncaughtError, NullType, ErrorType) //doesn't return

	DefineFunctionKeyArgs("json", ellJSON, StringType, []*LOB{AnyType, StringType}, []*LOB{EmptyString}, []*LOB{Intern("indent:")})

	DefineFunctionRestArgs("getfn", ellGetFn, FunctionType, AnyType, SymbolType)
	DefineFunction("method-signature", ellMethodSignature, TypeType, ListType)

	DefineFunction("now", ellNow, NumberType)
	DefineFunction("since", ellSince, NumberType, NumberType)
	DefineFunction("sleep", ellSleep, NumberType, NumberType)

	DefineFunctionKeyArgs("channel", ellChannel, ChannelType, []*LOB{StringType, NumberType}, []*LOB{EmptyString, Zero}, []*LOB{Intern("name:"), Intern("bufsize:")})
	DefineFunctionOptionalArgs("send", ellSend, NullType, []*LOB{ChannelType, AnyType, NumberType}, MinusOne)
	DefineFunctionOptionalArgs("recv", ellReceive, AnyType, []*LOB{ChannelType, NumberType}, MinusOne)
	DefineFunction("close", ellClose, NullType, AnyType)

	DefineFunction("set-random-seed!", ellSetRandomSeedBang, NullType, NumberType)
	DefineFunctionRestArgs("random", ellRandom, NumberType, NumberType)
	DefineFunctionRestArgs("random-list", ellRandomList, ListType, NumberType)

	DefineFunctionRestArgs("uuid", ellUUIDFromTime, StringType, StringType)
	DefineFunction("timestamp", ellTimestamp, StringType)

	DefineFunction("listen", ellListen, ChannelType, NumberType)
	DefineFunction("connect", ellConnect, AnyType, StringType, NumberType)

	DefineFunction("serve", ellHTTPServer, AnyType, NumberType, FunctionType)
	DefineFunctionKeyArgs("http", ellHTTPClient, StructType,
		[]*LOB{StringType, StringType, StructType, BlobType}, //(http "url" method: "PUT" headers: {} body: #[blob])
		[]*LOB{String("GET"), EmptyStruct, EmptyBlob},
		[]*LOB{Intern("method:"), Intern("headers:"), Intern("body:")})

	err := Load("ell")
	if err != nil {
		Fatal("*** ", err)
	}
	if extension != nil {
		err := extension.Init()
		if err != nil {
			Fatal("*** ", err)
		}
	}
}

func Cleanup() {
	if extension != nil {
		extension.Cleanup()
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
	return String(Version), nil
}

func ellDefinedP(argv []*LOB) (*LOB, error) {
	if IsDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSlurp(argv []*LOB) (*LOB, error) {
	url := argv[0].text
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		res, err := httpClientOperation("GET", url, nil, nil)
		if err != nil {
			return nil, err
		}
		status := int(structGet(res, Intern("status:")).fval)
		if status != 200 {
			return nil, Error(HTTPErrorKey, status, " ", http.StatusText(status))
		}
		s, _ := ToString(structGet(res, Intern("body:")))
		return s, nil
	}
	return SlurpFile(argv[0].text)
}

func ellSpit(argv []*LOB) (*LOB, error) {
	url := argv[0].text
	data := argv[1].text
	err := SpitFile(url, data)
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellRead(argv []*LOB) (*LOB, error) {
	return Read(argv[0], argv[1])
}

func ellReadAll(argv []*LOB) (*LOB, error) {
	return ReadAll(argv[0], argv[1])
}

func ellMacroexpand(argv []*LOB) (*LOB, error) {
	return Macroexpand(argv[0])
}

func ellCompile(argv []*LOB) (*LOB, error) {
	expanded, err := Macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return Compile(expanded)
}

func ellType(argv []*LOB) (*LOB, error) {
	return argv[0].Type, nil
}

func ellValue(argv []*LOB) (*LOB, error) {
	return Value(argv[0]), nil
}

func ellInstance(argv []*LOB) (*LOB, error) {
	return Instance(argv[0], argv[1])
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
	return Struct(argv)
}
func ellMakeStruct(argv []*LOB) (*LOB, error) {
	return MakeStruct(int(argv[0].fval)), nil
}

func ellToStruct(argv []*LOB) (*LOB, error) {
	//how about a keys: keyword argument to force a key type, like read does?
	return ToStruct(argv[0])
}

func ellIdenticalP(argv []*LOB) (*LOB, error) {
	if argv[0] == argv[1] {
		return True, nil
	}
	return False, nil
}

func ellEqualP(argv []*LOB) (*LOB, error) {
	if Equal(argv[0], argv[1]) {
		return True, nil
	}
	return False, nil
}

func ellNumEqual(argv []*LOB) (*LOB, error) {
	if NumberEqual(argv[0].fval, argv[1].fval) {
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
	return String(writeIndent(argv[0], argv[1].text)), nil
}

func ellWriteAll(argv []*LOB) (*LOB, error) {
	return String(writeAllIndent(argv[0], argv[1].text)), nil
}

func ellMakeError(argv []*LOB) (*LOB, error) {
	return MakeError(argv...), nil
}

func ellErrorP(argv []*LOB) (*LOB, error) {
	if IsError(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellUncaughtError(argv []*LOB) (*LOB, error) {
	return nil, argv[0]
}

func ellToString(argv []*LOB) (*LOB, error) {
	return ToString(argv[0])
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
				result = List(lst.car)
				tail = result
			} else {
				tail.cdr = List(lst.car)
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
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ellListLength(argv []*LOB) (*LOB, error) {
	return Number(float64(ListLength(argv[0]))), nil
}

func ellNumberP(argv []*LOB) (*LOB, error) {
	if IsNumber(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellToNumber(argv []*LOB) (*LOB, error) {
	return ToNumber(argv[0])
}

func ellIntP(argv []*LOB) (*LOB, error) {
	if IsInt(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFloatP(argv []*LOB) (*LOB, error) {
	if IsFloat(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellInt(argv []*LOB) (*LOB, error) {
	return ToInt(argv[0])
}

func ellFloor(argv []*LOB) (*LOB, error) {
	return Number(math.Floor(argv[0].fval)), nil
}

func ellCeiling(argv []*LOB) (*LOB, error) {
	return Number(math.Ceil(argv[0].fval)), nil
}

func ellInc(argv []*LOB) (*LOB, error) {
	return Number(argv[0].fval + 1), nil
}

func ellDec(argv []*LOB) (*LOB, error) {
	return Number(argv[0].fval - 1), nil
}

func ellAdd(argv []*LOB) (*LOB, error) {
	return Number(argv[0].fval + argv[1].fval), nil
}

func ellSub(argv []*LOB) (*LOB, error) {
	return Number(argv[0].fval - argv[1].fval), nil
}

func ellMul(argv []*LOB) (*LOB, error) {
	return Number(argv[0].fval * argv[1].fval), nil
}

func ellDiv(argv []*LOB) (*LOB, error) {
	return Number(argv[0].fval / argv[1].fval), nil
}

func ellQuotient(argv []*LOB) (*LOB, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "quotient: divide by zero")
	}
	n := int64(argv[0].fval) / denom
	return Number(float64(n)), nil
}

func ellRemainder(argv []*LOB) (*LOB, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "remainder: divide by zero")
	}
	n := int64(argv[0].fval) % denom
	return Number(float64(n)), nil
}

func ellAbs(argv []*LOB) (*LOB, error) {
	return Number(math.Abs(argv[0].fval)), nil
}

func ellExp(argv []*LOB) (*LOB, error) {
	return Number(math.Exp(argv[0].fval)), nil
}

func ellLog(argv []*LOB) (*LOB, error) {
	return Number(math.Log(argv[0].fval)), nil
}

func ellSin(argv []*LOB) (*LOB, error) {
	return Number(math.Sin(argv[0].fval)), nil
}

func ellCos(argv []*LOB) (*LOB, error) {
	return Number(math.Cos(argv[0].fval)), nil
}

func ellTan(argv []*LOB) (*LOB, error) {
	return Number(math.Tan(argv[0].fval)), nil
}

func ellAsin(argv []*LOB) (*LOB, error) {
	return Number(math.Asin(argv[0].fval)), nil
}

func ellAcos(argv []*LOB) (*LOB, error) {
	return Number(math.Acos(argv[0].fval)), nil
}

func ellAtan(argv []*LOB) (*LOB, error) {
	return Number(math.Atan(argv[0].fval)), nil
}

func ellAtan2(argv []*LOB) (*LOB, error) {
	return Number(math.Atan2(argv[0].fval, argv[1].fval)), nil
}

func ellVector(argv []*LOB) (*LOB, error) {
	return Vector(argv...), nil
}

func ellToVector(argv []*LOB) (*LOB, error) {
	return ToVector(argv[0])
}

func ellMakeVector(argv []*LOB) (*LOB, error) {
	vlen := int(argv[0].fval)
	init := argv[1]
	return MakeVector(vlen, init), nil
}

func ellVectorP(argv []*LOB) (*LOB, error) {
	if IsVector(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellVectorLength(argv []*LOB) (*LOB, error) {
	return Number(float64(len(argv[0].elements))), nil
}

func ellVectorRef(argv []*LOB) (*LOB, error) {
	el := argv[0].elements
	idx := int(argv[1].fval)
	if idx < 0 || idx >= len(el) {
		return nil, Error(ArgumentErrorKey, "Vector index out of range")
	}
	return el[idx], nil
}

func ellVectorSetBang(argv []*LOB) (*LOB, error) {
	sealed := int(argv[0].fval)
	if sealed != 0 {
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
	if NumberEqual(argv[0].fval, 0.0) {
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
	if IsBoolean(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbolP(argv []*LOB) (*LOB, error) {
	if IsSymbol(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbol(argv []*LOB) (*LOB, error) {
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	return Symbol(argv)
}

func ellKeywordP(argv []*LOB) (*LOB, error) {
	if IsKeyword(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellKeywordName(argv []*LOB) (*LOB, error) {
	return KeywordName(argv[0])
}

func ellTypeP(argv []*LOB) (*LOB, error) {
	if IsType(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellTypeName(argv []*LOB) (*LOB, error) {
	return TypeName(argv[0])
}

func ellStringP(argv []*LOB) (*LOB, error) {
	if IsString(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellCharacterP(argv []*LOB) (*LOB, error) {
	if IsCharacter(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellToCharacter(argv []*LOB) (*LOB, error) {
	return ToCharacter(argv[0])
}

func ellFunctionP(argv []*LOB) (*LOB, error) {
	if IsFunction(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFunctionSignature(argv []*LOB) (*LOB, error) {
	return String(functionSignature(argv[0])), nil
}

func ellListP(argv []*LOB) (*LOB, error) {
	if IsList(argv[0]) {
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
	return String(s), nil
}

func ellStringLength(argv []*LOB) (*LOB, error) {
	return Number(float64(StringLength(argv[0].text))), nil
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
	lst := argv[0]
	if lst == EmptyList {
		return nil, Error(ArgumentErrorKey, "set-car! expected a non-empty <list>")
	}
	sealed := int(lst.fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "set-car! on sealed list")
	}
	lst.car = argv[1]
	return Null, nil
}

func ellSetCdrBang(argv []*LOB) (*LOB, error) {
	lst := argv[0]
	if lst == EmptyList {
		return nil, Error(ArgumentErrorKey, "set-cdr! expected a non-empty <list>")
	}
	sealed := int(lst.fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "set-cdr! on sealed list")
	}
	lst.cdr = argv[1]
	return Null, nil
}

func ellCons(argv []*LOB) (*LOB, error) {
	return Cons(argv[0], argv[1]), nil
}

func ellStructP(argv []*LOB) (*LOB, error) {
	if IsStruct(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellGet(argv []*LOB) (*LOB, error) {
	return structGet(argv[0], argv[1]), nil
}

func ellStructLength(argv []*LOB) (*LOB, error) {
	return Number(float64(StructLength(argv[0]))), nil
}

func ellHasP(argv []*LOB) (*LOB, error) {
	b, err := Has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return True, nil
	}
	return False, nil
}

func ellSeal(argv []*LOB) (*LOB, error) {
	switch argv[0].Type {
	case StructType, VectorType, ListType:
		argv[0].fval = 1
		return argv[0], nil
	default:
		return nil, Error(ArgumentErrorKey, "cannot seal! ", argv[0])
	}
}

func ellPutBang(argv []*LOB) (*LOB, error) {
	key := argv[1]
	if !IsValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	sealed := int(argv[0].fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "put! on sealed struct")
	}
	Put(argv[0], key, argv[2])
	return Null, nil
}

func ellUnputBang(argv []*LOB) (*LOB, error) {
	key := argv[1]
	if !IsValidStructKey(key) {
		return nil, Error(ArgumentErrorKey, "Bad struct key: ", key)
	}
	sealed := int(argv[0].fval)
	if sealed != 0 {
		return nil, Error(ArgumentErrorKey, "unput! on sealed struct")
	}
	Unput(argv[0], key)
	return Null, nil
}

func ellToList(argv []*LOB) (*LOB, error) {
	return ToList(argv[0])
}

func ellSplit(argv []*LOB) (*LOB, error) {
	return StringSplit(argv[0], argv[1])
}

func ellJoin(argv []*LOB) (*LOB, error) {
	return StringJoin(argv[0], argv[1])
}

func ellJSON(argv []*LOB) (*LOB, error) {
	s, err := writeToString(argv[0], true, argv[1].text)
	if err != nil {
		return nil, err
	}
	return String(s), nil
}

func ellGetFn(argv []*LOB) (*LOB, error) {
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "getfn expected at least 1 argument, got none")
	}
	sym := argv[0]
	if sym.Type != SymbolType {
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
	return Channel(bufsize, name), nil
}

func ellClose(argv []*LOB) (*LOB, error) {
	switch argv[0].Type {
	case ChannelType:
		CloseChannel(argv[0])
	case Intern("<tcp-connection>"):
		closeConnection(argv[0])
	default:
		return nil, Error(ArgumentErrorKey, "close expected a channel or connection")
	}
	return Null, nil
}

func ellSend(argv []*LOB) (*LOB, error) {
	ch := ChannelValue(argv[0])
	if ch != nil { //not closed
		val := argv[1]
		timeout := argv[2].fval        //FIX: timeouts in seconds, floating point
		if NumberEqual(timeout, 0.0) { //non-blocking
			select {
			case ch <- val:
				return True, nil
			default:
			}
		} else if timeout > 0 { //with timeout
			dur := time.Millisecond * time.Duration(timeout*1000.0)
			select {
			case ch <- val:
				return True, nil
			case <-time.After(dur):
			}
		} else { //block forever
			ch <- val
			return True, nil
		}
	}
	return False, nil
}

func ellReceive(argv []*LOB) (*LOB, error) {
	ch := ChannelValue(argv[0])
	if ch != nil { //not closed
		timeout := argv[1].fval
		if NumberEqual(timeout, 0.0) { //non-blocking
			select {
			case val, ok := <-ch:
				if ok && val != nil {
					return val, nil
				}
			default:
			}
		} else if timeout > 0 { //with timeout
			dur := time.Millisecond * time.Duration(timeout*1000.0)
			select {
			case val, ok := <-ch:
				if ok && val != nil {
					return val, nil
				}
			case <-time.After(dur):
			}
		} else { //block forever
			val := <-ch
			if val != nil {
				return val, nil
			}
		}
	}
	return Null, nil
}

func ellSetRandomSeedBang(argv []*LOB) (*LOB, error) {
	RandomSeed(int64(argv[0].fval))
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
	return Random(min, max), nil
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
	return RandomList(count, min, max), nil
}

func ellUUIDFromTime(argv []*LOB) (*LOB, error) {
	var u uuid.UUID
	argc := len(argv)
	switch argc {
	case 0:
		u = uuid.NewUUID()
	case 1:
		u = uuid.NewMD5(uuid.NameSpace_URL, []byte(argv[0].text))
	case 2:
		ns := uuid.Parse(argv[0].text)
		if ns == nil {
			ns = uuid.NewMD5(uuid.NameSpace_URL, []byte(argv[0].text))
		}
		u = uuid.NewMD5(ns, []byte(argv[1].text))
	}
	if u == nil {
		return nil, Error(ArgumentErrorKey, "Expected 0-2 arguments, got: ", argc)
	}
	return String(u.String()), nil
}

func ellTimestamp(argv []*LOB) (*LOB, error) {
	t := time.Now().UTC()
	format := "%d-%02d-%02dT%02d:%02d:%02d.%03dZ"
	return String(fmt.Sprintf(format, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000)), nil
}

func ellBlobP(argv []*LOB) (*LOB, error) {
	if argv[0].Type == BlobType {
		return True, nil
	}
	return False, nil
}

func ellToBlob(argv []*LOB) (*LOB, error) {
	return toBlob(argv[0])
}

func ellMakeBlob(argv []*LOB) (*LOB, error) {
	size := int(argv[0].fval)
	return MakeBlob(size), nil
}

func ellBlobLength(argv []*LOB) (*LOB, error) {
	return Number(float64(len(BlobValue(argv[0])))), nil
}

func ellBlobRef(argv []*LOB) (*LOB, error) {
	el := BlobValue(argv[0])
	idx := int(argv[1].fval)
	if idx < 0 || idx >= len(el) {
		return nil, Error(ArgumentErrorKey, "Blob index out of range")
	}
	return Number(float64(el[idx])), nil
}

func ellListen(argv []*LOB) (*LOB, error) {
	port := fmt.Sprintf(":%d", int(argv[0].fval))
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	acceptChan := Channel(10, fmt.Sprintf("tcp listener on %s", port))
	go tcpListener(listener, acceptChan, port)
	return acceptChan, nil
}

func ellConnect(argv []*LOB) (*LOB, error) {
	host := argv[0].text
	port := int(argv[1].fval)
	endpoint := fmt.Sprintf("%s:%d", host, port)
	con, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	return Connection(con, endpoint), nil
}

func ellHTTPServer(argv []*LOB) (*LOB, error) {
	port := int(argv[0].fval)
	handler := argv[1] // a function of one <struct> argument
	if handler.code == nil || handler.code.argc != 1 {
		return nil, Error(ArgumentErrorKey, "Cannot use this function as a handler: ", handler)
	}
	glue := func(w http.ResponseWriter, r *http.Request) {
		headers := MakeStruct(10)
		for k, v := range r.Header {
			var values []*LOB
			for _, val := range v {
				values = append(values, String(val))
			}
			Put(headers, String(k), listFromValues(values))
		}
		var body *LOB
		println("method: ", r.Method)
		switch strings.ToUpper(r.Method) {
		case "POST", "PUT":
			println("try to read the body...")
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Cannot decode body for " + r.Method + " request"))
				return
			}
			println("... got ", bodyBytes)
			body = Blob(bodyBytes)
		}
		req, _ := Struct([]*LOB{Intern("headers:"), headers, Intern("body:"), body})
		println("req object: ", req)
		args := []*LOB{req}
		res, err := exec(handler.code, args)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		if !IsStruct(res) {
			w.WriteHeader(500)
			w.Write([]byte("Handler did not return a struct"))
			return
		}
		headers = structGet(res, Intern("headers:"))
		body = structGet(res, Intern("body:"))
		status := structGet(res, Intern("status:"))
		if status != nil {
			nstatus := int(status.fval)
			if nstatus != 0 && nstatus != 200 {
				w.WriteHeader(nstatus)
			}
		}
		if IsStruct(headers) {
			//fix: multiple values for a header
			for k, v := range headers.bindings {
				ks := headerString(k.toLOB())
				vs := v.String()
				w.Header().Set(ks, vs)
			}
		}
		if IsString(body) {
			bodylen := len(body.text)
			w.Header().Set("Content-length", fmt.Sprint(bodylen))
			if bodylen > 0 {
				w.Write([]byte(body.text))
			}
		}
	}
	http.HandleFunc("/", glue)
	println("web server running at ", fmt.Sprintf(":%d", port))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	//no way to stop it
	return Null, nil
}

func headerString(obj *LOB) string {
	switch obj.Type {
	case StringType, SymbolType:
		return obj.text
	case KeywordType:
		return unkeywordedString(obj)
	default:
		s, err := ToString(obj)
		if err != nil {
			return typeNameString(obj.Type.text)
		}
		return s.text
	}
}

func ellHTTPGet(url string, headers *LOB) (*LOB, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if headers != nil {
		for k, v := range headers.bindings {
			ks := k.toLOB().text
			if v.Type == ListType {
				vs := v.car.String()
				req.Header.Set(ks, vs)
				for v.cdr != EmptyList {
					v = v.cdr
					req.Header.Add(ks, v.car.String())
				}
			} else {
				req.Header.Set(ks, v.String())
			}
		}
	}
	res, err := client.Do(req)
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err == nil {
			status := Number(float64(res.StatusCode))
			headers := MakeStruct(len(res.Header))
			body := Blob(bodyBytes)
			for k, v := range res.Header {
				var values []*LOB
				for _, val := range v {
					values = append(values, String(val))
				}
				Put(headers, String(k), listFromValues(values))
			}
			s, _ := Struct([]*LOB{Intern("status:"), status, Intern("headers:"), headers, Intern("body:"), body})
			return s, nil
		}
	}
	return nil, err
}

func httpClientOperation(method string, url string, headers *LOB, data *LOB) (*LOB, error) {
	client := &http.Client{}
	var bodyReader io.Reader
	bodyLen := 0
	if data != nil {
		tmp := []byte(data.text)
		bodyLen = len(tmp)
		if bodyLen > 0 {
			bodyReader = bytes.NewBuffer(tmp)
		}
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if headers != nil {
		for k, v := range headers.bindings {
			ks := k.toLOB().text
			if v.Type == ListType {
				vs := v.car.String()
				req.Header.Set(ks, vs)
				for v.cdr != EmptyList {
					v = v.cdr
					req.Header.Add(ks, v.car.String())
				}
			} else {
				req.Header.Set(ks, v.String())
			}
		}
	}
	if bodyLen > 0 {
		req.Header.Set("Content-Length", fmt.Sprint(bodyLen))
	}
	res, err := client.Do(req)
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err == nil {
			s := MakeStruct(3)
			Put(s, Intern("status:"), Number(float64(res.StatusCode)))
			bodyLen := len(bodyBytes)
			if bodyLen > 0 {
				Put(s, Intern("body:"), Blob(bodyBytes))
			}
			if len(res.Header) > 0 {
				headers = MakeStruct(len(res.Header))
				for k, v := range res.Header {
					var values []*LOB
					for _, val := range v {
						values = append(values, String(val))
					}
					Put(headers, String(k), listFromValues(values))
				}
				Put(s, Intern("headers:"), headers)
			}
			return s, nil
		}
	}
	return nil, err
}

func ellHTTPClient(argv []*LOB) (*LOB, error) {
	url := argv[0].text
	method := strings.ToUpper(argv[1].text)
	headers := argv[2]
	body := argv[3]
	println("http ", method, " of ", url, ", headers: ", headers, ", body: ", body)
	switch method {
	case "GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS", "PATCH":
		return httpClientOperation(method, url, headers, body)
	default:
		return nil, Error(ErrorKey, "HTTP method not support: ", method)
	}
}

func Now() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / float64(time.Second)
}

func now() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / float64(time.Second)
}

func ellNow(argv []*LOB) (*LOB, error) {
	return Number(now()), nil
}

func ellSince(argv []*LOB) (*LOB, error) {
	then := argv[0].fval
	dur := now() - then
	return Number(dur), nil
}

func ellSleep(argv []*LOB) (*LOB, error) {
	dur := time.Duration(argv[0].fval * float64(time.Second))
	time.Sleep(dur) //!! this is not interruptable, fairly risky in a REPL
	return Number(now()), nil
}
