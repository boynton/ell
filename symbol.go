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

func Intern(name string) *LOB {
	sym, ok := symtab[name]
	if !ok {
		sym = new(LOB)
		sym.text = name
		if IsValidKeywordName(name) {
			sym.Type = KeywordType
		} else if IsValidTypeName(name) {
			sym.Type = TypeType
		} else if IsValidSymbolName(name) {
			sym.Type = SymbolType
		} else {
			panic("invalid symbol/type/keyword name passed to intern: " + name)
		}
		symtab[name] = sym
	}
	return sym
}

func IsValidSymbolName(name string) bool {
	return len(name) > 0
}

func IsValidTypeName(s string) bool {
	n := len(s)
	if n > 2 && s[0] == '<' && s[n-1] == '>' {
		return true
	}
	return false
}

func IsValidKeywordName(s string) bool {
	n := len(s)
	if n > 1 && s[n-1] == ':' {
		return true
	}
	return false
}

func ToKeyword(obj *LOB) (*LOB, error) {
	switch obj.Type {
	case KeywordType:
		return obj, nil
	case TypeType:
		return Intern(obj.text[1:len(obj.text)-1] + ":"), nil
	case SymbolType:
		return Intern(obj.text + ":"), nil
	case StringType:
		if IsValidKeywordName(obj.text) {
			return Intern(obj.text), nil
		} else if IsValidSymbolName(obj.text) {
			return Intern(obj.text + ":"), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-keyword expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type)
}

func keywordify(s *LOB) *LOB {
	switch s.Type {
	case StringType:
		if !IsValidKeywordName(s.text) {
			return Intern(s.text + ":")
		}
		return Intern(s.text)
	case SymbolType:
		return Intern(s.text + ":")
	}
	return s
}

func typeNameString(s string) string {
	return s[1 : len(s)-1]
}

// <type> -> <symbol>
func TypeName(t *LOB) (*LOB, error) {
	if !IsType(t) {
		return nil, Error(ArgumentErrorKey, "type-name expected a <type>, got a ", t.Type)
	}
	return Intern(typeNameString(t.text)), nil
}

// <keyword> -> <symbol>
func KeywordName(t *LOB) (*LOB, error) {
	if !IsKeyword(t) {
		return nil, Error(ArgumentErrorKey, "keyword-name expected a <keyword>, got a ", t.Type)
	}
	return unkeyworded(t)
}

func keywordNameString(s string) string {
	return s[:len(s)-1]
}

func unkeywordedString(k *LOB) string {
	if IsKeyword(k) {
		return keywordNameString(k.text)
	}
	return k.text
}

func unkeyworded(obj *LOB) (*LOB, error) {
	if IsSymbol(obj) {
		return obj, nil
	}
	if IsKeyword(obj) {
		return Intern(keywordNameString(obj.text)), nil
	}
	return nil, Error(ArgumentErrorKey, "Expected <keyword> or <symbol>, got ", obj.Type)
}

func ToSymbol(obj *LOB) (*LOB, error) {
	switch obj.Type {
	case KeywordType:
		return Intern(keywordNameString(obj.text)), nil
	case TypeType:
		return Intern(typeNameString(obj.text)), nil
	case SymbolType:
		return obj, nil
	case StringType:
		if IsValidSymbolName(obj.text) {
			return Intern(obj.text), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-symbol expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type)
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = initSymbolTable()

func initSymbolTable() map[string]*LOB {
	syms := make(map[string]*LOB, 0)
	TypeType = &LOB{text: "<type>"}
	TypeType.Type = TypeType //mutate to bootstrap type type
	syms[TypeType.text] = TypeType

	KeywordType = &LOB{Type: TypeType, text: "<keyword>"}
	syms[KeywordType.text] = KeywordType

	SymbolType = &LOB{Type: TypeType, text: "<symbol>"}
	syms[SymbolType.text] = SymbolType

	return syms
}

func Symbols() []*LOB {
	syms := make([]*LOB, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func Symbol(names []*LOB) (*LOB, error) {
	size := len(names)
	if size < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch o.Type {
		case StringType, SymbolType:
			s = o.text
		default:
			return nil, Error(ArgumentErrorKey, "symbol name component invalid: ", o)
		}
		name += s
	}
	return Intern(name), nil
}
