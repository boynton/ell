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

func Ell(module LModule) {
	module.Define("nil", NIL)
	module.Define("null", NIL)
	module.Define("true", TRUE)
	module.Define("false", FALSE)

	module.DefineMacro("define", ell_define)
	module.DefineMacro("let", ell_let)
	module.DefineMacro("letrec", ell_letrec)

	module.DefineFunction("type", ell_type)
	module.DefineFunction("equal?", ell_eq)
	module.DefineFunction("identical?", ell_identical_p)

	module.DefineFunction("null?", ell_null_p)
	module.DefineFunction("cons", ell_cons)
	module.DefineFunction("car", ell_car)
	module.DefineFunction("cdr", ell_cdr)

	module.DefineFunction("cadr", ell_cadr)
	module.DefineFunction("cddr", ell_cddr)
	module.DefineFunction("display", ell_display)
	module.DefineFunction("write", ell_write)
	module.DefineFunction("newline", ell_newline)
	module.DefineFunction("print", ell_print)
	module.DefineFunction("println", ell_println)
	module.DefineFunction("list", ell_list)
	module.DefineFunction("+", ell_plus)
	module.DefineFunction("-", ell_minus)
	module.DefineFunction("*", ell_times)
	module.DefineFunction("quotient", ell_quotient)
	module.DefineFunction("remainder", ell_remainder)
	module.DefineFunction("modulo", ell_remainder) //fix!
	module.DefineFunction("make-vector", ell_make_vector)
	module.DefineFunction("vector-set!", ell_vector_set_bang)
	module.DefineFunction("vector-ref", ell_vector_ref)
	module.DefineFunction("get", ell_get)
	module.DefineFunction("put!", ell_put_bang)
	module.DefineFunction("has?", ell_has_p)
	module.DefineFunction("=", ell_numeq)
	module.DefineFunction("<=", ell_le)
	module.DefineFunction(">=", ell_ge)
	module.DefineFunction(">", ell_gt)
	module.DefineFunction("<", ell_lt)
	module.DefineFunction("zero?", ell_zero_p)
	module.DefineFunction("number->string", ell_number_to_string)
	module.DefineFunction("string-length", ell_string_length)
	module.DefineFunction("error", ell_fatal)
	module.DefineFunction("length", ell_length)
	module.DefineFunction("json", ell_json)
}

func ell_type(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return argv[0].Type(), nil
}

func ell_identical_p(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b := Identical(argv[0], argv[1])
		if b {
			return TRUE, nil
		} else {
			return FALSE, nil
		}
	} else {
		return argcError()
	}
}

func ell_eq(argv []LObject, argc int) (LObject, error) {
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

func ell_numeq(argv []LObject, argc int) (LObject, error) {
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

func ell_display(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func ell_write(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func ell_newline(argv []LObject, argc int) (LObject, error) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func ell_fatal(argv []LObject, argc int) (LObject, error) {
	s := ""
	for _, o := range argv {
		s += fmt.Sprintf("%v", o)
	}
	return nil, Error(s)
}

func ell_print(argv []LObject, argc int) (LObject, error) {
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	return nil, nil
}

func ell_println(argv []LObject, argc int) (LObject, error) {
	ell_print(argv, argc)
	fmt.Println("")
	return nil, nil
}

func ell_list(argv []LObject, argc int) (LObject, error) {
	var p LObject
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ell_quotient(argv []LObject, argc int) (LObject, error) {
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
	} else {
		return argcError()
	}
}

func ell_remainder(argv []LObject, argc int) (LObject, error) {
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
	} else {
		return argcError()
	}
}

func ell_plus(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Add(argv[0], argv[1])
	} else {
		return Sum(argv, argc)
	}
}

func ell_minus(argv []LObject, argc int) (LObject, error) {
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

func ell_times(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Mul(argv[0], argv[1])
	} else {
		return Product(argv, argc)
	}
}

func ell_make_vector(argv []LObject, argc int) (LObject, error) {
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
	} else {
		return argcError()
	}
}

func ell_vector_set_bang(argv []LObject, argc int) (LObject, error) {
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

func ell_vector_ref(argv []LObject, argc int) (LObject, error) {
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

func ell_ge(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := GreaterOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	} else {
		return argcError()
	}
}

func ell_le(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := LessOrEqual(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	} else {
		return argcError()
	}
}

func ell_gt(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := Greater(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	} else {
		return argcError()
	}
}

func ell_lt(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b, err := Less(argv[0], argv[1])
		if err != nil {
			return nil, err
		}
		return b, nil
	} else {
		return argcError()
	}
}

func ell_zero_p(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		f, err := RealValue(argv[0])
		if err != nil {
			return nil, err
		}
		if f == 0 {
			return TRUE, nil
		} else {
			return FALSE, nil
		}
	} else {
		return argcError()
	}
}

func ell_number_to_string(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsNumber(argv[0]) {
		return nil, Error("Not a number:", argv[0])
	}
	return NewString(argv[0].String()), nil
}

func ell_string_length(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsString(argv[0]) {
		return typeError("string", 1)
	}
	i := Length(argv[0])
	return NewInteger(int64(i)), nil
}

func ell_length(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		return NewInteger(int64(Length(argv[0]))), nil
	} else {
		return argcError()
	}
}

func ell_null_p(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		if argv[0] == NIL {
			return TRUE, nil
		} else {
			return FALSE, nil
		}
	} else {
		return argcError()
	}
}

func ell_car(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Car(lst), nil
		}
		return typeError("pair", 1)
	} else {
		return argcError()
	}
}

func ell_cdr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cdr(lst), nil
		}
		return typeError("pair", 1)
	} else {
		return argcError()
	}
}

func ell_cadr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cadr(lst), nil
		}
		return typeError("pair", 1)
	} else {
		return argcError()
	}
}

func ell_cddr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cddr(lst), nil
		}
		return typeError("pair", 1)
	} else {
		return argcError()
	}
}

func ell_cons(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Cons(argv[0], argv[1]), nil
	} else {
		return argcError()
	}
}

func ell_define(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return ExpandDefine(argv[0])
}

func ell_letrec(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return ExpandLetrec(argv[0])
}

func ell_let(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	return ExpandLet(argv[0])
}

func ell_get(argv []LObject, argc int) (LObject, error) {
	if argc != 2 {
		return argcError()
	}
	return Get(argv[0], argv[1])
}

func ell_has_p(argv []LObject, argc int) (LObject, error) {
	if argc != 2 {
		return argcError()
	}
	b, err := Has(argv[0], argv[1])
	if err != nil {
		return nil, err
	} else if b {
		return TRUE, nil
	} else {
		return FALSE, nil
	}
}

func ell_put_bang(argv []LObject, argc int) (LObject, error) {
	if argc != 3 {
		return argcError()
	}
	return Put(argv[0], argv[1], argv[2])
}

func ell_json(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	s, err := JSON(argv[0])
	if err != nil {
		return nil, err
	}
	return NewString(s), nil
}
