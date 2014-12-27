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
	Global(sym LObject) LObject
	DefGlobal(sym LObject, val LObject)
	SetGlobal(sym LObject, val LObject) error
	Exports() []LSymbol
}

type lmodule struct {
	Name         string
	constantsMap map[LObject]int
	constants    []LObject
	globals      map[LSymbol]LObject
	exports      []LSymbol
}

func newModule(name string, primitives map[string]Primitive) (LModule, error) {
	constMap := make(map[LObject]int, 0)
	constants := make([]LObject, 0)
	globals := make(map[LSymbol]LObject, 0)
	exports := make([]LSymbol, 0)
	mod := lmodule{name, constMap, constants, globals, exports}
	if primitives != nil {
		for name, fun := range primitives {
			mod.RegisterPrimitive(name, fun)
		}
	}
	return &mod, nil
}

func (module *lmodule) RegisterPrimitive(name string, fun Primitive) {
	//need the current module!!!
	sym := Intern(name)
	if module.Global(sym) != nil {
		Println("*** Warning: redefining ", name)
		//check the argument signature. Define "primitiveN" differently than "primitive0" .. "primitive3"
	}
	prim := lprimitive{name, fun}
	module.DefGlobal(sym, &prim)
}

func (module *lmodule) Type() LSymbol {
	return Intern("module")
}

func (module *lmodule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", module.Name, module.constants)
}

func (module *lmodule) Global(sym LObject) LObject {
	val, ok := module.globals[sym]
	if !ok {
		return nil
	}
	return val
	//	return (sym.(*lsymbol)).value
}

func (module *lmodule) DefGlobal(sym LObject, val LObject) {
	module.globals[sym] = val
}

func (module *lmodule) SetGlobal(sym LObject, val LObject) error {
	val, ok := module.globals[sym]
	if !ok {
		return Error("*** Warning: set on undefined global ", sym)
	}
	module.globals[sym] = val
	return nil
}

func (module *lmodule) Exports() []LSymbol {
	return module.exports
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare...
//idea: I'd like all Ell objects to have a hashcode. Use that.
func (module *lmodule) putConstant(val LObject) int {
	idx, present := module.constantsMap[val]
	if !present {
		idx = len(module.constants)
		module.constants = append(module.constants, val)
		module.constantsMap[val] = idx
	}
	return idx
}

var verbose bool

func SetVerbose(b bool) {
	verbose = b
}

func (mod *lmodule) xExec(thunk LCode) (LObject, error) {
	code := thunk.(*lcode)
	vm := NewVM(DefaultStackSize)
	result, err := vm.exec(code)
	if err != nil {
		return NIL, err
	}
	exported := vm.Exported()
	if len(exported) > 0 {
		if verbose {
			Println("export these: ", exported)
		}
		//		for _, sym := range exported {
		//
		//		}
		//set up the module's exports
	}
	return result, nil
}

func RunModule(name string, primitives map[string]Primitive) (LObject, error) {
	thunk, err := LoadModule(name, primitives)
	if err != nil {
		return NIL, err
	}
	return Exec(thunk)
}

func FindModule(moduleName string) (string, error) {
	path := [...]string{".", "src/main/ell"} //fix
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

func LoadModule(name string, primitives map[string]Primitive) (LCode, error) {
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

func LoadFileModule(moduleName string, file string, primitives map[string]Primitive) (LCode, error) {
	if verbose {
		Println("; loadModule: " + moduleName + " from " + file)
	}
	module, err := newModule(moduleName, primitives)
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
	if verbose {
		Println("; read: ", Write(source))
	}
	if Length(source) == 2 {
		source = Cadr(source)
	}
	code, err := Compile(module, source)
	if err != nil {
		return nil, err
	}
	if verbose {
		Println("; compiled to: ", code)
		Println("; module: ", module)
	}
	return code, nil
}
