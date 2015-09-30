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
	"time"
)

var trace bool

func setTrace(b bool) {
	trace = b
}

func str(o interface{}) string {
	if lob, ok := o.(*LOB); ok {
		return lob.String()
	}
	return fmt.Sprintf("%v", o)
}

func print(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(str(args[i]))
	}
	fmt.Print(str(args[max]))
}

func println(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(str(args[i]))
	}
	fmt.Println(str(args[max]))
}

// Continuation - a substructure of LFunction for continuations
type Continuation struct {
	ops   []int
	stack []*LOB
	pc    int
}

// LFunction - all callable function subtypes are represented here
type LFunction struct {
	code         *LCode // closure
	frame        *Frame // closure
	primitive    *Primitive
	continuation *Continuation
	instruction  int
}

func newClosure(code *LCode, frame *Frame) *LOB {
	clo := newLOB(typeFunction)
	fun := new(LFunction)
	fun.code = code
	fun.frame = frame
	clo.function = fun
	return clo
}

func newContinuation(frame *Frame, ops []int, pc int, stack []*LOB) *LOB {
	lob := newLOB(typeFunction)
	cont := new(Continuation)
	cont.ops = ops
	cont.stack = make([]*LOB, len(stack))
	copy(cont.stack, stack)
	cont.pc = pc
	fun := new(LFunction)
	fun.frame = frame
	fun.continuation = cont
	lob.function = fun
	return lob
}

const defaultStackSize = 1000

var inExec = false
var conses = 0

func execCompileTime(code *LCode, args ...*LOB) (*LOB, error) {
	prev := verbose
	verbose = false
	res, err := exec(code, args...)
	verbose = prev
	return res, err
}

func exec(code *LCode, args ...*LOB) (*LOB, error) {
	if verbose {
		inExec = true
		conses = 0
	}
	vm := newVM(defaultStackSize)
	if len(args) != code.argc {
		return nil, Error(ArgumentErrorKey, "Wrong number of arguments")
	}
	startTime := time.Now()
	result, err := vm.exec(code, args)
	dur := time.Since(startTime)
	if verbose {
		inExec = false
		println("; executed in ", dur)
		println("; allocated list cells: ", conses)
	}
	if err != nil {
		return nil, err
	}
	if verbose {
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

const instructionNone = 0
const instructionApply = 1
const instructionCallCC = 2

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = &LOB{variant: typeFunction, function: &LFunction{instruction: instructionApply}}

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = &LOB{variant: typeFunction, function: &LFunction{instruction: instructionCallCC}}

func (f *LFunction) signature() string {
	switch f.instruction {
	case instructionNone:
		if f.primitive != nil { //primitive
			return f.primitive.signature
		}
		if f.code != nil { // closure
			return f.code.signature()
		}
	case instructionApply:
		return "(<any>*) <list>"
	case instructionCallCC:
		return "(<function>) <any>"
	default:
		panic("Bad function")
	}
	return ""
}

func (f LFunction) String() string {
	if f.instruction == instructionNone {
		if f.primitive != nil { //primitive
			return "#[function " + f.primitive.name + "]"
		}
		if f.code != nil { // closure
			n := f.code.name
			if n == "" {
				return fmt.Sprintf("#[function]")
			}
			return fmt.Sprintf("#[function %s]", n)
		}
		if f.continuation != nil { //continuation
			return "#[continuation]"
		}
		panic("Bad function")
	}
	if f.instruction == instructionApply {
		return "#[function apply]"
	}
	if f.instruction == instructionCallCC {
		return "#[function callcc]"
	}
	panic("Bad function")
}

// PrimCallable is the native go function signature for all Ell primitive functions
type PrimCallable func(argv []*LOB) (*LOB, error)

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
	locals    *Frame
	previous  *Frame
	code      *LCode
	ops       []int
	elements  []*LOB
	firstfive [5]*LOB
	pc        int
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
			return nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
		}
		if argc <= 5 {
			f.elements = f.firstfive[:]
		} else {
			f.elements = make([]*LOB, argc)
		}
		copy(f.elements, stack[sp:sp+argc])
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
			return nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected at least ", expectedArgc, ", got ", argc, ")")
		}
		return nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
	}
	totalArgc := expectedArgc + extra
	if totalArgc > 4 {
		panic("> 4 args")
	}
	el := make([]*LOB, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = listFromValues(restElements)
	} else if keys != nil {
		bindings := stack[sp+expectedArgc : sp+argc]
		if len(bindings)%2 != 0 {
			return nil, Error(ArgumentErrorKey, "Bad keyword argument(s): ", bindings)
		}
		copy(el, stack[sp:sp+expectedArgc]) //the required ones
		for i := expectedArgc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
		for i := expectedArgc; i < argc; i += 2 {
			key, err := toSymbol(stack[sp+i])
			if err != nil {
				return nil, Error(ArgumentErrorKey, "Bad keyword argument: ", stack[sp+1])
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
				return nil, Error(ArgumentErrorKey, "Undefined keyword argument: ", key)
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
	if e, ok := err.(*LOB); ok {
		if env.code != nil {
			if env.code.name != "throw" {
				e.text = env.code.name
			} else {
				e.text = env.previous.code.name
			}
		}
	}
	return err
}

func (vm *VM) funcall(fun *LOB, argc int, ops []int, savedPc int, stack []*LOB, sp int, env *Frame) ([]int, int, int, *Frame, error) {
opcodeCallAgain:
	if fun.variant == typeFunction {
		lfun := fun.function
		if lfun.code != nil {
			if checkInterrupt() {
				return nil, 0, 0, nil, addContext(env, Error(InterruptKey))
			}
			if lfun.code.defaults == nil {
				f := new(Frame)
				f.previous = env
				f.pc = savedPc
				f.ops = ops
				f.locals = lfun.frame
				f.code = lfun.code
				expectedArgc := lfun.code.argc
				if argc != expectedArgc {
					return nil, 0, 0, nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				if argc <= 5 {
					f.elements = f.firstfive[:argc]
				} else {
					f.elements = make([]*LOB, argc)
				}
				endSp := sp + argc
				copy(f.elements, stack[sp:endSp])
				return lfun.code.ops, 0, endSp, f, nil
			}
			f, err := buildFrame(env, savedPc, ops, lfun, argc, stack, sp)
			if err != nil {
				return nil, 0, 0, nil, addContext(env, err)
			}
			sp += argc
			env = f
			ops = lfun.code.ops
			return ops, 0, sp, env, err
		}
		if lfun.primitive != nil {
			val, err := lfun.primitive.fun(stack[sp : sp+argc])
			if err != nil {
				//to do: fix to throw an Ell continuation-based error
				return nil, 0, 0, nil, addContext(env, err)
			}
			sp = sp + argc - 1
			stack[sp] = val
			return ops, savedPc, sp, env, err
		}
		if fun == Apply {
			if argc < 2 {
				err := Error(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
				return nil, 0, 0, nil, addContext(env, err)
			}
			fun = stack[sp]
			args := stack[sp+argc-1]
			if !isList(args) {
				err := Error(ArgumentErrorKey, "apply expected a <list> as its final argument")
				return nil, 0, 0, nil, addContext(env, err)
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
		}
		if fun == CallCC {
			if argc != 1 {
				err := Error(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
				return nil, 0, 0, nil, addContext(env, err)
			}
			fun = stack[sp]
			stack[sp] = newContinuation(env, ops, savedPc, stack[sp+1:])
			goto opcodeCallAgain
		}
		if lfun.continuation != nil {
			println("funcall cont")
			if argc != 1 {
				return nil, 0, 0, nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
			}
			arg := stack[sp]
			sp = len(stack) - len(lfun.continuation.stack)
			segment := stack[sp:]
			copy(segment, lfun.continuation.stack)
			sp--
			stack[sp] = arg
			return lfun.continuation.ops, lfun.continuation.pc, sp, lfun.frame, nil
		}
		panic("unsupported instruction")
	}
	if fun.variant == typeKeyword {
		if argc != 1 {
			err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
			return nil, 0, 0, nil, addContext(env, err)
		}
		v, err := get(stack[sp], fun)
		if err != nil {
			return nil, 0, 0, nil, addContext(env, err)
		}
		stack[sp] = v
		return ops, savedPc, sp, env, err
	}
	return nil, 0, 0, nil, addContext(env, Error(ArgumentErrorKey, "Not a function: ", fun))
}

func (vm *VM) tailcall(fun *LOB, argc int, ops []int, stack []*LOB, sp int, env *Frame) ([]int, int, int, *Frame, error) {
opcodeTailCallAgain:
	if fun.variant == typeFunction {
		lfun := fun.function
		if lfun.code != nil {
			if lfun.code.defaults == nil && lfun.code == env.code { //self-tail-call - we can reuse the frame.
				expectedArgc := lfun.code.argc
				if argc != expectedArgc {
					return nil, 0, 0, nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				endSp := sp + argc
				copy(env.elements, stack[sp:endSp])
				return lfun.code.ops, 0, endSp, env, nil
			}
			f, err := buildFrame(env.previous, env.pc, env.ops, lfun, argc, stack, sp)
			if err != nil {
				return nil, 0, 0, nil, addContext(env, err)
			}
			sp += argc
			return lfun.code.ops, 0, sp, f, nil
		}
		if lfun.primitive != nil {
			val, err := lfun.primitive.fun(stack[sp : sp+argc])
			if err != nil {
				return nil, 0, 0, nil, addContext(env, err)
			}
			sp = sp + argc - 1
			stack[sp] = val
			return env.ops, env.pc, sp, env.previous, nil
		}
		if fun == Apply {
			if argc < 2 {
				err := Error(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
				return nil, 0, 0, nil, addContext(env, err)
			}
			fun = stack[sp]
			args := stack[sp+argc-1]
			if !isList(args) {
				err := Error(ArgumentErrorKey, "apply expected its last argument to be a <list>")
				return nil, 0, 0, nil, addContext(env, err)
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
		}
		if fun == CallCC {
			if argc != 1 {
				err := Error(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
				return nil, 0, 0, nil, addContext(env, err)
			}
			fun = stack[sp]
			stack[sp] = newContinuation(env.previous, env.ops, env.pc, stack[sp:])
			goto opcodeTailCallAgain
		}
		if lfun.continuation != nil {
			if argc != 1 {
				return nil, 0, 0, nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
			}
			arg := stack[sp]
			sp = len(stack) - len(lfun.continuation.stack)
			segment := stack[sp:]
			copy(segment, lfun.continuation.stack)
			sp--
			stack[sp] = arg
			return lfun.continuation.ops, lfun.continuation.pc, sp, lfun.frame, nil
		}
		panic("unsupported instruction")
	}
	if fun.variant == typeKeyword {
		if argc != 1 {
			err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
			return nil, 0, 0, nil, addContext(env, err)
		}
		v, err := get(stack[sp], fun)
		if err != nil {
			return nil, 0, 0, nil, addContext(env, err)
		}
		stack[sp] = v
		return env.ops, env.pc, sp, env.previous, nil
	}
	return nil, 0, 0, nil, addContext(env, Error(ArgumentErrorKey, "Not a function:", fun))
}

func (vm *VM) exec(code *LCode, args []*LOB) (*LOB, error) {
	if verbose || trace {
		return vm.instrumentedExec(code, args)
	}
	stack := make([]*LOB, vm.stackSize)
	sp := vm.stackSize
	env := new(Frame)
	env.elements = make([]*LOB, len(args))
	copy(env.elements, args)
	ops := code.ops
	pc := 0
	var err error
	for {
		op := ops[pc]
		if op == opcodeCall { // CALL
			argc := ops[pc+1]
			fun := stack[sp]
			if fun.variant == typeFunction {
				if fun.function.primitive != nil {
					nextSp := sp + argc
					val, err := fun.function.primitive.fun(stack[sp+1 : nextSp+1])
					if err != nil {
						return nil, addContext(env, err)
					}
					stack[nextSp] = val
					sp = nextSp
					pc += 2
				} else {
					ops, pc, sp, env, err = vm.funcall(fun, argc, ops, pc+2, stack, sp+1, env)
					if err != nil {
						return nil, err
					}
				}
			} else if fun.variant == typeKeyword {
				if argc != 1 {
					err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
					return nil, addContext(env, err)
				}
				sp++
				v, err := get(stack[sp], fun)
				if err != nil {
					return nil, addContext(env, err)
				}
				stack[sp] = v
				pc += 2
			} else {
				return nil, addContext(env, Error(ArgumentErrorKey, "Not callable: ", fun))
			}
		} else if op == opcodeGlobal { //GLOBAL
			sym := constants[ops[pc+1]]
			if sym.car == nil {
				return nil, addContext(env, Error(ErrorKey, "Undefined symbol: ", sym))
			}
			sp--
			stack[sp] = sym.car
			pc += 2
		} else if op == opcodeLocal {
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
		} else if op == opcodeJumpFalse {
			b := stack[sp]
			sp++
			if b == False {
				pc += ops[pc+1]
			} else {
				pc += 2
			}
		} else if op == opcodePop {
			sp++
			pc++
		} else if op == opcodeTailCall {
			if checkInterrupt() {
				return nil, addContext(env, Error(InterruptKey))
			}
			fun := stack[sp]
			argc := ops[pc+1]
			if fun.variant == fun.variant {
				if fun.function.primitive != nil {
					nextSp := sp + argc
					val, err := fun.function.primitive.fun(stack[sp+1 : nextSp+1])
					if err != nil {
						return nil, addContext(env, err)
					}
					stack[nextSp] = val
					sp = nextSp
					ops = env.ops
					pc = env.pc
					env = env.previous
					if env == nil {
						return stack[sp], nil
					}
				} else {
					ops, pc, sp, env, err = vm.tailcall(fun, argc, ops, stack, sp+1, env)
					if env == nil {
						return stack[sp], nil
					}
					if err != nil {
						return nil, err
					}
				}
			} else if fun.variant == typeKeyword {
				if argc != 1 {
					err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
					return nil, addContext(env, err)
				}
				sp++
				v, err := get(stack[sp], fun)
				if err != nil {
					return nil, addContext(env, err)
				}
				stack[sp] = v
				ops = env.ops
				pc = env.pc
				env = env.previous
				if env == nil {
					return stack[sp], nil
				}
			} else {
				return nil, addContext(env, Error(ArgumentErrorKey, "Not callable: ", fun))
			}
		} else if op == opcodeLiteral {
			sp--
			stack[sp] = constants[ops[pc+1]]
			pc += 2
		} else if op == opcodeSetLocal {
			tmpEnv := env
			i := ops[pc+1]
			for i > 0 {
				tmpEnv = tmpEnv.locals
				i--
			}
			j := ops[pc+2]
			tmpEnv.elements[j] = stack[sp]
			pc += 3
		} else if op == opcodeClosure {
			sp--
			stack[sp] = newClosure(constants[ops[pc+1]].code, env)
			pc = pc + 2
		} else if op == opcodeReturn {
			if checkInterrupt() {
				return nil, addContext(env, Error(InterruptKey))
			}
			if env.previous == nil {
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			env = env.previous
		} else if op == opcodeJump {
			pc += ops[pc+1]
		} else if op == opcodeDefGlobal {
			sym := constants[ops[pc+1]]
			defGlobal(sym, stack[sp])
			pc += 2
		} else if op == opcodeUndefGlobal {
			sym := constants[ops[pc+1]]
			undefGlobal(sym)
			pc += 2
		} else if op == opcodeDefMacro {
			sym := constants[ops[pc+1]]
			defMacro(sym, stack[sp])
			stack[sp] = sym
			pc += 2
		} else if op == opcodeUse {
			sym := constants[ops[pc+1]]
			err := use(sym)
			if err != nil {
				return nil, addContext(env, err)
			}
			sp--
			stack[sp] = sym
			pc += 2
		} else if op == opcodeVector {
			vlen := ops[pc+1]
			v := vector(stack[sp : sp+vlen]...)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else if op == opcodeStruct {
			vlen := ops[pc+1]
			v, _ := newStruct(stack[sp : sp+vlen])
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else {
			panic("Bad instruction")
		}
	}
}

const stackColumn = 40

func showInstruction(pc int, op int, args string, stack []*LOB, sp int) {
	var body string
	body = leftJustified(fmt.Sprintf("%d ", pc), 8) + leftJustified(opsyms[op].text, 10) + args
	println(leftJustified(body, stackColumn), showStack(stack, sp))
}

func leftJustified(s string, width int) string {
	padsize := width - len(s)
	for i := 0; i < padsize; i++ {
		s += " "
	}
	return s
}

func truncatedObjectString(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit] // ((defn foo (x) (if (not (stri
		firstN := s[:limit-3]
		for i := limit - 1; i >= 0; i-- {
			if isWhitespace(s[i]) {
				s = s[:i]
				break
			}
		}
		if s == "" {
			s = firstN + "..."
		} else {
			openParens := 0
			for _, c := range s {
				switch c {
				case '(':
					openParens++
				case ')':
					openParens--
				}
			}
			if openParens > 0 {
				s += " ..."
				for i := 0; i < openParens; i++ {
					s += ")"
				}
			} else {
				s += "..."
			}
		}
	}
	return s
}
func showStack(stack []*LOB, sp int) string {
	end := len(stack)
	s := "["
	limit := 5
	tail := ""
	if end-sp > limit {
		end = sp + limit
		tail = " ... "
	}
	for sp < end {
		tmp := fmt.Sprintf(" %v", write(stack[sp]))
		s = s + truncatedObjectString(tmp, 30)
		sp++
	}
	return s + tail + " ]"
}

func (vm *VM) instrumentedExec(code *LCode, args []*LOB) (*LOB, error) {
	stack := make([]*LOB, vm.stackSize)
	sp := vm.stackSize
	env := new(Frame)
	env.elements = make([]*LOB, len(args))
	copy(env.elements, args)
	ops := code.ops
	pc := 0
	for {
		op := ops[pc]
		if op == opcodeCall {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
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
					val, err := lfun.primitive.fun(stack[sp : sp+argc])
					if err != nil {
						//to do: fix to throw an Ell continuation-based error
						return nil, addContext(env, err)
					}
					sp = sp + argc - 1
					stack[sp] = val
					pc = savedPc
				} else if lfun.code != nil {
					if checkInterrupt() {
						return nil, addContext(env, Error(InterruptKey))
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
						err := Error(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
						return nil, addContext(env, err)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						err := Error(ArgumentErrorKey, "apply expected its last argument to be a list")
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
				} else if fun == CallCC {
					if argc != 1 {
						err := Error(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
						return nil, addContext(env, err)
					}
					fun = stack[sp]
					stack[sp] = newContinuation(env, ops, savedPc, stack[sp+1:])
					goto opcodeCallAgain
				} else if lfun.continuation != nil {
					if argc != 1 {
						return nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
					}
					arg := stack[sp]
					sp = len(stack) - len(lfun.continuation.stack)
					segment := stack[sp:]
					copy(segment, lfun.continuation.stack)
					env = lfun.frame
					sp--
					stack[sp] = arg
					pc = lfun.continuation.pc
					ops = lfun.continuation.ops
				} else {
					panic("bad instruction")
				}
			case typeKeyword:
				if argc != 1 {
					err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
					return nil, addContext(env, err)
				}
				v, err := get(stack[sp], fun)
				if err != nil {
					return nil, addContext(env, err)
				}
				stack[sp] = v
				pc = savedPc
			default:
				return nil, addContext(env, Error(ArgumentErrorKey, "Not a function: ", fun))
			}
		} else if op == opcodeGlobal {
			sym := constants[ops[pc+1]]
			if trace {
				showInstruction(pc, op, sym.String(), stack, sp)
			}
			if sym.car == nil {
				return nil, addContext(env, Error(ErrorKey, "Undefined symbol: ", sym))
			}
			sp--
			stack[sp] = sym.car
			pc += 2
		} else if op == opcodeLocal {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d, %d", ops[pc+1], ops[pc+2]), stack, sp)
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
		} else if op == opcodeJumpFalse {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", pc+ops[pc+1]), stack, sp)
			}
			b := stack[sp]
			sp++
			if b == False {
				pc += ops[pc+1]
			} else {
				pc += 2
			}
		} else if op == opcodePop {
			if trace {
				showInstruction(pc, op, "", stack, sp)
			}
			sp++
			pc++
		} else if op == opcodeTailCall {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			if checkInterrupt() {
				return nil, addContext(env, Error(InterruptKey))
			}
			fun := stack[sp]
			sp++
			argc := ops[pc+1]
		opcodeTailCallAgain:
			switch fun.variant {
			case typeFunction:
				lfun := fun.function
				if lfun.primitive != nil {
					val, err := lfun.primitive.fun(stack[sp : sp+argc])
					if err != nil {
						return nil, addContext(env, err)
					}
					sp = sp + argc - 1
					stack[sp] = val
					pc = env.pc
					ops = env.ops
					env = env.previous
					if env == nil {
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
						err := Error(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
						return nil, addContext(env, err)
					}
					fun = stack[sp]
					args := stack[sp+argc-1]
					if !isList(args) {
						err := Error(ArgumentErrorKey, "apply expected its last argument to be a list")
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
				} else if fun == CallCC {
					if argc != 1 {
						err := Error(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
						return nil, addContext(env, err)
					}
					fun = stack[sp]
					stack[sp] = newContinuation(env.previous, env.ops, env.pc, stack[sp:])
					goto opcodeTailCallAgain
				} else if lfun.continuation != nil {
					if argc != 1 {
						return nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
					}
					arg := stack[sp]
					sp = len(stack) - len(lfun.continuation.stack)
					segment := stack[sp:]
					copy(segment, lfun.continuation.stack)
					env = lfun.frame
					sp--
					stack[sp] = arg
					pc = lfun.continuation.pc
					ops = lfun.continuation.ops
				} else {
					panic("Bad instruction")
				}
			case typeKeyword:
				if argc != 1 {
					err := Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
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
					return stack[sp], nil
				}
			default:
				return nil, addContext(env, Error(ArgumentErrorKey, "Not a function:", fun))
			}
		} else if op == opcodeLiteral {
			if trace {
				showInstruction(pc, op, write(constants[ops[pc+1]].variant), stack, sp)
			}
			sp--
			stack[sp] = constants[ops[pc+1]]
			pc += 2
		} else if op == opcodeSetLocal {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d, %d", ops[pc+1], ops[pc+2]), stack, sp)
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
		} else if op == opcodeClosure {
			if trace {
				showInstruction(pc, op, "", stack, sp)
			}
			sp--
			stack[sp] = newClosure(constants[ops[pc+1]].code, env)
			pc = pc + 2
		} else if op == opcodeReturn {
			if trace {
				showInstruction(pc, op, "", stack, sp)
			}
			if checkInterrupt() {
				return nil, addContext(env, Error(InterruptKey))
			}
			if env.previous == nil {
				return stack[sp], nil
			}
			ops = env.ops
			pc = env.pc
			env = env.previous
		} else if op == opcodeJump {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", pc+ops[pc+1]), stack, sp)
			}
			pc += ops[pc+1]
		} else if op == opcodeDefGlobal {
			sym := constants[ops[pc+1]]
			if trace {
				showInstruction(pc, op, sym.text, stack, sp)
			}
			defGlobal(sym, stack[sp])
			pc += 2
		} else if op == opcodeUndefGlobal {
			sym := constants[ops[pc+1]]
			if trace {
				showInstruction(pc, op, sym.text, stack, sp)
			}
			undefGlobal(sym)
			pc += 2
		} else if op == opcodeDefMacro {
			sym := constants[ops[pc+1]]
			if trace {
				showInstruction(pc, op, sym.text, stack, sp)
			}
			defMacro(sym, stack[sp])
			stack[sp] = sym
			pc += 2
		} else if op == opcodeUse {
			sym := constants[ops[pc+1]]
			if trace {
				showInstruction(pc, op, sym.text, stack, sp)
			}
			err := use(sym)
			if err != nil {
				return nil, addContext(env, err)
			}
			sp--
			stack[sp] = sym
			pc += 2
		} else if op == opcodeVector {
			vlen := ops[pc+1]
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			v := vector(stack[sp : sp+vlen]...)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else if op == opcodeStruct {
			vlen := ops[pc+1]
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			v, _ := newStruct(stack[sp : sp+vlen])
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else {
			panic("Bad instruction")
		}
	}
}
