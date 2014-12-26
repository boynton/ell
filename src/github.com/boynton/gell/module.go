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
	"path/filepath"
	"strings"
)


type LModule interface {
	Type() LSymbol
	String() string
	RegisterPrimitive(name string, fun Primitive)
	Global(sym LSymbol) (LObject, bool)
}

type tModule struct {
	Name         string
	globals      map[LSymbol]LObject
	constantsMap map[LObject]int
	constants    []LObject
	exports      []LObject
}

func (tModule) Type() LSymbol {
	return Intern("module")
}

func (module tModule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", module.Name, module.constants)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare...
//idea: I'd like all Ell objects to have a hashcode. Use that.
func (module *tModule) putConstant(val LObject) int {
	idx, present := module.constantsMap[val]
	if !present {
		idx = len(module.constants)
		module.constants = append(module.constants, val)
		module.constantsMap[val] = idx
	}
	return idx
}

type Primitive func(argv []LObject, argc int) (LObject, LError)

type tPrimitive struct {
	name string
	fun  Primitive
}

var symPrimitive = newSymbol("primitive")

func (prim tPrimitive) Type() LSymbol {
	return symPrimitive
}

func (prim tPrimitive) String() string {
	return "<primitive " + prim.name + ">"
}

func (module tModule) globalRef(idx int) (LObject, bool) {
	sym := module.constants[idx]
	v := (sym.(*lsymbol)).value
	if v == nil {
		return nil, false
	}
	return v, true
}

func (module tModule) globalDefine(idx int, obj LObject) LObject {
	sym := module.constants[idx]
	(sym.(*lsymbol)).value = obj
	return sym
}

func (module tModule) Global(sym LSymbol) (LObject, bool) {
	v := (sym.(*lsymbol)).value
	b := v != nil
	return v, b
}

func (module tModule) SetGlobal(sym LSymbol, val LObject) {
	(sym.(*lsymbol)).value = val
}

func (module tModule) RegisterPrimitive(name string, fun Primitive) {
	sym := Intern(name)
	_, ok := module.Global(sym)
	if ok {
		Println("*** Warning: redefining ", name)
		//check the argument signature. Define "primitiveN" differently than "primitive0" .. "primitive3"
	}
	po := tPrimitive{name, fun}
	module.SetGlobal(sym, po)
}

func MakeModule(name string, primitives Primitives) (LModule, error) {
	globals := map[LSymbol]LObject{}
	constMap := map[LObject]int{}
	constants := make([]LObject, 0)
	mod := tModule{name, globals, constMap, constants, nil}
	if primitives != nil {
		err := primitives.Init(mod)
		if err != nil {
			return mod, nil
		}
	}
	return &mod, nil
}



type Primitives interface {
	Init(module LModule) error
}

func RunModule(name string, primitives Primitives) (LObject, error) {
	thunk, err := LoadModule(name, primitives)
	if err != nil {
		return NIL, err
	}
	code := thunk.(*lcode)
	Println("; begin execution")
	vm := LVM{DefaultStackSize, make([]LSymbol, 0)}
	result, err := vm.exec(code)
	if err != nil {
		return NIL, err
	}
	if len(vm.defs) > 0 {
		Println("export these: ", vm.defs)
	}
	if result != nil {
		Println("; end execution")
		Println("; => ", result)
	}
	return result, nil
}

func FindModule(moduleName string) (string, error) {
	path := [...]string{"src/main/ell"} //fix
	name := moduleName
	if !strings.HasSuffix(name, ".ell") {
		name = name + ".ell"
	}
	for _, dirname := range path {
		filename := filepath.Join(dirname, name)
		if FileReadable(filename) {
			return filename, nil
		}
	}
	return "", Error("not found")
}

func LoadModule(name string, primitives Primitives) (LCode, error) {
	file := name
	i := strings.Index(name, ".")
	if i < 0 {
		f, err := FindModule(name)
		if err != nil {
			return nil, Error("Module not found:", name)
		}
		file = f
	} else {
		if !FileReadable(name) {
			return nil, Error("Cannot read file:", name)
		}
		name = name[0:i]

	}
	return LoadFileModule(name, file, primitives)
}

func LoadFileModule(moduleName string, file string, primitives Primitives) (LCode, error) {
	Println("; loadModule: " + moduleName + " from " + file)
	module, err := MakeModule(moduleName, primitives)
	if err != nil {
		return nil, err
	}
	port, err := OpenInputFile(file)
	if err != nil {
		return nil, err
	}
	source := List(Intern("begin"))
	expr, err := port.Read()
	for {
		if err != nil {
			return nil, err
		}
		if expr == EOI {
			break
		}
		source, err = Concat(source, List(expr))
		if err == nil {
			expr, err = port.Read()
		}
	}
	port.Close()
	Println("; read: ", Write(source))
	code, err := Compile(module, source)
	if err != nil {
		return nil, err
	}
	Println("; compiled to: ", Write(code))
	Println("; module: ", module)
	return code, nil
}
