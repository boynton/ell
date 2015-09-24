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

type LString string // <string>

var typeString = intern("<string>")

func isString(obj LAny) bool {
	_, ok := obj.(LString)
	return ok
}

func stringValue(obj LAny) (string, error) {
	switch s := obj.(type) {
	case LString:
		return string(s), nil
	default:
		return "", TypeError(typeString, obj)
	}
}

// Type returns the type of the object
func (LString) Type() LAny {
	return typeString
}

// Value returns the object itself for primitive types
func (s LString) Value() LAny {
	return s
}

// Equal returns true if the object is equal to the argument
func (s LString) Equal(another LAny) bool {
	if a, ok := another.(LString); ok {
		return s == a
	}
	return false
}

func (s LString) Copy() LAny {
	return s
}

func encodeString(s string) string {
	buf := []byte{}
	buf = append(buf, '"')
	for _, c := range s {
		switch c {
		case '"':
			buf = append(buf, '\\')
			buf = append(buf, '"')
		case '\\':
			buf = append(buf, '\\')
			buf = append(buf, '\\')
		case '\n':
			buf = append(buf, '\\')
			buf = append(buf, 'n')
		case '\t':
			buf = append(buf, '\\')
			buf = append(buf, 't')
		case '\f':
			buf = append(buf, '\\')
			buf = append(buf, 'f')
		case '\b':
			buf = append(buf, '\\')
			buf = append(buf, 'b')
		case '\r':
			buf = append(buf, '\\')
			buf = append(buf, 'r')
		default:
			//to do: handle UTF8 correctly
			buf = append(buf, byte(c))
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

func (s LString) String() string {
	return string(s)
}

//
// LChar - Ell characters
//
type LChar rune // <char>

var typeChar = intern("<char>")

func isChar(obj LAny) bool {
	_, ok := obj.(LChar)
	if ok {
		return true
	}
	return ok
}

func newCharacter(c rune) LChar {
	v := LChar(c)
	return v
}

// Type returns the type of the object
func (LChar) Type() LAny {
	return typeChar
}

// Value returns the object itself for primitive types
func (i LChar) Value() LAny {
	return i
}

// Equal returns true if the object is equal to the argument
func (i LChar) Equal(another LAny) bool {
	if a, err := intValue(another); err == nil {
		return int(i) == a
	}
	return false
}

func (i LChar) String() string {
	buf := []rune{rune(i)}
	return string(buf)
}

func (i LChar) Copy() LAny {
	return i
}
