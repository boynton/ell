package main

// the primitive functions for the languages
import (
	"fmt"
	. "github.com/boynton/gell"
)

func argcError() (LObject, error) {
	return nil, Error("Wrong number of arguments")
}

func typeError(expected string, num int) (LObject, error) {
	return nil, Error("Argument ", num, " is not of type ", expected)
}

func rangeError(expected string, num int) (LObject, error) {
	return nil, Error("Argument ", num, " is out of range: ", expected)
}

// Ell defines the global functions for the top level environment
func Ell(module LModule) {
	module.Define("nil", NIL)
	module.Define("null", NIL)
	module.Define("true", TRUE)
	module.Define("false", FALSE)

	module.DefineMacro("let", ellLet)
	module.DefineMacro("letrec", ellLetrec)
	module.DefineMacro("do", ellDo)

	module.DefineFunction("type", ellType)
	module.DefineFunction("equal?", ellEq)
	module.DefineFunction("identical?", ellIdenticalP)

	module.DefineFunction("null?", ellNullP)
	module.DefineFunction("cons", ellCons)
	module.DefineFunction("car", ellCar)
	module.DefineFunction("cdr", ellCdr)

	module.DefineFunction("cadr", ellCadr)
	module.DefineFunction("cddr", ellCddr)
	module.DefineFunction("display", ellDisplay)
	module.DefineFunction("write", ellWrite)
	module.DefineFunction("newline", ellNewline)
	module.DefineFunction("print", ellPrint)
	module.DefineFunction("println", ellPrintln)
	module.DefineFunction("list", ellList)
	module.DefineFunction("+", ellPlus)
	module.DefineFunction("-", ellMinus)
	module.DefineFunction("*", ellTimes)
	module.DefineFunction("quotient", ellQuotient)
	module.DefineFunction("remainder", ellRemainder)
	module.DefineFunction("modulo", ellRemainder) //fix!
	module.DefineFunction("make-vector", ellMakeVector)
	module.DefineFunction("vector-set!", ellVectorSetBang)
	module.DefineFunction("vector-ref", ellVectorRef)
	module.DefineFunction("get", ellGet)
	module.DefineFunction("put!", ellPutBang)
	module.DefineFunction("has?", ellHasP)
	module.DefineFunction("=", ellNumeq)
	module.DefineFunction("<=", ellLe)
	module.DefineFunction(">=", ellGe)
	module.DefineFunction(">", ellGt)
	module.DefineFunction("<", ellLt)
	module.DefineFunction("zero?", ellZeroP)
	module.DefineFunction("number->string", ellNumberToString)
	module.DefineFunction("string-length", ellStringLength)
	module.DefineFunction("error", ellFatal)
	module.DefineFunction("length", ellLength)
	module.DefineFunction("json", ellJSON)
}

func ellType(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return argv[0].Type(), nil
}

func ellIdenticalP(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b := Identical(argv[0], argv[1])
		if b {
			return TRUE, nil
		}
		return FALSE, nil
	}
	return argcError()
}

func ellEq(argv []LObject, argc int) (LObject, error) {
	if argc < 1 {
		return argcError()
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if !Equal(obj, argv[1]) {
			return FALSE, nil
		}
	}
	return TRUE, nil
}

func ellNumeq(argv []LObject, argc int) (LObject, error) {
	if argc < 1 {
		return argcError()
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if b, err := NumericallyEqual(obj, argv[1]); err != nil || !b {
			return FALSE, err
		}
	}
	return TRUE, nil
}

func ellDisplay(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func ellWrite(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", Write(argv[0]))
	return nil, nil
}

func ellNewline(argv []LObject, argc int) (LObject, error) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func ellFatal(argv []LObject, argc int) (LObject, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}

func ellPrint(argv []LObject, argc int) (LObject, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return nil, nil
}

func ellPrintln(argv []LObject, argc int) (LObject, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return nil, nil
}

func ellList(argv []LObject, argc int) (LObject, error) {
	var p LObject
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ellQuotient(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		n1, err := IntegerValue(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := IntegerValue(argv[1])
		if err != nil {
			return nil, err
		}
		if n2 == 0 {
			return nil, Error("Quotient: divide by zero")
		}
		return NewInteger(n1 / n2), nil
	}
	return argcError()
}

func ellRemainder(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		n1, err := IntegerValue(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := IntegerValue(argv[1])
		if n2 == 0 {
			return nil, Error("Remainder: divide by zero")
		}
		if err != nil {
			return nil, err
		}
		return NewInteger(n1 % n2), nil
	}
	return argcError()
}

func ellPlus(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Add(argv[0], argv[1])
	}
	return Sum(argv, argc)
}

func ellMinus(argv []LObject, argc int) (LObject, error) {
	//hack
	if argc != 2 {
		return argcError()
	}
	n1, err := IntegerValue(argv[0])
	if err != nil {
		return nil, err
	}
	n2, err := IntegerValue(argv[1])
	if err != nil {
		return nil, err
	}
	return NewInteger(n1 - n2), nil
}

func ellTimes(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Mul(argv[0], argv[1])
	}
	return Product(argv, argc)
}

func ellMakeVector(argv []LObject, argc int) (LObject, error) {
	if argc > 0 {
		var initVal LObject = NIL
		vlen, err := IntegerValue(argv[0])
		if err != nil {
			return nil, err
		}
		if argc > 1 {
			if argc != 2 {
				return argcError()
			}
			initVal = argv[1]
		}
		return NewVector(int(vlen), initVal), nil
	}
	return argcError()
}

func ellVectorSetBang(argv []LObject, argc int) (LObject, error) {
	if argc == 3 {
		v := argv[0]
		idx, err := IntegerValue(argv[1])
		if err != nil {
			return nil, err
		}
		err = VectorSet(v, int(idx), argv[2])
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return argcError()
}

func ellVectorRef(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		v := argv[0]
		idx, err := IntegerValue(argv[1])
		if err != nil {
			return nil, err
		}
		val, err := VectorRef(v, int(idx))
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return argcError()
}

func ellGe(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := GreaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellLe(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := LessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellGt(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := Greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellLt(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := Less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellZeroP(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		f, err := RealValue(argv[0])
		if err != nil {
			return nil, err
		}
		if f == 0 {
			return TRUE, nil
		}
		return FALSE, nil
	}
	return argcError()
}

func ellNumberToString(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsNumber(argv[0]) {
		return nil, Error("Not a number:", argv[0])
	}
	return NewString(argv[0].String()), nil
}

func ellStringLength(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsString(argv[0]) {
		return typeError("string", 1)
	}
	i := Length(argv[0])
	return NewInteger(int64(i)), nil
}

func ellLength(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		return NewInteger(int64(Length(argv[0]))), nil
	}
	return argcError()
}

func ellNullP(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		if argv[0] == NIL {
			return TRUE, nil
		}
		return FALSE, nil
	}
	return argcError()
}

func ellCar(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Car(lst), nil
		}
		return typeError("pair", 1)
	}
	return argcError()
}

func ellCdr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cdr(lst), nil
		}
		return typeError("pair", 1)
	}
	return argcError()
}

func ellCadr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cadr(lst), nil
		}
		return typeError("pair", 1)
	}
	return argcError()
}

func ellCddr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cddr(lst), nil
		}
		return typeError("pair", 1)
	}
	return argcError()
}

func ellCons(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Cons(argv[0], argv[1]), nil
	}
	return argcError()
}

func ellLetrec(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return ExpandLetrec(argv[0])
}

func ellLet(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return ExpandLet(argv[0])
}

func ellDo(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return ExpandDo(argv[0])
}

func ellGet(argv []LObject, argc int) (LObject, error) {
	if argc != 2 {
		return argcError()
	}
	return Get(argv[0], argv[1])
}

func ellHasP(argv []LObject, argc int) (LObject, error) {
	if argc != 2 {
		return argcError()
	}
	b, err := Has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return TRUE, nil
	}
	return FALSE, nil
}

func ellPutBang(argv []LObject, argc int) (LObject, error) {
	if argc != 3 {
		return argcError()
	}
	return Put(argv[0], argv[1], argv[2])
}

func ellJSON(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	s, err := JSON(argv[0])
	if err != nil {
		return nil, err
	}
	return NewString(s), nil
}
