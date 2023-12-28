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

type Boolean struct {
	Value bool
}

var True *Boolean = &Boolean{Value: true}
var False *Boolean = &Boolean{Value: false}

func (data *Boolean) Type() Value {
	return BooleanType
}

func (data *Boolean) String() string {
	if data.Value {
		return "true"
	}
	return "false"
}

func (b1 *Boolean) Equals(another Value) bool {
	if b2, ok := another.(*Boolean); ok {
		return b1.Value == b2.Value
	}
	return false
}
