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

// generic file ops

func fileReadable(path string) bool {
	if info, err := os.Stat(path); err == nil {
		if info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

func slurpFile(path string) (*LOB, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return EmptyString, err
	}
	return newString(string(b)), nil
}

func spitFile(path string, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0644)
}

// --- reader

func keysOptionValue(options *LOB) (*LOB, error) {
	if options != nil {
		t, err := get(options, intern("keys:"))
		if err == nil && isType(t) {
			switch t {
			case SymbolType, KeywordType, StringType:
				return t, nil
			default:
				return nil, Error(ArgumentErrorKey, "Bad option value for keys: ", t)
			}
		}
	}
	return Null, nil
}

//only reads the first item in the input, along with how many characters it read
// for subsequence calls, you can slice the string to continue
func read(input *LOB) (*LOB, error) {
	if !isString(input) {
		return nil, Error(ArgumentErrorKey, "read invalid input: ", input)
	}
	r := strings.NewReader(input.text)
	reader := newDataReader(r)
	obj, err := reader.readData(AnyType)
	if err != nil {
		if err == io.EOF {
			return Null, nil
		}
		return nil, err
	}
	return obj, nil
}

func readAll(input *LOB, keys *LOB) (*LOB, error) {
	if !isString(input) {
		return nil, Error(ArgumentErrorKey, "read-all invalid input: ", input)
	}
	reader := newDataReader(strings.NewReader(input.text))
	lst := EmptyList
	tail := EmptyList
	val, err := reader.readData(keys)
	for err == nil {
		if lst == EmptyList {
			lst = list(val)
			tail = lst
		} else {
			tail.cdr = list(val)
			tail = tail.cdr
		}
		val, err = reader.readData(keys)
	}
	if err != io.EOF {
		return nil, err
	}
	return lst, nil
}

type dataReader struct {
	in  *bufio.Reader
	pos int
}

func newDataReader(in io.Reader) *dataReader {
	br := bufio.NewReader(in)
	return &dataReader{br, 0}
}

func (dr *dataReader) getChar() (byte, error) {
	b, e := dr.in.ReadByte()
	if e == nil {
		dr.pos++
	}
	return b, e
}

func (dr *dataReader) ungetChar() error {
	e := dr.in.UnreadByte()
	if e == nil {
		dr.pos--
	}
	return e
}

func (dr *dataReader) readData(keys *LOB) (*LOB, error) {
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
			o, err := dr.readData(keys)
			if err != nil {
				return nil, err
			}
			if o == nil {
				return o, nil
			}
			return list(symQuote, o), nil
		case '`':
			o, err := dr.readData(keys)
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
			o, err := dr.readData(keys)
			if err != nil {
				return nil, err
			}
			return list(sym, o), nil
		case '#':
			return dr.decodeReaderMacro(keys)
		case '(':
			return dr.decodeList(keys)
		case '[':
			return dr.decodeVector(keys)
		case '{':
			return dr.decodeStruct(keys)
		case '"':
			return dr.decodeString()
		case ')', ']', '}':
			return nil, Error(SyntaxErrorKey, "Unexpected '", string(c), "'")
		default:
			atom, err := dr.decodeAtom(c)
			return atom, err
		}
	}
	return nil, e
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

func (dr *dataReader) decodeString() (*LOB, error) {
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
	s := newString(string(buf))
	return s, e
}

func (dr *dataReader) decodeList(keys *LOB) (*LOB, error) {
	items, err := dr.decodeSequence(')', keys)
	if err != nil {
		return nil, err
	}
	return listFromValues(items), nil
}

func (dr *dataReader) decodeVector(keys *LOB) (*LOB, error) {
	items, err := dr.decodeSequence(']', keys)
	if err != nil {
		return nil, err
	}
	return vector(items...), nil
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

func (dr *dataReader) decodeStruct(keys *LOB) (*LOB, error) {
	items := []*LOB{}
	var err error
	var c byte
	for err == nil {
		c, err = dr.skipToData(false)
		if err != nil {
			return nil, err
		}
		if c == ':' {
			return nil, Error(SyntaxErrorKey, "Unexpected ':' in struct")
		}
		if c == '}' {
			return newStruct(items)
		}
		dr.ungetChar()
		element, err := dr.readData(nil)
		if err != nil {
			return nil, err
		}
		if keys != nil && keys != AnyType {
			switch keys {
			case KeywordType:
				element, err = toKeyword(element)
				if err != nil {
					return nil, err
				}
			case SymbolType:
				element, err = toSymbol(element)
				if err != nil {
					return nil, err
				}
			case StringType:
				element, err = toString(element)
				if err != nil {
					return nil, err
				}
			}
		}
		items = append(items, element)
		c, err = dr.skipToData(true)
		if err != nil {
			return nil, err
		}
		if c == '}' {
			return nil, Error(SyntaxErrorKey, "mismatched key/value in struct")
		}
		dr.ungetChar()
		element, err = dr.readData(keys)
		if err != nil {
			return nil, err
		}
		items = append(items, element)
	}
	return nil, err
}

func (dr *dataReader) decodeSequence(endChar byte, keys *LOB) ([]*LOB, error) {
	c, err := dr.getChar()
	items := []*LOB{}
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
		element, err := dr.readData(keys)
		if err != nil {
			return nil, err
		}
		items = append(items, element)
		c, err = dr.getChar()
	}
	return nil, err
}

func (dr *dataReader) decodeAtom(firstChar byte) (*LOB, error) {
	s, err := dr.decodeAtomString(firstChar)
	if err != nil {
		return nil, err
	}
	slen := len(s)
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
			return nil, Error(SyntaxErrorKey, "Keyword cannot have a name that looks like a number: ", s, ":")
		}
		return newFloat64(f), nil
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
			return "", Error(SyntaxErrorKey, "Invalid keyword: colons only valid at the end of symbols")
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

func (dr *dataReader) decodeType(firstChar byte) (string, error) {
	buf := []byte{}
	if firstChar != '<' {
		return "", Error(SyntaxErrorKey, "Invalid type name")
	}
	buf = append(buf, firstChar)
	c, e := dr.getChar()
	for e == nil {
		if isWhitespace(c) {
			break
		}
		if c == '>' {
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
		return 0, Error(SyntaxErrorKey, "Bad named character: #\\", name)
	}
}

func (dr *dataReader) decodeReaderMacro(keys *LOB) (*LOB, error) {
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
		atom, err := dr.decodeType(c)
		if err != nil {
			return nil, Error(SyntaxErrorKey, "Bad reader macro: #", string([]byte{c}), " ...")
		}
		if isValidTypeName(atom) {
			val, err := dr.readData(keys)
			if err != nil {
				return nil, Error(SyntaxErrorKey, "Bad reader macro: #", atom, " ...")
			}
			return instance(intern(atom), val)
		}
		return nil, Error(SyntaxErrorKey, "Bad reader macro: #", atom, " ...")
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ','
}

func isDelimiter(b byte) bool {
	return b == '(' || b == ')' || b == '[' || b == ']' || b == '{' || b == '}' || b == '"' || b == '\'' || b == ';' || b == ':'
}

//writer

const defaultIndentSize = "    "

func write(obj *LOB) string {
	return writeIndent(obj, "")
}

func pretty(obj *LOB) string {
	return writeIndent(obj, defaultIndentSize)
}

func writeIndent(obj *LOB, indentSize string) string {
	s, _ := writeToString(obj, false, indentSize)
	return s
}

func writeAll(obj *LOB) string {
	return writeAllIndent(obj, "")
}

func prettyAll(obj *LOB) string {
	return writeAllIndent(obj, "    ")
}

func writeAllIndent(obj *LOB, indent string) string {
	if isList(obj) {
		var buf bytes.Buffer
		for obj != EmptyList {
			o := car(obj)
			s, _ := writeToString(o, false, indent)
			buf.WriteString(s)
			buf.WriteString("\n")
			obj = cdr(obj)
		}
		return buf.String()
	}
	s, _ := writeToString(obj, false, indent)
	if indent == "" {
		return s + "\n"
	}
	return s
}

func writeToString(obj *LOB, json bool, indentSize string) (string, error) {
	elldn, err := writeData(obj, json, "", indentSize)
	if err != nil {
		return "", err
	}
	if indentSize != "" {
		return elldn + "\n", nil
	}
	return elldn, nil
}

func writeData(obj *LOB, json bool, indent string, indentSize string) (string, error) {
	//an error is never returned for non-json
	switch obj.variant {
	case BooleanType, NullType, NumberType:
		return obj.String(), nil
	case ListType:
		if json {
			return writeVector(listToVector(obj), json, indent, indentSize)
		}
		return writeList(obj, indent, indentSize), nil
	case KeywordType:
		if json {
			return encodeString(unkeywordedString(obj)), nil
		}
		return obj.String(), nil
	case SymbolType, TypeType:
		return obj.String(), nil
	case StringType:
		return encodeString(obj.text), nil
	case VectorType:
		return writeVector(obj, json, indent, indentSize)
	case StructType:
		return writeStruct(obj, json, indent, indentSize)
	case CharacterType:
		switch obj.ival {
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
			if obj.ival < 127 && obj.ival > 32 {
				return "#\\" + string(rune(obj.ival)), nil
			}
			return fmt.Sprintf("#\\x%04X", obj.ival), nil
		}
	default:
		if json {
			return "", Error(ArgumentErrorKey, "Data cannot be described in JSON: ", obj)
		}
		if obj == nil {
			panic("whoops: nil not allowed here!")
		}
		return obj.String(), nil
	}
}

func writeList(lst *LOB, indent string, indentSize string) string {
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
	delim := " "
	nextIndent := ""
	if indentSize != "" {
		nextIndent = indent + indentSize
		delim = "\n" + nextIndent
		buf.WriteString("\n" + nextIndent)
	}
	s, _ := writeData(lst.car, false, nextIndent, indentSize)
	buf.WriteString(s)
	lst = lst.cdr
	for lst != EmptyList {
		buf.WriteString(delim)
		s, _ := writeData(lst.car, false, nextIndent, indentSize)
		buf.WriteString(s)
		lst = lst.cdr
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString(")")
	return buf.String()
}

func writeVector(vec *LOB, json bool, indent string, indentSize string) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("[")
	vlen := len(vec.elements)
	if vlen > 0 {
		delim := ""
		if json {
			delim = ","
		}
		nextIndent := ""
		if indentSize != "" {
			nextIndent = indent + indentSize
			delim = delim + "\n" + nextIndent
			buf.WriteString("\n" + nextIndent)
		} else {
			delim = delim + " "
		}
		s, err := writeData(vec.elements[0], json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		for i := 1; i < vlen; i++ {
			s, err := writeData(vec.elements[i], json, nextIndent, indentSize)
			if err != nil {
				return "", err
			}
			buf.WriteString(delim)
			buf.WriteString(s)
		}
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString("]")
	return buf.String(), nil
}

func writeStruct(strct *LOB, json bool, indent string, indentSize string) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	size := len(strct.bindings)
	delim := ""
	sep := " "
	if json {
		delim = ","
		sep = ": "
	}
	nextIndent := ""
	if size > 0 {
		if indentSize != "" {
			nextIndent = indent + indentSize
			delim = delim + "\n" + nextIndent
			buf.WriteString("\n" + nextIndent)
		} else {
			delim = delim + " "
		}
	}
	first := true
	for k, v := range strct.bindings {
		if first {
			first = false
		} else {
			buf.WriteString(delim)
		}
		s, err := writeData(k.toLOB(), json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		buf.WriteString(sep)
		s, err = writeData(v, json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString("}")
	return buf.String(), nil
}
