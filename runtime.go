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
var optimize bool

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

// Continuation -
type Continuation struct {
	ops   []int
	stack []*LOB
	pc    int
}

func newClosure(code *Code, frame *Frame) *LOB {
	fun := newLOB(typeFunction)
	fun.code = code
	fun.frame = frame
	return fun
}

func newContinuation(frame *Frame, ops []int, pc int, stack []*LOB) *LOB {
	fun := newLOB(typeFunction)
	cont := new(Continuation)
	cont.ops = ops
	cont.stack = make([]*LOB, len(stack))
	copy(cont.stack, stack)
	cont.pc = pc
	fun.frame = frame
	fun.continuation = cont
	return fun
}

const defaultStackSize = 1000

func execCompileTime(code *Code, arg *LOB) (*LOB, error) {
	args := []*LOB{arg}
	prev := verbose
	verbose = false
	res, err := exec(code, args)
	verbose = prev
	return res, err
}

func spawn(code *Code, args []*LOB) {
	argcopy := make([]*LOB, len(args), len(args))
	copy(argcopy, args)
	go func() {
		_, err := exec(code, argcopy)
		if err != nil {
			println("; [*** error in spawned function '", code.name, "': ", err, "]")
		} else if verbose {
			println("; [spawned function '", code.name, "' exited cleanly]")
		}
	}()
}

//func exec(code *Code, args ...*LOB) (*LOB, error) {
func exec(code *Code, args []*LOB) (*LOB, error) {
	vm := newVM(defaultStackSize)
	if len(args) != code.argc {
		return nil, Error(ArgumentErrorKey, "Wrong number of arguments")
	}
	startTime := time.Now()
	result, err := vm.exec(code, args)
	dur := time.Since(startTime)
	if verbose {
		println("; executed in ", dur)
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

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = &LOB{variant: typeFunction}

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = &LOB{variant: typeFunction}

func functionSignature(f *LOB) string {
	if f.primitive != nil {
		return f.primitive.signature
	}
	if f.code != nil {
		return f.code.signature()
	}
	if f.continuation != nil {
		return "(<function>) <any>"
	}
	if f == Apply {
		return "(<any>*) <list>"
	}
	if f == CallCC {
		return "(<function>) <any>"
	}
	panic("Bad function")
}

func functionToString(f *LOB) string {
	if f.primitive != nil {
		return "#[function " + f.primitive.name + "]"
	}
	if f.code != nil {
		n := f.code.name
		if n == "" {
			return fmt.Sprintf("#[function]")
		}
		return fmt.Sprintf("#[function %s]", n)
	}
	if f.continuation != nil {
		return "#[continuation]"
	}
	if f == Apply {
		return "#[function apply]"
	}
	if f == CallCC {
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
	argc      int    // -1 means the primitive itself checks the args (legacy mode)
	result    *LOB   // if set the type of the result
	args      []*LOB // if set, the length must be for total args (both required and optional). The type (or <any>) for each
	rest      *LOB   // if set, then any number of this type can follow the normal args. Mutually incompatible with defaults/keys
	defaults  []*LOB // if set, then that many optional args beyond argc have these default values
	keys      []*LOB // if set, then it must match the size of defaults, and these are the keys
}

func functionSignatureFromTypes(result *LOB, args []*LOB, rest *LOB) string {
	sig := "("
	for i, t := range args {
		if !isType(t) {
			panic("not a type: " + t.String())
		}
		if i > 0 {
			sig += " "
		}
		sig += t.text
	}
	if rest != nil {
		if !isType(rest) {
			panic("not a type: " + rest.String())
		}
		if sig != "(" {
			sig += " "
		}
		sig += rest.text + "*"
	}
	sig += ") "
	if !isType(result) {
		panic("not a type: " + result.String())
	}
	sig += result.text
	return sig
}

func newPrimitive(name string, fun PrimCallable, result *LOB, args []*LOB, rest *LOB, defaults []*LOB, keys []*LOB) *LOB {
	//the rest type indicates arguments past the end of args will all have the given type. the length must be checked by primitive
	// -> they are all optional, then. So, (<any>+) must be expressed as (<any> <any>*)
	idx := len(primitives)
	argc := len(args)
	if defaults != nil {
		defc := len(defaults)
		if defc > argc {
			panic("more default argument values than types: " + name)
		}
		if keys != nil {
			if len(keys) != defc {
				panic("Argument keys must have same length as argument defaults")
			}
		}
		argc = argc - defc
		for i := 0; i < defc; i++ {
			t := args[argc+i]
			if t != typeAny && defaults[i].variant != t {
				panic("argument default's type (" + defaults[i].variant.text + ") doesn't match declared type (" + t.text + ")")
			}
		}
	} else {
		if keys != nil {
			panic("Cannot have argument keys without argument defaults")
		}
	}
	signature := functionSignatureFromTypes(result, args, rest)
	prim := &Primitive{name, fun, signature, idx, argc, result, args, rest, defaults, keys}
	primitives = append(primitives, prim)
	return &LOB{variant: typeFunction, primitive: prim}
}

//argc == 1: 1 or more args (depending on Defaults)
//argc == 0: zero or more args (depending on Defaults
//argc = -1: do no check, the primitive will
func newLegacyPrimitive(name string, fun PrimCallable, signature string) *LOB {
	idx := len(primitives)
	prim := &Primitive{name, fun, signature, idx, -1, nil, nil, nil, nil, nil}
	primitives = append(primitives, prim)
	return &LOB{variant: typeFunction, primitive: prim}
}

// Frame - a call frame in the VM, as well as en environment frame for lexical closures
type Frame struct {
	locals    *Frame
	previous  *Frame
	code      *Code
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

func buildFrame(env *Frame, pc int, ops []int, fun *LOB, argc int, stack []*LOB, sp int) (*Frame, error) {
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
			} else if env.previous != nil {
				if env.previous.code != nil {
					e.text = env.previous.code.name
				}
			}
		}
	}
	return err
}

func (vm *VM) keywordCall(fun *LOB, argc int, pc int, stack []*LOB, sp int) (int, int, error) {
	if argc != 1 {
		return 0, 0, Error(ArgumentErrorKey, fun.text, " expected 1 argument, got ", argc)
	}
	v, err := get(stack[sp], fun)
	if err != nil {
		return 0, 0, err
	}
	stack[sp] = v
	return pc, sp, nil
}

func argcError(name string, min int, max, provided int) error {
	s := "1 argument"
	if min == max {
		if min != 1 {
			s = fmt.Sprintf("%d arguments", min)
		}
	} else if max < 0 {
		s = fmt.Sprintf("%d or more arguments", min)
	} else {
		s = fmt.Sprintf("%d to %d arguments", min, max)
	}
	return Error(ArgumentErrorKey, fmt.Sprintf("%s expected %s, got %d", name, s, provided))
}

func (vm *VM) callPrimitive(prim *Primitive, argv []*LOB) (*LOB, error) {
	if prim.argc < 0 { //let the primitive itself figure it out
		return prim.fun(argv)
	}
	if prim.defaults != nil {
		return vm.callPrimitiveWithDefaults(prim, argv)
	}
	argc := len(argv)
	if argc != prim.argc {
		return nil, argcError(prim.name, prim.argc, prim.argc, argc)
	}
	for i, arg := range argv {
		t := prim.args[i]
		if t != typeAny && arg.variant != t {
			return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].text, i+1, argv[i].variant.text))
		}
	}
	return prim.fun(argv)
}

func (vm *VM) callPrimitiveWithDefaults(prim *Primitive, argv []*LOB) (*LOB, error) {
	provided := len(argv)
	minargc := prim.argc
	if len(prim.defaults) == 0 {
		rest := prim.rest
		if provided < minargc {
			return nil, argcError(prim.name, minargc, -1, provided)
		}
		for i := 0; i < minargc; i++ {
			t := prim.args[i]
			arg := argv[i]
			if t != typeAny && arg.variant != t {
				return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].text, i+1, argv[i].variant.text))
			}
		}
		if rest != typeAny {
			for i := minargc; i < provided; i++ {
				arg := argv[i]
				if arg.variant != rest {
					return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, rest.text, i+1, argv[i].variant.text))
				}
			}
		}
		return prim.fun(argv)
	}
	maxargc := len(prim.args)
	if provided < minargc {
		return nil, argcError(prim.name, minargc, maxargc, provided)
	}
	newargs := make([]*LOB, maxargc)
	if prim.keys != nil {
		j := 0
		copy(newargs, argv[:minargc])
		for i := minargc; i < maxargc; i++ {
			newargs[i] = prim.defaults[j]
			j++
		}
		j = minargc //the first key arg
		ndefaults := len(prim.defaults)
		for j < provided {
			k := argv[j]
			j++
			if j == provided {
				return nil, Error(ArgumentErrorKey, "mismatched keyword/value pair in argument list")
			}
			if k.variant != typeKeyword {
				return nil, Error(ArgumentErrorKey, "expected keyword, got a "+k.variant.text)
			}
			gotit := false
			for i := 0; i < ndefaults; i++ {
				if prim.keys[i] == k {
					gotit = true
					newargs[i+minargc] = argv[j]
					j++
					break
				}
			}
			if !gotit {
				return nil, Error(ArgumentErrorKey, prim.name, " accepts ", prim.keys, " as keyword arg(s), not ", k)
			}
		}
		argv = newargs
	} else {
		if provided > maxargc {
			return nil, argcError(prim.name, minargc, maxargc, provided)
		}
		copy(newargs, argv)
		j := 0
		for i := provided; i < maxargc; i++ {
			newargs[i] = prim.defaults[j]
			j++
		}
		argv = newargs
	}
	for i, arg := range argv {
		t := prim.args[i]
		if t != typeAny && arg.variant != t {
			return nil, Error(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].text, i+1, argv[i].variant.text))
		}
	}
	return prim.fun(argv)
}

func (vm *VM) funcall(fun *LOB, argc int, ops []int, savedPc int, stack []*LOB, sp int, env *Frame) ([]int, int, int, *Frame, error) {
opcodeCallAgain:
	if fun.variant == typeFunction {
		if fun.code != nil {
			if checkInterrupt() {
				return nil, 0, 0, nil, addContext(env, Error(InterruptKey))
			}
			if fun.code.defaults == nil {
				f := new(Frame)
				f.previous = env
				f.pc = savedPc
				f.ops = ops
				f.locals = fun.frame
				f.code = fun.code
				expectedArgc := fun.code.argc
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
				return fun.code.ops, 0, endSp, f, nil
			}
			f, err := buildFrame(env, savedPc, ops, fun, argc, stack, sp)
			if err != nil {
				return nil, 0, 0, nil, addContext(env, err)
			}
			sp += argc
			env = f
			ops = fun.code.ops
			return ops, 0, sp, env, err
		}
		if fun.primitive != nil {
			val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
			if err != nil {
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
			argc = listLength(arglist)
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
		if fun.continuation != nil {
			if argc != 1 {
				return nil, 0, 0, nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
			}
			arg := stack[sp]
			sp = len(stack) - len(fun.continuation.stack)
			segment := stack[sp:]
			copy(segment, fun.continuation.stack)
			sp--
			stack[sp] = arg
			return fun.continuation.ops, fun.continuation.pc, sp, fun.frame, nil
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
		if fun.code != nil {
			if fun.code.defaults == nil && fun.code == env.code { //self-tail-call - we can reuse the frame.
				expectedArgc := fun.code.argc
				if argc != expectedArgc {
					return nil, 0, 0, nil, Error(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				endSp := sp + argc
				copy(env.elements, stack[sp:endSp])
				return fun.code.ops, 0, endSp, env, nil
			}
			f, err := buildFrame(env.previous, env.pc, env.ops, fun, argc, stack, sp)
			if err != nil {
				return nil, 0, 0, nil, addContext(env, err)
			}
			sp += argc
			return fun.code.ops, 0, sp, f, nil
		}
		if fun.primitive != nil {
			val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
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
			argc = listLength(arglist)
			i := 0
			sp -= argc
			for arglist != EmptyList {
				stack[sp+i] = arglist.car
				i++
				arglist = arglist.cdr
			}
			goto opcodeTailCallAgain
		}
		if fun.continuation != nil {
			if argc != 1 {
				return nil, 0, 0, nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
			}
			arg := stack[sp]
			sp = len(stack) - len(fun.continuation.stack)
			segment := stack[sp:]
			copy(segment, fun.continuation.stack)
			sp--
			stack[sp] = arg
			return fun.continuation.ops, fun.continuation.pc, sp, fun.frame, nil
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
		panic("Bad function")
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

func (vm *VM) keywordTailcall(fun *LOB, argc int, ops []int, stack []*LOB, sp int, env *Frame) ([]int, int, int, *Frame, error) {
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

func (vm *VM) exec(code *Code, args []*LOB) (*LOB, error) {
	if !optimize || verbose || trace {
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
			if fun.primitive != nil {
				nextSp := sp + argc
				val, err := fun.primitive.fun(stack[sp+1 : nextSp+1])
				if err != nil {
					return nil, addContext(env, err)
				}
				stack[nextSp] = val
				sp = nextSp
				pc += 2
			} else if fun.variant == typeFunction {
				ops, pc, sp, env, err = vm.funcall(fun, argc, ops, pc+2, stack, sp+1, env)
				if err != nil {
					return nil, err
				}
			} else if fun.variant == typeKeyword {
				pc, sp, err = vm.keywordCall(fun, argc, pc+2, stack, sp+1)
				if err != nil {
					return nil, addContext(env, err)
				}
			} else {
				return nil, addContext(env, Error(ArgumentErrorKey, "Not callable: ", fun))
			}
		} else if op == opcodeGlobal { //GLOBAL
			sym := constants[ops[pc+1]]
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
			fun := stack[sp]
			argc := ops[pc+1]
			if fun.primitive != nil {
				nextSp := sp + argc
				val, err := vm.callPrimitive(fun.primitive, stack[sp+1:nextSp+1])
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
			} else if fun.variant == fun.variant {
				ops, pc, sp, env, err = vm.tailcall(fun, argc, ops, stack, sp+1, env)
				if env == nil {
					return stack[sp], nil
				}
				if err != nil {
					return nil, err
				}
			} else if fun.variant == typeKeyword {
				ops, pc, sp, env, err = vm.keywordTailcall(fun, argc, ops, stack, sp+1, env)
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

func (vm *VM) instrumentedExec(code *Code, args []*LOB) (*LOB, error) {
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
				if fun.primitive != nil {
					val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
					if err != nil {
						//to do: fix to throw an Ell continuation-based error
						return nil, addContext(env, err)
					}
					sp = sp + argc - 1
					stack[sp] = val
					pc = savedPc
				} else if fun.code != nil {
					if checkInterrupt() {
						return nil, addContext(env, Error(InterruptKey))
					}

					f, err := buildFrame(env, savedPc, ops, fun, argc, stack, sp)
					if err != nil {
						return nil, addContext(env, err)
					}
					sp += argc
					env = f
					ops = fun.code.ops
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
					argc = listLength(arglist)
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
				} else if fun.continuation != nil {
					if argc != 1 {
						return nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
					}
					arg := stack[sp]
					sp = len(stack) - len(fun.continuation.stack)
					segment := stack[sp:]
					copy(segment, fun.continuation.stack)
					env = fun.frame
					sp--
					stack[sp] = arg
					pc = fun.continuation.pc
					ops = fun.continuation.ops
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
			if tmpEnv == nil {
				return nil, addContext(env, Error(ErrorKey, "Closed over environment not available in this context"))
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
				if fun.primitive != nil {
					val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
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
				} else if fun.code != nil {
					f, err := buildFrame(env.previous, env.pc, env.ops, fun, argc, stack, sp)
					if err != nil {
						return nil, addContext(env, err)
					}
					sp += argc
					code = fun.code
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
					argc = listLength(arglist)
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
				} else if fun.continuation != nil {
					if argc != 1 {
						return nil, addContext(env, Error(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc))
					}
					arg := stack[sp]
					sp = len(stack) - len(fun.continuation.stack)
					segment := stack[sp:]
					copy(segment, fun.continuation.stack)
					env = fun.frame
					sp--
					stack[sp] = arg
					pc = fun.continuation.pc
					ops = fun.continuation.ops
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
