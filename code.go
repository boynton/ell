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
	opcodeArray // 15
	opcodeStruct

	//extended instructions
	opcodeNull
	opcodeCar
	opcodeCdr
	opcodeAdd // 20
	opcodeMul

	opcodeUndefGlobal
)

// Code - compiled Ell bytecode
type Code struct {
	name           string
	ops            []int
	argc           int
	defaults       []LAny
	keys           []LAny
	symClosure     LAny
	symFunction    LAny
	symLiteral     LAny
	symLocal       LAny
	symSetLocal    LAny
	symGlobal      LAny
	symJump        LAny
	symJumpFalse   LAny
	symCall        LAny
	symTailCall    LAny
	symReturn      LAny
	symPop         LAny
	symDefGlobal   LAny
	symUndefGlobal LAny
	symDefMacro    LAny
	symUse         LAny
	symCar         LAny
	symCdr         LAny
	symNull        LAny
	symAdd         LAny
	symMul         LAny
}

func newCode(argc int, defaults []LAny, keys []LAny, name string) *Code {
	var ops []int
	code := Code{
		name,
		ops,
		argc,
		defaults, //nil for normal procs, empty for rest, and non-empty for optional/keyword
		keys,
		intern("closure"),
		intern("function"),
		intern("literal"),
		intern("local"),
		intern("setlocal"),
		intern("global"),
		intern("jump"),
		intern("jumpfalse"),
		intern("call"),
		intern("tailcall"),
		intern("return"),
		intern("pop"),
		intern("defglobal"),
		intern("undefglobal"),
		intern("defmacro"),
		intern("use"),
		intern("car"),
		intern("cdr"),
		intern("null"),
		intern("add"),
		intern("mul"),
	}
	return &code
}

var typeCode = intern("<code>")

// Type returns the type of the code
func (*Code) Type() LAny {
	return typeCode
}

// Value returns the object itself for primitive types
func (code *Code) Value() LAny {
	return code
}

// Equal returns true if the object is equal to the argument
func (code *Code) Equal(another LAny) bool {
	if c, ok := another.(*Code); ok {
		return code == c
	}
	return false
}

func (code *Code) decompile(pretty bool) string {
	var buf bytes.Buffer
	code.decompileInto(&buf, "", pretty)
	s := buf.String()
	return strings.Replace(s, "(function (0 [] [])", "(lap", 1)
}

func (code *Code) decompileInto(buf *bytes.Buffer, indent string, pretty bool) {
	indentAmount := "   "
	offset := 0
	max := len(code.ops)
	begin := " "
	buf.WriteString(indent + "(function (")
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
		begin = "\n" + indent
	}
	for offset < max {
		switch code.ops[offset] {
		case opcodeLiteral:
			//fmt.Printf("%sL%03d:\t(literal %d)  \t; %v\n", indent, offset, code.ops[offset+1], constants[code.ops[offset+1]])
			buf.WriteString(begin + "(literal " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeDefGlobal:
			//fmt.Printf("%sL%03d:\t(global %v)\n", indent, offset, constants[code.ops[offset+1]])
			buf.WriteString(begin + "(defglobal " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeCall:
			//fmt.Printf("%sL%03d:\t(call %d)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(begin + "(call " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeTailCall:
			//fmt.Printf("%s%03d:\t(tailcall %d)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(begin + "(tailcall " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodePop:
			//fmt.Printf("%sL%03d:\t(pop)\n", indent, offset)
			buf.WriteString(begin + "(pop)")
			offset++
		case opcodeReturn:
			//fmt.Printf("%sL%03d:\t(return)\n", indent, offset)
			buf.WriteString(begin + "(return)")
			offset++
		case opcodeClosure:
			//fmt.Printf("%sL%03d:\t(closure %v)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(begin + "(closure")
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
		case opcodeLocal:
			//fmt.Printf("%sL%03d:\t(local %d %d)\n", indent, offset, code.ops[offset+1], code.ops[offset+2])
			buf.WriteString(begin + "(local " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opcodeGlobal:
			//fmt.Printf("%sL%03d:\t(global %v)\n", indent, offset, constants[code.ops[offset+1]])
			buf.WriteString(begin + "(global " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeUndefGlobal:
			//fmt.Printf("%sL%03d:\t(unglobal %v)\n", indent, offset, constants[code.ops[offset+1]])
			buf.WriteString(begin + "(undefglobal " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeDefMacro:
			//fmt.Printf("%sL%03d:\t(defmacro%6d ; %v)\n", indent, offset, code.ops[offset+1], constants[code.ops[offset+1]])
			buf.WriteString(begin + "(defmacro " + write(constants[code.ops[offset+1]]) + ")")
			offset += 2
		case opcodeSetLocal:
			//Println("%sL%03d:\t(setlocal %d %d)\n", indent, offset, code.ops[offset+1], code.ops[offset+2])
			buf.WriteString(begin + "(setlocal " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case opcodeJumpFalse:
			//fmt.Printf("%sL%03d:\t(jumpfalse %d)\t; L%03d\n", indent, offset, code.ops[offset+1], code.ops[offset+1] + offset)
			buf.WriteString(begin + "(jumpfalse " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeJump:
			//fmt.Printf("%sL%03d:\t(jump %d)    \t; L%03d\n", indent, offset, code.ops[offset+1], code.ops[offset+1] + offset)
			buf.WriteString(begin + "(jump " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeArray:
			buf.WriteString(begin + "(array " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeStruct:
			buf.WriteString(begin + "(struct " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case opcodeUse:
			buf.WriteString(begin + "(use " + constants[code.ops[offset+1]].String() + ")")
			offset += 2
		case opcodeCar:
			buf.WriteString(begin + "(car)")
			offset++
		case opcodeCdr:
			buf.WriteString(begin + "(cdr)")
			offset++
		case opcodeNull:
			buf.WriteString(begin + "(null)")
			offset++
		case opcodeAdd:
			buf.WriteString(begin + "(add)")
			offset++
		case opcodeMul:
			buf.WriteString(begin + "(mul)")
			offset++
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

func (code *Code) loadOps(lst LAny) error {
	name := ""
	for lst != EmptyList {
		instr := car(lst)
		op := car(instr)
		switch op {
		case code.symClosure:
			lstFunc := cadr(instr)
			if car(lstFunc) != code.symFunction {
				return Error("Bad argument for a closure: ", lstFunc)
			}
			lstFunc = cdr(lstFunc)
			funcParams := car(lstFunc)
			var argc int
			var defaults []LAny
			var keys []LAny
			var err error
			if isSymbol(funcParams) {
				//legacy form, just the argc
				argc, err = intValue(funcParams)
				if err != nil {
					return err
				}
				if argc < 0 {
					argc = -argc - 1
					defaults = make([]LAny, 0)
				}
			} else if isList(funcParams) && length(funcParams) == 3 {
				a := car(funcParams)
				argc, err = intValue(a)
				if err != nil {
					return Error("Bad lap format: ", funcParams)
				}
				b := cadr(funcParams)
				if ary, ok := b.(*LArray); ok {
					defaults = ary.elements
				}
				c := caddr(funcParams)
				if ary, ok := c.(*LArray); ok {
					keys = ary.elements
				}
			} else {
				return Error("Bad lap format: ", funcParams)
			}
			fun := newCode(argc, defaults, keys, name)
			fun.loadOps(cdr(lstFunc))
			code.emitClosure(fun)
		case code.symLiteral:
			code.emitLiteral(cadr(instr))
		case code.symLocal:
			i, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			j, err := intValue(caddr(instr))
			if err != nil {
				return err
			}
			code.emitLocal(i, j)
		case code.symSetLocal:
			i, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			j, err := intValue(caddr(instr))
			if err != nil {
				return err
			}
			code.emitSetLocal(i, j)
		case code.symGlobal:
			code.emitGlobal(cadr(instr))
		case code.symUndefGlobal:
			code.emitUndefGlobal(cadr(instr))
		case code.symJump:
			loc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitJump(loc)
		case code.symJumpFalse:
			loc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitJumpFalse(loc)
		case code.symCall:
			argc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitCall(argc)
		case code.symTailCall:
			argc, err := intValue(cadr(instr))
			if err != nil {
				return err
			}
			code.emitTailCall(argc)
		case code.symReturn:
			code.emitReturn()
		case code.symPop:
			code.emitPop()
		case code.symDefGlobal:
			code.emitDefGlobal(cadr(instr))
		case code.symDefMacro:
			code.emitDefMacro(cadr(instr))
		case code.symUse:
			code.emitUse(cadr(instr))
		case code.symCar:
			code.emitCar()
		case code.symCdr:
			code.emitCdr()
		case code.symNull:
			code.emitNull()
		case code.symAdd:
			code.emitAdd()
		case code.symMul:
			code.emitMul()
		default:
			panic(fmt.Sprintf("Bad instruction: %v", op))
		}
		lst = cdr(lst)
	}
	return nil
}

func (code *Code) emitLiteral(val LAny) {
	code.ops = append(code.ops, opcodeLiteral)
	code.ops = append(code.ops, putConstant(val))
}

func (code *Code) emitGlobal(sym LAny) {
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
func (code *Code) emitDefGlobal(sym LAny) {
	code.ops = append(code.ops, opcodeDefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitUndefGlobal(sym LAny) {
	code.ops = append(code.ops, opcodeUndefGlobal)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitDefMacro(sym LAny) {
	code.ops = append(code.ops, opcodeDefMacro)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitClosure(newCode *Code) {
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
func (code *Code) emitArray(alen int) {
	code.ops = append(code.ops, opcodeArray)
	code.ops = append(code.ops, alen)
}
func (code *Code) emitStruct(slen int) {
	code.ops = append(code.ops, opcodeStruct)
	code.ops = append(code.ops, slen)
}
func (code *Code) emitUse(sym LAny) {
	code.ops = append(code.ops, opcodeUse)
	code.ops = append(code.ops, putConstant(sym))
}
func (code *Code) emitCar() {
	code.ops = append(code.ops, opcodeCar)
}
func (code *Code) emitCdr() {
	code.ops = append(code.ops, opcodeCdr)
}
func (code *Code) emitNull() {
	code.ops = append(code.ops, opcodeNull)
}
func (code *Code) emitAdd() {
	code.ops = append(code.ops, opcodeAdd)
}
func (code *Code) emitMul() {
	code.ops = append(code.ops, opcodeMul)
}
