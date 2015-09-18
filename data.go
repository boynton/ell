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
// AnyType is the generic Ell object. It supports querying its symbolic type name at runtime
//
type AnyType interface {
	Type() AnyType
	Equal(another AnyType) bool
	String() string
}

//
// NullType is the type of the null value
//
type NullType int

var typeNull = intern("<null>")

// Null is Ell's version of nil. It means "nothing" and is not the same as EmptyList
const Null = NullType(0)

// Type returns the type of the object
func (NullType) Type() AnyType {
	return typeNull
}

// Equal returns true if the object is equal to the argument
func (NullType) Equal(another AnyType) bool {
	return another == Null
}

func (v NullType) String() string {
	return "null"
}

//
// EOFType is the type of the EOF marker
//
type EOFType int

// EOF is Ell's EOF object
const EOF = EOFType(0)

var typeEOF = intern("<eof>")

// Type returns the type of the object
func (EOFType) Type() AnyType {
	return typeEOF
}

// Equal returns true if the object is equal to the argument
func (EOFType) Equal(another AnyType) bool {
	return another == EOF
}

func (EOFType) String() string {
	return "#eof"
}

//
// BooleanType is the type of true and false
//
type BooleanType bool

//True is Ell's true constant
const True BooleanType = BooleanType(true)

//False is Ell's false constant
const False BooleanType = BooleanType(false)

var typeBoolean = intern("<boolean>")

func isBoolean(obj AnyType) bool {
	_, ok := obj.(BooleanType)
	return ok
}

// Type returns the type of the object
func (BooleanType) Type() AnyType {
	return typeBoolean
}

// Equal returns true if the object is equal to the argument
func (b BooleanType) Equal(another AnyType) bool {
	if a, ok := another.(BooleanType); ok {
		return b == a
	}
	return false
}

func (b BooleanType) String() string {
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

// Type returns the type of the object
func (sym *SymbolType) Type() AnyType {
	if sym.tag == keywordTag {
		return typeKeyword
	} else if sym.tag == typeTag {
		return typeType
	}
	return typeSymbol
}

// Equal returns true if the object is equal to the argument
func (sym *SymbolType) Equal(another AnyType) bool {
	if a, ok := another.(*SymbolType); ok {
		return sym == a
	}
	return false
}

func (sym *SymbolType) String() string {
	return sym.Name
}

func isSymbolKeyword(sym *SymbolType) bool {
	return sym.tag == keywordTag
}

func isSymbolType(sym *SymbolType) bool {
	return sym.tag == typeTag
}

func isSymbol(obj AnyType) bool {
	sym, ok := obj.(*SymbolType)
	return ok && sym.tag >= 0
}

func isType(obj AnyType) bool {
	sym, ok := obj.(*SymbolType)
	return ok && sym.tag == typeTag
}

func isKeyword(obj AnyType) bool {
	sym, ok := obj.(*SymbolType)
	return ok && sym.tag == keywordTag
}

func unKeywordString(sym *SymbolType) string {
	if isSymbolKeyword(sym) {
		return sym.Name[:len(sym.Name)-1]
	}
	return sym.Name
}

func typeName(obj AnyType) (*SymbolType, error) {
	switch t := obj.(type) {
	case *SymbolType:
		if !isSymbolType(t) {
			return nil, TypeError(typeType, obj)
		}
		return intern(t.Name[1 : len(t.Name)-1]), nil
	default:
		return nil, Error("Type error: expected symbol or string, got ", obj)
	}
}

func unkeyword(obj AnyType) (AnyType, error) {
	switch t := obj.(type) {
	case NullType:
		return obj, nil
	case *SymbolType:
		last := len(t.Name) - 1 //can never be less than 0, symbols always have len > 0
		if t.Name[last] != ':' {
			return obj, nil
		}
		return intern(t.Name[:last]), nil
	case StringType:
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
		return nil, Error("Type error: expected symbol or string, got ", obj)
	}
}

//the global symbol table. symbols for the basic types defined in this file are precached
var symtab = map[string]*SymbolType{}

func symbols() []AnyType {
	syms := make([]AnyType, 0, len(symtab))
	for _, sym := range symtab {
		syms = append(syms, sym)
	}
	return syms
}

func symbol(names []AnyType) (AnyType, error) {
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

func isString(obj AnyType) bool {
	_, ok := obj.(StringType)
	return ok
}

func stringValue(obj AnyType) (string, error) {
	switch s := obj.(type) {
	case StringType:
		return string(s), nil
	default:
		return "", TypeError(typeString, obj)
	}
}

// Type returns the type of the object
func (StringType) Type() AnyType {
	return typeString
}

// Equal returns true if the object is equal to the argument
func (s StringType) Equal(another AnyType) bool {
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

func (s StringType) encodedString() string {
	return encodeString(string(s))
}

func (s StringType) String() string {
	return string(s)
}

//
// CharType - Ell characters
//
type CharType rune

var symCharType = intern("<char>")

func isChar(obj AnyType) bool {
	_, ok := obj.(CharType)
	if ok {
		return true
	}
	return ok
}

func newCharacter(c rune) CharType {
	v := CharType(c)
	return v
}

// Type returns the type of the object
func (CharType) Type() AnyType {
	return symCharType
}

// Equal returns true if the object is equal to the argument
func (i CharType) Equal(another AnyType) bool {
	if a, err := intValue(another); err == nil {
		return int(i) == a
	}
	return false
}

func (i CharType) String() string {
	buf := []rune{rune(i)}
	return string(buf)
}

//
// NumberType - Ell numbers
//
type NumberType float64

var typeNumber = intern("<number>")

// Type returns the type of the object
func (NumberType) Type() AnyType {
	return typeNumber
}

func (f NumberType) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func theNumber(obj AnyType) (NumberType, error) {
	if n, ok := obj.(NumberType); ok {
		return n, nil
	}
	return 0, TypeError(typeNumber, obj)
}

func isInt(obj AnyType) bool {
	if n, ok := obj.(NumberType); ok {
		f := float64(n)
		if math.Trunc(f) == f {
			return true
		}
	}
	return false
}

func isFloat(obj AnyType) bool {
	return !isInt(obj)
}

func isNumber(obj AnyType) bool {
	_, ok := obj.(NumberType)
	return ok
}

func floatValue(obj AnyType) (float64, error) {
	switch n := obj.(type) {
	case NumberType:
		return float64(n), nil
	}
	return 0, TypeError(typeNumber, obj)
}

func int64Value(obj AnyType) (int64, error) {
	switch n := obj.(type) {
	case NumberType:
		return int64(n), nil
	case CharType:
		return int64(n), nil
	default:
		return 0, TypeError(typeNumber, obj)
	}
}

func intValue(obj AnyType) (int, error) {
	switch n := obj.(type) {
	case NumberType:
		return int(n), nil
	case CharType:
		return int(n), nil
	default:
		return 0, TypeError(typeNumber, obj)
	}
}

// Equal returns true if the object is equal to the argument
func greaterOrEqual(n1 AnyType, n2 AnyType) (AnyType, error) {
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

func lessOrEqual(n1 AnyType, n2 AnyType) (AnyType, error) {
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

func greater(n1 AnyType, n2 AnyType) (AnyType, error) {
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

func less(n1 AnyType, n2 AnyType) (AnyType, error) {
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

func equal(o1 AnyType, o2 AnyType) bool {
	if o1 == o2 {
		return true
	}
	return o1.Equal(o2)
}

func numericallyEqual(o1 AnyType, o2 AnyType) (bool, error) {
	switch n1 := o1.(type) {
	case NumberType:
		switch n2 := o2.(type) {
		case NumberType:
			return n1 == n2, nil
		default:
			return false, TypeError(typeNumber, o2)
		}
	default:
		return false, TypeError(typeNumber, o1)
	}
}

func identical(n1 AnyType, n2 AnyType) bool {
	return n1 == n2
}

// Equal returns true if the object is equal to the argument
func (f NumberType) Equal(another AnyType) bool {
	if a, err := floatValue(another); err == nil {
		return float64(f) == a
	}
	return false
}

func add(num1 AnyType, num2 AnyType) (AnyType, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return NumberType(n1 + n2), nil
}

func sum(nums []AnyType, argc int) (AnyType, error) {
	var sum float64
	for _, num := range nums {
		switch n := num.(type) {
		case NumberType:
			sum += float64(n)
		default:
			return nil, TypeError(typeNumber, num)
		}
	}
	return NumberType(sum), nil
}

func sub(num1 AnyType, num2 AnyType) (AnyType, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return NumberType(n1 - n2), nil
}

func minus(nums []AnyType, argc int) (AnyType, error) {
	if argc < 1 {
		return ArgcError("-", "1+", argc)
	}
	var fsum float64
	num := nums[0]
	switch n := num.(type) {
	case NumberType:
		fsum = float64(n)
	default:
		return nil, TypeError(typeNumber, num)
	}
	if argc == 1 {
		fsum = -fsum
	} else {
		for _, num := range nums[1:] {
			switch n := num.(type) {
			case NumberType:
				fsum -= float64(n)
			default:
				return nil, TypeError(typeNumber, num)
			}
		}
	}
	return NumberType(fsum), nil
}

func mul(num1 AnyType, num2 AnyType) (AnyType, error) {
	n1, err := floatValue(num1)
	if err != nil {
		return nil, err
	}
	n2, err := floatValue(num2)
	if err != nil {
		return nil, err
	}
	return NumberType(n1 * n2), nil
}

func product(argv []AnyType, argc int) (AnyType, error) {
	prod := 1.0
	for _, num := range argv {
		switch n := num.(type) {
		case NumberType:
			prod *= float64(n)
		default:
			return nil, TypeError(typeNumber, num)
		}
	}
	return NumberType(prod), nil
}

func div(argv []AnyType, argc int) (AnyType, error) {
	if argc < 1 {
		return ArgcError("/", "1+", argc)
	} else if argc == 1 {
		n1, err := floatValue(argv[0])
		if err != nil {
			return nil, err
		}
		return NumberType(1.0 / n1), nil
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
		return NumberType(quo), nil
	}
}

//
// ListType - Ell lists
//
type ListType struct {
	car AnyType
	cdr *ListType
}

var typeList = intern("<list>")

var symList = intern("list")
var symQuote = intern("quote")
var symQuasiquote = intern("quasiquote")
var symUnquote = intern("unquote")
var symUnquoteSplicing = intern("unquote-splicing")

// EmptyList - the value of (), terminates linked lists
var EmptyList = initEmpty()

func initEmpty() *ListType {
	var lst ListType
	return &lst
}

func isEmpty(col AnyType) bool {
	switch v := col.(type) {
	case NullType: //Do I really want this?
		return true
	case StringType:
		return len(v) == 0
	case *ArrayType:
		return len(v.elements) == 0
	case *ListType:
		return v == EmptyList
	case *StructType:
		return len(v.bindings) == 0
	default:
		return false
	}
}

func isList(obj AnyType) bool {
	_, ok := obj.(*ListType)
	return ok
}

// Type returns the type of the object
func (*ListType) Type() AnyType {
	return typeList
}

// Equal returns true if the object is equal to the argument
func (lst *ListType) Equal(another AnyType) bool {
	if a, ok := another.(*ListType); ok {
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

func (lst *ListType) String() string {
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

func (lst *ListType) length() int {
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

func newList(count int, val AnyType) *ListType {
	result := EmptyList
	for i := 0; i < count; i++ {
		result = cons(val, result)
	}
	return result
}

func cons(car AnyType, cdr *ListType) *ListType {
	if car == nil {
		panic("Assertion failure: don't call cons with nil as car")
	}
	if cdr == nil {
		panic("Assertion failure: don't call cons with nil as cdr")
	}
	if inExec {
		conses++
	}
	return &ListType{car, cdr}
}

func car(lst AnyType) AnyType {
	switch p := lst.(type) {
	case *ListType:
		if p != EmptyList {
			return p.car
		}
	}
	return Null
}

func setCar(lst AnyType, obj AnyType) {
	switch p := lst.(type) {
	case *ListType:
		if p != EmptyList {
			p.car = obj
		}
	}
}

func cdr(lst AnyType) *ListType {
	if lst != EmptyList {
		switch p := lst.(type) {
		case *ListType:
			return p.cdr
		}
	}
	return EmptyList
}

func setCdr(lst AnyType, obj AnyType) {
	switch p := lst.(type) {
	case *ListType:
		switch n := obj.(type) {
		case *ListType:
			p.cdr = n
		default:
			println("IGNORED: Setting cdr to non-list: ", obj)
		}
	default:
		println("IGNORED: Setting cdr of non-list: ", lst)
	}
}

func caar(lst AnyType) AnyType {
	return car(car(lst))
}
func cadr(lst AnyType) AnyType {
	return car(cdr(lst))
}
func cddr(lst AnyType) *ListType {
	return cdr(cdr(lst))
}
func caddr(lst AnyType) AnyType {
	return car(cdr(cdr(lst)))
}
func cdddr(lst AnyType) *ListType {
	return cdr(cdr(cdr(lst)))
}
func cadddr(lst AnyType) AnyType {
	return car(cdr(cdr(cdr(lst))))
}
func cddddr(lst AnyType) *ListType {
	return cdr(cdr(cdr(cdr(lst))))
}

func toList(values []AnyType) *ListType {
	p := EmptyList
	for i := len(values) - 1; i >= 0; i-- {
		v := values[i]
		p = cons(v, p)
	}
	return p
}

func list(values ...AnyType) *ListType {
	return toList(values)
}

func listToArray(lst *ListType) *ArrayType {
	var elems []AnyType
	for lst != EmptyList {
		elems = append(elems, lst.car)
		lst = lst.cdr
	}
	return &ArrayType{elems}
}

func arrayToList(ary AnyType) (AnyType, error) {
	v, ok := ary.(*ArrayType)
	if !ok {
		return nil, TypeError(typeArray, ary)
	}
	return toList(v.elements), nil
}

func length(seq AnyType) int {
	switch v := seq.(type) {
	case StringType:
		return len(v)
	case *ArrayType:
		return len(v.elements)
	case *ListType:
		return v.length()
	case *StructType:
		return len(v.bindings)
	default:
		return -1
	}
}

func reverse(lst *ListType) *ListType {
	rev := EmptyList
	for lst != EmptyList {
		rev = cons(lst.car, rev)
		lst = lst.cdr
	}
	return rev
}

func concat(seq1 *ListType, seq2 *ListType) (*ListType, error) {
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
// ArrayType - Ell Arrays
//
type ArrayType struct {
	elements []AnyType
}

var typeArray = intern("<array>")

func isArray(obj AnyType) bool {
	_, ok := obj.(*ArrayType)
	return ok
}

// Type returns the type of the object
func (*ArrayType) Type() AnyType {
	return typeArray
}

// Equal returns true if the object is equal to the argument
func (ary *ArrayType) Equal(another AnyType) bool {
	if a, ok := another.(*ArrayType); ok {
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

func (ary *ArrayType) String() string {
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

func newArray(size int, init AnyType) *ArrayType {
	elements := make([]AnyType, size)
	for i := 0; i < size; i++ {
		elements[i] = init
	}
	return &ArrayType{elements}
}

func array(elements ...AnyType) AnyType {
	size := len(elements)
	el := make([]AnyType, size)
	for i := 0; i < size; i++ {
		el[i] = elements[i]
	}
	return &ArrayType{el}
}

func toArray(elements []AnyType, count int) AnyType {
	el := make([]AnyType, count)
	copy(el, elements[0:count])
	return &ArrayType{el}
}

func (ary *ArrayType) length() int {
	return len(ary.elements)
}

func arrayLength(ary AnyType) (int, error) {
	if a, ok := ary.(*ArrayType); ok {
		return len(a.elements), nil
	}
	return 1, TypeError(typeArray, ary)
}

func arraySet(ary AnyType, idx int, obj AnyType) error {
	if a, ok := ary.(*ArrayType); ok {
		if idx < 0 || idx >= len(a.elements) {
			return Error("Array index out of range")
		}
		a.elements[idx] = obj
		return nil
	}
	return TypeError(typeArray, ary)
}

func arrayRef(ary AnyType, idx int) (AnyType, error) {
	if a, ok := ary.(*ArrayType); ok {
		if idx < 0 || idx >= len(a.elements) {
			return nil, Error("Array index out of range")
		}
		return a.elements[idx], nil
	}
	return nil, TypeError(typeArray, ary)
}

//
// StructType - Ell structs (objects). They are extensible, having a special type symbol in them.
//
type StructType struct {
	typesym  *SymbolType
	bindings map[AnyType]AnyType
}

var typeStruct = intern("<struct>")

// Type returns the type of the object
func (s *StructType) Type() AnyType {
	return s.typesym
}

func sliceContains(slice []AnyType, obj AnyType) bool {
	for _, o := range slice {
		if o == obj {
			return true
		}
	}
	return false
}

func normalizeKeywordArgs(args *ListType, keys []AnyType) (*ListType, error) {
	count := length(args)
	bindings := make(map[AnyType]AnyType, count/2)
	for args != EmptyList {
		key := car(args)
		switch t := key.(type) {
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
		case *StructType:
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
	lst := make([]AnyType, 0, count*2)
	for k, v := range bindings {
		lst = append(lst, k)
		lst = append(lst, v)
	}
	return toList(lst), nil
}

func newInstance(typesym *SymbolType, fieldvals []AnyType) (*StructType, error) {
	if !isType(typesym) {
		return nil, TypeError(typeType, typesym)
	}
	count := len(fieldvals)
	i := 0
	bindings := make(map[AnyType]AnyType, count/2) //optimal if all key/value pairs
	for i < count {
		o := fieldvals[i]
		i++
		switch t := o.(type) {
		case NullType:
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
			if !isSymbolKeyword(t) {
				o = intern(t.String() + ":")
			}
			bindings[o] = fieldvals[i]
			i++
		case *StructType:
			for k, v := range t.bindings {
				bindings[k] = v
			}
		default:
			return nil, Error("bad parameter to instance: ", o)
		}
	}
	return &StructType{typesym, bindings}, nil
}

func newStruct(pairwiseBindings ...AnyType) (*StructType, error) {
	return newInstance(typeStruct, pairwiseBindings)
}

func isStruct(obj AnyType) bool {
	_, ok := obj.(*StructType)
	return ok
}

func asStruct(o AnyType) (*StructType, error) {
	if o == Null {
		return newStruct()
	}
	switch t := o.(type) {
	case *StructType:
		if t.typesym == typeStruct {
			return t, nil
		}
		return &StructType{typeStruct, t.bindings}, nil
	}
	return nil, Error("Cannot convert to struct: ", o)
}

// Equal returns true if the object is equal to the argument
func (s *StructType) Equal(another AnyType) bool {
	if a, ok := another.(*StructType); ok {
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

func (s *StructType) length() int {
	return len(s.bindings)
}

func (s *StructType) String() string {
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

func (s *StructType) put(key AnyType, value AnyType) AnyType {
	s.bindings[key] = value
	return s
}

func (s *StructType) get(key AnyType) AnyType {
	if val, ok := s.bindings[key]; ok {
		return val
	}
	return Null
}

func (s *StructType) has(key AnyType) bool {
	_, ok := s.bindings[key]
	return ok
}

func has(obj AnyType, key AnyType) (bool, error) {
	if aStruct, ok := obj.(*StructType); ok {
		return aStruct.has(key), nil
	}
	return false, TypeError(typeStruct, obj)
}

func get(obj AnyType, key AnyType) (AnyType, error) {
	if aStruct, ok := obj.(*StructType); ok {
		return aStruct.get(key), nil
	}
	return nil, TypeError(typeStruct, obj)
}

func put(obj AnyType, key AnyType, value AnyType) (AnyType, error) {
	//not clear if I want to export this. Would prefer immutable values
	//i.e. (merge s {x: 23}) or (struct s x: 23)
	if aStruct, ok := obj.(*StructType); ok {
		return aStruct.put(key, value), nil
	}
	return nil, TypeError(typeStruct, obj)
}

func structToList(obj AnyType) (AnyType, error) {
	if aStruct, ok := obj.(*StructType); ok {
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
	if l, ok := arg1.(AnyType); ok {
		buf.WriteString(fmt.Sprintf("%v", write(l)))
	} else {
		buf.WriteString(fmt.Sprintf("%v", arg1))
		}
	for _, o := range args {
		if l, ok := o.(AnyType); ok {
			buf.WriteString(fmt.Sprintf("%v", write(l)))
		} else {
			buf.WriteString(fmt.Sprintf("%v", o))
		}
	}
	err := GenericError{buf.String()}
	return &err
}

// TypeError - an error indicating expected and actual value for a type mismatch
func TypeError(typeSym AnyType, obj AnyType) error {
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
