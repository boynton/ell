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
	Type() LObject
	String() string

	Keywords() []LObject
	Globals() []LObject
	Global(sym LObject) LObject
	DefGlobal(sym LObject, val LObject)
	SetGlobal(sym LObject, val LObject) error
	Macros() []LObject
	Macro(sym LObject) LObject
	DefMacro(sym LObject, val LObject)

	Define(name string, val LObject)
	DefineFunction(name string, fun Primitive)
	DefineMacro(name string, fun Primitive)

	Eval(expr LObject) (LObject, error)

	CheckInterrupt() bool
	CompileFile(filename string) (LObject, error)
	LoadFile(filename string) error
	LoadModule(filename string) error
	Exports() []LObject
}

type lmodule struct {
	Name         string
	constantsMap map[LObject]int
	constants    []LObject
	globals      []*binding
	macros       map[LObject]LObject
	exports      []LObject
	interrupts   chan os.Signal
}

type EnvironmentInitializer func(mod LModule)

var initializer EnvironmentInitializer

//var primitives map[string]Primitive
//var mprimitives map[string]Primitive

func NewEnvironment(name string, init EnvironmentInitializer, interrupts chan os.Signal) LModule {
	if initializer != nil {
		panic("Cannot define an environment twice.")
	}
	initializer = init
	return newModule(name, interrupts)
}

func (module *lmodule) CheckInterrupt() bool {
	if module.interrupts != nil {
		select {
		case msg := <-module.interrupts:
			return msg != nil
		default:
			return false
		}
	}
	return false
}

func newModule(name string, interrupts chan os.Signal) LModule {
	constMap := make(map[LObject]int, 0)
	constants := make([]LObject, 0)
	globals := make([]*binding, 100)
	macros := make(map[LObject]LObject, 0)
	exports := make([]LObject, 0)
	mod := lmodule{name, constMap, constants, globals, macros, exports, interrupts}
	if initializer != nil {
		initializer(&mod)
	}
	return &mod
}

func (module *lmodule) Define(name string, obj LObject) {
	module.DefGlobal(Intern(name), obj)
}

func (module *lmodule) DefineFunction(name string, fun Primitive) {
	sym := Intern(name)
	if module.Global(sym) != nil {
		Println("*** Warning: redefining ", name)
	}
	prim := lprimitive{name, fun}
	module.DefGlobal(sym, &prim)
}

func (module *lmodule) DefineMacro(name string, fun Primitive) {
	sym := Intern(name)
	if module.Macro(sym) != nil {
		Println("*** Warning: redefining macro ", name)
	}
	prim := lprimitive{name, fun}
	module.DefMacro(sym, &prim)
}

func (module *lmodule) Type() LObject {
	return Intern("module")
}

func (module *lmodule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", module.Name, module.constants)
}

func (module *lmodule) Keywords() []LObject {
	//keywords reserved for the base language that Ell compiles
	keywords := []LObject{
		Intern("quote"),
		Intern("define"),
		Intern("lambda"),
		Intern("if"),
		Intern("begin"),
		Intern("define-macro"),
		Intern("set!"),
		Intern("lap"),
		Intern("use"),
	}
	return keywords
}

func (module *lmodule) Globals() []LObject {
	syms := make([]LObject, 0, symtag)
	for _, b := range module.globals {
		if b != nil {
			syms = append(syms, b.sym)
		}
	}
	return syms
}

func (module *lmodule) Global(sym LObject) LObject {
	s := sym.(*lsymbol)
	if s.tag >= len(module.globals) {
		return nil
	}
	tmp := module.globals[s.tag]
	if tmp == nil {
		return nil
	}
	return tmp.val
}

type binding struct {
	sym LObject
	val LObject
}

func (module *lmodule) DefGlobal(sym LObject, val LObject) {
	s := sym.(*lsymbol)
	if s.tag >= len(module.globals) {
		glob := make([]*binding, s.tag+100)
		copy(glob, module.globals)
		module.globals = glob
	}
	b := binding{sym, val}
	module.globals[s.tag] = &b
}

func (module *lmodule) SetGlobal(sym LObject, val LObject) error {
	s := sym.(*lsymbol)
	if s.tag < len(module.globals) {
		if module.globals[s.tag] != nil {
			module.globals[s.tag].val = val
			return nil
		}

	}
	return Error("*** Warning: set on undefined global ", sym)
}

func (module *lmodule) Macros() []LObject {
	keys := make([]LObject, 0, len(module.macros))
	for k := range module.macros {
		keys = append(keys, k)
	}
	return keys
}

func (module *lmodule) Macro(sym LObject) LObject {
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

func (module *lmodule) Use(sym LObject) error {
	name := sym.String()
	return module.LoadModule(name)
}

func (module *lmodule) Import(thunk LCode) (LObject, error) {
	moduleToUse := thunk.Module()
	result, err := Exec(thunk)
	if err != nil {
		return nil, err
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
	return result, nil
}

func (module *lmodule) Exports() []LObject {
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
	lname := moduleName
	if !strings.HasSuffix(name, ".ell") {
		name = name + ".ell"
		lname = moduleName + ".lap"
	}
	for _, dirname := range path {
		filename := filepath.Join(dirname, name)
		if FileReadable(filename) {
			return filename, nil
		} else {
			filename = filepath.Join(dirname, lname)
			if FileReadable(filename) {
				return filename, nil
			}
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
		_, err = module.Eval(expr)
		if err != nil {
			return err
		}
		expr, err = port.Read()
	}
}

func (module *lmodule) Eval(expr LObject) (LObject, error) {
	if verbose {
		Println("; eval: ", Write(expr))
	}
	expanded, err := Macroexpand(module, expr)
	if err != nil {
		return nil, err
	}
	if verbose {
		Println("; expanded to: ", Write(expanded))
	}
	code, err := Compile(module, expanded)
	if err != nil {
		return nil, err
	}
	if verbose {
		Println("; compiled to: ", Write(code))
	}
	result, err := module.Import(code)
	return result, err
}

//caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func (module *lmodule) CompileFile(name string) (LObject, error) {
	pretty := true
	//without macros, this used towork fine. Just wrap the file's expressions in a big begin, and compile it
	// this only makes sense for files that contain definitions only (not executions of those definitions)
	// i.e. it is harmless to execute them
	//
	file := name
	i := strings.Index(name, ".")
	if i < 0 {
		f, err := module.FindModule(name)
		if err != nil {
			return nil, Error("Module not found: ", name)
		}
		file = f
	} else {
		if !FileReadable(name) {
			return nil, Error("Cannot read file: ", name)
		}
		name = name[0:i]

	}
	if verbose {
		Println("; loadFile: " + file)
	}
	port, err := OpenInputFile(file)
	if err != nil {
		return nil, err
	}

	expr, err := port.Read()
	defer port.Close()
	result := ""
	if pretty {
		result = ";\n; code generated from " + file + "\n;\n"
	}
	for {
		if err != nil {
			return nil, err
		}
		if expr == EOI {
			return NewString(result), nil
		}
		if verbose {
			Println("; compile: ", Write(expr))
		}
		expanded, err := Macroexpand(module, expr)
		if err != nil {
			return nil, err
		}
		if verbose {
			Println("; expanded to: ", Write(expanded))
		}
		code, err := Compile(module, expanded)
		if err != nil {
			return nil, err
		}
		if verbose {
			Println("; compiled to: ", Write(code))
		}
		if pretty {
			result = result + code.Decompile(true) + "\n"
		} else {
			result = result + " " + code.Decompile(true)
		}
		if false {
			//if the code contains macro defs, we need to run it. It may depend on other code. So, we run it all
			_, err = module.Import(code)
			if err != nil {
				return nil, err
			}
		}
		expr, err = port.Read()
	}
}
