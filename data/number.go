package data

import(
	"math"
	"strconv"
)

type Number struct {
	Value float64
}

func Float(f float64) *Number {
	return &Number{Value: f}
}

func Integer(i int) *Number {
	return &Number{Value: float64(i)}
}

const epsilon = 0.000000001

func NumberEqual(f1 float64, f2 float64) bool {
	if f1 == f2 {
		return true
	}
	if math.Abs(f1-f2) < epsilon {
		return true
	}
	return false
}

func (n *Number) Type() Value {
	return NumberType
}

func (n *Number) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

func (n *Number) Equals(another Value) bool {
	if another != nil {
		if n2, ok := another.(*Number); ok {
			return NumberEqual(n.Value, n2.Value)
		}
	}
	return false
}

func (n *Number) IntValue() int {
	return int(n.Value)
}

func (n *Number) Float64Value() float64 {
	return n.Value
}

func (n *Number) RuneValue() rune {
	return rune(n.Value)
}

