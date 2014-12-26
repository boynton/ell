package gell

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	LITERAL_OPCODE   = 1
	LOCAL_OPCODE     = 2
	JUMPFALSE_OPCODE = 3
	JUMP_OPCODE      = 4
	TAILCALL_OPCODE  = 5
	CALL_OPCODE      = 6
	RETURN_OPCODE    = 7
	CLOSURE_OPCODE   = 8
	POP_OPCODE       = 9
	GLOBAL_OPCODE    = 10
	DEFGLOBAL_OPCODE = 11
	SETLOCAL_OPCODE  = 12

	NULLP_OPCODE = 13
	CAR_OPCODE   = 14
	CDR_OPCODE   = 15
	ADD_OPCODE   = 16
	MUL_OPCODE   = 17
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

type LModule interface {
	Type() LSymbol
	String() string
	RegisterPrimitive(name string, fun Primitive)
	Global(sym LSymbol) (LObject, bool)
}

type tModule struct {
	Name         string
	globals      map[LSymbol]LObject
	constantsMap map[LObject]int
	constants    []LObject
	exports      []LObject
}

func (tModule) Type() LSymbol {
	return Intern("module")
}

func (module tModule) String() string {
	return fmt.Sprintf("<module %v, constants:%v>", module.Name, module.constants)
}

//note: unlike java, we cannot use maps or arrays as keys (they are not comparable).
//so, we will end up with duplicates, unless we do some deep compare...
//idea: I'd like all Ell objects to have a hashcode. Use that.
func (module *tModule) putConstant(val LObject) int {
	idx, present := module.constantsMap[val]
	if !present {
		idx = len(module.constants)
		module.constants = append(module.constants, val)
		module.constantsMap[val] = idx
	}
	return idx
}

type Primitive func(argv []LObject, argc int) (LObject, LError)

type tPrimitive struct {
	name string
	fun  Primitive
}

var symPrimitive = newSymbol("primitive")

func (prim tPrimitive) Type() LSymbol {
	return symPrimitive
}

func (prim tPrimitive) String() string {
	return "<primitive " + prim.name + ">"
}

func (module tModule) globalRef(idx int) (LObject, bool) {
	sym := module.constants[idx]
	v := (sym.(*lsymbol)).value
	if v == nil {
		return nil, false
	}
	return v, true
}

func (module tModule) globalDefine(idx int, obj LObject) LObject {
	sym := module.constants[idx]
	(sym.(*lsymbol)).value = obj
	return sym
}

func (module tModule) Global(sym LSymbol) (LObject, bool) {
	v := (sym.(*lsymbol)).value
	b := v != nil
	return v, b
}

func (module tModule) SetGlobal(sym LSymbol, val LObject) {
	(sym.(*lsymbol)).value = val
}

func (module tModule) RegisterPrimitive(name string, fun Primitive) {
	sym := Intern(name)
	_, ok := module.Global(sym)
	if ok {
		Println("*** Warning: redefining ", name)
		//check the argument signature. Define "primitiveN" differently than "primitive0" .. "primitive3"
	}
	po := tPrimitive{name, fun}
	module.SetGlobal(sym, po)
}

func MakeModule(name string, primitives Primitives) (LModule, error) {
	globals := map[LSymbol]LObject{}
	constMap := map[LObject]int{}
	constants := make([]LObject, 0)
	mod := tModule{name, globals, constMap, constants, nil}
	if primitives != nil {
		err := primitives.Init(mod)
		if err != nil {
			return mod, nil
		}
	}
	return &mod, nil
}

type LCode interface {
	Type() LSymbol
	String() string
	LoadOps(ops LObject) LError
	EmitLiteral(val LObject)
	EmitGlobal(sym LObject)
	EmitCall(argc int)
	EmitReturn()
	EmitTailCall(argc int)
	EmitPop()
	EmitLocal(i int, j int)
	EmitSetLocal(i int, j int)
	EmitDefGlobal(sym LObject)
	EmitClosure(code LCode)
	EmitJumpFalse(offset int) int
	EmitJump(offset int) int
	SetJumpLocation(loc int)
	EmitCar()
	EmitCdr()
	EmitNullP()
	EmitAdd()
	EmitMul()
}

type lcode struct {
	module *tModule
	ops    []int
	argc   int
	rest   LObject
}

func NewCode(module LModule, argc int, restArgs LObject) LCode {
	ops := make([]int, 0)
	mod := module.(*tModule)
	code := lcode{mod, ops, argc, restArgs}
	return &code
}

func (lcode) Type() LSymbol {
	return Intern("code")
}

func (code lcode) decompile(buf *bytes.Buffer, indent string) {
	offset := 0
	max := len(code.ops)
	buf.WriteString("(function ")
	buf.WriteString(strconv.Itoa(code.argc))
	if true {
		buf.WriteString(fmt.Sprintf(" %v)", code.ops)) //HACK
		return
	}
	for offset < max {
		switch code.ops[offset] {
		case LITERAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(literal %d)  \t; %v\n", indent, offset, code.ops[offset+1], code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (literal " + Write(code.module.constants[code.ops[offset+1]]) + ")")
			offset += 2
		case GLOBAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(global %v)\n", indent, offset, code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (global " + Write(code.module.constants[code.ops[offset+1]]) + ")")
			offset += 2
		case CALL_OPCODE:
			//fmt.Printf("%sL%03d:\t(call %d)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(" (call " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case TAILCALL_OPCODE:
			//fmt.Printf("%s%03d:\t(tailcall %d)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(" (tailcall " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case POP_OPCODE:
			//fmt.Printf("%sL%03d:\t(pop)\n", indent, offset)
			buf.WriteString(" (pop)")
			offset += 1
		case RETURN_OPCODE:
			//fmt.Printf("%sL%03d:\t(return)\n", indent, offset)
			buf.WriteString(" (return)")
			offset += 1
		case CLOSURE_OPCODE:
			//fmt.Printf("%sL%03d:\t(closure %v)\n", indent, offset, code.ops[offset+1])
			buf.WriteString(" (closure ")
			(code.module.constants[code.ops[offset+1]].(*lcode)).decompile(buf, indent+"\t")
			buf.WriteString(")")
			offset += 2
		case LOCAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(local %d %d)\n", indent, offset, code.ops[offset+1], code.ops[offset+2])
			buf.WriteString(" (local " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case DEFGLOBAL_OPCODE:
			//fmt.Printf("%sL%03d:\t(defglobal%6d ; %v)\n", indent, offset, code.ops[offset+1], code.module.constants[code.ops[offset+1]])
			buf.WriteString(" (defglobal " + Write(code.module.constants[code.ops[offset+1]]) + ")")
			offset += 2
		case SETLOCAL_OPCODE:
			//Println("%sL%03d:\t(setlocal %d %d)\n", indent, offset, code.ops[offset+1], code.ops[offset+2])
			buf.WriteString(" (setlocal " + strconv.Itoa(code.ops[offset+1]) + " " + strconv.Itoa(code.ops[offset+2]) + ")")
			offset += 3
		case JUMPFALSE_OPCODE:
			//fmt.Printf("%sL%03d:\t(jumpfalse %d)\t; L%03d\n", indent, offset, code.ops[offset+1], code.ops[offset+1] + offset)
			buf.WriteString(" (jumpfalse " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case JUMP_OPCODE:
			//fmt.Printf("%sL%03d:\t(jump %d)    \t; L%03d\n", indent, offset, code.ops[offset+1], code.ops[offset+1] + offset)
			buf.WriteString(" (jump " + strconv.Itoa(code.ops[offset+1]) + ")")
			offset += 2
		case CAR_OPCODE:
			buf.WriteString(" (car)")
			offset += 1
		case CDR_OPCODE:
			buf.WriteString(" (cdr)")
			offset += 1
		case NULLP_OPCODE:
			buf.WriteString(" (null?)")
			offset += 1
		case ADD_OPCODE:
			buf.WriteString(" (add)")
			offset += 1
		case MUL_OPCODE:
			buf.WriteString(" (mul)")
			offset += 1
		default:
			buf.WriteString("FIX ME: " + strconv.Itoa(code.ops[offset]))
			break
		}
	}
	buf.WriteString(")")
}

func (code lcode) String() string {
	var buf bytes.Buffer
	//buf.WriteString(fmt.Sprintf("<code ops:%v constants:%v>", code.ops, code.module.constants)) //HACK
	//buf.WriteString(fmt.Sprintf("<code ops:%v>", code.ops)) //HACK
	code.decompile(&buf, "")
	return buf.String()
}

func (code *lcode) LoadOps(lst LObject) LError {
	return Error("LoadOps NYI")
}

func (code *lcode) EmitLiteral(val LObject) {
	code.ops = append(code.ops, LITERAL_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(val))
}

func (code *lcode) EmitGlobal(sym LObject) {
	code.ops = append(code.ops, GLOBAL_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(sym))
}
func (code *lcode) EmitCall(argc int) {
	code.ops = append(code.ops, CALL_OPCODE)
	code.ops = append(code.ops, argc)
}
func (code *lcode) EmitReturn() {
	code.ops = append(code.ops, RETURN_OPCODE)
}
func (code *lcode) EmitTailCall(argc int) {
	code.ops = append(code.ops, TAILCALL_OPCODE)
	code.ops = append(code.ops, argc)
}
func (code *lcode) EmitPop() {
	code.ops = append(code.ops, POP_OPCODE)
}
func (code *lcode) EmitLocal(i int, j int) {
	code.ops = append(code.ops, LOCAL_OPCODE)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *lcode) EmitSetLocal(i int, j int) {
	code.ops = append(code.ops, SETLOCAL_OPCODE)
	code.ops = append(code.ops, i)
	code.ops = append(code.ops, j)
}
func (code *lcode) EmitDefGlobal(sym LObject) {
	code.ops = append(code.ops, DEFGLOBAL_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(sym))
}
func (code *lcode) EmitClosure(newCode LCode) {
	code.ops = append(code.ops, CLOSURE_OPCODE)
	code.ops = append(code.ops, code.module.putConstant(newCode))
}
func (code *lcode) EmitJumpFalse(offset int) int {
	code.ops = append(code.ops, JUMPFALSE_OPCODE)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *lcode) EmitJump(offset int) int {
	code.ops = append(code.ops, JUMP_OPCODE)
	loc := len(code.ops)
	code.ops = append(code.ops, offset)
	return loc
}
func (code *lcode) SetJumpLocation(loc int) {
	code.ops[loc] = len(code.ops) - loc + 1
}
func (code *lcode) EmitCar() {
	code.ops = append(code.ops, CAR_OPCODE)
}
func (code *lcode) EmitCdr() {
	code.ops = append(code.ops, CDR_OPCODE)
}
func (code *lcode) EmitNullP() {
	code.ops = append(code.ops, NULLP_OPCODE)
}
func (code *lcode) EmitAdd() {
	code.ops = append(code.ops, ADD_OPCODE)
}
func (code *lcode) EmitMul() {
	code.ops = append(code.ops, MUL_OPCODE)
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

type Primitives interface {
	Init(module LModule) error
}

func RunModule(name string, primitives Primitives) (LObject, error) {
	thunk, err := LoadModule(name, primitives)
	if err != nil {
		return NIL, err
	}
	code := thunk.(*lcode)
	Println("; begin execution")
	vm := LVM{DefaultStackSize, make([]LSymbol, 0)}
	result, err := vm.exec(code)
	if err != nil {
		return NIL, err
	}
	if len(vm.defs) > 0 {
		Println("export these: ", vm.defs)
	}
	if result != nil {
		Println("; end execution")
		Println("; => ", result)
	}
	return result, nil
}

func FindModule(moduleName string) (string, error) {
	path := [...]string{"src/main/ell"} //fix
	name := moduleName
	if !strings.HasSuffix(name, ".ell") {
		name = name + ".ell"
	}
	for _, dirname := range path {
		filename := filepath.Join(dirname, name)
		if FileReadable(filename) {
			return filename, nil
		}
	}
	return "", Error("not found")
}

func LoadModule(name string, primitives Primitives) (LCode, error) {
	file := name
	i := strings.Index(name, ".")
	if i < 0 {
		f, err := FindModule(name)
		if err != nil {
			return nil, Error("Module not found:", name)
		}
		file = f
	} else {
		if !FileReadable(name) {
			return nil, Error("Cannot read file:", name)
		}
		name = name[0:i]

	}
	return LoadFileModule(name, file, primitives)
}

func LoadFileModule(moduleName string, file string, primitives Primitives) (LCode, error) {
	Println("; loadModule: " + moduleName + " from " + file)
	module, err := MakeModule(moduleName, primitives)
	if err != nil {
		return nil, err
	}
	port, err := OpenInputFile(file)
	if err != nil {
		return nil, err
	}
	source := List(Intern("begin"))
	expr, err := port.Read()
	for {
		if err != nil {
			return nil, err
		}
		if expr == EOI {
			break
		}
		source, err = Concat(source, List(expr))
		if err == nil {
			expr, err = port.Read()
		}
	}
	port.Close()
	Println("; read: ", Write(source))
	code, err := Compile(module, source)
	if err != nil {
		return nil, err
	}
	Println("; compiled to: ", Write(code))
	Println("; module: ", module)
	return code, nil
}
