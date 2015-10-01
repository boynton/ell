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

// NumberType - the Type object for this kind of value
var NumberType = intern("<number>")

// LNumber - the concrete type for numbers
type LNumber struct { // <number>
	value float64
}

func (n *LNumber) String() string {
	return strconv.FormatFloat(n.value, 'f', -1, 64)
}

// Type returns the type of the object
func (n *LNumber) Type() LOB {
	return NumberType
}

// Value returns the object itself for primitive types
func (n *LNumber) Value() LOB {
	return n
}

// Equal returns true if the object is equal to the argument
func (n *LNumber) Equal(another LOB) bool {
	n2, ok := another.(*LNumber)
	if ok {
		return numberEqual(n.value, n2.value)
	}
	return false
}

func isNumber(obj LOB) bool {
	return obj.Type() == NumberType
}

// A commonly used constant
var One = newInt(1)

func newFloat64(f float64) *LNumber {
	return &LNumber{f}
}

func newInt64(i int64) *LNumber {
	return &LNumber{float64(i)}
}

func newInt(i int) *LNumber {
	return &LNumber{float64(i)}
}

func isInt(obj LOB) bool {
	n, ok := obj.(*LNumber)
	if ok {
		f := n.value
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj LOB) bool {
	n, ok := obj.(*LNumber)
	if ok {
		f := n.value
		if math.Trunc(f) != f {
			return true
		}
	}
	return false
}

func floatValue(obj LOB) (float64, error) {
	n, ok := obj.(*LNumber)
	if ok {
		return n.value, nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

func int64Value(obj LOB) (int64, error) {
	n, ok := obj.(*LNumber)
	if ok {
		return int64(n.value), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

func intValue(obj LOB) (int, error) {
	n, ok := obj.(*LNumber)
	if ok {
		return int(n.value), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

// Equal returns true if the object is equal to the argument
func greaterOrEqual(n1 LOB, n2 LOB) (LOB, error) {
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

func lessOrEqual(n1 LOB, n2 LOB) (LOB, error) {
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

func greater(n1 LOB, n2 LOB) (LOB, error) {
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

func less(n1 LOB, n2 LOB) (LOB, error) {
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

func numericallyEqual(n1 LOB, n2 LOB) (LOB, error) {
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

func inc(num LOB) (LOB, error) {
	n, err := floatValue(num)
	if err != nil {
		return nil, err
	}
	return newFloat64(n + 1), nil
}

func dec(num LOB) (LOB, error) {
	n, err := floatValue(num)
	if err != nil {
		return nil, err
	}
	return newFloat64(n - 1), nil
}

func add(num1 LOB, num2 LOB) (LOB, error) {
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

func sum(nums []LOB, argc int) (LOB, error) {
	var sum float64
	for i, num := range nums {
		n, ok := num.(*LNumber)
		if !ok {
			return nil, Error(ArgumentErrorKey, "+ expected a <number> for argument ", i+1, ", got a ", num.Type())
		}
		sum += float64(n.value)
	}
	return newFloat64(sum), nil
}

func sub(num1 LOB, num2 LOB) (LOB, error) {
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

func minus(nums []LOB, argc int) (LOB, error) {
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "- expected at least one argument")
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

func mul(num1 LOB, num2 LOB) (LOB, error) {
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

func product(argv []LOB, argc int) (LOB, error) {
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

func div(argv []LOB, argc int) (LOB, error) {
	if argc < 1 {
		return nil, Error(ArgumentErrorKey, "/ expected at least one argument")
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
