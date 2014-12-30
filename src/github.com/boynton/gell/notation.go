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
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
)

func FileReadable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		if info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

type LPort interface {
	IsBinary() bool
	IsInput() bool
	IsOutput() bool
	Read() (LObject, error)
	Write(obj LObject) error
	Close() error
}

type LInputPort struct {
	file   *os.File
	reader *DataReader
}

func (in LInputPort) IsBinary() bool {
	return false
}
func (in LInputPort) IsInput() bool {
	return true
}
func (in LInputPort) IsOutput() bool {
	return false
}
func (in LInputPort) Read() (LObject, error) {
	obj, err := in.reader.ReadData()
	if err != nil {
		if err == io.EOF {
			return EOI, nil
		}
		return nil, err
	}
	return obj, nil
}
func (in LInputPort) Write(obj LObject) error {
	return Error("Cannot write an input port")
}
func (in LInputPort) Close() error {
	if in.file != nil {
		return in.file.Close()
	}
	return nil
}

//todo: implement LOutputPort

const (
	READ  = "read"
	WRITE = "write"
)

func OpenInputFile(path string) (LPort, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(fi)
	dr := NewDataReader(r)
	port := LInputPort{fi, dr}
	return &port, nil
}

func OpenInputString(input string) LPort {
	r := strings.NewReader(input)
	dr := NewDataReader(r)
	port := LInputPort{nil, dr}
	return &port
}

func Decode(in io.Reader) (LObject, error) {
	br := bufio.NewReader(in)
	dr := DataReader{br}
	return dr.ReadData()
}

type DataReader struct {
	in *bufio.Reader
}

func NewDataReader(in io.Reader) *DataReader {
	br := bufio.NewReader(in)
	return &DataReader{br}
}

func (dr *DataReader) getChar() (byte, error) {
	return dr.in.ReadByte()
}

func (dr *DataReader) ungetChar() error {
	return dr.in.UnreadByte()
}

func (dr *DataReader) ReadData() (LObject, error) {
	//c, n, e := dr.in.ReadRune()
	c, e := dr.getChar()
	for e == nil {
		if IsWhitespace(c) {
			c, e = dr.in.ReadByte()
			continue
		}
		switch c {
		case ';':
			if dr.decodeComment() != nil {
				break
			} else {
				c, e = dr.getChar()
			}
		case '\'':
			o, err := dr.ReadData()
			if err != nil {
				return nil, err
			}
			if o == EOI || o == NIL {
				return o, nil
			}
			return List(Intern("quote"), o), nil
		case '(':
			return dr.decodeList()
		case '[':
			return dr.decodeVector()
		case '{':
			return dr.decodeMap()
		case '"':
			return dr.decodeString()
		case ')', ']', '}':
			return nil, Error("Unexpected '", string(c), "'")
		default:
			atom, err := dr.decodeAtom(c)
			return atom, err
		}
	}
	return EOI, e
}

func (dr *DataReader) decodeComment() error {
	c, e := dr.getChar()
	for e == nil {
		if c == '\n' {
			return nil
		} else {
			c, e = dr.getChar()
		}
	}
	return e
}

func (dr *DataReader) decodeString() (LObject, error) {
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
					return NIL, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return NIL, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return NIL, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return NIL, e
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
	s := NewString(string(buf))
	return s, e
}

func (dr *DataReader) decodeList() (LObject, error) {
	items, tail, err := dr.decodeSequence(')', '.')
	if err != nil {
		return nil, err
	}
	if tail != nil {
		return ToImproperList(items, tail), nil
	} else {
		return ToList(items), nil
	}
}

func (dr *DataReader) decodeVector() (LObject, error) {
	items, _, err := dr.decodeSequence(']', 0)
	if err != nil {
		return nil, err
	}
	return ToVector(items, len(items)), nil
}

func (dr *DataReader) decodeMap() (LObject, error) {
	items, _, err := dr.decodeSequence('}', 0)
	if err != nil {
		return nil, err
	}
	return ToMap(items, len(items))
}

func (dr *DataReader) decodeSequence(endChar byte, tailTag byte) ([]LObject, LObject, error) {
	c, err := dr.getChar()
	items := []LObject{}
	var tail LObject
	for err == nil {
		if IsWhitespace(c) {
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
			return items, tail, nil
		}
		if tail != nil {
			return nil, nil, Error("Syntax error: object beyond tail of dotted pair")
		}
		if c == tailTag {
			tail, err = dr.ReadData()
			if err != nil {
				return nil, nil, err
			}
		} else {
			dr.ungetChar()
			element, err := dr.ReadData()
			if err != nil {
				return nil, nil, err
			} else {
				items = append(items, element)
			}
		}
		c, err = dr.getChar()
	}
	return nil, nil, err
}

func (dr *DataReader) decodeAtom(firstChar byte) (LObject, error) {
	buf := []byte{}
	if firstChar != 0 {
		if firstChar == ':' {
			//leading colon is treated as a delimiter, letting us read JSON/EllDn directly
			return dr.ReadData()
		} else {
			buf = append(buf, firstChar)
		}
	}
	c, e := dr.getChar()
	for e == nil {
		if IsWhitespace(c) {
			break
		}
		if IsDelimiter(c) || c == ';' {
			dr.ungetChar()
			break
		}
		buf = append(buf, c)
		c, e = dr.getChar()
	}
	s := string(buf)
	if s[0] == '#' {
		//todo: generic reader macro dispatch to make this extensible
		if s == "#t" {
			return TRUE, nil
		} else if s == "#f" {
			return FALSE, nil
		} else {
			return nil, Error("Bad reader macro: ", s)
		}
	}
	if strings.HasSuffix(s, ":") {
		//macro for quoted symbol (rather than introduce keywords as types)
		s := s[:len(s)-1]
		if s == "" {
			return dr.ReadData()
		}
		return List(Intern("quote"), Intern(s)), nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return linteger(i), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return lreal(f), nil
	}
	sym := Intern(s)
	return sym, nil
}

func IsWhitespace(b byte) bool {
	if b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ',' {
		return true
	} else {
		return false
	}
}

func IsDelimiter(b byte) bool {
	if b == '(' || b == ')' || b == '[' || b == ']' || b == '{' || b == '}' || b == '"' || b == '\'' {
		return true
	} else {
		return false
	}
}

func Write(obj LObject) string {
	elldn, _ := writeData(obj, false)
	return elldn
}

func writeData(obj LObject, json bool) (string, error) {
	//an error is never returned for non-json
	if json {
		if obj == TRUE {
			return "true", nil
		} else if obj == FALSE {
			return "false", nil
		} else if obj == NIL {
			return "null", nil
		}
	}
	switch o := obj.(type) {
	case *lpair:
		if json {
			return "", Error("pair cannot be described in JSON: ", obj)
		}
		return writeList(o), nil
	case *lsymbol:
		if json {
			return "", Error("symbol cannot be described in JSON: ", obj)
		}
		return o.String(), nil
	case lstring:
		s := encodeString(string(o))
		return s, nil
	case *lcode:
		if json {
			return "", Error("code cannot be described in JSON: ", obj)
		}
		return o.String(), nil
	case *lvector:
		return writeVector(o, json)
	case *lmap:
		return writeMap(o, json)
	case linteger, lreal:
		return o.String(), nil
	default:
		if json {
			return "", Error("data cannot be described in JSON: ", obj)
		}
		return o.String(), nil
	}
}

func writeList(lst *lpair) string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(Write(lst.car))
	var tail LObject = lst.cdr
	b := true
	for b {
		if tail == NIL {
			b = false
		} else {
			p, isPair := tail.(*lpair)
			if isPair {
				tail = p.cdr
				buf.WriteString(" ")
				s, _ := writeData(p.car, false)
				buf.WriteString(s)
			} else {
				buf.WriteString(" . ")
				s, _ := writeData(tail, false)
				buf.WriteString(s)
				b = false
			}
		}
	}
	buf.WriteString(")")
	return buf.String()
}

func writeVector(vec *lvector, json bool) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("[")
	vlen := len(vec.elements)
	if vlen > 0 {
		s, err := writeData(vec.elements[0], json)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		delim := " "
		if json {
			delim = ", "
		}
		for i := 1; i < vlen; i++ {
			s, err := writeData(vec.elements[i], json)
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

func writeMap(m *lmap, json bool) (string, error) {
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
func JSON(obj LObject) (string, error) {
	return writeData(obj, true)
}
