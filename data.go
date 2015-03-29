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

package main

import (
	"bytes"
	"fmt"
	"strconv"
)

// for both ell and non-ell values
type any interface{}

//
// The generic Ell object, which can be queried for its symbolic type name at runtime
//
type lob interface {
	typeSymbol() lob
	equal(another lob) bool
	String() string
}

//
// ------------------- nil
//

// NIL is Ell's version of nil, not Go's
const NIL = lnil(0)

type lnil int

var symNull = newSymbol("null")

func (lnil) typeSymbol() lob {
	return symNull
}

func (lnil) equal(another lob) bool {
	return another == NIL
}

func (lnil) String() string {
	return "nil"
}

//
// ------------------- EOF marker
//

// EOF is Ell's EOF object
const EOF = leof(0)

type leof int

var symEOF = newSymbol("eof")

func (leof) typeSymbol() lob {
	return symEOF
}

func (leof) equal(another lob) bool {
	return another == EOF
}

func (leof) String() string {
	return "<EOF>"
}

//
// ------------------- boolean
//

//TRUE is Ell's true constant
const TRUE lboolean = lboolean(true)

//FALSE is Ell's flse constant
const FALSE lboolean = lboolean(false)

type lboolean bool

var symBoolean = newSymbol("boolean")

func isBoolean(obj lob) bool {
	_, ok := obj.(lboolean)
	return ok
}

func (lboolean) typeSymbol() lob {
	return symBoolean
}

func (b lboolean) equal(another lob) bool {
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

//func newSymbol(name string) *lsymbol {
func newSymbol(name string) lob {
	sym := lsymbol{name, symtag}
	symtag++
	return &sym
}

var symSymbol = newSymbol("symbol")

func isSymbol(obj lob) bool {
	_, ok := obj.(*lsymbol)
	return ok
}

func (*lsymbol) typeSymbol() lob {
	return symSymbol
}

func (sym *lsymbol) equal(another lob) bool {
	if a, ok := another.(*lsymbol); ok {
		return sym == a
	}
	return false
}

func (sym *lsymbol) String() string {
	return sym.Name
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]lob{
	"null":    symNull, //the type of NIL
	"boolean": symBoolean,
	"symbol":  symSymbol,
	"string":  symString,
	"number":  symNumber,
	"pair":    symPair,
	"vector":  symVector,
	"map":     symMap,
	"eof":     symEOF,
}

func symbols() []lob {
	syms := make([]lob, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func intern(name string) lob {
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

func newString(val string) lob {
	s := lstring(val)
	return s
}

func isString(obj lob) bool {
	_, ok := obj.(lstring)
	return ok
}

func (lstring) typeSymbol() lob {
	return symString
}

func (s lstring) equal(another lob) bool {
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

func (s lstring) encodedString() string {
	return encodeString(string(s))
}

func (s lstring) String() string {
	//return encodeString(string(s))
	return string(s)
}

//
// ------------------- character
//
var symCharacter = newSymbol("character")

func isCharacter(obj lob) bool {
	_, ok := obj.(lchar)
	if ok {
		return true
	}
	_, ok = obj.(lreal)
	return ok
}

func newCharacter(c rune) lob {
	v := lchar(c)
	return v
}

type lchar rune

func (lchar) typeSymbol() lob {
	return symCharacter
}

func (i lchar) equal(another lob) bool {
	if a, err := intValue(another); err == nil {
		return int(i) == a
	}
	return false
}

func (i lchar) String() string {
	buf := []rune{rune(i)}
	return string(buf)
}

//
// ------------------- number
//

var symNumber = newSymbol("number")

func isNumber(obj lob) bool {
	_, ok := obj.(linteger)
	if ok {
		return true
	}
	_, ok = obj.(lreal)
	return ok
}

func newInteger(n int64) lob {
	v := linteger(n)
	return v
}

func newReal(n float64) lob {
	v := lreal(n)
	return v
}

func realValue(obj lob) (float64, error) {
	switch n := obj.(type) {
	case linteger:
		return float64(n), nil
	case lreal:
		return float64(n), nil
	}
	return 0, typeError(symNumber, obj)
}

func integerValue(obj lob) (int64, error) {
	switch n := obj.(type) {
	case linteger:
		return int64(n), nil
	case lreal:
		return int64(n), nil
	case lchar:
		return int64(n), nil
	default:
		return 0, typeError(symNumber, obj)
	}
}

func intValue(obj lob) (int, error) {
	switch n := obj.(type) {
	case linteger:
		return int(n), nil
	case lreal:
		return int(n), nil
	case lchar:
		return int(n), nil
	default:
		return 0, typeError(symNumber, obj)
	}
}

func greaterOrEqual(n1 lob, n2 lob) (lob, error) {
	f1, err := realValue(n1)
	if err == nil {
		f2, err := realValue(n2)
		if err == nil {
			if f1 >= f2 {
				return TRUE, nil
			}
			return FALSE, nil
		}
		return nil, err
	}
	return nil, err
}

func lessOrEqual(n1 lob, n2 lob) (lob, error) {
	f1, err := realValue(n1)
	if err == nil {
		f2, err := realValue(n2)
		if err == nil {
			if f1 <= f2 {
				return TRUE, nil
			}
			return FALSE, nil
		}
		return nil, err
	}
	return nil, err
}

func greater(n1 lob, n2 lob) (lob, error) {
	f1, err := realValue(n1)
	if err == nil {
		f2, err := realValue(n2)
		if err == nil {
			if f1 > f2 {
				return TRUE, nil
			}
			return FALSE, nil
		}
		return nil, err
	}
	return nil, err
}

func less(n1 lob, n2 lob) (lob, error) {
	f1, err := realValue(n1)
	if err == nil {
		f2, err := realValue(n2)
		if err == nil {
			if f1 < f2 {
				return TRUE, nil
			}
			return FALSE, nil
		}
		return nil, err
	}
	return nil, err
}

func equal(o1 lob, o2 lob) bool {
	//value based
	if o1 == o2 {
		return true
	}
	return o1.equal(o2)
}

func numericallyEqual(o1 lob, o2 lob) (bool, error) {
	//for scheme, only accepts numbers, else error
	switch n1 := o1.(type) {
	case linteger:
		switch n2 := o2.(type) {
		case linteger:
			return n1 == n2, nil
		case lreal:
			return float64(n1) == float64(n2), nil
		default:
			return false, typeError(symNumber, o2)
		}
	case lreal:
		switch n2 := o2.(type) {
		case linteger:
			return float64(n2) == float64(n1), nil
		case lreal:
			return n1 == n2, nil
		default:
			return false, typeError(symNumber, o2)
		}
	default:
		return false, typeError(symNumber, o2)
	}
}

func identical(n1 lob, n2 lob) bool {
	return n1 == n2
}

type linteger int64

func (linteger) typeSymbol() lob {
	return symNumber
}

func (i linteger) equal(another lob) bool {
	if a, err := integerValue(another); err == nil {
		return int64(i) == a
	}
	return false
}

func (i linteger) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i linteger) integerValue() int64 {
	return int64(i)
}
func (i linteger) realValue() float64 {
	return float64(i)
}

type lreal float64

func (lreal) typeSymbol() lob {
	return symNumber
}

func (f lreal) equal(another lob) bool {
	if a, err := realValue(another); err == nil {
		return float64(f) == a
	}
	return false
}

func (f lreal) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f lreal) integerValue() int64 {
	return int64(f)
}

func (f lreal) realValue() float64 {
	return float64(f)
}

func add(num1 lob, num2 lob) (lob, error) {
	n1, err := realValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := realValue(num2)
	if err != nil {
		return nil, err
	}
	return newReal(n1 + n2), nil
}

func sum(nums []lob, argc int) (lob, error) {
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
			return nil, typeError(symNumber, num)
		}
	}
	if integral {
		return linteger(isum), nil
	}
	return lreal(fsum), nil
}

func mul(num1 lob, num2 lob) (lob, error) {
	n1, err := realValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := realValue(num2)
	if err != nil {
		return nil, err
	}
	return newReal(n1 * n2), nil
}

func product(argv []lob, argc int) (lob, error) {
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
			return nil, typeError(symNumber, num)
		}
	}
	if integral {
		return linteger(iprod), nil
	}
	return lreal(fprod), nil
}

//
// ------------------- list, pair
//
type lpair struct {
	car lob
	cdr lob
}

var symPair = newSymbol("pair")

func isPair(obj lob) bool {
	_, ok := obj.(*lpair)
	return ok
}

//this is the union list?, not the scheme-compatible one, which is IsProperList
func isList(obj lob) bool {
	return obj == NIL || isPair(obj)
}

//this is like Scheme's list? It protects against circularity
func isProperList(obj lob) bool {
	if obj == NIL {
		return true
	}
	first := obj
	for isPair(obj) {
		obj := cdr(obj)
		if obj == first {
			//circular list
			return true
		}
	}
	return obj == NIL
}

func (*lpair) typeSymbol() lob {
	return symPair
}

func (lst *lpair) equal(another lob) bool {
	if a, ok := another.(*lpair); ok {
		return equal(lst.car, a.car) && equal(lst.cdr, a.cdr)
	}
	return false
}

func (lst *lpair) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(lst.car.String())
	tail := lst.cdr
	b := true
	for b {
		if tail == NIL {
			b = false
		} else if isPair(tail) {
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

func (lst *lpair) length() int {
	count := 1
	o := lst.cdr
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

func newList(count int, val lob) lob {
	var result lob = NIL
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func cons(car lob, cdr lob) lob {
	lst := lpair{car, cdr}
	return &lst
}

func car(lst lob) lob {
	switch p := lst.(type) {
	case *lpair:
		return p.car
	} // unlike scheme, nil is returned, rather than an error, when applied to a non-pair
	return NIL
}

func setCar(lst lob, obj lob) {
	switch p := lst.(type) {
	case *lpair:
		p.car = obj
	}
}

func caar(lst lob) lob {
	return car(car(lst))
}
func cadr(lst lob) lob {
	return car(cdr(lst))
}
func cddr(lst lob) lob {
	return cdr(cdr(lst))
}
func caddr(lst lob) lob {
	return car(cdr(cdr(lst)))
}
func cdddr(lst lob) lob {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst lob) lob {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst lob) lob {
	return cdr(cdr(cdr(cdr(lst))))
}

func cdr(lst lob) lob {
	switch p := lst.(type) {
	case *lpair:
		return p.cdr
	} // unlike scheme, nil is returned, rather than an error, when applied to a non-pair
	return NIL
}

func setCdr(lst lob, obj lob) {
	switch p := lst.(type) {
	case *lpair:
		p.cdr = obj
	}
}

func toList(vec []lob) lob {
	var p lob
	p = NIL
	for i := len(vec) - 1; i >= 0; i-- {
		v := vec[i]
		p = cons(v, p)
	}
	return p
}

func toImproperList(vec []lob, rest lob) lob {
	var p lob
	p = rest
	for i := len(vec) - 1; i >= 0; i-- {
		v := vec[i]
		p = cons(v, p)
	}
	return p
}

func list(vec ...lob) lob {
	return toList(vec)
}

func vectorToList(vec lob) (lob, error) {
	v, ok := vec.(*lvector)
	if !ok {
		return nil, typeError(symVector, vec)
	}
	return toList(v.elements), nil
}

func length(seq lob) int {
	if seq == NIL {
		return 0
	}
	switch v := seq.(type) {
	case lstring:
		return len(v)
	case *lvector:
		return len(v.elements)
	case *lpair:
		return v.length()
	case *lmap:
		return len(v.bindings)
	default:
		return -1
	}
}

func reverse(lst lob) (lob, error) {
	var rev lob
	rev = NIL
	for lst != NIL {
		switch v := lst.(type) {
		case *lpair:
			rev = cons(v.car, rev)
			lst = v.cdr
		default:
			return nil, newError("Not a proper list: ", lst)
		}
	}
	return rev, nil
}

func concat(seq1 lob, seq2 lob) (lob, error) {
	rev, err := reverse(seq1)
	if err != nil {
		return nil, err
	}
	for rev != NIL {
		switch v := rev.(type) {
		case *lpair:
			seq2 = cons(v.car, seq2)
			rev = v.cdr
		}
	}
	return seq2, nil

}

//
// ------------------- vector
//

type lvector struct {
	elements []lob
}

func newVector(size int, init lob) lob {
	elements := make([]lob, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	vec := lvector{elements}
	return &vec
}

func vector(elements ...lob) lob {
	vec := lvector{elements}
	return &vec
}

func toVector(elements []lob, count int) lob {
	el := make([]lob, count)
	copy(el, elements[0:count])
	vec := lvector{el}
	return &vec
}

var symVector = newSymbol("vector")

func isVector(obj lob) bool {
	_, ok := obj.(*lvector)
	return ok
}

func (*lvector) typeSymbol() lob {
	return symVector
}

func (vec *lvector) equal(another lob) bool {
	if a, ok := another.(*lvector); ok {
		vlen := len(vec.elements)
		if vlen == len(a.elements) {
			for i := 0; i < vlen; i++ {
				if !equal(vec.elements[i], a.elements[i]) {
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

func (vec *lvector) length() int {
	return len(vec.elements)
}

func vectorLength(vec lob) (int, error) {
	if v, ok := vec.(*lvector); ok {
		return len(v.elements), nil
	}
	return 1, typeError(symVector, vec)
}

func vectorSet(vec lob, idx int, obj lob) error {
	if v, ok := vec.(*lvector); ok {
		if idx < 0 || idx >= len(v.elements) {
			return newError("Vector index out of range")
		}
		v.elements[idx] = obj
		return nil
	}
	return typeError(symVector, vec)
}

func vectorRef(vec lob, idx int) (lob, error) {
	if v, ok := vec.(*lvector); ok {
		if idx < 0 || idx >= len(v.elements) {
			return nil, newError("Vector index out of range")
		}
		return v.elements[idx], nil
	}
	return nil, typeError(symVector, vec)
}

//
// ------------------- map
//
type lmap struct {
	bindings map[lob]lob
}

func toMap(pairwiseBindings []lob, count int) (lob, error) {
	if count%2 != 0 {
		return nil, newError("Initializing a map requires an even number of elements")
	}
	bindings := make(map[lob]lob, count/2)
	for i := 0; i < count; i += 2 {
		bindings[pairwiseBindings[i]] = pairwiseBindings[i+1]
	}
	m := lmap{bindings}
	return &m, nil
}

func newMap(pairwiseBindings ...lob) (lob, error) {
	return toMap(pairwiseBindings, len(pairwiseBindings))
}

func isMap(obj lob) bool {
	_, ok := obj.(*lmap)
	return ok
}

var symMap = newSymbol("map")

func (*lmap) typeSymbol() lob {
	return symMap
}

func (m *lmap) equal(another lob) bool {
	if a, ok := another.(*lmap); ok {
		mlen := len(m.bindings)
		if mlen == len(a.bindings) {
			for k, v := range m.bindings {
				if v2, ok := a.bindings[k]; ok {
					if !equal(v, v2) {
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

func (m *lmap) length() int {
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

func (m *lmap) put(key lob, value lob) lob {
	m.bindings[key] = value
	return m
}

func (m *lmap) get(key lob) lob {
	if val, ok := m.bindings[key]; ok {
		return val
	}
	return NIL
}

func (m *lmap) has(key lob) bool {
	_, ok := m.bindings[key]
	return ok
}

func has(obj lob, key lob) (bool, error) {
	if aMap, ok := obj.(*lmap); ok {
		return aMap.has(key), nil
	}
	return false, typeError(symMap, obj)
}

func get(obj lob, key lob) (lob, error) {
	if aMap, ok := obj.(*lmap); ok {
		return aMap.get(key), nil
	}
	return nil, typeError(symMap, obj)
}

func put(obj lob, key lob, value lob) (lob, error) {
	if aMap, ok := obj.(*lmap); ok {
		return aMap.put(key, value), nil
	}
	return nil, typeError(symMap, obj)
}

//
// ------------------- error
//

func newError(arg1 any, args ...any) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%v", arg1))
	for _, o := range args {
		if l, ok := o.(lob); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
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

func typeError(typeSym lob, obj lob) error {
	return newError("Type error: expected ", typeSym, ", got ", obj)
}
