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

package gell

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	LITERAL_OPCODE   = 1
	LOCAL_OPCODE     = 2
	JUMPFALSE_OPCODE = 3
	JUMP_OPCODE      = 4
	TAILCALL_OPCODE  = 5
	CALL_OPCODE      = 6
	RETURN_OPCODE    = 7
	CLOSURE_OPCODE   = 8
	POP_OPCODE       = 9
	GLOBAL_OPCODE    = 10
	DEFGLOBAL_OPCODE = 11
	SETLOCAL_OPCODE  = 12

	NULLP_OPCODE = 13
	CAR_OPCODE   = 14
	CDR_OPCODE   = 15
	ADD_OPCODE   = 16
	MUL_OPCODE   = 17
)

type LCode interface {
	Type() LSymbol
	String() string
	LoadOps(ops LObject) LError
	EmitLiteral(val LObject)
	EmitGlobal(sym LObject)
	EmitCall(argc int)
	EmitReturn()
	EmitTailCall(argc int)
	EmitPop()
	EmitLocal(i int, j int)
	EmitSetLocal(i int, j int)
	EmitDefGlobal(sym LObject)
	EmitClosure(code LCode)
	EmitJumpFalse(offset int) int
	EmitJump(offset int) int
	SetJumpLocation(loc int)
	EmitCar()
	EmitCdr()
	EmitNullP()
	EmitAdd()
	EmitMul()
}

type lcode struct {
	module *tModule
	ops    []int
	argc   int
	rest   LObject
}

func NewCode(module LModule, argc int, restArgs LObject) LCode {
	ops := make([]int, 0)
	mod := module.(*tModule)
	code := lcode{mod, ops, argc, restArgs}
	return &code
}

func (lcode) Type() LSymbol {
	return Intern("code")
}

func (code lcode) decompile(buf *bytes.Buffer, indent string) {
	offset := 0
	max := len(code.ops)
	buf.WriteString("(function ")
	buf.WriteString(strconv.Itoa(code.argc))
	if true {
		buf.WriteString(fmt.Sprintf(" %v)", code.ops)) //HACK
		return
	}
	for offset < max {
		switch code.ops[offset] {
		case LITERAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(literal %d)  \t; %v\n", indent, offset, code.ops[offset+1], code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (literal " + Write(code.module.constants[code.ops[offset+1]]) + ")")
			offset += 2
		case GLOBAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(global %v)\n", indent, offset, code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (global " + Write(code.module.constants[code.ops[offset+1]]) + ")")
			offset += 2
		case CALL_OPCODE:
			//fmt.Printf("%sL%03d:\t(call %d)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(" (call " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case TAILCALL_OPCODE:
			//fmt.Printf("%s%03d:\t(tailcall %d)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(" (tailcall " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case POP_OPCODE:
			//fmt.Printf("%sL%03d:\t(pop)\n", indent, offset)
			buf.WriteString(" (pop)")
			offset += 1
		case RETURN_OPCODE:
			//fmt.Printf("%sL%03d:\t(return)\n", indent, offset)
			buf.WriteString(" (return)")
			offset += 1
		case CLOSURE_OPCODE:
			//fmt.Printf("%sL%03d:\t(closure %v)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(" (closure ")
			(code.module.constants[code.ops[offset+1]].(*lcode)).decompile(buf, indent+"\t")
			buf.WriteString(")")
			offset += 2
		case LOCAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(local %d %d)\n", indent, offset, code.ops[offset+1], code.ops[offset+2])
			buf.WriteString(" (local " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case DEFGLOBAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(defglobal%6d ; %v)\n", indent, offset, code.ops[offset+1], code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (defglobal " + Write(code.module.constants[code.ops[offset+1]]) + ")")
			offset += 2
		case SETLOCAL_OPCODE:
			//Println("%sL%03d:\t(setlocal %d %d)\n", indent, offset, code.ops[offset+1], code.ops[offset+2])
			buf.WriteString(" (setlocal " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case JUMPFALSE_OPCODE:
			//fmt.Printf("%sL%03d:\t(jumpfalse %d)\t; L%03d\n", indent, offset, code.ops[offset+1], code.ops[offset+1] + offset)
			buf.WriteString(" (jumpfalse " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case JUMP_OPCODE:
			//fmt.Printf("%sL%03d:\t(jump %d)    \t; L%03d\n", indent, offset, code.ops[offset+1], code.ops[offset+1] + offset)
			buf.WriteString(" (jump " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case CAR_OPCODE:
			buf.WriteString(" (car)")
			offset += 1
		case CDR_OPCODE:
			buf.WriteString(" (cdr)")
			offset += 1
		case NULLP_OPCODE:
			buf.WriteString(" (null?)")
			offset += 1
		case ADD_OPCODE:
			buf.WriteString(" (add)")
			offset += 1
		case MUL_OPCODE:
			buf.WriteString(" (mul)")
			offset += 1
		default:
			buf.WriteString("FIX ME: " + strconv.Itoa(code.ops[offset]))
			break
		}
	}
	buf.WriteString(")")
}

func (code lcode) String() string {
	var buf bytes.Buffer
	//buf.WriteString(fmt.Sprintf("<code ops:%v constants:%v>", code.ops, code.module.constants)) //HACK
	//buf.WriteString(fmt.Sprintf("<code ops:%v>", code.ops)) //HACK
	code.decompile(&buf, "")
	return buf.String()
}

func (code *lcode) LoadOps(lst LObject) LError {
	return Error("LoadOps NYI")
}

func (code *lcode) EmitLiteral(val LObject) {
	code.ops = append(code.ops, LITERAL_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(val))
}

func (code *lcode) EmitGlobal(sym LObject) {
	code.ops = append(code.ops, GLOBAL_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(sym))
}
func (code *lcode) EmitCall(argc int) {
	code.ops = append(code.ops, CALL_OPCODE)
	code.ops = append(code.ops, argc)
}
func (code *lcode) EmitReturn() {
	code.ops = append(code.ops, RETURN_OPCODE)
}
func (code *lcode) EmitTailCall(argc int) {
	code.ops = append(code.ops, TAILCALL_OPCODE)
	code.ops = append(code.ops, argc)
}
func (code *lcode) EmitPop() {
	code.ops = append(code.ops, POP_OPCODE)
}
func (code *lcode) EmitLocal(i int, j int) {
	code.ops = append(code.ops, LOCAL_OPCODE)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *lcode) EmitSetLocal(i int, j int) {
	code.ops = append(code.ops, SETLOCAL_OPCODE)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *lcode) EmitDefGlobal(sym LObject) {
	code.ops = append(code.ops, DEFGLOBAL_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(sym))
}
func (code *lcode) EmitClosure(newCode LCode) {
	code.ops = append(code.ops, CLOSURE_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(newCode))
}
func (code *lcode) EmitJumpFalse(offset int) int {
	code.ops = append(code.ops, JUMPFALSE_OPCODE)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *lcode) EmitJump(offset int) int {
	code.ops = append(code.ops, JUMP_OPCODE)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *lcode) SetJumpLocation(loc int) {
	code.ops[loc] = len(code.ops) - loc + 1
}
func (code *lcode) EmitCar() {
	code.ops = append(code.ops, CAR_OPCODE)
}
func (code *lcode) EmitCdr() {
	code.ops = append(code.ops, CDR_OPCODE)
}
func (code *lcode) EmitNullP() {
	code.ops = append(code.ops, NULLP_OPCODE)
}
func (code *lcode) EmitAdd() {
	code.ops = append(code.ops, ADD_OPCODE)
}
func (code *lcode) EmitMul() {
	code.ops = append(code.ops, MUL_OPCODE)
}
