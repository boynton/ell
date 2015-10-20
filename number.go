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

package ell

import (
	"math"
	"math/rand"
	"strconv"
)

// Zero is the Ell 0 value
var Zero = Number(0)

// One is the Ell 1 value
var One = Number(1)

// MinusOne is the Ell -1 value
var MinusOne = Number(-1)

func Number(f float64) *LOB {
	num := new(LOB)
	num.Type = NumberType
	num.fval = f
	return num
}

func Round(f float64) float64 {
	if f > 0 {
		return math.Floor(f + 0.5)
	}
	return math.Ceil(f - 0.5)
}

func ToNumber(o *LOB) (*LOB, error) {
	switch o.Type {
	case NumberType:
		return o, nil
	case CharacterType:
		return Number(o.fval), nil
	case BooleanType:
		return Number(o.fval), nil
	case StringType:
		f, err := strconv.ParseFloat(o.text, 64)
		if err == nil {
			return Number(f), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "cannot convert to an number: ", o)
}

func ToInt(o *LOB) (*LOB, error) {
	switch o.Type {
	case NumberType:
		return Number(Round(o.fval)), nil
	case CharacterType:
		return Number(o.fval), nil
	case BooleanType:
		return Number(o.fval), nil
	case StringType:
		n, err := strconv.ParseInt(o.text, 10, 64)
		if err == nil {
			return Number(float64(n)), nil
		}
	}
	return nil, Error(ArgumentErrorKey, "cannot convert to an integer: ", o)
}

func IsInt(obj *LOB) bool {
	if obj.Type == NumberType {
		f := obj.fval
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func IsFloat(obj *LOB) bool {
	if obj.Type == NumberType {
		return !IsInt(obj)
	}
	return false
}

func AsFloat64Value(obj *LOB) (float64, error) {
	if obj.Type == NumberType {
		return obj.fval, nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

func AsInt64Value(obj *LOB) (int64, error) {
	if obj.Type == NumberType {
		return int64(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

func AsIntValue(obj *LOB) (int, error) {
	if obj.Type == NumberType {
		return int(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

func AsByteValue(obj *LOB) (byte, error) {
	if obj.Type == NumberType {
		return byte(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.Type)
}

const epsilon = 0.000000001

// Equal returns true if the object is equal to the argument, within epsilon
func NumberEqual(f1 float64, f2 float64) bool {
	if f1 == f2 {
		return true
	}
	if math.Abs(f1-f2) < epsilon {
		return true
	}
	return false
}

var randomGenerator = rand.New(rand.NewSource(1))

func RandomSeed(n int64) {
	randomGenerator = rand.New(rand.NewSource(n))
}

func Random(min float64, max float64) *LOB {
	return Number(min + (randomGenerator.Float64() * (max - min)))
}

func RandomList(size int, min float64, max float64) *LOB {
	result := EmptyList
	tail := EmptyList
	for i := 0; i < size; i++ {
		tmp := List(Random(min, max))
		if result == EmptyList {
			result = tmp
			tail = tmp
		} else {
			tail.cdr = tmp
			tail = tmp
		}
	}
	return result
}
