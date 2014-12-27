package main

// the primitive functions for the languages
import (
	"fmt"
	. "github.com/boynton/gell"
)

func SchemePrimitives() map[string]Primitive {
	m := make(map[string]Primitive)
	m["display"] = scheme_display
	m["newline"] = scheme_newline
	m["list"] = scheme_list
	m["+"] = scheme_plus
	m["-"] = scheme_minus
	m["*"] = scheme_times
	m["quotient"] = scheme_quotient
	m["remainder"] = scheme_remainder
	m["modulo"] = scheme_remainder //fix!
	m["make-vector"] = scheme_make_vector
	m["vector-set!"] = scheme_vector_set_bang
	m["vector-ref"] = scheme_vector_ref
	m["="] = scheme_eq
	m["<="] = scheme_le
	m[">="] = scheme_ge
	m[">"] = scheme_gt
	m["<"] = scheme_lt
	m["zero?"] = scheme_zero_p
	m["number->string"] = scheme_number_to_string
	m["string-length"] = scheme_string_length
	return m
}

func scheme_display(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("%v", argv[0])
	return nil, nil
}

func scheme_newline(argv []LObject, argc int) (LObject, error) {
	if argc != 0 {
		//todo: add the optional port argument like schema
		return argcError()
	}
	fmt.Printf("\n")
	return nil, nil
}

func scheme_list(argv []LObject, argc int) (LObject, error) {
	var p LObject
	p = NIL
	for i := argc - 1; i >= 0; i-- {
		p = Cons(argv[i], p)
	}
	return p, nil
}

func scheme_quotient(argv []LObject, argc int) (LObject, error) {
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

func scheme_remainder(argv []LObject, argc int) (LObject, error) {
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

func scheme_plus(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Add(argv[0], argv[1])
	} else {
		return Sum(argv, argc)
	}
}

func scheme_minus(argv []LObject, argc int) (LObject, error) {
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

func scheme_times(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		return Mul(argv[0], argv[1])
	} else {
		return Product(argv, argc)
	}
}

func scheme_make_vector(argv []LObject, argc int) (LObject, error) {
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

func scheme_vector_set_bang(argv []LObject, argc int) (LObject, error) {
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

func scheme_vector_ref(argv []LObject, argc int) (LObject, error) {
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

func scheme_ge(argv []LObject, argc int) (LObject, error) {
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

func scheme_le(argv []LObject, argc int) (LObject, error) {
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

func scheme_gt(argv []LObject, argc int) (LObject, error) {
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

func scheme_lt(argv []LObject, argc int) (LObject, error) {
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

func scheme_eq(argv []LObject, argc int) (LObject, error) {
	if argc == 2 {
		b := Equal(argv[0], argv[1])
		return b, nil
	} else {
		return argcError()
	}
}

func scheme_zero_p(argv []LObject, argc int) (LObject, error) {
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

func scheme_number_to_string(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsNumber(argv[0]) {
		return nil, Error("Not a number:", argv[0])
	}
	return NewString(argv[0].String()), nil
}

func scheme_string_length(argv []LObject, argc int) (LObject, error) {
	if argc != 1 {
		return argcError()
	}
	if !IsString(argv[0]) {
		return typeError("string", 1)
	}
	i := Length(argv[0])
	return NewInteger(int64(i)), nil
}
