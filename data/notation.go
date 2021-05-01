package data

import(
	"bufio"
	"bytes"
	"io"
	"strconv"
)

type ReaderExtension interface {
	HandleChar(c byte) (Value, error, bool)
	HandleReaderMacro(c byte) (Value, error, bool)
}

type Reader struct {
	Input  *bufio.Reader
	Position int
	Extension ReaderExtension
}

func (reader *Reader) Read() (Value, error) {
	obj, err := reader.ReadValue()
	if err != nil {
		if err == io.EOF {
			return Null, nil
		}
		return nil, err
	}
	return obj, nil
}

// ReadAll - read all items in the input, returning a list of them.
func (reader *Reader) ReadAll() (*List, error) {
	lst := EmptyList
	tail := EmptyList
	val, err := reader.ReadValue()
	for err == nil {
		if lst == EmptyList {
			lst = NewList(val)
			tail = lst
		} else {
			tail.Cdr = NewList(val)
			tail = tail.Cdr
		}
		val, err = reader.ReadValue()
	}
	if err != io.EOF {
		return nil, err
	}
	return lst, nil
}


func IsWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == ','
}

func IsDelimiter(b byte) bool {
	return b == '(' || b == ')' || b == '[' || b == ']' || b == '{' || b == '}' || b == '"' || b == '\'' || b == ';' || b == ':'
}

func (dr *Reader) GetChar() (byte, error) {
	b, e := dr.Input.ReadByte()
	if e == nil {
		dr.Position++
	}
	return b, e
}

func (dr *Reader) UngetChar() error {
	e := dr.Input.UnreadByte()
	if e == nil {
		dr.Position--
	}
	return e
}

func (dr *Reader) ReadValue() (Value, error) {
	c, e := dr.GetChar()
	for e == nil {
		if IsWhitespace(c) {
			c, e = dr.GetChar()
			continue
		}
		switch c {
		case ';':
			e = dr.DecodeComment()
			if e != nil {
				break
			} else {
				c, e = dr.GetChar()
			}
		case '#':
			return dr.DecodeReaderMacro()
		case '(':
			return dr.DecodeList()
		case '[':
			return dr.DecodeVector()
		case '{':
			return dr.DecodeStruct()
		case '"':
			return dr.DecodeString()
		case ')', ']', '}':
			return nil, NewError(SyntaxErrorKey, "Unexpected '", string(c), "'")
		default:
			if dr.Extension != nil {
				o, err, done := dr.Extension.HandleChar(c)
				if done || err != nil {
					return o, err
				}
			}
			atom, err := dr.DecodeAtom(c)
			return atom, err
		}
	}
	return Null, e
}

func (dr *Reader) DecodeComment() error {
	c, e := dr.GetChar()
	for e == nil {
		if c == '\n' {
			return nil
		}
		c, e = dr.GetChar()
	}
	return e
}

func (dr *Reader) DecodeString() (Value, error) {
	var buf []byte
	c, e := dr.GetChar()
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
				c, e = dr.GetChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
				c, e = dr.GetChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
				c, e = dr.GetChar()
				if e != nil {
					return nil, e
				}
				buf = append(buf, c)
				c, e = dr.GetChar()
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
		c, e = dr.GetChar()
	}
	return NewString(string(buf)), e
}

func (dr *Reader) DecodeList() (Value, error) {
	items, err := dr.DecodeSequence(')')
	if err != nil {
		return nil, err
	}
	return ListFromValues(items), nil
}

func (dr *Reader) DecodeVector() (Value, error) {
	items, err := dr.DecodeSequence(']')
	if err != nil {
		return nil, err
	}
	return NewVector(items...), nil
}


func (dr *Reader) SkipToData(skipColon bool) (byte, error) {
	c, err := dr.GetChar()
	for err == nil {
		if IsWhitespace(c) || (skipColon && c == ':') {
			c, err = dr.GetChar()
			continue
		}
		if c == ';' {
			err = dr.DecodeComment()
			if err == nil {
				c, err = dr.GetChar()
			}
			continue
		}
		return c, nil
	}
	return 0, err
}

func (dr *Reader) DecodeStruct() (Value, error) {
	var items []Value
	var err error
	var c byte
	for err == nil {
		c, err = dr.SkipToData(false)
		if err != nil {
			return nil, err
		}
		if c == ':' {
			return nil, NewError(SyntaxErrorKey, "Unexpected ':' in struct")
		}
		if c == '}' {
			return MakeStruct(items)
		}
		dr.UngetChar()
		element, err := dr.ReadValue()
		if err != nil {
			return nil, err
		}
		items = append(items, element)
		c, err = dr.SkipToData(true)
		if err != nil {
			return nil, err
		}
		if c == '}' {
			return nil, NewError(SyntaxErrorKey, "mismatched key/value in struct")
		}
		dr.UngetChar()
		element, err = dr.ReadValue()
		if err != nil {
			return nil, err
		}
		items = append(items, element)
	}
	return nil, err
}

func (dr *Reader) DecodeSequence(endChar byte) ([]Value, error) {
	c, err := dr.GetChar()
	var items []Value
	for err == nil {
		if IsWhitespace(c) {
			c, err = dr.GetChar()
			continue
		}
		if c == ';' {
			err = dr.DecodeComment()
			if err == nil {
				c, err = dr.GetChar()
			}
			continue
		}
		if c == endChar {
			return items, nil
		}
		dr.UngetChar()
		element, err := dr.ReadValue()
		if err != nil {
			return nil, err
		}
		items = append(items, element)
		c, err = dr.GetChar()
	}
	return nil, err
}

func (dr *Reader) DecodeAtom(firstChar byte) (Value, error) {
	s, err := dr.DecodeAtomString(firstChar)
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
			return nil, NewError(SyntaxErrorKey, "Keyword cannot have a name that looks like a number: ", s, ":")
		}
		return Float(f), nil
	}
	if keyword {
		s += ":"
	}
	sym := Intern(s)
	return sym, nil
}

func (dr *Reader) DecodeAtomString(firstChar byte) (string, error) {
	var buf []byte
	if firstChar != 0 {
		if firstChar == ':' {
			return "", NewError(SyntaxErrorKey, "Invalid keyword: colons only valid at the end of symbols")
		}
		buf = append(buf, firstChar)
	}
	c, e := dr.GetChar()
	for e == nil {
		if IsWhitespace(c) {
			break
		}
		if c == ':' {
			buf = append(buf, c)
			break
		}
		if IsDelimiter(c) {
			dr.UngetChar()
			break
		}
		buf = append(buf, c)
		c, e = dr.GetChar()
	}
	if e != nil && e != io.EOF {
		return "", e
	}
	s := string(buf)
	return s, nil
}

func (dr *Reader) DecodeType(firstChar byte) (string, error) {
	var buf []byte
	if firstChar != '<' {
		panic("here!")
		return "", NewError(SyntaxErrorKey, "Invalid type name")
	}
	buf = append(buf, firstChar)
	c, e := dr.GetChar()
	for e == nil {
		if IsWhitespace(c) {
			break
		}
		if c == '>' {
			buf = append(buf, c)
			break
		}
		if IsDelimiter(c) {
			dr.UngetChar()
			break
		}
		buf = append(buf, c)
		c, e = dr.GetChar()
	}
	if e != nil && e != io.EOF {
		return "", e
	}
	s := string(buf)
	return s, nil
}

func (dr *Reader) DecodeReaderMacro() (Value, error) {
	c, e := dr.GetChar()
	if e != nil {
		return nil, e
	}
	switch c {
	case '[': //all non-printable objects are displayed like #[foo ... ]
		s, err := dr.DecodeAtomString(0)
		if err != nil {
			return nil, err
		}
		return nil, NewError(SyntaxErrorKey, "Unreadable object: #[", s, "]")
	default:
		if dr.Extension != nil {
			o, err, done := dr.Extension.HandleReaderMacro(c)
			if done || err != nil {
				return o, err
			}
		}
		atom, err := dr.DecodeType(c)
		if err != nil {
			return nil, err
		}
		if IsValidTypeName(atom) {
			val, err := dr.ReadValue()
			if err != nil {
				return nil, NewError(SyntaxErrorKey, "Bad reader macro: #", atom, " ...")
			}
			return NewInstance(Intern(atom), val)
		}
		return nil, NewError(SyntaxErrorKey, "Bad reader macro: #", atom, " ...")
	}
}

type WriterExtension interface {
	HandleValue(v Value) (string, error, bool)
}

type Writer struct {
	Json bool
	Indent string
	Extension WriterExtension
}

func (writer *Writer) Write(val Value) (string, error) {
	return writer.writeToString(val)
}

func (writer *Writer) WriteAll(lst *List) (string, error) {
	var buf bytes.Buffer
	for lst != EmptyList {
		o := lst.Car
		s, err := writer.writeToString(o)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		buf.WriteString("\n")
		lst = lst.Cdr
	}
	return buf.String(), nil
}

/*
func (writer *Writer) WriteAllIndent(lst *List, indent string) string {
	var buf bytes.Buffer
	for lst != EmptyList {
		o := lst.Car
		s, _ := WriteToString(o, false, indent)
		buf.WriteString(s)
		buf.WriteString("\n")
		lst = lst.Cdr
	}
	return buf.String()
}

func Pretty(obj Value) string {
	return WriteIndent(obj, defaultIndentSize)
}

func Json(obj Value, indent string) (string, error) {
	return WriteToString(obj, true, indent)
}

func WriteIndent(obj Value, indentSize string) string {
	s, _ := WriteToString(obj, false, indentSize)
	return s
}

func WriteToString(obj Value, json bool, indentSize string) (string, error) {
	elldn, err := writeData(obj, json, "", indentSize)
	if err != nil {
		return "", err
	}
	if indentSize != "" {
		return elldn + "\n", nil
	}
	return elldn, nil
}

*/

func (writer *Writer) writeToString(obj Value) (string, error) {
	elldn, err := writer.WriteData(obj, writer.Json, "", writer.Indent)
	if err != nil {
		return "", err
	}
	if writer.Indent != "" {
		return elldn + "\n", nil
	}
	return elldn, nil
}


func (writer *Writer) WriteData(o Value, json bool, indent string, indentSize string) (string, error) {
	//an error is never returned for non-json
	if writer.Extension != nil {
		s, err, done := writer.Extension.HandleValue(o)
		if done || err != nil {
			return s, err
		}
	}
	if o == Null {
		return "null", nil
	}
	switch p := o.(type) {
	case *Boolean:
		return p.String(), nil
	case *Number:
		return p.String(), nil
	case *List:
		if json {
			return writer.WriteVector(ListToVector(p), json, indent, indentSize)
		}
		return writer.WriteList(p, indent, indentSize), nil
	case *Keyword:
		if json {
			return EncodeString(p.Name()), nil
		}
		return p.String(), nil
	case *Symbol:
		if json {
			return EncodeString(p.Name()), nil
		}
		return o.String(), nil
	case *Type:
		if json {
			return EncodeString(p.Name()), nil
		}
		return o.String(), nil
	case *String:
		return EncodeString(p.Value), nil
	case *Vector:
		return writer.WriteVector(p, json, indent, indentSize)
	case *Struct:
		return writer.WriteStruct(p, json, indent, indentSize)
	case *Instance:
		if json {
			return p.Value.String(), nil
		}
		return o.String(), nil
	default:
		if json {
			return "", NewError(ArgumentErrorKey, "Data cannot be described in JSON: ", o)
		}
		return o.String(), nil
	}
}

func (writer *Writer) WriteVector(vec *Vector, json bool, indent string, indentSize string) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("[")
	vlen := len(vec.Elements)
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
		s, err := writer.WriteData(vec.Elements[0], json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		for i := 1; i < vlen; i++ {
			s, err := writer.WriteData(vec.Elements[i], json, nextIndent, indentSize)
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

func (writer *Writer) WriteStruct(strct *Struct, json bool, indent string, indentSize string) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("{")
	size := len(strct.Bindings)
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
	for k, v := range strct.Bindings {
		if first {
			first = false
		} else {
			buf.WriteString(delim)
		}
		s, err := writer.WriteData(k.ToValue(), json, nextIndent, indentSize)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
		buf.WriteString(sep)
		s, err = writer.WriteData(v, json, nextIndent, indentSize)
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

func (writer *Writer) WriteList(lst *List, indent string, indentSize string) string {
	if lst == EmptyList {
		return "()"
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
	s, _ := writer.WriteData(lst.Car, false, nextIndent, indentSize)
	buf.WriteString(s)
	lst = lst.Cdr
	for lst != EmptyList {
		buf.WriteString(delim)
		s, _ := writer.WriteData(lst.Car, false, nextIndent, indentSize)
		buf.WriteString(s)
		lst = lst.Cdr
	}
	if indentSize != "" {
		buf.WriteString("\n" + indent)
	}
	buf.WriteString(")")
	return buf.String()
}

// EncodeString - return the encoded form of a string value
func EncodeString(s string) string {
	var buf []rune
	buf = append(buf, '"')
	for _, c := range s {
		switch c {
		case '"':
			buf = append(buf, '\\')
			buf = append(buf, '"')
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
			buf = append(buf, c)
		}
	}
	buf = append(buf, '"')
	return string(buf)
}
