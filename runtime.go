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

func ArgcError(name string, expected string, got int) (AnyType, error) {
	return nil, Error("Wrong number of arguments to ", name, " (expected ", expected, ", got ", got, ")")
}

func ArgTypeError(expected string, num int, arg AnyType) (AnyType, error) {
	return nil, Error("Argument ", num, " is not of type <", expected, ">: ", arg)
}

func isFunction(obj AnyType) bool {
	switch obj.(type) {
	//	case *Code, *Closure, *Primitive, Instruction:
	case *Closure, *Primitive, Instruction:
		return true
	default:
		return false
	}
}

const defaultStackSize = 1000

var inExec = false
var conses = 0

func exec(thunk code, args ...AnyType) (AnyType, error) {
	if verbose {
		println("; begin execution")
		inExec = true
		conses = 0
	}
	code := thunk.(*Code)
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
	exports := vm.exported()
	module := code.module().(*Module)
	module.exported = exports[:]
	if verbose {
		//Println("; end execution")
		if err == nil && result != nil {
			println("; => ", result)
		}
	}
	return result, err
}

type LVM struct {
	stackSize int
	defs      []AnyType
}

func newVM(stackSize int) *LVM {
	var defs []AnyType
	vm := LVM{stackSize, defs}
	return &vm
}

type Instruction int

// APPLY is a primitive instruction to apply a function to a list of arguments
var APPLY = Instruction(0)

// CALLCC is a primitive instruction to executable (restore) a continuation
var CALLCC = Instruction(1)

// Type returns the type of the object
func (Instruction) Type() AnyType {
	return typeFunction
}

// Equal returns true if the object is equal to the argument
func (s Instruction) Equal(another AnyType) bool {
	if a, ok := another.(Instruction); ok {
		return s == a
	}
	return false
}

func (s Instruction) String() string {
	switch s {
	case 0:
		return "<function apply>"
	case 1:
		return "<function callcc>"
	}
	return "<instr ?>"
}

type primitive func(argv []AnyType, argc int) (AnyType, error)

type Primitive struct {
	name string
	fun  primitive
}

var typeFunction = internSymbol("<function>")

// Type returns the type of the object
func (prim *Primitive) Type() AnyType {
	return typeFunction
}

// Equal returns true if the object is equal to the argument
func (prim *Primitive) Equal(another AnyType) bool {
	if a, ok := another.(*Primitive); ok {
		return prim == a
	}
	return false
}

func (prim *Primitive) String() string {
	return "<function " + prim.name + ">"
}

type frame struct {
	previous  *frame
	pc        int
	ops       []int
	locals    *frame
	elements  []AnyType
	module    *Module
	constants []AnyType
}

func (frame frame) String() string {
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

type Closure struct {
	code  *Code
	frame *frame
}

// Type returns the type of the object
func (Closure) Type() AnyType {
	return typeFunction
}

// Equal returns true if the object is equal to the argument
func (closure *Closure) Equal(another AnyType) bool {
	if a, ok := another.(*Closure); ok {
		return closure == a
	}
	return false
}

func (closure Closure) String() string {
	n := closure.code.name
	if n == "" {
		return "<function>"
	}
	return "<function " + n + ">"
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

func showStack(stack []AnyType, sp int) string {
	end := len(stack)
	s := "["
	for sp < end {
		s = s + fmt.Sprintf(" %v", stack[sp])
		sp++
	}
	return s + " ]"
}

func (vm *LVM) exported() []AnyType {
	return vm.defs
}

func buildFrame(env *frame, pc int, ops []int, module *Module, fun *Closure, argc int, stack []AnyType, sp int) (*frame, error) {
	f := new(frame)
	f.previous = env
	f.pc = pc
	f.ops = ops
	f.module = module
	f.locals = fun.frame
	expectedArgc := fun.code.argc
	defaults := fun.code.defaults
	if defaults == nil {
		if argc != expectedArgc {
			return nil, Error("Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
		}
		el := make([]AnyType, argc)
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
	el := make([]AnyType, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = toList(restElements)
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
			key, err := unkeyword(stack[sp+i])
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

func (vm *LVM) exec(code *Code, args []AnyType) (AnyType, error) {
	topmod := code.mod
	stack := make([]AnyType, vm.stackSize)
	sp := vm.stackSize
	env := new(frame)
	env.elements = make([]AnyType, len(args))
	copy(env.elements, args)
	module := code.mod
	currentModule = module
	ops := code.ops
	pc := 0
	if len(ops) == 0 {
		return nil, Error("No code to execute")
	}
	if trace {
		println("------------------ BEGIN EXECUTION of ", code)
		println("    stack: ", showStack(stack, sp))
		println("    module: ", module)
	}
	for {
		switch ops[pc] {
		case opcodeLiteral:
			if trace {
				println(pc, "\tconst\t", module.constants[ops[pc+1]])
			}
			sp--
			stack[sp] = module.constants[ops[pc+1]]
			pc += 2
		case opcodeGlobal:
			if trace {
				println(pc, "\tgAnyType\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			val := module.global(sym)
			if val == nil {
				return nil, Error("Undefined symbol: ", sym)
			}
			sp--
			stack[sp] = val
			pc += 2
		case opcodeDefGlobal:
			if trace {
				println(pc, "\tdefgAnyType\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.defGAnyTypeal(sym, stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
			pc += 2
		case opcodeUndefGlobal:
			if trace {
				println(pc, "\tungAnyType\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.undefGAnyTypeal(sym)
			pc += 2
		case opcodeDefMacro:
			if trace {
				println(pc, "\tdefmacro\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.defMacro(sym, stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
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
			case *Primitive:
				val, err := tfun.fun(stack[sp:sp+argc], argc)
				if err != nil {
					//to do: fix to throw an Ell continuation-based error
					return nil, err
				}
				sp = sp + argc - 1
				stack[sp] = val
				pc = savedPc
			case *Closure:
				if topmod.checkInterrupt() {
					return nil, Error("Interrupt")
				}
				f, err := buildFrame(env, savedPc, ops, module, tfun, argc, stack, sp)
				if err != nil {
					return nil, err
				}
				sp += argc
				env = f
				ops = tfun.code.ops
				module = tfun.code.mod
				pc = 0
			case Instruction:
				if tfun == APPLY {
					if argc < 2 {
						return ArgcError("apply", "2+", argc)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						return ArgTypeError("list", argc, args)
					}
					arglist := args.(*ListType)
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
					return nil, Error("unsupported instruction", tfun)
				}
			case *SymbolType:
				if isKeyword(tfun) {
					if argc != 1 {
						return ArgcError(tfun.Name, "1", argc)
					}
					v, err := get(stack[sp], fun)
					if err != nil {
						return nil, err
					}
					stack[sp] = v
					pc = savedPc
				} else {
					return nil, Error("Not a function: ", tfun)
				}
			default:
				return nil, Error("Not a function: ", tfun)
			}
		case opcodeTailCall:
			if topmod.checkInterrupt() {
				return nil, Error("Interrupt")
			}
			if trace {
				println(pc, "\ttcall\t", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
		opcodeTailCallAgain:
			switch tfun := fun.(type) {
			case *Primitive:
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
						println("------------------ END EXECUTION of ", module)
					}
					return stack[sp], nil
				}
			case *Closure:
				f, err := buildFrame(env.previous, env.pc, env.ops, env.module, tfun, argc, stack, sp)
				if err != nil {
					return nil, err
				}
				sp += argc
				ops = tfun.code.ops
				module = tfun.code.mod
				pc = 0
				env = f
			case Instruction:
				if tfun == APPLY {
					if argc < 2 {
						return ArgcError("apply", "2+", argc)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						return ArgTypeError("list", argc, args)
					}
					arglist := args.(*ListType)
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
					return nil, Error("unsupported instruction", tfun)
				}
			case *SymbolType:
				if isKeyword(tfun) {
					if argc != 1 {
						return ArgcError(tfun.Name, "1", argc)
					}
					v, err := get(stack[sp], fun)
					if err != nil {
						return nil, err
					}
					stack[sp] = v
					pc = env.pc
					ops = env.ops
					module = env.module
					env = env.previous
					if env == nil {
						if trace {
							println("------------------ END EXECUTION of ", module)
						}
						return stack[sp], nil
					}
				} else {
					return nil, Error("Not a function: ", tfun)
				}
			default:
				return nil, Error("Not a function:", tfun)
			}
		case opcodeReturn:
			if topmod.checkInterrupt() {
				return nil, Error("Interrupt")
			}
			if trace {
				println(pc, "\tret")
			}
			if env.previous == nil {
				if trace {
					println("------------------ END EXECUTION of ", module)
				}
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			module = env.module
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
				println(pc, "\tclosure\t", module.constants[ops[pc+1]])
			}
			sp--
			stack[sp] = &Closure{module.constants[ops[pc+1]].(*Code), env}
			pc = pc + 2
		case opcodeUse:
			if trace {
				println(pc, "\tuse\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			err := module.use(sym)
			if err != nil {
				return nil, err
			}
			sp--
			stack[sp] = sym
			pc += 2
		case opcodeArray:
			if trace {
				println(pc, "\tarray\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v := toArray(stack[sp:], vlen)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		case opcodeStruct:
			if trace {
				println(pc, "\tstruct\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			typesym := stack[sp].(*SymbolType)
			v, _ := newInstance(typesym, stack[sp+1:sp+vlen]) //To do: extend to other types
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		case opcodeCar:
			if trace {
				println(pc, "\tcar")
			}
			stack[sp] = car(stack[sp])
			pc++
		case opcodeCdr:
			if trace {
				println(pc, "\tcdr")
			}
			stack[sp] = cdr(stack[sp])
			pc++
		case opcodeNull:
			if trace {
				println(pc, "\tnull")
			}
			if stack[sp] == Null {
				stack[sp] = True
			} else {
				stack[sp] = False
			}
			pc++
		case opcodeAdd:
			if trace {
				println(pc, "\tadd")
			}
			v, err := add(stack[sp], stack[sp+1])
			if err != nil {
				return nil, err
			}
			sp++
			stack[sp] = v
			pc++
		case opcodeMul:
			if trace {
				println(pc, "\tmul")
			}
			v, err := mul(stack[sp], stack[sp+1])
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
}
