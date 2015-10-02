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
		gf = value(gf)
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

func isEmpty(obj *LOB) bool {
	seq := value(obj)
	switch seq.variant {
	case typeList:
		return seq == EmptyList
	case typeString:
		return len(seq.text) == 0
	case typeVector:
		return len(seq.elements) == 0
	case typeStruct:
		return len(seq.elements) == 0
	case typeNull: //?
		return true
	}
	return false
}

func length(obj *LOB) int {
	seq := value(obj)
	switch seq.variant {
	case typeString:
		return stringLength(seq.text)
	case typeVector:
		return len(seq.elements)
	case typeList:
		return listLength(seq)
	case typeStruct:
		return len(seq.elements) / 2
	default:
		return -1
	}
}

func reverse(lst *LOB) *LOB {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func flatten(lst *LOB) *LOB {
	result := EmptyList
	tail := EmptyList
	for lst != EmptyList {
		item := lst.car
		switch item.variant {
		case typeList:
			item = flatten(item)
		case typeVector:
			litem, _ := toList(item)
			item = flatten(litem)
		default:
			item = list(item)
		}
		if tail == EmptyList {
			result = item
			tail = result
		} else {
			tail.cdr = item
		}
		for tail.cdr != EmptyList {
			tail = tail.cdr
		}
		lst = lst.cdr
	}
	return result
}

func concat(seq1 *LOB, seq2 *LOB) (*LOB, error) {
	rev := reverse(seq1)
	if rev == EmptyList {
		return seq2, nil
	}
	lst := seq2
	for rev != EmptyList {
		lst = cons(rev.car, lst)
		rev = rev.cdr
	}
	return lst, nil
}

func nextRangeValue(val int, end int, step int) (int, bool) {
	val += step
	return val, isValidRange(val, end, step)
}

func isValidRange(val int, end int, step int) bool {
	if step >= 0 {
		if val < end {
			return true
		}
	} else {
		if val > end {
			return true
		}
	}
	return false
}

func assoc(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return assocStruct(s, rest)
	case typeVector:
		return assocVector(s, rest)
	default:
		return nil, Error(ArgumentErrorKey, "Expected a <struct>|<vector>, got a ", obj.variant)
	}
}

func dissoc(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return dissocStruct(s, rest)
	default:
		return nil, Error(ArgumentErrorKey, "Expected a <struct>|<vector>, got a ", obj.variant)
	}
}

func dissocBang(obj *LOB, rest ...*LOB) (*LOB, error) {
	s := value(obj)
	switch s.variant {
	case typeStruct:
		return dissocBangStruct(s, rest)
	default:
		return nil, Error(ArgumentErrorKey, "Expected a <struct>|<vector>, got a ", obj.variant)
	}
}
