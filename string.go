/*
Copyright 2015 Lee Boynton

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

package ell

import (
	"strings"

	. "github.com/boynton/ell/data"
)

// AsStringValue - return the native string representation of the object, if possible
func AsStringValue(obj Value) (string, error) {
	if p, ok := obj.(*String); ok {
		return p.Value, nil
	}
	return "", NewError(ArgumentErrorKey, StringType, obj)
}

// ToString - convert the object to a string, if possible
func ToString(a Value) (*String, error) {
	switch p := a.(type) {
	case *Character:
		return NewString(string([]rune{p.Value})), nil
	case *String:
		return p, nil
	case *Blob:
		return NewString(string(p.Value)), nil
	case *Symbol:
		return NewString(p.Text), nil
	case *Keyword:
		return NewString(p.Name()), nil
	case *Type:
		return NewString(p.Name()), nil
	case *Number:
		return NewString(p.String()), nil
	case *Boolean:
		return NewString(p.String()), nil
	case *Vector:
		var chars []rune
		for _, c := range p.Elements {
			if pc, ok := c.(*Character); ok {
				chars = append(chars, pc.Value)
			} else {
				return nil, NewError(ArgumentErrorKey, "to-string: vector element is not a <character>: ", c)
			}
		}
		return NewString(string(chars)), nil
	case *List:
		var chars []rune
		for p != EmptyList {
			c := p.Car
			if pc, ok := c.(*Character); ok {
				chars = append(chars, pc.Value)
			} else {
				return nil, NewError(ArgumentErrorKey, "to-string: list element is not a <character>: ", c)
			}
			p = p.Cdr
		}
		return NewString(string(chars)), nil
	default:
		return nil, NewError(ArgumentErrorKey, "to-string: cannot convert argument to <string>: ", a)
	}
}

// ToCharacter - convert object to a <character> object, if possible
func ToCharacter(c Value) (*Character, error) {
	switch p := c.(type) {
	case *Character:
		return p, nil
	case *String:
		if len(p.Value) == 1 {
			for _, r := range p.Value {
				return NewCharacter(r), nil
			}
		}
	case *Number:
		r := rune(int(p.Value))
		return NewCharacter(r), nil
	}
	return nil, NewError(ArgumentErrorKey, "Cannot convert to <character>: ", c)
}

// StringCharacters - return a slice of <character> objects that represent the string
func StringCharacters(s *String) []Value {
	var chars []Value
	for _, c := range s.Value {
		chars = append(chars, NewCharacter(c))
	}
	return chars
}

// StringRef - return the <character> object at the specified string index
func StringRef(s *String, idx int) Value {
	//utf8 requires a scan
	for i, r := range s.Value {
		if i == idx {
			return NewCharacter(r)
		}
	}
	return Null
}

func StringToVector(s *String) *Vector {
	return NewVector(StringCharacters(s)...)
}

func StringToList(s *String) *List {
	return NewList(StringCharacters(s)...)
}

func StringSplit(obj Value, delims Value) (*List, error) {
	str, ok := obj.(*String)
	if !ok {
		return nil, NewError(ArgumentErrorKey, "split expected a <string> for argument 1, got ", obj)
	}
	del, ok := delims.(*String)
	if !ok {
		return nil, NewError(ArgumentErrorKey, "split expected a <string> for argument 2, got ", delims)
	}
	lst := EmptyList
	tail := EmptyList
	for _, s := range strings.Split(str.Value, del.Value) {
		if lst == EmptyList {
			lst = NewList(NewString(s))
			tail = lst
		} else {
			tail.Cdr = NewList(NewString(s))
			tail = tail.Cdr
		}
	}
	return lst, nil
}

func StringJoin(seq Value, delims Value) (*String, error) {
	del, ok := delims.(*String)
	if !ok {
		return nil, NewError(ArgumentErrorKey, "join expected a <string> for argument 2, got ", delims)
	}
	switch p := seq.(type) {
	case *List:
		result := ""
		for p != EmptyList {
			o := p.Car
			if o != EmptyString && o != Null {
				if result != "" {
					result += del.Value
				}
				result += o.String()
			}
			p = p.Cdr
		}
		return NewString(result), nil
	case *Vector:
		result := ""
		elements := p.Elements
		count := len(elements)
		for i := 0; i < count; i++ {
			o := elements[i]
			if o != EmptyString && o != Null {
				if result != "" {
					result += del.Value
				}
				result += o.String()
			}
		}
		return NewString(result), nil
	default:
		return nil, NewError(ArgumentErrorKey, "join expected a <list> or <vector> for argument 1, got a ", seq.Type)
	}
}

// RuneValue - return native rune value of the object
func RuneValue(obj Value) rune {
	if p, ok := obj.(*Character); ok {
		return p.Value
	}
	return 0
}

// StringValue - return native string value of the object
func StringValue(obj Value) string {
	switch p := obj.(type) {
	case *String:
		return p.Value
	case *Symbol:
		return p.Name()
	case *Keyword:
		return p.Name()
	case *Type:
		return p.Name()
	default:
		return p.String()
	}
}
