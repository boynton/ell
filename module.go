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

// Module - a package of code. Every module has its own namespace that symbols, types, and kerywords are interned into
type Module struct {
	name         string
	constantsMap map[AnyType]int
	constants    []AnyType
	globalMap    []*binding
	macroMap     map[AnyType]*macro
	exported     []AnyType
	interrupts   chan os.Signal
}

type environmentInitializer func(mod *Module)

var initializer environmentInitializer

func newEnvironment(name string, init environmentInitializer, interrupts chan os.Signal) *Module {
	if initializer != nil {
		panic("Cannot define an environment twice.")
	}
	initializer = init
	return newModule(name, interrupts)
}

func (mod *Module) checkInterrupt() bool {
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

func newModule(name string, interrupts chan os.Signal) *Module {
	constMap := make(map[AnyType]int, 0)
	var constants []AnyType
	globalMap := make([]*binding, 100)
	macroMap := make(map[AnyType]*macro, 0)
	var exported []AnyType
	mod := Module{name, constMap, constants, globalMap, macroMap, exported, interrupts}
	if initializer != nil {
		initializer(&mod)
	}
	return &mod
}

func (mod *Module) define(name string, obj AnyType) {
	sym := intern(name)
	if sym == nil {
		panic("Cannot define a value for this symbol: " + name)
	}
	mod.defGlobal(sym, obj)
}

func (mod *Module) defineFunction(name string, fun primitive) {
	sym := intern(name)
	if mod.global(sym) != nil {
		println("*** Warning: redefining ", name)
	}
	prim := Primitive{name, fun}
	mod.defGlobal(sym, &prim)
}

func (mod *Module) defineMacro(name string, fun primitive) {
	sym := intern(name)
	if mod.macro(sym) != nil {
		println("*** Warning: redefining macro ", name, " -> ", mod.macro(sym))
	}
	prim := Primitive{name, fun}
	mod.defMacro(sym, &prim)
}

// Name - the name fo the module
func (mod *Module) Name() string {
	return mod.name
}

func (mod *Module) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", mod.Name, mod.constants)
}

func (mod *Module) keywords() []AnyType {
	//keywords reserved for the base language that Ell compiles
	keywords := []AnyType{
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

func (mod *Module) globals() []AnyType {
	syms := make([]AnyType, 0, symtag)
	for _, b := range mod.globalMap {
		if b != nil {
			syms = append(syms, b.sym)
		}
	}
	return syms
}

func (mod *Module) global(sym AnyType) AnyType {
	s := sym.(*SymbolType)
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
	sym AnyType
	val AnyType
}

func (mod *Module) defGlobal(sym AnyType, val AnyType) {
	s := sym.(*SymbolType)
	if s.tag >= len(mod.globalMap) {
		gAnyType := make([]*binding, s.tag+100)
		copy(gAnyType, mod.globalMap)
		mod.globalMap = gAnyType
	}
	b := binding{sym, val}
	mod.globalMap[s.tag] = &b
}

func (mod *Module) isDefined(sym AnyType) bool {
	s := sym.(*SymbolType)
	if s.tag < len(mod.globalMap) {
		return mod.globalMap[s.tag] != nil
	}
	return false
}

func (mod *Module) undefGlobal(sym AnyType) {
	s := sym.(*SymbolType)
	if s.tag < len(mod.globalMap) {
		mod.globalMap[s.tag] = nil
	}
}

func (mod *Module) setGlobal(sym AnyType, val AnyType) error {
	s := sym.(*SymbolType)
	if s.tag < len(mod.globalMap) {
		if mod.globalMap[s.tag] != nil {
			mod.globalMap[s.tag].val = val
			return nil
		}

	}
	return Error("*** Warning: set on undefined global ", sym)
}

func (mod *Module) macros() []AnyType {
	keys := make([]AnyType, 0, len(mod.macroMap))
	for k := range mod.macroMap {
		keys = append(keys, k)
	}
	return keys
}

func (mod *Module) macro(sym AnyType) *macro {
	mac, ok := mod.macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func (mod *Module) defMacro(sym AnyType, val AnyType) {
	mod.macroMap[sym] = newMacro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func (mod *Module) putConstant(val AnyType) int {
	idx, present := mod.constantsMap[val]
	if !present {
		idx = len(mod.constants)
		mod.constants = append(mod.constants, val)
		mod.constantsMap[val] = idx
	}
	return idx
}

func (mod *Module) use(sym AnyType) error {
	name := sym.String()
	return mod.loadModule(name)
}

func (mod *Module) importCode(thunk *Code) (AnyType, error) {
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

func (mod *Module) exports() []AnyType {
	return mod.exported
}

func (mod *Module) findModuleByName(moduleName string) (string, error) {
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
	return "", Error("Module not found: ", moduleName)
}

func (mod *Module) loadModule(name string) error {
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

func (mod *Module) loadFile(file string) error {
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

func (mod *Module) eval(expr AnyType) (AnyType, error) {
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

func (mod *Module) findModuleFile(name string) (string, error) {
	i := strings.Index(name, ".")
	if i < 0 {
		file, err := mod.findModuleByName(name)
		if err != nil {
			return "", err
		}
		return file, nil
	}
	if !fileReadable(name) {
		return "", Error("Cannot read file: ", name)
	}
	return name, nil
}

func (mod *Module) compileExpr(expr AnyType) (string, error) {
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
func (mod *Module) compileFile(name string) (AnyType, error) {
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
			return StringType(result), nil
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
