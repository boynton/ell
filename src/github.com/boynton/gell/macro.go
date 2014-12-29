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

package gell

import (
	"fmt"
)

type lmacro struct {
	name     LObject
	expander LObject //a function of one argument
}

func NewMacro(name LObject, expander LObject) LObject {
	mac := lmacro{name, expander}
	return &mac
}

func (*lmacro) Type() LObject {
	return Intern("macro")
}

func (macro *lmacro) String() string {
	return fmt.Sprintf("(macro %v %v)", macro.name, macro.expander)
}

func (macro *lmacro) Expand(expr LObject) (LObject, error) {
	expander := macro.expander
	switch fun := expander.(type) {
	case *lclosure:
		if fun.code.argc == 1 {
			expanded, err := Exec(fun.code, expr)
			if err != nil {
				return nil, Error("macro error in '", macro.name, "': ", err)
			}
			return expanded, nil
		}
	case *lprimitive:
		args := []LObject{expr}
		expanded, err := fun.fun(args, 1)
		if err != nil {
			return nil, Error("macro error in '", macro.name, "': ", err)
		}
		return expanded, nil
	}
	return nil, Error("Bad macro expander function")
}

func Macroexpand(module LModule, expr LObject) (LObject, error) {
	if IsPair(expr) {
		var head LObject = NIL
		fn := Car(expr)
		if IsSymbol(fn) {
			macro := module.Macro(fn)
			if macro != nil {
				return (macro.(*lmacro)).Expand(expr)
			} else {
				head = Cons(fn, NIL)
			}
		} else {
			expanded, err := Macroexpand(module, fn)
			if err != nil {
				return nil, err
			}
			head = Cons(expanded, NIL)
		}
		var tail LObject = head
		rest := Cdr(expr)
		for IsPair(rest) {
			expanded, err := Macroexpand(module, Car(rest))
			if err != nil {
				return nil, err
			}
			tmp := Cons(expanded, NIL)
			(tail.(*lpair)).cdr = tmp
			tail = tmp
			rest = Cdr(rest)
		}
		if rest != NIL {
			(tail.(*lpair)).cdr = rest
		}
		return head, nil
	}
	return expr, nil
}
