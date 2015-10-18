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
var EmptyString = newString("")

func newString(s string) *LOB {
	str := newLOB(StringType)
	str.text = s
	return str
}

func asString(obj *LOB) (string, error) {
	if !isString(obj) {
		return "", Error(ArgumentErrorKey, StringType, obj)
	}
	return obj.text, nil
}

func toString(a *LOB) (*LOB, error) {
	switch a.Type {
	case CharacterType:
		return newString(string([]rune{rune(a.fval)})), nil
	case StringType:
		return a, nil
	case SymbolType, KeywordType, TypeType, BlobType:
		return newString(a.text), nil
	case NumberType, BooleanType:
		return newString(a.String()), nil
	case VectorType:
		var chars []rune
		for _, c := range a.elements {
			if !isCharacter(c) {
				return nil, Error(ArgumentErrorKey, "to-string: vector element is not a <character>: ", c)
			}
			chars = append(chars, rune(c.fval))
		}
		return newString(string(chars)), nil
	case ListType:
		var chars []rune
		for a != EmptyList {
			c := car(a)
			if !isCharacter(c) {
				return nil, Error(ArgumentErrorKey, "to-string: list element is not a <character>: ", c)
			}
			chars = append(chars, rune(c.fval))
			a = a.cdr
		}
		return newString(string(chars)), nil
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

func newCharacter(c rune) *LOB {
	char := newLOB(CharacterType)
	char.fval = float64(c)
	return char
}

func toCharacter(c *LOB) (*LOB, error) {
	switch c.Type {
	case CharacterType:
		return c, nil
	case StringType:
		if len(c.text) == 1 {
			for _, r := range c.text {
				return newCharacter(r), nil
			}
		}
	case NumberType:
		r := rune(int(c.fval))
		return newCharacter(r), nil
	}
	return nil, Error(ArgumentErrorKey, "Cannot convert to <character>: ", c)
}

func asCharacter(c *LOB) (rune, error) {
	if !isCharacter(c) {
		return 0, Error(ArgumentErrorKey, "Not a <character>", c)
	}
	return rune(c.fval), nil
}

func stringCharacters(s *LOB) []*LOB {
	var chars []*LOB
	for _, c := range s.text {
		chars = append(chars, newCharacter(c))
	}
	return chars
}

func stringRef(s *LOB, idx int) *LOB {
	//utf8 requires a scan
	for i, r := range s.text {
		if i == idx {
			return newCharacter(r)
		}
	}
	return Null
}

func stringToVector(s *LOB) *LOB {
	return vector(stringCharacters(s)...)
}

func stringToList(s *LOB) *LOB {
	return list(stringCharacters(s)...)
}

func stringSplit(obj *LOB, delims *LOB) (*LOB, error) {
	if !isString(obj) {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 1, got ", obj)
	}
	if !isString(delims) {
		return nil, Error(ArgumentErrorKey, "split expected a <string> for argument 2, got ", delims)
	}
	lst := EmptyList
	tail := EmptyList
	for _, s := range strings.Split(obj.text, delims.text) {
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

func stringJoin(seq *LOB, delims *LOB) (*LOB, error) {
	if !isString(delims) {
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
		return newString(result), nil
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
		return newString(result), nil
	default:
		return nil, Error(ArgumentErrorKey, "join expected a <list> or <vector> for argument 1, got a ", seq.Type)
	}
}
