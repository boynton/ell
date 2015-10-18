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
	Type      *LOB          // i.e. <string>
	code         *Code         // non-nil for closure, code
	frame        *Frame        // non-nil for closure, continuation
	primitive    *Primitive    // non-nil for primitives
	continuation *Continuation // non-nil for continuation
	car          *LOB          // non-nil for instances and lists
	cdr          *LOB          // non-nil for slists, nil for everything else
	bindings     map[Key]*LOB  // non-nil for struct
	elements     []*LOB        // non-nil for vector
	fval         float64       // number
	text         string        // string, symbol, keyword, type, blob, channel
	Value         interface{}   // the rest of the data for more complex things
}

func BoolValue(obj *LOB) bool {
	if obj == True {
		return true
	}
	return false
}

func RuneValue(obj *LOB) rune {
	return rune(obj.fval)
}

func StringValue(obj *LOB) string {
	return obj.text
}

func BlobValue(obj *LOB) []byte {
	return []byte(obj.text)
}


type Channel struct {
	bufsize        int
	channel      chan *LOB     // non-nil for channels
}

func ChannelValue (obj *LOB) *Channel {
	if obj.Value == nil {
		return nil
	}
	v, _ := obj.Value.(*Channel)
	return v
}

func NewObject(variant *LOB, value interface{}) *LOB {
	lob := newLOB(variant)
	lob.Value = value
	return lob
}

func newLOB(variant *LOB) *LOB {
	lob := new(LOB)
	lob.Type = variant
	return lob
}

func identical(o1 *LOB, o2 *LOB) bool {
	return o1 == o2
}

func (lob *LOB) String() string {
	switch lob.Type {
	case NullType:
		return "null"
	case BooleanType:
		if lob == True {
			return "true"
		}
		return "false"
	case CharacterType:
		return string([]rune{rune(lob.fval)})
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
		ch, _ := lob.Value.(*Channel)
		if ch.bufsize > 0 {
			s += fmt.Sprintf(" [%d]", ch.bufsize)
		}
		if ch.channel == nil {
			s += " CLOSED"
		}
		return s + "]"
	default:
		return "#" + lob.Type.text + write(lob.car)
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
	switch o.Type {
	case StringType, SymbolType, KeywordType, TypeType:
		return true
	}
	return false
}

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList. It is a singleton.
var Null = &LOB{Type: NullType}

func isNull(obj *LOB) bool {
	return obj == Null
}

// True is the singleton boolean true value
var True = &LOB{Type: BooleanType, fval: 1}

// False is the singleton boolean false value
var False = &LOB{Type: BooleanType, fval: 0}

func isBoolean(obj *LOB) bool {
	return obj.Type == BooleanType
}

func isCharacter(obj *LOB) bool {
	return obj.Type == CharacterType
}
func isNumber(obj *LOB) bool {
	return obj.Type == NumberType
}
func isString(obj *LOB) bool {
	return obj.Type == StringType
}
func isList(obj *LOB) bool {
	return obj.Type == ListType
}
func isVector(obj *LOB) bool {
	return obj.Type == VectorType
}
func isStruct(obj *LOB) bool {
	return obj.Type == StructType
}
func isFunction(obj *LOB) bool {
	return obj.Type == FunctionType
}
func isCode(obj *LOB) bool {
	return obj.Type == CodeType
}
func isSymbol(obj *LOB) bool {
	return obj.Type == SymbolType
}
func isKeyword(obj *LOB) bool {
	return obj.Type == KeywordType
}
func isType(obj *LOB) bool {
	return obj.Type == TypeType
}

//instances have arbitrary Type symbols, all we can check is that the instanceValue is set
func isInstance(obj *LOB) bool {
	return obj.car != nil && obj.cdr == nil
}

func equal(o1 *LOB, o2 *LOB) bool {
	if o1 == o2 {
		return true
	}
	if o1.Type != o2.Type {
		return false
	}
	switch o1.Type {
	case BooleanType, CharacterType:
		return int(o1.fval) == int(o2.fval)
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
	if errkey.Type != KeywordType {
		errkey = ErrorKey
	}
	return newError(errkey, newString(buf.String()))
}

func newError(elements ...*LOB) *LOB {
	data := vector(elements...)
	return &LOB{Type: ErrorType, car: data}
}

func theError(o interface{}) (*LOB, bool) {
	if o == nil {
		return nil, false
	}
	if err, ok := o.(*LOB); ok {
		if err.Type == ErrorType {
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
	if lob.Type == ErrorType {
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
	lob.Value = &Channel{bufsize: bufsize, channel: make(chan *LOB, bufsize)}
	lob.text = name
	return lob
}
