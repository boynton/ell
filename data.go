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

//
// LAny is the generic Ell object. It supports querying its symbolic type name at runtime
//
type LAny interface {
	Type() LAny
	Value() LAny
	Equal(another LAny) bool
	String() string
}

//
// LNull is the type of the null value
//
type LNull int

var typeNull = intern("<null>")

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList
const Null = LNull(0)

// Type returns the type of the object
func (LNull) Type() LAny {
	return typeNull
}

// Value returns the object itself for primitive types
func (LNull) Value() LAny {
	return Null
}

// Equal returns true if the object is equal to the argument
func (LNull) Equal(another LAny) bool {
	return another == Null
}

func (v LNull) String() string {
	return "null"
}

//
// LEOF is the type of the EOF marker
//
type LEOF int

// EOF is Ell's EOF object
const EOF = LEOF(0)

var typeEOF = intern("<eof>")

// Type returns the type of the object
func (LEOF) Type() LAny {
	return typeEOF
}

// Value returns the object itself for primitive types
func (LEOF) Value() LAny {
	return EOF
}

// Equal returns true if the object is equal to the argument
func (LEOF) Equal(another LAny) bool {
	return another == EOF
}

func (LEOF) String() string {
	return "#eof"
}

//
// LBoolean is the type of true and false
//
type LBoolean bool

//True is Ell's true constant
const True LBoolean = LBoolean(true)

//False is Ell's false constant
const False LBoolean = LBoolean(false)

var typeBoolean = intern("<boolean>")

func isBoolean(obj LAny) bool {
	_, ok := obj.(LBoolean)
	return ok
}

// Type returns the type of the object
func (LBoolean) Type() LAny {
	return typeBoolean
}

// Value returns the object itself for primitive types
func (b LBoolean) Value() LAny {
	return b
}

// Equal returns true if the object is equal to the argument
func (b LBoolean) Equal(another LAny) bool {
	if a, ok := another.(LBoolean); ok {
		return b == a
	}
	return false
}

func (b LBoolean) String() string {
	return strconv.FormatBool(bool(b))
}

//
// SymbolType holds symbols, keywords, and types. Use the tag to distinguish between them
//
type SymbolType struct {
	Name string
	tag  int //an incrementing sequence number for symbols, -1 for types, and -2 for keywords
}

const typeTag = -1
const keywordTag = -2

var symtag int

func intern(name string) *SymbolType {
	sym, ok := symtab[name]
	if !ok {
		if isValidKeywordName(name) {
			sym = &SymbolType{name, keywordTag}
		} else if isValidTypeName(name) {
			sym = &SymbolType{name, typeTag}
		} else if isValidSymbolName(name) {
			sym = &SymbolType{name, symtag}
			symtag++
		}
		if sym == nil {
			panic("invalid symbol/type/keyword name passed to intern: " + name)
		}
		symtab[name] = sym
	}
	return sym
}

func isValidSymbolName(name string) bool {
	return len(name) > 0
}

func isValidTypeName(s string) bool {
	n := len(s)
	if n > 2 && s[0] == '<' && s[n-1] == '>' {
		return true
	}
	return false
}

func isValidKeywordName(s string) bool {
	n := len(s)
	if n > 1 && s[n-1] == ':' {
		return true
	}
	return false
}

var typeSymbol = intern("<symbol>")
var typeKeyword = intern("<keyword>")
var typeType = intern("<type>")

// Type returns the type of the object. Since SymbolType represents keywords, types, and regular
// symbols, it could return any of those three values
func (sym *SymbolType) Type() LAny {
	if sym.tag == keywordTag {
		return typeKeyword
	} else if sym.tag == typeTag {
		return typeType
	}
	return typeSymbol
}

// Value returns the object itself for primitive types
func (sym *SymbolType) Value() LAny {
	return sym
}

// Equal returns true if the object is equal to the argument
func (sym *SymbolType) Equal(another LAny) bool {
	if a, ok := another.(*SymbolType); ok {
		return sym == a
	}
	return false
}

func (sym *SymbolType) String() string {
	return sym.Name
}

func isSymbol(obj LAny) bool {
	sym, ok := obj.(*SymbolType)
	return ok && sym.tag >= 0
}

func isType(obj LAny) bool {
	sym, ok := obj.(*SymbolType)
	return ok && sym.tag == typeTag
}

func isKeyword(obj LAny) bool {
	sym, ok := obj.(*SymbolType)
	return ok && sym.tag == keywordTag
}

func typeName(obj LAny) (*SymbolType, error) {
	sym, ok := obj.(*SymbolType)
	if ok && sym.tag == typeTag {
		return intern(sym.Name[1 : len(sym.Name)-1]), nil
	}
	return nil, Error("Type error: expected <type>, got ", obj)
}

func unkeywordedString(sym *SymbolType) string {
	if sym.tag == keywordTag {
		return sym.Name[:len(sym.Name)-1]
	}
	return sym.Name
}

func unkeyworded(obj LAny) (LAny, error) {
	sym, ok := obj.(*SymbolType)
	if ok {
		switch sym.tag {
		case keywordTag:
			return intern(sym.Name[:len(sym.Name)-1]), nil
		case typeTag:
			//nothing
		default: //already a regular symbol
			return obj, nil
		}
	}
	return nil, Error("Type error: expected <keyword> or <symbol>, got ", obj)
}

func keywordToSymbol(obj LAny) (LAny, error) {
	sym, ok := obj.(*SymbolType)
	if ok && sym.tag == keywordTag {
		return intern(sym.Name[:len(sym.Name)-1]), nil
	}
	return nil, Error("Type error: expected <keyword>, got ", obj)
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]*SymbolType{}

func symbols() []LAny {
	syms := make([]LAny, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func symbol(names []LAny) (LAny, error) {
	size := len(names)
	if size < 1 {
		return ArgcError("symbol", "1+", size)
	}
	name := ""
	for i := 0; i < size; i++ {
		o := names[i]
		s := ""
		switch t := o.(type) {
		case StringType:
			s = string(t)
		case *SymbolType:
			s = t.Name
		default:
			return nil, Error("symbol name component invalid: ", o)
		}
		name += s
	}
	return intern(name), nil
}

//
// StringType - Ell Strings
//
type StringType string

var typeString = intern("<string>")

func isString(obj LAny) bool {
	_, ok := obj.(StringType)
	return ok
}

func stringValue(obj LAny) (string, error) {
	switch s := obj.(type) {
	case StringType:
		return string(s), nil
	default:
		return "", TypeError(typeString, obj)
	}
}

// Type returns the type of the object
func (StringType) Type() LAny {
	return typeString
}

// Value returns the object itself for primitive types
func (s StringType) Value() LAny {
	return s
}

// Equal returns true if the object is equal to the argument
func (s StringType) Equal(another LAny) bool {
	if a, ok := another.(StringType); ok {
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

func (s StringType) String() string {
	return string(s)
}

//
// LChar - Ell characters
//
type LChar rune

var typeChar = intern("<char>")

func isChar(obj LAny) bool {
	_, ok := obj.(LChar)
	if ok {
		return true
	}
	return ok
}

func newCharacter(c rune) LChar {
	v := LChar(c)
	return v
}

// Type returns the type of the object
func (LChar) Type() LAny {
	return typeChar
}

// Value returns the object itself for primitive types
func (i LChar) Value() LAny {
	return i
}

// Equal returns true if the object is equal to the argument
func (i LChar) Equal(another LAny) bool {
	if a, err := intValue(another); err == nil {
		return int(i) == a
	}
	return false
}

func (i LChar) String() string {
	buf := []rune{rune(i)}
	return string(buf)
}

//
// LNumber - Ell numbers
//
type LNumber float64

var typeNumber = intern("<number>")

// Type returns the type of the object
func (LNumber) Type() LAny {
	return typeNumber
}

// Value returns the object itself for primitive types
func (f LNumber) Value() LAny {
	return f
}

func (f LNumber) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func theNumber(obj LAny) (LNumber, error) {
	if n, ok := obj.(LNumber); ok {
		return n, nil
	}
	return 0, TypeError(typeNumber, obj)
}

func isInt(obj LAny) bool {
	if n, ok := obj.(LNumber); ok {
		f := float64(n)
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj LAny) bool {
	return !isInt(obj)
}

func isNumber(obj LAny) bool {
	_, ok := obj.(LNumber)
	return ok
}

func floatValue(obj LAny) (float64, error) {
	switch n := obj.(type) {
	case LNumber:
		return float64(n), nil
	}
	return 0, TypeError(typeNumber, obj)
}

func int64Value(obj LAny) (int64, error) {
	switch n := obj.(type) {
	case LNumber:
		return int64(n), nil
	case LChar:
		return int64(n), nil
	default:
		return 0, TypeError(typeNumber, obj)
	}
}

func intValue(obj LAny) (int, error) {
	switch n := obj.(type) {
	case LNumber:
		return int(n), nil
	case LChar:
		return int(n), nil
	default:
		return 0, TypeError(typeNumber, obj)
	}
}

// Equal returns true if the object is equal to the argument
func greaterOrEqual(n1 LAny, n2 LAny) (LAny, error) {
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

func lessOrEqual(n1 LAny, n2 LAny) (LAny, error) {
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

func greater(n1 LAny, n2 LAny) (LAny, error) {
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

func less(n1 LAny, n2 LAny) (LAny, error) {
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

func equal(o1 LAny, o2 LAny) bool {
	if o1 == o2 {
		return true
	}
	return o1.Equal(o2)
}

func numericallyEqual(o1 LAny, o2 LAny) (bool, error) {
	switch n1 := o1.(type) {
	case LNumber:
		switch n2 := o2.(type) {
		case LNumber:
			return n1 == n2, nil
		default:
			return false, TypeError(typeNumber, o2)
		}
	default:
		return false, TypeError(typeNumber, o1)
	}
}

func identical(n1 LAny, n2 LAny) bool {
	return n1 == n2
}

// Equal returns true if the object is equal to the argument
func (f LNumber) Equal(another LAny) bool {
	if a, err := floatValue(another); err == nil {
		return float64(f) == a
	}
	return false
}

func add(num1 LAny, num2 LAny) (LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return LNumber(n1 + n2), nil
}

func sum(nums []LAny, argc int) (LAny, error) {
	var sum float64
	for _, num := range nums {
		switch n := num.(type) {
		case LNumber:
			sum += float64(n)
		default:
			return nil, TypeError(typeNumber, num)
		}
	}
	return LNumber(sum), nil
}

func sub(num1 LAny, num2 LAny) (LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return LNumber(n1 - n2), nil
}

func minus(nums []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return ArgcError("-", "1+", argc)
	}
	var fsum float64
	num := nums[0]
	switch n := num.(type) {
	case LNumber:
		fsum = float64(n)
	default:
		return nil, TypeError(typeNumber, num)
	}
	if argc == 1 {
		fsum = -fsum
	} else {
		for _, num := range nums[1:] {
			switch n := num.(type) {
			case LNumber:
				fsum -= float64(n)
			default:
				return nil, TypeError(typeNumber, num)
			}
		}
	}
	return LNumber(fsum), nil
}

func mul(num1 LAny, num2 LAny) (LAny, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return LNumber(n1 * n2), nil
}

func product(argv []LAny, argc int) (LAny, error) {
	prod := 1.0
	for _, num := range argv {
		switch n := num.(type) {
		case LNumber:
			prod *= float64(n)
		default:
			return nil, TypeError(typeNumber, num)
		}
	}
	return LNumber(prod), nil
}

func div(argv []LAny, argc int) (LAny, error) {
	if argc < 1 {
		return ArgcError("/", "1+", argc)
	} else if argc == 1 {
		n1, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		return LNumber(1.0 / n1), nil
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
		return LNumber(quo), nil
	}
}

//
// LList - Ell lists
//
type LList struct {
	car LAny
	cdr *LList
}

var typeList = intern("<list>")

var symList = intern("list")
var symQuote = intern("quote")
var symQuasiquote = intern("quasiquote")
var symUnquote = intern("unquote")
var symUnquoteSplicing = intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *LList {
	var lst LList
	return &lst
}

func isEmpty(col LAny) bool {
	switch v := col.(type) {
	case LNull: //Do I really want this?
		return true
	case StringType:
		return len(v) == 0
	case *LArray:
		return len(v.elements) == 0
	case *LList:
		return v == EmptyList
	case *LStruct:
		return len(v.bindings) == 0
	default:
		return false
	}
}

func isList(obj LAny) bool {
	_, ok := obj.(*LList)
	return ok
}

// Type returns the type of the object
func (*LList) Type() LAny {
	return typeList
}

// Value returns the object itself for primitive types
func (lst *LList) Value() LAny {
	return lst
}

// Equal returns true if the object is equal to the argument
func (lst *LList) Equal(another LAny) bool {
	if a, ok := another.(*LList); ok {
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

func (lst *LList) String() string {
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

func listLength(lst *LList) int {
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

func newList(count int, val LAny) *LList {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func cons(car LAny, cdr *LList) *LList {
	if car == nil {
		panic("Assertion failure: don't call cons with nil as car")
	}
	if cdr == nil {
		panic("Assertion failure: don't call cons with nil as cdr")
	}
	if inExec {
		conses++
	}
	return &LList{car, cdr}
}

func car(lst LAny) LAny {
	switch p := lst.(type) {
	case *LList:
		if p != EmptyList {
			return p.car
		}
	}
	return Null
}

func setCar(lst LAny, obj LAny) {
	switch p := lst.(type) {
	case *LList:
		if p != EmptyList {
			p.car = obj
		}
	}
}

func cdr(lst LAny) *LList {
	if lst != EmptyList {
		switch p := lst.(type) {
		case *LList:
			return p.cdr
		}
	}
	return EmptyList
}

func setCdr(lst LAny, obj LAny) {
	switch p := lst.(type) {
	case *LList:
		switch n := obj.(type) {
		case *LList:
			p.cdr = n
		default:
			println("IGNORED: Setting cdr to non-list: ", obj)
		}
	default:
		println("IGNORED: Setting cdr of non-list: ", lst)
	}
}

func caar(lst LAny) LAny {
	return car(car(lst))
}
func cadr(lst LAny) LAny {
	return car(cdr(lst))
}
func cddr(lst LAny) *LList {
	return cdr(cdr(lst))
}
func caddr(lst LAny) LAny {
	return car(cdr(cdr(lst)))
}
func cdddr(lst LAny) *LList {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst LAny) LAny {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst LAny) *LList {
	return cdr(cdr(cdr(cdr(lst))))
}

func toList(values []LAny) *LList {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = cons(v, p)
	}
	return p
}

func list(values ...LAny) *LList {
	return toList(values)
}

func listToArray(lst *LList) *LArray {
	var elems []LAny
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return &LArray{elems}
}

func arrayToList(ary LAny) (LAny, error) {
	v, ok := ary.(*LArray)
	if !ok {
		return nil, TypeError(typeArray, ary)
	}
	return toList(v.elements), nil
}

func length(seq LAny) int {
	switch v := seq.Value().(type) {
	case StringType:
		return len(v)
	case *LArray:
		return len(v.elements)
	case *LList:
		return listLength(v)
	case *LStruct:
		return len(v.bindings)
	default:
		return -1
	}
}

func assoc(seq LAny, key LAny, val LAny) (LAny, error) {
	switch s := seq.(type) {
	case *LStruct:
		s2 := copyStruct(s)
		s2.bindings[key] = val
		return s2, nil
	case *LArray:
		if idx, ok := key.(LNumber); ok {
			a := copyArray(s)
			a.elements[int(idx)] = val
			return a, nil
		}
		return nil, TypeError(typeNumber, key)
	default:
		return nil, Error("Cannot assoc with this value: ", seq)
	}
}

func dissoc(seq LAny, key LAny) (LAny, error) {
	switch s := seq.(type) {
	case *LStruct:
		s2 := copyStruct(s)
		delete(s2.bindings, key)
		return s2, nil
	default:
		return nil, Error("Cannot dissoc with this value: ", seq)
	}
}

func reverse(lst *LList) *LList {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func concat(seq1 *LList, seq2 *LList) (*LList, error) {
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
// LArray - Ell Arrays
//
type LArray struct {
	elements []LAny
}

var typeArray = intern("<array>")

func isArray(obj LAny) bool {
	_, ok := obj.(*LArray)
	return ok
}

// Type returns the type of the object
func (*LArray) Type() LAny {
	return typeArray
}

// Value returns the object itself for primitive types
func (ary *LArray) Value() LAny {
	return ary
}

// Equal returns true if the object is equal to the argument
func (ary *LArray) Equal(another LAny) bool {
	if a, ok := another.(*LArray); ok {
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

func (ary *LArray) String() string {
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

func newArray(size int, init LAny) *LArray {
	elements := make([]LAny, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return &LArray{elements}
}

func array(elements ...LAny) LAny {
	return toArray(elements, len(elements))
}

func toArray(elements []LAny, count int) LAny {
	el := make([]LAny, count)
	copy(el, elements[0:count])
	return &LArray{el}
}

func copyArray(a *LArray) *LArray {
	elements := make([]LAny, len(a.elements))
	copy(elements, a.elements)
	return &LArray{elements}
}

func arrayLength(ary LAny) (int, error) {
	if a, ok := ary.(*LArray); ok {
		return len(a.elements), nil
	}
	return 1, TypeError(typeArray, ary)
}

func arraySet(ary LAny, idx int, obj LAny) error {
	if a, ok := ary.(*LArray); ok {
		if idx < 0 || idx >= len(a.elements) {
			return Error("Array index out of range")
		}
		a.elements[idx] = obj
		return nil
	}
	return TypeError(typeArray, ary)
}

func arrayRef(ary LAny, idx int) (LAny, error) {
	if a, ok := ary.(*LArray); ok {
		if idx < 0 || idx >= len(a.elements) {
			return nil, Error("Array index out of range")
		}
		return a.elements[idx], nil
	}
	return nil, TypeError(typeArray, ary)
}

// LInstance is a typed value
type LInstance struct {
	tag   *SymbolType
	value LAny
}

func instance(tag LAny, val LAny) (LAny, error) {
	sym, ok := tag.(*SymbolType)
	if !ok || !isValidTypeName(sym.Name) {
		return nil, TypeError(typeType, tag)
	}
	switch sym {
	case typeString, typeNumber, typeNull, typeBoolean, typeChar, typeEOF:
		return val, nil
	case typeStruct, typeList, typeArray, typeSymbol, typeFunction, typeInput, typeOutput:
		return val, nil
	default:
		return &LInstance{tag: sym, value: val}, nil
	}
}

// Type returns the type of the object
func (s *LInstance) Type() LAny {
	return s.tag
}

// Value returns the value of the object
func (s *LInstance) Value() LAny {
	return s.value
}

// Equal returns true if the object is equal to the argument
func (s *LInstance) Equal(another LAny) bool {
	if a, ok := another.(*LInstance); ok {
		return s.tag == a.tag && s.value.Equal(a.value)
	}
	return false
}

// String of a instance, i.e. #<point>{x: 1 y: 2} or #<uuid>"0bbbc94a-5e14-11e5-81e6-003ee1be85f9"
func (s *LInstance) String() string {
	return "#" + s.tag.String() + write(s.value)
}

//
// LStruct - Ell structs (objects). They are extensible, having a special type symbol in them.
//
type LStruct struct {
	bindings map[LAny]LAny
}

var typeStruct = intern("<struct>")

// Type returns the type of the object
func (s *LStruct) Type() LAny {
	return typeStruct
}

// Value returns the object itself for primitive types
func (s *LStruct) Value() LAny {
	return s
}

func sliceContains(slice []LAny, obj LAny) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func normalizeKeywordArgs(args *LList, keys []LAny) (*LList, error) {
	count := length(args)
	bindings := make(map[LAny]LAny, count/2)
	for args != EmptyList {
		key := car(args)
		switch t := key.Value().(type) {
		case *SymbolType:
			if !isKeyword(key) {
				key = intern(t.String() + ":")
			}
			if !sliceContains(keys, key) {
				return nil, Error(key, " bad keyword parameter")
			}
			args = args.cdr
			if args == EmptyList {
				return nil, Error(key, " mismatched keyword/value pair in parameter")
			}
			bindings[key] = car(args)
		case *LStruct:
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
	lst := make([]LAny, 0, count*2)
	for k, v := range bindings {
		lst = append(lst, k)
		lst = append(lst, v)
	}
	return toList(lst), nil
}

func copyStruct(s *LStruct) *LStruct {
	bindings := make(map[LAny]LAny, len(s.bindings))
	for k, v := range s.bindings {
		bindings[k] = v
	}
	return &LStruct{bindings}
}

func newStruct(fieldvals []LAny) (*LStruct, error) {
	count := len(fieldvals)
	i := 0
	bindings := make(map[LAny]LAny, count/2) //optimal if all key/value pairs
	for i < count {
		o := fieldvals[i]
		i++
		switch t := o.Value().(type) {
		case LNull:
			//ignore
		case StringType:
			if i == count {
				return nil, Error("mismatched keyword/value in arglist: ", o)
			}
			bindings[o] = fieldvals[i]
			i++
		case *SymbolType:
			if i == count {
				return nil, Error("mismatched keyword/value in arglist: ", o)
			}
			bindings[o] = fieldvals[i]
			i++
		case *LStruct:
			for k, v := range t.bindings {
				bindings[k] = v
			}
		default:
			return nil, Error("bad parameter to instance: ", o)
		}
	}
	return &LStruct{bindings}, nil
}

func isStruct(obj LAny) bool {
	_, ok := obj.(*LStruct)
	return ok
}

// Equal returns true if the object is equal to the argument
func (s *LStruct) Equal(another LAny) bool {
	if a, ok := another.(*LStruct); ok {
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
	return false
}

func (s *LStruct) String() string {
	var buf bytes.Buffer
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

func has(obj LAny, key LAny) (bool, error) {
	o := obj.Value()
	if s, ok := o.(*LStruct); ok {
		_, ok := s.bindings[key]
		return ok, nil
	}
	return false, TypeError(typeStruct, obj)
}

func get(obj LAny, key LAny) (LAny, error) {
	o := obj.Value()
	if s, ok := o.(*LStruct); ok {
		if val, ok := s.bindings[key]; ok {
			return val, nil
		}
		return Null, nil
	}
	return nil, TypeError(typeStruct, obj)
}

func put(obj LAny, key LAny, value LAny) (LAny, error) {
	if aStruct, ok := obj.(*LStruct); ok {
		aStruct.bindings[key] = value
		return aStruct, nil
	}
	return nil, TypeError(typeStruct, obj)
}

func structToList(obj LAny) (LAny, error) {
	if aStruct, ok := obj.(*LStruct); ok {
		result := EmptyList
		tail := EmptyList
		for k, v := range aStruct.bindings {
			tmp := list(k, v)
			if result == EmptyList {
				result = list(tmp)
				tail = result
			} else {
				tail.cdr = list(tmp)
				tail = tail.cdr
			}
		}
		return result, nil
	}
	return nil, TypeError(typeStruct, obj)
}

//
// Error - creates a new Error from the arguments
//
func Error(arg1 interface{}, args ...interface{}) error {
	var buf bytes.Buffer
	if l, ok := arg1.(LAny); ok {
		buf.WriteString(fmt.Sprintf("%v", write(l)))
	} else {
		buf.WriteString(fmt.Sprintf("%v", arg1))
	}
	for _, o := range args {
		if l, ok := o.(LAny); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	err := GenericError{buf.String()}
	return &err
}

// TypeError - an error indicating expected and actual value for a type mismatch
func TypeError(typeSym LAny, obj LAny) error {
	return Error("Type error: expected ", typeSym, ", got ", obj)
}

// GenericError - most Ell errors are one of these
type GenericError struct {
	msg string
}

func (e *GenericError) Error() string {
	return e.msg
}

func (e *GenericError) String() string {
	return fmt.Sprintf("<Error: %s>", e.msg)
}
