/*
Copyright 2015 Lee Boynton

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
	"fmt"
	"github.com/pborman/uuid"
	"math"
	"net"
	"net/http"
	"strings"
	"time"
)

// InitEnvironment - defines the global functions/variables/macros for the top level environment
func InitPrimitives() {
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
	DefineFunctionOptionalArgs("make-vector", ellMakeVector, VectorType, []*Object{NumberType, AnyType}, Null)
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
	DefineFunctionKeyArgs("read", ellRead, AnyType, []*Object{StringType, TypeType}, []*Object{AnyType}, []*Object{Intern("keys:")})
	DefineFunctionKeyArgs("read-all", ellReadAll, AnyType, []*Object{StringType, TypeType}, []*Object{AnyType}, []*Object{Intern("keys:")})
	DefineFunction("spit", ellSpit, NullType, StringType, StringType)
	DefineFunctionKeyArgs("write", ellWrite, NullType, []*Object{AnyType, StringType}, []*Object{EmptyString}, []*Object{Intern("indent:")})
	DefineFunctionKeyArgs("write-all", ellWriteAll, NullType, []*Object{AnyType, StringType}, []*Object{EmptyString}, []*Object{Intern("indent:")})
	DefineFunctionRestArgs("print", ellPrint, NullType, AnyType)
	DefineFunctionRestArgs("println", ellPrintln, NullType, AnyType)
	DefineFunction("macroexpand", ellMacroexpand, AnyType, AnyType)
	DefineFunction("compile", ellCompile, CodeType, AnyType)

	DefineFunctionRestArgs("make-error", ellMakeError, ErrorType, AnyType)
	DefineFunction("error?", ellErrorP, BooleanType, AnyType)
	DefineFunction("uncaught-error", ellUncaughtError, NullType, ErrorType) //doesn't return

	DefineFunctionKeyArgs("json", ellJSON, StringType, []*Object{AnyType, StringType}, []*Object{EmptyString}, []*Object{Intern("indent:")})

	DefineFunctionRestArgs("getfn", ellGetFn, FunctionType, AnyType, SymbolType)
	DefineFunction("method-signature", ellMethodSignature, TypeType, ListType)

	DefineFunction("now", ellNow, NumberType)
	DefineFunction("since", ellSince, NumberType, NumberType)
	DefineFunction("sleep", ellSleep, NumberType, NumberType)

	DefineFunctionKeyArgs("channel", ellChannel, ChannelType, []*Object{StringType, NumberType}, []*Object{EmptyString, Zero}, []*Object{Intern("name:"), Intern("bufsize:")})
	DefineFunctionOptionalArgs("send", ellSend, NullType, []*Object{ChannelType, AnyType, NumberType}, MinusOne)
	DefineFunctionOptionalArgs("recv", ellReceive, AnyType, []*Object{ChannelType, NumberType}, MinusOne)
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
		[]*Object{StringType, StringType, StructType, BlobType}, //(http "url" method: "PUT" headers: {} body: #[blob])
		[]*Object{String("GET"), EmptyStruct, EmptyBlob},
		[]*Object{Intern("method:"), Intern("headers:"), Intern("body:")})

	err := Load("ell")
	if err != nil {
		Fatal("*** ", err)
	}
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *LList
//

func ellLetrec(argv []*Object) (*Object, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []*Object) (*Object, error) {
	return expandLet(argv[0])
}

func ellCond(argv []*Object) (*Object, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []*Object) (*Object, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(argv []*Object) (*Object, error) {
	s := Version
	if len(extensions) > 0 {
		s += " (with "
		for i, ext := range extensions {
			if i > 0 {
				s += ", "
			}
			s += ext.String()
		}
		s += ")"
	}
	return String(s), nil
}

func ellDefinedP(argv []*Object) (*Object, error) {
	if IsDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSlurp(argv []*Object) (*Object, error) {
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

func ellSpit(argv []*Object) (*Object, error) {
	url := argv[0].text
	data := argv[1].text
	err := SpitFile(url, data)
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellRead(argv []*Object) (*Object, error) {
	return Read(argv[0], argv[1])
}

func ellReadAll(argv []*Object) (*Object, error) {
	return ReadAll(argv[0], argv[1])
}

func ellMacroexpand(argv []*Object) (*Object, error) {
	return Macroexpand(argv[0])
}

func ellCompile(argv []*Object) (*Object, error) {
	expanded, err := Macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return Compile(expanded)
}

func ellType(argv []*Object) (*Object, error) {
	return argv[0].Type, nil
}

func ellValue(argv []*Object) (*Object, error) {
	return Value(argv[0]), nil
}

func ellInstance(argv []*Object) (*Object, error) {
	return Instance(argv[0], argv[1])
}

func ellValidateKeywordArgList(argv []*Object) (*Object, error) {
	//(validate-keyword-arg-list '(x: 23) x: y:) -> (x:)
	//(validate-keyword-arg-list '(x: 23 z: 100) x: y:) -> error("bad keyword z: in argument list")
	return validateKeywordArgList(argv[0], argv[1:])
}

func ellKeys(argv []*Object) (*Object, error) {
	return structKeyList(argv[0]), nil
}

func ellValues(argv []*Object) (*Object, error) {
	return structValueList(argv[0]), nil
}

func ellStruct(argv []*Object) (*Object, error) {
	return Struct(argv)
}
func ellMakeStruct(argv []*Object) (*Object, error) {
	return MakeStruct(int(argv[0].fval)), nil
}

func ellToStruct(argv []*Object) (*Object, error) {
	//how about a keys: keyword argument to force a key type, like read does?
	return ToStruct(argv[0])
}

func ellIdenticalP(argv []*Object) (*Object, error) {
	if argv[0] == argv[1] {
		return True, nil
	}
	return False, nil
}

func ellEqualP(argv []*Object) (*Object, error) {
	if Equal(argv[0], argv[1]) {
		return True, nil
	}
	return False, nil
}

func ellNumEqual(argv []*Object) (*Object, error) {
	if NumberEqual(argv[0].fval, argv[1].fval) {
		return True, nil
	}
	return False, nil
}

func ellNumLess(argv []*Object) (*Object, error) {
	if argv[0].fval < argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellNumLessEqual(argv []*Object) (*Object, error) {
	if argv[0].fval <= argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellNumGreater(argv []*Object) (*Object, error) {
	if argv[0].fval > argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellNumGreaterEqual(argv []*Object) (*Object, error) {
	if argv[0].fval >= argv[1].fval {
		return True, nil
	}
	return False, nil
}

func ellWrite(argv []*Object) (*Object, error) {
	return String(writeIndent(argv[0], argv[1].text)), nil
}

func ellWriteAll(argv []*Object) (*Object, error) {
	return String(writeAllIndent(argv[0], argv[1].text)), nil
}

func ellMakeError(argv []*Object) (*Object, error) {
	return MakeError(argv...), nil
}

func ellErrorP(argv []*Object) (*Object, error) {
	if IsError(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellUncaughtError(argv []*Object) (*Object, error) {
	return nil, argv[0]
}

func ellToString(argv []*Object) (*Object, error) {
	return ToString(argv[0])
}

func ellPrint(argv []*Object) (*Object, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []*Object) (*Object, error) {
	ellPrint(argv)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []*Object) (*Object, error) {
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

func ellReverse(argv []*Object) (*Object, error) {
	return Reverse(argv[0]), nil
}

func ellFlatten(argv []*Object) (*Object, error) {
	return Flatten(argv[0]), nil
}

func ellList(argv []*Object) (*Object, error) {
	argc := len(argv)
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ellListLength(argv []*Object) (*Object, error) {
	return Number(float64(ListLength(argv[0]))), nil
}

func ellNumberP(argv []*Object) (*Object, error) {
	if IsNumber(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellToNumber(argv []*Object) (*Object, error) {
	return ToNumber(argv[0])
}

func ellIntP(argv []*Object) (*Object, error) {
	if IsInt(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFloatP(argv []*Object) (*Object, error) {
	if IsFloat(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellInt(argv []*Object) (*Object, error) {
	return ToInt(argv[0])
}

func ellFloor(argv []*Object) (*Object, error) {
	return Number(math.Floor(argv[0].fval)), nil
}

func ellCeiling(argv []*Object) (*Object, error) {
	return Number(math.Ceil(argv[0].fval)), nil
}

func ellInc(argv []*Object) (*Object, error) {
	return Number(argv[0].fval + 1), nil
}

func ellDec(argv []*Object) (*Object, error) {
	return Number(argv[0].fval - 1), nil
}

func ellAdd(argv []*Object) (*Object, error) {
	return Number(argv[0].fval + argv[1].fval), nil
}

func ellSub(argv []*Object) (*Object, error) {
	return Number(argv[0].fval - argv[1].fval), nil
}

func ellMul(argv []*Object) (*Object, error) {
	return Number(argv[0].fval * argv[1].fval), nil
}

func ellDiv(argv []*Object) (*Object, error) {
	return Number(argv[0].fval / argv[1].fval), nil
}

func ellQuotient(argv []*Object) (*Object, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "quotient: divide by zero")
	}
	n := int64(argv[0].fval) / denom
	return Number(float64(n)), nil
}

func ellRemainder(argv []*Object) (*Object, error) {
	denom := int64(argv[1].fval)
	if denom == 0 {
		return nil, Error(ArgumentErrorKey, "remainder: divide by zero")
	}
	n := int64(argv[0].fval) % denom
	return Number(float64(n)), nil
}

func ellAbs(argv []*Object) (*Object, error) {
	return Number(math.Abs(argv[0].fval)), nil
}

func ellExp(argv []*Object) (*Object, error) {
	return Number(math.Exp(argv[0].fval)), nil
}

func ellLog(argv []*Object) (*Object, error) {
	return Number(math.Log(argv[0].fval)), nil
}

func ellSin(argv []*Object) (*Object, error) {
	return Number(math.Sin(argv[0].fval)), nil
}

func ellCos(argv []*Object) (*Object, error) {
	return Number(math.Cos(argv[0].fval)), nil
}

func ellTan(argv []*Object) (*Object, error) {
	return Number(math.Tan(argv[0].fval)), nil
}

func ellAsin(argv []*Object) (*Object, error) {
	return Number(math.Asin(argv[0].fval)), nil
}

func ellAcos(argv []*Object) (*Object, error) {
	return Number(math.Acos(argv[0].fval)), nil
}

func ellAtan(argv []*Object) (*Object, error) {
	return Number(math.Atan(argv[0].fval)), nil
}

func ellAtan2(argv []*Object) (*Object, error) {
	return Number(math.Atan2(argv[0].fval, argv[1].fval)), nil
}

func ellVector(argv []*Object) (*Object, error) {
	return Vector(argv...), nil
}

func ellToVector(argv []*Object) (*Object, error) {
	return ToVector(argv[0])
}

func ellMakeVector(argv []*Object) (*Object, error) {
	vlen := int(argv[0].fval)
	init := argv[1]
	return MakeVector(vlen, init), nil
}

func ellVectorP(argv []*Object) (*Object, error) {
	if IsVector(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellVectorLength(argv []*Object) (*Object, error) {
	return Number(float64(len(argv[0].elements))), nil
}

func ellVectorRef(argv []*Object) (*Object, error) {
	el := argv[0].elements
	idx := int(argv[1].fval)
	if idx < 0 || idx >= len(el) {
		return nil, Error(ArgumentErrorKey, "Vector index out of range")
	}
	return el[idx], nil
}

func ellVectorSetBang(argv []*Object) (*Object, error) {
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

func ellZeroP(argv []*Object) (*Object, error) {
	if NumberEqual(argv[0].fval, 0.0) {
		return True, nil
	}
	return False, nil
}

func ellNot(argv []*Object) (*Object, error) {
	if argv[0] == False {
		return True, nil
	}
	return False, nil
}

func ellNullP(argv []*Object) (*Object, error) {
	if argv[0] == Null {
		return True, nil
	}
	return False, nil
}

func ellBooleanP(argv []*Object) (*Object, error) {
	if IsBoolean(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbolP(argv []*Object) (*Object, error) {
	if IsSymbol(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbol(argv []*Object) (*Object, error) {
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	return Symbol(argv)
}

func ellKeywordP(argv []*Object) (*Object, error) {
	if IsKeyword(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellKeywordName(argv []*Object) (*Object, error) {
	return KeywordName(argv[0])
}

func ellTypeP(argv []*Object) (*Object, error) {
	if IsType(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellTypeName(argv []*Object) (*Object, error) {
	return TypeName(argv[0])
}

func ellStringP(argv []*Object) (*Object, error) {
	if IsString(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellCharacterP(argv []*Object) (*Object, error) {
	if IsCharacter(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellToCharacter(argv []*Object) (*Object, error) {
	return ToCharacter(argv[0])
}

func ellFunctionP(argv []*Object) (*Object, error) {
	if IsFunction(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFunctionSignature(argv []*Object) (*Object, error) {
	return String(functionSignature(argv[0])), nil
}

func ellListP(argv []*Object) (*Object, error) {
	if IsList(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellEmptyP(argv []*Object) (*Object, error) {
	if argv[0] == EmptyList {
		return True, nil
	}
	return False, nil
}

func ellString(argv []*Object) (*Object, error) {
	s := ""
	for _, ss := range argv {
		s += ss.String()
	}
	return String(s), nil
}

func ellStringLength(argv []*Object) (*Object, error) {
	return Number(float64(StringLength(argv[0].text))), nil
}

func ellCar(argv []*Object) (*Object, error) {
	lst := argv[0]
	if lst == EmptyList {
		return Null, nil
	}
	return lst.car, nil
}

func ellCdr(argv []*Object) (*Object, error) {
	lst := argv[0]
	if lst == EmptyList {
		return lst, nil
	}
	return lst.cdr, nil
}

func ellSetCarBang(argv []*Object) (*Object, error) {
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

func ellSetCdrBang(argv []*Object) (*Object, error) {
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

func ellCons(argv []*Object) (*Object, error) {
	return Cons(argv[0], argv[1]), nil
}

func ellStructP(argv []*Object) (*Object, error) {
	if IsStruct(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellGet(argv []*Object) (*Object, error) {
	return structGet(argv[0], argv[1]), nil
}

func ellStructLength(argv []*Object) (*Object, error) {
	return Number(float64(StructLength(argv[0]))), nil
}

func ellHasP(argv []*Object) (*Object, error) {
	b, err := Has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return True, nil
	}
	return False, nil
}

func ellSeal(argv []*Object) (*Object, error) {
	switch argv[0].Type {
	case StructType, VectorType, ListType:
		argv[0].fval = 1
		return argv[0], nil
	default:
		return nil, Error(ArgumentErrorKey, "cannot seal! ", argv[0])
	}
}

func ellPutBang(argv []*Object) (*Object, error) {
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

func ellUnputBang(argv []*Object) (*Object, error) {
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

func ellToList(argv []*Object) (*Object, error) {
	return ToList(argv[0])
}

func ellSplit(argv []*Object) (*Object, error) {
	return StringSplit(argv[0], argv[1])
}

func ellJoin(argv []*Object) (*Object, error) {
	return StringJoin(argv[0], argv[1])
}

func ellJSON(argv []*Object) (*Object, error) {
	s, err := writeToString(argv[0], true, argv[1].text)
	if err != nil {
		return nil, err
	}
	return String(s), nil
}

func ellGetFn(argv []*Object) (*Object, error) {
	if len(argv) < 1 {
		return nil, Error(ArgumentErrorKey, "getfn expected at least 1 argument, got none")
	}
	sym := argv[0]
	if sym.Type != SymbolType {
		return nil, Error(ArgumentErrorKey, "getfn expected a <symbol> for argument 1, got ", sym)
	}
	return getfn(sym, argv[1:])
}

func ellMethodSignature(argv []*Object) (*Object, error) {
	return methodSignature(argv[0])
}

func ellChannel(argv []*Object) (*Object, error) {
	name := argv[0].text
	bufsize := int(argv[1].fval)
	return Channel(bufsize, name), nil
}

func ellClose(argv []*Object) (*Object, error) {
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

func ellSend(argv []*Object) (*Object, error) {
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

func ellReceive(argv []*Object) (*Object, error) {
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

func ellSetRandomSeedBang(argv []*Object) (*Object, error) {
	RandomSeed(int64(argv[0].fval))
	return Null, nil
}

func ellRandom(argv []*Object) (*Object, error) {
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

func ellRandomList(argv []*Object) (*Object, error) {
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

func ellUUIDFromTime(argv []*Object) (*Object, error) {
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

func Timestamp(t time.Time) *Object {
	format := "%d-%02d-%02dT%02d:%02d:%02d.%03dZ"
	return String(fmt.Sprintf(format, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000))
}

func ellTimestamp(argv []*Object) (*Object, error) {
	return Timestamp(time.Now().UTC()), nil
}

func ellBlobP(argv []*Object) (*Object, error) {
	if argv[0].Type == BlobType {
		return True, nil
	}
	return False, nil
}

func ellToBlob(argv []*Object) (*Object, error) {
	return ToBlob(argv[0])
}

func ellMakeBlob(argv []*Object) (*Object, error) {
	size := int(argv[0].fval)
	return MakeBlob(size), nil
}

func ellBlobLength(argv []*Object) (*Object, error) {
	return Number(float64(len(BlobValue(argv[0])))), nil
}

func ellBlobRef(argv []*Object) (*Object, error) {
	el := BlobValue(argv[0])
	idx := int(argv[1].fval)
	if idx < 0 || idx >= len(el) {
		return nil, Error(ArgumentErrorKey, "Blob index out of range")
	}
	return Number(float64(el[idx])), nil
}

func ellListen(argv []*Object) (*Object, error) {
	port := fmt.Sprintf(":%d", int(argv[0].fval))
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	acceptChan := Channel(10, fmt.Sprintf("tcp listener on %s", port))
	go tcpListener(listener, acceptChan, port)
	return acceptChan, nil
}

func ellConnect(argv []*Object) (*Object, error) {
	host := argv[0].text
	port := int(argv[1].fval)
	endpoint := fmt.Sprintf("%s:%d", host, port)
	con, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	return Connection(con, endpoint), nil
}

func ellHTTPServer(argv []*Object) (*Object, error) {
	port := int(argv[0].fval)
	handler := argv[1] // a function of one <struct> argument
	if handler.code == nil || handler.code.argc != 1 {
		return nil, Error(ArgumentErrorKey, "Cannot use this function as a handler: ", handler)
	}
	return httpServer(port, handler)
}

func ellHTTPClient(argv []*Object) (*Object, error) {
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

func ellNow(argv []*Object) (*Object, error) {
	return Number(Now()), nil
}

func ellSince(argv []*Object) (*Object, error) {
	then := argv[0].fval
	dur := Now() - then
	return Number(dur), nil
}

func Sleep(delayInSeconds float64) {
	dur := time.Duration(delayInSeconds * float64(time.Second))
	time.Sleep(dur) //!! this is not interruptable, fairly risky in a REPL
}

func ellSleep(argv []*Object) (*Object, error) {
	Sleep(argv[0].fval)
	return Number(Now()), nil
}
