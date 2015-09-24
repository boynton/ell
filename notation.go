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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func fileReadable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		if info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

// Input - an object to read from
type Input struct {
	file   *os.File
	reader *dataReader
	name   string
}

var typeInput = intern("<input>")
var typeOutput = intern("<input>")

func isInput(obj LAny) bool {
	_, ok := obj.(*Input)
	return ok
}

// Type returns the type of the object
func (*Input) Type() LAny {
	return typeInput
}

// Value returns the object itself for primitive types
func (in *Input) Value() LAny {
	return in
}

func (in *Input) String() string {
	if in.reader == nil {
		return fmt.Sprintf("<input: CLOSED %s>", in.name)
	}
	return fmt.Sprintf("<input: %s>", in.name)
}

func (in *Input) Copy() LAny {
	//not a copy!
	return in
}

// Equal returns true if the object is equal to the argument
func (in *Input) Equal(another LAny) bool {
	if a, ok := another.(*Input); ok {
		return in == a
	}
	return false
}

func readInput(input LAny) (LAny, error) {
	switch in := input.(type) {
	case *Input:
		return in.read()
	default:
		return nil, TypeError(intern("input"), input)
	}
}
func closeInput(input LAny) error {
	switch in := input.(type) {
	case *Input:
		return in.close()
	default:
		return TypeError(intern("input"), input)
	}
}
func (in *Input) read() (LAny, error) {
	if in.reader == nil {
		return nil, Error("Input is closed: ", in)
	}
	obj, err := in.reader.readData()
	if err != nil {
		if err == io.EOF {
			return EOF, nil
		}
		return nil, err
	}
	return obj, nil
}
func (in *Input) close() error {
	var err error
	if in.file != nil {
		err = in.file.Close()
		in.file = nil
	}
	in.reader = nil
	return err
}

func fileContents(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//todo: implement Output

func openInputFile(path string) (*Input, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(fi)
	dr := newDataReader(r)
	return &Input{fi, dr, path}, nil
}

func openInputString(input string) *Input {
	r := strings.NewReader(input)
	dr := newDataReader(r)
	return &Input{nil, dr, input}
}

func decode(in io.Reader) (LAny, error) {
	br := bufio.NewReader(in)
	dr := dataReader{br}
	return dr.readData()
}

type dataReader struct {
	in *bufio.Reader
}

func newDataReader(in io.Reader) *dataReader {
	br := bufio.NewReader(in)
	return &dataReader{br}
}

func (dr *dataReader) getChar() (byte, error) {
	return dr.in.ReadByte()
}

func (dr *dataReader) ungetChar() error {
	return dr.in.UnreadByte()
}

func (dr *dataReader) readData() (LAny, error) {
	//c, n, e := dr.in.ReadRune()
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			c, e = dr.getChar()
			continue
		}
		switch c {
		case ';':
			e = dr.decodeComment()
			if e != nil {
				break
			} else {
				c, e = dr.getChar()
			}
		case '\'':
			o, err := dr.readData()
			if err != nil {
				return nil, err
			}
			if o == EOF {
				return o, nil
			}
			return list(symQuote, o), nil
		case '`':
			o, err := dr.readData()
			if err != nil {
				return nil, err
			}
			return list(symQuasiquote, o), nil
		case '~':
			c, e := dr.getChar()
			if e != nil {
				return nil, e
			}
			sym := symUnquote
			if c != '@' {
				dr.ungetChar()
			} else {
				sym = symUnquoteSplicing
			}
			o, err := dr.readData()
			if err != nil {
				return nil, err
			}
			return list(sym, o), nil
		case '#':
			return dr.decodeReaderMacro()
		case '(':
			return dr.decodeList()
		case '[':
			return dr.decodeVector()
		case '{':
			return dr.decodeStruct()
		case '"':
			return dr.decodeString()
		case ')', ']', '}':
			return nil, Error("Unexpected '", string(c), "'")
		default:
			atom, err := dr.decodeAtom(c)
			return atom, err
		}
	}
	return EOF, e
}

func (dr *dataReader) decodeComment() error {
	c, e := dr.getChar()
	for e == nil {
		if c == '\n' {
			return nil
		}
		c, e = dr.getChar()
	}
	return e
}

func (dr *dataReader) decodeString() (LAny, error) {
	buf := []byte{}
	c, e := dr.getChar()
	escape := false
	for e == nil {
		if escape {
			escape = false
			switch c {
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'f':
				buf = append(buf, '\f')
			case 'b':
				buf = append(buf, '\b')
			case 'r':
				buf = append(buf, '\r')
			case 'u', 'U':
				c, e = dr.getChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
			}
		} else if c == '"' {
			break
		} else if c == '\\' {
			escape = true
		} else {
			escape = false
			buf = append(buf, c)
		}
		c, e = dr.getChar()
	}
	s := LString(string(buf))
	return s, e
}

func (dr *dataReader) decodeList() (LAny, error) {
	items, err := dr.decodeSequence(')')
	if err != nil {
		return nil, err
	}
	return toList(items), nil
}

func (dr *dataReader) decodeVector() (LAny, error) {
	items, err := dr.decodeSequence(']')
	if err != nil {
		return nil, err
	}
	return toVector(items, len(items)), nil
}

func (dr *dataReader) skipToData(skipColon bool) (byte, error) {
	c, err := dr.getChar()
	for err == nil {
		if isWhitespace(c) || (skipColon && c == ':') {
			c, err = dr.getChar()
			continue
		}
		if c == ';' {
			err = dr.decodeComment()
			if err == nil {
				c, err = dr.getChar()
			}
			continue
		}
		return c, nil
	}
	return 0, err
}

func (dr *dataReader) decodeStruct() (LAny, error) {
	items := []LAny{}
	var err error
	var c byte
	for err == nil {
		c, err = dr.skipToData(false)
		if err != nil {
			return nil, err
		}
		if c == ':' {
			return nil, Error("Bad syntax, unexpected ':' in struct")
		}
		if c == '}' {
			return newStruct(items)
		}
		dr.ungetChar()
		element, err := dr.readData()
		if err != nil {
			return nil, err
		}
		items = append(items, element)
		c, err = dr.skipToData(true)
		if err != nil {
			return nil, err
		}
		if c == '}' {
			return nil, Error("mismatched key/value in struct")
		}
		dr.ungetChar()
		element, err = dr.readData()
		if err != nil {
			return nil, err
		}
		items = append(items, element)
	}
	return nil, err
}

func (dr *dataReader) decodeSequence(endChar byte) ([]LAny, error) {
	c, err := dr.getChar()
	items := []LAny{}
	for err == nil {
		if isWhitespace(c) {
			c, err = dr.getChar()
			continue
		}
		if c == ';' {
			err = dr.decodeComment()
			if err == nil {
				c, err = dr.getChar()
			}
			continue
		}
		if c == endChar {
			return items, nil
		}
		dr.ungetChar()
		element, err := dr.readData()
		if err != nil {
			return nil, err
		}
		items = append(items, element)
		c, err = dr.getChar()
	}
	return nil, err
}

func (dr *dataReader) decodeAtom(firstChar byte) (LAny, error) {
	s, err := dr.decodeAtomString(firstChar)
	if err != nil {
		return nil, err
	}
	slen := len(s)
	if slen == 0 { //?
		panic("whuh?")
		return dr.readData()
	}
	keyword := false
	if s[slen-1] == ':' {
		keyword = true
		s = s[:slen-1]
	} else {
		//reserved words. We could do without this by using #n, #f, #t reader macros like scheme does
		//but EllDn is JSON compatible, so these symbols need to be reserved
		if s == "null" {
			return Null, nil
		} else if s == "true" {
			return True, nil
		} else if s == "false" {
			return False, nil
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		if keyword {
			return nil, Error("Keyword cannot have a name that looks like a number: ", s, ":")
		}
		return LNumber(f), nil
	}
	if keyword {
		s += ":"
	}
	sym := intern(s)
	return sym, nil
}

func (dr *dataReader) decodeAtomString(firstChar byte) (string, error) {
	buf := []byte{}
	if firstChar != 0 {
		if firstChar == ':' {
			return "", Error("Invalid keyword: colons go at the end of symbols, not at the beginning")
		}
		buf = append(buf, firstChar)
	}
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			break
		}
		if c == ':' {
			buf = append(buf, c)
			break
		}
		if isDelimiter(c) {
			dr.ungetChar()
			break
		}
		buf = append(buf, c)
		c, e = dr.getChar()
	}
	if e != nil && e != io.EOF {
		return "", e
	}
	s := string(buf)
	return s, nil
}

func namedChar(name string) (rune, error) {
	switch name {
	case "null":
		return 0, nil
	case "alarm":
		return 7, nil
	case "backspace":
		return 8, nil
	case "tab":
		return 9, nil
	case "newline":
		return 10, nil
	case "return":
		return 13, nil
	case "escape":
		return 27, nil
	case "space":
		return 32, nil
	case "delete":
		return 127, nil
	default:
		if strings.HasPrefix(name, "x") {
			hex := name[1:]
			i, err := strconv.ParseInt(hex, 16, 64)
			if err != nil {
				return 0, err
			}
			return rune(i), nil
		}
		return 0, Error("bad named character: #\\", name)
	}
}

func (dr *dataReader) decodeReaderMacro() (LAny, error) {
	c, e := dr.getChar()
	if e != nil {
		return nil, e
	}
	switch c {
	case '\\': // to handle character literals.
		c, e = dr.getChar()
		if e != nil {
			return nil, e
		}
		if isWhitespace(c) || isDelimiter(c) {
			return newCharacter(rune(c)), nil
		}
		c2, e := dr.getChar()
		if e != nil {
			if e != io.EOF {
				return nil, e
			}
			c2 = 32
		}
		if !isWhitespace(c2) && !isDelimiter(c2) {
			var name []byte
			name = append(name, c)
			name = append(name, c2)
			c, e = dr.getChar()
			for (e == nil || e != io.EOF) && !isWhitespace(c) && !isDelimiter(c) {
				name = append(name, c)
				c, e = dr.getChar()
			}
			if e != io.EOF && e != nil {
				return nil, e
			}
			dr.ungetChar()
			r, e := namedChar(string(name))
			if e != nil {
				return nil, e
			}
			return newCharacter(r), nil
		} else if e == nil {
			dr.ungetChar()
		}
		return newCharacter(rune(c)), nil
	case '!': //to handle shell scripts, handle #! as a comment
		err := dr.decodeComment()
		return Null, err
	default:
		atom, err := dr.decodeAtomString(c)
		if err != nil {
			return nil, Error("Bad reader macro: #", string([]byte{c}), " ...")
		}
		if isValidTypeName(atom) {
			val, err := dr.readData()
			if err != nil {
				return nil, Error("Bad reader macro: #", atom, " ...")
			}
			return instance(intern(atom), val)
		}
		return nil, Error("Bad reader macro: #", atom, " ...")
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ','
}

func isDelimiter(b byte) bool {
	return b == '(' || b == ')' || b == '[' || b == ']' || b == '{' || b == '}' || b == '"' || b == '\'' || b == ';' || b == ':'
}

func write(obj LAny) string {
	if obj == nil {
		panic("null pointer")
	}
	elldn, _ := writeData(obj, false)
	return elldn
}

func writeData(obj LAny, json bool) (string, error) {
	//an error is never returned for non-json
	switch o := obj.(type) {
	case LBoolean, LNull:
		return o.String(), nil
	case *LList:
		if json {
			return writeVector(listToVector(o), json)
		}
		return writeList(o), nil
	case *LSymbol:
		if json {
			return encodeString(unkeywordedString(o)), nil
		}
		return o.String(), nil
	case LString:
		s := encodeString(string(o))
		return s, nil
	case *LVector:
		return writeVector(o, json)
	case *LStruct:
		return writeStruct(o, json)
	case LNumber:
		return o.String(), nil
	case LChar:
		switch o {
		case 0:
			return "#\\null", nil
		case 7:
			return "#\\alarm", nil
		case 8:
			return "#\\backspace", nil
		case 9:
			return "#\\tab", nil
		case 10:
			return "#\\newline", nil
		case 13:
			return "#\\return", nil
		case 27:
			return "#\\escape", nil
		case 32:
			return "#\\space", nil
		case 127:
			return "#\\delete", nil
		default:
			if o < 127 {
				return "#\\" + string(rune(o)), nil
			}
			return fmt.Sprintf("#\\x%04X", int(o)), nil
		}
	default:
		if json {
			return "", Error("data cannot be described in JSON: ", obj)
		}
		if o == nil {
			panic("whoops: nil not allowed here!")
		}
		return o.String(), nil
	}
}

func writeList(lst *LList) string {
	if lst == EmptyList {
		return "()"
	}
	if lst.cdr != EmptyList {
		if lst.car == symQuote {
			return "'" + cadr(lst).String()
		} else if lst.car == symQuasiquote {
			return "`" + cadr(lst).String()
		} else if lst.car == symUnquote {
			return "~" + cadr(lst).String()
		} else if lst.car == symUnquoteSplicing {
			return "~@" + cadr(lst).String()
		}
	}
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(write(car(lst)))
	lst = lst.cdr
	for lst != EmptyList {
		buf.WriteString(" ")
		s, _ := writeData(lst.car, false)
		buf.WriteString(s)
		lst = lst.cdr
	}
	buf.WriteString(")")
	return buf.String()
}

func writeVector(ary *LVector, json bool) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("[")
	vlen := len(ary.elements)
	if vlen > 0 {
		s, err := writeData(ary.elements[0], json)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		delim := " "
		if json {
			delim = ", "
		}
		for i := 1; i < vlen; i++ {
			s, err := writeData(ary.elements[i], json)
			if err != nil {
				return "", err
			}
			buf.WriteString(delim)
			buf.WriteString(s)
		}
	}
	buf.WriteString("]")
	return buf.String(), nil
}

func writeStruct(m *LStruct, json bool) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	first := true
	delim := ", "
	sep := " "
	if json {
		delim = ", "
		sep = ": "
	}
	for k, v := range m.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(delim)
		}
		s, err := writeData(k, json)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		buf.WriteString(sep)
		s, err = writeData(v, json)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	buf.WriteString("}")
	return buf.String(), nil

}

//return a JSON string of the object, or an error if it cannot be expressed in JSON
//this is very close to EllDn, the standard output format. Exceptions:
//  1. the only symbols allowed are true, false, null
func toJSON(obj LAny) (string, error) {
	return writeData(obj, true)
}
