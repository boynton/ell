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

//
// LAny is the generic Ell object. It supports querying its symbolic type name at runtime
//
type LAny interface { // <any> - the interface all Lxxx types must implement
	Type() LAny
	Value() LAny
	Equal(another LAny) bool
	String() string
	Copy() LAny
}

func equal(o1 LAny, o2 LAny) bool {
	if o1 == o2 {
		return true
	}
	return o1.Equal(o2)
}

//
// LNull is the type of the null value
//
type LNull int // <null>

var typeNull = intern("<null>")

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList
const Null = LNull(0)

// Type returns the type of the object
func (LNull) Type() LAny {
	return typeNull
}

// Value returns the object itself for primitive types
func (LNull) Value() LAny {
	return Null
}

// Equal returns true if the object is equal to the argument
func (LNull) Equal(another LAny) bool {
	return another == Null
}

func (v LNull) String() string {
	return "null"
}

func (v LNull) Copy() LAny {
	return v
}

//
// LEOF is the type of the EOF marker
//
type LEOF int // <eof>

// EOF is Ell's EOF object
const EOF = LEOF(0)

var typeEOF = intern("<eof>")

// Type returns the type of the object
func (LEOF) Type() LAny {
	return typeEOF
}

// Value returns the object itself for primitive types
func (LEOF) Value() LAny {
	return EOF
}

// Equal returns true if the object is equal to the argument
func (LEOF) Equal(another LAny) bool {
	return another == EOF
}

func (LEOF) String() string {
	return "#[eof]"
}

func (LEOF) Copy() LAny {
	return EOF
}

//
// LBoolean is the type of true and false
//
type LBoolean bool // <boolean>

//True is Ell's true constant
const True LBoolean = LBoolean(true)

//False is Ell's false constant
const False LBoolean = LBoolean(false)

var typeBoolean = intern("<boolean>")

func isBoolean(obj LAny) bool {
	_, ok := obj.(LBoolean)
	return ok
}

// Type returns the type of the object
func (LBoolean) Type() LAny {
	return typeBoolean
}

// Value returns the object itself for primitive types
func (b LBoolean) Value() LAny {
	return b
}

// Equal returns true if the object is equal to the argument
func (b LBoolean) Equal(another LAny) bool {
	if a, ok := another.(LBoolean); ok {
		return b == a
	}
	return false
}

func (b LBoolean) String() string {
	return strconv.FormatBool(bool(b))
}

func (b LBoolean) Copy() LAny {
	return b
}

// LInstance is a typed value
type LInstance struct { // <user-defined-type>
	tag   *LSymbol
	value LAny
}

func instance(tag LAny, val LAny) (LAny, error) {
	sym, ok := tag.(*LSymbol)
	if !ok || !isValidTypeName(sym.Name) {
		return nil, TypeError(typeType, tag)
	}
	switch sym {
	case typeString, typeNumber, typeNull, typeBoolean, typeChar, typeEOF:
		return val, nil
	case typeStruct, typeList, typeVector, typeSymbol, typeFunction, typeInput, typeOutput:
		return val, nil
	default:
		return &LInstance{tag: sym, value: val}, nil
	}
}

// Type returns the type of the object
func (s *LInstance) Type() LAny {
	return s.tag
}

// Value returns the value of the object
func (s *LInstance) Value() LAny {
	return s.value
}

// Equal returns true if the object is equal to the argument
func (s *LInstance) Equal(another LAny) bool {
	if a, ok := another.(*LInstance); ok {
		return s.tag == a.tag && s.value.Equal(a.value)
	}
	return false
}

// String of a instance, i.e. #<point>{x: 1 y: 2} or #<uuid>"0bbbc94a-5e14-11e5-81e6-003ee1be85f9"
func (s *LInstance) String() string {
	return "#" + s.tag.String() + write(s.value)
}

func (i *LInstance) Copy() LAny {
	c, _ := instance(i.tag, i.value.Copy())
	return c
}

//
// Error - creates a new Error from the arguments
//
func Error(arg1 interface{}, args ...interface{}) error {
	var buf bytes.Buffer
	if l, ok := arg1.(LAny); ok {
		buf.WriteString(fmt.Sprintf("%v", write(l)))
	} else {
		buf.WriteString(fmt.Sprintf("%v", arg1))
	}
	for _, o := range args {
		if l, ok := o.(LAny); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	err := GenericError{buf.String()}
	return &err
}

// TypeError - an error indicating expected and actual value for a type mismatch
func TypeError(typeSym LAny, obj LAny) error {
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
