package main

// the primitive functions for the languages
import (
	"fmt"
	. "github.com/boynton/gell"
)

type SchemePrimitives struct {
}

func (ell SchemePrimitives) Init(module LModule) error {
	module.RegisterPrimitive("display", prim_display)
	module.RegisterPrimitive("newline", prim_newline)
	module.RegisterPrimitive("list", prim_list)
	module.RegisterPrimitive("+", prim_plus)
	module.RegisterPrimitive("-", prim_minus)
	module.RegisterPrimitive("*", prim_times)
	module.RegisterPrimitive("quotient", prim_quotient)
	module.RegisterPrimitive("remainder", prim_remainder)
	module.RegisterPrimitive("modulo", prim_remainder) //fix!
	module.RegisterPrimitive("make-vector", prim_make_vector)
	module.RegisterPrimitive("vector-set!", prim_vector_set_bang)
	module.RegisterPrimitive("vector-ref", prim_vector_ref)
	module.RegisterPrimitive("=", prim_eq)
	module.RegisterPrimitive("<=", prim_le)
	module.RegisterPrimitive(">=", prim_ge)
	module.RegisterPrimitive(">", prim_gt)
	module.RegisterPrimitive("<", prim_lt)
	module.RegisterPrimitive("zero?", prim_zero_p)
	module.RegisterPrimitive("number->string", prim_number_to_string)
	module.RegisterPrimitive("string-length", prim_string_length)
	return nil
}

func prim_display(argv []LObject, argc int) (LObject, LError) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func prim_newline(argv []LObject, argc int) (LObject, LError) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func prim_print(argv []LObject, argc int) (LObject, LError) {
	fmt.Printf(red)
	for _, o := range argv {
		fmt.Printf("%v", o)
	}
	fmt.Printf(black)
	return nil, nil
}

func prim_println(argv []LObject, argc int) (LObject, LError) {
	prim_print(argv, argc)
	fmt.Println("")
	return nil, nil
}

func prim_list(argv []LObject, argc int) (LObject, LError) {
	var p LObject
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func prim_quotient(argv []LObject, argc int) (LObject, LError) {
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

func prim_remainder(argv []LObject, argc int) (LObject, LError) {
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

func prim_plus(argv []LObject, argc int) (LObject, LError) {
	if argc == 2 {
		return Add(argv[0], argv[1])
	} else {
		return Sum(argv, argc)
	}
}

func prim_minus(argv []LObject, argc int) (LObject, LError) {
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

func prim_times(argv []LObject, argc int) (LObject, LError) {
	if argc == 2 {
		return Mul(argv[0], argv[1])
	} else {
		return Product(argv, argc)
	}
}

func prim_make_vector(argv []LObject, argc int) (LObject, LError) {
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

func prim_vector_set_bang(argv []LObject, argc int) (LObject, LError) {
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

func prim_vector_ref(argv []LObject, argc int) (LObject, LError) {
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

func prim_ge(argv []LObject, argc int) (LObject, LError) {
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

func prim_le(argv []LObject, argc int) (LObject, LError) {
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

func prim_gt(argv []LObject, argc int) (LObject, LError) {
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

func prim_lt(argv []LObject, argc int) (LObject, LError) {
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

func prim_eq(argv []LObject, argc int) (LObject, LError) {
	if argc == 2 {
		b := Equal(argv[0], argv[1])
		return b, nil
	} else {
		return argcError()
	}
}

func prim_zero_p(argv []LObject, argc int) (LObject, LError) {
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

func prim_number_to_string(argv []LObject, argc int) (LObject, LError) {
	if argc != 1 {
		return argcError()
	}
	if !IsNumber(argv[0]) {
		return nil, Error("Not a number:", argv[0])
	}
	return NewString(argv[0].String()), nil
}

func prim_string_length(argv []LObject, argc int) (LObject, LError) {
	if argc != 1 {
		return argcError()
	}
	if !IsString(argv[0]) {
		return typeError("string", 1)
	}
	i := Length(argv[0])
	return NewInteger(int64(i)), nil
}

func NewSchemePrimitives() Primitives {
	return SchemePrimitives{}
}
