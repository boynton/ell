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

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type module interface {
	typeSymbol() lob
	String() string

	keywords() []lob
	globals() []lob
	global(sym lob) lob
	defGlobal(sym lob, val lob)
	setGlobal(sym lob, val lob) error
	macros() []lob
	macro(sym lob) lob
	defMacro(sym lob, val lob)

	define(name string, val lob)
	defineFunction(name string, fun primitive)
	defineMacro(name string, fun primitive)

	eval(expr lob) (lob, error)

	checkInterrupt() bool
	compileFile(filename string) (lob, error)
	loadFile(filename string) error
	loadModule(filename string) error
	exports() []lob
}

type lmodule struct {
	Name         string
	constantsMap map[lob]int
	constants    []lob
	globalMap    []*binding
	macroMap     map[lob]lob
	exported     []lob
	interrupts   chan os.Signal
}

type environmentInitializer func(mod module)

var initializer environmentInitializer

func newEnvironment(name string, init environmentInitializer, interrupts chan os.Signal) module {
	if initializer != nil {
		panic("Cannot define an environment twice.")
	}
	initializer = init
	return newModule(name, interrupts)
}

func (module *lmodule) checkInterrupt() bool {
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

func newModule(name string, interrupts chan os.Signal) module {
	constMap := make(map[lob]int, 0)
	var constants []lob
	globalMap := make([]*binding, 100)
	macroMap := make(map[lob]lob, 0)
	var exported []lob
	mod := lmodule{name, constMap, constants, globalMap, macroMap, exported, interrupts}
	if initializer != nil {
		initializer(&mod)
	}
	return &mod
}

func (module *lmodule) define(name string, obj lob) {
	module.defGlobal(intern(name), obj)
}

func (module *lmodule) defineFunction(name string, fun primitive) {
	sym := intern(name)
	if module.global(sym) != nil {
		println("*** Warning: redefining ", name)
	}
	prim := lprimitive{name, fun}
	module.defGlobal(sym, &prim)
}

func (module *lmodule) defineMacro(name string, fun primitive) {
	sym := intern(name)
	if module.macro(sym) != nil {
		println("*** Warning: redefining macro ", name)
	}
	prim := lprimitive{name, fun}
	module.defMacro(sym, &prim)
}

func (module *lmodule) typeSymbol() lob {
	return intern("module")
}

func (module *lmodule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", module.Name, module.constants)
}

func (module *lmodule) keywords() []lob {
	//keywords reserved for the base language that Ell compiles
	keywords := []lob{
		intern("quote"),
		intern("define"),
		intern("lambda"),
		intern("if"),
		intern("begin"),
		intern("define-macro"),
		intern("set!"),
		intern("lap"),
		intern("use"),
	}
	return keywords
}

func (module *lmodule) globals() []lob {
	syms := make([]lob, 0, symtag)
	for _, b := range module.globalMap {
		if b != nil {
			syms = append(syms, b.sym)
		}
	}
	return syms
}

func (module *lmodule) global(sym lob) lob {
	s := sym.(*lsymbol)
	if s.tag >= len(module.globalMap) {
		return nil
	}
	tmp := module.globalMap[s.tag]
	if tmp == nil {
		return nil
	}
	return tmp.val
}

type binding struct {
	sym lob
	val lob
}

func (module *lmodule) defGlobal(sym lob, val lob) {
	s := sym.(*lsymbol)
	if s.tag >= len(module.globalMap) {
		glob := make([]*binding, s.tag+100)
		copy(glob, module.globalMap)
		module.globalMap = glob
	}
	b := binding{sym, val}
	module.globalMap[s.tag] = &b
}

func (module *lmodule) setGlobal(sym lob, val lob) error {
	s := sym.(*lsymbol)
	if s.tag < len(module.globalMap) {
		if module.globalMap[s.tag] != nil {
			module.globalMap[s.tag].val = val
			return nil
		}

	}
	return newError("*** Warning: set on undefined global ", sym)
}

func (module *lmodule) macros() []lob {
	keys := make([]lob, 0, len(module.macroMap))
	for k := range module.macroMap {
		keys = append(keys, k)
	}
	return keys
}

func (module *lmodule) macro(sym lob) lob {
	mac, ok := module.macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func (module *lmodule) defMacro(sym lob, val lob) {
	module.macroMap[sym] = newMacro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func (module *lmodule) putConstant(val lob) int {
	idx, present := module.constantsMap[val]
	if !present {
		idx = len(module.constants)
		module.constants = append(module.constants, val)
		module.constantsMap[val] = idx
	}
	return idx
}

func (module *lmodule) use(sym lob) error {
	name := sym.String()
	return module.loadModule(name)
}

func (module *lmodule) importCode(thunk code) (lob, error) {
	moduleToUse := thunk.module()
	result, err := exec(thunk)
	if err != nil {
		return nil, err
	}
	exported := moduleToUse.exports()
	for _, sym := range exported {
		val := moduleToUse.global(sym)
		if val == nil {
			val = moduleToUse.macro(sym)
			module.defMacro(sym, (val.(*lmacro)).expander)
		} else {
			module.defGlobal(sym, val)
		}
	}
	return result, nil
}

func (module *lmodule) exports() []lob {
	return module.exported
}

func (module *lmodule) findModule(moduleName string) (string, error) {
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
		if fileReadable(filename) {
			return filename, nil
		}
		filename = filepath.Join(dirname, lname)
		if fileReadable(filename) {
			return filename, nil
		}
	}
	return "", newError("not found")
}

func (module *lmodule) loadModule(name string) error {
	file := name
	i := strings.Index(name, ".")
	if i < 0 {
		f, err := module.findModule(name)
		if err != nil {
			return newError("Module not found: ", name)
		}
		file = f
	} else {
		if !fileReadable(name) {
			return newError("Cannot read file: ", name)
		}
		name = name[0:i]

	}
	return module.loadFile(file)
}

func (module *lmodule) loadFile(file string) error {
	if verbose {
		println("; loadFile: " + file)
	}
	port, err := openInputFile(file)
	if err != nil {
		return err
	}

	expr, err := port.read()
	defer port.close()

	for {
		if err != nil {
			return err
		}
		if expr == EOF {
			return nil
		}
		_, err = module.eval(expr)
		if err != nil {
			return err
		}
		expr, err = port.read()
	}
}

func (module *lmodule) eval(expr lob) (lob, error) {
	if verbose {
		println("; eval: ", write(expr))
	}
	expanded, err := macroexpand(module, expr)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; expanded to: ", write(expanded))
	}
	code, err := compile(module, expanded)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; compiled to: ", write(code))
	}
	result, err := module.importCode(code)
	return result, err
}

//caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func (module *lmodule) compileFile(name string) (lob, error) {
	pretty := true
	//without macros, this used towork fine. Just wrap the file's expressions in a big begin, and compile it
	// this only makes sense for files that contain definitions only (not executions of those definitions)
	// i.e. it is harmless to execute them
	//
	file := name
	i := strings.Index(name, ".")
	if i < 0 {
		f, err := module.findModule(name)
		if err != nil {
			return nil, newError("Module not found: ", name)
		}
		file = f
	} else {
		if !fileReadable(name) {
			return nil, newError("Cannot read file: ", name)
		}
		name = name[0:i]

	}
	if verbose {
		println("; loadFile: " + file)
	}
	port, err := openInputFile(file)
	if err != nil {
		return nil, err
	}

	expr, err := port.read()
	defer port.close()
	result := ""
	if pretty {
		result = ";\n; code generated from " + file + "\n;\n"
	}
	for {
		if err != nil {
			return nil, err
		}
		if expr == EOF {
			return newString(result), nil
		}
		if verbose {
			println("; compile: ", write(expr))
		}
		expanded, err := macroexpand(module, expr)
		if err != nil {
			return nil, err
		}
		if verbose {
			println("; expanded to: ", write(expanded))
		}
		code, err := compile(module, expanded)
		if err != nil {
			return nil, err
		}
		if verbose {
			println("; compiled to: ", write(code))
		}
		if pretty {
			result = result + code.decompile(true) + "\n"
		} else {
			result = result + " " + code.decompile(true)
		}
		if false {
			//if the code contains macro defs, we need to run it. It may depend on other code. So, we run it all
			_, err = module.importCode(code)
			if err != nil {
				return nil, err
			}
		}
		expr, err = port.read()
	}
}
