package main

// the primitive functions for the languages
import (
	"fmt"
)

func argcError() (lob, error) {
	return nil, newError("Wrong number of arguments")
}

func ellTypeError(expected string, num int) (lob, error) {
	return nil, newError("Argument ", num, " is not of type ", expected)
}

// Ell defines the global functions for the top level environment
func Ell(module module) {
	module.define("nil", NIL)
	module.define("null", NIL)
	module.define("true", TRUE)
	module.define("false", FALSE)

	module.defineMacro("let", ellLet)
	module.defineMacro("letrec", ellLetrec)
	module.defineMacro("do", ellDo)

	module.defineFunction("type", ellType)
	module.defineFunction("equal?", ellEq)
	module.defineFunction("identical?", ellIdenticalP)

	module.defineFunction("null?", ellNullP)
	module.defineFunction("cons", ellCons)
	module.defineFunction("car", ellCar)
	module.defineFunction("cdr", ellCdr)

	module.defineFunction("cadr", ellCadr)
	module.defineFunction("cddr", ellCddr)
	module.defineFunction("display", ellDisplay)
	module.defineFunction("write", ellWrite)
	module.defineFunction("newline", ellNewline)
	module.defineFunction("print", ellPrint)
	module.defineFunction("println", ellPrintln)
	module.defineFunction("list", ellList)
	module.defineFunction("+", ellPlus)
	module.defineFunction("-", ellMinus)
	module.defineFunction("*", ellTimes)
	module.defineFunction("quotient", ellQuotient)
	module.defineFunction("remainder", ellRemainder)
	module.defineFunction("modulo", ellRemainder) //fix!
	module.defineFunction("make-vector", ellMakeVector)
	module.defineFunction("vector-set!", ellVectorSetBang)
	module.defineFunction("vector-ref", ellVectorRef)
	module.defineFunction("get", ellGet)
	module.defineFunction("put!", ellPutBang)
	module.defineFunction("has?", ellHasP)
	module.defineFunction("=", ellNumeq)
	module.defineFunction("<=", ellLe)
	module.defineFunction(">=", ellGe)
	module.defineFunction(">", ellGt)
	module.defineFunction("<", ellLt)
	module.defineFunction("zero?", ellZeroP)
	module.defineFunction("number->string", ellNumberToString)
	module.defineFunction("string-length", ellStringLength)
	module.defineFunction("error", ellFatal)
	module.defineFunction("length", ellLength)
	module.defineFunction("json", ellJSON)
}

func ellType(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	return argv[0].typeSymbol(), nil
}

func ellIdenticalP(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b := identical(argv[0], argv[1])
		if b {
			return TRUE, nil
		}
		return FALSE, nil
	}
	return argcError()
}

func ellEq(argv []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError()
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if !equal(obj, argv[1]) {
			return FALSE, nil
		}
	}
	return TRUE, nil
}

func ellNumeq(argv []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError()
	}
	obj := argv[0]
	for i := 1; i < argc; i++ {
		if b, err := numericallyEqual(obj, argv[1]); err != nil || !b {
			return FALSE, err
		}
	}
	return TRUE, nil
}

func ellDisplay(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func ellWrite(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", write(argv[0]))
	return nil, nil
}

func ellNewline(argv []lob, argc int) (lob, error) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func ellFatal(argv []lob, argc int) (lob, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, newError(s)
}

func ellPrint(argv []lob, argc int) (lob, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return nil, nil
}

func ellPrintln(argv []lob, argc int) (lob, error) {
	ellPrint(argv, argc)
	fmt.Println("")
	return nil, nil
}

func ellList(argv []lob, argc int) (lob, error) {
	var p lob
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = cons(argv[i], p)
	}
	return p, nil
}

func ellQuotient(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		n1, err := integerValue(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := integerValue(argv[1])
		if err != nil {
			return nil, err
		}
		if n2 == 0 {
			return nil, newError("Quotient: divide by zero")
		}
		return newInteger(n1 / n2), nil
	}
	return argcError()
}

func ellRemainder(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		n1, err := integerValue(argv[0])
		if err != nil {
			return nil, err
		}
		n2, err := integerValue(argv[1])
		if n2 == 0 {
			return nil, newError("Remainder: divide by zero")
		}
		if err != nil {
			return nil, err
		}
		return newInteger(n1 % n2), nil
	}
	return argcError()
}

func ellPlus(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		return add(argv[0], argv[1])
	}
	return sum(argv, argc)
}

func ellMinus(argv []lob, argc int) (lob, error) {
	//hack
	if argc != 2 {
		return argcError()
	}
	n1, err := integerValue(argv[0])
	if err != nil {
		return nil, err
	}
	n2, err := integerValue(argv[1])
	if err != nil {
		return nil, err
	}
	return newInteger(n1 - n2), nil
}

func ellTimes(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		return mul(argv[0], argv[1])
	}
	return product(argv, argc)
}

func ellMakeVector(argv []lob, argc int) (lob, error) {
	if argc > 0 {
		var initVal lob = NIL
		vlen, err := integerValue(argv[0])
		if err != nil {
			return nil, err
		}
		if argc > 1 {
			if argc != 2 {
				return argcError()
			}
			initVal = argv[1]
		}
		return newVector(int(vlen), initVal), nil
	}
	return argcError()
}

func ellVectorSetBang(argv []lob, argc int) (lob, error) {
	if argc == 3 {
		v := argv[0]
		idx, err := integerValue(argv[1])
		if err != nil {
			return nil, err
		}
		err = vectorSet(v, int(idx), argv[2])
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return argcError()
}

func ellVectorRef(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		v := argv[0]
		idx, err := integerValue(argv[1])
		if err != nil {
			return nil, err
		}
		val, err := vectorRef(v, int(idx))
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return argcError()
}

func ellGe(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := greaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellLe(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := lessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellGt(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellLt(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		b, err := less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return argcError()
}

func ellZeroP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		f, err := realValue(argv[0])
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

func ellNumberToString(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	if !isNumber(argv[0]) {
		return ellTypeError("number", 1)
	}
	return newString(argv[0].String()), nil
}

func ellStringLength(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	if !isString(argv[0]) {
		return ellTypeError("string", 1)
	}
	i := length(argv[0])
	return newInteger(int64(i)), nil
}

func ellLength(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		return newInteger(int64(length(argv[0]))), nil
	}
	return argcError()
}

func ellNullP(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		if argv[0] == NIL {
			return TRUE, nil
		}
		return FALSE, nil
	}
	return argcError()
}

func ellCar(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return car(lst), nil
		}
		return ellTypeError("pair", 1)
	}
	return argcError()
}

func ellCdr(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cdr(lst), nil
		}
		return ellTypeError("pair", 1)
	}
	return argcError()
}

func ellCadr(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cadr(lst), nil
		}
		return ellTypeError("pair", 1)
	}
	return argcError()
}

func ellCddr(argv []lob, argc int) (lob, error) {
	if argc == 1 {
		lst := argv[0]
		if isList(lst) {
			return cddr(lst), nil
		}
		return ellTypeError("pair", 1)
	}
	return argcError()
}

func ellCons(argv []lob, argc int) (lob, error) {
	if argc == 2 {
		return cons(argv[0], argv[1]), nil
	}
	return argcError()
}

func ellLetrec(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	return expandLetrec(argv[0])
}

func ellLet(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	return expandLet(argv[0])
}

func ellDo(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	return expandDo(argv[0])
}

func ellGet(argv []lob, argc int) (lob, error) {
	if argc != 2 {
		return argcError()
	}
	return get(argv[0], argv[1])
}

func ellHasP(argv []lob, argc int) (lob, error) {
	if argc != 2 {
		return argcError()
	}
	b, err := has(argv[0], argv[1])
	if err != nil {
		return nil, err
	}
	if b {
		return TRUE, nil
	}
	return FALSE, nil
}

func ellPutBang(argv []lob, argc int) (lob, error) {
	if argc != 3 {
		return argcError()
	}
	return put(argv[0], argv[1], argv[2])
}

func ellJSON(argv []lob, argc int) (lob, error) {
	if argc != 1 {
		return argcError()
	}
	s, err := toJSON(argv[0])
	if err != nil {
		return nil, err
	}
	return newString(s), nil
}
