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

// Commonly used constants
var Zero = newInt(0)
var One = newInt(1)

func newFloat64(f float64) *LOB {
	num := newLOB(typeNumber)
	num.fval = f
	return num
}

func newInt64(i int64) *LOB {
	num := newLOB(typeNumber)
	num.fval = float64(i)
	return num
}

func newInt(i int) *LOB {
	num := newLOB(typeNumber)
	num.fval = float64(i)
	return num
}

func isInt(obj *LOB) bool {
	if obj.variant == typeNumber {
		f := obj.fval
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj *LOB) bool {
	if obj.variant == typeNumber {
		return !isInt(obj)
	}
	return false
}

func floatValue(obj *LOB) (float64, error) {
	if obj.variant == typeNumber {
		return obj.fval, nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.variant)
}

func int64Value(obj *LOB) (int64, error) {
	if obj.variant == typeNumber {
		return int64(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.variant)
}

func intValue(obj *LOB) (int, error) {
	if obj.variant == typeNumber {
		return int(obj.fval), nil
	}
	return 0, Error(ArgumentErrorKey, "Expected a <number>, got a ", obj.variant)
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
