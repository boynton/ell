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
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	opcodeBad     = iota
	opcodeLiteral // 1
	opcodeLocal
	opcodeJumpFalse
	opcodeJump
	opcodeTailCall // 5
	opcodeCall
	opcodeReturn
	opcodeClosure
	opcodePop
	opcodeGlobal // 10
	opcodeDefGlobal
	opcodeSetLocal
	opcodeUse
	opcodeDefMacro
	opcodeVector // 15
	opcodeStruct
	opcodeUndefGlobal
	opcodeCount //18
)

//all syms for ops should be 7 chars or less
var symOpLiteral = intern("literal").(*LSymbol)
var symOpLocal = intern("local").(*LSymbol)
var symOpJumpFalse = intern("jumpfalse").(*LSymbol)
var symOpJump = intern("jump").(*LSymbol)
var symOpTailCall = intern("tailcall").(*LSymbol)
var symOpCall = intern("call").(*LSymbol)
var symOpReturn = intern("return").(*LSymbol)
var symOpClosure = intern("closure").(*LSymbol)
var symOpPop = intern("pop").(*LSymbol)
var symOpGlobal = intern("global").(*LSymbol)
var symOpDefGlobal = intern("defglobal").(*LSymbol)
var symOpSetLocal = intern("setlocal").(*LSymbol)
var symOpUse = intern("use").(*LSymbol)
var symOpDefMacro = intern("defmacro").(*LSymbol)
var symOpVector = intern("vector").(*LSymbol)
var symOpStruct = intern("struct").(*LSymbol)
var symOpUndefGlobal = intern("undefine").(*LSymbol)

var symOpFunction = intern("func").(*LSymbol)

var opsyms = initOpsyms()

func initOpsyms() []*LSymbol {
	syms := make([]*LSymbol, opcodeCount)
	syms[opcodeLiteral] = symOpLiteral
	syms[opcodeLocal] = symOpLocal
	syms[opcodeJumpFalse] = symOpJumpFalse
	syms[opcodeJump] = symOpJump
	syms[opcodeTailCall] = symOpTailCall
	syms[opcodeCall] = symOpCall
	syms[opcodeReturn] = symOpReturn
	syms[opcodeClosure] = symOpClosure
	syms[opcodePop] = symOpPop
	syms[opcodeGlobal] = symOpGlobal
	syms[opcodeDefGlobal] = symOpDefGlobal
	syms[opcodeSetLocal] = symOpSetLocal
	syms[opcodeUse] = symOpUse
	syms[opcodeDefMacro] = symOpDefMacro
	syms[opcodeVector] = symOpVector
	syms[opcodeStruct] = symOpStruct
	syms[opcodeUndefGlobal] = symOpUndefGlobal
	return syms
}

// LCode - compiled Ell bytecode
type LCode struct { // <code>
	name     string
	ops      []int
	argc     int
	defaults []LOB
	keys     []LOB
}

// CodeType - the Type object for this kind of value
var CodeType = intern("<code>")

// Type returns the type of the object
func (*LCode) Type() LOB {
	return CodeType
}

// Value returns the object itself for primitive types
func (code *LCode) Value() LOB {
	return code
}

// Equal returns true if the object is equal to the argument
func (code *LCode) Equal(another LOB) bool {
	return another == code
}

// String returns the string representation of the object
func (code *LCode) String() string {
	return code.decompile(true)
}

func newCode(argc int, defaults []LOB, keys []LOB, name string) *LCode {
	var ops []int
	//defaults == nil for normal procs, empty for rest, and non-empty for optional/keyword
	return &LCode{name, ops, argc, defaults, keys}
}

func (code *LCode) signature() string {
	//
	//experimental: external annotations on the functions: *declarations* is a map from symbol to string
	//
	//a macro to declare could be as follows:
	// (defmacro declare (name args)
	//   (or (symbol? name) (error "expected a <symbol> got " name))
	//   (let ((sig (string args)))
	//     `(assoc! *declarations* '~name '~sig)))
	//used as:
	// (declare cons (<any> <list>) <list>)
	if code.name != "" {
		sym, _ := intern("*declarations*").(*LSymbol)
		val := global(sym) //so if this this has not been defined, we'll just skip it
		if val != nil {
			strct, ok := val.(*LStruct)
			if ok {
				sig, _ := get(strct, intern(code.name))
				if sig != Null {
					return sig.String()
				}
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

func (code *LCode) decompile(pretty bool) string {
	var buf bytes.Buffer
	code.decompileInto(&buf, "", pretty)
	s := buf.String()
	return strings.Replace(s, "(function (\"\" 0 [] [])", "(code", 1)
}

func (code *LCode) decompileInto(buf *bytes.Buffer, indent string, pretty bool) {
	indentAmount := "   "
	offset := 0
	max := len(code.ops)
	prefix := " "
	buf.WriteString(indent + "(function (")
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
		s := prefix + "(" + opsyms[op].String()
		switch op {
		case opcodePop, opcodeReturn:
			buf.WriteString(s + ")")
			offset++
		case opcodeLiteral, opcodeDefGlobal, opcodeUse, opcodeGlobal, opcodeUndefGlobal, opcodeDefMacro:
			buf.WriteString(s + " " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeCall, opcodeTailCall, opcodeJumpFalse, opcodeJump, opcodeVector, opcodeStruct:
			buf.WriteString(s + " " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeLocal, opcodeSetLocal:
			buf.WriteString(s + " " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opcodeClosure:
			buf.WriteString(s + ")")
			if pretty {
				buf.WriteString("\n")
			} else {
				buf.WriteString(" ")
			}
			indent2 := ""
			if pretty {
				indent2 = indent + indentAmount
			}
			c, _ := constants[code.ops[offset+1]].(*LCode)
			c.decompileInto(buf, indent2, pretty)
			buf.WriteString(")")
			offset += 2
		default:
			panic(fmt.Sprintf("Bad instruction: %d", code.ops[offset]))
		}
	}
	buf.WriteString(")")
}

func (code *LCode) loadOps(lst *LList) error {
	for lst != EmptyList {
		o := car(lst)
		instr, ok := o.(*LList)
		if !ok {
			return Error(SyntaxErrorKey, o)
		}
		op := car(instr)
		switch op {
		case symOpClosure:
			lstFunc, ok := cadr(instr).(*LList)
			if !ok {
				return Error(SyntaxErrorKey, o)
			}
			if car(lstFunc) != symOpFunction {
				return Error(SyntaxErrorKey, instr)
			}
			lstFunc = cdr(lstFunc)
			funcParams := car(lstFunc)
			var argc int
			var name string
			var defaults []LOB
			var keys []LOB
			var err error
			tmp, ok := funcParams.(*LList)
			if ok && listLength(tmp) == 4 {
				a := car(tmp)
				tmp = cdr(tmp)
				name, err = asString(a)
				if err != nil {
					return Error(SyntaxErrorKey, funcParams)
				}
				a = car(tmp)
				tmp = cdr(tmp)
				argc, err = intValue(a)
				if err != nil {
					return Error(SyntaxErrorKey, funcParams)
				}
				a = car(tmp)
				tmp = cdr(tmp)
				if vec, ok := a.(*LVector); ok {
					defaults = vec.elements
				}
				a = car(tmp)
				if vec, ok := a.(*LVector); ok {
					keys = vec.elements
				}
			} else {
				return Error(SyntaxErrorKey, funcParams)
			}
			fun := newCode(argc, defaults, keys, name)
			fun.loadOps(cdr(lstFunc))
			code.emitClosure(fun)
		case symOpLiteral:
			code.emitLiteral(cadr(instr))
		case symOpLocal:
			i, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			j, err := intValue(caddr(instr))
			if err != nil {
				return err
			}
			code.emitLocal(i, j)
		case symOpSetLocal:
			i, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			j, err := intValue(caddr(instr))
			if err != nil {
				return err
			}
			code.emitSetLocal(i, j)
		case symOpGlobal:
			tmp := cadr(instr)
			if sym, ok := tmp.(*LSymbol); ok {
				code.emitGlobal(sym)
			} else {
				return Error(symOpGlobal, " argument 1 not a symbol: ", tmp)
			}
		case symOpUndefGlobal:
			code.emitUndefGlobal(cadr(instr))
		case symOpJump:
			loc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitJump(loc)
		case symOpJumpFalse:
			loc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitJumpFalse(loc)
		case symOpCall:
			argc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitCall(argc)
		case symOpTailCall:
			argc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitTailCall(argc)
		case symOpReturn:
			code.emitReturn()
		case symOpPop:
			code.emitPop()
		case symOpDefGlobal:
			code.emitDefGlobal(cadr(instr))
		case symOpDefMacro:
			code.emitDefMacro(cadr(instr))
		case symOpUse:
			code.emitUse(cadr(instr))
		default:
			panic(fmt.Sprintf("Bad instruction: %v", op))
		}
		lst = cdr(lst)
	}
	return nil
}

func (code *LCode) emitLiteral(val LOB) {
	code.ops = append(code.ops, opcodeLiteral)
	code.ops = append(code.ops, putConstant(val))
}

func (code *LCode) emitGlobal(sym LOB) {
	code.ops = append(code.ops, opcodeGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitCall(argc int) {
	code.ops = append(code.ops, opcodeCall)
	code.ops = append(code.ops, argc)
}
func (code *LCode) emitReturn() {
	code.ops = append(code.ops, opcodeReturn)
}
func (code *LCode) emitTailCall(argc int) {
	code.ops = append(code.ops, opcodeTailCall)
	code.ops = append(code.ops, argc)
}
func (code *LCode) emitPop() {
	code.ops = append(code.ops, opcodePop)
}
func (code *LCode) emitLocal(i int, j int) {
	code.ops = append(code.ops, opcodeLocal)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *LCode) emitSetLocal(i int, j int) {
	code.ops = append(code.ops, opcodeSetLocal)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *LCode) emitDefGlobal(sym LOB) {
	code.ops = append(code.ops, opcodeDefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitUndefGlobal(sym LOB) {
	code.ops = append(code.ops, opcodeUndefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitDefMacro(sym LOB) {
	code.ops = append(code.ops, opcodeDefMacro)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitClosure(newCode *LCode) {
	code.ops = append(code.ops, opcodeClosure)
	code.ops = append(code.ops, putConstant(newCode))
}
func (code *LCode) emitJumpFalse(offset int) int {
	code.ops = append(code.ops, opcodeJumpFalse)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *LCode) emitJump(offset int) int {
	code.ops = append(code.ops, opcodeJump)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *LCode) setJumpLocation(loc int) {
	code.ops[loc] = len(code.ops) - loc + 1
}
func (code *LCode) emitVector(alen int) {
	code.ops = append(code.ops, opcodeVector)
	code.ops = append(code.ops, alen)
}
func (code *LCode) emitStruct(slen int) {
	code.ops = append(code.ops, opcodeStruct)
	code.ops = append(code.ops, slen)
}
func (code *LCode) emitUse(name LOB) {
	code.ops = append(code.ops, opcodeUse)
	code.ops = append(code.ops, putConstant(name))
}
