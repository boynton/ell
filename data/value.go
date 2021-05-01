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
