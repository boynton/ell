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
)

type LOB interface { // <any> - the interface all *Lxxx types must implement
	Type() LOB
	Value() LOB
	Equal(another LOB) bool
	String() string
}

func isIdentical(o1 LOB, o2 LOB) bool {
	return o1 == o2
}

func isEqual(o1 LOB, o2 LOB) bool {
	if o1 == o2 {
		return true
	}
	return o1.Equal(o2)
}

func value(obj LOB) LOB {
	return obj.Value()
}


// --- Null

type LNull struct {  // <null>
}

var NullType = intern("<null>")

// Null is Ell's version of nil. It means "nothing" and is not used for anything else (it is different than EmptyList)
var Null = LOB(&LNull{})

// Type returns the type of the object
func (*LNull) Type() LOB {
	return NullType
}

// Value returns the object itself for primitive types
func (*LNull) Value() LOB {
	return Null
}

// Equal returns true if the object is equal to the argument
func (*LNull) Equal(another LOB) bool {
	return another == Null
}

// String returns the string representation of the object
func (*LNull) String() string {
	return "null"
}

func isNull(obj LOB) bool {
	return obj == Null
}



// --- Boolean

var BooleanType = intern("<boolean>")

type LBoolean struct {
	value bool
}

// Type returns the type of the object
func (*LBoolean) Type() LOB {
	return BooleanType
}

// Value returns the object itself for primitive types
func (b *LBoolean) Value() LOB {
	return b
}

// Equal returns true if the object is equal to the argument
func (b *LBoolean) Equal(another LOB) bool {
	return another == b
}

// String returns the string representation of the object
func (b *LBoolean) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

// True is the singleton boolean true value
var True = &LBoolean{value: true}

// False is the singleton boolean false value
var False = &LBoolean{value: false}

func isBoolean(obj LOB) bool {
	_, ok := obj.(*LBoolean)
	return ok
}


// LOB type is the Ell object: a union of all possible primitive types. Which fields are used depends on the variant
// the variant is a type object i.e. intern("<string>")
/*
type LOB struct {
	variant  *LOB       // i.e. <string>
	function *LFunction // <function>
	car      *LOB       // non-nil for instances and <list>
	cdr      *LOB       // non-nil for <list>, nil for everything else
	code     *LCode     //<code>
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
*/

type LInstance struct {
	variant *LType
	value LOB
}

func (inst *LInstance) Type() LOB {
	return inst.variant
}

func (inst *LInstance) Value() LOB {
	return inst.value
}

func (inst *LInstance) Equal(another LOB) bool {
	if i2, ok := another.(*LInstance); ok {
		if isEqual(inst.variant, i2.variant) {
			return isEqual(inst.value, i2.value)
		}
	}
	return false
}

func (inst *LInstance) String() string {
	return "#" + inst.variant.text + write(inst.value)
}

//instances have arbitrary variant symbols, all we can check is that the instanceValue is set
func isInstance(obj LOB) bool {
	_, ok := obj.(*LInstance)
	return ok
}

func isPrimitiveType(variant *LType) bool {
	switch variant {
	case NullType, BooleanType, CharacterType, NumberType, StringType, ListType, VectorType, StructType:
		return true
	case SymbolType, KeywordType, TypeType, FunctionType:
		return true
	default:
		return false
	}
}

func instance(variant LOB, val LOB) (LOB, error) {
	v, ok := variant.(*LType)
	if !ok {
		return nil, Error(ArgumentErrorKey, TypeType.text, variant)
	}
	if isPrimitiveType(v) {
		return val, nil
	}
	return &LInstance{v, val}, nil
}


var ErrorType = intern("<error>")

type LError struct {
	data LOB
	text string
}

// Type returns the type of the object
func (*LError) Type() LOB {
	return ErrorType
}

// Value returns the object itself for primitive types
func (err *LError) Value() LOB {
	return err
}

// Equal returns true if the object is equal to the argument
func (err *LError) Equal(another LOB) bool {
	return LOB(err) == another
}

// String returns the string representation of the object
func (err *LError) String() string {
	return "#<error>" + write(err.data)
}

func (err *LError) Error() string {
	s := err.String()
	if err.text != "" {
		s += " [in " + err.text + "]"
	}
	return s
}


func newError(elements ...LOB) *LError {
	data := vector(elements...)
	return &LError{data, ""}
}


//
// Error - creates a new Error from the arguments. The first is an actual Ell object, the rest are interpreted as/converted to strings
//
func Error(errkey LOB, args ...interface{}) error {
	var buf bytes.Buffer
	for _, o := range args {
		if l, ok := o.(LOB); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	if !isKeyword(errkey) {
		errkey = ErrorKey
	}
	return newError(errkey, LOB(newString(buf.String())))
}

func theError(o interface{}) (*LError, bool) {
	if o == nil {
		return nil, false
	}
	if err, ok := o.(*LError); ok {
		return err, true
	}
	return nil, false

}

func isError(o interface{}) bool {
	_, ok := theError(o)
	return ok
}

func errorData(err *LError) LOB {
	return err.data
}

// Error
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
