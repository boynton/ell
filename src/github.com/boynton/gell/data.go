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

// for both ell and non-ell values
type LAny interface{}

//
// The generic Ell object, which can be queried for its symbolic type name at runtime
//
type LObject interface {
	Type() LObject
	String() string
	Equal(another LObject) bool
}

//
// ------------------- nil
//

const NIL = lnil(0)

type lnil int

var symNull = newSymbol("null")

func (lnil) Type() LObject {
	return symNull
}

func (lnil) Equal(another LObject) bool {
	return another == NIL
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

func (leoi) Type() LObject {
	return symEoi
}

func (leoi) Equal(another LObject) bool {
	return another == EOI
}

func (leoi) String() string {
	return "<end-of-input>"
}

//
// ------------------- boolean
//

const TRUE lboolean = lboolean(true)
const FALSE lboolean = lboolean(false)

type lboolean bool

var symBoolean = newSymbol("boolean")

func IsBoolean(obj LObject) bool {
	_, ok := obj.(lboolean)
	return ok
}

func (lboolean) Type() LObject {
	return symBoolean
}

func (b lboolean) Equal(another LObject) bool {
	if a, ok := another.(lboolean); ok {
		return b == a
	}
	return false
}

func (b lboolean) String() string {
	return strconv.FormatBool(bool(b))
}

//
// ------------------- symbol
//

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

func (*lsymbol) Type() LObject {
	return symSymbol
}

func (sym *lsymbol) Equal(another LObject) bool {
	if a, ok := another.(*lsymbol); ok {
		return sym == a
	}
	return false
}

func (sym *lsymbol) String() string {
	return sym.Name
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]*lsymbol{
	"null":    symNull, //the type of NIL
	"boolean": symBoolean,
	"symbol":  symSymbol,
	"string":  symString,
	"number":  symNumber,
	"pair":    symPair,
	"vector":  symVector,
	"map":     symMap,
	"eoi":     symEoi, //End Of Information
}

func Symbols() []LObject {
	syms := make([]LObject, 0, len(symtab))
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
// ------------------- string
//

type lstring string

var symString = newSymbol("string")

func NewString(val string) LObject {
	s := lstring(val)
	return s
}

func IsString(obj LObject) bool {
	_, ok := obj.(lstring)
	return ok
}

func (lstring) Type() LObject {
	return symString
}

func (s lstring) Equal(another LObject) bool {
	if a, ok := another.(lstring); ok {
		return s == a
	}
	return false
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

var symNumber = newSymbol("number")

func IsNumber(obj LObject) bool {
	_, ok := obj.(linteger)
	if ok {
		return true
	}
	_, ok = obj.(lreal)
	return ok
}

func NewInteger(n int64) LObject {
	v := linteger(n)
	return v
}

func NewReal(n float64) LObject {
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
	return 0, Error("Not a real number: ", obj)
}

func IntegerValue(obj LObject) (int64, error) {
	switch n := obj.(type) {
	case linteger:
		return int64(n), nil
	case lreal:
		return int64(n), nil
	default:
		return 0, Error("Not an integer: ", obj)
	}
}

func IntValue(obj LObject) (int, error) {
	switch n := obj.(type) {
	case linteger:
		return int(n), nil
	case lreal:
		return int(n), nil
	default:
		return 0, Error("Not an integer: ", obj)
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

func Equal(o1 LObject, o2 LObject) bool {
	//value based
	if o1 == o2 {
		return true
	} else {
		return o1.Equal(o2)
	}
}

func NumericallyEqual(o1 LObject, o2 LObject) (bool, error) {
	//for scheme, only accepts numbers, else error
	switch n1 := o1.(type) {
	case linteger:
		switch n2 := o2.(type) {
		case linteger:
			return n1 == n2, nil
		case lreal:
			return float64(n1) == float64(n2), nil
		default:
			return false, Error("Not a number: ", o2)
		}
	case lreal:
		switch n2 := o2.(type) {
		case linteger:
			return float64(n2) == float64(n1), nil
		case lreal:
			return n1 == n2, nil
		default:
			return false, Error("Not a number: ", o2)
		}
	default:
		return false, Error("Not a number: ", o1)
	}
}

func Identical(n1 LObject, n2 LObject) bool {
	if n1 == n2 {
		return true
	} else {
		return false
	}
}

type linteger int64

func (linteger) Type() LObject {
	return symNumber
}

func (i linteger) Equal(another LObject) bool {
	if a, err := IntegerValue(another); err == nil {
		return int64(i) == a
	}
	return false
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

func (lreal) Type() LObject {
	return symNumber
}

func (f lreal) Equal(another LObject) bool {
	if a, err := RealValue(another); err == nil {
		return float64(f) == a
	}
	return false
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
			return nil, Error("Not a number: ", num)
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
			return nil, Error("Not a number: ", num)
		}
	}
	if integral {
		return linteger(iprod), nil
	} else {
		return lreal(fprod), nil
	}
}

//
// ------------------- list, pair
//
type lpair struct {
	car LObject
	cdr LObject
}

var symPair = newSymbol("pair")

func IsPair(obj LObject) bool {
	_, ok := obj.(*lpair)
	return ok
}

//this is the union list?, not the scheme-compatible one, which is IsProperList
func IsList(obj LObject) bool {
	return obj == NIL || IsPair(obj)
}

//this is like Scheme's list? It protects against circularity
func IsProperList(obj LObject) bool {
	if obj == NIL {
		return true
	}
	first := obj
	for IsPair(obj) {
		obj := Cdr(obj)
		if obj == first {
			//circular list
			return true
		}
	}
	return obj == NIL
}

func (*lpair) Type() LObject {
	return symPair
}

func (lst *lpair) Equal(another LObject) bool {
	if a, ok := another.(*lpair); ok {
		return Equal(lst.car, a.car) && Equal(lst.cdr, a.cdr)
	}
	return false
}

func (lst *lpair) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(lst.car.String())
	var tail LObject = lst.cdr
	b := true
	for b {
		if tail == NIL {
			b = false
		} else if IsPair(tail) {
			lst = tail.(*lpair)
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

func (lst *lpair) Length() int {
	count := 1
	var o LObject = lst.cdr
	for o != NIL {
		if p, ok := o.(*lpair); ok {
			count++
			o = p.cdr
		} else {
			return -1 //not a proper list
		}
	}
	return count
}

func NewList(count int, val LObject) LObject {
	var result LObject = NIL
	for i := 0; i < count; i++ {
		result = Cons(val, result)
	}
	return result
}

func Cons(car LObject, cdr LObject) LObject {
	lst := lpair{car, cdr}
	return &lst
}

func Car(lst LObject) LObject {
	switch p := lst.(type) {
	case *lpair:
		return p.car
	} // unlike scheme, nil is returned, rather than an error, when applied to a non-pair
	return NIL
}

func SetCar(lst LObject, obj LObject) {
	switch p := lst.(type) {
	case *lpair:
		p.car = obj
	}
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
	switch p := lst.(type) {
	case *lpair:
		return p.cdr
	} // unlike scheme, nil is returned, rather than an error, when applied to a non-pair
	return NIL
}

func SetCdr(lst LObject, obj LObject) {
	switch p := lst.(type) {
	case *lpair:
		p.cdr = obj
	}
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

func ToImproperList(vec []LObject, rest LObject) LObject {
	var p LObject
	p = rest
	for i := len(vec) - 1; i >= 0; i-- {
		v := vec[i]
		p = Cons(v, p)
	}
	return p
}

func List(vec ...LObject) LObject {
	return ToList(vec)
}

func VectorToList(vec LObject) (LObject, error) {
	v, ok := vec.(*lvector)
	if !ok {
		return nil, TypeError(symVector, vec)
	}
	return ToList(v.elements), nil
}

func Length(seq LObject) int {
	if seq == NIL {
		return 0
	} else {
		switch v := seq.(type) {
		case lstring:
			return len(v)
		case *lvector:
			return len(v.elements)
		case *lpair:
			return v.Length()
		case *lmap:
			return len(v.bindings)
		default:
			return -1
		}
	}
}

func Reverse(lst LObject) (LObject, error) {
	var rev LObject
	rev = NIL
	for lst != NIL {
		switch v := lst.(type) {
		case *lpair:
			rev = Cons(v.car, rev)
			lst = v.cdr
		default:
			return nil, Error("Not a proper list: ", lst)
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
		case *lpair:
			seq2 = Cons(v.car, seq2)
			rev = v.cdr
		}
	}
	return seq2, nil

}

//
// ------------------- vector
//

type lvector struct {
	elements []LObject
}

func NewVector(size int, init LObject) LObject {
	elements := make([]LObject, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	vec := lvector{elements}
	return &vec
}

func Vector(elements ...LObject) LObject {
	vec := lvector{elements}
	return &vec
}

func ToVector(elements []LObject, count int) LObject {
	el := make([]LObject, count)
	copy(el, elements[0:count])
	vec := lvector{el}
	return &vec
}

var symVector = newSymbol("vector")

func IsVector(obj LObject) bool {
	_, ok := obj.(*lvector)
	return ok
	//	return obj.Type() == symVector
}

func (*lvector) Type() LObject {
	return symVector
}

func (vec *lvector) Equal(another LObject) bool {
	if a, ok := another.(*lvector); ok {
		vlen := len(vec.elements)
		if vlen == len(a.elements) {
			for i := 0; i < vlen; i++ {
				if !Equal(vec.elements[i], a.elements[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
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

func VectorLength(vec LObject) (int, error) {
	if v, ok := vec.(*lvector); ok {
		return len(v.elements), nil
	}
	return 0, TypeError(symVector, vec)
	//	return Error("Not a vector: ", vec)
}

func VectorSet(vec LObject, idx int, obj LObject) error {
	if v, ok := vec.(*lvector); ok {
		if idx < 0 || idx >= len(v.elements) {
			return Error("Vector index out of range")
		}
		v.elements[idx] = obj
		return nil
	}
	return Error("Not a vector: ", vec)
}

func VectorRef(vec LObject, idx int) (LObject, error) {
	if v, ok := vec.(*lvector); ok {
		if idx < 0 || idx >= len(v.elements) {
			return nil, Error("Vector index out of range")
		}
		return v.elements[idx], nil
	}
	return nil, Error("Not a vector: ", vec)
}

//
// ------------------- map
//
type lmap struct {
	bindings map[LObject]LObject
}

func ToMap(pairwiseBindings []LObject, count int) (LObject, error) {
	if count%2 != 0 {
		return nil, Error("Initializing a map requires an even number of elements")
	}
	bindings := make(map[LObject]LObject, count/2)
	for i := 0; i < count; i += 2 {
		bindings[pairwiseBindings[i]] = pairwiseBindings[i+1]
	}
	m := lmap{bindings}
	return &m, nil
}

func Map(pairwiseBindings ...LObject) (LObject, error) {
	return ToMap(pairwiseBindings, len(pairwiseBindings))
}

var symMap = newSymbol("map")

func (*lmap) Type() LObject {
	return symMap
}

func (m *lmap) Equal(another LObject) bool {
	if a, ok := another.(*lmap); ok {
		mlen := len(m.bindings)
		if mlen == len(a.bindings) {
			for k, v := range m.bindings {
				if v2, ok := a.bindings[k]; ok {
					if !Equal(v, v2) {
						return false
					}
				} else {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (m *lmap) Length() int {
	return len(m.bindings)
}

func (m *lmap) String() string {
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

func (m *lmap) Put(key LObject, value LObject) LObject {
	m.bindings[key] = value
	return m
}

func (m *lmap) Get(key LObject) LObject {
	if val, ok := m.bindings[key]; ok {
		return val
	} else {
		return NIL
	}
}

func (m *lmap) Has(key LObject) bool {
	_, ok := m.bindings[key]
	return ok
}

func Has(obj LObject, key LObject) (bool, error) {
	if aMap, ok := obj.(*lmap); ok {
		return aMap.Has(key), nil
	}
	return false, TypeError(symMap, obj)
}

func Get(obj LObject, key LObject) (LObject, error) {
	if aMap, ok := obj.(*lmap); ok {
		return aMap.Get(key), nil
	}
	return nil, TypeError(symMap, obj)
}

func Put(obj LObject, key LObject, value LObject) (LObject, error) {
	if aMap, ok := obj.(*lmap); ok {
		return aMap.Put(key, value), nil
	}
	return nil, TypeError(symMap, obj)
}

//
// ------------------- error
//

func Error(arg1 LAny, args ...LAny) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%v", arg1))
	for _, o := range args {
		if l, ok := o.(LObject); ok {
			buf.WriteString(fmt.Sprintf("%v", Write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
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

func TypeError(typeSym LObject, obj LObject) error {
	return Error("Type error: expected ", typeSym, ", got ", obj)
}
