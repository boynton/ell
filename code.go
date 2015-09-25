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
	opcodePrimCall
)

var symOpClosure = intern("closure")
var symOpFunction = intern("function")
var symOpLiteral = intern("literal")
var symOpLocal = intern("local")
var symOpSetLocal = intern("setlocal")
var symOpGlobal = intern("global")
var symOpJump = intern("jump")
var symOpJumpFalse = intern("jumpfalse")
var symOpCall = intern("call")
var symOpTailCall = intern("tailcall")
var symOpPrimCall = intern("primcall")
var symOpReturn = intern("return")
var symOpPop = intern("pop")
var symOpDefGlobal = intern("defglobal")
var symOpUndefGlobal = intern("undefglobal")
var symOpDefMacro = intern("defmacro")
var symOpUse = intern("use")

// LCode - compiled Ell bytecode
type LCode struct {
	name     string
	ops      []int
	argc     int
	defaults []*LAny
	keys     []*LAny
}

func newCode(argc int, defaults []*LAny, keys []*LAny, name string) *LAny {
	var ops []int
	code := &LCode{
		name,
		ops,
		argc,
		defaults, //nil for normal procs, empty for rest, and non-empty for optional/keyword
		keys,
	}
	result := new(LAny)
	result.ltype = typeCode
	result.code = code
	return result
}

func (code *LCode) signature() string {
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
		val := global(intern("*declarations*")) //so if this this has not been defined, we'll just skip it
		if val != nil && isStruct(val) {
			sig, _ := get(val, intern(code.name))
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
		switch code.ops[offset] {
		case opcodeLiteral:
			buf.WriteString(prefix + "(literal " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeDefGlobal:
			buf.WriteString(prefix + "(defglobal " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeCall:
			buf.WriteString(prefix + "(call " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeTailCall:
			buf.WriteString(prefix + "(tailcall " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodePrimCall:
			buf.WriteString(prefix + "(primcall " + strconv.Itoa(code.ops[offset+1]) + " " + primitives[code.ops[offset+2]].name + ")")
			offset += 3
		case opcodePop:
			buf.WriteString(prefix + "(pop)")
			offset++
		case opcodeReturn:
			buf.WriteString(prefix + "(return)")
			offset++
		case opcodeClosure:
			buf.WriteString(prefix + "(closure")
			if pretty {
				buf.WriteString("\n")
			} else {
				buf.WriteString(" ")
			}
			indent2 := ""
			if pretty {
				indent2 = indent + indentAmount
			}
			constants[code.ops[offset+1]].code.decompileInto(buf, indent2, pretty)
			buf.WriteString(")")
			offset += 2
		case opcodeLocal:
			buf.WriteString(prefix + "(local " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opcodeGlobal:
			buf.WriteString(prefix + "(global " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeUndefGlobal:
			buf.WriteString(prefix + "(undefglobal " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeDefMacro:
			buf.WriteString(prefix + "(defmacro " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeSetLocal:
			buf.WriteString(prefix + "(setlocal " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opcodeJumpFalse:
			buf.WriteString(prefix + "(jumpfalse " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeJump:
			buf.WriteString(prefix + "(jump " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeVector:
			buf.WriteString(prefix + "(vector " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeStruct:
			buf.WriteString(prefix + "(struct " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeUse:
			buf.WriteString(prefix + "(use " + constants[code.ops[offset+1]].String() + ")")
			offset += 2
		default:
			panic(fmt.Sprintf("Bad instruction: %d", code.ops[offset]))
		}
	}
	buf.WriteString(")")
}

func (code *LCode) String() string {
	return code.decompile(true)
	//	return fmt.Sprintf("(function (%d %v %s) %v)", code.argc, code.defaults, code.keys, code.ops)
}

func (code *LCode) loadOps(lst *LAny) error {
	for lst != EmptyList {
		instr := car(lst)
		op := car(instr)
		switch op {
		case symOpClosure:
			lstFunc := cadr(instr)
			if car(lstFunc) != symOpFunction {
				return Error("Bad argument for a closure: ", lstFunc)
			}
			lstFunc = cdr(lstFunc)
			funcParams := car(lstFunc)
			var argc int
			var name string
			var defaults []*LAny
			var keys []*LAny
			var err error
			if isSymbol(funcParams) {
				//legacy form, just the argc
				argc, err = intValue(funcParams)
				if err != nil {
					return err
				}
				if argc < 0 {
					argc = -argc - 1
					defaults = make([]*LAny, 0)
				}
			} else if isList(funcParams) && length(funcParams) == 4 {
				tmp := funcParams
				a := car(tmp)
				tmp = cdr(tmp)
				name, err = asString(a)
				if err != nil {
					return Error("Bad code format: ", funcParams)
				}
				a = car(tmp)
				tmp = cdr(tmp)
				argc, err = intValue(a)
				if err != nil {
					return Error("Bad code format: ", funcParams)
				}
				a = car(tmp)
				tmp = cdr(tmp)
				if isVector(a) {
					defaults = a.elements
				}
				a = car(tmp)
				if isVector(a) {
					keys = a.elements
				}
			} else {
				return Error("Bad code format: ", funcParams)
			}
			fun := newCode(argc, defaults, keys, name)
			fun.code.loadOps(cdr(lstFunc))
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
			sym := cadr(instr)
			if isSymbol(sym) {
				code.emitGlobal(sym)
			} else {
				return Error(symOpGlobal, " argument 1 not a symbol: ", sym)
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

func (code *LCode) emitLiteral(val *LAny) {
	code.ops = append(code.ops, opcodeLiteral)
	code.ops = append(code.ops, putConstant(val))
}

func (code *LCode) emitGlobal(sym *LAny) {
	code.ops = append(code.ops, opcodeGlobal)
	code.ops = append(code.ops, int(sym.ival))
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
func (code *LCode) emitPrimCall(prim *Primitive, argc int) {
	code.ops = append(code.ops, opcodePrimCall)
	code.ops = append(code.ops, prim.idx)
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
func (code *LCode) emitDefGlobal(sym *LAny) {
	code.ops = append(code.ops, opcodeDefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitUndefGlobal(sym *LAny) {
	code.ops = append(code.ops, opcodeUndefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitDefMacro(sym *LAny) {
	code.ops = append(code.ops, opcodeDefMacro)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *LCode) emitClosure(newCode *LAny) {
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
func (code *LCode) emitUse(sym *LAny) {
	code.ops = append(code.ops, opcodeUse)
	code.ops = append(code.ops, putConstant(sym))
}
