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

package ell

import (
	"strings"
)

// EmptyString
var EmptyString = String("")

func String(s string) *LOB {
	str := new(LOB)
	str.Type = StringType
	str.text = s
	return str
}

func AsStringValue(obj *LOB) (string, error) {
	if !IsString(obj) {
		return "", Error(ArgumentErrorKey, StringType, obj)
	}
	return obj.text, nil
}

func ToString(a *LOB) (*LOB, error) {
	switch a.Type {
	case CharacterType:
		return String(string([]rune{rune(a.fval)})), nil
	case StringType:
		return a, nil
	case SymbolType, KeywordType, TypeType, BlobType:
		return String(a.text), nil
	case NumberType, BooleanType:
		return String(a.String()), nil
	case VectorType:
		var chars []rune
		for _, c := range a.elements {
			if !IsCharacter(c) {
				return nil, Error(ArgumentErrorKey, "to-string: vector element is not a <character>: ", c)
			}
			chars = append(chars, rune(c.fval))
		}
		return String(string(chars)), nil
	case ListType:
		var chars []rune
		for a != EmptyList {
			c := Car(a)
			if !IsCharacter(c) {
				return nil, Error(ArgumentErrorKey, "to-string: list element is not a <character>: ", c)
			}
			chars = append(chars, rune(c.fval))
			a = a.cdr
		}
		return String(string(chars)), nil
	default:
		return nil, Error(ArgumentErrorKey, "to-string: cannot convert argument to <string>: ", a)
	}
}

func StringLength(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}

func EncodeString(s string) string {
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

func Character(c rune) *LOB {
	char := new(LOB)
	char.Type = CharacterType
	char.fval = float64(c)
	return char
}

func ToCharacter(c *LOB) (*LOB, error) {
	switch c.Type {
	case CharacterType:
		return c, nil
	case StringType:
		if len(c.text) == 1 {
			for _, r := range c.text {
				return Character(r), nil
			}
		}
	case NumberType:
		r := rune(int(c.fval))
		return Character(r), nil
	}
	return nil, Error(ArgumentErrorKey, "Cannot convert to <character>: ", c)
}

func AsCharacter(c *LOB) (rune, error) {
	if !IsCharacter(c) {
		return 0, Error(ArgumentErrorKey, "Not a <character>", c)
	}
	return rune(c.fval), nil
}

func StringCharacters(s *LOB) []*LOB {
	var chars []*LOB
	for _, c := range s.text {
		chars = append(chars, Character(c))
	}
	return chars
}

func StringRef(s *LOB, idx int) *LOB {
	//utf8 requires a scan
	for i, r := range s.text {
		if i == idx {
			return Character(r)
		}
	}
	return Null
}

func stringToVector(s *LOB) *LOB {
	return Vector(StringCharacters(s)...)
}

func stringToList(s *LOB) *LOB {
	return List(StringCharacters(s)...)
}

func StringSplit(obj *LOB, delims *LOB) (*LOB, error) {
	if !IsString(obj) {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 1, got ", obj)
	}
	if !IsString(delims) {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 2, got ", delims)
	}
	lst := EmptyList
	tail := EmptyList
	for _, s := range strings.Split(obj.text, delims.text) {
		if lst == EmptyList {
			lst = List(String(s))
			tail = lst
		} else {
			tail.cdr = List(String(s))
			tail = tail.cdr
		}
	}
	return lst, nil
}

func StringJoin(seq *LOB, delims *LOB) (*LOB, error) {
	if !IsString(delims) {
		return nil, Error(ArgumentErrorKey, "join expected a <string> for argument 2, got ", delims)
	}
	switch seq.Type {
	case ListType:
		result := ""
		for seq != EmptyList {
			o := seq.car
			if o != EmptyString && o != Null {
				if result != "" {
					result += delims.text
				}
				result += o.String()
			}
			seq = seq.cdr
		}
		return String(result), nil
	case VectorType:
		result := ""
		elements := seq.elements
		count := len(elements)
		for i := 0; i < count; i++ {
			o := elements[i]
			if o != EmptyString && o != Null {
				if result != "" {
					result += delims.text
				}
				result += o.String()
			}
		}
		return String(result), nil
	default:
		return nil, Error(ArgumentErrorKey, "join expected a <list> or <vector> for argument 1, got a ", seq.Type)
	}
}
