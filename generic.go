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

import(
	. "github.com/boynton/ell/data"
)

func methodSignature(formalArgs *List) (Value, error) {
	sig := ""
	for formalArgs != EmptyList {
		s := formalArgs.Car //might be a symbol, might be a list
		tname := ""
		if lst, ok := s.(*List); ok { //specialized
			t := Cadr(lst)
			if tp, ok := t.(*Type); ok {
				tname = tp.Name()
			} else {
				return nil, NewError(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
			}
		} else if s.Type() == SymbolType { //unspecialized
			tname = "<any>"
		} else {
			return nil, NewError(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
		}
		sig += tname
		formalArgs = formalArgs.Cdr
	}
	return Intern(sig), nil
}

func arglistSignature(args []Value) string {
	sig := ""
	for _, arg := range args {
		sig += (arg.Type().(*Type)).Name()
	}
	return sig
}

func signatureCombos(argtypes []Value) []string {
	switch len(argtypes) {
	case 0:
		return []string{}
	case 1:
		return []string{TypeNameString(argtypes[0]), TypeNameString(AnyType)}
	default:
		//get the combinations of the tail, and concat both the type and <any> onto each of those combos
		rest := signatureCombos(argtypes[1:]) // ["<number>" "<any>"]
		result := make([]string, 0, len(rest)*2)
		this := argtypes[0]
		for _, s := range rest {
			result = append(result, TypeNameString(this)+s)
		}
		for _, s := range rest {
			result = append(result, TypeNameString(AnyType)+s)
		}
		return result
	}
}

func TypeNameString(tval Value) string {
	if tp, ok := tval.(*Type); ok {
		return tp.Name()
	}
	return ""
}

var cachedSigs = make(map[string][]Value)

func arglistSignatures(args []Value) []Value {
	key := arglistSignature(args)
	sigs, ok := cachedSigs[key]
	if !ok {
		var argtypes []Value
		for _, arg := range args {
			argtypes = append(argtypes, arg.Type())
		}
		stringSigs := signatureCombos(argtypes)
		sigs = make([]Value, 0, len(stringSigs))
		for _, sig := range stringSigs {
			sigs = append(sigs, Intern(sig))
		}
		cachedSigs[key] = sigs
	}
	return sigs
}

var GenfnsSymbol = Intern("*genfns*")
var MethodsKeyword = Intern("methods:")

func getfn(sym Value, args []Value) (Value, error) {
	sigs := arglistSignatures(args)
	gfs := GetGlobal(GenfnsSymbol)
	if p, ok := gfs.(*Struct); ok {
		gf := p.Get(sym)
		if p2, ok := gf.(*Instance); ok {
			if p3, ok := p2.Value.(*Struct); ok {
				methods := p3.Get(MethodsKeyword)
				if p3, ok := methods.(*Struct); ok {
					for _, sig := range sigs {
						fun := p3.Get(sig)
						if fun != Null {
							return fun, nil
						}
					}
				}
			}
		} else {
			return nil, NewError(ErrorKey, "Not a generic function: ", sym)
		}
	}
	return nil, NewError(ErrorKey, "Generic function ", sym, ", has no matching method for: ", args)
}
