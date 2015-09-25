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
func ArgTypeError(expected string, num int, arg *LOB) error {
	return Error("Argument ", num, " is not of type <", expected, ">: ", arg)
}

// LFunction - all callable function subtypes are represented here
type LFunction struct {
	code        *LCode // closure
	frame       *Frame // closure
	instruction int
	primitive   *Primitive
}

func newClosure(code *LCode, frame *Frame) *LOB {
	clo := newLOB(typeFunction)
	fun := new(LFunction)
	fun.code = code
	fun.frame = frame
	clo.function = fun
	return clo
}

const defaultStackSize = 1000

var inExec = false
var conses = 0

func exec(code *LCode, args ...*LOB) (*LOB, error) {
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

const instructionNone = 0
const instructionApply = 1
const instructionCallCC = 2

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = &LOB{variant: typeFunction, function: &LFunction{instruction: instructionApply}}

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = &LOB{variant: typeFunction, function: &LFunction{instruction: instructionCallCC}}

func (f LFunction) String() string {
	if f.instruction == instructionNone {
		if f.primitive != nil { //primitive
			return "#[primitive-function " + f.primitive.name + " " + f.primitive.signature + "]"
		}
		if f.code != nil { // closure
			n := f.code.name
			s := f.code.signature()
			if n == "" {
				return fmt.Sprintf("#[function %s]", s)
			}
			return fmt.Sprintf("#[function %s %s]", n, s)
		}
		panic("Bad function")
	}
	if f.instruction == instructionApply {
		return "#[function apply (<function> <any>* <list>)]"
	}
	if f.instruction == instructionCallCC {
		return "#[function callcc (<function>)]"
	}
	panic("Bad function")
}

// PrimCallable is the native go function signature for all Ell primitive functions
type PrimCallable func(argv []*LOB, argc int) (*LOB, error)

// Primitive - a primitive function, written in Go, callable by VM
type Primitive struct { // <function>
	name      string
	fun       PrimCallable
	signature string
	idx       int
}

func newPrimitive(name string, fun PrimCallable, signature string) *LOB {
	idx := len(primitives)
	prim := &Primitive{name, fun, signature, idx}
	primitives = append(primitives, prim)
	return &LOB{variant: typeFunction, function: &LFunction{primitive: prim}}
}

// Frame - a call frame in the VM, as well as en environment frame for lexical closures
type Frame struct {
	previous  *Frame
	pc        int
	ops       []int
	locals    *Frame
	elements  []*LOB
	constants []*LOB
	code      *LCode
}

func (frame *Frame) String() string {
	var buf bytes.Buffer
	buf.WriteString("#[frame")
	tmpEnv := frame
	for tmpEnv != nil {
		buf.WriteString(fmt.Sprintf(" %v", tmpEnv.elements))
		tmpEnv = tmpEnv.locals
	}
	buf.WriteString("]")
	return buf.String()
}

func showEnv(f *Frame) string {
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

func showStack(stack []*LOB, sp int) string {
	end := len(stack)
	s := "["
	for sp < end {
		s = s + fmt.Sprintf(" %v", stack[sp])
		sp++
	}
	return s + " ]"
}

func buildFrame(env *Frame, pc int, ops []int, fun *LFunction, argc int, stack []*LOB, sp int) (*Frame, error) {
	f := new(Frame)
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
		el := make([]*LOB, argc)
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
	el := make([]*LOB, totalArgc)
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

func addContext(env *Frame, err error) error {
	if env.code == nil || env.code.name == "" {
		return err
	}
	return Error("[", env.code.name, "] ", err.Error())
}

func (vm *VM) exec(code *LCode, args []*LOB) (*LOB, error) {
	stack := make([]*LOB, vm.stackSize)
	sp := vm.stackSize
	env := new(Frame)
	env.elements = make([]*LOB, len(args))
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
				println(pc, "\tglob\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			if sym.car == nil {
				return nil, addContext(env, Error("Undefined symbol: ", sym))
			}
			sp--
			stack[sp] = sym.car
			pc += 2
		case opcodeDefGlobal:
			if trace {
				println(pc, "\tdef\t", constants[ops[pc+1]])
			}
			sym := constants[ops[pc+1]]
			defGlobal(sym, stack[sp])
			pc += 2
		case opcodeUndefGlobal:
			if trace {
				println(pc, "\tundef\t", constants[ops[pc+1]])
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
		case opcodePrimCall:
			if trace {
				println(pc, "\tprimcall\t", primitives[ops[pc+2]], " ", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			tfun := primitives[ops[pc+1]]
			argc := ops[pc+2]
			val, err := tfun.fun(stack[sp:sp+argc], argc)
			if err != nil {
				return nil, addContext(env, err)
			}
			sp = sp + argc - 1
			stack[sp] = val
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
			switch fun.variant {
			case typeFunction:
				lfun := fun.function
				if lfun.primitive != nil {
					val, err := lfun.primitive.fun(stack[sp:sp+argc], argc)
					if err != nil {
						//to do: fix to throw an Ell continuation-based error
						return nil, addContext(env, err)
					}
					sp = sp + argc - 1
					stack[sp] = val
					pc = savedPc
				} else if lfun.code != nil {
					if checkInterrupt() {
						return nil, addContext(env, Error("Interrupt"))
					}
					f, err := buildFrame(env, savedPc, ops, lfun, argc, stack, sp)
					if err != nil {
						return nil, addContext(env, err)
					}
					sp += argc
					env = f
					ops = lfun.code.ops
					pc = 0
				} else if fun == Apply {
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
					arglist := args
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
					return nil, addContext(env, Error("unsupported instruction", fun))
				}
			case typeKeyword:
				if argc != 1 {
					err := ArgcError(fun.text, "1", argc)
					return nil, addContext(env, err)
				}
				v, err := get(stack[sp], fun)
				if err != nil {
					return nil, addContext(env, err)
				}
				stack[sp] = v
				pc = savedPc
			default:
				return nil, addContext(env, Error("Not a function: ", fun))
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
			switch fun.variant {
			case typeFunction:
				lfun := fun.function
				if lfun.primitive != nil {
					val, err := lfun.primitive.fun(stack[sp:sp+argc], argc)
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
				} else if lfun.code != nil {
					f, err := buildFrame(env.previous, env.pc, env.ops, lfun, argc, stack, sp)
					if err != nil {
						return nil, addContext(env, err)
					}
					sp += argc
					code = lfun.code
					ops = code.ops
					pc = 0
					env = f
				} else if fun == Apply {
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
					arglist := args
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
					return nil, addContext(env, Error("unsupported instruction", fun))
				}
			case typeKeyword:
				if argc != 1 {
					err := ArgcError(fun.text, "1", argc)
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
			default:
				return nil, addContext(env, Error("Not a function:", fun))
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
			stack[sp] = newClosure(constants[ops[pc+1]].code, env)
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
			v := vector(stack[sp : sp+vlen]...)
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
