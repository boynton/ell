package data

import(
	"fmt"
)

type String struct {
	Value string
}

func NewString(s string) *String {
	return &String{Value: s}
}

var EmptyString Value = NewString("")

func (s *String) Type() Value {
	return StringType
}

func (s *String) String() string {
	return fmt.Sprintf("%s", s.Value)
}

func (s *String) Equals(another Value) bool {
	if another != nil {
		if s2, ok := another.(*String); ok {
			return s.Value == s2.Value
		}
	}
	return false
}

type Character struct {
	Value rune
}

func NewCharacter(c rune) *Character {
	return &Character{Value: c}
}

func (c *Character) Type() Value {
	return CharacterType
}
func (c *Character) String() string {
	return string([]rune{c.Value})
}
func (c *Character) Equals(another Value) bool {
	if another != nil {
		if p, ok := another.(*Character); ok {
			return c.Value == p.Value
		}
	}
	return false
}

