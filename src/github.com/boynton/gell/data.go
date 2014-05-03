package gell

import (
	"bytes"
	"fmt"
	"strconv"
)

type Type int

const (
	NULL_TYPE Type = iota
	BOOLEAN_TYPE
	NUMBER_TYPE
	STRING_TYPE
	SYMBOL_TYPE
	KEYWORD_TYPE
	PAIR_TYPE
	VECTOR_TYPE
	MAP_TYPE
	FUNCTION_TYPE
	ERROR_TYPE
)

func (tag Type) String() string {
	switch tag {
	case NULL_TYPE:
		return "null"
	case BOOLEAN_TYPE:
		return "boolean"
	case NUMBER_TYPE:
		return "number"
	case STRING_TYPE:
		return "string"
	case SYMBOL_TYPE:
		return "symbol"
	case PAIR_TYPE:
		return "pair"
	default:
		return "<?>"
	}
}

type Object interface {
	Type() Type
	String() string
}

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}
func (e Error) Type() Type {
	return ERROR_TYPE
}
func (e Error) String() string {
	return fmt.Sprintf("<Error: %s>", e.msg)
}

type Boolean bool

func (Boolean) Type() Type {
	return BOOLEAN_TYPE
}
func (b Boolean) String() string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

var TRUE = Boolean(true)
var FALSE = Boolean(false)

//Char?

type Null struct{}

func (Null) Type() Type {
	return NULL_TYPE
}
func (b Null) String() string {
	return "()"
}

var null = Null{}

func NullP(o Object) bool {
	return o == null
}
func NULL() Object {
	return null
}

type Number float64

func NumberP(o Object) bool {
	return o.Type() == NUMBER_TYPE
}
func (Number) Type() Type {
	return NUMBER_TYPE
}
func (n Number) String() string {
	return strconv.FormatFloat(float64(n), 'f', -1, 64)
}

type String string

func (String) Type() Type {
	return STRING_TYPE
}
func (s String) String() string {
	return fmt.Sprintf("\"%s\"", string(s))
}

type Symbol struct {
	Name string
}

func (v Symbol) Type() Type {
	return SYMBOL_TYPE
}
func (v Symbol) String() string {
	return v.Name
}

type Namespace map[string]Symbol

func Intern(ns Namespace, name string) Symbol {
	sym, ok := ns[name]
	if !ok {
		sym = Symbol{name}
		ns[name] = sym
	}
	return sym
}

type Pair struct {
	car Object
	cdr Object
}

func (Pair) Type() Type {
	return PAIR_TYPE
}
func (p Pair) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(p.car.String())
	tail := p.cdr
	b := true
	for b {
		switch tail.(type) {
		case Pair:
			p = tail.(Pair)
			tail = p.cdr
			buf.WriteString(" ")
			buf.WriteString(p.car.String())
		case Null:
			b = false
		default:
			buf.WriteString(" . ")
			buf.WriteString(tail.String())
			b = false
		}
	}
	buf.WriteString(")")
	return buf.String()
}

type Vector []Object

func (Vector) Type() Type {
	return VECTOR_TYPE
}

type Map map[Object]Object

func (Map) DataType() Type {
	return MAP_TYPE
}

func Cons(car Object, cdr Object) Pair {
	return Pair{car, cdr}
}
