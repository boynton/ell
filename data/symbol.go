/*
Copyright 2021 Lee Boynton

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
package data

import (
	"sync"
)

// Symbols are symbolic identifiers, i.e. Intern("foo") == Intern("foo"), the same objects.
type Symbol struct {
	Text  string //the textual representation of the Symbol
	Value Value  //A hook for a value to bound to the symbol. Used by Ell, not actually part of EllDn Spec.
}

func (data *Symbol) Type() Value {
	return SymbolType
}

func (data *Symbol) String() string {
	return data.Text
}
func (s1 *Symbol) Equals(another Value) bool {
	if s2, ok := another.(*Symbol); ok {
		return s1 == s2
	}
	return false
}

func (data *Symbol) Name() string {
	return data.Text
}

// Keywords are symbolic identifiers with a trailing ':', i.e. `foo:`
type Keyword struct {
	Text string //the textual representation of the Keyword
}

func (kw *Keyword) Type() Value {
	return KeywordType
}
func (kw *Keyword) String() string {
	return kw.Text
}

func (kw *Keyword) Equals(another Value) bool {
	if kw2, ok := another.(*Keyword); ok {
		return kw == kw2
	}
	return false
}

func (kw *Keyword) Name() string {
	return kw.Text[:len(kw.Text)-1]
}

func IsValidSymbolName(name string) bool {
	//todo: restrict to match EllDn spec: must start with `a-zA-Z_`, followed by `a-zA-Z0-9_-?!`, with
	// an embedded '/' (with nonempty on either side).
	//in SADL, it is simpler, it doesn't even have symbols except for the hard coded true and false (no null specified)
	// -> but null means "absence of a  value", i.e. cannot be modeled as a type. Hmm.
	// -> an smithy list (an array) can have null values only if the @sparse trait is present, otherwise illegal. If
	//    a null value is encountered, it should be discarded. This seems pretty bogus to me!
	//I want something more general than SADL for the reasonable subset that maps to Smithy
	return len(name) > 0
}

func IsValidKeywordName(s string) bool {
	n := len(s)
	if n > 1 && s[n-1] == ':' {
		return IsValidSymbolName(s[:n-1])
	}
	return false
}

func ToSymbol(obj Value) (Value, error) {
	switch p := obj.(type) {
	case *Keyword:
		return Intern(p.Name()), nil
	case *Type:
		return Intern(p.Name()), nil
	case *Symbol:
		return obj, nil
	case *String:
		if IsValidSymbolName(p.Value) {
			return Intern(p.Value), nil
		}
		return nil, NewError(ArgumentErrorKey, "to-symbol cannot convert the <string> to a valid <symbol>", obj)
	}
	return nil, NewError(ArgumentErrorKey, "to-symbol expected a <keyword>, <type>, <symbol>, or <string>, got a ", obj.Type())
}

func ToKeyword(o Value) (Value, error) { //return *Keyword instead?
	switch p := o.(type) {
	case *Keyword:
		return p, nil
	case *Type:
		return Intern(p.Name() + ":"), nil
	case *Symbol:
		return Intern(p.Name() + ":"), nil
	case *String:
		if IsValidKeywordName(p.Value) {
			return Intern(p.Value), nil
		} else if IsValidSymbolName(p.Value) {
			return Intern(p.Value + ":"), nil
		} else {
			return nil, NewError(ArgumentErrorKey, "to-keyword cannot convert the <string> to a valid <keyword>", o)
		}
	}
	return nil, NewError(ArgumentErrorKey, "to-keyword expected a <keyword>, <type>, <symbol>, or <string>, got a ", o.Type())
}

// The following may be something to allow to be pluggable by different runtimes, as it is global mutable state.
var symtabMutex sync.Mutex
var symtab map[string]Value = initSymtableTable()

func initSymtableTable() map[string]Value {
	m := make(map[string]Value)
	//	m[TypeType.String()] = TypeType
	return m
}

func Symbols() []Value {
	syms := make([]Value, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func Intern(name string) Value {
	symtabMutex.Lock()
	defer symtabMutex.Unlock()
	sym, ok := symtab[name]
	if !ok {
		if IsValidKeywordName(name) {
			sym = &Keyword{Text: name}
		} else if IsValidTypeName(name) {
			sym = &Type{Text: name}
		} else if IsValidSymbolName(name) {
			sym = &Symbol{Text: name}
		} else {
			panic("invalid symbol/type/keyword name passed to intern: '" + name + "'")
		}
		symtab[name] = sym
	}
	return sym
}
