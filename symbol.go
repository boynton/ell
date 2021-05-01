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
	"fmt"
	
	. "github.com/boynton/ell/data"
)

func NewSymbol(names []Value) (Value, error) {
	size := len(names)
	if size < 1 {
		return nil, NewError(ArgumentErrorKey, "symbol expected at least 1 argument, got none")
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		switch p := o.(type) {
		case *String:
			name += p.Value
		case *Symbol:
			name += p.Text
		default:
			fmt.Println("o:", o)
			return nil, NewError(ArgumentErrorKey, "symbol name component invalid: ", o)
		}
	}
	return Intern(name), nil
}

func SymbolName(obj Value) string {
	if sym, ok := obj.(*Symbol); ok {
		return sym.Name()
	}
	return ""
}

func IsSymbol(obj Value) bool {
	if _, ok := obj.(*Symbol); ok {
		return true
	}
	return false
}

func Unkeyworded(symOrKeyword Value) (Value, error) {
	if _, ok := symOrKeyword.(*Symbol); ok {
		return symOrKeyword, nil
	}
	if kw, ok := symOrKeyword.(*Keyword); ok {
		return Intern(kw.Name()), nil
	}
    return Null, NewError(ArgumentErrorKey, "Expected <keyword> or <symbol>, got ", symOrKeyword.Type())
}
