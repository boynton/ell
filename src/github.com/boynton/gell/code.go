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
	"strings"
)

const (
	BAD_OPCODE       = iota
	LITERAL_OPCODE   = iota // 1
	LOCAL_OPCODE     = iota
	JUMPFALSE_OPCODE = iota
	JUMP_OPCODE      = iota
	TAILCALL_OPCODE  = iota // 5
	CALL_OPCODE      = iota
	RETURN_OPCODE    = iota
	CLOSURE_OPCODE   = iota
	POP_OPCODE       = iota
	GLOBAL_OPCODE    = iota // 10
	DEFGLOBAL_OPCODE = iota
	SETLOCAL_OPCODE  = iota
	NULL_OPCODE      = iota
	CAR_OPCODE       = iota
	CDR_OPCODE       = iota // 15
	ADD_OPCODE       = iota
	MUL_OPCODE       = iota
	USE_OPCODE       = iota
	DEFMACRO_OPCODE  = iota
	VECTOR_OPCODE    = iota // 20
	MAP_OPCODE       = iota
)

type LCode interface {
	Type() LObject
	String() string
	Equal(another LObject) bool
	Module() LModule
	LoadOps(ops LObject) error
	Decompile() string
	EmitLiteral(val LObject)
	EmitGlobal(sym LObject)
	EmitCall(argc int)
	EmitReturn()
	EmitTailCall(argc int)
	EmitPop()
	EmitLocal(i int, j int)
	EmitSetLocal(i int, j int)
	EmitDefGlobal(sym LObject)
	EmitDefMacro(sym LObject)
	EmitClosure(code LCode)
	EmitJumpFalse(offset int) int
	EmitJump(offset int) int
	SetJumpLocation(loc int)
	EmitVector(length int)
	EmitMap(length int)
	EmitUse(sym LObject)
	EmitCar()
	EmitCdr()
	EmitNull()
	EmitAdd()
	EmitMul()
}

type lcode struct {
	module       *lmodule
	ops          []int
	argc         int
	rest         LObject
	symClosure   LObject
	symFunction  LObject
	symLiteral   LObject
	symLocal     LObject
	symSetLocal  LObject
	symGlobal    LObject
	symJump      LObject
	symJumpFalse LObject
	symCall      LObject
	symTailCall  LObject
	symReturn    LObject
	symPop       LObject
	symDefGlobal LObject
	symDefMacro  LObject
	symUse       LObject
	symCar       LObject
	symCdr       LObject
	symNull      LObject
	symAdd       LObject
	symMul       LObject
}

func NewCode(module LModule, argc int, restArgs LObject) LCode {
	ops := make([]int, 0)
	mod := module.(*lmodule)
	code := lcode{
		mod,
		ops,
		argc,
		restArgs,
		Intern("closure"),
		Intern("function"),
		Intern("literal"),
		Intern("local"),
		Intern("setlocal"),
		Intern("global"),
		Intern("jump"),
		Intern("jumpfalse"),
		Intern("call"),
		Intern("tailcall"),
		Intern("return"),
		Intern("pop"),
		Intern("defglobal"),
		Intern("defmacro"),
		Intern("use"),
		Intern("car"),
		Intern("cdr"),
		Intern("null"),
		Intern("add"),
		Intern("mul"),
	}
	return &code
}

func (*lcode) Type() LObject {
	return Intern("code")
}

func (code *lcode) Equal(another LObject) bool {
	if c, ok := another.(*lcode); ok {
		return code == c
	}
	return false
}

func (code *lcode) Decompile() string {
	var buf bytes.Buffer
	code.decompile(&buf, "")
	s := buf.String()
	return strings.Replace(s, "function 0", "lap", 1)
}

func (code *lcode) decompile(buf *bytes.Buffer, indent string) {
	offset := 0
	max := len(code.ops)
	buf.WriteString("(function ")
	buf.WriteString(strconv.Itoa(code.argc))
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
		case DEFMACRO_OPCODE:
			//fmt.Printf("%sL%03d:\t(defmacro%6d ; %v)\n", indent, offset, code.ops[offset+1], code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (defmacro " + Write(code.module.constants[code.ops[offset+1]]) + ")")
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
		case USE_OPCODE:
			buf.WriteString(" (use " + code.module.constants[code.ops[offset+1]].String() + ")")
			offset += 2
		case CAR_OPCODE:
			buf.WriteString(" (car)")
			offset += 1
		case CDR_OPCODE:
			buf.WriteString(" (cdr)")
			offset += 1
		case NULL_OPCODE:
			buf.WriteString(" (null)")
			offset += 1
		case ADD_OPCODE:
			buf.WriteString(" (add)")
			offset += 1
		case MUL_OPCODE:
			buf.WriteString(" (mul)")
			offset += 1
		default:
			panic(fmt.Sprintf("Bad instruction: %d", code.ops[offset]))
			break
		}
	}
	buf.WriteString(")")
}

func (code *lcode) String() string {
	//	return code.Decompile()
	return fmt.Sprintf("(function %d %v)", code.argc, code.ops)
}

func (code *lcode) Module() LModule {
	return code.module
}

func (code *lcode) LoadOps(lst LObject) error {
	for lst != NIL {
		instr := Car(lst)
		op := Car(instr)
		switch op {
		case code.symClosure:
			lstFunc := Cadr(instr)
			if Car(lstFunc) != code.symFunction {
				return Error("Bad argument for a closure: ", lstFunc)
			}
			lstFunc = Cdr(lstFunc)
			ac, err := IntValue(Car(lstFunc))
			if err != nil {
				return err
			}
			fun := NewCode(code.module, ac, nil)
			fun.LoadOps(Cdr(lstFunc))
			code.EmitClosure(fun)
		case code.symLiteral:
			code.EmitLiteral(Cadr(instr))
		case code.symLocal:
			i, err := IntValue(Cadr(instr))
			if err != nil {
				return err
			}
			j, err := IntValue(Caddr(instr))
			if err != nil {
				return err
			}
			code.EmitLocal(i, j)
		case code.symSetLocal:
			i, err := IntValue(Cadr(instr))
			if err != nil {
				return err
			}
			j, err := IntValue(Caddr(instr))
			if err != nil {
				return err
			}
			code.EmitSetLocal(i, j)
		case code.symGlobal:
			code.EmitGlobal(Cadr(instr))
		case code.symJump:
			loc, err := IntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.EmitJump(loc)
		case code.symJumpFalse:
			loc, err := IntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.EmitJumpFalse(loc)
		case code.symCall:
			argc, err := IntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.EmitCall(argc)
		case code.symTailCall:
			argc, err := IntValue(Cadr(instr))
			if err != nil {
				return err
			}
			code.EmitTailCall(argc)
		case code.symReturn:
			code.EmitReturn()
		case code.symPop:
			code.EmitPop()
		case code.symDefGlobal:
			code.EmitDefGlobal(Cadr(instr))
		case code.symDefMacro:
			code.EmitDefMacro(Cadr(instr))
		case code.symUse:
			code.EmitUse(Cadr(instr))
		case code.symCar:
			code.EmitCar()
		case code.symCdr:
			code.EmitCdr()
		case code.symNull:
			code.EmitNull()
		case code.symAdd:
			code.EmitAdd()
		case code.symMul:
			code.EmitMul()
		default:
			panic(fmt.Sprintf("Bad instruction: %v", op))
		}
		lst = Cdr(lst)
	}
	return nil
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
func (code *lcode) EmitDefMacro(sym LObject) {
	code.ops = append(code.ops, DEFMACRO_OPCODE)
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
func (code *lcode) EmitVector(vlen int) {
	code.ops = append(code.ops, VECTOR_OPCODE)
	code.ops = append(code.ops, vlen)
}
func (code *lcode) EmitMap(vlen int) {
	code.ops = append(code.ops, MAP_OPCODE)
	code.ops = append(code.ops, vlen)
}
func (code *lcode) EmitUse(sym LObject) {
	code.ops = append(code.ops, USE_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(sym))
}
func (code *lcode) EmitCar() {
	code.ops = append(code.ops, CAR_OPCODE)
}
func (code *lcode) EmitCdr() {
	code.ops = append(code.ops, CDR_OPCODE)
}
func (code *lcode) EmitNull() {
	code.ops = append(code.ops, NULL_OPCODE)
}
func (code *lcode) EmitAdd() {
	code.ops = append(code.ops, ADD_OPCODE)
}
func (code *lcode) EmitMul() {
	code.ops = append(code.ops, MUL_OPCODE)
}
