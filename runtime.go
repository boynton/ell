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
)

var trace bool

func setTrace(b bool) {
	trace = b
}

func print(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Print(args[max])
}

func println(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Println(args[max])
}

// ArgcError - returns an error describing an argument count mismatch.
func ArgcError(name string, expected string, got int) error {
	return Error("Wrong number of arguments to ", name, " (expected ", expected, ", got ", got, ")")
}

// ArgTypeError - returns an error describing an argument type mismatch
func ArgTypeError(expected string, num int, arg LAny) error {
	return Error("Argument ", num, " is not of type <", expected, ">: ", arg)
}

func isFunction(obj LAny) bool {
	switch obj.(type) {
	case *LClosure, *LPrimitive, LInstruction:
		return true
	default:
		return false
	}
}

const defaultStackSize = 1000

var inExec = false
var conses = 0

func exec(code *Code, args ...LAny) (LAny, error) {
	if verbose {
		println("; begin execution")
		inExec = true
		conses = 0
	}
	vm := newVM(defaultStackSize)
	if len(args) != code.argc {
		return nil, Error("Wrong number of arguments")
	}
	result, err := vm.exec(code, args)
	if verbose {
		inExec = false
		println("; end execution")
		println("; total cons cells allocated: ", conses)
	}
	if err != nil {
		return nil, err
	}
	if verbose {
		//Println("; end execution")
		if err == nil && result != nil {
			println("; => ", result)
		}
	}
	return result, err
}

// VM - the Ell VM
type VM struct {
	stackSize int
}

func newVM(stackSize int) *VM {
	return &VM{stackSize}
}

// LInstruction - a primitive instruction for the VM
type LInstruction int // <function>

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = LInstruction(0)

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = LInstruction(1)

// Type returns the type of the object
func (LInstruction) Type() LAny {
	return typeFunction
}

// Value returns the object itself for primitive types
func (i LInstruction) Value() LAny {
	return i
}

// Equal returns true if the object is equal to the argument
func (i LInstruction) Equal(another LAny) bool {
	if a, ok := another.(LInstruction); ok {
		return i == a
	}
	return false
}

func (i LInstruction) String() string {
	switch i {
	case 0:
		return "#[function apply (<function> <any>* <list>)]"
	case 1:
		return "#[function callcc (<function>)]"
	}
	return fmt.Sprintf("#[function UNDEFINED]", i)
}

func (i LInstruction) Copy() LAny {
	return i
}

type primitive func(argv []LAny, argc int) (LAny, error)

// LPrimitive - a primitive function, written in Go, callable by VM
type LPrimitive struct { // <function>
	name string
	fun  primitive
	signature string
}

var typeFunction = intern("<function>")

// Type returns the type of the object
func (prim *LPrimitive) Type() LAny {
	return typeFunction
}

// Value returns the object itself for primitive types
func (prim *LPrimitive) Value() LAny {
	return prim
}

// Equal returns true if the object is equal to the argument
func (prim *LPrimitive) Equal(another LAny) bool {
	if a, ok := another.(*LPrimitive); ok {
		return prim == a
	}
	return false
}

func (prim *LPrimitive) String() string {
	return "#[function " + prim.name + " " + prim.signature + "]"
}

func (prim *LPrimitive) Copy() LAny {
	return prim
}

type frame struct {
	previous  *frame
	pc        int
	ops       []int
	locals    *frame
	elements  []LAny
	constants []LAny
	code *Code
}

func (frame frame) String() string {
	var buf bytes.Buffer
	buf.WriteString("#[frame")
	tmpEnv := &frame
	for tmpEnv != nil {
		buf.WriteString(fmt.Sprintf(" %v", tmpEnv.elements))
		tmpEnv = tmpEnv.locals
	}
	buf.WriteString("]")
	return buf.String()
}

// LClosure - an Ell closure formed over some compiled code and the current environment
type LClosure struct { // <function>
	code  *Code
	frame *frame
}

// Type returns the type of the object
func (LClosure) Type() LAny {
	return typeFunction
}

// Value returns the object itself for primitive types
func (closure *LClosure) Value() LAny {
	return closure
}

// Equal returns true if the object is equal to the argument
func (closure *LClosure) Equal(another LAny) bool {
	if a, ok := another.(*LClosure); ok {
		return closure == a
	}
	return false
}

func (closure *LClosure) Copy() LAny {
	//note: this isn't really a copy! closed-over state can still be mutated
	return closure
}

func (closure LClosure) String() string {
	n := closure.code.name
	s := closure.code.signature()
	if n == "" {
		return fmt.Sprintf("#[function %s]", s)
	}
	return fmt.Sprintf("#[function %s %s]", n, s)
}

func showEnv(f *frame) string {
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

func showStack(stack []LAny, sp int) string {
	end := len(stack)
	s := "["
	for sp < end {
		s = s + fmt.Sprintf(" %v", stack[sp])
		sp++
	}
	return s + " ]"
}

func buildFrame(env *frame, pc int, ops []int, fun *LClosure, argc int, stack []LAny, sp int) (*frame, error) {
	f := new(frame)
	f.previous = env
	f.pc = pc
	f.ops = ops
	f.locals = fun.frame
	f.code = fun.code
	expectedArgc := fun.code.argc
	defaults := fun.code.defaults
	if defaults == nil {
		if argc != expectedArgc {
			return nil, Error("Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
		}
		el := make([]LAny, argc)
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
		if extra > 0 {
			return nil, Error("Wrong number of args to ", fun, " (expected at least ", expectedArgc, ", got ", argc, ")")
		}
		return nil, Error("Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
	}
	totalArgc := expectedArgc + extra
	el := make([]LAny, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = listFromValues(restElements)
	} else if keys != nil {
		bindings := stack[sp+expectedArgc : sp+argc]
		if len(bindings)%2 != 0 {
			return nil, Error("Bad keyword argument(s): ", bindings)
		}
		copy(el, stack[sp:sp+expectedArgc]) //the required ones
		for i := expectedArgc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
		for i := expectedArgc; i < argc; i += 2 {
			key, err := keywordToSymbol(stack[sp+i])
			if err != nil {
				return nil, Error("Bad keyword argument: ", stack[sp+1])
			}
			gotit := false
			for j := 0; j < extra; j++ {
				if keys[j] == key {
					el[expectedArgc+j] = stack[sp+i+1]
					gotit = true
					break
				}
			}
			if !gotit {
				return nil, Error("Undefined keyword argument: ", key)
			}
		}
	} else {
		copy(el, stack[sp:sp+argc])
		for i := argc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
	}
	f.elements = el
	return f, nil
}

func addContext(env *frame, err error) error {
	if env.code == nil || env.code.name == "" {
		return err
	}
	return Error("[", env.code.name, "] ", err.Error())
}

func (vm *VM) exec(code *Code, args []LAny) (LAny, error) {
	stack := make([]LAny, vm.stackSize)
	sp := vm.stackSize
	env := new(frame)
	env.elements = make([]LAny, len(args))
	copy(env.elements, args)
	ops := code.ops
	pc := 0
	if len(ops) == 0 {
		return nil, addContext(env, Error("No code to execute"))
	}
	if trace {
		println("------------------ BEGIN EXECUTION of ", code)
		println("    stack: ", showStack(stack, sp))
	}
	for {
		switch ops[pc] {
		case opcodeLiteral:
			if trace {
				println(pc, "\tconst\t", constants[ops[pc+1]])
			}
			sp--
			stack[sp] = constants[ops[pc+1]]
			pc += 2
		case opcodeGlobal:
			if trace {
				println(pc, "\tgLAny\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			val := global(sym)
			if val == nil {
				return nil, addContext(env, Error("Undefined symbol: ", sym))
			}
			sp--
			stack[sp] = val
			pc += 2
		case opcodeDefGlobal:
			if trace {
				println(pc, "\tdefgLAny\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			defGlobal(sym, stack[sp])
			pc += 2
		case opcodeUndefGlobal:
			if trace {
				println(pc, "\tungLAny\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			undefGlobal(sym)
			pc += 2
		case opcodeDefMacro:
			if trace {
				println(pc, "\tdefmacro\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			defMacro(sym, stack[sp])
			stack[sp] = sym
			pc += 2
		case opcodeLocal:
			if trace {
				println(pc, "\tgetloc\t", +ops[pc+1], " ", ops[pc+2])
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
		case opcodeSetLocal:
			if trace {
				println(pc, "\tsetloc\t", +ops[pc+1], " ", ops[pc+2])
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
		case opcodeCall:
			if trace {
				println(pc, "\tcall\t", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
			savedPc := pc + 2
		opcodeCallAgain:
			switch tfun := fun.(type) {
			case *LPrimitive:
				val, err := tfun.fun(stack[sp:sp+argc], argc)
				if err != nil {
					//to do: fix to throw an Ell continuation-based error
					return nil, addContext(env, err)
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = savedPc
			case *LClosure:
				if checkInterrupt() {
					return nil, addContext(env, Error("Interrupt"))
				}
				f, err := buildFrame(env, savedPc, ops, tfun, argc, stack, sp)
				if err != nil {
					return nil, addContext(env, err)
				}
				sp += argc
				env = f
				ops = tfun.code.ops
				pc = 0
			case LInstruction:
				if tfun == Apply {
					if argc < 2 {
						err := ArgcError("apply", "2+", argc)
						return nil, addContext(env, err)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						err := ArgTypeError("list", argc, args)
						return nil, addContext(env, err)
					}
					arglist := args.(*LList)
					for i := argc - 2; i > 0; i-- {
						arglist = cons(stack[sp+i], arglist)
					}
					sp += argc
					argc = length(arglist)
					i := 0
					sp -= argc
					for arglist != EmptyList {
						stack[sp+i] = arglist.car
						i++
						arglist = arglist.cdr
					}
					goto opcodeCallAgain
				} else {
					return nil, addContext(env, Error("unsupported instruction", tfun))
				}
			case *LSymbol:
				if isKeyword(tfun) {
					if argc != 1 {
						err := ArgcError(tfun.Name, "1", argc)
						return nil, addContext(env, err)
					}
					v, err := get(stack[sp], fun)
					if err != nil {
						return nil, addContext(env, err)
					}
					stack[sp] = v
					pc = savedPc
				} else {
					return nil, addContext(env, Error("Not a function: ", tfun))
				}
			default:
				return nil, addContext(env, Error("Not a function: ", tfun))
			}
		case opcodeTailCall:
			if checkInterrupt() {
				return nil, addContext(env, Error("Interrupt"))
			}
			if trace {
				println(pc, "\ttcall\t", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
		opcodeTailCallAgain:
			switch tfun := fun.(type) {
			case *LPrimitive:
				val, err := tfun.fun(stack[sp:sp+argc], argc)
				if err != nil {
					return nil, addContext(env, err)
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = env.pc
				ops = env.ops
				env = env.previous
				if env == nil {
					if trace {
						println("------------------ END EXECUTION")
					}
					return stack[sp], nil
				}
			case *LClosure:
				f, err := buildFrame(env.previous, env.pc, env.ops, tfun, argc, stack, sp)
				if err != nil {
					return nil, addContext(env, err)
				}
				sp += argc
				ops = tfun.code.ops
				code = tfun.code
				pc = 0
				env = f
			case LInstruction:
				if tfun == Apply {
					if argc < 2 {
						err := ArgcError("apply", "2+", argc)
						return nil, addContext(env, err)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						err := ArgTypeError("list", argc, args)
						return nil, addContext(env, err)
					}
					arglist := args.(*LList)
					for i := argc - 2; i > 0; i-- {
						arglist = cons(stack[sp+i], arglist)
					}
					sp += argc
					argc = length(arglist)
					i := 0
					sp -= argc
					for arglist != EmptyList {
						stack[sp+i] = arglist.car
						i++
						arglist = arglist.cdr
					}
					goto opcodeTailCallAgain
				} else {
					return nil, addContext(env, Error("unsupported instruction", tfun))
				}
			case *LSymbol:
				if isKeyword(tfun) {
					if argc != 1 {
						err := ArgcError(tfun.Name, "1", argc)
						return nil, addContext(env, err)
					}
					v, err := get(stack[sp], fun)
					if err != nil {
						return nil, addContext(env, err)
					}
					stack[sp] = v
					pc = env.pc
					ops = env.ops
					env = env.previous
					if env == nil {
						if trace {
							println("------------------ END EXECUTION")
						}
						return stack[sp], nil
					}
				} else {
					return nil, addContext(env, Error("Not a function: ", tfun))
				}
			default:
				return nil, addContext(env, Error("Not a function:", tfun))
			}
		case opcodeReturn:
			if checkInterrupt() {
				return nil, addContext(env, Error("Interrupt"))
			}
			if trace {
				println(pc, "\tret")
			}
			if env.previous == nil {
				if trace {
					println("------------------ END EXECUTION")
				}
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			env = env.previous
		case opcodeJumpFalse:
			if trace {
				println(pc, "\tfjmp\t", pc+ops[pc+1])
			}
			b := stack[sp]
			sp++
			if b == False {
				pc += ops[pc+1]
			} else {
				pc += 2
			}
		case opcodeJump:
			if trace {
				println(pc, "\tjmp\t", pc+ops[pc+1])
			}
			pc += ops[pc+1]
		case opcodePop:
			if trace {
				println(pc, "\tpop")
			}
			sp++
			pc++
		case opcodeClosure:
			if trace {
				println(pc, "\tclosure\t", constants[ops[pc+1]])
			}
			sp--
			stack[sp] = &LClosure{constants[ops[pc+1]].(*Code), env}
			pc = pc + 2
		case opcodeUse:
			if trace {
				println(pc, "\tuse\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			err := use(sym)
			if err != nil {
				return nil, addContext(env, err)
			}
			sp--
			stack[sp] = sym
			pc += 2
		case opcodeVector:
			if trace {
				println(pc, "\tvector\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v := vector(stack[sp:sp+vlen]...)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		case opcodeStruct:
			if trace {
				println(pc, "\tstruct\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v, _ := newStruct(stack[sp : sp+vlen])
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		default:
			return nil, addContext(env, Error("Bad instruction: ", strconv.Itoa(ops[pc])))
		}
	}
}
