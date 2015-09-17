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
	Name() string
	String() string

	keywords() []lob
	globals() []lob
	global(sym lob) lob
	isDefined(sym lob) bool
	defGlobal(sym lob, val lob)
	undefGlobal(sym lob)
	setGlobal(sym lob, val lob) error
	macros() []lob
	macro(sym lob) *macro
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
	name         string
	constantsMap map[lob]int
	constants    []lob
	globalMap    []*binding
	macroMap     map[lob]*macro
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

func (mod *lmodule) checkInterrupt() bool {
	if mod.interrupts != nil {
		select {
		case msg := <-mod.interrupts:
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
	macroMap := make(map[lob]*macro, 0)
	var exported []lob
	mod := lmodule{name, constMap, constants, globalMap, macroMap, exported, interrupts}
	if initializer != nil {
		initializer(&mod)
	}
	return &mod
}

func (mod *lmodule) define(name string, obj lob) {
	mod.defGlobal(intern(name), obj)
}

func (mod *lmodule) defineFunction(name string, fun primitive) {
	sym := intern(name)
	if mod.global(sym) != nil {
		println("*** Warning: redefining ", name)
	}
	prim := lprimitive{name, fun}
	mod.defGlobal(sym, &prim)
}

func (mod *lmodule) defineMacro(name string, fun primitive) {
	sym := intern(name)
	if mod.macro(sym) != nil {
		println("*** Warning: redefining macro ", name)
	}
	prim := lprimitive{name, fun}
	mod.defMacro(sym, &prim)
}

func (mod *lmodule) typeSymbol() lob {
	return intern("module")
}

func (mod *lmodule) Name() string {
	return mod.name
}

func (mod *lmodule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", mod.Name, mod.constants)
}

func (mod *lmodule) keywords() []lob {
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

func (mod *lmodule) globals() []lob {
	syms := make([]lob, 0, symtag)
	for _, b := range mod.globalMap {
		if b != nil {
			syms = append(syms, b.sym)
		}
	}
	return syms
}

func (mod *lmodule) global(sym lob) lob {
	s := sym.(*lsymbol)
	if s.tag >= len(mod.globalMap) {
		return nil
	}
	tmp := mod.globalMap[s.tag]
	if tmp == nil {
		return nil
	}
	return tmp.val
}

type binding struct {
	sym lob
	val lob
}

func (mod *lmodule) defGlobal(sym lob, val lob) {
	s := sym.(*lsymbol)
	if s.tag >= len(mod.globalMap) {
		glob := make([]*binding, s.tag+100)
		copy(glob, mod.globalMap)
		mod.globalMap = glob
	}
	b := binding{sym, val}
	mod.globalMap[s.tag] = &b
}

func (mod *lmodule) isDefined(sym lob) bool {
	s := sym.(*lsymbol)
	if s.tag < len(mod.globalMap) {
		return mod.globalMap[s.tag] != nil
	}
	return false
}

func (mod *lmodule) undefGlobal(sym lob) {
	s := sym.(*lsymbol)
	if s.tag < len(mod.globalMap) {
		mod.globalMap[s.tag] = nil
	}
}

func (mod *lmodule) setGlobal(sym lob, val lob) error {
	s := sym.(*lsymbol)
	if s.tag < len(mod.globalMap) {
		if mod.globalMap[s.tag] != nil {
			mod.globalMap[s.tag].val = val
			return nil
		}

	}
	return newError("*** Warning: set on undefined global ", sym)
}

func (mod *lmodule) macros() []lob {
	keys := make([]lob, 0, len(mod.macroMap))
	for k := range mod.macroMap {
		keys = append(keys, k)
	}
	return keys
}

func (mod *lmodule) macro(sym lob) *macro {
	mac, ok := mod.macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func (mod *lmodule) defMacro(sym lob, val lob) {
	mod.macroMap[sym] = newMacro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func (mod *lmodule) putConstant(val lob) int {
	idx, present := mod.constantsMap[val]
	if !present {
		idx = len(mod.constants)
		mod.constants = append(mod.constants, val)
		mod.constantsMap[val] = idx
	}
	return idx
}

func (mod *lmodule) use(sym lob) error {
	name := sym.String()
	return mod.loadModule(name)
}

func (mod *lmodule) importCode(thunk code) (lob, error) {
	moduleToUse := thunk.module()
	result, err := exec(thunk)
	if err != nil {
		return nil, err
	}
	exported := moduleToUse.exports()
	for _, sym := range exported {
		val := moduleToUse.global(sym)
		if val == nil { //shouldn't syntax take priority over global defs?
			mac := moduleToUse.macro(sym)
			mod.defMacro(sym, mac.expander)
		} else {
			mod.defGlobal(sym, val)
		}
	}
	return result, nil
}

func (mod *lmodule) exports() []lob {
	return mod.exported
}

func (mod *lmodule) findModuleByName(moduleName string) (string, error) {
	path := strings.Split(EllPath, ":")
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
	return "", newError("Module not found: ", moduleName)
}

func (mod *lmodule) loadModule(name string) error {
	file := name
	if !fileReadable(name) {
		f, err := mod.findModuleFile(name)
		if err != nil {
			return err
		}
		file = f
	}
	return mod.loadFile(file)
}

func (mod *lmodule) loadFile(file string) error {
	if verbose {
		println("; loadFile: " + file)
	} else {
		println("[loading " + file + "]")
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
		_, err = mod.eval(expr)
		if err != nil {
			return err
		}
		expr, err = port.read()
	}
}

func (mod *lmodule) eval(expr lob) (lob, error) {
	if verbose {
		println("; eval: ", write(expr))
	}
	expanded, err := macroexpandObject(mod, expr)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; expanded to: ", write(expanded))
	}
	code, err := compile(mod, expanded)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; compiled to: ", write(code))
	}
	result, err := mod.importCode(code)
	return result, err
}

func (mod *lmodule) findModuleFile(name string) (string, error) {
	i := strings.Index(name, ".")
	if i < 0 {
		file, err := mod.findModuleByName(name)
		if err != nil {
			return "", err
		}
		return file, nil
	}
	if !fileReadable(name) {
		return "", newError("Cannot read file: ", name)
	}
	return name, nil
}

func (mod *lmodule) compileExpr(expr lob) (string, error) {
	if verbose {
		println("; compile: ", write(expr))
	}
	expanded, err := macroexpandObject(mod, expr)
	if err != nil {
		return "", err
	}
	if verbose {
		println("; expanded to: ", write(expanded))
	}
	code, err := compile(mod, expanded)
	if err != nil {
		return "", err
	}
	if verbose {
		println("; compiled to: ", write(code))
	}
	return code.decompile(true) + "\n", nil
}

//caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func (mod *lmodule) compileFile(name string) (lob, error) {
	file, err := mod.findModuleFile(name)
	if err != nil {
		return nil, err
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
	result := ";\n; code generated from " + file + "\n;\n"
	var lap string
	for err == nil {
		if expr == EOF {
			return lstring(result), nil
		}
		lap, err = mod.compileExpr(expr)
		if err != nil {
			return nil, err
		}
		result += lap
		expr, err = port.read()
	}
	return nil, err
}
