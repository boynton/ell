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

var constantsMap = make(map[*LOB]int, 0)
var constants = make([]*LOB, 0, 1000)
var macroMap = make(map[*LOB]*macro, 0)
var primitives = make([]*Primitive, 0, 1000)

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

func define(name string, obj *LOB) {
	sym := intern(name)
	if sym == nil {
		panic("Cannot define a value for this symbol: " + name)
	}
	defGlobal(sym, obj)
}

//Need to pass a "signature" string to document usage
// "(x y [z])" or "(x {y: default})" or "(x & y)" or whatever
func defineFunction(name string, fun PrimCallable, signature string) {
	sym := intern(name)
	if global(sym) != nil {
		println("*** Warning: redefining ", name, " with a primitive")
	}

	prim := newPrimitive(name, fun, signature)
	defGlobal(sym, prim)
}

func defineMacro(name string, fun PrimCallable) {
	sym := intern(name)
	if getMacro(sym) != nil {
		println("*** Warning: redefining macro ", name, " -> ", getMacro(sym))
	}
	prim := newPrimitive(name, fun, "(<any>)")
	defMacro(sym, prim)
}

func getKeywords() []*LOB {
	//keywords reserved for the base language that Ell compiles
	keywords := []*LOB{
		intern("quote"),
		intern("fn"),
		intern("if"),
		intern("do"),
		intern("def"),
		intern("defn"),
		intern("defmacro"),
		intern("set!"),
		intern("lap"),
		intern("use"),
	}
	return keywords
}

func getGlobals() []*LOB {
	var syms []*LOB
	for _, sym := range symtab {
		if sym.car != nil {
			syms = append(syms, sym)
		}
	}
	return syms
}

func global(sym *LOB) *LOB {
	if isSymbol(sym) {
		return sym.car
	}
	return nil
}

type binding struct {
	sym *LOB
	val *LOB
}

func defGlobal(sym *LOB, val *LOB) {
	sym.car = val
	delete(macroMap, sym)
}

func isDefined(sym *LOB) bool {
	return sym.car != nil
}

func undefGlobal(sym *LOB) {
	sym.car = nil
}

func macros() []*LOB {
	keys := make([]*LOB, 0, len(macroMap))
	for k := range macroMap {
		keys = append(keys, k)
	}
	return keys
}

func getMacro(sym *LOB) *macro {
	mac, ok := macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func defMacro(sym *LOB, val *LOB) {
	macroMap[sym] = newMacro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func putConstant(val *LOB) int {
	idx, present := constantsMap[val]
	if !present {
		idx = len(constants)
		constants = append(constants, val)
		constantsMap[val] = idx
	}
	return idx
}

func use(sym *LOB) error {
	return loadModule(sym.text)
}

func importCode(thunk *LOB) (*LOB, error) {
	result, err := exec(thunk.code)
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
	} else if interactive {
		println("[loading " + file + "]")
	}
	port, err := openInputFile(file)
	if err != nil {
		return err
	}

	expr, err := readInputPort(port, nil)
	defer closeInputPort(port)

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
		expr, err = readInputPort(port, nil)
	}
}

func eval(expr *LOB) (*LOB, error) {
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
		val := strings.Replace(write(code), "\n", "\n; ", -1)
		println("; compiled to:\n;  ", val)
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

func compileObject(expr *LOB) (string, error) {
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
	thunk, err := compile(expanded)
	if err != nil {
		return "", err
	}
	if verbose {
		println("; compiled to: ", write(thunk))
	}
	return thunk.code.decompile(true) + "\n", nil
}

//caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func compileFile(name string) (*LOB, error) {
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

	expr, err := readInputPort(port, nil)
	defer closeInputPort(port)
	result := ";\n; code generated from " + file + "\n;\n"
	var lap string
	for err == nil {
		if expr == EOF {
			return newString(result), nil
		}
		lap, err = compileObject(expr)
		if err != nil {
			return nil, err
		}
		result += lap
		expr, err = readInputPort(port, nil)
	}
	return nil, err
}
