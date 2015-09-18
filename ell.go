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
	defineMacro("do", ellDo) //scheme's do. I don't like it, will replace with... "for" of some sort
	//note clojure uses "do" instead of "begin". I rather prefer that.
	defineMacro("cond", ellCond)
	defineMacro("quasiquote", ellQuasiquote)

	define("null", Null)
	define("true", True)
	define("false", False)

	define("apply", Apply)

	defineFunction("version", ellVersion)

	defineFunction("defined?", ellDefinedP)

	defineFunction("file-contents", ellFileContents)
	defineFunction("open-input-string", ellOpenInputString)
	defineFunction("open-input-file", ellOpenInputFile)
	defineFunction("read", ellRead)
	defineFunction("close-input", ellCloseInput)

	defineFunction("macroexpand", ellMacroexpand)
	defineFunction("type", ellType)
	defineFunction("value", ellValue)
	defineFunction("instance", ellInstance)
	defineFunction("normalize-keyword-args", ellNormalizeKeywordArgs)

	defineFunction("type?", ellTypeP)
	defineFunction("type-name", ellTypeName)

	defineFunction("struct", ellStruct)
	defineFunction("equal?", ellEq)
	defineFunction("identical?", ellIdenticalP)
	defineFunction("not", ellNot)

	defineFunction("boolean?", ellBooleanP)
	defineFunction("null?", ellNullP)
	defineFunction("symbol?", elSymbolTypeP)
	defineFunction("symbol", elSymbolType)

	defineFunction("keyword?", ellKeywordP)
	defineFunction("string?", elStringTypeP)
	defineFunction("char?", ellCharP)
	defineFunction("function?", ellFunctionP)
	defineFunction("eof?", ellFunctionP)

	defineFunction("list?", elLListP)
	defineFunction("cons", ellCons)
	defineFunction("car", ellCar)
	defineFunction("cdr", ellCdr)
	defineFunction("list", elLList)
	defineFunction("concat", ellConcat)
	defineFunction("reverse", ellReverse)
	defineFunction("set-car!", ellSetCarBang) //mutate!
	defineFunction("set-cdr!", ellSetCdrBang) //mutate!

	defineFunction("array?", ellArrayP)
	defineFunction("array", ellArray)
	defineFunction("make-array", ellMakeArray)
	defineFunction("array-set!", ellArraySetBang) //mutate!
	defineFunction("array-ref", ellArrayRef)

	defineFunction("struct?", ellStructP)
	defineFunction("has?", ellHasP)
	defineFunction("get", ellGet)
	defineFunction("assoc", ellAssoc)
	defineFunction("dissoc", ellDissoc)
	defineFunction("put!", ellPutBang) //mutate!
	defineFunction("struct->list", ellStructToList)

	defineFunction("empty?", ellEmptyP)

	defineFunction("string", elStringType)
	defineFunction("display", ellDisplay)
	defineFunction("write", ellWrite)
	defineFunction("newline", ellNewline)
	defineFunction("print", ellPrint)
	defineFunction("println", ellPrintln)
	defineFunction("write-to-string", ellWriteToString)

	defineFunction("number?", elLNumberP) // either float or int
	defineFunction("int?", ellIntP)       //int only
	defineFunction("float?", ellFloatP)   //float only
	defineFunction("+", ellPlus)
	defineFunction("-", ellMinus)
	defineFunction("*", ellTimes)
	defineFunction("/", ellDiv)
	defineFunction("quotient", ellQuotient)
	defineFunction("remainder", ellRemainder)
	defineFunction("modulo", ellRemainder) //fix!

	defineFunction("=", ellNumeq)
	defineFunction("<=", ellLe)
	defineFunction(">=", ellGe)
	defineFunction(">", ellGt)
	defineFunction("<", ellLt)
	defineFunction("zero?", ellZeroP)
	defineFunction("number->string", elLNumberToString)
	defineFunction("string-length", elStringTypeLength)

	defineFunction("error", ellFatal)
	defineFunction("length", ellLength)
	defineFunction("json", ellJSON)

	err := loadModule("ell")
	if err != nil {
		fatal("*** ", err)
	}
}

//
//expanders - these only gets called from the macro expander itself, so we know the single arg is an *LList
//

func ellLetrec(argv []LAny, argc int) (LAny, error) {
	return expandLetrec(argv[0])
}

func ellLet(argv []LAny, argc int) (LAny, error) {
	return expandLet(argv[0])
}

func ellDo(argv []LAny, argc int) (LAny, error) {
	return expandDo(argv[0])
}

func ellCond(argv []LAny, argc int) (LAny, error) {
	return expandCond(argv[0])
}

func ellQuasiquote(argv []LAny, argc int) (LAny, error) {
	return expandQuasiquote(argv[0])
}

// functions

func ellVersion(argv []LAny, argc int) (LAny, error) {
	return StringType(Version), nil
}

func ellDefinedP(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("defined?", "1", argc)
	}
	if !isSymbol(argv[0]) {
		return ArgTypeError("symbol", 1, argv[0])
	}
	if isDefined(argv[0]) {
		return True, nil
	}
	return False, nil
}

func ellFileContents(argv []LAny, argc int) (LAny, error) {
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

func ellOpenInputString(argv []LAny, argc int) (LAny, error) {
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

func ellOpenInputFile(argv []LAny, argc int) (LAny, error) {
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

func ellRead(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("read", "1", argc)
	}
	return readInput(argv[0])
}
func ellCloseInput(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("read", "1", argc)
	}
	return nil, closeInput(argv[0])
}

func ellMacroexpand(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("macroexpand", "1", argc)
	}
	return macroexpand(argv[0])
}

func ellType(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("type", "1", argc)
	}
	return argv[0].Type(), nil
}

func ellValue(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("value", "1", argc)
	}
	return argv[0].Value(), nil
}

func ellInstance(argv []LAny, argc int) (LAny, error) {
	if argc != 2 {
		return ArgcError("instance", "2", argc)
	}
	return instance(argv[0], argv[1])
}

func ellNormalizeKeywordArgs(argv []LAny, argc int) (LAny, error) {
	//(normalize-keyword-args '(x: 23) '(x: y:) -> (x:)
	//(normalize-keyword-args '(x: 23 z: 100) '(x: y:) -> error("bad keyword z: in argument list")
	if argc < 1 {
		return ArgcError("normalizeKeywordArgs", "1+", argc)
	}
	if args, ok := argv[0].(*LList); ok {
		return normalizeKeywordArgs(args, argv[1:argc])
	}
	return ArgTypeError("list", 1, argv[0])
}

func ellStruct(argv []LAny, argc int) (LAny, error) {
	return newStruct(argv[:argc])
}

func ellIdenticalP(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("identical?", "2", argc)
}

func ellEq(argv []LAny, argc int) (LAny, error) {
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

func ellNumeq(argv []LAny, argc int) (LAny, error) {
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

func ellDisplay(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		//todo: add the optional output argument like scheme
		return ArgcError("display", "1", argc)
	}
	fmt.Printf("%v", argv[0])
	return Null, nil
}

func ellWrite(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		//todo: add the optional output argument like scheme
		return ArgcError("write", "1", argc)
	}
	fmt.Printf("%v", write(argv[0]))
	return Null, nil
}

func ellNewline(argv []LAny, argc int) (LAny, error) {
	if argc != 0 {
		//todo: add the optional output argument like scheme
		return ArgcError("newline", "0", argc)
	}
	fmt.Printf("\n")
	return Null, nil
}

func ellFatal(argv []LAny, argc int) (LAny, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}

func ellWriteToString(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("write-to-string", "1", argc)
	}
	s := write(argv[0])
	return StringType(s), nil

}

func ellPrint(argv []LAny, argc int) (LAny, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return Null, nil
}

func ellPrintln(argv []LAny, argc int) (LAny, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return Null, nil
}

func ellConcat(argv []LAny, argc int) (LAny, error) {
	result := EmptyList
	tail := result
	for i := 0; i < argc; i++ {
		o := argv[i]
		switch lst := o.(type) {
		case *LList:
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

func ellReverse(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("reverse", "1", argc)
	}
	o := argv[0]
	switch lst := o.(type) {
	case *LList:
		return reverse(lst), nil
	default:
		return ArgTypeError("list", 1, o)
	}
}

func elLList(argv []LAny, argc int) (LAny, error) {
	p := EmptyList
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func elLNumberP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isNumber(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("number?", "1", argc)
}

func ellIntP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isInt(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("int?", "1", argc)
}

func ellFloatP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isFloat(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("float?", "1", argc)
}

func ellQuotient(argv []LAny, argc int) (LAny, error) {
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
		return LNumber(n1 / n2), nil
	}
	return ArgcError("quotient", "2", argc)
}

func ellRemainder(argv []LAny, argc int) (LAny, error) {
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
		return LNumber(n1 % n2), nil
	}
	return ArgcError("remainder", "2", argc)
}

func ellPlus(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []LAny, argc int) (LAny, error) {
	return minus(argv, argc)
}

func ellTimes(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellDiv(argv []LAny, argc int) (LAny, error) {
	return div(argv, argc)
}

func ellArray(argv []LAny, argc int) (LAny, error) {
	return array(argv...), nil
}

func ellMakeArray(argv []LAny, argc int) (LAny, error) {
	if argc > 0 {
		initVal := LAny(Null)
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

func ellArrayP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isArray(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("array?", "1", argc)
}

func ellArraySetBang(argv []LAny, argc int) (LAny, error) {
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

func ellArrayRef(argv []LAny, argc int) (LAny, error) {
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

func ellGe(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError(">=", "2", argc)
}

func ellLe(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError("<=", "2", argc)
}

func ellGt(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError(">", "2", argc)
}

func ellLt(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return ArgcError("<", "2", argc)
}

func ellZeroP(argv []LAny, argc int) (LAny, error) {
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

func elLNumberToString(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("number->string", "1", argc)
	}
	if !isNumber(argv[0]) {
		return ArgTypeError("number", 1, argv[0])
	}
	return StringType(argv[0].String()), nil
}

func elStringTypeLength(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("string-length", "1", argc)
	}
	if !isString(argv[0]) {
		return ArgTypeError("string", 1, argv[0])
	}
	i := length(argv[0])
	return LNumber(int64(i)), nil
}

func ellLength(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		return LNumber(int64(length(argv[0]))), nil
	}
	return ArgcError("length", "1", argc)
}

func ellNot(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if argv[0] == False {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("not", "1", argc)
}

func ellNullP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if argv[0] == Null {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("nil?", "1", argc)
}

func ellBooleanP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isBoolean(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("boolean?", "1", argc)
}

func elSymbolTypeP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isSymbol(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("symbol?", "1", argc)
}

func elSymbolType(argv []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return ArgcError("symbol", "1+", argc)
	}
	return symbol(argv[:argc])
}

func ellKeywordP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("keyword?", "1", argc)
}

func ellTypeP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("type?", "1", argc)
}

func ellTypeName(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isType(argv[0]) {
			return typeName(argv[0])
		}
		return False, nil
	}
	return ArgcError("type-name", "1", argc)
}

func elStringTypeP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isString(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("string?", "1", argc)
}

func ellCharP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isChar(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("character?", "1", argc)
}

func ellFunctionP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isFunction(argv[0]) || isKeyword(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("function?", "1", argc)
}

func elLListP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isList(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("list?", "1", argc)
}

func ellEmptyP(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("empty?", "1", argc)
	}
	if isEmpty(argv[0]) {
		return True, nil
	}
	return False, nil
}

func elStringType(argv []LAny, argc int) (LAny, error) {
	s := ""
	for i := 0; i < argc; i++ {
		s += argv[i].String()
	}
	return StringType(s), nil
}

func ellCar(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("car", "1", argc)
}

func ellCdr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cdr", "1", argc)
}

func ellSetCarBang(argv []LAny, argc int) (LAny, error) {
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

func ellSetCdrBang(argv []LAny, argc int) (LAny, error) {
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

func ellCaar(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caar(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("caar", "1", argc)
}

func ellCadr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cadr", "1", argc)
}

func ellCddr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cddr", "1", argc)
}

func ellCadar(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(cdr(car(lst))), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cadar", "1", argc)
}

func ellCaddr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return caddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("caddr", "1", argc)
}
func ellCdddr(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdddr(lst), nil
		}
		return ArgTypeError("list", 1, lst)
	}
	return ArgcError("cdddr", "1", argc)
}

func ellCons(argv []LAny, argc int) (LAny, error) {
	if argc == 2 {
		switch lst := argv[1].(type) {
		case *LList:
			return cons(argv[0], lst), nil
		default:
			return ArgTypeError("list", 2, lst)
		}
	}
	return ArgcError("cons", "2", argc)
}

func ellStructP(argv []LAny, argc int) (LAny, error) {
	if argc == 1 {
		if isStruct(argv[0]) {
			return True, nil
		}
		return False, nil
	}
	return ArgcError("struct?", "1", argc)
}

func ellGet(argv []LAny, argc int) (LAny, error) {
	if argc != 2 {
		return ArgcError("get", "2", argc)
	}
	return get(argv[0], argv[1])
}

func ellHasP(argv []LAny, argc int) (LAny, error) {
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

func ellPutBang(argv []LAny, argc int) (LAny, error) {
	if argc != 3 {
		return ArgcError("put!", "3", argc)
	}
	return put(argv[0], argv[1], argv[2])
}

func ellAssoc(argv []LAny, argc int) (LAny, error) {
	if argc != 3 {
		return ArgcError("assoc", "3", argc)
	}
	return assoc(argv[0], argv[1], argv[2])
}

func ellDissoc(argv []LAny, argc int) (LAny, error) {
	if argc != 3 {
		return ArgcError("dissoc", "2", argc)
	}
	return dissoc(argv[0], argv[1])
}

func ellStructToList(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("struct->list", "1", argc)
	}
	return structToList(argv[0])
}

func ellJSON(argv []LAny, argc int) (LAny, error) {
	if argc != 1 {
		return ArgcError("json", "1", argc)
	}
	s, err := toJSON(argv[0])
	if err != nil {
		return nil, err
	}
	return StringType(s), nil
}
