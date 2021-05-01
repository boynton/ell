package data

import(
	"fmt"
)

// Instance - a type/data value pair, i.e. `#<point>{x: 23 y: 57}` which is a struct tagged with the <point> type
// The 'type' of an instance is determined by a tag, i.e. it is not a primitive type
type Instance struct {
	TypeTag Value
	Value Value
}

//this is not a primitive type, it is determined by the tag.
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
		Value: value,
	}, nil
}
