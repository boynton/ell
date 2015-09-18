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
	"path/filepath"
	"strings"
)

var constantsMap = make(map[LAny]int, 0)
var constants = make([]LAny, 0, 1000)
var globals = make([]*binding, 0, 1000)
var macroMap = make(map[LAny]*macro, 0)

func checkInterrupt() bool {
	if interrupts != nil {
		select {
		case msg := <-interrupts:
			return msg != nil
		default:
			return false
		}
	}
	return false
}

func define(name string, obj LAny) {
	sym := intern(name)
	if sym == nil {
		panic("Cannot define a value for this symbol: " + name)
	}
	defGlobal(sym, obj)
}

func defineFunction(name string, fun primitive) {
	sym := intern(name)
	if global(sym) != nil {
		println("*** Warning: redefining ", name)
	}
	prim := LPrimitive{name, fun}
	defGlobal(sym, &prim)
}

func defineMacro(name string, fun primitive) {
	sym := intern(name)
	if getMacro(sym) != nil {
		println("*** Warning: redefining macro ", name, " -> ", getMacro(sym))
	}
	prim := LPrimitive{name, fun}
	defMacro(sym, &prim)
}

func getKeywords() []LAny {
	//keywords reserved for the base language that Ell compiles
	keywords := []LAny{
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

func getGlobals() []LAny {
	syms := make([]LAny, 0, symtag)
	for _, b := range globals {
		if b != nil {
			syms = append(syms, b.sym)
		}
	}
	return syms
}

func global(sym LAny) LAny {
	s := sym.(*SymbolType)
	if s.tag >= len(globals) {
		return nil
	}
	tmp := globals[s.tag]
	if tmp == nil {
		return nil
	}
	return tmp.val
}

type binding struct {
	sym LAny
	val LAny
}

func defGlobal(sym LAny, val LAny) {
	s := sym.(*SymbolType)
	if s.tag >= len(globals) {
		gLAny := make([]*binding, s.tag+100)
		copy(gLAny, globals)
		globals = gLAny
	}
	b := binding{sym, val}
	globals[s.tag] = &b
}

func isDefined(sym LAny) bool {
	s := sym.(*SymbolType)
	if s.tag < len(globals) {
		return globals[s.tag] != nil
	}
	return false
}

func undefGlobal(sym LAny) {
	s := sym.(*SymbolType)
	if s.tag < len(globals) {
		globals[s.tag] = nil
	}
}

func setGlobal(sym LAny, val LAny) error {
	s := sym.(*SymbolType)
	if s.tag < len(globals) {
		if globals[s.tag] != nil {
			globals[s.tag].val = val
			return nil
		}

	}
	return Error("*** Warning: set on undefined global ", sym)
}

func macros() []LAny {
	keys := make([]LAny, 0, len(macroMap))
	for k := range macroMap {
		keys = append(keys, k)
	}
	return keys
}

func getMacro(sym LAny) *macro {
	mac, ok := macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func defMacro(sym LAny, val LAny) {
	macroMap[sym] = newMacro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func putConstant(val LAny) int {
	idx, present := constantsMap[val]
	if !present {
		idx = len(constants)
		constants = append(constants, val)
		constantsMap[val] = idx
	}
	return idx
}

func use(sym LAny) error {
	name := sym.String()
	return loadModule(name)
}

func importCode(thunk *Code) (LAny, error) {
	result, err := exec(thunk)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func findModuleByName(moduleName string) (string, error) {
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

func loadModule(name string) error {
	file := name
	if !fileReadable(name) {
		f, err := findModuleFile(name)
		if err != nil {
			return err
		}
		file = f
	}
	return loadFile(file)
}

func loadFile(file string) error {
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
		_, err = eval(expr)
		if err != nil {
			return err
		}
		expr, err = port.read()
	}
}

func eval(expr LAny) (LAny, error) {
	if verbose {
		println("; eval: ", write(expr))
	}
	expanded, err := macroexpandObject(expr)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; expanded to: ", write(expanded))
	}
	code, err := compile(expanded)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; compiled to: ", write(code))
	}
	result, err := importCode(code)
	return result, err
}

func findModuleFile(name string) (string, error) {
	i := strings.Index(name, ".")
	if i < 0 {
		file, err := findModuleByName(name)
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

func compileObject(expr LAny) (string, error) {
	if verbose {
		println("; compile: ", write(expr))
	}
	expanded, err := macroexpandObject(expr)
	if err != nil {
		return "", err
	}
	if verbose {
		println("; expanded to: ", write(expanded))
	}
	code, err := compile(expanded)
	if err != nil {
		return "", err
	}
	if verbose {
		println("; compiled to: ", write(code))
	}
	return code.decompile(true) + "\n", nil
}

//caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func compileFile(name string) (LAny, error) {
	file, err := findModuleFile(name)
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
		lap, err = compileObject(expr)
		if err != nil {
			return nil, err
		}
		result += lap
		expr, err = port.read()
	}
	return nil, err
}
