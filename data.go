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

import (
	"bytes"
	"fmt"
	"strconv"
)

// LAny type is a union of all possible primitive types. Which fields are used depends on the variant
// the variant is a type object i.e. intern("<string>")
type LAny struct {
	ltype     *LAny      // i.e. <string>
	boolean   bool       // <boolean>
	character rune       // <character>
	ival      int        // <number>, <symbol>
	fval      float64    //<number>
	text      string     // <string>, <symbol>, <keyword>, <type>
	car       *LAny      // non-nil for instances and <list>
	cdr       *LAny      // non-nil for <list>, nil for everything else
	elements  []*LAny    // <vector>, <struct>
	function  *LFunction // <function>
	code      *LCode     //<code>
	port      *LPort     // <port>
}

func identical(o1 *LAny, o2 *LAny) bool {
	return o1 == o2
}

func (a *LAny) String() string {
	if a == Null {
		return "null"
	} else if a == EOF {
		return "#[eof]"
	} else {
		switch a.ltype {
		case typeBoolean:
			return strconv.FormatBool(a.boolean)
		case typeCharacter:
			return string([]rune{a.character})
		case typeNumber:
			return strconv.FormatFloat(a.fval, 'f', -1, 64)
		case typeString, typeSymbol, typeKeyword, typeType:
			return a.text
		case typeList:
			return listToString(a)
		case typeVector:
			return vectorToString(a)
		case typeStruct:
			return structToString(a)
		case typeFunction:
			return a.function.String()
		case typePort:
			return a.port.String()
		default:
			return "#" + a.ltype.text + write(a.car)
		}
	}
}

var typeType *LAny    // bootstrapped in initSymbolTable => intern("<type>")
var typeKeyword *LAny // bootstrapped in initSymbolTable => intern("<keyword>")
var typeSymbol *LAny  // bootstrapped in initSymbolTable = intern("<symbol>")

var typeNull = intern("<null>")
var typeEOF = intern("<eof>")
var typeBoolean = intern("<boolean>")
var typeCharacter = intern("<character>")
var typeNumber = intern("<number>")
var typeString = intern("<string>")
var typeList = intern("<list>")
var typeVector = intern("<vector>")
var typeStruct = intern("<struct>")
var typeFunction = intern("<function>")
var typeCode = intern("<code>")
var typePort = intern("<port>")

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList. It is a singleton.
var Null = &LAny{ltype: typeNull}

func isNull(obj *LAny) bool {
	return obj == Null
}

// EOF is Ell's singleton EOF object
var EOF = &LAny{ltype: typeEOF}

func isEOF(obj *LAny) bool {
	return obj == EOF
}

// True is the singleton boolean true value
var True = &LAny{ltype: typeBoolean, boolean: true}

// False is the singleton boolean false value
var False = &LAny{ltype: typeBoolean, boolean: false}

func isBoolean(obj *LAny) bool {
	return obj.ltype == typeBoolean
}

func isCharacter(obj *LAny) bool {
	return obj.ltype == typeCharacter
}
func isNumber(obj *LAny) bool {
	return obj.ltype == typeNumber
}
func isString(obj *LAny) bool {
	return obj.ltype == typeString
}
func isList(obj *LAny) bool {
	return obj.ltype == typeList
}
func isVector(obj *LAny) bool {
	return obj.ltype == typeVector
}
func isStruct(obj *LAny) bool {
	return obj.ltype == typeStruct
}
func isFunction(obj *LAny) bool {
	return obj.ltype == typeFunction
}
func isCode(obj *LAny) bool {
	return obj.ltype == typeCode
}
func isSymbol(obj *LAny) bool {
	return obj.ltype == typeSymbol
}
func isKeyword(obj *LAny) bool {
	return obj.ltype == typeKeyword
}
func isType(obj *LAny) bool {
	return obj.ltype == typeType
}
func isPort(obj *LAny) bool {
	return obj.ltype == typePort
}

//instances have arbitrary variant symbols, all we can check is that the instanceValue is set
func isInstance(obj *LAny) bool {
	return obj.car != nil && obj.cdr == nil
}

func equal(o1 *LAny, o2 *LAny) bool {
	if o1 == o2 {
		return true
	}
	if o1.ltype != o2.ltype {
		return false
	}
	switch o1.ltype {
	case typeBoolean:
		return o1.boolean == o2.boolean
	case typeCharacter:
		return o1.character == o2.character
	case typeNumber:
		return o1.fval == o2.fval
	case typeString:
		return o1.text == o2.text
	case typeList:
		return listEqual(o1, o2)
	case typeVector:
		return vectorEqual(o1, o2)
	case typeStruct:
		return structEqual(o1, o2)
	case typeSymbol, typeKeyword, typeType:
		return o1 == o2
	case typeNull, typeEOF:
		return true // singletons
	default:
		return false
	}
}

func isPrimitiveType(tag *LAny) bool {
	switch tag {
	case typeNull, typeEOF, typeBoolean, typeCharacter, typeNumber, typeString, typeList, typeVector, typeStruct:
		return true
	case typeSymbol, typeKeyword, typeType, typeFunction, typePort:
		return true
	default:
		return false
	}
}

func instance(tag *LAny, val *LAny) (*LAny, error) {
	if !isType(tag) {
		return nil, TypeError(typeType, tag)
	}
	if isPrimitiveType(tag) {
		return val, nil
	}
	return &LAny{ltype: tag, car: val}, nil
}

func value(obj *LAny) *LAny {
	if obj.cdr == nil && obj.car != nil {
		return obj.car
	}
	return obj
}

//
// Error - creates a new Error from the arguments
//
func Error(arg1 interface{}, args ...interface{}) error {
	var buf bytes.Buffer
	if l, ok := arg1.(*LAny); ok {
		buf.WriteString(fmt.Sprintf("%v", write(l)))
	} else {
		buf.WriteString(fmt.Sprintf("%v", arg1))
	}
	for _, o := range args {
		if l, ok := o.(*LAny); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	err := GenericError{buf.String()}
	return &err
}

// TypeError - an error indicating expected and actual value for a type mismatch
func TypeError(typeSym *LAny, obj *LAny) error {
	return Error("Type error: expected ", typeSym, ", got ", obj)
}

// GenericError - most Ell errors are one of these
type GenericError struct {
	msg string
}

func (e *GenericError) Error() string {
	return e.msg
}

func (e *GenericError) String() string {
	return fmt.Sprintf("<Error: %s>", e.msg)
}
