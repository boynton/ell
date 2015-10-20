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

package ell

import (
	"flag"
	//"github.com/davecheney/profile"
	"os"
	"path/filepath"
	"strings"
)

var verbose bool
var debug bool
var interactive bool

func SetFlags(o bool, v bool, d bool, t bool, i bool) {
	optimize = o
	verbose = v
	debug = d
	trace = t
	interactive = i
}

// Version - this version of ell
const Version = "ell v0.2"

// EllPath is the path where the library *.ell files can be found
var EllPath string

var constantsMap = make(map[*LOB]int, 0)
var constants = make([]*LOB, 0, 1000)
var macroMap = make(map[*LOB]*macro, 0)
var primitives = make([]*primitive, 0, 1000)

func DefineGlobal(name string, obj *LOB) {
	sym := Intern(name)
	if sym == nil {
		panic("Cannot define a value for this symbol: " + name)
	}
	defGlobal(sym, obj)
}

func DefinePrimitive(name string, prim *LOB) {
	sym := Intern(name)
	if GetGlobal(sym) != nil {
		println("*** Warning: redefining ", name, " with a primitive")
	}
	defGlobal(sym, prim)
}

func DefineFunction(name string, fun PrimitiveFunction, result *LOB, args ...*LOB) {
	prim := Primitive(name, fun, result, args, nil, nil, nil)
	DefinePrimitive(name, prim)
}

func DefineFunctionRestArgs(name string, fun PrimitiveFunction, result *LOB, rest *LOB, args ...*LOB) {
	prim := Primitive(name, fun, result, args, rest, []*LOB{}, nil)
	DefinePrimitive(name, prim)
}

func DefineFunctionOptionalArgs(name string, fun PrimitiveFunction, result *LOB, args []*LOB, defaults ...*LOB) {
	prim := Primitive(name, fun, result, args, nil, defaults, nil)
	DefinePrimitive(name, prim)
}

func DefineFunctionKeyArgs(name string, fun PrimitiveFunction, result *LOB, args []*LOB, defaults []*LOB, keys []*LOB) {
	prim := Primitive(name, fun, result, args, nil, defaults, keys)
	DefinePrimitive(name, prim)
}

func DefineMacro(name string, fun PrimitiveFunction) {
	sym := Intern(name)
	if GetMacro(sym) != nil {
		println("*** Warning: redefining macro ", name, " -> ", GetMacro(sym))
	}
	prim := Primitive(name, fun, AnyType, []*LOB{AnyType}, nil, nil, nil)
	defMacro(sym, prim)
}

func GetKeywords() []*LOB {
	//keywords reserved for the base language that Ell compiles
	keywords := []*LOB{
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

func Globals() []*LOB {
	var syms []*LOB
	for _, sym := range symtab {
		if sym.car != nil {
			syms = append(syms, sym)
		}
	}
	return syms
}

func GetGlobal(sym *LOB) *LOB {
	if IsSymbol(sym) {
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

func IsDefined(sym *LOB) bool {
	return sym.car != nil
}

func undefGlobal(sym *LOB) {
	sym.car = nil
}

func Macros() []*LOB {
	keys := make([]*LOB, 0, len(macroMap))
	for k := range macroMap {
		keys = append(keys, k)
	}
	return keys
}

func GetMacro(sym *LOB) *macro {
	mac, ok := macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func defMacro(sym *LOB, val *LOB) {
	macroMap[sym] = Macro(sym, val)
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

func Use(sym *LOB) error {
	return Load(sym.text)
}

func importCode(thunk *LOB) (*LOB, error) {
	var args []*LOB
	result, err := exec(thunk.code, args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func FindModuleByName(moduleName string) (string, error) {
	path := strings.Split(EllPath, ":")
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

func Eval(expr *LOB) (*LOB, error) {
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

func compileObject(expr *LOB) (string, error) {
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
func CompileFile(name string) (*LOB, error) {
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

func Main(ext Extension) {
	pCompile := flag.Bool("c", false, "compile the file and output lap")
	pOptimize := flag.Bool("o", false, "optimize execution speed, should work for correct code, but doesn't check everything")
	pVerbose := flag.Bool("v", false, "verbose mode, print extra information")
	pDebug := flag.Bool("d", false, "debug mode, print extra information about compilation")
	pTrace := flag.Bool("t", false, "trace VM instructions as they get executed")
	pNoInit := flag.Bool("i", false, "disable initialization from the $HOME/.ell file")
	flag.Parse()
	args := flag.Args()

	EllPath = os.Getenv("ELL_PATH")
	home := os.Getenv("HOME")
	ellini := filepath.Join(home, ".ell")
	if EllPath == "" {
		EllPath = "."
		homelib := filepath.Join(home, "lib/ell")
		_, err := os.Stat(homelib)
		if err == nil {
			EllPath += ":" + homelib
		}
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			golibdir := filepath.Join(gopath, "src/github.com/boynton/gell/lib")
			_, err := os.Stat(golibdir)
			if err == nil {
				EllPath += ":" + golibdir
			}
		}
	}
	Init(ext)
	if len(args) < 1 {
		if !*pNoInit {
			_, err := os.Stat(ellini)
			if err == nil {
				err := Load(ellini)
				if err != nil {
					Fatal("*** ", err)
				}
			}
		}
		SetFlags(*pOptimize, *pVerbose, *pDebug, *pTrace, true)
		ReadEvalPrintLoop()
	} else {
		SetFlags(*pOptimize, *pVerbose, *pDebug, *pTrace, false)
		/*
			if len(os.Args) > 2 {
				cfg := profile.Config{
					CPUProfile:     true,
					ProfilePath:    ".",  // store profiles in current directory
					NoShutdownHook: true, // do not hook SIGINT
				}
				defer profile.Start(&cfg).Stop()
			}
		*/
		for _, filename := range args {
			if *pCompile {
				//just compile and print LAP code
				lap, err := CompileFile(filename)
				if err != nil {
					Fatal("*** ", err)
				}
				println(lap)
			} else {
				//this executes the file
				err := Load(filename)
				if err != nil {
					Fatal("*** ", err.Error())
				}
			}
		}
	}
	Cleanup()
}
