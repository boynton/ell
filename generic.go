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

func methodSignature(formalArgs *LOB) (*LOB, error) {
	sig := ""
	for formalArgs != EmptyList {
		s := formalArgs.car //might be a symbol, might be a list
		tname := ""
		if s.variant == typeList { //specialized
			t := cadr(s)
			if t.variant != typeType {
				return nil, Error(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
			}
			tname = t.text
		} else if s.variant == typeSymbol { //unspecialized
			tname = "<any>"
		} else {
			return nil, Error(SyntaxErrorKey, "Specialized argument must be of the form <symbol> or (<symbol> <type>), got ", s)
		}
		sig += tname
		formalArgs = formalArgs.cdr
	}
	return intern(sig), nil
}

func arglistSignature(args []*LOB) string {
	sig := ""
	for _, arg := range args {
		sig += arg.variant.text
	}
	return sig
}

var typeAny = intern("<any>")

func signatureCombos(argtypes []*LOB) []string {
	switch len(argtypes) {
	case 0:
		return []string{}
	case 1:
		return []string{argtypes[0].text, typeAny.text}
	default:
		//get the combinations of the tail, and concat both the type and <any> onto each of those combos
		rest := signatureCombos(argtypes[1:]) // ["<number>" "<any>"]
		result := make([]string, 0, len(rest)*2)
		this := argtypes[0]
		for _, s := range rest {
			result = append(result, this.text+s)
		}
		for _, s := range rest {
			result = append(result, typeAny.text+s)
		}
		return result
	}
}

var cachedSigs = make(map[string][]*LOB)

func arglistSignatures(args []*LOB) []*LOB {
	key := arglistSignature(args)
	sigs, ok := cachedSigs[key]
	if !ok {
		var argtypes []*LOB
		for _, arg := range args {
			argtypes = append(argtypes, arg.variant)
		}
		stringSigs := signatureCombos(argtypes)
		sigs = make([]*LOB, 0, len(stringSigs))
		for _, sig := range stringSigs {
			sigs = append(sigs, intern(sig))
		}
		cachedSigs[key] = sigs
	}
	return sigs
}

var symGenfns = intern("*genfns*")
var keyMethods = intern("methods:")

func getfn(sym *LOB, args []*LOB) (*LOB, error) {
	sigs := arglistSignatures(args)
	gfs := global(symGenfns)
	if gfs != nil && gfs.variant == typeStruct {
		gf := structGet(gfs, sym)
		if gf == Null {
			return nil, Error(ErrorKey, "Not a generic function: ", sym)
		}
		methods := structGet(gf, keyMethods)
		if methods.variant == typeStruct {
			for _, sig := range sigs {
				fun := structGet(methods, sig)
				if fun != Null {
					return fun, nil
				}
			}
		}
	}
	return nil, Error(ErrorKey, "Generic function ", sym, ", has no matching method for: ", args)
}
