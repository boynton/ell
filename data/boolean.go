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

