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

var trace bool

func SetTrace(b bool) {
	trace = b
}

func Print(args ...LAny) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Print(args[max])
}
func Println(args ...LAny) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Println(args[max])
}

func IsFunction(obj LObject) bool {
	switch obj.(type) {
	case *lcode, *lclosure, *lprimitive:
		return true
	default:
		return false
	}
}

const DefaultStackSize = 1000

func Exec(thunk LCode, args ...LObject) (LObject, error) {
	if verbose {
		Println("; begin execution")
	}
	code := thunk.(*lcode)
	vm := newVM(DefaultStackSize)
	if len(args) != code.argc {
		return nil, Error("Wrong number of arguments")
	}
	result, err := vm.exec(code, args)
	if verbose {
		Println("; end execution")
	}
	if err != nil {
		return nil, err
	}
	exports := vm.Exported()
	module := code.module
	module.exports = exports[:]
	if verbose {
		//Println("; end execution")
		if err == nil && result != nil {
			Println("; => ", result)
		}
	}
	return result, err
}

type lvm struct {
	stackSize int
	defs      []LObject
}

func newVM(stackSize int) *lvm {
	defs := make([]LObject, 0)
	vm := lvm{stackSize, defs}
	return &vm
}

type Primitive func(argv []LObject, argc int) (LObject, error)

type lprimitive struct {
	name string
	fun  Primitive
}

var symPrimitive = newSymbol("primitive")

func (prim *lprimitive) Type() LObject {
	return symPrimitive
}

func (prim *lprimitive) Equal(another LObject) bool {
	if a, ok := another.(*lprimitive); ok {
		return prim == a
	}
	return false
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

func (lclosure) Type() LObject {
	return Intern("closure")
}

func (closure *lclosure) Equal(another LObject) bool {
	if a, ok := another.(*lclosure); ok {
		return closure == a
	}
	return false
}

func (closure lclosure) String() string {
	//	return "<closure: " + closure.code.String() + ">"
	if closure.code.defaults != nil {
		return fmt.Sprintf("<function of %d or more arguments>", closure.code.argc)
	} else if closure.code.argc == 1 {
		return fmt.Sprintf("<function of 1 argument>")
	} else {
		return fmt.Sprintf("<function of %d arguments>", closure.code.argc)
	}
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

func (vm *lvm) Exported() []LObject {
	return vm.defs
}

func buildFrame(env *lframe, pc int, ops []int, module *lmodule, fun *lclosure, argc int, stack []LObject, sp int) (*lframe, error) {
	f := new(lframe)
	f.previous = env
	f.pc = pc
	f.ops = ops
	f.module = module
	f.locals = fun.frame
	expectedArgc := fun.code.argc
	defaults := fun.code.defaults
	if defaults == nil {
		if argc != expectedArgc {
			return nil, Error("Wrong number of args (", argc, ") to ", fun)
		}
		el := make([]LObject, argc)
		copy(el, stack[sp:sp+argc])
		f.elements = el
		return f, nil
	}
	keys := fun.code.keys
	rest := false
	extra := len(defaults)
	if extra == 0 {
		rest = true
		extra = 1
	}
	if argc < expectedArgc {
		return nil, Error("Wrong number of args (", argc, ") to ", fun)
	}
	totalArgc := expectedArgc + extra
	el := make([]LObject, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = ToList(restElements)
	} else if keys != nil {
		bindings := stack[sp+expectedArgc : sp+argc]
		if len(bindings)%2 != 0 {
			return nil, Error("Bad keyword argument(s): ", bindings)
		}
		copy(el, stack[sp:sp+expectedArgc]) //the required ones
		for i := expectedArgc; i < totalArgc; i++ {
			el[i+expectedArgc-1] = defaults[i-1]
		}
		for i := expectedArgc; i < argc; i += 2 {
			key := stack[sp+i]
			if !IsSymbol(key) {
				return nil, Error("Bad keyword for argument: ", key)
			}
			for j := 0; j < extra; j++ {
				if keys[j] == key {
					el[expectedArgc+j] = stack[sp+i+1]
					break
				}
			}
		}
	} else {
		copy(el, stack[sp:sp+argc])
		for i := argc; i < totalArgc; i++ {
			el[i+expectedArgc-1] = defaults[i-1]
		}
	}
	f.elements = el
	return f, nil
}

func (vm *lvm) exec(code *lcode, args []LObject) (LObject, error) {
	topmod := code.module
	stack := make([]LObject, vm.stackSize)
	sp := vm.stackSize
	env := new(lframe)
	env.elements = make([]LObject, len(args))
	copy(env.elements, args)
	module := code.module
	ops := code.ops
	pc := 0
	if len(ops) == 0 {
		return nil, Error("No code to execute")
	}
	if trace {
		Println("------------------ BEGIN EXECUTION of ", code)
		Println("    stack: ", showStack(stack, sp))
		Println("    module: ", module)
	}
	for {
		switch ops[pc] {
		case LITERAL_OPCODE:
			if trace {
				Println(pc, "\tconst\t", module.constants[ops[pc+1]])
			}
			sp--
			stack[sp] = module.constants[ops[pc+1]]
			pc += 2
		case GLOBAL_OPCODE:
			if trace {
				Println(pc, "\tglob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			val := module.Global(sym)
			if val == nil {
				return nil, Error("Undefined symbol: ", sym)
			}
			sp--
			stack[sp] = val
			pc += 2
		case DEFGLOBAL_OPCODE:
			if trace {
				Println(pc, "\tdefglob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.DefGlobal(sym, stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
			pc += 2
		case DEFMACRO_OPCODE:
			if trace {
				Println(pc, "\tdefmacro\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.DefMacro(sym, stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
			pc += 2
		case LOCAL_OPCODE:
			if trace {
				Println(pc, "\tgetloc\t", +ops[pc+1], " ", ops[pc+2])
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
				Println(pc, "\tsetloc\t", +ops[pc+1], " ", ops[pc+2])
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
				Println(pc, "\tcall\t", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			savedPc := pc + 2
			switch tfun := fun.(type) {
			case *lprimitive:
				val, err := tfun.fun(stack[sp:sp+argc], argc)
				if err != nil {
					//to do: fix to throw an Ell continuation-based error
					return nil, err
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = savedPc
			case *lclosure:
				if topmod.CheckInterrupt() {
					return nil, Error("Interrupt")
				}
				f, err := buildFrame(env, savedPc, ops, module, tfun, argc, stack, sp)
				if err != nil {
					return nil, err
				}
				sp += argc
				env = f
				ops = tfun.code.ops
				module = tfun.code.module
				pc = 0
			default:
				return nil, Error("Not a function: ", tfun)
			}
		case TAILCALL_OPCODE:
			if topmod.CheckInterrupt() {
				return nil, Error("Interrupt")
			}
			if trace {
				Println(pc, "\ttcall\t", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			switch tfun := fun.(type) {
			case *lprimitive:
				val, err := tfun.fun(stack[sp:sp+argc], argc)
				if err != nil {
					return nil, err
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = env.pc
				ops = env.ops
				module = env.module
				env = env.previous
				if env == nil {
					if trace {
						Println("------------------ END EXECUTION of ", module)
						Println(" -> sp:", sp)
					}
					return stack[sp], nil
				}
			case *lclosure:
				if env.previous == nil {
					if trace {
						Println("------------------ END EXECUTION of ", module)
						Println(" -> sp:", sp)
					}
					return stack[sp], nil
				}
				f, err := buildFrame(env.previous, env.pc, env.ops, env.module, tfun, argc, stack, sp)
				if err != nil {
					return nil, err
				}
				sp += argc
				ops = tfun.code.ops
				module = tfun.code.module
				pc = 0
				env = f
			default:
				return nil, Error("Not a function:", tfun)
			}
		case RETURN_OPCODE:
			if topmod.CheckInterrupt() {
				return nil, Error("Interrupt")
			}
			if trace {
				Println(pc, "\tret")
			}
			if env.previous == nil {
				if trace {
					Println("------------------ END EXECUTION of ", module)
					Println(" -> sp:", sp)
					Println(" = ", stack[sp])
				}
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			module = env.module
			env = env.previous
		case JUMPFALSE_OPCODE:
			if trace {
				Println(pc, "\tfjmp\t", pc+ops[pc+1])
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
				Println(pc, "\tjmp\t", pc+ops[pc+1])
			}
			pc += ops[pc+1]
		case POP_OPCODE:
			if trace {
				Println(pc, "\tpop")
			}
			sp++
			pc++
		case CLOSURE_OPCODE:
			if trace {
				Println(pc, "\tclosure\t", module.constants[ops[pc+1]])
			}
			sp--
			stack[sp] = &lclosure{module.constants[ops[pc+1]].(*lcode), env}
			pc = pc + 2
		case USE_OPCODE:
			if trace {
				Println(pc, "\tuse\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			err := module.Use(sym)
			if err != nil {
				return nil, err
			}
			sp--
			stack[sp] = sym
			pc += 2
		case VECTOR_OPCODE:
			if trace {
				Println(pc, "\tvec\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v := ToVector(stack[sp:], vlen)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		case MAP_OPCODE:
			if trace {
				Println(pc, "\tmap\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v, _ := ToMap(stack[sp:], vlen)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		case CAR_OPCODE:
			if trace {
				Println(pc, "\tcar")
			}
			stack[sp] = Car(stack[sp])
			pc++
		case CDR_OPCODE:
			if trace {
				Println(pc, "\tcdr")
			}
			stack[sp] = Cdr(stack[sp])
			pc++
		case NULL_OPCODE:
			if trace {
				Println(pc, "\tnull")
			}
			if stack[sp] == NIL {
				stack[sp] = TRUE
			} else {
				stack[sp] = FALSE
			}
			pc++
		case ADD_OPCODE:
			if trace {
				Println(pc, "\tadd")
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
				Println(pc, "\tmul")
			}
			v, err := Mul(stack[sp], stack[sp+1])
			if err != nil {
				return nil, err
			}
			sp++
			stack[sp] = v
			pc++
		default:
			return nil, Error("Bad instruction: ", strconv.Itoa(ops[pc]))
		}
	}
	return nil, nil //never happens
}
