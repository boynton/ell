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
)

func FileReadable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		if info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

type Port interface {
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

func OpenInputFile(path string) (Port, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(fi)
	dr := NewDataReader(r)
	port := LInputPort{fi, dr}
	return &port, nil
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
		if isWhitespace(c) {
			c, e = dr.in.ReadByte()
			continue
		} else if c == ';' {
			if dr.decodeComment() != nil {
				break
			} else {
				c, e = dr.getChar()
				continue
			}
		} else if c == '\'' {
			o, err := dr.ReadData()
			if err != nil {
				return nil, err
			}
			if o == EOI || o == NIL {
				return o, nil
			}
			return List(Intern("quote"), o), nil
		} else if c == '(' {
			return dr.decodeList()
		} else if c == '"' {
			return dr.decodeString()
		} else {
			atom, err := dr.decodeAtom(c)
			return atom, err
		}
	}
	//fixme: discern between EOF and other errors
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
	c, e := dr.getChar()
	items := []LObject{}
	for e == nil {
		if isWhitespace(c) {
			c, e = dr.getChar()
			continue
		}
		if c == ';' {
			e = dr.decodeComment()
			if e == nil {
				c, e = dr.getChar()
			}
			continue
		}
		if c == ')' {
			break
		}
		dr.ungetChar()
		element, err := dr.ReadData()
		if err != nil {
			break
		} else {
			items = append(items, element)
		}
		c, e = dr.getChar()
	}
	if e != nil {
		return NIL, e
	}
	return ToList(items), nil
}

func (dr *DataReader) decodeAtom(firstChar byte) (LObject, error) {
	buf := []byte{}
	if firstChar != 0 {
		buf = append(buf, firstChar)
	}
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			break
		}
		if isDelimiter(c) || c == ';' {
			dr.ungetChar()
			break
		}
		buf = append(buf, c)
		if c == ':' {
			break
		}
		c, e = dr.getChar()
	}
	s := string(buf)
	i, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return linteger(i), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return lreal(f), nil
	}
	if s == "true" || s == "#t" {
		return TRUE, nil
	} else if s == "false" || s == "#f" {
		return FALSE, nil
	}
	sym := Intern(s)
	return sym, e
}

func isWhitespace(b byte) bool {
	if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
		return true
	} else {
		return false
	}
}

func isDelimiter(b byte) bool {
	if b == ')' || b == ')' || b == '[' || b == ']' || b == '{' || b == '}' || b == '"' || b == '\'' {
		return true
	} else {
		return false
	}
}

func Write(obj LObject) string {
	//finish this
	switch o := obj.(type) {
	case *llist:
		return writeList(o)
	case *lsymbol:
		return o.String()
	case lstring:
		s := encodeString(string(o))
		return s
	case *lcode:
		return o.String()
	//map?
	//vector?
	default:
		return o.String()
	}
}

func writeList(lst *llist) string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(lst.car.String())
	var tail LObject = lst.cdr
	b := true
	for b {
		if tail == NIL {
			b = false
		} else if IsList(tail) {
			lst = tail.(*llist)
			tail = lst.cdr
			buf.WriteString(" ")
			buf.WriteString(Write(lst.car))
		} else {
			buf.WriteString(" . ")
			buf.WriteString(Write(tail))
			b = false
		}
	}
	buf.WriteString(")")
	return buf.String()
}
