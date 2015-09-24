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

//
// LSymbol holds symbols, keywords, and types. Use the tag to distinguish between them
//
type LSymbol struct {
	Name string
	tag  int //an incrementing sequence number for symbols, -1 for types, and -2 for keywords
}

const typeTag = -1
const keywordTag = -2

var symtag int

func intern(name string) *LSymbol {
	sym, ok := symtab[name]
	if !ok {
		if isValidKeywordName(name) {
			sym = &LSymbol{name, keywordTag}
		} else if isValidTypeName(name) {
			sym = &LSymbol{name, typeTag}
		} else if isValidSymbolName(name) {
			sym = &LSymbol{name, symtag}
			symtag++
		}
		if sym == nil {
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

var typeSymbol = intern("<symbol>")
var typeKeyword = intern("<keyword>")
var typeType = intern("<type>")

// Type returns the type of the object. Since LSymbol represents keywords, types, and regular
// symbols, it could return any of those three values
func (sym *LSymbol) Type() LAny {
	if sym.tag == keywordTag {
		return typeKeyword
	} else if sym.tag == typeTag {
		return typeType
	}
	return typeSymbol
}

// Value returns the object itself for primitive types
func (sym *LSymbol) Value() LAny {
	return sym
}

// Equal returns true if the object is equal to the argument
func (sym *LSymbol) Equal(another LAny) bool {
	if a, ok := another.(*LSymbol); ok {
		return sym == a
	}
	return false
}

func (sym *LSymbol) String() string {
	return sym.Name
}

func (sym *LSymbol) Copy() LAny {
	return sym
}

func isSymbol(obj LAny) bool {
	sym, ok := obj.(*LSymbol)
	return ok && sym.tag >= 0
}

func isType(obj LAny) bool {
	sym, ok := obj.(*LSymbol)
	return ok && sym.tag == typeTag
}

func isKeyword(obj LAny) bool {
	sym, ok := obj.(*LSymbol)
	return ok && sym.tag == keywordTag
}

func typeName(obj LAny) (*LSymbol, error) {
	sym, ok := obj.(*LSymbol)
	if ok && sym.tag == typeTag {
		return intern(sym.Name[1 : len(sym.Name)-1]), nil
	}
	return nil, Error("Type error: expected <type>, got ", obj)
}

func unkeywordedString(sym *LSymbol) string {
	if sym.tag == keywordTag {
		return sym.Name[:len(sym.Name)-1]
	}
	return sym.Name
}

func unkeyworded(obj LAny) (LAny, error) {
	sym, ok := obj.(*LSymbol)
	if ok {
		switch sym.tag {
		case keywordTag:
			return intern(sym.Name[:len(sym.Name)-1]), nil
		case typeTag:
			//nothing
		default: //already a regular symbol
			return obj, nil
		}
	}
	return nil, Error("Type error: expected <keyword> or <symbol>, got ", obj)
}

func keywordToSymbol(obj LAny) (LAny, error) {
	sym, ok := obj.(*LSymbol)
	if ok && sym.tag == keywordTag {
		return intern(sym.Name[:len(sym.Name)-1]), nil
	}
	return nil, Error("Type error: expected <keyword>, got ", obj)
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]*LSymbol{}

func symbols() []LAny {
	syms := make([]LAny, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func symbol(names []LAny) (LAny, error) {
	size := len(names)
	if size < 1 {
		return nil, ArgcError("symbol", "1+", size)
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch t := o.(type) {
		case LString:
			s = string(t)
		case *LSymbol:
			s = t.Name
		default:
			return nil, Error("symbol name component invalid: ", o)
		}
		name += s
	}
	return intern(name), nil
}
