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
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/boynton/cli"
	. "github.com/boynton/ell/data"
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
var Version = "(development version)"

var constantsMap = make(map[Value]int, 0)
var constants = make([]Value, 0, 1000)
var macroMap = make(map[Value]*macro, 0)
var primitives = make([]*Primitive, 0, 1000)

// Bind the value to the global name
func DefineGlobal(name string, obj Value) {
	sym := Intern(name)
	if p, ok := sym.(*Symbol); ok {
		defGlobal(p, obj)
	} else {
		panic("Cannot define a value for this symbol: " + name)
	}
}

func definePrimitive(name string, prim *Function) {
	sym := Intern(name)
	if GetGlobal(sym) != nil {
		println("*** Warning: redefining ", name, " with a primitive")
	}
	if p, ok := sym.(*Symbol); ok {
		defGlobal(p, prim)
	} else {
		panic("Cannot define a value for this symbol: " + name)
	}
}

// Register a primitive function to the specified global name
func DefineFunction(name string, fun PrimitiveFunction, result Value, args ...Value) {
	prim := NewPrimitive(name, fun, result, args, nil, nil, nil)
	definePrimitive(name, prim)
}

// Register a primitive function with Rest arguments to the specified global name
func DefineFunctionRestArgs(name string, fun PrimitiveFunction, result Value, rest Value, args ...Value) {
	prim := NewPrimitive(name, fun, result, args, rest, []Value{}, nil)
	definePrimitive(name, prim)
}

// Register a primitive function with optional arguments to the specified global name
func DefineFunctionOptionalArgs(name string, fun PrimitiveFunction, result Value, args []Value, defaults ...Value) {
	prim := NewPrimitive(name, fun, result, args, nil, defaults, nil)
	definePrimitive(name, prim)
}

// Register a primitive function with keyword arguments to the specified global name
func DefineFunctionKeyArgs(name string, fun PrimitiveFunction, result Value, args []Value, defaults []Value, keys []Value) {
	prim := NewPrimitive(name, fun, result, args, nil, defaults, keys)
	definePrimitive(name, prim)
}

// Register a primitive macro with the specified name.
func DefineMacro(name string, fun PrimitiveFunction) {
	sym := Intern(name)
	if GetMacro(sym) != nil {
		println("*** Warning: redefining macro ", name, " -> ", GetMacro(sym))
	}
	prim := NewPrimitive(name, fun, AnyType, []Value{AnyType}, nil, nil, nil)
	defMacro(sym, prim)
}

// GetKeywords - return a slice of Ell primitive reserved words
func GetKeywords() []Value {
	//keywords reserved for the base language that Ell compiles
	keywords := []Value{
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
func Globals() []*Symbol {
	var syms []*Symbol
	for _, sym := range Symbols() {
		if p, ok := sym.(*Symbol); ok && p.Value != nil {
			syms = append(syms, p)
		}
	}
	return syms
}

// GetGlobal - return the global value for the specified symbol, or nil if the symbol is not defined.
func GetGlobal(sym Value) Value {
	if p, ok := sym.(*Symbol); ok {
		return p.Value
	}
	return nil
}

func defGlobal(sym *Symbol, val Value) {
	sym.Value = val
	delete(macroMap, sym)
}

// IsDefined - return true if the there is a global value defined for the symbol
func IsDefined(sym *Symbol) bool {
	return sym.Value != nil
}

func undefGlobal(sym *Symbol) {
	sym.Value = nil
}

// Macros - return a slice of all defined macros
func Macros() []Value {
	keys := make([]Value, 0, len(macroMap))
	for k := range macroMap {
		keys = append(keys, k)
	}
	return keys
}

// GetMacro - return the macro for the symbol, or nil if not defined
func GetMacro(sym Value) *macro {
	mac, ok := macroMap[sym]
	if !ok {
		return nil
	}
	return mac
}

func defMacro(sym Value, val *Function) {
	macroMap[sym] = NewMacro(sym, val)
}

// note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
// so, we will end up with duplicates, unless we do some deep compare, when putting map or array constants
func putConstant(val Value) int {
	idx, present := constantsMap[val]
	if !present {
		idx = len(constants)
		constants = append(constants, val)
		constantsMap[val] = idx
	}
	return idx
}

func Use(sym *Symbol) error {
	return Load(sym.Text)
}

func importCode(thunk *Code) (Value, error) {
	var args []Value
	result, err := exec(thunk, args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

var loadPathSymbol = Intern("*load-path*")

func FindModuleByName(moduleName string) (string, error) {
	if moduleName == "ell" || moduleName == "ell.ell" {
		return "@/ell.ell", nil
	}
	loadPath := GetGlobal(loadPathSymbol)
	if loadPath == nil {
		loadPath = NewString(".")
	}
	path := strings.Split(StringValue(loadPath), ":")
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
	return "", NewError(IOErrorKey, "Module not found: ", moduleName)
}

func Load(name string) error {
	if verbose {
		fmt.Println("; [loading " + name + "]")
	}
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
	exprs, err := ReadAllFromString(fileText)
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

func Eval(expr Value) (Value, error) {
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
		return "", NewError(IOErrorKey, "Cannot read file: ", name)
	}
	return name, nil
}

func compileValue(expr Value) (string, error) {
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
	return thunk.decompile(true) + "\n", nil
}

// caveats: when you compile a file, you actually run it. This is so we can handle imports and macros correctly.
func CompileFile(name string) (Value, error) {
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

	exprs, err := ReadAllFromString(fileText)
	result := ";\n; code generated from " + file + "\n;\n"
	var lvm string
	for exprs != EmptyList {
		expr := Car(exprs)
		lvm, err = compileValue(expr)
		if err != nil {
			return nil, err
		}
		result += lvm
		exprs = Cdr(exprs)
	}
	return NewString(result), nil
}

type Extension interface {
	Init() error
	Cleanup()
	String() string
}

var extensions []Extension

func AddEllDirectory(dirname string) {
	loadPath := dirname
	tmp := GetGlobal(loadPathSymbol)
	if tmp != nil {
		loadPath = dirname + ":" + StringValue(tmp)
	}
	DefineGlobal(StringValue(loadPathSymbol), NewString(loadPath))
}

func Init(extns ...Extension) {
	extensions = extns
	loadPath := os.Getenv("ELL_PATH")
	home := os.Getenv("HOME")
	if loadPath == "" {
		loadPath = "."
		homelib := filepath.Join(home, "lib/ell")
		_, err := os.Stat(homelib)
		if err == nil {
			loadPath += ":" + homelib
		}
	}
	loadPath += ":@/"
	DefineGlobal(StringValue(loadPathSymbol), NewString(loadPath))
	InitPrimitives()
	for _, ext := range extensions {
		err := ext.Init()
		if err != nil {
			Fatal("*** ", err)
		}
	}
}

func Cleanup() {
	for _, ext := range extensions {
		ext.Cleanup()
	}
}

func Run(args ...string) {
	for _, filename := range args {
		err := Load(filename)
		if err != nil {
			Fatal("*** ", err.Error())
		}
	}
}

func Main(extns ...Extension) {
	var help, compile, optimize, verbose, debug, trace, noInit bool
	var path string
	cmd := cli.New("ell", "The Ell Language compiler, VM, and runtime")
	cmd.BoolOption(&help, "help", false, "Show help")
	cmd.BoolOption(&compile, "compile", false, "compile the file and output lap")
	cmd.BoolOption(&optimize, "optimize", false, "optimize execution speed, should work for correct code, relax some checks")
	cmd.BoolOption(&verbose, "verbose", false, "verbose mode, print extra information")
	cmd.BoolOption(&debug, "debug", false, "debug mode, print extra information about compilation")
	cmd.BoolOption(&trace, "trace", false, "trace VM instructions as they get executed")
	cmd.BoolOption(&noInit, "noinit", false, "disable initialization from the $HOME/.ell file")
	var prof string
	cmd.StringOption(&prof, "profile", "", "profile the code to the specified file")
	cmd.StringOption(&path, "path", "", "add directories to ell load path")
	args, _ := cmd.Parse()
	if help {
		fmt.Println(cmd.Usage())
		os.Exit(1)
	}
	interactive := len(args) == 0
	SetFlags(optimize, verbose, debug, trace, interactive)
	Init(extns...)
	if path != "" {
		for _, p := range strings.Split(path, ":") {
			expandedPath := ExpandFilePath(p)
			if IsDirectoryReadable(expandedPath) {
				AddEllDirectory(expandedPath)
				if debug {
					Println("[added directory to path: '", expandedPath, "']")
				}
			} else if debug {
				Println("[directory not readable, cannot add to path: '", expandedPath, "']")
			}
		}
	}
	if len(args) > 0 {
		if compile {
			SetFlags(optimize, verbose, debug, trace, interactive)
			//just compile and print LVM code
			for _, filename := range args {
				lap, err := CompileFile(filename)
				if err != nil {
					Fatal("*** ", err)
				}
				Println(lap)
			}
		} else {
			if prof != "" {
				f, err := os.Create(prof)
				if err != nil {
					Fatal("*** ", err)
				}
				pprof.StartCPUProfile(f)
				defer pprof.StopCPUProfile()
			}
			SetFlags(optimize, verbose, debug, trace, interactive)
			Run(args...)
		}
	} else {
		if !noInit {
			home := os.Getenv("HOME")
			ellini := filepath.Join(home, ".ell")
			_, err := os.Stat(ellini)
			if err == nil {
				err := Load(ellini)
				if err != nil {
					Fatal("*** ", err)
				}
			}
		}
		SetFlags(optimize, verbose, debug, trace, interactive)
		ReadEvalPrintLoop()
	}
	Cleanup()
}
