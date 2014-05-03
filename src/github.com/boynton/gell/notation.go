package gell

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

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

const (
	WHITESPACE_TAG = iota
	QUOTE_TAG
	ATOM_TAG
	STRING_TAG
	BEGIN_LIST_TAG
	END_LIST_TAG
	BEGIN_VECTOR_TAG
	END_VECTOR_TAG
	BEGIN_MAP_TAG
	END_MAP_TAG
)

type DataReader struct {
	in        io.Reader
	scanner   *bufio.Scanner
	ns        Namespace
	lastToken []byte
}

func MakeDataReader(in io.Reader, ns Namespace) DataReader {
	dr := DataReader{in, bufio.NewScanner(in), ns, nil}
	tokenizer := func(data []byte, eofp bool) (adv int, token []byte, err error) {
		switch data[0] {
		case '(':
			adv, token, err = 1, []byte{BEGIN_LIST_TAG}, nil
		case ')':
			adv, token, err = 1, []byte{END_LIST_TAG}, nil
		case '[':
			adv, token, err = 1, []byte{BEGIN_VECTOR_TAG}, nil
		case ']':
			adv, token, err = 1, []byte{END_VECTOR_TAG}, nil
		case '{':
			adv, token, err = 1, []byte{BEGIN_MAP_TAG}, nil
		case '}':
			adv, token, err = 1, []byte{END_MAP_TAG}, nil
		case ' ', '\n', '\r', '\t':
			adv, token, err = skipWhitespace(data)
		case '"':
			adv, token, err = parseString(data)
		case '\'':
			adv, token, err = 1, []byte{QUOTE_TAG}, nil
		default:
			adv, token, err = parseAtom(data, eofp)
		}
		return
	}
	dr.scanner.Split(tokenizer)
	return dr
}

func skipWhitespace(data []byte) (int, []byte, error) {
	for i, b := range data {
		if isWhitespace(b) {
		} else {
			return i, []byte{WHITESPACE_TAG}, nil
		}
	}
	return 0, nil, nil
}

func parseAtom(data []byte, eofp bool) (int, []byte, error) {
	buf := []byte{ATOM_TAG}
	for i, b := range data {
		if isWhitespace(b) || isDelimiter(b) {
			return i, buf, nil
		} else {
			buf = append(buf, b)
		}
	}
	if eofp {
		return len(data), buf, nil
	} else {
		return 0, nil, nil
	}
}

func parseString(data []byte) (int, []byte, error) {
	delim := data[0]
	skip := false
	buf := []byte{STRING_TAG}
	for i, b := range data[1:] {
		if b == delim && !skip {
			return i + 2, buf, nil
		}
		skip = false
		if b == '\\' {
			skip = true
			continue
		}
		buf = append(buf, b)
	}
	return 0, nil, nil
}

func (dr *DataReader) getToken() ([]byte, error) {
	if len(dr.lastToken) > 0 {
		tmp := dr.lastToken
		dr.lastToken = nil
		return tmp, nil
	} else {
		if dr.scanner.Scan() {
			b := dr.scanner.Bytes()
			return b, nil
		} else {
			return nil, io.EOF
		}
	}
}
func (dr *DataReader) ungetToken(tok []byte) {
	dr.lastToken = tok
}

func cons(first Object, rest Object) Pair {
	return Pair{first, rest}
}

func ToList(vec []Object) Object {
	var p Object
	p = NULL()
	for i := len(vec) - 1; i >= 0; i-- {
		v := vec[i]
		p = cons(v, p)
	}
	return p
}

func (dr *DataReader) ReadData() (Object, error) {
	tok, err := dr.getToken()
	for err == nil {

		switch tok[0] {
		case WHITESPACE_TAG:
			tok, err = dr.getToken()
			continue
		case ATOM_TAG:
			repr := string(tok[1:])
			if repr[0] == '#' {
				mac := repr[1:]
				if mac == "f" {
					return FALSE, nil
				}
				if mac == "t" {
					return TRUE, nil
				}
				if mac[0] == '\\' {
					mac = mac[1:]
					if mac == "newline" {
						return String("\n"), nil
					}
					if mac == "space" {
						return String(" "), nil
					}
					if len(mac) == 1 {
						return String(mac), nil
					}
				}
				return nil, Error{fmt.Sprintf("Unhandled reader macro: %s", repr)}
			} else {
				num, nerr := strconv.ParseFloat(repr, 64)
				if nerr == nil {
					return Number(num), nil
				}
			}
			//keyword?
			return Intern(dr.ns, repr), nil
		case STRING_TAG:
			return String(tok[1:]), nil
		case QUOTE_TAG:
			tmp, err2 := dr.ReadData()
			if err2 != nil {
				return nil, err2
			}
			return cons(Intern(dr.ns, "quote"), cons(tmp, NULL())), nil
		case BEGIN_LIST_TAG:
			vec := make([]Object, 0)
			tok, err = dr.getToken()
			for err == nil {
				if len(tok) == 0 {
					err = Error{"Unterminated list"}
					break
				}
				if len(tok) > 0 && tok[0] == END_LIST_TAG {
					return ToList(vec), nil
				}
				dr.ungetToken(tok)
				tmp, err := dr.ReadData()
				if err != nil {
					return nil, err
				}
				vec = append(vec, tmp)
				tok, err = dr.getToken()
			}
			return nil, err
		case END_LIST_TAG:
			return nil, Error{"Unexpected list terminator"}
		default:
			fmt.Println("Hmm:", string(tok[1:]))
			return nil, Error{"Hmm"}
		}
		fmt.Println("returning nil")
		return nil, nil
	}
	return nil, err
}

func Decode(in io.Reader, ns Namespace) (Object, error) {
	dr := MakeDataReader(in, ns)
	return dr.ReadData()
}
