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

func intern(name string) *LOB {
	sym, ok := symtab[name]
	if !ok {
		sym = new(LOB)
		sym.text = name
		if isValidKeywordName(name) {
			sym.variant = typeKeyword
		} else if isValidTypeName(name) {
			sym.variant = typeType
		} else if isValidSymbolName(name) {
			sym.variant = typeSymbol
		} else {
			panic("invalid symbol/type/keyword name passed to intern: " + name)
		}
		symtab[name] = sym
	}
	return sym
}

func isValidSymbolName(name string) bool {
	return len(name) > 0
}

func isValidTypeName(s string) bool {
	n := len(s)
	if n > 2 && s[0] == '<' && s[n-1] == '>' {
		return true
	}
	return false
}

func isValidKeywordName(s string) bool {
	n := len(s)
	if n > 1 && s[n-1] == ':' {
		return true
	}
	return false
}

func toKeyword(obj *LOB) (*LOB, error) {
	switch obj.variant {
	case typeKeyword:
		return obj, nil
	case typeType:
		return intern(obj.text[1:len(obj.text)-1] + ":"), nil
	case typeSymbol:
		return intern(obj.text + ":"), nil
	case typeString:
		if isValidKeywordName(obj.text) {
			return intern(obj.text), nil
		} else if isValidSymbolName(obj.text) {
			return intern(obj.text + ":"), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-keyword expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.variant)
}

func keywordify(s *LOB) *LOB {
	switch s.variant {
	case typeString:
		if !isValidKeywordName(s.text) {
			return intern(s.text + ":")
		}
		return intern(s.text)
	case typeSymbol:
		return intern(s.text + ":")
	}
	return s
}

func typeNameString(s string) string {
	return s[1 : len(s)-1]
}

// <type> -> <symbol>
func typeName(t *LOB) (*LOB, error) {
	if !isType(t) {
		return nil, Error(ArgumentErrorKey, "type-name expected a <type>, got a ", t.variant)
	}
	return intern(typeNameString(t.text)), nil
}

// <keyword> -> <symbol>
func keywordName(t *LOB) (*LOB, error) {
	if !isKeyword(t) {
		return nil, Error(ArgumentErrorKey, "keyword-name expected a <keyword>, got a ", t.variant)
	}
	return unkeyworded(t)
}

func keywordNameString(s string) string {
	return s[:len(s)-1]
}

func unkeywordedString(k *LOB) string {
	if isKeyword(k) {
		return keywordNameString(k.text)
	}
	return k.text
}

func unkeyworded(obj *LOB) (*LOB, error) {
	if isSymbol(obj) {
		return obj, nil
	}
	if isKeyword(obj) {
		return intern(keywordNameString(obj.text)), nil
	}
	return nil, Error(ArgumentErrorKey, "Expected <keyword> or <symbol>, got ", obj.variant)
}

func toSymbol(obj *LOB) (*LOB, error) {
	switch obj.variant {
	case typeKeyword:
		return intern(keywordNameString(obj.text)), nil
	case typeType:
		return intern(typeNameString(obj.text)), nil
	case typeSymbol:
		return obj, nil
	case typeString:
		if isValidSymbolName(obj.text) {
			return intern(obj.text), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-symbol expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.variant)
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = initSymbolTable()

func initSymbolTable() map[string]*LOB {
	syms := make(map[string]*LOB, 0)
	typeType = &LOB{text: "<type>"}
	typeType.variant = typeType //mutate to bootstrap type type
	syms[typeType.text] = typeType

	typeKeyword = &LOB{variant: typeType, text: "<keyword>"}
	syms[typeKeyword.text] = typeKeyword

	typeSymbol = &LOB{variant: typeType, text: "<symbol>"}
	syms[typeSymbol.text] = typeSymbol

	return syms
}

func symbols() []*LOB {
	syms := make([]*LOB, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func symbol(names []*LOB) (*LOB, error) {
	size := len(names)
	if size < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch o.variant {
		case typeString, typeSymbol:
			s = o.text
		default:
			return nil, Error(ArgumentErrorKey, "symbol name component invalid: ", o)
		}
		name += s
	}
	return intern(name), nil
}
