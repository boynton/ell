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

var SymbolType *LType // <symbol>

type LSymbol struct {
	text string
	binding LOB
}

func (sym *LSymbol) Type() LOB {
	return SymbolType
}

func (sym *LSymbol) Value() LOB {
	return sym
}

func (sym *LSymbol) String() string {
	return sym.text
}

func (sym *LSymbol) Equal(another LOB) bool {
	return LOB(sym) ==  another
}

func isSymbol(obj LOB) bool {
	return obj.Type() == SymbolType
}

var KeywordType *LType // <keyword>

type LKeyword struct {
	text string
}

func (key *LKeyword) Type() LOB {
	return KeywordType
}

func (key *LKeyword) String() string {
	return key.text
}

func (key *LKeyword) Equal(another LOB) bool {
	return LOB(key) == another
}

func (key *LKeyword) Value() LOB {
	return key
}

func isKeyword(obj LOB) bool {
	return obj.Type() == KeywordType
}

var TypeType *LType // <type>

type LType struct {
	text string
}

func (variant *LType) String() string {
	return variant.text
}

func (variant *LType) Type() LOB {
	return TypeType
}

func (variant *LType) Value() LOB {
	return variant
}

func (variant *LType) Equal(another LOB) bool {
	return LOB(variant) == another
}

func isType(obj LOB) bool {
	return obj.Type() == TypeType
}


//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = initSymbolTable()

func initSymbolTable() map[string]LOB {
	syms := make(map[string]LOB, 0)
	TypeType = &LType{"<type>"}
	syms[TypeType.text] = TypeType
	KeywordType = &LType{"<keyword>"}
	syms[KeywordType.text] = KeywordType
	SymbolType = &LType{"<symbol>"}
	syms[SymbolType.text] = SymbolType
	return syms
}

func internSymbol(name string) *LSymbol {
	entry := intern(name)
	sym, ok := entry.(*LSymbol)
	if !ok {
		panic("internSymbol: not a symbol")
	}
	return sym
}

func internKeyword(name string) *LKeyword {
	entry := intern(name)
	sym, ok := entry.(*LKeyword)
	if !ok {
		panic("internKeyword: not a keyword")
	}
	return sym
}

func internType(name string) *LType {
	entry := intern(name)
	sym, ok := entry.(*LType)
	if !ok {
		panic("internType: not a type")
	}
	return sym
}

func intern(name string) LOB {
	entry, ok := symtab[name]
	if !ok {
		if isValidKeywordName(name) {
			entry = &LKeyword{name}
		} else if isValidTypeName(name) {
			entry = &LType{name}
		} else if isValidSymbolName(name) {
			entry = &LSymbol{name, nil}
		} else {
			panic("invalid symbol/type/keyword name passed to intern: " + name)
		}
		symtab[name] = entry
	}
	return entry
}

func isValidSymbolName(name string) bool {
	//and not a type or keyword?
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

func toKeyword(obj LOB) (*LKeyword, error) {
	switch t := obj.(type) {
	case *LKeyword:
		return t, nil
	case *LType:
		return internKeyword(t.text[1:len(t.text)-1] + ":"), nil
	case *LSymbol:
		return internKeyword(t.text + ":"), nil
	case *LString:
		if isValidKeywordName(t.value) {
			return internKeyword(t.value), nil
		} else if isValidSymbolName(t.value) {
			return internKeyword(t.value + ":"), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-keyword expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type())
}

func keywordify(s LOB) LOB {
	switch t := s.(type) {
	case *LString:
		if !isValidKeywordName(t.value) {
			return intern(t.value + ":")
		}
		return intern(t.value)
	case *LSymbol:
		return intern(t.text + ":")
	}
	return s
}

func typeNameString(s string) string {
	return s[1 : len(s)-1]
}

// <type> -> <symbol>
func typeName(obj LOB) (LOB, error) {
	t, ok := obj.(*LType)
	if !ok {
		return nil, Error(ArgumentErrorKey, "type-name expected a <type>, got a ", obj.Type())
	}
	return intern(typeNameString(t.text)), nil
}

// <keyword> -> <symbol>
func keywordName(obj LOB) (LOB, error) {
	t, ok := obj.(*LKeyword)
	if !ok {
		return nil, Error(ArgumentErrorKey, "keyword-name expected a <keyword>, got a ", obj.Type())
	}
	return unkeyworded(t)
}

func keywordNameString(s string) string {
	return s[:len(s)-1]
}

func unkeywordedString(obj LOB) string {
	switch k := obj.(type) {
	case *LKeyword:
		return keywordNameString(k.text)
	case *LString:
		return keywordNameString(k.value)
	default:
		panic("unkeywordedString expected a <keyword> or <string>, got neither")
	}
}

func unkeyworded(obj LOB) (*LSymbol, error) {
	switch t := obj.(type) {
	case *LSymbol:
		return t, nil
	case *LKeyword:
		return internSymbol(keywordNameString(t.text)), nil
	}
	return nil, Error(ArgumentErrorKey, "Expected <keyword> or <symbol>, got ", obj.Type())
}

func toSymbol(obj LOB) (*LSymbol, error) {
	switch t := obj.(type) {
	case *LKeyword:
		return internSymbol(keywordNameString(t.text)), nil
	case *LType:
		return internSymbol(typeNameString(t.text)), nil
	case *LSymbol:
		return t, nil
	case *LString:
		if isValidSymbolName(t.value) {
			return internSymbol(t.value), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "to-symbol expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type())
}

func symbols() []LOB {
	syms := make([]LOB, 0, len(symtab))
	for _, entry := range symtab {
		syms = append(syms, LOB(entry))
	}
	return syms
}

func symbol(names []LOB) (LOB, error) {
	size := len(names)
	if size < 1 {
		return nil, Error(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch t := o.(type) {
		case *LString:
			s = t.value
		case *LSymbol:
			s = t.text
		default:
			return nil, Error(ArgumentErrorKey, "symbol name component invalid: ", o)
		}
		name += s
	}
	return intern(name), nil
}
