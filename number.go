/*
Copyright 2014 Lee Boynton

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

package main

import (
	"math"
	"strconv"
)

type LNumber float64 // <number>

var typeNumber = intern("<number>")

// Type returns the type of the object
func (LNumber) Type() LAny {
	return typeNumber
}

// Value returns the object itself for primitive types
func (f LNumber) Value() LAny {
	return f
}

func (f LNumber) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f LNumber) Copy() LAny {
	return f
}

func theNumber(obj LAny) (LNumber, error) {
	if n, ok := obj.(LNumber); ok {
		return n, nil
	}
	return 0, TypeError(typeNumber, obj)
}

func isInt(obj LAny) bool {
	if n, ok := obj.(LNumber); ok {
		f := float64(n)
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj LAny) bool {
	return !isInt(obj)
}

func isNumber(obj LAny) bool {
	_, ok := obj.(LNumber)
	return ok
}

func floatValue(obj LAny) (float64, error) {
	switch n := obj.(type) {
	case LNumber:
		return float64(n), nil
	}
	return 0, TypeError(typeNumber, obj)
}

func int64Value(obj LAny) (int64, error) {
	switch n := obj.(type) {
	case LNumber:
		return int64(n), nil
	case LChar:
		return int64(n), nil
	default:
		return 0, TypeError(typeNumber, obj)
	}
}

func intValue(obj LAny) (int, error) {
	switch n := obj.(type) {
	case LNumber:
		return int(n), nil
	case LChar:
		return int(n), nil
	default:
		return 0, TypeError(typeNumber, obj)
	}
}

// Equal returns true if the object is equal to the argument
func greaterOrEqual(n1 LAny, n2 LAny) (LAny, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 >= f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func lessOrEqual(n1 LAny, n2 LAny) (LAny, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 <= f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func greater(n1 LAny, n2 LAny) (LAny, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 > f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func less(n1 LAny, n2 LAny) (LAny, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 < f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func numericallyEqual(o1 LAny, o2 LAny) (bool, error) {
	switch n1 := o1.(type) {
	case LNumber:
		switch n2 := o2.(type) {
		case LNumber:
			return n1 == n2, nil
		default:
			return false, TypeError(typeNumber, o2)
		}
	default:
		return false, TypeError(typeNumber, o1)
	}
}

func identical(n1 LAny, n2 LAny) bool {
	return n1 == n2
}

// Equal returns true if the object is equal to the argument
func (f LNumber) Equal(another LAny) bool {
	if a, err := floatValue(another); err == nil {
		return float64(f) == a
	}
	return false
}

func add(num1 LAny, num2 LAny) (LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return LNumber(n1 + n2), nil
}

func sum(nums []LAny, argc int) (LAny, error) {
	var sum float64
	for _, num := range nums {
		switch n := num.(type) {
		case LNumber:
			sum += float64(n)
		default:
			return nil, TypeError(typeNumber, num)
		}
	}
	return LNumber(sum), nil
}

func sub(num1 LAny, num2 LAny) (LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return LNumber(n1 - n2), nil
}

func minus(nums []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return nil, ArgcError("-", "1+", argc)
	}
	var fsum float64
	num := nums[0]
	switch n := num.(type) {
	case LNumber:
		fsum = float64(n)
	default:
		return nil, TypeError(typeNumber, num)
	}
	if argc == 1 {
		fsum = -fsum
	} else {
		for _, num := range nums[1:] {
			switch n := num.(type) {
			case LNumber:
				fsum -= float64(n)
			default:
				return nil, TypeError(typeNumber, num)
			}
		}
	}
	return LNumber(fsum), nil
}

func mul(num1 LAny, num2 LAny) (LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return LNumber(n1 * n2), nil
}

func product(argv []LAny, argc int) (LAny, error) {
	prod := 1.0
	for _, num := range argv {
		switch n := num.(type) {
		case LNumber:
			prod *= float64(n)
		default:
			return nil, TypeError(typeNumber, num)
		}
	}
	return LNumber(prod), nil
}

func div(argv []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return nil, ArgcError("/", "1+", argc)
	} else if argc == 1 {
		n1, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		return LNumber(1.0 / n1), nil
	} else {
		quo, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		for i := 1; i < argc; i++ {
			n, err := floatValue(argv[i])
			if err != nil {
				return nil, err
			}
			quo /= n
		}
		return LNumber(quo), nil
	}
}

