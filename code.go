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
	"bytes"
	"fmt"
	"strconv"
	"strings"

	. "github.com/boynton/ell/data"
)

const (
	opcodeLiteral = iota
	opcodeLocal
	opcodeJumpFalse
	opcodeJump
	opcodeTailCall
	opcodeCall
	opcodeReturn
	opcodeClosure
	opcodePop
	opcodeGlobal
	opcodeDefGlobal
	opcodeSetLocal
	opcodeUse
	opcodeDefMacro
	opcodeVector
	opcodeStruct
	opcodeUndefGlobal
	opcodeCount
)

var LiteralSymbol = Intern("literal")
var LocalSymbol = Intern("local")
var JumpfalseSymbol = Intern("jumpfalse")
var JumpSymbol = Intern("jump")
var TailcallSymbol = Intern("tailcall")
var CallSymbol = Intern("call")
var ReturnSymbol = Intern("return")
var ClosureSymbol = Intern("closure")
var PopSymbol = Intern("pop")
var GlobalSymbol = Intern("global")
var DefglobalSymbol = Intern("defglobal")
var SetlocalSymbol = Intern("setlocal")
var UseSymbol = Intern("use")
var DefmacroSymbol = Intern("defmacro")
var VectorSymbol = Intern("vector")
var StructSymbol = Intern("struct")
var UndefineSymbol = Intern("undefine")
var FuncSymbol = Intern("func")

var opsyms = initOpsyms()

func initOpsyms() []Value {
	syms := make([]Value, opcodeCount)
	syms[opcodeLiteral] = LiteralSymbol
	syms[opcodeLocal] = LocalSymbol
	syms[opcodeJumpFalse] = JumpfalseSymbol
	syms[opcodeJump] = JumpSymbol
	syms[opcodeTailCall] = TailcallSymbol
	syms[opcodeCall] = CallSymbol
	syms[opcodeReturn] = ReturnSymbol
	syms[opcodeClosure] = ClosureSymbol
	syms[opcodePop] = PopSymbol
	syms[opcodeGlobal] = GlobalSymbol
	syms[opcodeDefGlobal] = DefglobalSymbol
	syms[opcodeSetLocal] = SetlocalSymbol
	syms[opcodeUse] = UseSymbol
	syms[opcodeDefMacro] = DefmacroSymbol
	syms[opcodeVector] = VectorSymbol
	syms[opcodeStruct] = StructSymbol
	syms[opcodeUndefGlobal] = UndefineSymbol
	return syms
}

var CodeType Value = Intern("<code>")

// Code - compiled Ell bytecode
type Code struct {
	name     string
	ops      []int
	argc     int
	defaults []Value
	keys     []Value
}

func MakeCode(argc int, defaults []Value, keys []Value, name string) *Code {
	var ops []int
	code := &Code{
		name,
		ops,
		argc,
		defaults, //nil for normal procs, empty for rest, and non-empty for optional/keyword
		keys,
	}
	return code
}

func (code *Code) Type() Value {
	return CodeType
}

func (code1 *Code) Equals(another Value) bool {
	if code2, ok := another.(*Code); ok {
		return code1 == code2
	}
	return false
}


func (code *Code) signature() string {
	//
	//experimental: external annotations on the functions: *declarations* is a map from symbol to string
	//
	//a macro to declare could be as follows:
	// (defmacro declare (name args)
	//   (or (symbol? name) (error "expected a <symbol> got " name))
	//   (let ((sig (string args)))
	//     `(put! *declarations* '~name '~sig)))
	//used as:
	// (declare cons (<any> <list>) <list>)
	if code.name != "" {
		val := GetGlobal(Intern("*declarations*")) //so if this this has not been defined, we'll just skip it
		if val != nil && IsStruct(val) {
			sig, _ := Get(val, Intern(code.name))
			if sig != Null {
				return sig.String()
			}
		}
	}
	//the following has no type info
	tmp := ""
	for i := 0; i < code.argc; i++ {
		tmp += " <any>"
	}
	if code.defaults != nil {
		tmp += " <any>*"
	}
	if tmp != "" {
		tmp = "(" + tmp[1:] + ")"
	} else {
		tmp = "()"
	}
	return tmp
}

func (code *Code) decompile(pretty bool) string {
	var buf bytes.Buffer
	code.decompileInto(&buf, "", pretty)
	s := buf.String()
	return strings.Replace(s, "("+SymbolName(FuncSymbol)+" (\"\" 0 [] [])", "(code", 1)
}

func (code *Code) decompileInto(buf *bytes.Buffer, indent string, pretty bool) {
	indentAmount := "   "
	offset := 0
	max := len(code.ops)
	prefix := " "
	buf.WriteString(indent + "(" + SymbolName(FuncSymbol) + " (")
	buf.WriteString(fmt.Sprintf("%q ", code.name))
	buf.WriteString(strconv.Itoa(code.argc))
	if code.defaults != nil {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%v", code.defaults))
	} else {
		buf.WriteString(" []")
	}
	if code.keys != nil {
		buf.WriteString(" ")
		buf.WriteString(fmt.Sprintf("%v", code.keys))
	} else {
		buf.WriteString(" []")
	}
	buf.WriteString(")")
	if pretty {
		indent = indent + indentAmount
		prefix = "\n" + indent
	}
	for offset < max {
		op := code.ops[offset]
		s := prefix + "(" + SymbolName(opsyms[op])
		switch op {
		case opcodePop, opcodeReturn:
			buf.WriteString(s + ")")
			offset++
		case opcodeLiteral, opcodeDefGlobal, opcodeUse, opcodeGlobal, opcodeUndefGlobal, opcodeDefMacro:
			buf.WriteString(s + " " + Write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeCall, opcodeTailCall, opcodeJumpFalse, opcodeJump, opcodeVector, opcodeStruct:
			buf.WriteString(s + " " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeLocal, opcodeSetLocal:
			buf.WriteString(s + " " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opcodeClosure:
			buf.WriteString(s)
			if pretty {
				buf.WriteString("\n")
			} else {
				buf.WriteString(" ")
			}
			indent2 := ""
			if pretty {
				indent2 = indent + indentAmount
			}
			(constants[code.ops[offset+1]].(*Code)).decompileInto(buf, indent2, pretty)
			buf.WriteString(")")
			offset += 2
		default:
			panic(fmt.Sprintf("Bad instruction: %d", code.ops[offset]))
		}
	}
	buf.WriteString(")")
}

func (code *Code) String() string {
	return code.decompile(true)
	//	return fmt.Sprintf("(function (%d %v %s) %v)", code.argc, code.defaults, code.keys, code.ops)
}

func (code *Code) loadOps(lst *List) error {
	for lst != EmptyList {
		instr := Car(lst)
		op := Car(instr)
		switch op {
		case ClosureSymbol:
			lstFunc := Cadr(instr)
			if Car(lstFunc) != FuncSymbol {
				return NewError(SyntaxErrorKey, instr)
			}
			lstFunc = Cdr(lstFunc)
			funcParams := Car(lstFunc)
			var argc int
			var name string
			var defaults []Value
			var keys []Value
			var err error
			/*			if IsSymbol(funcParams) {
				//legacy form, just the argc
				argc, err = AsIntValue(funcParams)
				if err != nil {
					return err
				}
				if argc < 0 {
					argc = -argc - 1
					defaults = make([]*Object, 0)
				}
			} else */
			if lst, ok := funcParams.(*List); ok && lst.Length() == 4 {
				//			if IsList(funcParams) && ListLength(funcParams) == 4 {
				a := lst.Car
				lst = lst.Cdr
				name, err = AsStringValue(a)
				if err != nil {
					return NewError(SyntaxErrorKey, funcParams)
				}
				a = lst.Car
				lst = lst.Cdr
				if n, ok := a.(*Number); ok {
					argc = n.IntValue()
				} else {
					return NewError(SyntaxErrorKey, funcParams)
				}
				a = lst.Car
				lst = lst.Cdr
				if v, ok := a.(*Vector); ok {
					defaults = v.Elements
				}
				a = lst.Car
				if v, ok := a.(*Vector); ok {
					keys = v.Elements
				}
			} else {
				return NewError(SyntaxErrorKey, funcParams)
			}
			fun := MakeCode(argc, defaults, keys, name)
			fun.loadOps(Cdr(lstFunc))
			code.emitClosure(fun)
		case LiteralSymbol:
			code.emitLiteral(Cadr(instr))
		case LocalSymbol:
			i, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			j, err := AsIntValue(Caddr(instr))
			if err != nil {
				return err
			}
			code.emitLocal(i, j)
		case SetlocalSymbol:
			i, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			j, err := AsIntValue(Caddr(instr))
			if err != nil {
				return err
			}
			code.emitSetLocal(i, j)
		case GlobalSymbol:
			sym := Cadr(instr)
			if IsSymbol(sym) {
				code.emitGlobal(sym)
			} else {
				return NewError(GlobalSymbol, " argument 1 not a symbol: ", sym)
			}
		case UndefineSymbol:
			code.emitUndefGlobal(Cadr(instr))
		case JumpSymbol:
			loc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitJump(loc)
		case JumpfalseSymbol:
			loc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitJumpFalse(loc)
		case CallSymbol:
			argc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitCall(argc)
		case TailcallSymbol:
			argc, err := AsIntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.emitTailCall(argc)
		case ReturnSymbol:
			code.emitReturn()
		case PopSymbol:
			code.emitPop()
		case DefglobalSymbol:
			code.emitDefGlobal(Cadr(instr))
		case DefmacroSymbol:
			code.emitDefMacro(Cadr(instr))
		case UseSymbol:
			code.emitUse(Cadr(instr))
		default:
			panic(fmt.Sprintf("Bad instruction: %v", op))
		}
		lst = Cdr(lst)
	}
	return nil
}

func (code *Code) emitLiteral(val Value) {
	code.ops = append(code.ops, opcodeLiteral)
	code.ops = append(code.ops, putConstant(val))
}

func (code *Code) emitGlobal(sym Value) {
	code.ops = append(code.ops, opcodeGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitCall(argc int) {
	code.ops = append(code.ops, opcodeCall)
	code.ops = append(code.ops, argc)
}
func (code *Code) emitReturn() {
	code.ops = append(code.ops, opcodeReturn)
}
func (code *Code) emitTailCall(argc int) {
	code.ops = append(code.ops, opcodeTailCall)
	code.ops = append(code.ops, argc)
}
func (code *Code) emitPop() {
	code.ops = append(code.ops, opcodePop)
}
func (code *Code) emitLocal(i int, j int) {
	code.ops = append(code.ops, opcodeLocal)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *Code) emitSetLocal(i int, j int) {
	code.ops = append(code.ops, opcodeSetLocal)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *Code) emitDefGlobal(sym Value) {
	code.ops = append(code.ops, opcodeDefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitUndefGlobal(sym Value) {
	code.ops = append(code.ops, opcodeUndefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitDefMacro(sym Value) {
	code.ops = append(code.ops, opcodeDefMacro)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitClosure(newCode Value) {
	code.ops = append(code.ops, opcodeClosure)
	code.ops = append(code.ops, putConstant(newCode))
}
func (code *Code) emitJumpFalse(offset int) int {
	code.ops = append(code.ops, opcodeJumpFalse)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *Code) emitJump(offset int) int {
	code.ops = append(code.ops, opcodeJump)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *Code) setJumpLocation(loc int) {
	code.ops[loc] = len(code.ops) - loc + 1
}
func (code *Code) emitVector(alen int) {
	code.ops = append(code.ops, opcodeVector)
	code.ops = append(code.ops, alen)
}
func (code *Code) emitStruct(slen int) {
	code.ops = append(code.ops, opcodeStruct)
	code.ops = append(code.ops, slen)
}
func (code *Code) emitUse(sym Value) {
	code.ops = append(code.ops, opcodeUse)
	code.ops = append(code.ops, putConstant(sym))
}
