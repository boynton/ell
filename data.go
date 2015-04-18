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

func (v lnil) String() string {
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

//FALSE is Ell's false constant
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
// ------------------- symbol, keyword
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

func isSymbol(obj lob) bool {
	_, ok := obj.(*lsymbol)
	return ok
}

func isKeyword(obj lob) bool {
	sym, ok := obj.(*lsymbol)
	if ok {
		return (sym.Name[len(sym.Name)-1] == ':' || sym.Name[0] == ':')
	}
	return false
}

func unkeyword(obj lob) lob {
	sym, ok := obj.(*lsymbol)
	if ok {
		last := len(sym.Name) - 1
		if sym.Name[last] == ':' {
			return intern(sym.Name[0:last])
		} else if sym.Name[0] == ':' {
			return intern(sym.Name[1:])
		}
	}
	return obj
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
	"list":    symList,
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
	if len(name) == 0 {
		panic("empty symbol!")
	}
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

func newString(val string) lstring {
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
		default:
			//to do: handle UTF8 correctly
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

func newCharacter(c rune) lchar {
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

func isInteger(obj lob) bool {
	_, ok := obj.(linteger)
	return ok
}

func isNumber(obj lob) bool {
	switch obj.(type) {
	case linteger:
		return true
	case lreal:
		return true
	default:
		return false
	}
}

func newInteger(n int64) linteger {
	v := linteger(n)
	return v
}

func newReal(n float64) lreal {
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
	if o1 == o2 {
		return true
	}
	return o1.equal(o2)
}

func numericallyEqual(o1 lob, o2 lob) (bool, error) {
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
			if integral {
				isum += int64(n)
			} else {
				fsum += float64(n)
			}
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

func sub(num1 lob, num2 lob) (lob, error) {
	n1, err := realValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := realValue(num2)
	if err != nil {
		return nil, err
	}
	return newReal(n1 - n2), nil
}

func minus(nums []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError("-", "1+", argc)
	}
	var isum int64
	var fsum float64
	integral := true
	num := nums[0]
	switch n := num.(type) {
	case linteger:
		isum = int64(n)
	case lreal:
		integral = false
		fsum = float64(n)
	default:
		return nil, typeError(symNumber, num)
	}
	if argc == 1 {
		if integral {
			isum = -isum
		} else {
			fsum = -fsum
		}
	} else {
		for _, num := range nums[1:] {
			switch n := num.(type) {
			case linteger:
				if integral {
					isum -= int64(n)
				} else {
					fsum -= float64(n)
				}
			case lreal:
				if integral {
					fsum = float64(isum)
					integral = false
				}
				fsum -= float64(n)
			default:
				return nil, typeError(symNumber, num)
			}
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

func div(argv []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError("/", "1+", argc)
	} else if argc == 1 {
		n1, err := realValue(argv[0])
		if err != nil {
			return nil, err
		}
		return lreal(1.0 / n1), nil
	} else {
		quo, err := realValue(argv[0])
		if err != nil {
			return nil, err
		}
		for i := 1; i < argc; i++ {
			n, err := realValue(argv[i])
			if err != nil {
				return nil, err
			}
			quo /= n
		}
		return lreal(quo), nil
	}
}

//
// ------------------- list
//
type llist struct {
	car lob
	cdr *llist
}

var symList = newSymbol("list")
var EMPTY_LIST = initEmpty()

func initEmpty() *llist {
	lst := &llist{NIL, nil}
	lst.cdr = lst
	return lst
}

func isList(obj lob) bool {
	_, ok := obj.(*llist)
	return ok
}

func (*llist) typeSymbol() lob {
	return symList
}

func (lst *llist) equal(another lob) bool {
	if a, ok := another.(*llist); ok {
		for lst != EMPTY_LIST {
			if a == EMPTY_LIST {
				return false
			}
			if !equal(lst.car, a.car) {
				return false
			}
			lst = lst.cdr
			a = a.cdr
		}
		if lst == a {
			return true
		}
	}
	return false
}

func (lst *llist) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	delim := ""
	if lst == lst.cdr {
		return "()"
	}
	for lst != EMPTY_LIST {
		buf.WriteString(delim)
		delim = " "
		buf.WriteString(lst.car.String())
		lst = lst.cdr
	}
	buf.WriteString(")")
	return buf.String()
}

func (lst *llist) length() int {
	if lst == EMPTY_LIST {
		return 0
	}
	count := 1
	o := lst.cdr
	for o != EMPTY_LIST {
		count++
		o = o.cdr
	}
	return count
}

func newList(count int, val lob) *llist {
	result := EMPTY_LIST
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func cons(car lob, cdr *llist) *llist {
	return &llist{car, cdr}
}

func car(lst lob) lob {
	switch p := lst.(type) {
	case *llist:
		return p.car
	}
	return NIL
}

func setCar(lst lob, obj lob) {
	switch p := lst.(type) {
	case *llist:
		p.car = obj
	}
}

func cdr(lst lob) *llist {
	if lst != EMPTY_LIST {
		switch p := lst.(type) {
		case *llist:
			return p.cdr
		}
	}
	return EMPTY_LIST
}

func setCdr(lst lob, obj lob) {
	switch p := lst.(type) {
	case *llist:
		switch n := obj.(type) {
		case *llist:
			p.cdr = n
		default:
			println("IGNORED: Setting cdr to non-list: ", obj)
		}
	default:
		println("IGNORED: Setting cdr of non-list: ", lst)
	}
}

func caar(lst lob) lob {
	return car(car(lst))
}
func cadr(lst lob) lob {
	return car(cdr(lst))
}
func cddr(lst lob) *llist {
	return cdr(cdr(lst))
}
func caddr(lst lob) lob {
	return car(cdr(cdr(lst)))
}
func cdddr(lst lob) *llist {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst lob) lob {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst lob) *llist {
	return cdr(cdr(cdr(cdr(lst))))
}

func toList(vec []lob) *llist {
	p := EMPTY_LIST
	for i := len(vec) - 1; i >= 0; i-- {
		v := vec[i]
		p = cons(v, p)
	}
	return p
}

func list(vec ...lob) *llist {
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
	case *llist:
		return v.length()
	case *lmap:
		return len(v.bindings)
	default:
		return -1
	}
}

func reverse(lst *llist) (*llist, error) {
	rev := EMPTY_LIST
	for lst != EMPTY_LIST {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev, nil
}

func concat(seq1 *llist, seq2 *llist) (*llist, error) {
	rev, err := reverse(seq1)
	if err != nil {
		return nil, err
	}
	if rev == EMPTY_LIST {
		return seq2, nil
	}
	lst := seq2
	for rev != EMPTY_LIST {
		lst = cons(rev.car, lst)
		rev = rev.cdr
	}
	return lst, nil
}

//
// ------------------- vector
//

type lvector struct {
	elements []lob
}

func newVector(size int, init lob) *lvector {
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

func toMap(pairwiseBindings []lob, count int) (*lmap, error) {
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

func newMap(pairwiseBindings ...lob) (*lmap, error) {
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
