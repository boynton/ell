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

type port interface {
	isBinary() bool
	isInput() bool
	isOutput() bool
	read() (lob, error)
	write(obj lob) error
	close() error
}

type linport struct {
	file   *os.File
	reader *dataReader
}

func (in linport) isBinary() bool {
	return false
}
func (in linport) isInput() bool {
	return true
}
func (in linport) isOutput() bool {
	return false
}
func (in linport) read() (lob, error) {
	obj, err := in.reader.readData()
	if err != nil {
		if err == io.EOF {
			return EOF, nil
		}
		return nil, err
	}
	return obj, nil
}
func (in linport) write(obj lob) error {
	return newError("Cannot write an input port")
}
func (in linport) close() error {
	if in.file != nil {
		return in.file.Close()
	}
	return nil
}

//todo: implement LOutputPort

func openInputFile(path string) (port, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(fi)
	dr := newDataReader(r)
	port := linport{fi, dr}
	return &port, nil
}

func openInputString(input string) port {
	r := strings.NewReader(input)
	dr := newDataReader(r)
	port := linport{nil, dr}
	return &port
}

func decode(in io.Reader) (lob, error) {
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

func (dr *dataReader) readData() (lob, error) {
	//c, n, e := dr.in.ReadRune()
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
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
			o, err := dr.readData()
			if err != nil {
				return nil, err
			}
			if o == EOF || o == Nil {
				return o, nil
			}
			return list(intern("quote"), o), nil
		case '#':
			o, e := dr.decodeReaderMacro()
			if e != nil || o != nil {
				return o, e
			}
		case '(':
			return dr.decodeList()
		case '[':
			return dr.decodeVector()
		case '{':
			return dr.decodeMap()
		case '"':
			return dr.decodeString()
		case ')', ']', '}':
			return nil, newError("Unexpected '", string(c), "'")
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

func (dr *dataReader) decodeString() (lob, error) {
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
					return Nil, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return Nil, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return Nil, e
				}
				buf = append(buf, c)
				c, e = dr.getChar()
				if e != nil {
					return Nil, e
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
	s := newString(string(buf))
	return s, e
}

func (dr *dataReader) decodeList() (lob, error) {
	items, err := dr.decodeSequence(')')
	if err != nil {
		return nil, err
	}
	return toList(items), nil
}

func (dr *dataReader) decodeVector() (lob, error) {
	items, err := dr.decodeSequence(']')
	if err != nil {
		return nil, err
	}
	return toVector(items, len(items)), nil
}

func (dr *dataReader) decodeMap() (lob, error) {
	items, err := dr.decodeSequence('}')
	if err != nil {
		return nil, err
	}
	return toMap(items, len(items))
}

func (dr *dataReader) decodeSequence(endChar byte) ([]lob, error) {
	c, err := dr.getChar()
	items := []lob{}
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

func (dr *dataReader) decodeAtom(firstChar byte) (lob, error) {
	buf := []byte{}
	if firstChar != 0 {
		buf = append(buf, firstChar)
	}
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
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
		return nil, e
	}
	s := string(buf)
	if len(s) == 0 {
		return dr.readData()
	}
	if s == ":" {
		return nil, newError("Bad token: :")
	}
	//reserved words. We could do without this by using #n, #f, #t reader macros like scheme does
	if s == "nil" || s == "null" { //EllDN is upwards compatible with JSON, so we need to handle null
		return Nil, nil
	} else if s == "true" {
		return True, nil
	} else if s == "false" {
		return False, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return linteger(i), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return lreal(f), nil
	}
	sym := intern(s)
	return sym, nil
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
		return 0, newError("bad named character: #\\", name)
	}
}

func (dr *dataReader) decodeReaderMacro() (lob, error) {
	c, e := dr.getChar()
	if e != nil {
		return nil, e
	}
	switch c {
	case '\\':
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
	case 'f':
		return False, nil
	case 't':
		return True, nil
	case 'n':
		return Nil, nil
	case '!': //to handle shell scripts, handle #! as a comment
		err := dr.decodeComment()
		return Nil, err
	default:
		return nil, newError("Bad reader macro: #", string([]byte{c}), " ...")
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ','
}

func isDelimiter(b byte) bool {
	return b == '(' || b == ')' || b == '[' || b == ']' || b == '{' || b == '}' || b == '"' || b == '\'' || b == ';'
}

func write(obj lob) string {
	if obj == nil {
		panic("null pointer")
	}
	elldn, _ := writeData(obj, false)
	return elldn
}

func writeData(obj lob, json bool) (string, error) {
	//an error is never returned for non-json
	if json {
		if obj == True {
			return "true", nil
		} else if obj == False {
			return "false", nil
		} else if obj == Nil {
			return "null", nil
		}
	}
	switch o := obj.(type) {
	case *llist:
		if json {
			return "", newError("list cannot be described in JSON: ", obj)
		}
		return writeList(o), nil
	case *lsymbol:
		if json {
			return "", newError("symbol cannot be described in JSON: ", obj)
		}
		return o.String(), nil
	case lstring:
		s := encodeString(string(o))
		return s, nil
	case *lcode:
		if json {
			return "", newError("code cannot be described in JSON: ", obj)
		}
		return o.String(), nil
	case *lvector:
		return writeVector(o, json)
	case *lmap:
		return writeMap(o, json)
	case linteger, lreal:
		return o.String(), nil
	case lchar:
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
			return "", newError("data cannot be described in JSON: ", obj)
		}
		if o == nil {
			panic("whoops: nil not allowed here!")
		}
		return o.String(), nil
	}
}

func writeList(lst *llist) string {
	if lst == EmptyList {
		return "()"
	}
	if lst.car == intern("quote") {
		if lst.cdr != EmptyList {
			return "'" + cadr(lst).String()
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
func toJSON(obj lob) (string, error) {
	return writeData(obj, true)
}
