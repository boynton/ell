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
	if verbose {
		Println("; begin execution")
	}
	code := thunk.(*lcode)
	vm := NewVM(DefaultStackSize)
	result, err := vm.exec(code)
	if err == nil {
		exports := vm.Exported()
		module := code.module
		module.exports = exports[:]
	}
	if verbose {
		//Println("; end execution")
		if err == nil && result != nil {
			Println("; => ", result)
		}
	}
	return result, err
}

type LVM interface {
	exec(code *lcode) (LObject, error)
	Exported() []LSymbol
}

type lvm struct {
	stackSize int
	defs      []LSymbol
}

func NewVM(stackSize int) LVM {
	defs := make([]LSymbol, 0)
	vm := lvm{stackSize, defs}
	return &vm
}

//pity: the module is rarely used, adds noticeable overhead
//the only current use for it is the "use" primitive, which must load and run more code.
//Perhaps that can be special cased as an op, so the other primitives can run faster. But,
//it is language-dependent (ell vs scheme)
type Primitive func(argv []LObject, argc int) (LObject, error)

type lprimitive struct {
	name string
	fun  Primitive
}

var symPrimitive = newSymbol("primitive")

func (prim *lprimitive) Type() LSymbol {
	return symPrimitive
}

func (prim *lprimitive) String() string {
	return "<primitive " + prim.name + ">"
}

type lframe struct {
	previous  *lframe
	pc        int
	ops       []int
	locals    *lframe
	elements  []LObject
	module    *lmodule
	constants []LObject
}

func (frame lframe) String() string {
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

type lclosure struct {
	code  *lcode
	frame *lframe
}

func (lclosure) Type() LSymbol {
	return Intern("closure")
}

func (closure lclosure) String() string {
	return "<closure: " + closure.code.String() + ">"
}

func showEnv(f *lframe) string {
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

func (vm *lvm) Exported() []LSymbol {
	return vm.defs
}

func (vm *lvm) exec(code *lcode) (LObject, error) {
	stack := make([]LObject, vm.stackSize)
	sp := vm.stackSize
	env := new(lframe)
	module := code.module
	ops := code.ops
	pc := 0
	if len(ops) == 0 {
		return nil, Error("No code to execute")
	}
	trace := false
	if trace {
		Println("------------------ BEGIN EXECUTION of ", module)
		Println(" ops: ", ops)
	}
	for {
		switch ops[pc] {
		case LITERAL_OPCODE:
			if trace {
				Println("const\t", module.constants[ops[pc+1]])
			}
			sp--
			stack[sp] = module.constants[ops[pc+1]]
			pc += 2
		case GLOBAL_OPCODE:
			if trace {
				Println("glob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			//val := (sym.(*lsymbol)).value
			val := module.Global(sym)
			if val == nil {
				return nil, Error("Undefined symbol:", sym)
			}
			sp--
			stack[sp] = val
			pc += 2
		case DEFGLOBAL_OPCODE:
			if trace {
				Println("defglob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			//(sym.(*lsymbol)).value = stack[sp]
			module.DefGlobal(sym, stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
			pc += 2
		case LOCAL_OPCODE:
			if trace {
				Println("getloc\t", +ops[pc+1], " ", ops[pc+2])
			}
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
			if trace {
				Println("setloc\t", +ops[pc+1], " ", ops[pc+2])
			}
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
			if trace {
				Println("call\t", ops[pc+1])
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			savedPc := pc + 2
			switch tfun := fun.(type) {
			case *lprimitive:
				//context for error reporting: tfun.name
				val, err := tfun.fun(stack[sp:], argc)
				if err != nil {
					//to do: fix to throw an Ell continuation-based error
					return nil, err
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = savedPc
			case *lclosure:
				f := new(lframe)
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
			if trace {
				Println("tcall\t", ops[pc+1])
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			switch tfun := fun.(type) {
			case *lprimitive:
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
			case *lclosure:
				newEnv := new(lframe)
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
			if trace {
				Println("ret")
			}
			if env.previous == nil {
				if trace {
					Println("------------------ END EXECUTION of ", module)
				}
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			module = env.module
			env = env.previous
		case JUMPFALSE_OPCODE:
			if trace {
				Println("fjmp\t", ops[pc+1])
			}
			b := stack[sp]
			sp++
			if b == FALSE {
				pc += ops[pc+1]
			} else {
				pc += 2
			}
		case JUMP_OPCODE:
			if trace {
				Println("jmp\t", ops[pc+1])
			}
			pc += ops[pc+1]
		case POP_OPCODE:
			if trace {
				Println("pop")
			}
			sp++
			pc++
		case CLOSURE_OPCODE:
			if trace {
				Println("closure\t", module.constants[ops[pc+1]])
			}
			sp--
			stack[sp] = &lclosure{module.constants[ops[pc+1]].(*lcode), env}
			pc = pc + 2
		case USE_OPCODE:
			if trace {
				Println("use\t", module.constants[ops[pc+1]])
			}
			if trace {
				Println(" -> pc before:", pc, ", ops:", ops)
			}
			sym := module.constants[ops[pc+1]]
			err := module.Use(sym)
			if err != nil {
				return nil, err
			}
			if trace {
				Println(" -> pc after:", pc, ", ops:", ops)
			}
			pc += 2
		case CAR_OPCODE:
			if trace {
				Println("car")
			}
			stack[sp] = Car(stack[sp])
			pc++
		case CDR_OPCODE:
			if trace {
				Println("cdr")
			}
			stack[sp] = Cdr(stack[sp])
			pc++
		case NULL_OPCODE:
			if trace {
				Println("null")
			}
			if stack[sp] == NIL {
				stack[sp] = TRUE
			} else {
				stack[sp] = FALSE
			}
			pc++
		case ADD_OPCODE:
			if trace {
				Println("add")
			}
			v, err := Add(stack[sp], stack[sp+1])
			if err != nil {
				return nil, err
			}
			sp++
			stack[sp] = v
			pc++
		case MUL_OPCODE:
			if trace {
				Println("mul")
			}
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
