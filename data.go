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
	//"unsafe"
)

// LOB type is the Ell object: a union of all possible primitive types. Which fields are used depends on the variant
// the variant is a type object i.e. intern("<string>")
type LOB struct {
	variant      *LOB   // i.e. <string>
	code         *Code  // closure, code
	frame        *Frame // closure, continuation
	primitive    *Primitive
	continuation *Continuation
	car          *LOB    // non-nil for instances and <list>
	cdr          *LOB    // non-nil for <list>, nil for everything else
	ival         int64   // <boolean>, <character>
	fval         float64 //<number>
	text         string  // <string>, <symbol>, <keyword>, <type>
	elements     []*LOB  // <vector>, <struct>
	padTo120     int64
}

func newLOB(variant *LOB) *LOB {
	lob := new(LOB)
	//	println("lob size: ",unsafe.Sizeof(*lob))
	lob.variant = variant
	return lob
}

func identical(o1 *LOB, o2 *LOB) bool {
	return o1 == o2
}

func (lob *LOB) String() string {
	if lob == Null {
		return "null"
	}
	switch lob.variant {
	case typeBoolean:
		if lob.ival == 0 {
			return "false"
		}
		return "true"
	case typeCharacter:
		return string([]rune{rune(lob.ival)})
	case typeNumber:
		return strconv.FormatFloat(lob.fval, 'f', -1, 64)
	case typeString, typeSymbol, typeKeyword, typeType:
		return lob.text
	case typeList:
		return listToString(lob)
	case typeVector:
		return vectorToString(lob)
	case typeStruct:
		return structToString(lob)
	case typeFunction:
		return functionToString(lob)
	case typeCode:
		return lob.code.String()
	case typeError:
		return "#<error>" + write(lob.car)
	default:
		return "#" + lob.variant.text + write(lob.car)
	}
}

var typeType *LOB    // bootstrapped in initSymbolTable => intern("<type>")
var typeKeyword *LOB // bootstrapped in initSymbolTable => intern("<keyword>")
var typeSymbol *LOB  // bootstrapped in initSymbolTable = intern("<symbol>")

var typeNull = intern("<null>")
var typeBoolean = intern("<boolean>")
var typeCharacter = intern("<character>")
var typeNumber = intern("<number>")
var typeString = intern("<string>")
var typeList = intern("<list>")
var typeVector = intern("<vector>")
var typeStruct = intern("<struct>")
var typeFunction = intern("<function>")
var typeCode = intern("<code>")
var typeError = intern("<error>")

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList. It is a singleton.
var Null = &LOB{variant: typeNull}

func isNull(obj *LOB) bool {
	return obj == Null
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
		return numberEqual(o1.fval, o2.fval)
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
	case typeNull:
		return true // singleton
	default:
		o1a := value(o1)
		if o1a != o1 {
			o2a := value(o2)
			return equal(o1a, o2a)
		}
		return false
	}
}

func isPrimitiveType(tag *LOB) bool {
	switch tag {
	case typeNull, typeBoolean, typeCharacter, typeNumber, typeString, typeList, typeVector, typeStruct:
		return true
	case typeSymbol, typeKeyword, typeType, typeFunction:
		return true
	default:
		return false
	}
}

func instance(tag *LOB, val *LOB) (*LOB, error) {
	if !isType(tag) {
		return nil, Error(ArgumentErrorKey, typeType.text, tag)
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
// Error - creates a new Error from the arguments. The first is an actual Ell object, the rest are interpreted as/converted to strings
//
func Error(errkey *LOB, args ...interface{}) error {
	var buf bytes.Buffer
	for _, o := range args {
		if l, ok := o.(*LOB); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	if errkey.variant != typeKeyword {
		errkey = ErrorKey
	}
	return newError(errkey, newString(buf.String()))
}

func newError(elements ...*LOB) *LOB {
	data := vector(elements...)
	return &LOB{variant: typeError, car: data}
}

func theError(o interface{}) (*LOB, bool) {
	if o == nil {
		return nil, false
	}
	if err, ok := o.(*LOB); ok {
		if err.variant == typeError {
			return err, true
		}
	}
	return nil, false

}

func isError(o interface{}) bool {
	_, ok := theError(o)
	return ok
}

func errorData(err *LOB) *LOB {
	return err.car
}

// Error
func (lob *LOB) Error() string {
	if lob.variant == typeError {
		s := lob.car.String()
		if lob.text != "" {
			s += " [in " + lob.text + "]"
		}
		return s
	}
	return lob.String()
}

// ErrorKey - used to generic errors
var ErrorKey = intern("error:")

// ArgumentErrorKey
var ArgumentErrorKey = intern("argument-error:")

// SyntaxErrorKey
var SyntaxErrorKey = intern("syntax-error:")

// MacroErrorKey
var MacroErrorKey = intern("macro-error:")

// IOErrorKey
var IOErrorKey = intern("io-error:")

// InterruptKey
var InterruptKey = intern("interrupt:")
