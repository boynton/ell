/*
Copyright 2021 Lee Boynton

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
package data

type Value interface {
	Type() Value
	String() string
	Equals(another Value) bool
}

func Equal(o1 Value, o2 Value) bool {
	if o1 == o2 {
		return true
	}
	if o1 == nil || o2 == nil {
		return false
	}
	return o1.Equals(o2)
}

var Null Value = &NullValue{}

type NullValue struct {
}

func (v *NullValue) Type() Value {
	return NullType
}
func (v *NullValue) String() string {
	return "null"
}
func (v *NullValue) Equals(another Value) bool {
	return another == Null
}
