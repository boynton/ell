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

// LOB type is the Ell object: a union of all possible primitive types. Which fields are used depends on the variant
// the variant is a type object i.e. intern("<string>")
type LOB struct {
	variant  *LOB       // i.e. <string>
	function *LFunction // <function>
	car      *LOB       // non-nil for instances and <list>
	cdr      *LOB       // non-nil for <list>, nil for everything else
	code     *LCode     //<code>
	port     *LPort     // <port>
	ival     int64      // <boolean>, <character>
	fval     float64    //<number>
	text     string     // <string>, <symbol>, <keyword>, <type>
	elements []*LOB     // <vector>, <struct>
}

func newLOB(variant *LOB) *LOB {
	lob := new(LOB)
	lob.variant = variant
	return lob
}

func identical(o1 *LOB, o2 *LOB) bool {
	return o1 == o2
}

func (a *LOB) String() string {
	if a == Null {
		return "null"
	} else if a == EOF {
		return "#[eof]"
	} else {
		switch a.variant {
		case typeBoolean:
			if a.ival == 0 {
				return "false"
			}
			return "true"
		case typeCharacter:
			return string([]rune{rune(a.ival)})
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
			return "#" + a.variant.text + write(a.car)
		}
	}
}

var typeType *LOB    // bootstrapped in initSymbolTable => intern("<type>")
var typeKeyword *LOB // bootstrapped in initSymbolTable => intern("<keyword>")
var typeSymbol *LOB  // bootstrapped in initSymbolTable = intern("<symbol>")

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
var Null = &LOB{variant: typeNull}

func isNull(obj *LOB) bool {
	return obj == Null
}

// EOF is Ell's singleton EOF object
var EOF = &LOB{variant: typeEOF}

func isEOF(obj *LOB) bool {
	return obj == EOF
}

// True is the singleton boolean true value
var True = &LOB{variant: typeBoolean, ival: 1}

// False is the singleton boolean false value
var False = &LOB{variant: typeBoolean, ival: 0}

func isBoolean(obj *LOB) bool {
	return obj.variant == typeBoolean
}

func isCharacter(obj *LOB) bool {
	return obj.variant == typeCharacter
}
func isNumber(obj *LOB) bool {
	return obj.variant == typeNumber
}
func isString(obj *LOB) bool {
	return obj.variant == typeString
}
func isList(obj *LOB) bool {
	return obj.variant == typeList
}
func isVector(obj *LOB) bool {
	return obj.variant == typeVector
}
func isStruct(obj *LOB) bool {
	return obj.variant == typeStruct
}
func isFunction(obj *LOB) bool {
	return obj.variant == typeFunction
}
func isCode(obj *LOB) bool {
	return obj.variant == typeCode
}
func isSymbol(obj *LOB) bool {
	return obj.variant == typeSymbol
}
func isKeyword(obj *LOB) bool {
	return obj.variant == typeKeyword
}
func isType(obj *LOB) bool {
	return obj.variant == typeType
}
func isPort(obj *LOB) bool {
	return obj.variant == typePort
}

//instances have arbitrary variant symbols, all we can check is that the instanceValue is set
func isInstance(obj *LOB) bool {
	return obj.car != nil && obj.cdr == nil
}

func equal(o1 *LOB, o2 *LOB) bool {
	if o1 == o2 {
		return true
	}
	if o1.variant != o2.variant {
		return false
	}
	switch o1.variant {
	case typeBoolean, typeCharacter:
		return o1.ival == o2.ival
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

func isPrimitiveType(tag *LOB) bool {
	switch tag {
	case typeNull, typeEOF, typeBoolean, typeCharacter, typeNumber, typeString, typeList, typeVector, typeStruct:
		return true
	case typeSymbol, typeKeyword, typeType, typeFunction, typePort:
		return true
	default:
		return false
	}
}

func instance(tag *LOB, val *LOB) (*LOB, error) {
	if !isType(tag) {
		return nil, TypeError(typeType, tag)
	}
	if isPrimitiveType(tag) {
		return val, nil
	}
	result := newLOB(tag)
	result.car = val
	return result, nil
}

func value(obj *LOB) *LOB {
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
	if l, ok := arg1.(*LOB); ok {
		buf.WriteString(fmt.Sprintf("%v", write(l)))
	} else {
		buf.WriteString(fmt.Sprintf("%v", arg1))
	}
	for _, o := range args {
		if l, ok := o.(*LOB); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	err := GenericError{buf.String()}
	return &err
}

// TypeError - an error indicating expected and actual value for a type mismatch
func TypeError(typeSym *LOB, obj *LOB) error {
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
