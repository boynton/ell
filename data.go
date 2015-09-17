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
	"math"
	"strconv"
)

// for both ell and non-ell values
type any interface{}

//
// The generic Ell object, which can be queried for its symbolic type name at runtime
//
type lob interface {
	Type() lob
	Equal(another lob) bool
	String() string
}

//
// ------------------- null
//

type lnull int

var typeNull = newSymbol("<null>")
var symNull = newSymbol("null")

func (lnull) Type() lob {
	return typeNull
}

func (lnull) Equal(another lob) bool {
	return another == Null
}

func (v lnull) String() string {
	return "null"
}

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList
const Null = lnull(0)

//
// ------------------- EOF marker
//

// EOF is Ell's EOF object
const EOF = leof(0)

type leof int

var typeEOF = newSymbol("<eof>")
var symEOF = newSymbol("eof")

func (leof) Type() lob {
	return typeEOF
}

func (leof) Equal(another lob) bool {
	return another == EOF
}

func (leof) String() string {
	return "#eof"
}

//
// ------------------- boolean
//

//TRUE is Ell's true constant
const True lboolean = lboolean(true)

//FALSE is Ell's false constant
const False lboolean = lboolean(false)

type lboolean bool

var typeBoolean = newSymbol("<boolean>")
var symBoolean = newSymbol("boolean")

func isBoolean(obj lob) bool {
	_, ok := obj.(lboolean)
	return ok
}

func (lboolean) Type() lob {
	return typeBoolean
}

func (b lboolean) Equal(another lob) bool {
	if a, ok := another.(lboolean); ok {
		return b == a
	}
	return false
}

func (b lboolean) String() string {
	return strconv.FormatBool(bool(b))
}

//
// ------------------- symbol, keyword, type
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

var typeSymbol = newSymbol("<symbol>")
var typeKeyword = newSymbol("<keyword>")
var typeType = newSymbol("<type>")

func (sym *lsymbol) Type() lob {
	if isSymbolKeyword(sym) {
		return typeKeyword
	} else if isSymbolType(sym) {
		return typeType
	}
	return typeSymbol
}

func (sym *lsymbol) Equal(another lob) bool {
	if a, ok := another.(*lsymbol); ok {
		return sym == a
	}
	return false
}

func (sym *lsymbol) String() string {
	return sym.Name
}

func isSymbolKeyword(sym *lsymbol) bool {
	n := len(sym.Name)
	return sym.Name[n-1] == ':'
}

func isSymbolType(sym *lsymbol) bool {
	n := len(sym.Name)
	if n > 2 && sym.Name[0] == '<' && sym.Name[n-1] == '>' {
		return true
	}
	return false
}

func isSymbol(obj lob) bool {
	sym, ok := obj.(*lsymbol)
	return ok && !isSymbolKeyword(sym) && !isSymbolType(sym)
}

func isType(obj lob) bool {
	sym, ok := obj.(*lsymbol)
	if ok {
		return isSymbolType(sym)
	}
	return false
}

func isKeyword(obj lob) bool {
	sym, ok := obj.(*lsymbol)
	if ok {
		return isSymbolKeyword(sym)
	}
	return false
}

func unKeywordString(sym *lsymbol) string {
	if isSymbolKeyword(sym) {
		return sym.Name[:len(sym.Name)-1]
	}
	return sym.Name
}

func keyword(obj lob) (lob, error) {
	switch t := obj.(type) {
	case lnull:
		return obj, nil
	case *lsymbol:
		last := len(t.Name) - 1 //can never be less than 0, symbols always have len > 0
		if t.Name[last] == ':' {
			return obj, nil
		}
		return intern(t.Name + ":"), nil
	case lstring:
		s := string(t)
		last := len(s) - 1
		if last < 0 {
			return obj, nil
		}
		if s[last] != ':' {
			s += ":"
		}
		return intern(s), nil
	default:
		return nil, newError("Type error: expected symbol or string, got ", obj)
	}
}

func typeName(obj lob) (*lsymbol, error) {
	switch t := obj.(type) {
	case *lsymbol:
		if !isSymbolType(t) {
			return nil, typeError(typeType, obj)
		}
		return internSymbol(t.Name[1:len(t.Name)-1]), nil
	default:
		return nil, newError("Type error: expected symbol or string, got ", obj)
	}
}

func unkeyword(obj lob) (lob, error) {
	switch t := obj.(type) {
	case lnull:
		return obj, nil
	case *lsymbol:
		last := len(t.Name) - 1 //can never be less than 0, symbols always have len > 0
		if t.Name[last] != ':' {
			return obj, nil
		}
		return intern(t.Name[:last]), nil
	case lstring:
		s := string(t)
		last := len(s) - 1
		if last < 0 {
			return obj, nil
		}
		if s[last] == ':' {
			s = s[:last]
		}
		return intern(s), nil
	default:
		return nil, newError("Type error: expected symbol or string, got ", obj)
	}
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]*lsymbol {
	"<null>":           typeNull,
	"<boolean>":          typeBoolean,
	"<symbol>":           typeSymbol,
	"<string>":           typeString,
	"<number>":           typeNumber,
	"<list>":             typeList,
	"<array>":            typeArray,
	"<struct>":           typeStruct,
	"<eof>":              typeEOF,
	"quote":            symQuote,
	"quasiquote":       symQuasiquote,
	"unquote":          symUnquote,
	"unquote-splicing": symUnquoteSplicing,
}

func symbols() []lob {
	syms := make([]lob, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func symbol(names []lob) (lob, error) {
	size := len(names)
	if size < 1 {
		return argcError("symbol", "1+", size)
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch t := o.(type) {
		case lstring:
			s = string(t)
		case *lsymbol:
			s = t.Name
		default:
			return nil, newError("symbol name component invalid: ", o)
		}
		name += s
	}
	return intern(name), nil
}

func internSymbol(name string) *lsymbol {
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

func intern(name string) lob {
	return internSymbol(name)
}

//
// ------------------- string
//

type lstring string

var typeString = newSymbol("<string>")
//var typeString = newSymbol("string")

func isString(obj lob) bool {
	_, ok := obj.(lstring)
	return ok
}

func stringValue(obj lob) (string, error) {
	switch s := obj.(type) {
	case lstring:
		return string(s), nil
	default:
		return "", typeError(typeString, obj)
	}
}

func (lstring) Type() lob {
	return typeString
}

func (s lstring) Equal(another lob) bool {
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
var symCharType = newSymbol("<char>")
var symChar = newSymbol("char")

func isChar(obj lob) bool {
	_, ok := obj.(lchar)
	if ok {
		return true
	}
	return ok
}

func newCharacter(c rune) lchar {
	v := lchar(c)
	return v
}

type lchar rune

func (lchar) Type() lob {
	return symCharType
}

func (i lchar) Equal(another lob) bool {
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

var typeNumber = newSymbol("<number>")

type lnumber float64

func (lnumber) Type() lob {
	return typeNumber
}

func (f lnumber) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func theNumber(obj lob) (lnumber, error) {
	if n, ok := obj.(lnumber); ok {
		return n, nil
	}
	return 0, typeError(typeNumber, obj)
}

func isInt(obj lob) bool {
	if n, ok := obj.(lnumber); ok {
		f := float64(n)
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj lob) bool {
	return !isInt(obj)
}

func isNumber(obj lob) bool {
	_, ok := obj.(lnumber)
	return ok
}

func floatValue(obj lob) (float64, error) {
	switch n := obj.(type) {
	case lnumber:
		return float64(n), nil
	}
	return 0, typeError(typeNumber, obj)
}

func int64Value(obj lob) (int64, error) {
	switch n := obj.(type) {
	case lnumber:
		return int64(n), nil
	case lchar:
		return int64(n), nil
	default:
		return 0, typeError(typeNumber, obj)
	}
}

func intValue(obj lob) (int, error) {
	switch n := obj.(type) {
	case lnumber:
		return int(n), nil
	case lchar:
		return int(n), nil
	default:
		return 0, typeError(typeNumber, obj)
	}
}

func greaterOrEqual(n1 lob, n2 lob) (lob, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 >= f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func lessOrEqual(n1 lob, n2 lob) (lob, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 <= f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func greater(n1 lob, n2 lob) (lob, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 > f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func less(n1 lob, n2 lob) (lob, error) {
	f1, err := floatValue(n1)
	if err == nil {
		f2, err := floatValue(n2)
		if err == nil {
			if f1 < f2 {
				return True, nil
			}
			return False, nil
		}
		return nil, err
	}
	return nil, err
}

func equal(o1 lob, o2 lob) bool {
	if o1 == o2 {
		return true
	}
	return o1.Equal(o2)
}

func numericallyEqual(o1 lob, o2 lob) (bool, error) {
	switch n1 := o1.(type) {
	case lnumber:
		switch n2 := o2.(type) {
		case lnumber:
			return n1 == n2, nil
		default:
			return false, typeError(typeNumber, o2)
		}
	default:
		return false, typeError(typeNumber, o1)
	}
}

func identical(n1 lob, n2 lob) bool {
	return n1 == n2
}

func (f lnumber) Equal(another lob) bool {
	if a, err := floatValue(another); err == nil {
		return float64(f) == a
	}
	return false
}

func add(num1 lob, num2 lob) (lob, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return lnumber(n1 + n2), nil
}

func sum(nums []lob, argc int) (lob, error) {
	var sum float64
	for _, num := range nums {
		switch n := num.(type) {
		case lnumber:
			sum += float64(n)
		default:
			return nil, typeError(typeNumber, num)
		}
	}
	return lnumber(sum), nil
}

func sub(num1 lob, num2 lob) (lob, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return lnumber(n1 - n2), nil
}

func minus(nums []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError("-", "1+", argc)
	}
	var fsum float64
	num := nums[0]
	switch n := num.(type) {
	case lnumber:
		fsum = float64(n)
	default:
		return nil, typeError(typeNumber, num)
	}
	if argc == 1 {
		fsum = -fsum
	} else {
		for _, num := range nums[1:] {
			switch n := num.(type) {
			case lnumber:
				fsum -= float64(n)
			default:
				return nil, typeError(typeNumber, num)
			}
		}
	}
	return lnumber(fsum), nil
}

func mul(num1 lob, num2 lob) (lob, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return lnumber(n1 * n2), nil
}

func product(argv []lob, argc int) (lob, error) {
	prod := 0.0
	for _, num := range argv {
		switch n := num.(type) {
		case lnumber:
			prod *= float64(n)
		default:
			return nil, typeError(typeNumber, num)
		}
	}
	return lnumber(prod), nil
}

func div(argv []lob, argc int) (lob, error) {
	if argc < 1 {
		return argcError("/", "1+", argc)
	} else if argc == 1 {
		n1, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		return lnumber(1.0 / n1), nil
	} else {
		quo, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		for i := 1; i < argc; i++ {
			n, err := floatValue(argv[i])
			if err != nil {
				return nil, err
			}
			quo /= n
		}
		return lnumber(quo), nil
	}
}

//
// ------------------- list
//
type llist struct {
	car lob
	cdr *llist
}

var typeList = newSymbol("<list>")
//var symList = newSymbol("list")
var symQuote = newSymbol("quote")
var symQuasiquote = newSymbol("quasiquote")
var symUnquote = newSymbol("unquote")
var symUnquoteSplicing = newSymbol("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *llist {
	var lst llist
	return &lst
}

func isEmpty(col lob) bool {
	switch v := col.(type) {
	case lnull: //Do I really want this?
		return true
	case lstring:
		return len(v) == 0
	case *larray:
		return len(v.elements) == 0
	case *llist:
		return v == EmptyList
	case *lstruct:
		return len(v.bindings) == 0
	default:
		return false
	}
}

func isList(obj lob) bool {
	_, ok := obj.(*llist)
	return ok
}

func (*llist) Type() lob {
	return typeList
}

func (lst *llist) Equal(another lob) bool {
	if a, ok := another.(*llist); ok {
		for lst != EmptyList {
			if a == EmptyList {
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
	if lst != EmptyList && lst.cdr != EmptyList && cddr(lst) == EmptyList {
		if lst.car == symQuote {
			buf.WriteString("'")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		} else if lst.car == symQuasiquote {
			buf.WriteString("`")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		} else if lst.car == symUnquote {
			buf.WriteString("~")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		} else if lst.car == symUnquoteSplicing {
			buf.WriteString("~")
			buf.WriteString(cadr(lst).String())
			return buf.String()
		}
	}
	buf.WriteString("(")
	delim := ""
	for lst != EmptyList {
		buf.WriteString(delim)
		delim = " "
		buf.WriteString(lst.car.String())
		lst = lst.cdr
	}
	buf.WriteString(")")
	return buf.String()
}

func (lst *llist) length() int {
	if lst == EmptyList {
		return 0
	}
	count := 1
	o := lst.cdr
	for o != EmptyList {
		count++
		o = o.cdr
	}
	return count
}

func newList(count int, val lob) *llist {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func cons(car lob, cdr *llist) *llist {
	if car == nil {
		panic("Assertion failure: don't call cons with nil as car")
	}
	if cdr == nil {
		panic("Assertion failure: don't call cons with nil as cdr")
	}
	if inExec {
		conses++
	}
	return &llist{car, cdr}
}

func car(lst lob) lob {
	switch p := lst.(type) {
	case *llist:
		if p != EmptyList {
			return p.car
		}
	}
	return Null
}

func setCar(lst lob, obj lob) {
	switch p := lst.(type) {
	case *llist:
		if p != EmptyList {
			p.car = obj
		}
	}
}

func cdr(lst lob) *llist {
	if lst != EmptyList {
		switch p := lst.(type) {
		case *llist:
			return p.cdr
		}
	}
	return EmptyList
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

func toList(values []lob) *llist {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = cons(v, p)
	}
	return p
}

func list(values ...lob) *llist {
	return toList(values)
}

func listToArray(lst *llist) *larray {
	var elems []lob
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return &larray{elems}
}

func arrayToList(ary lob) (lob, error) {
	v, ok := ary.(*larray)
	if !ok {
		return nil, typeError(typeArray, ary)
	}
	return toList(v.elements), nil
}

func length(seq lob) int {
	switch v := seq.(type) {
	case lstring:
		return len(v)
	case *larray:
		return len(v.elements)
	case *llist:
		return v.length()
	case *lstruct:
		return len(v.bindings)
	default:
		return -1
	}
}

func reverse(lst *llist) *llist {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func concat(seq1 *llist, seq2 *llist) (*llist, error) {
	rev := reverse(seq1)
	if rev == EmptyList {
		return seq2, nil
	}
	lst := seq2
	for rev != EmptyList {
		lst = cons(rev.car, lst)
		rev = rev.cdr
	}
	return lst, nil
}

//
// ------------------- array
//

type larray struct {
	elements []lob
}

var typeArray = newSymbol("<array>")

func isArray(obj lob) bool {
	_, ok := obj.(*larray)
	return ok
}

func (*larray) Type() lob {
	return typeArray
}

func (ary *larray) Equal(another lob) bool {
	if a, ok := another.(*larray); ok {
		alen := len(ary.elements)
		if alen == len(a.elements) {
			for i := 0; i < alen; i++ {
				if !equal(ary.elements[i], a.elements[i]) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (ary *larray) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	count := len(ary.elements)
	if count > 0 {
		buf.WriteString(ary.elements[0].String())
		for i := 1; i < count; i++ {
			buf.WriteString(" ")
			buf.WriteString(ary.elements[i].String())
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func newArray(size int, init lob) *larray {
	elements := make([]lob, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return &larray{elements}
}

func array(elements ...lob) lob {
	size := len(elements)
	el := make([]lob, size)
	for i := 0; i < size; i++ {
		el[i] = elements[i]
	}
	return &larray{el}
}

func toArray(elements []lob, count int) lob {
	el := make([]lob, count)
	copy(el, elements[0:count])
	return &larray{el}
}

func (ary *larray) length() int {
	return len(ary.elements)
}

func arrayLength(ary lob) (int, error) {
	if a, ok := ary.(*larray); ok {
		return len(a.elements), nil
	}
	return 1, typeError(typeArray, ary)
}

func arraySet(ary lob, idx int, obj lob) error {
	if a, ok := ary.(*larray); ok {
		if idx < 0 || idx >= len(a.elements) {
			return newError("Array index out of range")
		}
		a.elements[idx] = obj
		return nil
	}
	return typeError(typeArray, ary)
}

func arrayRef(ary lob, idx int) (lob, error) {
	if a, ok := ary.(*larray); ok {
		if idx < 0 || idx >= len(a.elements) {
			return nil, newError("Array index out of range")
		}
		return a.elements[idx], nil
	}
	return nil, typeError(typeArray, ary)
}

//
// ------------------- struct
//
type lstruct struct {
	typesym  *lsymbol
	bindings map[lob]lob
}

var typeStruct = newSymbol("<struct>")

func (s *lstruct) Type() lob {
	return s.typesym
}

func sliceContains(slice []lob, obj lob) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func normalizeKeywordArgs(args *llist, keys []lob) (*llist, error) {
	count := length(args)
	bindings := make(map[lob]lob, count/2)
	for args != EmptyList {
		key := car(args)
		switch t := key.(type) {
		case *lsymbol:
			if !isKeyword(key) {
				key = intern(t.String() + ":")
			}
			if !sliceContains(keys, key) {
				return nil, newError(key, " bad keyword parameter")
			}
			args = args.cdr
			if args == EmptyList {
				return nil, newError(key, " mismatched keyword/value pair in parameter")
			}
			bindings[key] = car(args)
		case *lstruct:
			for k, v := range t.bindings {
				if sliceContains(keys, k) {
					bindings[k] = v
				}
			}
		}
		args = args.cdr
	}
	count = len(bindings)
	if count == 0 {
		return EmptyList, nil
	}
	lst := make([]lob, 0, count*2)
	for k, v := range bindings {
		lst = append(lst, k)
		lst = append(lst, v)
	}
	return toList(lst), nil
}

func newInstance(typesym *lsymbol, fieldvals []lob) (*lstruct, error) {
//	typename, err := typeName(typesym)
//	if err != nil {
	if !isType(typesym) {
		return nil, typeError(typeType, typesym)
	}
	count := len(fieldvals)
	i := 0
	bindings := make(map[lob]lob, count/2) //optimal if all key/value pairs
	for i < count {
		o := fieldvals[i]
		i++
		switch t := o.(type) {
		case lnull:
			//ignore
		case lstring:
			if i == count {
				return nil, newError("mismatched keyword/value in arglist: ", o)
			}
			bindings[o] = fieldvals[i]
			i++
		case *lsymbol:
			if i == count {
				return nil, newError("mismatched keyword/value in arglist: ", o)
			}
			if !isSymbolKeyword(t) {
				o = intern(t.String() + ":")
			}
			bindings[o] = fieldvals[i]
			i++
		case *lstruct:
			for k, v := range t.bindings {
				bindings[k] = v
			}
		default:
			return nil, newError("bad parameter to instance: ", o)
		}
	}
	return &lstruct{typesym, bindings}, nil
}

func newStruct(pairwiseBindings ...lob) (*lstruct, error) {
	return newInstance(typeStruct, pairwiseBindings)
}

func isStruct(obj lob) bool {
	_, ok := obj.(*lstruct)
	return ok
}

func asStruct(o lob) (*lstruct, error) {
	if o == Null {
		return newStruct()
	}
	switch t := o.(type) {
	case *lstruct:
		if t.typesym == typeStruct {
			return t, nil
		}
		return &lstruct{typeStruct, t.bindings}, nil
	}
	return nil, newError("Cannot convert to struct: ", o)
}

func (s *lstruct) Equal(another lob) bool {
	if a, ok := another.(*lstruct); ok {
		if s.typesym == a.typesym {
			slen := len(s.bindings)
			if slen == len(a.bindings) {
				for k, v := range s.bindings {
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
	}
	return false
}

func (s *lstruct) length() int {
	return len(s.bindings)
}

func (s *lstruct) String() string {
	var buf bytes.Buffer
	if s.typesym != typeStruct {
		buf.WriteString("#")
		n, _ := typeName(s.typesym)
		buf.WriteString(n.String())
	}
	buf.WriteString("{")
	first := true
	for k, v := range s.bindings {
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

func (s *lstruct) put(key lob, value lob) lob {
	s.bindings[key] = value
	return s
}

func (s *lstruct) get(key lob) lob {
	if val, ok := s.bindings[key]; ok {
		return val
	}
	return Null
}

func (s *lstruct) has(key lob) bool {
	_, ok := s.bindings[key]
	return ok
}

func has(obj lob, key lob) (bool, error) {
	if aStruct, ok := obj.(*lstruct); ok {
		return aStruct.has(key), nil
	}
	return false, typeError(typeStruct, obj)
}

func get(obj lob, key lob) (lob, error) {
	if aStruct, ok := obj.(*lstruct); ok {
		return aStruct.get(key), nil
	}
	return nil, typeError(typeStruct, obj)
}

func put(obj lob, key lob, value lob) (lob, error) {
	//not clear if I want to export this. Would prefer immutable values
	//i.e. (merge s {x: 23}) or (struct s x: 23)
	if aStruct, ok := obj.(*lstruct); ok {
		return aStruct.put(key, value), nil
	}
	return nil, typeError(typeStruct, obj)
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
