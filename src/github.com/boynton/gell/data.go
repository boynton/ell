/*
Copyright 2014 Lee Boynton

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

package gell

import (
	"bytes"
	"fmt"
	"strconv"
)

//
// The generic Ell object, which can be queried for its symbolic type name at runtime
//
type LObject interface {
	Type() LSymbol
	String() string
}

//
// ------------------- nil
//

const NIL = lnil(0)

type lnil int

var symNil = newSymbol("nil")

func IsNil(obj LObject) bool {
	return obj == NIL
}

func (lnil) Type() LSymbol {
	return symNil
}

func (lnil) String() string {
	return "nil"
}

//
// ------------------- EOI (end-of-information) marker
//

const EOI = leoi(0)

type leoi int

var symEoi = newSymbol("eoi")

func IsEOI(obj LObject) bool {
	return obj == EOI
}

func (leoi) Type() LSymbol {
	return symEoi
}

func (leoi) String() string {
	return "<end-of-input>"
}

//
// ------------------- boolean
//

type LBoolean interface {
	Type() LSymbol
	String() string
}

const TRUE lboolean = lboolean(true)
const FALSE lboolean = lboolean(false)

type lboolean bool

var symBoolean = newSymbol("boolean")

func IsBoolean(obj LObject) bool {
	_, ok := obj.(lboolean)
	return ok
}

func (lboolean) Type() LSymbol {
	return symBoolean
}

func (b lboolean) String() string {
	return strconv.FormatBool(bool(b))
}

//
// ------------------- symbol
//

type LSymbol interface {
	Type() LSymbol
	String() string
}

type lsymbol struct {
	Name string
	tag  int
}

var symtag int

func newSymbol(name string) *lsymbol {
	sym := lsymbol{name, symtag}
	symtag++
	return &sym
}

var symSymbol = newSymbol("symbol")

func IsSymbol(obj LObject) bool {
	_, ok := obj.(*lsymbol)
	return ok
}

func (*lsymbol) Type() LSymbol {
	return symSymbol
}

func (sym *lsymbol) String() string {
	return sym.Name
}

/*
func Global(sym LSymbol) LObject {
	return (sym.(*lsymbol)).value
}

func SetGlobal(sym LSymbol, val LObject) {
	(sym.(*lsymbol)).value = val
}
*/

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]*lsymbol{
	"nil":     symNil,
	"boolean": symBoolean,
	"symbol":  symSymbol,
	"keyword": symKeyword,
	"string":  symString,
	"number":  symNumber,
	"list":    symList,
	"vector":  symVector,
	"map":     symMap,
	"eoi":     symEoi,
}

func Symbols() []LSymbol {
	syms := make([]LSymbol, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func Intern(name string) LObject {
	//to do: validate the symbol name, based on EllDN spec
	v, ok := symtab[name]
	if !ok {
		v = newSymbol(name)
		symtab[name] = v
	}
	return v
}

//
// ------------------- keyword
//

type LKeyword interface {
	Type() LSymbol
	String() string
	Symbol() LSymbol
}

type lkeyword struct {
	sym *lsymbol
}

var symKeyword = newSymbol("keyword")

func IsKeyword(obj LObject) bool {
	_, ok := obj.(lkeyword)
	return ok
}

func (lkeyword) Type() LSymbol {
	return symKeyword
}

func (key lkeyword) String() string {
	return key.sym.Name + ":"
}

//
// ------------------- string
//

type LString interface {
	Type() LSymbol
	String() string
}

type lstring string

var symString = newSymbol("string")

func NewString(val string) LString {
	s := lstring(val)
	return s
}

func IsString(obj LObject) bool {
	_, ok := obj.(lstring)
	return ok
}

func (lstring) Type() LSymbol {
	return symString
}

func encodeString(s string) string {
	buf := []byte{}
	buf = append(buf, '"')
	for _, c := range s {
		switch c {
		case '\\':
			buf = append(buf, '\\')
			buf = append(buf, '\\')
		case '\n':
			buf = append(buf, '\\')
			buf = append(buf, 'n')
		case '\t':
			buf = append(buf, '\\')
			buf = append(buf, 't')
		case '\f':
			buf = append(buf, '\\')
			buf = append(buf, 'f')
		case '\b':
			buf = append(buf, '\\')
			buf = append(buf, 'b')
		case '\r':
			buf = append(buf, '\\')
			buf = append(buf, 'r')
			//to do: handle non-byte unicode by encoding as "\uhhhh"
		default:
			buf = append(buf, byte(c))
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

func (s lstring) EncodedString() string {
	return encodeString(string(s))
}

func (s lstring) String() string {
	//return encodeString(string(s))
	return string(s)
}

//
// ------------------- number
//

type LNumber interface {
	Type() LSymbol
	String() string
	IntegerValue() int64
	RealValue() float64
}

var symNumber = newSymbol("number")

func IsNumber(obj LObject) bool {
	_, ok := obj.(linteger)
	if ok {
		return true
	}
	_, ok = obj.(lreal)
	return ok
}

func NewInteger(n int64) LNumber {
	v := linteger(n)
	return v
}

func NewReal(n float64) LNumber {
	v := lreal(n)
	return v
}

func RealValue(obj LObject) (float64, error) {
	switch n := obj.(type) {
	case linteger:
		return float64(n), nil
	case lreal:
		return float64(n), nil
	}
	return 0, Error("Not a real number:", obj)
}

func IntegerValue(obj LObject) (int64, error) {
	switch n := obj.(type) {
	case linteger:
		return int64(n), nil
	case lreal:
		return int64(n), nil
	default:
		return 0, Error("Not an integer:", obj)
	}
}

func IntValue(obj LObject) (int, error) {
	switch n := obj.(type) {
	case linteger:
		return int(n), nil
	case lreal:
		return int(n), nil
	default:
		return 0, Error("Not an integer:", obj)
	}
}

func GreaterOrEqual(n1 LObject, n2 LObject) (LObject, error) {
	f1, err := RealValue(n1)
	if err == nil {
		f2, err := RealValue(n2)
		if err == nil {
			if f1 >= f2 {
				return TRUE, nil
			} else {
				return FALSE, nil
			}
		}
		return nil, err
	}
	return nil, err
}

func LessOrEqual(n1 LObject, n2 LObject) (LObject, error) {
	f1, err := RealValue(n1)
	if err == nil {
		f2, err := RealValue(n2)
		if err == nil {
			if f1 <= f2 {
				return TRUE, nil
			} else {
				return FALSE, nil
			}
		}
		return nil, err
	}
	return nil, err
}

func Greater(n1 LObject, n2 LObject) (LObject, error) {
	f1, err := RealValue(n1)
	if err == nil {
		f2, err := RealValue(n2)
		if err == nil {
			if f1 > f2 {
				return TRUE, nil
			} else {
				return FALSE, nil
			}
		}
		return nil, err
	}
	return nil, err
}

func Less(n1 LObject, n2 LObject) (LObject, error) {
	f1, err := RealValue(n1)
	if err == nil {
		f2, err := RealValue(n2)
		if err == nil {
			if f1 < f2 {
				return TRUE, nil
			} else {
				return FALSE, nil
			}
		}
		return nil, err
	}
	return nil, err
}

func Equal(n1 LObject, n2 LObject) LObject {
	if n1 == n2 {
		return TRUE
	} else {
		return FALSE
	}
}

type linteger int64

func (linteger) Type() LSymbol {
	return symNumber
}

func (i linteger) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i linteger) IntegerValue() int64 {
	return int64(i)
}
func (i linteger) RealValue() float64 {
	return float64(i)
}

type lreal float64

func (lreal) Type() LSymbol {
	return symNumber
}

func (f lreal) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f lreal) IntegerValue() int64 {
	return int64(f)
}

func (f lreal) RealValue() float64 {
	return float64(f)
}

func Add(num1 LObject, num2 LObject) (LObject, error) {
	n1, err := RealValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := RealValue(num2)
	if err != nil {
		return nil, err
	}
	return NewReal(n1 + n2), nil
}

func Sum(nums []LObject, argc int) (LObject, error) {
	var isum int64
	var fsum float64
	integral := true
	isum = 0
	for _, num := range nums {
		switch n := num.(type) {
		case linteger:
			isum += int64(n)
		case lreal:
			if integral {
				fsum = float64(isum)
				integral = false
			}
			fsum += float64(n)
		default:
			return nil, Error("Not a number", num)
		}
	}
	if integral {
		return linteger(isum), nil
	} else {
		return lreal(fsum), nil
	}
}

func Mul(num1 LObject, num2 LObject) (LObject, error) {
	n1, err := RealValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := RealValue(num2)
	if err != nil {
		return nil, err
	}
	return NewReal(n1 * n2), nil
}

func Product(argv []LObject, argc int) (LObject, error) {
	var iprod int64
	var fprod float64
	integral := true
	iprod = 1
	for _, num := range argv {
		switch n := num.(type) {
		case linteger:
			iprod = iprod * int64(n)
		case lreal:
			if integral {
				fprod = float64(iprod)
				integral = false
			}
			fprod *= float64(n)
		default:
			return nil, Error("Not a number", num)
		}
	}
	if integral {
		return linteger(iprod), nil
	} else {
		return lreal(fprod), nil
	}
}

//
// ------------------- list
//
type LList interface {
	Type() LSymbol
	String() string
	Length() int
	Car() LObject
	Cdr() LObject
	//append
	//reverse
}

type llist struct {
	car LObject
	cdr LObject
}

var symList = newSymbol("list")

func IsList(obj LObject) bool {
	//	return obj.Type() == symList
	_, ok := obj.(*llist)
	return ok
}
func (*llist) Type() LSymbol {
	return symList
}

func (lst *llist) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(lst.car.String())
	var tail LObject = lst.cdr
	b := true
	for b {
		if tail == NIL {
			b = false
		} else if IsList(tail) {
			lst = tail.(*llist)
			tail = lst.cdr
			buf.WriteString(" ")
			buf.WriteString(lst.car.String())
		} else {
			buf.WriteString(" . ")
			buf.WriteString(tail.String())
			b = false
		}
	}
	buf.WriteString(")")
	return buf.String()
}

func (lst *llist) Length() int {
	count := 1
	var o LObject = lst.cdr
	for o != NIL {
		if !IsList(o) {
			return -1 //not a proper list
		}
		count++
		o = o.(*llist).cdr
	}
	return count
}

func Cons(car LObject, cdr LObject) LObject {
	lst := llist{car, cdr}
	return &lst
}

func Car(lst LObject) LObject {
	if IsList(lst) {
		return lst.(*llist).car
	}
	return NIL
}

func Caar(lst LObject) LObject {
	return Car(Car(lst))
}
func Cadr(lst LObject) LObject {
	return Car(Cdr(lst))
}
func Cddr(lst LObject) LObject {
	return Cdr(Cdr(lst))
}
func Caddr(lst LObject) LObject {
	return Car(Cdr(Cdr(lst)))
}
func Cdddr(lst LObject) LObject {
	return Cdr(Cdr(Cdr(lst)))
}
func Cadddr(lst LObject) LObject {
	return Car(Cdr(Cdr(Cdr(lst))))
}
func Cddddr(lst LObject) LObject {
	return Cdr(Cdr(Cdr(Cdr(lst))))
}

func Cdr(lst LObject) LObject {
	if IsList(lst) {
		return lst.(*llist).cdr
	}
	return NIL
}

func ToList(vec []LObject) LObject {
	var p LObject
	p = NIL
	for i := len(vec) - 1; i >= 0; i-- {
		v := vec[i]
		p = Cons(v, p)
	}
	return p
}

func List(vec ...LObject) LObject {
	return ToList(vec)
}

func Length(seq LObject) int {
	if seq == NIL {
		return 0
	} else if IsString(seq) {
		return len(seq.(lstring))
	} else if IsVector(seq) {
		return seq.(*lvector).Length()
	} else if IsList(seq) {
		return seq.(*llist).Length()
	} else {
		return -1
	}
}

func Reverse(lst LObject) (LObject, error) {
	var rev LObject
	rev = NIL
	for lst != NIL {
		switch v := lst.(type) {
		case *llist:
			rev = Cons(v.car, rev)
			lst = v.cdr
		default:
			return nil, Error("Not a proper list:", lst)
		}
	}
	return rev, nil
}

func Concat(seq1 LObject, seq2 LObject) (LObject, error) {
	rev, err := Reverse(seq1)
	if err != nil {
		return nil, err
	}
	for rev != NIL {
		switch v := rev.(type) {
		case *llist:
			seq2 = Cons(v.car, seq2)
			rev = v.cdr
		}
	}
	return seq2, nil

}

//
// ------------------- vector
//

type LVector interface {
	Type() LSymbol
	String() string
	Length() int
	Set(idx int, obj LObject)
	Ref(idx int) LObject
}

type lvector struct {
	elements []LObject
}

func NewVector(size int, init LObject) LVector {
	elements := make([]LObject, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	vec := lvector{elements}
	return &vec
}

func Vector(elements ...LObject) LVector {
	vec := lvector{elements}
	return &vec
}

var symVector = newSymbol("vector")

func IsVector(obj LObject) bool {
	_, ok := obj.(*lvector)
	return ok
	//	return obj.Type() == symVector
}

func (*lvector) Type() LSymbol {
	return symVector
}
func (vec *lvector) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	count := len(vec.elements)
	if count > 0 {
		buf.WriteString(vec.elements[0].String())
		for i := 1; i < count; i++ {
			buf.WriteString(" ")
			buf.WriteString(vec.elements[i].String())
		}
	}
	buf.WriteString("]")
	return buf.String()
}
func (vec *lvector) Length() int {
	return len(vec.elements)
}

func (vec *lvector) Ref(idx int) LObject {
	return vec.elements[idx]
}

func (vec *lvector) Set(idx int, obj LObject) {
	vec.elements[idx] = obj
}

func Ref(vec LObject, idx int) LObject {
	if IsVector(vec) {
		return vec.(*lvector).Ref(idx)
	}
	return NIL
}

func theVector(obj LObject) (*lvector, bool) {
	vec, ok := obj.(*lvector)
	return vec, ok
}

func VectorSet(vec LObject, idx int, obj LObject) error {
	if v, ok := theVector(vec); ok {
		v.elements[idx] = obj
		return nil
	}
	return Error("Not a vector:", vec)
}

func VectorRef(vec LObject, idx int) (LObject, error) {
	if v, ok := theVector(vec); ok {
		return v.elements[idx], nil //maybe should range check the index
	}
	return nil, Error("Not a vector:", vec)
}

//
// ------------------- map
//
type LMap interface {
	Type() LSymbol
	String() string
	Length() int
	Has(key LObject) bool
	Get(key LObject) LObject
	Put(key LObject, value LObject) LMap
}

type tMap struct {
	bindings map[LObject]LObject
}

var symMap = newSymbol("map")

//func IsMap(obj LObject) bool {
//	return obj.Type() == symMap
//}
func (tMap) Type() LSymbol {
	return symMap
}
func (m tMap) Length() int {
	return len(m.bindings)
}
func (m tMap) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	for k, v := range m.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(k.String())
		buf.WriteString(" ")
		buf.WriteString(v.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func Map(pairwiseBindings ...LObject) LMap {
	count := len(pairwiseBindings)
	Println("map pairwise count is", count)
	m := tMap{map[LObject]LObject{}}
	for i := 0; i < count; i += 2 {
		m.Put(pairwiseBindings[i], pairwiseBindings[i+1])
	}
	return m
}

func (m tMap) Put(key LObject, value LObject) LMap {
	m.bindings[key] = value
	return m
}

func (m tMap) Get(key LObject) LObject {
	if val, ok := m.bindings[key]; ok {
		return val
	} else {
		return NIL
	}
}

func (m tMap) Has(key LObject) bool {
	_, ok := m.bindings[key]
	return ok
}

//
// ------------------- error
//

func Error(arg1 interface{}, args ...interface{}) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%v", arg1))
	for _, o := range args {
		buf.WriteString(fmt.Sprintf(" %v", o))
	}
	err := lerror{buf.String()}
	return &err
}

type lerror struct {
	msg string
}

func (e *lerror) Error() string {
	return e.msg
}

func (e *lerror) String() string {
	return fmt.Sprintf("<Error: %s>", e.msg)
}
