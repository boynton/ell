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

func EllPrimitives() map[string]Primitive {
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
	m["use"] = ell_use
	return m
}

const terminalRed = "\033[0;31m"
const terminalBlack = "\033[0;0m"

func ell_display(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%s%v%s", terminalRed, argv[0], terminalBlack)
	//fmt.Printf("%v", argv[0])
	return nil, nil
}

func ell_newline(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func ell_print(module LModule, argv []LObject, argc int) (LObject, error) {
	fmt.Printf(terminalRed)
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	fmt.Printf(terminalBlack)
	return nil, nil
}

func ell_println(module LModule, argv []LObject, argc int) (LObject, error) {
	ell_print(module, argv, argc)
	fmt.Println("")
	return nil, nil
}

func ell_list(module LModule, argv []LObject, argc int) (LObject, error) {
	var p LObject
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ell_quotient(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_remainder(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_plus(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Add(argv[0], argv[1])
	} else {
		return Sum(argv, argc)
	}
}

func ell_minus(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_times(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Mul(argv[0], argv[1])
	} else {
		return Product(argv, argc)
	}
}

func ell_make_vector(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_vector_set_bang(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_vector_ref(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_ge(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_le(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_gt(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_lt(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_eq(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b := Equal(argv[0], argv[1])
		return b, nil
	} else {
		return argcError()
	}
}

func ell_zero_p(module LModule, argv []LObject, argc int) (LObject, error) {
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

func ell_number_to_string(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsNumber(argv[0]) {
		return nil, Error("Not a number:", argv[0])
	}
	return NewString(argv[0].String()), nil
}

func ell_string_length(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsString(argv[0]) {
		return typeError("string", 1)
	}
	i := Length(argv[0])
	return NewInteger(int64(i)), nil
}

func ell_use(module LModule, argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsSymbol(argv[0]) {
		return typeError("symbol", 1)
	}
	name := argv[0].String()
	thunk, err := LoadModule(name, EllPrimitives())
	if err != nil {
		return nil, err
	}
	moduleToUse := thunk.Module()
	_, err = Exec(thunk)
	exports := moduleToUse.Exports()
	for _, sym := range exports {
		val := moduleToUse.Global(sym)
		module.DefGlobal(sym, val)
	}
	return argv[0], nil
}
