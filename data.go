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
	variant      *LOB          // i.e. <string>
	code         *Code         // non-nil for closure, code
	frame        *Frame        // non-nil for closure, continuation
	primitive    *Primitive    // non-nil for primitives
	continuation *Continuation // non-nil for continuation
	car          *LOB          // non-nil for instances and lists
	cdr          *LOB          // non-nil for slists, nil for everything else
	channel      chan *LOB     // non-nil for channels
	text         string        // string, symbol, keyword, type, blob, channel
	elements     []*LOB        // non-nil for vector
	fval         float64       // number
	bindings     map[Key]*LOB  // non-nil for struct
	ival         int64         // boolean, character, channel
}

func newLOB(variant *LOB) *LOB {
	lob := new(LOB)
	lob.variant = variant
	return lob
}

func identical(o1 *LOB, o2 *LOB) bool {
	return o1 == o2
}

func (lob *LOB) String() string {
	switch lob.variant {
	case NullType:
		return "null"
	case BooleanType:
		if lob == True {
			return "true"
		}
		return "false"
	case CharacterType:
		return string([]rune{rune(lob.ival)})
	case NumberType:
		return strconv.FormatFloat(lob.fval, 'f', -1, 64)
	case BlobType:
		return fmt.Sprintf("#[blob %d bytes]", len(lob.text))
	case StringType, SymbolType, KeywordType, TypeType:
		return lob.text
	case ListType:
		return listToString(lob)
	case VectorType:
		return vectorToString(lob)
	case StructType:
		return structToString(lob)
	case FunctionType:
		return functionToString(lob)
	case CodeType:
		return lob.code.String()
	case ErrorType:
		return "#<error>" + write(lob.car)
	case ChannelType:
		s := "#[channel"
		if lob.text != "" {
			s += " " + lob.text
		}
		if lob.ival > 0 {
			s += fmt.Sprintf(" [%d]", lob.ival)
		}
		if lob.channel == nil {
			s += " CLOSED"
		}
		return s + "]"
	default:
		return "#" + lob.variant.text + write(lob.car)
	}
}

// TypeType is the metatype, the type of all types
var TypeType *LOB // bootstrapped in initSymbolTable => intern("<type>")

// KeywordType is the type of all keywords
var KeywordType *LOB // bootstrapped in initSymbolTable => intern("<keyword>")

// SymbolType is the type of all symbols
var SymbolType *LOB // bootstrapped in initSymbolTable = intern("<symbol>")

// NullType the type of the null object
var NullType = intern("<null>")

// BooleanType is the type of true and false
var BooleanType = intern("<boolean>")

// CharacterType is the type of all characters
var CharacterType = intern("<character>")

// NumberType is the type of all numbers
var NumberType = intern("<number>")

// StringType is the type of all strings
var StringType = intern("<string>")

// BlobType is the type of all bytearrays
var BlobType = intern("<blob>")

// ListType is the type of all lists
var ListType = intern("<list>")

// VectorType is the type of all vectors
var VectorType = intern("<vector>")

// VectorType is the type of all structs
var StructType = intern("<struct>")

// FunctionType is the type of all functions
var FunctionType = intern("<function>")

// CodeType is the type of compiled code
var CodeType = intern("<code>")

// ErrorType is the type of all errors
var ErrorType = intern("<error>")

// ChannelType is the type for channels
var ChannelType = intern("<channel>")

// AnyType is a pseudo type specifier indicating any type
var AnyType = intern("<any>")

func isTextual(o *LOB) bool {
	switch o.variant {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList. It is a singleton.
var Null = &LOB{variant: NullType}

func isNull(obj *LOB) bool {
	return obj == Null
}

// True is the singleton boolean true value
var True = &LOB{variant: BooleanType, ival: 1}

// False is the singleton boolean false value
var False = &LOB{variant: BooleanType, ival: 0}

func isBoolean(obj *LOB) bool {
	return obj.variant == BooleanType
}

func isCharacter(obj *LOB) bool {
	return obj.variant == CharacterType
}
func isNumber(obj *LOB) bool {
	return obj.variant == NumberType
}
func isString(obj *LOB) bool {
	return obj.variant == StringType
}
func isList(obj *LOB) bool {
	return obj.variant == ListType
}
func isVector(obj *LOB) bool {
	return obj.variant == VectorType
}
func isStruct(obj *LOB) bool {
	return obj.variant == StructType
}
func isFunction(obj *LOB) bool {
	return obj.variant == FunctionType
}
func isCode(obj *LOB) bool {
	return obj.variant == CodeType
}
func isSymbol(obj *LOB) bool {
	return obj.variant == SymbolType
}
func isKeyword(obj *LOB) bool {
	return obj.variant == KeywordType
}
func isType(obj *LOB) bool {
	return obj.variant == TypeType
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
	case BooleanType, CharacterType:
		return o1.ival == o2.ival
	case NumberType:
		return numberEqual(o1.fval, o2.fval)
	case StringType:
		return o1.text == o2.text
	case ListType:
		return listEqual(o1, o2)
	case VectorType:
		return vectorEqual(o1, o2)
	case StructType:
		return structEqual(o1, o2)
	case SymbolType, KeywordType, TypeType:
		return o1 == o2
	case NullType:
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
	case NullType, BooleanType, CharacterType, NumberType, StringType, ListType, VectorType, StructType:
		return true
	case SymbolType, KeywordType, TypeType, FunctionType:
		return true
	default:
		return false
	}
}

func instance(tag *LOB, val *LOB) (*LOB, error) {
	if !isType(tag) {
		return nil, Error(ArgumentErrorKey, TypeType.text, tag)
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
	if errkey.variant != KeywordType {
		errkey = ErrorKey
	}
	return newError(errkey, newString(buf.String()))
}

func newError(elements ...*LOB) *LOB {
	data := vector(elements...)
	return &LOB{variant: ErrorType, car: data}
}

func theError(o interface{}) (*LOB, bool) {
	if o == nil {
		return nil, false
	}
	if err, ok := o.(*LOB); ok {
		if err.variant == ErrorType {
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
	if lob.variant == ErrorType {
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

// HttpErrorKey
var HTTPErrorKey = intern("http-error:")

// InterruptKey
var InterruptKey = intern("interrupt:")

// channels

func newChannel(bufsize int, name string) *LOB {
	lob := newLOB(ChannelType)
	lob.channel = make(chan *LOB, bufsize)
	lob.ival = int64(bufsize)
	lob.text = name
	return lob
}
