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

import (
	"fmt"
)

// Instance - a type/data value pair, i.e. `#<point>{x: 23 y: 57}` which is a struct tagged with the <point> type
// The 'type' of an instance is determined by a tag, i.e. it is not a primitive type
type Instance struct {
	TypeTag Value
	Value   Value
}

// this is not a primitive type, it is determined by the tag.
func (data *Instance) Type() Value {
	return data.TypeTag
}

func (data *Instance) String() string {
	return fmt.Sprintf("#%s%v", data.TypeTag, data.Value.String())
}

func (i1 *Instance) Equals(another Value) bool {
	if i2, ok := another.(*Instance); ok {
		if i1.TypeTag != i2.TypeTag {
			return false
		}
		return Equal(i1.Value, i2.Value)
	}
	return false
}

func NewInstance(tag Value, value Value) (Value, error) {
	if !IsType(tag) {
		return nil, NewError(ArgumentErrorKey, TypeType.String(), tag)
	}
	switch tag {
	case NullType, BooleanType, NumberType, SymbolType, KeywordType, StringType, VectorType, StructType, ListType, TypeType:
		return nil, NewError(ArgumentErrorKey, tag, NewString("Cannot tag instance as a builtin type"))
	}
	return &Instance{
		TypeTag: tag,
		Value:   value,
	}, nil
}
