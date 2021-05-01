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
	. "github.com/boynton/ell/data"
)

// ToVector - convert the object to a <vector>, if possible
func ToVector(obj Value) (*Vector, error) {
	switch p := obj.(type) {
	case *Vector:
		return p, nil
	case *List:
		return ListToVector(p), nil
	case *Struct:
		return StructToVector(p), nil
	case *String:
		return StringToVector(p), nil
	}
	return nil, NewError(ArgumentErrorKey, "to-vector expected <vector>, <list>, <struct>, or <string>, got a ", obj.Type())
}

func IsVector(obj Value) bool {
	if _, ok := obj.(*Vector); ok {
		return true
	}
	return false
}
