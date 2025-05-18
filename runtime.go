/*
Copyright 2015 Lee Boynton

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

package ell

import (
	"bytes"
	"fmt"
	"os"
	"time"

	. "github.com/boynton/ell/data"
)

var trace bool
var optimize bool

var interrupted = false
var interrupts chan os.Signal
var InterruptKey = Intern("interrupt:")

func checkInterrupt() bool {
	if interrupts != nil {
		select {
		case msg := <-interrupts:
			if msg != nil {
				interrupted = true
				return true
			}
		default:
			return false
		}
	}
	return false
}

func str(o interface{}) string {
	if lob, ok := o.(Value); ok {
		return lob.String()
	}
	return fmt.Sprintf("%v", o)
}

func Print(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(str(args[i]))
	}
	fmt.Print(str(args[max]))
}

func Println(args ...interface{}) {
	max := len(args) - 1
	for i := 0; i < max; i++ {
		fmt.Print(str(args[i]))
	}
	fmt.Println(str(args[max]))
}

func Fatal(args ...interface{}) {
	Println(args...)
	Cleanup()
	exit(1)
}

// Continuation -
type Continuation struct {
	ops   []int
	stack []Value
	pc    int
}

func Closure(code *Code, frame *Frame) *Function {
	return &Function{
		code:  code,
		frame: frame,
	}
}

func NewContinuation(frame *Frame, ops []int, pc int, stack []Value) *Function {
	cont := new(Continuation)
	cont.ops = ops
	cont.stack = make([]Value, len(stack))
	copy(cont.stack, stack)
	cont.pc = pc
	return &Function{
		frame:        frame,
		continuation: cont,
	}
}

const defaultStackSize = 1000

// VM - the Ell VM
type vm struct {
	stackSize int
}

func VM(stackSize int) *vm {
	return &vm{stackSize}
}

var FunctionType Value = Intern("<function>")

type Function struct {
	name         string
	code         *Code
	frame        *Frame
	primitive    *Primitive
	continuation *Continuation
}

func (f *Function) Type() Value {
	return FunctionType
}

func IsFunction(v Value) bool {
	return v.Type() == FunctionType
}

func (f *Function) Equals(another Value) bool {
	return false
}

func (f *Function) String() string {
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
	if f == Spawn {
		return "#[function spawn]"
	}
	panic("Bad function")
}

// Apply is a primitive instruction to apply a function to a list of arguments
var Apply = &Function{}

// CallCC is a primitive instruction to executable (restore) a continuation
var CallCC = &Function{}

// Apply is a primitive instruction to apply a function to a list of arguments
var Spawn = &Function{}

func functionSignature(f *Function) string {
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
	if f == Spawn {
		return "(<function> <any>*) <null>"
	}
	panic("Bad function")
}

// PrimitiveFunction is the native go function signature for all Ell primitive functions
type PrimitiveFunction func(argv []Value) (Value, error)

// Primitive - a primitive function, written in Go, callable by VM
type Primitive struct { // <function>
	name      string
	fun       PrimitiveFunction
	signature string
	//	idx       int
	argc     int     // -1 means the primitive itself checks the args (legacy mode)
	result   Value   // if set the type of the result
	args     []Value // if set, the length must be for total args (both required and optional). The type (or <any>) for each
	rest     Value   // if set, then any number of this type can follow the normal args. Mutually incompatible with defaults/keys
	defaults []Value // if set, then that many optional args beyond argc have these default values
	keys     []Value // if set, then it must match the size of defaults, and these are the keys
}

func functionSignatureFromTypes(result Value, args []Value, rest Value) string {
	sig := "("
	for i, t := range args {
		if !IsType(t) {
			panic("not a type: " + t.String())
		}
		if i > 0 {
			sig += " "
		}
		sig += t.String()
	}
	if rest != nil {
		if !IsType(rest) {
			panic("not a type: " + rest.String())
		}
		if sig != "(" {
			sig += " "
		}
		sig += rest.String() + "*"
	}
	sig += ") "
	if !IsType(result) {
		panic("not a type: " + result.String())
	}
	sig += result.String()
	return sig
}

func NewPrimitive(name string, fun PrimitiveFunction, result Value, args []Value, rest Value, defaults []Value, keys []Value) *Function {
	//the rest type indicates arguments past the end of args will all have the given type. the length must be checked by primitive
	// -> they are all optional, then. So, (<any>+) must be expressed as (<any> <any>*)
	//	idx := len(primitives)
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
			if t != AnyType && defaults[i].Type() != t {
				panic("argument default's type (" + TypeNameOf(defaults[i]) + ") doesn't match declared type (" + t.String() + ")")
			}
		}
	} else {
		if keys != nil {
			panic("Cannot have argument keys without argument defaults")
		}
	}
	signature := functionSignatureFromTypes(result, args, rest)
	prim := &Primitive{name, fun, signature, argc, result, args, rest, defaults, keys}
	primitives = append(primitives, prim)
	return &Function{primitive: prim}
}

type Frame struct {
	locals    *Frame
	previous  *Frame
	code      *Code
	ops       []int
	elements  []Value
	firstfive [5]Value
	pc        int
}

func (frame *Frame) String() string {
	var buf bytes.Buffer
	buf.WriteString("#[frame ")
	if frame.code != nil {
		if frame.code.name != "" {
			buf.WriteString(" " + frame.code.name)
		} else {
			buf.WriteString(" (anonymous code)")
		}
	} else {
		buf.WriteString(" (no code)")
	}
	buf.WriteString(fmt.Sprintf(" previous: %v", frame.previous))
	buf.WriteString("]")
	return buf.String()
}

func buildFrame(env *Frame, pc int, ops []int, fun *Function, argc int, stack []Value, sp int) (*Frame, error) {
	f := &Frame{
		previous: env,
		pc:       pc,
		ops:      ops,
		locals:   fun.frame,
		code:     fun.code,
	}
	expectedArgc := fun.code.argc
	defaults := fun.code.defaults
	if defaults == nil {
		if argc != expectedArgc {
			return nil, NewError(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
		}
		if argc <= 5 {
			f.elements = f.firstfive[:]
		} else {
			f.elements = make([]Value, argc)
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
			return nil, NewError(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected at least ", expectedArgc, ", got ", argc, ")")
		}
		return nil, NewError(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
	}
	totalArgc := expectedArgc + extra
	el := make([]Value, totalArgc)
	end := sp + expectedArgc
	if rest {
		copy(el, stack[sp:end])
		restElements := stack[end : sp+argc]
		el[expectedArgc] = ListFromValues(restElements)
	} else if keys != nil {
		bindings := stack[sp+expectedArgc : sp+argc]
		if len(bindings)%2 != 0 {
			return nil, NewError(ArgumentErrorKey, "Bad keyword argument(s): ", bindings)
		}
		copy(el, stack[sp:sp+expectedArgc]) //the required ones
		for i := expectedArgc; i < totalArgc; i++ {
			el[i] = defaults[i-expectedArgc]
		}
		for i := expectedArgc; i < argc; i += 2 {
			key, err := ToSymbol(stack[sp+i])
			if err != nil {
				return nil, NewError(ArgumentErrorKey, "Bad keyword argument: ", stack[sp+1])
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
				return nil, NewError(ArgumentErrorKey, "Undefined keyword argument: ", key)
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
	if _, ok := err.(*Error); ok {
		if env.code != nil {
			if env.code.name != "throw" {
				//FIX e.Data = append(e.Data, env.code.name)
			} else if env.previous != nil {
				if env.previous.code != nil {
					//FIX e.Data = append(e.Data, env.previous.code.name)
				}
			}
		}
	}
	return err
}

func (vm *vm) keywordCall(fun *Keyword, argc int, pc int, stack []Value, sp int) (int, int, error) {
	if argc != 1 {
		return 0, 0, NewError(ArgumentErrorKey, fun.Text, " expected 1 argument, got ", argc)
	}
	v, err := Get(stack[sp], fun)
	if err != nil {
		return 0, 0, err
	}
	stack[sp] = v
	return pc, sp, nil
}

func argcError(name string, min int, max int, provided int) error {
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
	return NewError(ArgumentErrorKey, fmt.Sprintf("%s expected %s, got %d", name, s, provided))
}

func (vm *vm) callPrimitive(prim *Primitive, argv []Value) (Value, error) {
	if prim.defaults != nil {
		return vm.callPrimitiveWithDefaults(prim, argv)
	}
	argc := len(argv)
	if argc != prim.argc {
		return nil, argcError(prim.name, prim.argc, prim.argc, argc)
	}
	for i, arg := range argv {
		t := prim.args[i]
		if t != AnyType && arg.Type() != t {
			return nil, NewError(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].String(), i+1, TypeNameOf(argv[i])))
		}
	}
	return prim.fun(argv)
}

func (vm *vm) callPrimitiveWithDefaults(prim *Primitive, argv []Value) (Value, error) {
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
			if t != AnyType && arg.Type() != t {
				return nil, NewError(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].String(), i+1, TypeNameOf(argv[i])))
			}
		}
		if rest != AnyType {
			for i := minargc; i < provided; i++ {
				arg := argv[i]
				if arg.Type() != rest {
					return nil, NewError(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, rest.String(), i+1, TypeNameOf(argv[i])))
				}
			}
		}
		return prim.fun(argv)
	}
	maxargc := len(prim.args)
	if provided < minargc {
		return nil, argcError(prim.name, minargc, maxargc, provided)
	}
	newargs := make([]Value, maxargc)
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
				return nil, NewError(ArgumentErrorKey, "mismatched keyword/value pair in argument list")
			}
			if k.Type() != KeywordType {
				return nil, NewError(ArgumentErrorKey, "expected keyword, got a "+TypeNameOf(k))
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
				return nil, NewError(ArgumentErrorKey, prim.name, " accepts ", prim.keys, " as keyword arg(s), not ", k)
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
		if t != AnyType && arg.Type() != t {
			return nil, NewError(ArgumentErrorKey, fmt.Sprintf("%s expected a %s for argument %d, got a %s", prim.name, prim.args[i].String(), i+1, TypeNameOf(argv[i])))
		}
	}
	return prim.fun(argv)
}

func (vm *vm) funcall(callable Value, argc int, ops []int, savedPc int, stack []Value, sp int, env *Frame) ([]int, int, int, *Frame, error) {
opcodeCallAgain:
	if fun, ok := callable.(*Function); ok {
		if fun.code != nil {
			if interrupted || checkInterrupt() {
				return nil, 0, 0, nil, addContext(env, NewError(InterruptKey)) //not catchable
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
					return nil, 0, 0, nil, NewError(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				if argc <= 5 {
					f.elements = f.firstfive[:argc]
				} else {
					f.elements = make([]Value, argc)
				}
				endSp := sp + argc
				copy(f.elements, stack[sp:endSp])
				return fun.code.ops, 0, endSp, f, nil
			}
			f, err := buildFrame(env, savedPc, ops, fun, argc, stack, sp)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp += argc
			env = f
			ops = fun.code.ops
			return ops, 0, sp, env, err
		}
		if fun.primitive != nil {
			val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = val
			return ops, savedPc, sp, env, err
		}
		if fun == Apply {
			if argc < 2 {
				err := NewError(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
				return vm.catch(err, stack, env)
			}
			callable = stack[sp]
			args := stack[sp+argc-1]
			if !IsList(args) {
				err := NewError(ArgumentErrorKey, "apply expected a <list> as its final argument")
				return vm.catch(err, stack, env)
			}
			arglist := args.(*List)
			for i := argc - 2; i > 0; i-- {
				arglist = Cons(stack[sp+i], arglist)
			}
			sp += argc
			argc = ListLength(arglist)
			i := 0
			sp -= argc
			for arglist != EmptyList {
				stack[sp+i] = arglist.Car
				i++
				arglist = arglist.Cdr
			}
			goto opcodeCallAgain
		}
		if fun == CallCC {
			if argc != 1 {
				err := NewError(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			callable = stack[sp]
			stack[sp] = NewContinuation(env, ops, savedPc, stack[sp+1:])
			goto opcodeCallAgain
		}
		if fun.continuation != nil {
			if argc != 1 {
				err := NewError(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			arg := stack[sp]
			sp = len(stack) - len(fun.continuation.stack)
			segment := stack[sp:]
			copy(segment, fun.continuation.stack)
			sp--
			stack[sp] = arg
			return fun.continuation.ops, fun.continuation.pc, sp, fun.frame, nil
		}
		if fun == Spawn {
			err := vm.spawn(stack[sp], argc-1, stack, sp+1)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = Null
			return ops, savedPc, sp, env, err
		}
		panic("unsupported instruction")
	}
	if kw, ok := callable.(*Keyword); ok {
		if argc != 1 {
			err := NewError(ArgumentErrorKey, kw.Text, " expected 1 argument, got ", argc)
			return vm.catch(err, stack, env)
		}
		v, err := Get(stack[sp], kw)
		if err != nil {
			return vm.catch(err, stack, env)
		}
		stack[sp] = v
		return ops, savedPc, sp, env, err
	}
	err := NewError(ArgumentErrorKey, "Not callable: ", callable)
	return vm.catch(err, stack, env)
}

func (vm *vm) tailcall(callable Value, argc int, stack []Value, sp int, env *Frame) ([]int, int, int, *Frame, error) {
opcodeTailCallAgain:
	if fun, ok := callable.(*Function); ok {
		if fun.code != nil {
			if fun.code.defaults == nil && fun.code == env.code { //self-tail-call - we can reuse the frame.
				expectedArgc := fun.code.argc
				if argc != expectedArgc {
					return nil, 0, 0, nil, NewError(ArgumentErrorKey, "Wrong number of args to ", fun, " (expected ", expectedArgc, ", got ", argc, ")")
				}
				endSp := sp + argc
				copy(env.elements, stack[sp:endSp])
				return fun.code.ops, 0, endSp, env, nil
			}
			f, err := buildFrame(env.previous, env.pc, env.ops, fun, argc, stack, sp)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp += argc
			return fun.code.ops, 0, sp, f, nil
		}
		if fun.primitive != nil {
			val, err := vm.callPrimitive(fun.primitive, stack[sp:sp+argc])
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = val
			return env.ops, env.pc, sp, env.previous, nil
		}
		if fun == Apply {
			if argc < 2 {
				err := NewError(ArgumentErrorKey, "apply expected at least 2 arguments, got ", argc)
				return vm.catch(err, stack, env)
			}
			callable = stack[sp]
			args := stack[sp+argc-1]
			if !IsList(args) {
				err := NewError(ArgumentErrorKey, "apply expected its last argument to be a <list>")
				return vm.catch(err, stack, env)
			}
			arglist := args.(*List)
			for i := argc - 2; i > 0; i-- {
				arglist = Cons(stack[sp+i], arglist)
			}
			sp += argc
			argc = ListLength(arglist)
			i := 0
			sp -= argc
			for arglist != EmptyList {
				stack[sp+i] = arglist.Car
				i++
				arglist = arglist.Cdr
			}
			goto opcodeTailCallAgain
		}
		if fun.continuation != nil {
			if argc != 1 {
				err := NewError(ArgumentErrorKey, "#[continuation] expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
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
				err := NewError(ArgumentErrorKey, "callcc expected 1 argument, got ", argc)
				return vm.catch(err, stack, env)
			}
			callable = stack[sp]
			stack[sp] = NewContinuation(env.previous, env.ops, env.pc, stack[sp:])
			goto opcodeTailCallAgain
		}
		if fun == Spawn {
			err := vm.spawn(stack[sp], argc-1, stack, sp+1)
			if err != nil {
				return vm.catch(err, stack, env)
			}
			sp = sp + argc - 1
			stack[sp] = Null
			return env.ops, env.pc, sp, env.previous, nil
		}
		panic("Bad function")
	}
	if kw, ok := callable.(*Keyword); ok {
		if argc != 1 {
			err := NewError(ArgumentErrorKey, kw.Text, " expected 1 argument, got ", argc)
			return vm.catch(err, stack, env)
		}
		v, err := Get(stack[sp], kw)
		if err != nil {
			return vm.catch(err, stack, env)
		}
		stack[sp] = v
		return env.ops, env.pc, sp, env.previous, nil
	}
	err := NewError(ArgumentErrorKey, "Not callable:", callable)
	return vm.catch(err, stack, env)
}

func (vm *vm) keywordTailcall(fun *Keyword, argc int, stack []Value, sp int, env *Frame) ([]int, int, int, *Frame, error) {
	if argc != 1 {
		err := NewError(ArgumentErrorKey, fun.Text, " expected 1 argument, got ", argc)
		return vm.catch(err, stack, env)
	}
	v, err := Get(stack[sp], fun)
	if err != nil {
		return vm.catch(err, stack, env)
	}
	stack[sp] = v
	return env.ops, env.pc, sp, env.previous, nil
}

func execCompileTime(code *Code, arg Value) (Value, error) {
	args := []Value{arg}
	prev := verbose
	verbose = false
	res, err := exec(code, args)
	verbose = prev
	return res, err
}

func (vm *vm) catch(err error, stack []Value, env *Frame) ([]int, int, int, *Frame, error) {
	errobj, ok := err.(Value)
	if !ok {
		errobj = MakeError(ErrorKey, NewString(err.Error()))
	}
	ghandler := GetGlobal(Intern("*top-handler*"))
	if ghandler != nil {
		if handler, ok := ghandler.(*Function); ok {
			if handler.code != nil {
				if handler.code.argc == 1 {
					sp := len(stack) - 1
					stack[sp] = errobj
					return vm.funcall(handler, 1, nil, 0, stack, sp, nil)
				}
			}
		}
	}
	return nil, 0, 0, nil, addContext(env, err)
}

func (vm *vm) spawn(callable Value, argc int, stack []Value, sp int) error {
	if fun, ok := callable.(*Function); ok {
		if fun.code != nil {
			env, err := buildFrame(nil, 0, nil, fun, argc, stack, sp)
			if err != nil {
				return err
			}
			go func(code *Code, env *Frame) {
				vm := VM(defaultStackSize)
				_, err := vm.exec(code, env)
				if err != nil {
					println("; [*** error in spawned function '", code.name, "': ", err, "]")
				} else if verbose {
					println("; [spawned function '", code.name, "' exited cleanly]")
				}
			}(fun.code, env)
			return nil
		}
		// spawning callcc, apply, and spawn instructions not supported.
		//? spawning primitives not supported. Is that important?
	}
	return NewError(ArgumentErrorKey, "Bad function for spawn: ", callable)
}

func exec(code *Code, args []Value) (Value, error) {
	vm := VM(defaultStackSize)
	if len(args) != code.argc {
		return nil, NewError(ArgumentErrorKey, "Wrong number of arguments")
	}
	env := new(Frame)
	env.elements = make([]Value, len(args))
	copy(env.elements, args)
	env.code = code
	startTime := time.Now()
	result, err := vm.exec(code, env)
	dur := time.Since(startTime)
	if err != nil {
		return nil, err
	}
	if result == nil {
		panic("result should never be nil if no error")
	}
	if verbose {
		println("; executed in ", dur)
		if !interactive {
			println("; => ", result)
		}
	}
	return result, err
}

func (vm *vm) exec(code *Code, env *Frame) (Value, error) {
	if !optimize || verbose || trace {
		return vm.instrumentedExec(code, env)
	}
	stack := make([]Value, vm.stackSize)
	sp := vm.stackSize
	ops := code.ops
	pc := 0
	var val Value
	var err error
	for {
		op := ops[pc]
		if op == opcodeCall {
			argc := ops[pc+1]
			callable := stack[sp]
			if fun, ok := callable.(*Function); ok {
				if fun.primitive != nil {
					nextSp := sp + argc
					prim := fun.primitive
					argv := stack[sp+1 : nextSp+1]
					if prim.defaults != nil {
						val, err = vm.callPrimitiveWithDefaults(prim, argv)
					} else {
						val, err = prim.fun(argv)
					}
					if err != nil {
						ops, pc, _, env, err = vm.catch(err, stack, env)
						if err != nil {
							return nil, err
						}
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
			} else if kw, ok := callable.(*Keyword); ok {
				pc, sp, err = vm.keywordCall(kw, argc, pc+2, stack, sp+1)
				if err != nil {
					ops, pc, sp, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				}
			} else {
				ops, pc, sp, env, err = vm.catch(NewError(ArgumentErrorKey, "Not callable: ", callable), stack, env)
				if err != nil {
					return nil, err
				}
			}
		} else if op == opcodeGlobal {
			sym := constants[ops[pc+1]]
			sp--
			stack[sp] = (sym.(*Symbol)).Value
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
			callable := stack[sp]
			argc := ops[pc+1]
			if fun, ok := callable.(*Function); ok {
				if fun.primitive != nil {
					nextSp := sp + argc
					prim := fun.primitive
					argv := stack[sp+1 : nextSp+1]
					if prim.defaults != nil {
						val, err = vm.callPrimitiveWithDefaults(prim, argv)
					} else {
						val, err = prim.fun(argv)
					}
					if err != nil {
						_, _, _, env, err = vm.catch(err, stack, env)
						if err != nil {
							return nil, err
						}
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
					ops, pc, sp, env, err = vm.tailcall(fun, argc, stack, sp+1, env)
					if err != nil {
						return nil, err
					}
					if env == nil {
						return stack[sp], nil
					}
				}
			} else if kw, ok := callable.(*Keyword); ok {
				ops, pc, sp, env, err = vm.keywordTailcall(kw, argc, stack, sp+1, env)
				if err != nil {
					ops, pc, sp, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				} else {
					if env == nil {
						return stack[sp], nil
					}
				}
			} else {
				ops, pc, sp, env, err = vm.catch(NewError(ArgumentErrorKey, "Not callable: ", fun), stack, env)
				if err != nil {
					return nil, err
				}
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
			stack[sp] = Closure(constants[ops[pc+1]].(*Code), env)
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
			sym := constants[ops[pc+1]].(*Symbol)
			defGlobal(sym, stack[sp])
			pc += 2
		} else if op == opcodeUndefGlobal {
			sym := constants[ops[pc+1]].(*Symbol)
			undefGlobal(sym)
			pc += 2
		} else if op == opcodeDefMacro {
			sym := constants[ops[pc+1]].(*Symbol)
			defMacro(sym, stack[sp].(*Function))
			stack[sp] = sym
			pc += 2
		} else if op == opcodeUse {
			sym := constants[ops[pc+1]].(*Symbol)
			err := Use(sym)
			if err != nil {
				ops, pc, sp, env, err = vm.catch(err, stack, env)
				if err != nil {
					return nil, err
				}
			} else {
				sp--
				stack[sp] = sym
				pc += 2
			}
		} else if op == opcodeVector {
			vlen := ops[pc+1]
			v := NewVector(stack[sp : sp+vlen]...)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else if op == opcodeStruct {
			vlen := ops[pc+1]
			v, _ := MakeStruct(stack[sp : sp+vlen])
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else {
			panic("Bad instruction")
		}
	}
}

const stackColumn = 40

func showInstruction(pc int, op int, args string, stack []Value, sp int) {
	var body string
	body = leftJustified(fmt.Sprintf("%d ", pc), 8) + leftJustified(opsyms[op].String(), 10) + args
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
			if IsWhitespace(s[i]) {
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
func showStack(stack []Value, sp int) string {
	end := len(stack)
	s := fmt.Sprintf("%d [", sp)
	limit := 5
	tail := ""
	if end-sp > limit {
		end = sp + limit
		tail = " ... "
	}
	for sp < end {
		tmp := fmt.Sprintf(" %v", Write(stack[sp]))
		s = s + truncatedObjectString(tmp, 30)
		sp++
	}
	return s + tail + " ]"
}

func primopName(v Value) string {
	f, _ := v.(*Function)
	return f.primitive.name
}

func (vm *vm) instrumentedExec(code *Code, env *Frame) (Value, error) {
	stack := make([]Value, vm.stackSize)
	sp := vm.stackSize
	ops := code.ops
	pc := 0
	var err, err2 error
	for {
		op := ops[pc]
		if op == opcodeCall { // CALL
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			argc := ops[pc+1]
			callable := stack[sp]
			if fun, ok := callable.(*Function); ok {
				if fun.primitive != nil {
					nextSp := sp + argc
					val, err := vm.callPrimitive(fun.primitive, stack[sp+1:nextSp+1])
					if err != nil {
						ops, pc, sp, env, err = vm.catch(err, stack, env)
						if err != nil {
							return nil, err
						}
					} else {
						stack[nextSp] = val
						sp = nextSp
						pc += 2
					}
				} else {
					ops, pc, sp, env, err = vm.funcall(fun, argc, ops, pc+2, stack, sp+1, env)
					if err != nil {
						return nil, err
					}
				}
			} else if kw, ok := callable.(*Keyword); ok {
				pc, sp, err = vm.keywordCall(kw, argc, pc+2, stack, sp+1)
				if err != nil {
					ops, pc, sp, env, err = vm.catch(err, stack, env)
					if err != nil {
						return nil, err
					}
				}
			} else {
				err := NewError(ArgumentErrorKey, "Not callable: ", fun)
				ops, pc, sp, env, err2 = vm.catch(err, stack, env)
				if err2 != nil {
					return nil, err2
				}
			}
		} else if op == opcodeGlobal { //GObjectAL
			sym := constants[ops[pc+1]].(*Symbol)
			if sym.Value == nil {
				err := NewError(ErrorKey, "Undefined symbol: ", sym)
				ops, pc, sp, env, err2 = vm.catch(err, stack, env)
				if err2 != nil {
					return nil, err2
				}
			} else {
				if trace {
					showInstruction(pc, op, sym.Text, stack, sp)
				}
				sp--
				stack[sp] = sym.Value
				pc += 2
			}
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
			if interrupted || checkInterrupt() {
				return nil, addContext(env, NewError(InterruptKey)) //not catchable
			}
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			callable := stack[sp]
			argc := ops[pc+1]
			if fun, ok := callable.(*Function); ok {
				if fun.primitive != nil {
					nextSp := sp + argc
					val, err := vm.callPrimitive(fun.primitive, stack[sp+1:nextSp+1])
					if err != nil {
						ops, pc, sp, env, err = vm.catch(err, stack, env)
						if err != nil {
							return nil, err
						}
					} else {
						stack[nextSp] = val
						sp = nextSp
						ops = env.ops
						pc = env.pc
						env = env.previous
						if env == nil {
							return stack[sp], nil
						}
					}
				} else {
					ops, pc, sp, env, err = vm.tailcall(fun, argc, stack, sp+1, env)
					if err != nil {
						return nil, err
					}
					if env == nil {
						return stack[sp], nil
					}
				}
			} else if kw, ok := callable.(*Keyword); ok {
				ops, pc, sp, env, err = vm.keywordTailcall(kw, argc, stack, sp+1, env)
				if err != nil {
					return nil, err
				}
				if env.previous == nil {
					return stack[sp], nil
				}
			} else {
				return nil, addContext(env, NewError(ArgumentErrorKey, "Not callable: ", fun))
			}
		} else if op == opcodeLiteral {
			if trace {
				showInstruction(pc, op, Write(constants[ops[pc+1]].Type()), stack, sp)
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
			stack[sp] = Closure((constants[ops[pc+1]].(*Code)), env)
			pc = pc + 2
		} else if op == opcodeReturn {
			if interrupted || checkInterrupt() {
				return nil, addContext(env, NewError(InterruptKey)) //not catchable
			}
			if trace {
				showInstruction(pc, op, "", stack, sp)
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
			sym := constants[ops[pc+1]].(*Symbol)
			if trace {
				showInstruction(pc, op, sym.Text, stack, sp)
			}
			defGlobal(sym, stack[sp])
			//fmt.Println(";", sym)
			pc += 2
		} else if op == opcodeUndefGlobal {
			sym := constants[ops[pc+1]].(*Symbol)
			if trace {
				showInstruction(pc, op, sym.Text, stack, sp)
			}
			undefGlobal(sym)
			pc += 2
		} else if op == opcodeDefMacro {
			sym := constants[ops[pc+1]].(*Symbol)
			if trace {
				showInstruction(pc, op, sym.Text, stack, sp)
			}
			defMacro(sym, stack[sp].(*Function))
			stack[sp] = sym
			pc += 2
		} else if op == opcodeUse {
			sym := constants[ops[pc+1]].(*Symbol)
			if trace {
				showInstruction(pc, op, sym.Text, stack, sp)
			}
			err := Use(sym)
			if err != nil {
				ops, pc, sp, env, err = vm.catch(err, stack, env)
				if err != nil {
					return nil, err
				}
			}
			sp--
			stack[sp] = sym
			pc += 2
		} else if op == opcodeVector {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			vlen := ops[pc+1]
			v := NewVector(stack[sp : sp+vlen]...)
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else if op == opcodeStruct {
			if trace {
				showInstruction(pc, op, fmt.Sprintf("%d", ops[pc+1]), stack, sp)
			}
			vlen := ops[pc+1]
			v, _ := MakeStruct(stack[sp : sp+vlen])
			sp = sp + vlen - 1
			stack[sp] = v
			pc += 2
		} else {
			panic("Bad instruction")
		}
	}
}
