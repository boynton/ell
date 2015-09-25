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
)

func newFloat64(f float64) *LAny {
	num := new(LAny)
	num.ltype = typeNumber
	num.fval = f
	return num
}

func newInt64(i int64) *LAny {
	num := new(LAny)
	num.ltype = typeNumber
	num.fval = float64(i)
	return num
}

func newInt(i int) *LAny {
	num := new(LAny)
	num.ltype = typeNumber
	num.fval = float64(i)
	return num
}

func isInt(obj *LAny) bool {
	if obj.ltype == typeNumber {
		f := obj.fval
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj *LAny) bool {
	if obj.ltype == typeNumber {
		return !isInt(obj)
	}
	return false
}

func floatValue(obj *LAny) (float64, error) {
	if obj.ltype == typeNumber {
		return obj.fval, nil
	}
	return 0, TypeError(typeNumber, obj)
}

func int64Value(obj *LAny) (int64, error) {
	if obj.ltype == typeNumber {
		return int64(obj.fval), nil
	}
	return 0, TypeError(typeNumber, obj)
}

func intValue(obj *LAny) (int, error) {
	if obj.ltype == typeNumber {
		return int(obj.fval), nil
	}
	return 0, TypeError(typeNumber, obj)
}

// Equal returns true if the object is equal to the argument
func greaterOrEqual(n1 *LAny, n2 *LAny) (*LAny, error) {
	f1, err := floatValue(n1)
	if err != nil {
		return nil, err
	}
	f2, err := floatValue(n2)
	if err != nil {
		return nil, err
	}
	if f1 >= f2 {
		return True, nil
	}
	return False, nil
}

func lessOrEqual(n1 *LAny, n2 *LAny) (*LAny, error) {
	f1, err := floatValue(n1)
	if err != nil {
		return nil, err
	}
	f2, err := floatValue(n2)
	if err != nil {
		return nil, err
	}
	if f1 <= f2 {
		return True, nil
	}
	return False, nil
}

func greater(n1 *LAny, n2 *LAny) (*LAny, error) {
	f1, err := floatValue(n1)
	if err != nil {
		return nil, err
	}
	f2, err := floatValue(n2)
	if err != nil {
		return nil, err
	}
	if f1 > f2 {
		return True, nil
	}
	return False, nil
}

func less(n1 *LAny, n2 *LAny) (*LAny, error) {
	f1, err := floatValue(n1)
	if err != nil {
		return nil, err
	}
	f2, err := floatValue(n2)
	if err != nil {
		return nil, err
	}
	if f1 < f2 {
		return True, nil
	}
	return False, nil
}

const epsilon = 0.000000001

// Equal returns true if the object is equal to the argument, within epsilon
func numberEqual(f1 float64, f2 float64) bool {
	if f1 == f2 {
		return true
	}
	if math.Abs(f1-f2) < epsilon {
		return true
	}
	return false
}

func numericallyEqual(n1 *LAny, n2 *LAny) (*LAny, error) {
	f1, err := floatValue(n1)
	if err != nil {
		return nil, err
	}
	f2, err := floatValue(n2)
	if err != nil {
		return nil, err
	}
	if numberEqual(f1, f2) {
		return True, nil
	}
	return False, nil
}

func add(num1 *LAny, num2 *LAny) (*LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return newFloat64(n1 + n2), nil
}

func sum(nums []*LAny, argc int) (*LAny, error) {
	var sum float64
	for _, num := range nums {
		if !isNumber(num) {
			return nil, TypeError(typeNumber, num)
		}
		sum += float64(num.fval)
	}
	return newFloat64(sum), nil
}

func sub(num1 *LAny, num2 *LAny) (*LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return newFloat64(n1 - n2), nil
}

func minus(nums []*LAny, argc int) (*LAny, error) {
	if argc < 1 {
		return nil, ArgcError("-", "1+", argc)
	}
	fsum, err := floatValue(nums[0])
	if err != nil {
		return nil, err
	}
	if argc == 1 {
		fsum = -fsum
	} else {
		for _, num := range nums[1:] {
			f, err := floatValue(num)
			if err != nil {
				return nil, err
			}
			fsum -= f
		}
	}
	return newFloat64(fsum), nil
}

func mul(num1 *LAny, num2 *LAny) (*LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return newFloat64(n1 * n2), nil
}

func product(argv []*LAny, argc int) (*LAny, error) {
	prod := 1.0
	for _, num := range argv {
		f, err := floatValue(num)
		if err != nil {
			return nil, err
		}
		prod *= f
	}
	return newFloat64(prod), nil
}

func div(argv []*LAny, argc int) (*LAny, error) {
	if argc < 1 {
		return nil, ArgcError("/", "1+", argc)
	} else if argc == 1 {
		n1, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		return newFloat64(1.0 / n1), nil
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
		return newFloat64(quo), nil
	}
}
