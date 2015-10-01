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
	"strings"
)

// LString - the concrete type for lists
type LString struct {
	value string
}

// StringType - the Type object for this kind of value
var StringType = intern("<string>")

// Type returns the type of the object
func (*LString) Type() LOB {
	return StringType
}

// Value returns the object itself for primitive types
func (s *LString) Value() LOB {
	return s
}

// Equal returns true if the object is equal to the argument
func (s *LString) Equal(another LOB) bool {
	s2, ok := another.(*LString)
	if ok {
		return s.value == s2.value
	}
	return false
}

// String returns the string representation of the object
func (s *LString) String() string {
	return s.value
}

func isString(obj LOB) bool {
	return obj.Type() == StringType
}

// EmptyString
var EmptyString = newString("")

func newString(s string) *LString {
	return &LString{s}
}

func asString(obj LOB) (string, error) {
	s, ok := obj.(*LString)
	if !ok {
		return "", Error(ArgumentErrorKey, StringType, obj)
	}
	return s.value, nil
}

func toString(a LOB) (*LString, error) {
	switch t := a.(type) {
	case *LString:
		return t, nil
	case *LVector:
		var chars []rune
		for _, e := range t.elements {
			c, ok := e.(*LCharacter)
			if !ok {
				return nil, Error(ArgumentErrorKey, "to-string: vector element is not a <character>: ", e)
			}
			chars = append(chars, c.value)
		}
		return newString(string(chars)), nil
	case *LList:
		var chars []rune
		for t != EmptyList {
			c, ok := t.car.(*LCharacter)
			if !ok {
				return nil, Error(ArgumentErrorKey, "to-string: list element is not a <character>: ", t.car)
			}
			chars = append(chars, c.value)
			t = t.cdr
		}
		return newString(string(chars)), nil
	case *LNumber, *LBoolean, *LCharacter:
		return newString(a.String()), nil
	default:
		return nil, Error(ArgumentErrorKey, "to-string: cannot convert argument to <string>: ", a)
	}
}

func stringLength(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}

func encodeString(s string) string {
	buf := []rune{}
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
			buf = append(buf, c)
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

// LCharacter - the concrete type for characters
type LCharacter struct { // <character>
	value rune
}

// CHaracterType - the Type object for this kind of value
var CharacterType = intern("<character>")

func isCharacter(obj LOB) bool {
	return obj.Type() == CharacterType
}

// String returns the string representation of the object
func (c *LCharacter) String() string {
	return string([]rune{c.value})
}

// Type returns the type of the object
func (*LCharacter) Type() LOB {
	return CharacterType
}

// Value returns the object itself for primitive types
func (c *LCharacter) Value() LOB {
	return c
}

// Equal returns true if the object is equal to the argument
func (c *LCharacter) Equal(another LOB) bool {
	c2, ok := another.(*LCharacter)
	if ok {
		return c.value == c2.value
	}
	return false
}

func newCharacter(c rune) LOB {
	return &LCharacter{c}
}

func toCharacter(o LOB) (LOB, error) {
	switch t := o.(type) {
	case *LCharacter:
		return t, nil
	case *LString:
		if len(t.value) == 1 {
			for _, r := range t.value {
				return newCharacter(r), nil
			}
		}
	case *LNumber:
		r := rune(int(t.value))
		return newCharacter(r), nil
	}
	return nil, Error(ArgumentErrorKey, "Cannot convert to <character>: ", o)
}

func asCharacter(o LOB) (rune, error) {
	c, ok := o.(*LCharacter)
	if !ok {
		return 0, Error(ArgumentErrorKey, "Not a <character>", o)
	}
	return c.value, nil
}

func stringCharacters(s *LString) []LOB {
	var chars []LOB
	for _, c := range s.value {
		chars = append(chars, newCharacter(c))
	}
	return chars
}

func stringRef(s *LString, idx int) LOB {
	//utf8 requires a scan
	for i, r := range s.value {
		if i == idx {
			return newCharacter(r)
		}
	}
	return Null
}

func stringToVector(s *LString) *LVector {
	return vector(stringCharacters(s)...)
}

func stringToList(s *LString) *LList {
	return list(stringCharacters(s)...)
}

//func stringSplit(obj *LString, delims *LString) (LOB, error) {
func stringSplit(obj LOB, delims LOB) (*LList, error) {
	s, ok := obj.(*LString)
	if !ok {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 1, got ", obj)
	}
	d, ok := delims.(*LString)
	if !ok {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 2, got ", delims)
	}
	lst := EmptyList
	tail := EmptyList
	for _, s := range strings.Split(s.value, d.value) {
		if lst == EmptyList {
			lst = list(newString(s))
			tail = lst
		} else {
			tail.cdr = list(newString(s))
			tail = tail.cdr
		}
	}
	return lst, nil
}

func stringJoin(seq LOB, delims LOB) (LOB, error) {
	d, ok := delims.(*LString)
	if !ok {
		return nil, Error(ArgumentErrorKey, "join expected a <string> for argument 2, got ", delims)
	}
	switch t := seq.(type) {
	case *LList:
		result := ""
		for t != EmptyList {
			o := t.car
			if o != EmptyString && o != Null {
				if result != "" {
					result += d.value
				}
				result += o.String()
			}
			t = t.cdr
		}
		return newString(result), nil
	case *LVector:
		result := ""
		count := len(t.elements)
		for i := 0; i < count; i++ {
			o := t.elements[i]
			if o != EmptyString && o != Null {
				if result != "" {
					result += d.value
				}
				result += o.String()
			}
		}
		return newString(result), nil
	default:
		return nil, Error(ArgumentErrorKey, "join expected a <list> or <vector> for argument 1, got a ", seq.Type())
	}
}
