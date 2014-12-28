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
	return nil, Error("Argument", num, "is not of type", expected)
}

func EllPrimitiveFunctions() map[string]Primitive {
	m := make(map[string]Primitive)
	m["display"] = ell_display
	m["newline"] = ell_newline
	m["print"] = ell_print
	m["println"] = ell_println
	m["list"] = ell_list
	m["+"] = ell_plus
	m["-"] = ell_minus
	m["*"] = ell_times
	m["quotient"] = ell_quotient
	m["remainder"] = ell_remainder
	m["modulo"] = ell_remainder //fix!
	m["make-vector"] = ell_make_vector
	m["vector-set!"] = ell_vector_set_bang
	m["vector-ref"] = ell_vector_ref
	m["="] = ell_eq
	m["<="] = ell_le
	m[">="] = ell_ge
	m[">"] = ell_gt
	m["<"] = ell_lt
	m["zero?"] = ell_zero_p
	m["number->string"] = ell_number_to_string
	m["string-length"] = ell_string_length
	m["error"] = ell_fatal
	m["length"] = ell_length
	m["cadr"] = ell_cadr
	m["cddr"] = ell_cddr
	m["cons"] = ell_cons
	return m
}
func EllPrimitiveMacros() map[string]Primitive {
	m := make(map[string]Primitive)
	m["define"] = ell_define
	return m
}

const terminalRed = "\033[0;31m"
const terminalBlack = "\033[0;0m"

func ell_display(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%s%v%s", terminalRed, argv[0], terminalBlack)
	//fmt.Printf("%v", argv[0])
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
	fmt.Printf(terminalRed)
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	fmt.Printf(terminalBlack)
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

func ell_eq(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b := Equal(argv[0], argv[1])
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

func ell_cadr(argv []LObject, argc int) (LObject, error) {
	if argc == 1 {
		lst := argv[0]
		if IsList(lst) {
			return Cadr(lst), nil
		}
		return typeError("list", 1)
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
		return typeError("list", 1)
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
	expr := argv[0]
	exprLen := Length(expr)
	if exprLen < 3 {
		return nil, Error("syntax error: ", expr)
	}
	sym := Cadr(expr)
	if !IsList(sym) {
		//let it pass through, let the compiler syntax check the primitive define form
		return expr, nil
	}
	args := Cdr(sym)
	sym = Car(sym)
	return List(Car(expr), sym, Cons(Intern("lambda"), Cons(args, Cddr(expr)))), nil
}
