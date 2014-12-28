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
	"os"
	"path/filepath"
	"strings"
)

type LModule interface {
	Type() LSymbol
	String() string
	Global(sym LObject) LObject
	DefGlobal(sym LObject, val LObject)
	SetGlobal(sym LObject, val LObject) error
	Macro(sym LObject) LMacro
	DefMacro(sym LObject, val LObject)
	CompileFile(filename string) (LObject, error)
	LoadFile(filename string) error
	LoadModule(filename string) error
	Exports() []LSymbol
}

type lmodule struct {
	Name         string
	constantsMap map[LObject]int
	constants    []LObject
	globals      []LSymbol
	macros       map[LSymbol]LMacro
	exports      []LSymbol
}

var primitives map[string]Primitive
var mprimitives map[string]Primitive

func NewEnvironment(name string, prims map[string]Primitive, mprims map[string]Primitive) LModule {
	if primitives != nil {
		panic("Cannot define an environment twice.")
	}
	primitives = prims
	mprimitives = mprims
	return newModule(name)
}

func newModule(name string) LModule {
	constMap := make(map[LObject]int, 0)
	constants := make([]LObject, 0)
	globals := make([]LSymbol, 100) //!
	macros := make(map[LSymbol]LMacro, 0)
	exports := make([]LSymbol, 0)
	mod := lmodule{name, constMap, constants, globals, macros, exports}
	if primitives != nil {
		for name, fun := range primitives {
			mod.RegisterPrimitive(name, fun)
		}
	}
	if mprimitives != nil {
		for name, fun := range mprimitives {
			mod.RegisterPrimitiveMacro(name, fun)
		}
	}
	return &mod
}

func (module *lmodule) RegisterPrimitive(name string, fun Primitive) {
	sym := Intern(name)
	if module.Global(sym) != nil {
		Println("*** Warning: redefining ", name)
	}
	prim := lprimitive{name, fun}
	module.DefGlobal(sym, &prim)
}

func (module *lmodule) RegisterPrimitiveMacro(name string, fun Primitive) {
	sym := Intern(name)
	if module.Macro(sym) != nil {
		Println("*** Warning: redefining macro ", name)
	}
	prim := lprimitive{name, fun}
	module.DefMacro(sym, &prim)
}

func (module *lmodule) Type() LSymbol {
	return Intern("module")
}

func (module *lmodule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", module.Name, module.constants)
}

func (module *lmodule) Global(sym LObject) LObject {
	s := sym.(*lsymbol)
	if s.tag >= len(module.globals) {
		return nil
	}
	return module.globals[s.tag]
}

func (module *lmodule) DefGlobal(sym LObject, val LObject) {
	s := sym.(*lsymbol)
	if s.tag >= len(module.globals) {
		glob := make([]LSymbol, s.tag+100)
		copy(glob, module.globals)
		module.globals = glob
	}
	module.globals[s.tag] = val
}

func (module *lmodule) SetGlobal(sym LObject, val LObject) error {
	s := sym.(*lsymbol)
	if s.tag < len(module.globals) {
		if module.globals[s.tag] != nil {
			module.globals[s.tag] = val
			return nil
		}

	}
	return Error("*** Warning: set on undefined global ", sym)
}

func (module *lmodule) Macro(sym LObject) LMacro {
	mac, ok := module.macros[sym]
	if !ok {
		return nil
	}
	return mac
}

func (module *lmodule) DefMacro(sym LObject, val LObject) {
	module.macros[sym] = NewMacro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
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

func (module *lmodule) Use(sym LSymbol) error {
	name := sym.String()
	return module.LoadModule(name)
}

func (module *lmodule) Import(thunk LCode) error {
	moduleToUse := thunk.Module()
	_, err := Exec(thunk)
	if err != nil {
		return err
	}
	exports := moduleToUse.Exports()
	for _, sym := range exports {
		val := moduleToUse.Global(sym)
		if val == nil {
			val = moduleToUse.Macro(sym)
			module.DefMacro(sym, (val.(*lmacro)).expander)
		} else {
			module.DefGlobal(sym, val)
		}
	}
	return nil
}

func (module *lmodule) Exports() []LSymbol {
	return module.exports
}

func (module *lmodule) FindModule(moduleName string) (string, error) {
	var path []string
	spath := os.Getenv("ELL_PATH")
	if spath != "" {
		path = strings.Split(spath, ":")
	} else {
		path = []string{".", "tests/"}
	}
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

func (module *lmodule) LoadModule(name string) error {
	file := name
	i := strings.Index(name, ".")
	if i < 0 {
		f, err := module.FindModule(name)
		if err != nil {
			return Error("Module not found: ", name)
		}
		file = f
	} else {
		if !FileReadable(name) {
			return Error("Cannot read file: ", name)
		}
		name = name[0:i]

	}
	return module.LoadFile(file)
}

func (module *lmodule) LoadFile(file string) error {
	if verbose {
		Println("; loadFile: " + file)
	}
	port, err := OpenInputFile(file)
	if err != nil {
		return err
	}

	expr, err := port.Read()
	defer port.Close()

	for {
		if err != nil {
			return err
		}
		if expr == EOI {
			return nil
		}
		if verbose {
			Println("; read: ", Write(expr))
		}
		expanded, err := Macroexpand(module, expr)
		if err != nil {
			return err
		}
		if verbose {
			Println("; expanded to: ", expanded)
		}
		code, err := Compile(module, expanded)
		if err != nil {
			return err
		}
		if verbose {
			Println("; compiled to: ", code)
		}
		err = module.Import(code)
		if err != nil {
			return err
		}
		expr, err = port.Read()
	}
}

func (module *lmodule) CompileFile(file string) (LObject, error) {
	return nil, Error("CompileFile NYI")
}
