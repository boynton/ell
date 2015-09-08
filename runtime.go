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

func print(args ...any) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Print(args[max])
}
func println(args ...any) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(args[i])
	}
	fmt.Println(args[max])
}

func argcError(name string, expected string, got int) (lob, error) {
	return nil, newError("Wrong number of arguments to ", name, " (expected ", expected, ", got ", got, ")")
}

func argTypeError(expected string, num int, arg lob) (lob, error) {
	return nil, newError("Argument ", num, " is not of type ", expected, ": ", arg)
}

func isFunction(obj lob) bool {
	switch obj.(type) {
	case *lcode, *lclosure, *lprimitive, *linstr:
		return true
	default:
		return false
	}
}

const defaultStackSize = 1000

func exec(thunk code, args ...lob) (lob, error) {
	if verbose {
		println("; begin execution")
	}
	code := thunk.(*lcode)
	vm := newVM(defaultStackSize)
	if len(args) != code.argc {
		return nil, newError("Wrong number of arguments")
	}
	result, err := vm.exec(code, args)
	if verbose {
		println("; end execution")
	}
	if err != nil {
		return nil, err
	}
	exports := vm.exported()
	module := code.module().(*lmodule)
	module.exported = exports[:]
	if verbose {
		//Println("; end execution")
		if err == nil && result != nil {
			println("; => ", result)
		}
	}
	return result, err
}

type lvm struct {
	stackSize int
	defs      []lob
}

func newVM(stackSize int) *lvm {
	var defs []lob
	vm := lvm{stackSize, defs}
	return &vm
}

type linstr struct {
	op int
}

// APPLY is a primitive instruction to apply a function to a list of arguments
var APPLY = &linstr{op: 0}

// CALLCC is a primitive instruction to executable (restore) a continuation
var CALLCC = &linstr{op: 1}

func (*linstr) typeSymbol() lob {
	return symFunction
}

func (s *linstr) equal(another lob) bool {
	if a, ok := another.(*linstr); ok {
		return s.op == a.op
	}
	return false
}

func (s *linstr) String() string {
	switch s.op {
	case 0:
		return "<function apply>"
	case 1:
		return "<function callcc>"
	}
	return "<instr ?>"
}

type primitive func(argv []lob, argc int) (lob, error)

type lprimitive struct {
	name string
	fun  primitive
}

var symFunction = newSymbol("function")

func (prim *lprimitive) typeSymbol() lob {
	return symFunction
}

func (prim *lprimitive) equal(another lob) bool {
	if a, ok := another.(*lprimitive); ok {
		return prim == a
	}
	return false
}

func (prim *lprimitive) String() string {
	return "<function " + prim.name + ">"
}

type lframe struct {
	previous  *lframe
	pc        int
	ops       []int
	locals    *lframe
	elements  []lob
	module    *lmodule
	constants []lob
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

func (lclosure) typeSymbol() lob {
	return symFunction
}

func (closure *lclosure) equal(another lob) bool {
	if a, ok := another.(*lclosure); ok {
		return closure == a
	}
	return false
}

func (closure lclosure) String() string {
	n := closure.code.name
	if n == "" {
		return "<function>"
	}
	return "<function " + n + ">"
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

func showStack(stack []lob, sp int) string {
	end := len(stack)
	s := "["
	for sp < end {
		s = s + fmt.Sprintf(" %v", stack[sp])
		sp++
	}
	return s + " ]"
}

func (vm *lvm) exported() []lob {
	return vm.defs
}

func buildFrame(env *lframe, pc int, ops []int, module *lmodule, fun *lclosure, argc int, stack []lob, sp int) (*lframe, error) {
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
			return nil, newError("Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
		}
		el := make([]lob, argc)
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
			return nil, newError("Wrong number of args to ", fun, " (expected at least ", expectedArgc, ", got ", argc, ")")
		}
		return nil, newError("Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
	}
	totalArgc := expectedArgc + extra
	el := make([]lob, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = toList(restElements)
	} else if keys != nil {
		bindings := stack[sp+expectedArgc : sp+argc]
		if len(bindings)%2 != 0 {
			return nil, newError("Bad keyword argument(s): ", bindings)
		}
		copy(el, stack[sp:sp+expectedArgc]) //the required ones
		for i := expectedArgc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
		for i := expectedArgc; i < argc; i += 2 {
			key := unkeyword(stack[sp+i])
			if !isSymbol(key) {
				return nil, newError("Bad keyword argument: ", key)
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
				return nil, newError("Undefined keyword argument: ", key)
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

func (vm *lvm) exec(code *lcode, args []lob) (lob, error) {
	topmod := code.mod
	stack := make([]lob, vm.stackSize)
	sp := vm.stackSize
	env := new(lframe)
	env.elements = make([]lob, len(args))
	copy(env.elements, args)
	module := code.mod
	ops := code.ops
	pc := 0
	if len(ops) == 0 {
		return nil, newError("No code to execute")
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
				println(pc, "\tglob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			val := module.global(sym)
			if val == nil {
				return nil, newError("Undefined symbol: ", sym)
			}
			sp--
			stack[sp] = val
			pc += 2
		case opcodeDefGlobal:
			if trace {
				println(pc, "\tdefglob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.defGlobal(sym, stack[sp])
			if vm.defs != nil {
				vm.defs = append(vm.defs, sym)
			}
			pc += 2
		case opcodeUndefGlobal:
			if trace {
				println(pc, "\tunglob\t", module.constants[ops[pc+1]])
			}
			sym := module.constants[ops[pc+1]]
			module.undefGlobal(sym)
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
				if topmod.checkInterrupt() {
					return nil, newError("Interrupt")
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
			case *linstr:
				if tfun == APPLY {
					if argc < 2 {
						return argcError("apply", "2+", argc)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						return argTypeError("list", argc, args)
					}
					arglist := args.(*llist)
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
					return nil, newError("unsupported instruction", tfun)
				}
			case *lsymbol:
				if isKeyword(tfun) {
					if argc != 1 {
						return argcError(tfun.Name, "1", argc)
					}
					v, err := get(stack[sp], fun)
					if err != nil {
						return nil, err
					}
					stack[sp] = v
					pc = savedPc
				} else {
					return nil, newError("Not a function: ", tfun)
				}
			default:
				return nil, newError("Not a function: ", tfun)
			}
		case opcodeTailCall:
			if topmod.checkInterrupt() {
				return nil, newError("Interrupt")
			}
			if trace {
				println(pc, "\ttcall\t", ops[pc+1], "\tstack: ", showStack(stack, sp))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
		opcodeTailCallAgain:
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
						println("------------------ END EXECUTION of ", module)
						println(" -> sp:", sp)
					}
					return stack[sp], nil
				}
			case *lclosure:
				f, err := buildFrame(env.previous, env.pc, env.ops, env.module, tfun, argc, stack, sp)
				if err != nil {
					return nil, err
				}
				sp += argc
				ops = tfun.code.ops
				module = tfun.code.mod
				pc = 0
				env = f
			case *linstr:
				if tfun == APPLY {
					if argc < 2 {
						return argcError("apply", "2+", argc)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						return argTypeError("list", argc, args)
					}
					arglist := args.(*llist)
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
					return nil, newError("unsupported instruction", tfun)
				}
			case *lsymbol:
				if isKeyword(tfun) {
					if argc != 1 {
						return argcError(tfun.Name, "1", argc)
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
							println(" -> sp:", sp)
						}
						return stack[sp], nil
					}
				} else {
					return nil, newError("Not a function: ", tfun)
				}
			default:
				return nil, newError("Not a function:", tfun)
			}
		case opcodeReturn:
			if topmod.checkInterrupt() {
				return nil, newError("Interrupt")
			}
			if trace {
				println(pc, "\tret")
			}
			if env.previous == nil {
				if trace {
					println("------------------ END EXECUTION of ", module)
					println(" -> sp:", sp)
					println(" = ", stack[sp])
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
			stack[sp] = &lclosure{module.constants[ops[pc+1]].(*lcode), env}
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
		case opcodeVector:
			if trace {
				println(pc, "\tvec\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v := toVector(stack[sp:], vlen)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		case opcodeMap:
			if trace {
				println(pc, "\tmap\t", ops[pc+1])
			}
			vlen := ops[pc+1]
			v, _ := toMap(stack[sp:], vlen)
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
			if stack[sp] == Nil {
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
			return nil, newError("Bad instruction: ", strconv.Itoa(ops[pc]))
		}
	}
}
