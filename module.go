/*
Copyright 2015 Lee Boynton

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

package ell

import (
	"os"
	"path/filepath"
	"strings"
)

var verbose bool
var debug bool
var interactive bool

// SetFlags - set various flags controlling the runtime
func SetFlags(o bool, v bool, d bool, t bool, i bool) {
	optimize = o
	verbose = v
	debug = d
	trace = t
	interactive = i
}

// Version - this version of ell
const Version = "ell v0.2"

// LoadPath is the path where the library *.ell files can be found
var LoadPath string

var constantsMap = make(map[*Object]int, 0)
var constants = make([]*Object, 0, 1000)
var macroMap = make(map[*Object]*macro, 0)
var primitives = make([]*primitive, 0, 1000)

// Bind the value to the global name
func DefineGlobal(name string, obj *Object) {
	sym := Intern(name)
	if sym == nil {
		panic("Cannot define a value for this symbol: " + name)
	}
	defGlobal(sym, obj)
}

func definePrimitive(name string, prim *Object) {
	sym := Intern(name)
	if GetGlobal(sym) != nil {
		println("*** Warning: redefining ", name, " with a primitive")
	}
	defGlobal(sym, prim)
}

// Register a primitive function to the specified global name
func DefineFunction(name string, fun PrimitiveFunction, result *Object, args ...*Object) {
	prim := Primitive(name, fun, result, args, nil, nil, nil)
	definePrimitive(name, prim)
}

// Register a primitive function with Rest arguments to the specified global name
func DefineFunctionRestArgs(name string, fun PrimitiveFunction, result *Object, rest *Object, args ...*Object) {
	prim := Primitive(name, fun, result, args, rest, []*Object{}, nil)
	definePrimitive(name, prim)
}

// Register a primitive function with optional arguments to the specified global name
func DefineFunctionOptionalArgs(name string, fun PrimitiveFunction, result *Object, args []*Object, defaults ...*Object) {
	prim := Primitive(name, fun, result, args, nil, defaults, nil)
	definePrimitive(name, prim)
}

// Register a primitive function with keyword arguments to the specified global name
func DefineFunctionKeyArgs(name string, fun PrimitiveFunction, result *Object, args []*Object, defaults []*Object, keys []*Object) {
	prim := Primitive(name, fun, result, args, nil, defaults, keys)
	definePrimitive(name, prim)
}

// Register a primitive macro with the specified name.
func DefineMacro(name string, fun PrimitiveFunction) {
	sym := Intern(name)
	if GetMacro(sym) != nil {
		println("*** Warning: redefining macro ", name, " -> ", GetMacro(sym))
	}
	prim := Primitive(name, fun, AnyType, []*Object{AnyType}, nil, nil, nil)
	defMacro(sym, prim)
}

// GetKeywords - return a slice of Ell primitive reserved words
func GetKeywords() []*Object {
	//keywords reserved for the base language that Ell compiles
	keywords := []*Object{
		Intern("quote"),
		Intern("fn"),
		Intern("if"),
		Intern("do"),
		Intern("def"),
		Intern("defn"),
		Intern("defmacro"),
		Intern("set!"),
		Intern("code"),
		Intern("use"),
	}
	return keywords
}

// Globals - return a slice of all defined global symbols
func Globals() []*Object {
	var syms []*Object
	for _, sym := range symtab {
		if sym.car != nil {
			syms = append(syms, sym)
		}
	}
	return syms
}

// GetGlobal - return the global value for the specified symbol, or nil if the symbol is not defined.
func GetGlobal(sym *Object) *Object {
	if IsSymbol(sym) {
		return sym.car
	}
	return nil
}

type binding struct {
	sym *Object
	val *Object
}

func defGlobal(sym *Object, val *Object) {
	sym.car = val
	delete(macroMap, sym)
}

// IsDefined - return true if the there is a global value defined for the symbol
func IsDefined(sym *Object) bool {
	return sym.car != nil
}

func undefGlobal(sym *Object) {
	sym.car = nil
}

// Macros - return a slice of all defined macros
func Macros() []*Object {
	keys := make([]*Object, 0, len(macroMap))
	for k := range macroMap {
		keys = append(keys, k)
	}
	return keys
}

// GetMacro - return the macro for the symbol, or nil if not defined
func GetMacro(sym *Object) *macro {
	mac, ok := macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func defMacro(sym *Object, val *Object) {
	macroMap[sym] = Macro(sym, val)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func putConstant(val *Object) int {
	idx, present := constantsMap[val]
	if !present {
		idx = len(constants)
		constants = append(constants, val)
		constantsMap[val] = idx
	}
	return idx
}

func Use(sym *Object) error {
	return Load(sym.text)
}

func importCode(thunk *Object) (*Object, error) {
	var args []*Object
	result, err := exec(thunk.code, args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func FindModuleByName(moduleName string) (string, error) {
	path := strings.Split(LoadPath, ":")
	name := moduleName
	var lname string
	if strings.HasSuffix(name, ".ell") {
		lname = name[:len(name)-3] + ".lvm"
	} else {
		lname = name + ".lvm"
		name = name + ".ell"
	}
	for _, dirname := range path {
		filename := filepath.Join(dirname, lname)
		if IsFileReadable(filename) {
			return filename, nil
		}
		filename = filepath.Join(dirname, name)
		if IsFileReadable(filename) {
			return filename, nil
		}
	}
	return "", Error(IOErrorKey, "Module not found: ", moduleName)
}

func Load(name string) error {
	file, err := FindModuleFile(name)
	if err != nil {
		return err
	}
	return LoadFile(file)
}

func LoadFile(file string) error {
	if verbose {
		println("; loadFile: " + file)
	} else if interactive {
		println("[loading " + file + "]")
	}
	fileText, err := SlurpFile(file)
	if err != nil {
		return err
	}
	exprs, err := ReadAll(fileText, nil)
	if err != nil {
		return err
	}
	for exprs != EmptyList {
		expr := Car(exprs)
		_, err = Eval(expr)
		if err != nil {
			return err
		}
		exprs = Cdr(exprs)
	}
	return nil
}

func Eval(expr *Object) (*Object, error) {
	if debug {
		println("; eval: ", Write(expr))
	}
	expanded, err := macroexpandObject(expr)
	if err != nil {
		return nil, err
	}
	if debug {
		println("; expanded to: ", Write(expanded))
	}
	code, err := Compile(expanded)
	if err != nil {
		return nil, err
	}
	if debug {
		val := strings.Replace(Write(code), "\n", "\n; ", -1)
		println("; compiled to:\n;  ", val)
	}
	return importCode(code)
}

func FindModuleFile(name string) (string, error) {
	i := strings.Index(name, ".")
	if i < 0 {
		file, err := FindModuleByName(name)
		if err != nil {
			return "", err
		}
		return file, nil
	}
	if !IsFileReadable(name) {
		return "", Error(IOErrorKey, "Cannot read file: ", name)
	}
	return name, nil
}

func compileObject(expr *Object) (string, error) {
	if debug {
		println("; compile: ", Write(expr))
	}
	expanded, err := macroexpandObject(expr)
	if err != nil {
		return "", err
	}
	if debug {
		println("; expanded to: ", Write(expanded))
	}
	thunk, err := Compile(expanded)
	if err != nil {
		return "", err
	}
	if debug {
		println("; compiled to: ", Write(thunk))
	}
	return thunk.code.decompile(true) + "\n", nil
}

//caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func CompileFile(name string) (*Object, error) {
	file, err := FindModuleFile(name)
	if err != nil {
		return nil, err
	}
	if verbose {
		println("; loadFile: " + file)
	}
	fileText, err := SlurpFile(file)
	if err != nil {
		return nil, err
	}

	exprs, err := ReadAll(fileText, nil)
	result := ";\n; code generated from " + file + "\n;\n"
	var lvm string
	for exprs != EmptyList {
		expr := Car(exprs)
		lvm, err = compileObject(expr)
		if err != nil {
			return nil, err
		}
		result += lvm
		exprs = Cdr(exprs)
	}
	return String(result), nil
}

type Extension interface {
	Init() error
	Cleanup()
}

var extension Extension

func Init(ext Extension) {
	extension = ext
	LoadPath = os.Getenv("ELL_PATH")
	home := os.Getenv("HOME")
	if LoadPath == "" {
		LoadPath = "."
		homelib := filepath.Join(home, "lib/ell")
		_, err := os.Stat(homelib)
		if err == nil {
			LoadPath += ":" + homelib
		}
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			golibdir := filepath.Join(gopath, "src/github.com/boynton/ell/lib")
			_, err := os.Stat(golibdir)
			if err == nil {
				LoadPath += ":" + golibdir
			}
		}
	}
	if verbose || debug {
		Println("[LoadPath ", LoadPath, "]")
	}
	InitPrimitives()
	if extension != nil {
		err := extension.Init()
		if err != nil {
			Fatal("*** ", err)
		}
	}
}

func Cleanup() {
	if extension != nil {
		extension.Cleanup()
	}
}

func Run(args []string) {
	for _, filename := range args {
		err := Load(filename)
		if err != nil {
			Fatal("*** ", err.Error())
		}
	}
}
