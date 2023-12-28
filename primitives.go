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
	"os"
	"strings"
	"time"

	. "github.com/boynton/ell/data"
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
	DefineFunction("to-keyword", ellToKeyword, KeywordType, AnyType)
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
	DefineFunction("substring", ellSubstring, StringType, StringType, NumberType, NumberType)

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
	DefineFunctionOptionalArgs("make-vector", ellMakeVector, VectorType, []Value{NumberType, AnyType}, Null)
	DefineFunction("make-vector2", ellMakeVector, VectorType, NumberType, AnyType) //fixed number of args is faster. Compiler should figure it out!
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
	DefineFunction("read", ellRead, AnyType, StringType)
	DefineFunction("read-all", ellReadAll, AnyType, StringType)
	DefineFunction("spit", ellSpit, NullType, StringType, StringType)
	DefineFunctionKeyArgs("write", ellWrite, NullType, []Value{AnyType, StringType}, []Value{EmptyString}, []Value{Intern("indent:")})
	DefineFunctionKeyArgs("write-all", ellWriteAll, NullType, []Value{AnyType, StringType}, []Value{EmptyString}, []Value{Intern("indent:")})
	DefineFunctionRestArgs("print", ellPrint, NullType, AnyType)
	DefineFunctionRestArgs("println", ellPrintln, NullType, AnyType)
	DefineFunction("macroexpand", ellMacroexpand, AnyType, AnyType)
	DefineFunction("compile", ellCompile, CodeType, AnyType)

	DefineFunctionRestArgs("make-error", ellMakeError, ErrorType, AnyType)
	DefineFunction("error?", ellErrorP, BooleanType, AnyType)
	DefineFunction("error-data", ellErrorData, AnyType, ErrorType)
	DefineFunction("uncaught-error", ellUncaughtError, NullType, ErrorType) //doesn't return

	DefineFunctionKeyArgs("json", ellJSON, StringType, []Value{AnyType, StringType}, []Value{EmptyString}, []Value{Intern("indent:")})

	DefineFunctionRestArgs("getfn", ellGetFn, FunctionType, AnyType, SymbolType)
	DefineFunction("method-signature", ellMethodSignature, TypeType, ListType)

	DefineFunction("now", ellNow, NumberType)
	DefineFunction("since", ellSince, NumberType, NumberType)
	DefineFunction("sleep", ellSleep, NumberType, NumberType)

	DefineFunctionKeyArgs("channel", ellChannel, ChannelType, []Value{StringType, NumberType}, []Value{EmptyString, Zero}, []Value{Intern("name:"), Intern("bufsize:")})
	DefineFunctionOptionalArgs("send", ellSend, NullType, []Value{ChannelType, AnyType, NumberType}, MinusOne)
	DefineFunctionOptionalArgs("recv", ellReceive, AnyType, []Value{ChannelType, NumberType}, MinusOne)
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
		[]Value{StringType, StringType, StructType, BlobType}, //(http "url" method: "PUT" headers: {} body: #[blob])
		[]Value{NewString("GET"), EmptyStruct, EmptyBlob},
		[]Value{Intern("method:"), Intern("headers:"), Intern("body:")})

	DefineFunction("getenv", ellGetenv, StringType, StringType)
	DefineFunction("load", ellLoad, StringType, AnyType)

	if true {
		err := Load("ell")
		if err != nil {
			Fatal("*** ", err)
		}
	}
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *LList
//

func ellLetrec(argv []Value) (Value, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []Value) (Value, error) {
	return expandLet(argv[0])
}

func ellCond(argv []Value) (Value, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []Value) (Value, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(_ []Value) (Value, error) {
	s := "ell " + Version
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
	return NewString(s), nil
}

func ellDefinedP(argv []Value) (Value, error) {
	if p, ok := argv[0].(*Symbol); ok {
		if IsDefined(p) {
			return True, nil
		}
	}
	return False, nil
}

func ellSlurp(argv []Value) (Value, error) {
	url := StringValue(argv[0])
	if strings.HasPrefix(url, "http:") || strings.HasPrefix(url, "https:") {
		res, err := httpClientOperation("GET", url, nil, nil)
		if err != nil {
			return nil, err
		}
		status := IntValue(res.Get(Intern("status:")))
		if status != 200 {
			return nil, NewError(HTTPErrorKey, status, " ", http.StatusText(status))
		}
		s, _ := ToString(res.Get(Intern("body:")))
		return s, nil
	}
	s, err := SlurpFile(StringValue(argv[0]))
	if err != nil {
		return nil, err
	}
	return NewString(s), nil
}

func ellSpit(argv []Value) (Value, error) {
	url := StringValue(argv[0])
	data := StringValue(argv[1])
	err := SpitFile(url, data)
	if err != nil {
		return nil, err
	}
	return Null, nil
}

func ellRead(argv []Value) (Value, error) {
	return ReadFromString(StringValue(argv[0]))
}

func ellReadAll(argv []Value) (Value, error) {
	return ReadAllFromString(StringValue(argv[0]))
}

func ellMacroexpand(argv []Value) (Value, error) {
	return Macroexpand(argv[0])
}

func ellCompile(argv []Value) (Value, error) {
	expanded, err := Macroexpand(argv[0])
	if err != nil {
		return nil, err
	}
	return Compile(expanded)
}

func ellLoad(argv []Value) (Value, error) {
	err := Load(StringValue(argv[0]))
	return argv[0], err
}

func ellType(argv []Value) (Value, error) {
	return argv[0].Type(), nil
}

func ellValue(argv []Value) (Value, error) {
	return Value(argv[0]), nil
}

func ellInstance(argv []Value) (Value, error) {
	return NewInstance(argv[0], argv[1])
}

func ellValidateKeywordArgList(argv []Value) (Value, error) {
	//(validate-keyword-arg-list '(x: 23) x: y:) -> (x:)
	//(validate-keyword-arg-list '(x: 23 z: 100) x: y:) -> error("bad keyword z: in argument list")
	return validateKeywordArgList(argv[0].(*List), argv[1:])
}

func ellKeys(argv []Value) (Value, error) {
	return structKeyList(argv[0].(*Struct)), nil
}

func ellValues(argv []Value) (Value, error) {
	return structValueList(argv[0].(*Struct)), nil
}

func ellStruct(argv []Value) (Value, error) {
	return MakeStruct(argv)
}
func ellMakeStruct(argv []Value) (Value, error) {
	return NewStruct(), nil
}

func ellToStruct(argv []Value) (Value, error) {
	//how about a keys: keyword argument to force a key type, like read does?
	return ToStruct(argv[0])
}

func ellIdenticalP(argv []Value) (Value, error) {
	if argv[0] == argv[1] {
		return True, nil
	}
	return False, nil
}

func ellEqualP(argv []Value) (Value, error) {
	if Equal(argv[0], argv[1]) {
		return True, nil
	}
	return False, nil
}

func numeq(n1 Value, n2 Value) bool {
	if f1, ok := n1.(*Number); ok {
		if f2, ok := n2.(*Number); ok {
			return NumberEqual(f1.Value, f2.Value)
		}
	}
	return false
}

func numericPair(argv []Value) (float64, float64, error) {
	return (argv[0].(*Number)).Value, (argv[1].(*Number)).Value, nil
	/*	f1, err := AsFloat64Value(argv[0])
		if err != nil {
			return 0, 0, err
		}
		f2, err := AsFloat64Value(argv[1])
		if err != nil {
			return 0, 0, err
		}
		return f1, f2, nil
	*/
}

func ellNumEqual(argv []Value) (Value, error) {
	f1, f2, err := numericPair(argv)
	if err == nil {
		if NumberEqual(f1, f2) {
			return True, nil
		}
	}
	return False, err
}

func ellNumLess(argv []Value) (Value, error) {
	f1, f2, err := numericPair(argv)
	if err == nil {
		if f1 < f2 {
			return True, nil
		}
	}
	return False, err
}

func ellNumLessEqual(argv []Value) (Value, error) {
	f1, f2, err := numericPair(argv)
	if err == nil {
		if f1 <= f2 {
			return True, nil
		}
	}
	return False, err
}

func ellNumGreater(argv []Value) (Value, error) {
	f1, f2, err := numericPair(argv)
	if err == nil {
		if f1 > f2 {
			return True, nil
		}
	}
	return False, err
}

func ellNumGreaterEqual(argv []Value) (Value, error) {
	f1, f2, err := numericPair(argv)
	if err == nil {
		if f1 >= f2 {
			return True, nil
		}
	}
	return False, err
}

func ellWrite(argv []Value) (Value, error) {
	return NewString(WriteIndent(argv[0], StringValue(argv[1]))), nil
}

func ellWriteAll(argv []Value) (Value, error) {
	if lst, ok := argv[0].(*List); ok {
		return NewString(WriteAllIndent(lst, StringValue(argv[1]))), nil
	}
	return nil, NewError(ArgumentErrorKey, "Expected a <list>, but got a ", argv[0].Type())
}

func ellMakeError(argv []Value) (Value, error) {
	return MakeError(argv...), nil
}

func ellErrorP(argv []Value) (Value, error) {
	if _, ok := argv[0].(*Error); ok {
		return True, nil
	}
	return False, nil
}

func ellErrorData(argv []Value) (Value, error) {
	if p, ok := argv[0].(*Error); ok {
		return p.Data, nil
	}
	return nil, NewError(ArgumentErrorKey, "Expected an <error>, but got a ", argv[0].Type())
}

func ellUncaughtError(argv []Value) (Value, error) {
	if p, ok := argv[0].(*Error); ok {
		return nil, p
	}
	return Null, nil
}

func ellToString(argv []Value) (Value, error) {
	return ToString(argv[0])
}

func ellPrint(argv []Value) (Value, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []Value) (Value, error) {
	ellPrint(argv)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []Value) (Value, error) {
	result := EmptyList
	tail := result
	for _, obj := range argv {
		lst, _ := obj.(*List)
		for lst != EmptyList {
			if tail == EmptyList {
				result = NewList(lst.Car)
				tail = result
			} else {
				tail.Cdr = NewList(lst.Car)
				tail = tail.Cdr
			}
			lst = lst.Cdr
		}
	}
	return result, nil
}

func AsList(obj Value) *List {
	lst, _ := obj.(*List)
	return lst
}

func ellReverse(argv []Value) (Value, error) {
	return Reverse(AsList(argv[0])), nil
}

func ellFlatten(argv []Value) (Value, error) {
	return Flatten(AsList(argv[0])), nil
}

func ellList(argv []Value) (Value, error) {
	argc := len(argv)
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ellListLength(argv []Value) (Value, error) {
	return Integer(ListLength(argv[0])), nil
}

func IsNumber(val Value) bool {
	return val.Type() == NumberType
}

func ellNumberP(argv []Value) (Value, error) {
	if IsNumber(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellToNumber(argv []Value) (Value, error) {
	return ToNumber(argv[0])
}

func ellIntP(argv []Value) (Value, error) {
	if IsInt(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFloatP(argv []Value) (Value, error) {
	if IsFloat(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellInt(argv []Value) (Value, error) {
	return ToInt(argv[0])
}

func ellFloor(argv []Value) (Value, error) {
	return Float(math.Floor(Float64Value(argv[0]))), nil
}

func ellCeiling(argv []Value) (Value, error) {
	return Float(math.Ceil(Float64Value(argv[0]))), nil
}

func ellInc(argv []Value) (Value, error) {
	return Integer(IntValue(argv[0]) + 1), nil
}

func ellDec(argv []Value) (Value, error) {
	return Integer(IntValue(argv[0]) - 1), nil
}

func ellAdd(argv []Value) (Value, error) {
	//f1, f2, _ := numericPair(argv) //
	return Float((argv[0].(*Number)).Value + (argv[1].(*Number)).Value), nil //interesting: inlining like this instead of numericPair is no faster
}

func ellSub(argv []Value) (Value, error) {
	return Float((argv[0].(*Number)).Value - (argv[1].(*Number)).Value), nil
}

func ellMul(argv []Value) (Value, error) {
	return Float((argv[0].(*Number)).Value * (argv[1].(*Number)).Value), nil
}

func ellDiv(argv []Value) (Value, error) {
	return Float((argv[0].(*Number)).Value / (argv[1].(*Number)).Value), nil
}

func ellQuotient(argv []Value) (Value, error) {
	f1 := (argv[0].(*Number)).Value
	f2 := (argv[1].(*Number)).Value
	return Float(math.Floor(f1 / f2)), nil
}

func ellRemainder(argv []Value) (Value, error) {
	return Integer(int((argv[0].(*Number)).Value) % int((argv[1].(*Number)).Value)), nil
}

func ellAbs(argv []Value) (Value, error) {
	return Float(math.Abs((argv[0].(*Number)).Value)), nil
}

func ellExp(argv []Value) (Value, error) {
	return Float(math.Exp((argv[0].(*Number)).Value)), nil
}

func ellLog(argv []Value) (Value, error) {
	return Float(math.Log(Float64Value(argv[0]))), nil
}

func ellSin(argv []Value) (Value, error) {
	return Float(math.Sin(Float64Value(argv[0]))), nil
}

func ellCos(argv []Value) (Value, error) {
	return Float(math.Cos(Float64Value(argv[0]))), nil
}

func ellTan(argv []Value) (Value, error) {
	return Float(math.Tan(Float64Value(argv[0]))), nil
}

func ellAsin(argv []Value) (Value, error) {
	return Float(math.Asin(Float64Value(argv[0]))), nil
}

func ellAcos(argv []Value) (Value, error) {
	return Float(math.Acos(Float64Value(argv[0]))), nil
}

func ellAtan(argv []Value) (Value, error) {
	return Float(math.Atan(Float64Value(argv[0]))), nil
}

func ellAtan2(argv []Value) (Value, error) {
	f1, f2, err := numericPair(argv)
	return Float(math.Atan2(f1, f2)), err
}

func ellVector(argv []Value) (Value, error) {
	return NewVector(argv...), nil
}

func ellToVector(argv []Value) (Value, error) {
	return ToVector(argv[0])
}

func ellMakeVector(argv []Value) (Value, error) {
	vlen := IntValue(argv[0])
	init := argv[1]
	return MakeVector(vlen, init), nil
}

func ellVectorP(argv []Value) (Value, error) {
	if IsVector(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellVectorLength(argv []Value) (Value, error) {
	vec, _ := argv[0].(*Vector)
	return Integer(len(vec.Elements)), nil
}

func ellVectorRef(argv []Value) (Value, error) {
	vec, _ := argv[0].(*Vector)
	el := vec.Elements
	idx := IntValue(argv[1])
	if idx < 0 || idx >= len(el) {
		return nil, NewError(ArgumentErrorKey, "Vector index out of range")
	}
	return el[idx], nil
}

func ellVectorSetBang(argv []Value) (Value, error) {
	vec, _ := argv[0].(*Vector)
	el := vec.Elements
	idx := IntValue(argv[1])
	if idx < 0 || idx > len(el) {
		return nil, NewError(ArgumentErrorKey, "Vector index out of range")
	}
	el[idx] = argv[2]
	return Null, nil
}

func ellZeroP(argv []Value) (Value, error) {
	if NumberEqual(Float64Value(argv[0]), 0.0) {
		return True, nil
	}
	return False, nil
}

func ellNot(argv []Value) (Value, error) {
	if argv[0] == False {
		return True, nil
	}
	return False, nil
}

func ellNullP(argv []Value) (Value, error) {
	if argv[0] == Null {
		return True, nil
	}
	return False, nil
}

func ellBooleanP(argv []Value) (Value, error) {
	if _, ok := argv[0].(*Boolean); ok {
		return True, nil
	}
	return False, nil
}

func ellSymbolP(argv []Value) (Value, error) {
	if IsSymbol(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellSymbol(argv []Value) (Value, error) {
	if len(argv) < 1 {
		return nil, NewError(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	return NewSymbol(argv)
}

func ellKeywordP(argv []Value) (Value, error) {
	if argv[0].Type() == KeywordType {
		return True, nil
	}
	return False, nil
}

func ellKeywordName(argv []Value) (Value, error) {
	return Intern((argv[0].(*Keyword)).Name()), nil
}

func ellToKeyword(argv []Value) (Value, error) {
	return ToKeyword(argv[0])
}

func ellTypeP(argv []Value) (Value, error) {
	if IsType(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellTypeName(argv []Value) (Value, error) {
	return Intern((argv[0].(*Type)).Name()), nil
}

func ellStringP(argv []Value) (Value, error) {
	if argv[0].Type() == StringType {
		return True, nil
	}
	return False, nil
}

func ellCharacterP(argv []Value) (Value, error) {
	if argv[0].Type() == CharacterType {
		return True, nil
	}
	return False, nil
}

func ellToCharacter(argv []Value) (Value, error) {
	return ToCharacter(argv[0])
}

func ellSubstring(argv []Value) (Value, error) {
	s := StringValue(argv[0])
	start := IntValue(argv[1])
	end := IntValue(argv[2])
	if start < 0 {
		start = 0
	} else if start > len(s) {
		return EmptyString, nil
	}
	if end < start {
		return EmptyString, nil
	} else if end > len(s) {
		end = len(s)
	}
	return NewString(s[start:end]), nil
}

func ellFunctionP(argv []Value) (Value, error) {
	if argv[0].Type() == FunctionType {
		return True, nil
	}
	return False, nil
}

func ellFunctionSignature(argv []Value) (Value, error) {
	fun, _ := argv[0].(*Function)
	return NewString(functionSignature(fun)), nil
}

func ellListP(argv []Value) (Value, error) {
	if argv[0].Type() == ListType {
		return True, nil
	}
	return False, nil
}

func ellEmptyP(argv []Value) (Value, error) {
	if argv[0] == EmptyList {
		return True, nil
	}
	return False, nil
}

func ellString(argv []Value) (Value, error) {
	s := ""
	for _, ss := range argv {
		s += ss.String()
	}
	return NewString(s), nil
}

func ellStringLength(argv []Value) (Value, error) {
	s, _ := argv[0].(*String)
	return Integer(len(s.Value)), nil
}

func ellCar(argv []Value) (Value, error) {
	lst := argv[0].(*List)
	if lst == EmptyList {
		return Null, nil
	}
	return lst.Car, nil
}

func ellCdr(argv []Value) (Value, error) {
	lst := argv[0].(*List)
	if lst == EmptyList {
		return lst, nil
	}
	return lst.Cdr, nil
}

func ellSetCarBang(argv []Value) (Value, error) {
	lst := argv[0].(*List)
	if lst == EmptyList {
		return nil, NewError(ArgumentErrorKey, "set-car! expected a non-empty <list>")
	}
	lst.Car = argv[1]
	return Null, nil
}

func ellSetCdrBang(argv []Value) (Value, error) {
	lst := argv[0].(*List)
	if lst == EmptyList {
		return nil, NewError(ArgumentErrorKey, "set-cdr! expected a non-empty <list>")
	}
	lst.Cdr = argv[1].(*List)
	return Null, nil
}

func ellCons(argv []Value) (Value, error) {
	return Cons(argv[0], argv[1].(*List)), nil
}

func ellStructP(argv []Value) (Value, error) {
	if IsStruct(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellGet(argv []Value) (Value, error) {
	return Get(argv[0], argv[1])
}

func ellStructLength(argv []Value) (Value, error) {
	s := argv[0].(*Struct)
	return Integer(len(s.Bindings)), nil
}

func ellHasP(argv []Value) (Value, error) {
	b, err := Has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return True, nil
	}
	return False, nil
}

func ellPutBang(argv []Value) (Value, error) {
	key := argv[1]
	if !IsValidStructKey(key) {
		return nil, NewError(ArgumentErrorKey, "Bad struct key: ", key)
	}
	Put(argv[0], key, argv[2])
	return Null, nil
}

func ellUnputBang(argv []Value) (Value, error) {
	key := argv[1]
	if !IsValidStructKey(key) {
		return nil, NewError(ArgumentErrorKey, "Bad struct key: ", key)
	}
	Unput(argv[0], key)
	return Null, nil
}

func ellToList(argv []Value) (Value, error) {
	return ToList(argv[0])
}

func ellSplit(argv []Value) (Value, error) {
	return StringSplit(argv[0], argv[1])
}

func ellJoin(argv []Value) (Value, error) {
	return StringJoin(argv[0], argv[1])
}

func ellJSON(argv []Value) (Value, error) {
	s, err := Json(argv[0], StringValue(argv[1]))
	if err != nil {
		return nil, err
	}
	return NewString(s), nil
}

func ellGetFn(argv []Value) (Value, error) {
	if len(argv) < 1 {
		return nil, NewError(ArgumentErrorKey, "getfn expected at least 1 argument, got none")
	}
	sym := argv[0]
	if sym.Type() != SymbolType {
		return nil, NewError(ArgumentErrorKey, "getfn expected a <symbol> for argument 1, got ", sym)
	}
	return getfn(sym, argv[1:])
}

func ellMethodSignature(argv []Value) (Value, error) {
	return methodSignature(argv[0].(*List))
}

func ellChannel(argv []Value) (Value, error) {
	name := StringValue(argv[0])
	bufsize := IntValue(argv[1])
	return NewChannel(bufsize, name), nil
}

func ellClose(argv []Value) (Value, error) {
	switch p := argv[0].(type) {
	case *Channel:
		CloseChannel(p)
	case *Connection:
		closeConnection(p)
	default:
		return nil, NewError(ArgumentErrorKey, "close expected a channel or connection")
	}
	return Null, nil
}

func ellSend(argv []Value) (Value, error) {
	ch := ChannelValue(argv[0])
	if ch != nil { //not closed
		val := argv[1]
		timeout := Float64Value(argv[2]) //FIX: timeouts in seconds, floating point
		if NumberEqual(timeout, 0.0) {   //non-blocking
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

func ellReceive(argv []Value) (Value, error) {
	ch := ChannelValue(argv[0])
	if ch != nil { //not closed
		timeout := Float64Value(argv[1])
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

func ellSetRandomSeedBang(argv []Value) (Value, error) {
	RandomSeed(int64(IntValue(argv[0])))
	return Null, nil
}

func ellRandom(argv []Value) (Value, error) {
	min := 0.0
	max := 1.0
	argc := len(argv)
	switch argc {
	case 0:
	case 1:
		max = Float64Value(argv[0])
	case 2:
		min = Float64Value(argv[0])
		max = Float64Value(argv[1])
	default:
		return nil, NewError(ArgumentErrorKey, "random expected 0 to 2 arguments, got ", argc)
	}
	return Random(min, max), nil
}

func ellRandomList(argv []Value) (Value, error) {
	count := IntValue(argv[0])
	min := 0.0
	max := 1.0
	argc := len(argv)
	switch argc {
	case 1:
	case 2:
		max = Float64Value(argv[0])
	case 3:
		min = Float64Value(argv[0])
		max = Float64Value(argv[1])
	default:
		return nil, NewError(ArgumentErrorKey, "random-list expected 1 to 3 arguments, got ", argc)
	}
	return RandomList(count, min, max), nil
}

func ellUUIDFromTime(argv []Value) (Value, error) {
	var u uuid.UUID
	argc := len(argv)
	switch argc {
	case 0:
		u = uuid.NewUUID()
	case 1:
		u = uuid.NewMD5(uuid.NameSpace_URL, []byte(StringValue(argv[0])))
	case 2:
		ns := uuid.Parse(StringValue(argv[0]))
		if ns == nil {
			ns = uuid.NewMD5(uuid.NameSpace_URL, []byte(StringValue(argv[0])))
		}
		u = uuid.NewMD5(ns, []byte(StringValue(argv[1])))
	}
	if u == nil {
		return nil, NewError(ArgumentErrorKey, "Expected 0-2 arguments, got: ", argc)
	}
	return NewString(u.String()), nil
}

func CurrentTimestamp(t time.Time) Value {
	format := "%d-%02d-%02dT%02d:%02d:%02d.%03dZ"
	return NewString(fmt.Sprintf(format, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000))
}

func ellTimestamp(_ []Value) (Value, error) {
	return CurrentTimestamp(time.Now().UTC()), nil
}

func ellBlobP(argv []Value) (Value, error) {
	if argv[0].Type() == BlobType {
		return True, nil
	}
	return False, nil
}

func ellToBlob(argv []Value) (Value, error) {
	return ToBlob(argv[0])
}

func ellMakeBlob(argv []Value) (Value, error) {
	size := IntValue(argv[0])
	return MakeBlob(size), nil
}

func ellBlobLength(argv []Value) (Value, error) {
	blob := argv[0].(*Blob)
	return Integer(len(blob.Value)), nil
}

func ellBlobRef(argv []Value) (Value, error) {
	blob := argv[0].(*Blob)
	el := blob.Value
	idx := IntValue(argv[1])
	if idx < 0 || idx >= len(el) {
		return nil, NewError(ArgumentErrorKey, "Blob index out of range")
	}
	return Integer(int(el[idx])), nil
}

func ellListen(argv []Value) (Value, error) {
	port := fmt.Sprintf(":%d", IntValue(argv[0]))
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}
	acceptChan := NewChannel(10, fmt.Sprintf("tcp listener on %s", port))
	go tcpListener(listener, acceptChan, port)
	return acceptChan, nil
}

func ellConnect(argv []Value) (Value, error) {
	host := StringValue(argv[0])
	port := IntValue(argv[1])
	endpoint := fmt.Sprintf("%s:%d", host, port)
	con, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	return NewConnection(con, endpoint), nil
}

func ellHTTPServer(argv []Value) (Value, error) {
	port := IntValue(argv[0])
	handler := argv[1].(*Function) // a function of one <struct> argument
	if handler.code == nil || handler.code.argc != 1 {
		return nil, NewError(ArgumentErrorKey, "Cannot use this function as a handler: ", handler)
	}
	return httpServer(port, handler)
}

func ellHTTPClient(argv []Value) (Value, error) {
	url := StringValue(argv[0])
	method := strings.ToUpper(StringValue(argv[1]))
	headers := argv[2].(*Struct)
	body := argv[3].(*String)
	println("http ", method, " of ", url, ", headers: ", headers, ", body: ", body)
	switch method {
	case "GET", "PUT", "POST", "DELETE", "HEAD", "OPTIONS", "PATCH":
		return httpClientOperation(method, url, headers, body)
	default:
		return nil, NewError(ErrorKey, "HTTP method not support: ", method)
	}
}

func Now() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / float64(time.Second)
}

func ellNow(_ []Value) (Value, error) {
	return Float(Now()), nil
}

func ellSince(argv []Value) (Value, error) {
	then := Float64Value(argv[0])
	dur := Now() - then
	return Float(dur), nil
}

func Sleep(delayInSeconds float64) {
	dur := time.Duration(delayInSeconds * float64(time.Second))
	time.Sleep(dur) //!! this is not interruptable, fairly risky in a REPL
}

func ellSleep(argv []Value) (Value, error) {
	Sleep(Float64Value(argv[0]))
	return Float(Now()), nil
}

func ellGetenv(argv []Value) (Value, error) {
	s := os.Getenv(StringValue(argv[0]))
	if s == "" {
		return Null, nil
	}
	return NewString(s), nil
}
