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

var NullType = primitiveType("<null>")
var BooleanType Value = primitiveType("<boolean>")
var CharacterType Value = Intern("<character>")
var NumberType Value = primitiveType("<number>")
var StringType Value = primitiveType("<string>")
var VectorType Value = MakeType("<vector>")
var StructType Value = primitiveType("<struct>")
var ListType Value = MakeType("<list>")
var SymbolType = primitiveType("<symbol>")
var KeywordType = primitiveType("<keyword>")
var TypeType Value = primitiveType("<type>")

//not a concrete type, used as a type assertion that any type is ok.
//Note that the Read function lets the caller choose whether or not
//the keys should be coerced to a specific type (<keyword>, <symbol>, <string>) or not <any>
//
var AnyType Value = Intern("<any>")


//Type is a type tag, of the form <foo> for type Foo. Types are part of the data notation.
type Type struct {
	Text string //only the Name is needed for builtin types
	//TO DO: model non-primitive types here with schema info
}

func MakeType(name string) Value {
	if !IsValidTypeName(name) {
		return NewError(ArgumentErrorKey, NewString("Not a valid type name: " + name))
	}
	return Intern(name)
}

func primitiveType(name string) Value {
	return Intern(name)
}

func IsValidTypeName(s string) bool {
	n := len(s)
	if n > 2 && s[0] == '<' && s[n-1] == '>' {
		return true
	}
	return false
}

func (data *Type) Type() Value {
	return TypeType
}

func (data *Type) String() string {
	return data.Text
}

func (t1 *Type) Equals(another Value) bool {
	if t2, ok := another.(*Type); ok {
		return t1 == t2
	}
	return false
}

func (data *Type) Name() string {
	return data.Text[1 : len(data.Text)-1]
}

func IsType(o Value) bool {
	return o.Type() == TypeType
}

func IsPrimitiveType(tag Value) bool {
	switch tag.Type() {
	case NullType,BooleanType, NumberType, StringType, VectorType, StructType, ListType, SymbolType, KeywordType:
		return true
	default:
		return false
	}
}

func TypeNameOf (val Value) string {
	return val.Type().(*Type).Name()
}
