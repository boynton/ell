package main

// the primitive functions for the languages
import (
	"fmt"
	. "github.com/boynton/gell"
)

type EllPrimitives struct {
}

func argcError() (LObject, LError) {
	return nil, Error("Wrong number of arguments")
}

func typeError(expected string, num int) (LObject, LError) {
	return nil, Error("Argument", num, "is not of type", expected)
}

func (ell EllPrimitives) Init(module LModule) error {
	module.RegisterPrimitive("display", ell_display)
	module.RegisterPrimitive("newline", ell_newline)
	module.RegisterPrimitive("print", ell_print)
	module.RegisterPrimitive("println", ell_println)
	module.RegisterPrimitive("list", ell_list)
	module.RegisterPrimitive("+", ell_plus)
	module.RegisterPrimitive("-", ell_minus)
	module.RegisterPrimitive("*", ell_times)
	module.RegisterPrimitive("quotient", ell_quotient)
	module.RegisterPrimitive("remainder", ell_remainder)
	module.RegisterPrimitive("modulo", ell_remainder) //fix!
	module.RegisterPrimitive("make-vector", ell_make_vector)
	module.RegisterPrimitive("vector-set!", ell_vector_set_bang)
	module.RegisterPrimitive("vector-ref", ell_vector_ref)
	module.RegisterPrimitive("=", ell_eq)
	module.RegisterPrimitive("<=", ell_le)
	module.RegisterPrimitive(">=", ell_ge)
	module.RegisterPrimitive(">", ell_gt)
	module.RegisterPrimitive("<", ell_lt)
	module.RegisterPrimitive("zero?", ell_zero_p)
	module.RegisterPrimitive("number->string", ell_number_to_string)
	module.RegisterPrimitive("string-length", ell_string_length)
	return nil
}

const red = "\033[0;31m"
const black = "\033[0;0m"

func ell_display(argv []LObject, argc int) (LObject, LError) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%s%v%s", red, argv[0], black)
	//fmt.Printf("%v", argv[0])
	return nil, nil
}

func ell_newline(argv []LObject, argc int) (LObject, LError) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func ell_print(argv []LObject, argc int) (LObject, LError) {
	fmt.Printf(red)
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	fmt.Printf(black)
	return nil, nil
}

func ell_println(argv []LObject, argc int) (LObject, LError) {
	ell_print(argv, argc)
	fmt.Println("")
	return nil, nil
}

func ell_list(argv []LObject, argc int) (LObject, LError) {
	var p LObject
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func ell_quotient(argv []LObject, argc int) (LObject, LError) {
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

func ell_remainder(argv []LObject, argc int) (LObject, LError) {
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

func ell_plus(argv []LObject, argc int) (LObject, LError) {
	if argc == 2 {
		return Add(argv[0], argv[1])
	} else {
		return Sum(argv, argc)
	}
}

func ell_minus(argv []LObject, argc int) (LObject, LError) {
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

func ell_times(argv []LObject, argc int) (LObject, LError) {
	if argc == 2 {
		return Mul(argv[0], argv[1])
	} else {
		return Product(argv, argc)
	}
}

func ell_make_vector(argv []LObject, argc int) (LObject, LError) {
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

func ell_vector_set_bang(argv []LObject, argc int) (LObject, LError) {
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

func ell_vector_ref(argv []LObject, argc int) (LObject, LError) {
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

func ell_ge(argv []LObject, argc int) (LObject, LError) {
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

func ell_le(argv []LObject, argc int) (LObject, LError) {
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

func ell_gt(argv []LObject, argc int) (LObject, LError) {
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

func ell_lt(argv []LObject, argc int) (LObject, LError) {
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

func ell_eq(argv []LObject, argc int) (LObject, LError) {
	if argc == 2 {
		b := Equal(argv[0], argv[1])
		return b, nil
	} else {
		return argcError()
	}
}

func ell_zero_p(argv []LObject, argc int) (LObject, LError) {
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

func ell_number_to_string(argv []LObject, argc int) (LObject, LError) {
	if argc != 1 {
		return argcError()
	}
	if !IsNumber(argv[0]) {
		return nil, Error("Not a number:", argv[0])
	}
	return NewString(argv[0].String()), nil
}

func ell_string_length(argv []LObject, argc int) (LObject, LError) {
	if argc != 1 {
		return argcError()
	}
	if !IsString(argv[0]) {
		return typeError("string", 1)
	}
	i := Length(argv[0])
	return NewInteger(int64(i)), nil
}

func NewEllPrimitives() Primitives {
	return EllPrimitives{}
}
