/*
Copyright 2015 Lee Boynton

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

package ell

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	
	. "github.com/boynton/ell/data"
)

// IsDirectoryReadable - return true of the directory is readable
func IsDirectoryReadable(path string) bool {
	if strings.HasPrefix(path, "@/") {
		p := "lib" + path[1:]
		if info, err := fs.Stat(sysFS, p); err == nil {
			if info.Mode().IsDir() {
				return true
			}
		}
	}
	if info, err := os.Stat(path); err == nil {
		if info.Mode().IsDir() {
			return true
		}
	}
	return false
}

// IsFileReadable - return true of the file is readable
func IsFileReadable(path string) bool {
	if strings.HasPrefix(path, "@/") {
		p := "lib" + path[1:]
		if info, err := fs.Stat(sysFS, p); err == nil {
			if info.Mode().IsRegular() {
				return true
			}
		}
	}
	if info, err := os.Stat(path); err == nil {
		if info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

func ExpandFilePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		if home != "" {
			return home + path[1:]
		}
	}
	return path
}

//go:embed lib
var sysFS embed.FS

// SlurpFile - return the file contents as a string
func SlurpFile(path string) (string, error) {
	path = ExpandFilePath(path)
	var b []byte
	var err error
	if strings.HasPrefix(path, "@/") {
		b, err = fs.ReadFile(sysFS, "lib"+path[1:])
	} else {
		b, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SpitFile - write the string to the file.
func SpitFile(path string, data string) error {
	path = ExpandFilePath(path)
	return ioutil.WriteFile(path, []byte(data), 0644)
}

func ReadFromString(s string) (Value, error) {
	reader := &Reader{
		Input: bufio.NewReader(strings.NewReader(s)),
		Position: 0,
	}
	reader.Extension = &EllReaderExtension{r: reader}
	return reader.Read()
}

func ReadAllFromString(s string) (*List, error) {
	reader := &Reader{
		Input: bufio.NewReader(strings.NewReader(s)),
		Position: 0,
	}
	reader.Extension = &EllReaderExtension{r: reader}
	return reader.ReadAll()
	//	return ReadAll(strings.NewReader(s))
}

type EllReaderExtension struct {
	r *Reader
}

var QuoteSymbol = Intern("quote")
var QuasiquoteSymbol = Intern("quasiquote")
var UnquoteSymbol = Intern("unquote")
var UnquoteSymbolSplicing = Intern("unquote-splicing")

func (ext *EllReaderExtension) HandleReaderMacro(c byte) (Value, error, bool) {
	dr := ext.r
	var e error
	switch c {
	case '\\': // character literals.
		c, e = dr.GetChar()
		if e != nil {
			return nil, e, true
		}
		if IsWhitespace(c) || IsDelimiter(c) {
			return NewCharacter(rune(c)), nil, true
		}
		c2, e := dr.GetChar()
		if e != nil {
			if e != io.EOF {
				return nil, e, true
			}
			c2 = 32
		}
		if !IsWhitespace(c2) && !IsDelimiter(c2) {
			var name []byte
			name = append(name, c)
			name = append(name, c2)
			c, e = dr.GetChar()
			for (e == nil || e != io.EOF) && !IsWhitespace(c) && !IsDelimiter(c) {
				name = append(name, c)
				c, e = dr.GetChar()
			}
			if e != io.EOF && e != nil {
				return nil, e, true
			}
			dr.UngetChar()
			r, e := NamedChar(string(name))
			if e != nil {
				return nil, e, true
			}
			return NewCharacter(r), nil, true
		} else if e == nil {
			dr.UngetChar()
		}
		return NewCharacter(rune(c)), nil, true
	case '!': //to handle shell scripts, handle #! as a comment
		err := dr.DecodeComment()
		return Null, err, true
	}
	return Null, nil, false
}

func NamedChar(name string) (rune, error) {
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
		return 0, NewError(SyntaxErrorKey, "Bad named character: #\\", name)
	}
}

func (ext *EllReaderExtension) HandleChar(c byte) (Value, error, bool) {
	switch c {
	case '\'':
		o, err := ext.r.ReadValue()
		if err != nil {
			return nil, err, true
		}
		if o == nil {
			return o, nil, true
		}
		return NewList(QuoteSymbol, o), nil, true
	case '`':
		o, err := ext.r.ReadValue()
		if err != nil {
			return nil, err, true
		}
		return NewList(QuasiquoteSymbol, o), nil, true
	case '~':
		c, e := ext.r.GetChar()
		if e != nil {
			return nil, e, true
		}
		sym := UnquoteSymbol
		if c != '@' {
			ext.r.UngetChar()
		} else {
			sym = UnquoteSymbolSplicing
		}
		o, err := ext.r.ReadValue()
		if err != nil {
			return nil, err, true
		}
		return NewList(sym, o), nil, true
	}
	return Null, nil, false
}

type EllWriterExtension struct {
	writer *Writer
}

func newWriter(indent string, json bool) *EllWriterExtension {
	writer := &Writer{Indent: indent, Json: json}
	ext := &EllWriterExtension{writer: writer}
	writer.Extension = ext
	return ext
}

func (ext *EllWriterExtension) write(val Value) string {
	s, err := ext.writer.Write(val)
	if err != nil {
		panic("unhandled object in Write: " + val.String())
	}
	return s
}

func (ext *EllWriterExtension) writeAll(lst *List) string {
	s, err := ext.writer.WriteAll(lst)
	if err != nil {
		panic("unhandled object in Write: " + lst.String())
	}
	return s
}

func (ext *EllWriterExtension) HandleValue(val Value) (string, error, bool) {
	switch p := val.(type) {
	case *List:
		if p.Cdr != EmptyList {
			if p.Car == QuoteSymbol {
				return "'" + Cadr(val).String(), nil, true
			} else if p.Car == QuasiquoteSymbol {
				return "`" + Cadr(val).String(), nil, true
			} else if p.Car == UnquoteSymbol {
				return "~" + Cadr(val).String(), nil, true
			} else if p.Car == UnquoteSymbolSplicing {
				return "~@" + Cadr(val).String(), nil, true
			}
		}
		return "", nil, false
	case *Character: //move this out of here
		c := p.Value
		switch c {
		case 0:
			return "#\\null", nil, true
		case 7:
			return "#\\alarm", nil, true
		case 8:
			return "#\\backspace", nil, true
		case 9:
			return "#\\tab", nil, true
		case 10:
			return "#\\newline", nil, true
		case 13:
			return "#\\return", nil, true
		case 27:
			return "#\\escape", nil, true
		case 32:
			return "#\\space", nil, true
		case 127:
			return "#\\delete", nil, true
		default:
			if c < 127 && c > 32 {
				return "#\\" + string(c), nil, true
			}
			return fmt.Sprintf("#\\x%04X", c), nil, true
		}
	}
	return "", nil, false
}

const defaultIndentSize = "    "

func Write(val Value) string {
	return newWriter("", false).write(val)
}

func Pretty(val Value) string {
	return newWriter(defaultIndentSize, false).write(val)
}

func WriteIndent(val Value, indent string) string {
	return newWriter(indent, false).write(val)
}

func WriteAll(lst *List) string {
	return newWriter("", false).writeAll(lst)
}

func WriteAllIndent(lst *List, indent string) string {
	return newWriter(indent, false).writeAll(lst)
}

func Json(val Value, indent string) (string, error) {
	ext := newWriter(indent, true)
	return ext.writer.Write(val)
}
