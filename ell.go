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

// Ell defines the global functions for the top level environment
func Ell(module *Module) {

	module.defineMacro("let", ellLet)
	module.defineMacro("letrec", ellLetrec)
	module.defineMacro("do", ellDo) //scheme's do. I don't like it, will replace with... "for" of some sort
	//note clojure uses "do" instead of "begin". I rather prefer that.
	module.defineMacro("cond", ellCond)
	module.defineMacro("quasiquote", ellQuasiquote)

	module.define("null", Null)
	module.define("true", True)
	module.define("false", False)

	module.define("apply", Apply)

	module.defineFunction("version", ellVersion)

	module.defineFunction("defined?", ellDefinedP)

	module.defineFunction("file-contents", ellFileContents)
	module.defineFunction("open-input-string", ellOpenInputString)
	module.defineFunction("open-input-file", ellOpenInputFile)
	module.defineFunction("read", ellRead)
	module.defineFunction("close-input-port", ellCloseInputPort)

	module.defineFunction("macroexpand", ellMacroexpand)
	module.defineFunction("type", ellType)
	module.defineFunction("normalize-keyword-args", ellNormalizeKeywordArgs)

	module.defineFunction("type?", ellTypeP)
	module.defineFunction("type-name", ellTypeName)

	module.defineFunction("struct", ellStruct)
	module.defineFunction("instance", ellInstance)
	module.defineFunction("equal?", ellEq)
	module.defineFunction("identical?", ellIdenticalP)
	module.defineFunction("not", ellNot)

	module.defineFunction("boolean?", ellBooleanP)
	module.defineFunction("null?", ellNullP)
	module.defineFunction("symbol?", elSymbolTypeP)
	module.defineFunction("symbol", elSymbolType)

	module.defineFunction("keyword?", ellKeywordP)
	//	module.defineFunction("keyword", ellKeyword)
	module.defineFunction("unkeyword", ellUnkeyword)
	module.defineFunction("string?", elStringTypeP)
	module.defineFunction("char?", ellCharP)
	module.defineFunction("function?", ellFunctionP)
	module.defineFunction("eof?", ellFunctionP)

	module.defineFunction("list?", elListTypeP)
	module.defineFunction("cons", ellCons)
	module.defineFunction("car", ellCar)
	module.defineFunction("cdr", ellCdr)
	module.defineFunction("list", elListType)
	module.defineFunction("concat", ellConcat)
	module.defineFunction("reverse", ellReverse)
	module.defineFunction("set-car!", ellSetCarBang)
	module.defineFunction("set-cdr!", ellSetCdrBang)

	module.defineFunction("array?", ellArrayP)
	module.defineFunction("array", ellArray)
	module.defineFunction("make-array", ellMakeArray)
	module.defineFunction("array-set!", ellArraySetBang)
	module.defineFunction("array-ref", ellArrayRef)

	module.defineFunction("struct?", ellStructP)
	module.defineFunction("has?", ellHasP)
	module.defineFunction("get", ellGet)
	module.defineFunction("put!", ellPutBang)
	module.defineFunction("struct->list", ellStructToList)

	module.defineFunction("empty?", ellEmptyP)

	module.defineFunction("string", elStringType)
	module.defineFunction("display", ellDisplay)
	module.defineFunction("write", ellWrite)
	module.defineFunction("newline", ellNewline)
	module.defineFunction("print", ellPrint)
	module.defineFunction("println", ellPrintln)
	module.defineFunction("write-to-string", ellWriteToString)

	module.defineFunction("number?", elNumberTypeP) // either float or int
	module.defineFunction("int?", ellIntP)          //int only
	module.defineFunction("float?", ellFloatP)      //float only
	module.defineFunction("+", ellPlus)
	module.defineFunction("-", ellMinus)
	module.defineFunction("*", ellTimes)
	module.defineFunction("/", ellDiv)
	module.defineFunction("quotient", ellQuotient)
	module.defineFunction("remainder", ellRemainder)
	module.defineFunction("modulo", ellRemainder) //fix!

	module.defineFunction("=", ellNumeq)
	module.defineFunction("<=", ellLe)
	module.defineFunction(">=", ellGe)
	module.defineFunction(">", ellGt)
	module.defineFunction("<", ellLt)
	module.defineFunction("zero?", ellZeroP)
	module.defineFunction("number->string", elNumberTypeToString)
	module.defineFunction("string-length", elStringTypeLength)

	module.defineFunction("error", ellFatal)
	module.defineFunction("length", ellLength)
	module.defineFunction("json", ellJSON)
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *ListType
//

func ellLetrec(argv []AnyType, argc int) (AnyType, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []AnyType, argc int) (AnyType, error) {
	return expandLet(argv[0])
}

func ellDo(argv []AnyType, argc int) (AnyType, error) {
	return expandDo(argv[0])
}

func ellCond(argv []AnyType, argc int) (AnyType, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []AnyType, argc int) (AnyType, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(argv []AnyType, argc int) (AnyType, error) {
	return StringType(Version), nil
}

func ellDefinedP(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("defined?", "1", argc)
	}
	if !isSymbol(argv[0]) {
		return ArgTypeError("symbol", 1, argv[0])
	}
	if currentModule != nil {
		if currentModule.isDefined(argv[0]) {
			return True, nil
		}
	}
	return False, nil
}

func ellFileContents(argv []AnyType, argc int) (AnyType, error) {
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
	return StringType(s), nil
}

func ellOpenInputString(argv []AnyType, argc int) (AnyType, error) {
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

func ellOpenInputFile(argv []AnyType, argc int) (AnyType, error) {
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

func ellRead(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("read", "1", argc)
	}
	return readInputPort(argv[0])
}
func ellCloseInputPort(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("read", "1", argc)
	}
	return nil, closeInputPort(argv[0])
}

func ellMacroexpand(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("macroexpand", "1", argc)
	}
	return macroexpand(argv[0])
}

func ellType(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("type", "1", argc)
	}
	return argv[0].Type(), nil
}

func ellNormalizeKeywordArgs(argv []AnyType, argc int) (AnyType, error) {
	//(normalize-keyword-args '(x: 23) '(x: y:) -> (x:)
	//(normalize-keyword-args '(x: 23 z: 100) '(x: y:) -> error("bad keyword z: in argument list")
	if argc < 1 {
		return ArgcError("normalizeKeywordArgs", "1+", argc)
	}
	if args, ok := argv[0].(*ListType); ok {
		return normalizeKeywordArgs(args, argv[1:argc])
	}
	return ArgTypeError("list", 1, argv[0])
}

func ellStruct(argv []AnyType, argc int) (AnyType, error) {
	return newInstance(typeStruct, argv[:argc])
}

func ellInstance(argv []AnyType, argc int) (AnyType, error) {
	if argc < 1 {
		return ArgcError("instance", "1+", argc)
	}
	switch s := argv[0].(type) {
	case *SymbolType:
		if isSymbolType(s) {
			return newInstance(s, argv[1:argc])
		}
	}
	return ArgTypeError("type", 1, argv[0])
}

func ellIdenticalP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("identical?", "2", argc)
}

func ellEq(argv []AnyType, argc int) (AnyType, error) {
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

func ellNumeq(argv []AnyType, argc int) (AnyType, error) {
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

func ellDisplay(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		//todo: add the optional port argument like scheme
		return ArgcError("display", "1", argc)
	}
	fmt.Printf("%v", argv[0])
	return Null, nil
}

func ellWrite(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		//todo: add the optional port argument like scheme
		return ArgcError("write", "1", argc)
	}
	fmt.Printf("%v", write(argv[0]))
	return Null, nil
}

func ellNewline(argv []AnyType, argc int) (AnyType, error) {
	if argc != 0 {
		//todo: add the optional port argument like scheme
		return ArgcError("newline", "0", argc)
	}
	fmt.Printf("\n")
	return Null, nil
}

func ellFatal(argv []AnyType, argc int) (AnyType, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}


func ellWriteToString(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("write-to-string", "1", argc)
	}
	s := write(argv[0])
	return StringType(s), nil
	
}

func ellPrint(argv []AnyType, argc int) (AnyType, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []AnyType, argc int) (AnyType, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []AnyType, argc int) (AnyType, error) {
	result := EmptyList
	tail := result
	for i := 0; i < argc; i++ {
		o := argv[i]
		switch lst := o.(type) {
		case *ListType:
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

func ellReverse(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("reverse", "1", argc)
	}
	o := argv[0]
	switch lst := o.(type) {
	case *ListType:
		return reverse(lst), nil
	default:
		return ArgTypeError("list", 1, o)
	}
}

func elListType(argv []AnyType, argc int) (AnyType, error) {
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func elNumberTypeP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("number?", "1", argc)
}

func ellIntP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isInt(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("int?", "1", argc)
}

func ellFloatP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isFloat(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("float?", "1", argc)
}

func ellQuotient(argv []AnyType, argc int) (AnyType, error) {
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
		return NumberType(n1 / n2), nil
	}
	return ArgcError("quotient", "2", argc)
}

func ellRemainder(argv []AnyType, argc int) (AnyType, error) {
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
		return NumberType(n1 % n2), nil
	}
	return ArgcError("remainder", "2", argc)
}

func ellPlus(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []AnyType, argc int) (AnyType, error) {
	return minus(argv, argc)
}

func ellTimes(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellDiv(argv []AnyType, argc int) (AnyType, error) {
	return div(argv, argc)
}

func ellArray(argv []AnyType, argc int) (AnyType, error) {
	return array(argv...), nil
}

func ellMakeArray(argv []AnyType, argc int) (AnyType, error) {
	if argc > 0 {
		initVal := AnyType(Null)
		vlen, err := intValue(argv[0])
		if err != nil {
			return nil, err
		}
		if argc > 1 {
			if argc != 2 {
				return ArgcError("make-array", "1-2", argc)
			}
			initVal = argv[1]
		}
		return newArray(int(vlen), initVal), nil
	}
	return ArgcError("make-array", "1-2", argc)
}

func ellArrayP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isArray(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("array?", "1", argc)
}

func ellArraySetBang(argv []AnyType, argc int) (AnyType, error) {
	if argc == 3 {
		a := argv[0]
		idx, err := intValue(argv[1])
		if err != nil {
			return nil, err
		}
		err = arraySet(a, int(idx), argv[2])
		if err != nil {
			return nil, err
		}
		return a, nil
	}
	return ArgcError("array-set!", "3", argc)
}

func ellArrayRef(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		a := argv[0]
		idx, err := intValue(argv[1])
		if err != nil {
			return nil, err
		}
		val, err := arrayRef(a, int(idx))
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return ArgcError("array-ref", "2", argc)
}

func ellGe(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError(">=", "2", argc)
}

func ellLe(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError("<=", "2", argc)
}

func ellGt(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError(">", "2", argc)
}

func ellLt(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError("<", "2", argc)
}

func ellZeroP(argv []AnyType, argc int) (AnyType, error) {
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

func elNumberTypeToString(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("number->string", "1", argc)
	}
	if !isNumber(argv[0]) {
		return ArgTypeError("number", 1, argv[0])
	}
	return StringType(argv[0].String()), nil
}

func elStringTypeLength(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("string-length", "1", argc)
	}
	if !isString(argv[0]) {
		return ArgTypeError("string", 1, argv[0])
	}
	i := length(argv[0])
	return NumberType(int64(i)), nil
}

func ellLength(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		return NumberType(int64(length(argv[0]))), nil
	}
	return ArgcError("length", "1", argc)
}

func ellNot(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("not", "1", argc)
}

func ellNullP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if argv[0] == Null {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("nil?", "1", argc)
}

func ellBooleanP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("boolean?", "1", argc)
}

func elSymbolTypeP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("symbol?", "1", argc)
}

func elSymbolType(argv []AnyType, argc int) (AnyType, error) {
	if argc < 1 {
		return ArgcError("symbol", "1+", argc)
	}
	return symbol(argv[:argc])
}

func ellKeywordP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("keyword?", "1", argc)
}

func ellTypeP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("type?", "1", argc)
}

func ellTypeName(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return typeName(argv[0])
		}
		return False, nil
	}
	return ArgcError("type-name", "1", argc)
}

/*
func ellKeyword(argv []AnyType, argc int) (AnyType, error) {
	if argc < 1 {
		return ArgcError("symbol", "1+", argc)
	}
	return keyword(argv)
}
*/

func ellUnkeyword(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("unkeyword", "1", argc)
	}
	return unkeyword(argv[0])
}

func elStringTypeP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("string?", "1", argc)
}

func ellCharP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isChar(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("character?", "1", argc)
}

func ellFunctionP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isFunction(argv[0]) || isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("function?", "1", argc)
}

func elListTypeP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("list?", "1", argc)
}

func ellEmptyP(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("empty?", "1", argc)
	}
	if isEmpty(argv[0]) {
		return True, nil
	}
	return False, nil
}

func elStringType(argv []AnyType, argc int) (AnyType, error) {
	s := ""
	for i := 0; i < argc; i++ {
		s += argv[i].String()
	}
	return StringType(s), nil
}

func ellCar(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("car", "1", argc)
}

func ellCdr(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cdr", "1", argc)
}

func ellSetCarBang(argv []AnyType, argc int) (AnyType, error) {
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

func ellSetCdrBang(argv []AnyType, argc int) (AnyType, error) {
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

func ellCaar(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caar(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("caar", "1", argc)
}

func ellCadr(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cadr", "1", argc)
}

func ellCddr(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cddr", "1", argc)
}

func ellCadar(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(cdr(car(lst))), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cadar", "1", argc)
}

func ellCaddr(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("caddr", "1", argc)
}
func ellCdddr(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cdddr", "1", argc)
}

func ellCons(argv []AnyType, argc int) (AnyType, error) {
	if argc == 2 {
		switch lst := argv[1].(type) {
		case *ListType:
			return cons(argv[0], lst), nil
		default:
			return ArgTypeError("list", 2, lst)
		}
	}
	return ArgcError("cons", "2", argc)
}

func ellStructP(argv []AnyType, argc int) (AnyType, error) {
	if argc == 1 {
		if isStruct(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("struct?", "1", argc)
}

func ellGet(argv []AnyType, argc int) (AnyType, error) {
	if argc != 2 {
		return ArgcError("get", "2", argc)
	}
	return get(argv[0], argv[1])
}

func ellHasP(argv []AnyType, argc int) (AnyType, error) {
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

func ellPutBang(argv []AnyType, argc int) (AnyType, error) {
	if argc != 3 {
		return ArgcError("put!", "3", argc)
	}
	return put(argv[0], argv[1], argv[2])
}

func ellStructToList(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("struct->list", "1", argc)
	}
	return structToList(argv[0])
}

func ellJSON(argv []AnyType, argc int) (AnyType, error) {
	if argc != 1 {
		return ArgcError("json", "1", argc)
	}
	s, err := toJSON(argv[0])
	if err != nil {
		return nil, err
	}
	return StringType(s), nil
}
