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
// The basic Ell object, which can be queried for its symbolic type
//
type LObject interface {
	Type() LSymbol
	String() string
}

//
// ------------------- nil
//

type LNil int

// here is the single representation of nil.
const NIL = LNil(0)

var symNil = LSymbol{"nil"}

func (LNil) Type() LSymbol {
	return symNil
}

//nil can be considered a sequence of length 0
func (LNil) Length() int {
	return 0
}

// the external representation is nil, same as the java implementation. Confusing in Go, sorry!
func (LNil) String() string {
	return "nil"
}
func IsNil(obj LObject) bool {
	return obj == NIL
}

//
// ------------------- end-of-information marker
//

type LEOI int

// here is the single representation of nil.
const EOI = LEOI(0)

// nil has type nil
var symEoi = LSymbol{"eoi"}

func (LEOI) Type() LSymbol {
	return symEoi
}
func (LEOI) String() string {
	return "<end-of-input>"
}
func IsEOI(obj LObject) bool {
	return obj == EOI
}

//
// ------------------- boolean
//

type LBoolean bool

var symBoolean = LSymbol{"boolean"}

const TRUE LBoolean = LBoolean(true)
const FALSE LBoolean = LBoolean(false)

func (LBoolean) Type() LSymbol {
	return symBoolean
}
func IsBoolean(obj LObject) bool {
	return obj.Type() == symBoolean
}
func (b LBoolean) String() string {
	return strconv.FormatBool(bool(b))
}

//
// ------------------- symbol
//

type LSymbol struct {
	Name string
}

var symSymbol = LSymbol{"symbol"}

func IsSymbol(obj LObject) bool {
	return obj.Type() == symSymbol
}
func (LSymbol) Type() LSymbol {
	return symSymbol
}
func (sym LSymbol) String() string {
	return sym.Name
}

var symtab = map[string]LSymbol {
	"nil":     symNil,
	"boolean": symBoolean,
	"symbol":  symSymbol,
	"keyword":  symKeyword,
	"string":  symString,
	"number":  symNumber,
	"list":  symList,
	"vector":  symVector,
	"map":  symMap,
	"eoi": symEoi,
	"error": symError,
}

func Intern(name string) LSymbol {
	//to do: validate the symbol name, based on EllDN spec
	v, ok := symtab[name]
	if !ok {
		v = LSymbol{name}
		symtab[name] = v
	}
	return v
}

//
// ------------------- keyword
//

type LKeyword struct {
	sym LSymbol
}

var symKeyword = LSymbol{"keyword"}

func IsKeyword(obj LObject) bool {
	return obj.Type() == symKeyword
}
func (LKeyword) Type() LSymbol {
	return symKeyword
}
func (key LKeyword) String() string {
	return key.sym.Name + ":"
}

//
// ------------------- string
//

type LString string

var symString = LSymbol{"string"}

func IsString(obj LObject) bool {
	return obj.Type() == symString
}
func (LString) Type() LSymbol {
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

func (s LString) String() string {
	return encodeString(string(s))
}

//
// ------------------- number
//

type LNumber float64

var symNumber = LSymbol{"number"}

func IsNumber(obj LObject) bool {
	return obj.Type() == symNumber
}
func (LNumber) Type() LSymbol {
	return symNumber
}

func (n LNumber) String() string {
	return strconv.FormatFloat(float64(n), 'f', -1, 64)
}

//
// ------------------- list
//
type LList struct {
	car LObject
	cdr LObject
}

var symList = LSymbol{"list"}

func IsList(obj LObject) bool {
	return obj.Type() == symList
}
func (LList) Type() LSymbol {
	return symList
}

func (lst LList) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(lst.car.String())
	var tail LObject = lst.cdr
	b := true
	for b {
		if tail == NIL {
			b = false
		} else if IsList(tail) {
			lst = tail.(LList)
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

func (lst LList) Length() int {
	if IsNil(lst) {
		return 0
	}
	count := 1
	var o LObject = lst.cdr
	for o != NIL {
		if !IsList(o) {
			return -1 //not a proper list
		}
		count++
		o = o.(LList).cdr
	}
	return count
}

func Cons(car LObject, cdr LObject) LList {
	return LList{car, cdr}
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

type LSequence interface {
	Length() int
}

func Length(seq LSequence) int {
	return seq.Length()
	/*
		if seq == NIL {
			return 0;
		} else if IsList(seq) {
			return seq.(LList).Length()
		} else if IsList(seq) {
			return seq.(LList).Length()
		} else {
			return -1;
		}
	*/
}

//
// ------------------- vector
//
type LVector struct {
	elements []LObject
}

var symVector = LSymbol{"vector"}

func IsVector(obj LObject) bool {
	return obj.Type() == symVector
}
func (LVector) Type() LSymbol {
	return symVector
}
func Vector(vec ...LObject) LVector {
	return LVector{vec}
}
func (vec LVector) Length() int {
	return len(vec.elements)
}
func (vec LVector) String() string {
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
func (vec LVector) Ref(idx int) LObject {
	return vec.elements[idx]
}
func Ref(vec LVector, idx int) LObject {
	return vec.elements[idx]
}

//
// ------------------- map
//
type LMap struct {
	bindings map[LObject]LObject
}

var symMap = LSymbol{"map"}

func IsMap(obj LObject) bool {
	return obj.Type() == symMap
}
func (LMap) Type() LSymbol {
	return symMap
}
func (m LMap) Length() int {
	return len(m.bindings)
}
func (m LMap) String() string {
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
	m := LMap{map[LObject]LObject{}}
	for i := 0; i < count; i += 2 {
		m.Put(pairwiseBindings[i], pairwiseBindings[i+1])
	}
	return m
}

func (m LMap) Put(key LObject, value LObject) LMap {
	m.bindings[key] = value
	Println("map.PUT:", key, value)
	return m
}

func (m LMap) Get(key LObject) LObject {
	val, ok := m.bindings[key]
	if ok {
		return val
	} else {
		return NIL
	}
}

func (m LMap) Has(key LObject) bool {
	_, ok := m.bindings[key]
	return ok
}

//
// ------------------- error
//

type LError struct {
	msg string
}

var symError = LSymbol{"error"}

func (e LError) Error() string {
	return e.msg
}
func (e LError) Type() LSymbol {
	return symError
}
func (e LError) String() string {
	return fmt.Sprintf("<Error: %s>", e.msg)
}
