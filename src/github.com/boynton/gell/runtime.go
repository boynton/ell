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

func Print(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Print(args[max])
}
func Println(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Println(args[max])
}

const DefaultStackSize = 1000

func Exec(thunk LCode) (LObject, error) {
	code := thunk.(*lcode)
	vm := LVM{DefaultStackSize, nil}
	return vm.exec(code)
}

type LVM struct {
	stackSize int
	defs      []LSymbol
}

type tFrame struct {
	previous  *tFrame
	pc        int
	ops       []int
	locals    *tFrame
	elements  []LObject
	module    *tModule
	constants []LObject
}

func (frame tFrame) String() string {
	var buf bytes.Buffer
	buf.WriteString("<frame")
	tmpEnv := &frame
	for tmpEnv != nil {
		buf.WriteString(fmt.Sprintf(" %v", tmpEnv.elements))
		tmpEnv = tmpEnv.locals
	}
	buf.WriteString(">")
	return buf.String()
}

type tClosure struct {
	code  *lcode
	frame *tFrame
}

func (tClosure) Type() LSymbol {
	return Intern("closure") //optimize!
}
func (closure tClosure) String() string {
	return "<closure: " + closure.code.String() + ">"
}

func showEnv(f *tFrame) string {
	tmp := f
	s := ""
	for {
		s = s + fmt.Sprintf(" %v", tmp.elements)
		if tmp.locals == nil {
			break
		}
		tmp = tmp.locals
	}
	return s
}
func showStack(stack []LObject, sp int) string {
	end := len(stack)
	s := "["
	for sp < end {
		s = s + fmt.Sprintf(" %v", stack[sp])
		sp++
	}
	return s + " ]"
}

func (vm LVM) exec(code *lcode) (LObject, LError) {
	stack := make([]LObject, vm.stackSize)
	sp := vm.stackSize
	env := new(tFrame)
	module := code.module
	ops := code.ops
	pc := 0
	if len(ops) == 0 {
		return nil, Error("No code to execute")
	}
	for {
		switch ops[pc] {
		case LITERAL_OPCODE:
			sp--
			stack[sp] = module.constants[ops[pc+1]]
			pc += 2
		case GLOBAL_OPCODE:
			sym := module.constants[ops[pc+1]]
			val := (sym.(*lsymbol)).value
			if val == nil {
				return nil, Error("Undefined symbol:", sym)
			}
			sp--
			stack[sp] = val
			pc += 2
		case DEFGLOBAL_OPCODE:
			sym := module.globalDefine(ops[pc+1], stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
			pc += 2
		case LOCAL_OPCODE:
			tmpEnv := env
			i := ops[pc+1]
			for i > 0 {
				tmpEnv = tmpEnv.locals
				i--
			}
			j := ops[pc+2]
			val := tmpEnv.elements[j]
			sp--
			stack[sp] = val
			pc += 3
		case SETLOCAL_OPCODE:
			tmpEnv := env
			i := ops[pc+1]
			for i > 0 {
				tmpEnv = tmpEnv.locals
				i--
			}
			j := ops[pc+2]
			tmpEnv.elements[j] = stack[sp]
			pc += 3
		case CALL_OPCODE:
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			savedPc := pc + 2
			switch tfun := fun.(type) {
			case tPrimitive:
				//context for error reporting: tfun.name
				val, err := tfun.fun(stack[sp:], argc)
				if err != nil {
					//to do: fix to throw an Ell continuation-based error
					return nil, err
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = savedPc
			case tClosure:
				f := new(tFrame)
				f.previous = env
				f.pc = savedPc
				f.ops = ops
				f.module = module
				f.locals = tfun.frame
				if tfun.code.argc >= 0 {
					if tfun.code.argc != argc {
						return nil, Error("Wrong number of args ("+strconv.Itoa(ops[pc+1])+") to", tfun)
					}
					f.elements = make([]LObject, argc)
					if argc > 0 {
						copy(f.elements, stack[sp:sp+argc])
						sp += argc
					}
				} else {
					return nil, Error("rest args NYI")
				}
				env = f
				ops = tfun.code.ops
				module = tfun.code.module
				pc = 0
			default:
				return nil, Error("Not a function:", tfun)
			}
		case TAILCALL_OPCODE:
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			switch tfun := fun.(type) {
			case tPrimitive:
				//context for error reporting: tfun.name
				val, err := tfun.fun(stack[sp:], argc)
				if err != nil {
					return nil, err
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = env.pc
				ops = env.ops
				module = env.module
				env = env.previous
			case tClosure:
				newEnv := new(tFrame)
				newEnv.previous = env.previous
				newEnv.pc = env.pc
				newEnv.ops = env.ops
				newEnv.module = env.module
				newEnv.locals = tfun.frame
				if tfun.code.argc >= 0 {
					if tfun.code.argc != argc {
						return nil, Error("Wrong number of args ("+strconv.Itoa(ops[pc+1])+") to", tfun)
					}
					newEnv.elements = make([]LObject, argc)
					copy(newEnv.elements, stack[sp:sp+argc])
					sp += argc
				} else {
					return nil, Error("rest args NYI")
				}
				ops = tfun.code.ops
				module = tfun.code.module
				pc = 0
				env = newEnv
			default:
				return nil, Error("Not a function:", tfun)
			}
		case RETURN_OPCODE:
			if env.previous == nil {
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			module = env.module
			env = env.previous
		case JUMPFALSE_OPCODE:
			b := stack[sp]
			sp++
			if b == FALSE {
				pc += ops[pc+1]
			} else {
				pc += 2
			}
		case JUMP_OPCODE:
			pc += ops[pc+1]
		case POP_OPCODE:
			sp++
			pc++
		case CLOSURE_OPCODE:
			sp--
			stack[sp] = tClosure{module.constants[ops[pc+1]].(*lcode), env}
			pc = pc + 2
		case CAR_OPCODE:
			stack[sp] = Car(stack[sp])
			pc++
		case CDR_OPCODE:
			stack[sp] = Cdr(stack[sp])
			pc++
		case NULLP_OPCODE:
			if stack[sp] == NIL {
				stack[sp] = TRUE
			} else {
				stack[sp] = FALSE
			}
			pc++
		case ADD_OPCODE:
			v, err := Add(stack[sp], stack[sp+1])
			if err != nil {
				return nil, err
			}
			sp++
			stack[sp] = v
			pc++
		case MUL_OPCODE:
			v, err := Mul(stack[sp], stack[sp+1])
			if err != nil {
				return nil, err
			}
			sp++
			stack[sp] = v
			pc++
		default:
			return nil, Error("Bad instruction:", strconv.Itoa(ops[pc]))
		}
	}
	return nil, nil //never happens
}
