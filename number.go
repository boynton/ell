/*
Copyright 2015 Lee Boynton

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

package ell

import (
	"math"
	"math/rand"
	"strconv"

	. "github.com/boynton/ell/data"
)

// Zero is the Ell 0 value
var Zero = Integer(0)

// One is the Ell 1 value
var One = Integer(1)

// MinusOne is the Ell -1 value
var MinusOne = Integer(-1)

func Int(n int64) *Number {
	return Integer(int(n))
}

// Round - return the closest integer value to the float value
func Round(f float64) float64 {
	if f > 0 {
		return math.Floor(f + 0.5)
	}
	return math.Ceil(f - 0.5)
}

// ToNumber - convert object to a number, if possible
func ToNumber(o Value) (*Number, error) {
	switch p := o.(type) {
	case *Number:
		return p, nil
	case *Character:
		return Integer(int(p.Value)), nil
	case *Boolean:
		if p.Value {
			return One, nil
		}
		return Zero, nil
	case *String:
		f, err := strconv.ParseFloat(p.Value, 64)
		if err == nil {
			return Float(f), nil
		}
	}
	return nil, NewError(ArgumentErrorKey, "cannot convert to an number: ", o)
}

// ToInt - convert the object to an integer number, if possible
func ToInt(o Value) (*Number, error) {
	switch p := o.(type) {
	case *Number:
		return Float(Round(p.Value)), nil
	case *Character:
		return Integer(int(p.Value)), nil
	case *Boolean:
		if p.Value {
			return One, nil
		}
		return Zero, nil
	case *String:
		n, err := strconv.ParseInt(p.Value, 10, 64)
		if err == nil {
			return Integer(int(n)), nil
		}
	}
	return nil, NewError(ArgumentErrorKey, "cannot convert to an integer: ", o)
}

func IsInt(obj Value) bool {
	if p, ok := obj.(*Number); ok {
		f := p.Value
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func IsFloat(obj Value) bool {
	if obj.Type() == NumberType {
		return !IsInt(obj)
	}
	return false
}

func AsFloat64Value(obj Value) (float64, error) {
	if p, ok := obj.(*Number); ok {
		return p.Value, nil
	}
	return 0, NewError(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

func AsInt64Value(obj Value) (int64, error) {
	if p, ok := obj.(*Number); ok {
		return int64(p.Value), nil
	}
	return 0, NewError(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

func AsIntValue(obj Value) (int, error) {
	if p, ok := obj.(*Number); ok {
		return int(p.Value), nil
	}
	return 0, NewError(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

func AsByteValue(obj Value) (byte, error) {
	if p, ok := obj.(*Number); ok {
		return byte(p.Value), nil
	}
	return 0, NewError(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type())
}

var randomGenerator = rand.New(rand.NewSource(1))

func RandomSeed(n int64) {
	randomGenerator = rand.New(rand.NewSource(n))
}

func Random(min float64, max float64) *Number {
	return Float(min + (randomGenerator.Float64() * (max - min)))
}

func RandomList(size int, min float64, max float64) *List {
	//fix this!
	result := EmptyList
	tail := EmptyList
	for i := 0; i < size; i++ {
		tmp := NewList(Random(min, max))
		if result == EmptyList {
			result = tmp
			tail = tmp
		} else {
			tail.Cdr = tmp
			tail = tmp
		}
	}
	return result
}

// IntValue - return native int value of the object
func IntValue(obj Value) int {
	if p, ok := obj.(*Number); ok {
		return int(p.Value)
	}
	return 0
}

// Int64Value - return native int64 value of the object
func Int64Value(obj Value) int64 {
	if p, ok := obj.(*Number); ok {
		return int64(p.Value)
	}
	return 0
}

// Float64Value - return native float64 value of the object
func Float64Value(obj Value) float64 {
	if p, ok := obj.(*Number); ok {
		return p.Value
	}
	return 0
}
