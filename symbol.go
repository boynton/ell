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

var symTag int

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
			sym.ival = symTag
			symTag++
		} else {
			panic("invalid symbol/type/keyword name passed to intern: " + name)
		}
		symtab[name] = sym
	}
	return sym
}

func symbolTag(sym *LOB) int {
	return sym.ival
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

// <type> -> <symbol>
func typeName(t *LOB) (*LOB, error) {
	if !isType(t) {
		return nil, Error("Type error: expected <type>, got ", t)
	}
	return intern(t.text[1 : len(t.text)-1]), nil
}

func unkeywordedString(k *LOB) string {
	if isKeyword(k) {
		return k.text[:len(k.text)-1]
	}
	return k.text
}

func unkeyworded(obj *LOB) (*LOB, error) {
	if isSymbol(obj) {
		return obj, nil
	}
	if isKeyword(obj) {
		return intern(obj.text[:len(obj.text)-1]), nil
	}
	return nil, Error("Type error: expected <keyword> or <symbol>, got ", obj)
}

func keywordToSymbol(obj *LOB) (*LOB, error) {
	if isKeyword(obj) {
		return intern(obj.text[:len(obj.text)-1]), nil
	}
	return nil, Error("Type error: expected <keyword>, got ", obj)
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
		return nil, ArgcError("symbol", "1+", size)
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch o.variant {
		case typeString, typeSymbol:
			s = o.text
		default:
			return nil, Error("symbol name component invalid: ", o)
		}
		name += s
	}
	return intern(name), nil
}
