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

func newString(s string) *LAny {
	return &LAny{ltype: typeString, text: s}
}

func asString(obj *LAny) (string, error) {
	if !isString(obj) {
		return "", TypeError(typeString, obj)
	}
	return obj.text, nil
}

func toString(a *LAny) *LAny {
	return newString(a.String())
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

func newCharacter(c rune) *LAny {
	return &LAny{ltype: typeCharacter, character: c}
}

func asCharacter(c *LAny) (rune, error) {
	if !isCharacter(c) {
		return 0, TypeError(typeCharacter, c)
	}
	return c.character, nil
}

func stringCharacters(s *LAny) []*LAny {
	chars := make([]*LAny, len(s.text))
	for i, c := range s.text {
		chars[i] = newCharacter(c)
	}
	return chars
}

func stringToVector(s *LAny) *LAny {
	return vector(stringCharacters(s)...)
}

func stringToList(s *LAny) *LAny {
	return list(stringCharacters(s)...)
}
