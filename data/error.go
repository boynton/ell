/*
Copyright 2021 Lee Boynton

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
package data

import (
	"bytes"
	"fmt"
)

type Error struct {
	Data Value
}

// Q: do I really need this? It is not part of EllDN. It has Instance syntax anyway. So...like UUID/Timestamp, right?
var ErrorType Value = primitiveType("<error>")

var SyntaxErrorKey = Intern("syntax-error:")
var IOErrorKey = Intern("io-error:")
var ArgumentErrorKey = Intern("argument-error:")

// ErrorKey - used for generic errors.
// By convention, error data (which is a vector of values), should have a keyword as the first element, and then a message
// as the second, and then optional other data after that.
var ErrorKey = Intern("error:")

func NewError(errkey Value, args ...interface{}) *Error {
	var buf bytes.Buffer
	for _, o := range args {
		if l, ok := o.(Value); ok {
			buf.WriteString(fmt.Sprintf("%s", l.String()))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	if errkey.Type() != KeywordType {
		errkey = ErrorKey
	}
	return MakeError(errkey, NewString(buf.String()))
}

func MakeError(elements ...Value) *Error {
	data := NewVector(elements...)
	return &Error{Data: data}
}

func (err *Error) Type() Value {
	return ErrorType
}

func (err *Error) String() string {
	return fmt.Sprintf("#<error>%s", err.Data.String())
}

func (err1 *Error) Equals(another Value) bool {
	if err2, ok := another.(*Error); ok {
		return Equal(err1.Data, err2.Data)
	}
	return false
}

// for golang error
func (err *Error) Error() string {
	return err.String()
}
