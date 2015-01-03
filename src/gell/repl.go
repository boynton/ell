package main

import (
	"errors"
	. "github.com/boynton/gell"
	"github.com/boynton/repl"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EllHandler struct {
	env LModule
	buf string
}

func (ell *EllHandler) Eval(expr string) (string, bool, error) {
	//return result, needMore, error
	for ell.env.CheckInterrupt() {
	} //to clear out any that happened while sitting in getc
	whole := strings.Trim(ell.buf+expr, " ")
	opens := len(strings.Split(whole, "("))
	closes := len(strings.Split(whole, ")"))
	if opens > closes {
		ell.buf = whole + " "
		return "", true, nil
	} else if closes > opens {
		ell.buf = ""
		return "", false, errors.New("Unbalanced ')'")
	} else {
		//this is the normal case
		if whole == "" {
			return "", false, nil
		}
		lexpr, err := OpenInputString(whole).Read()
		ell.buf = ""
		if err == nil {
			val, err := ell.env.Eval(lexpr)
			if err == nil {
				result := ""
				if val != nil {
					result = "= " + Write(val)
				}
				return result, false, nil
			} else {
				return "", false, err
			}
		}
		return "", false, err
	}
}

func (ell *EllHandler) Reset() {
	ell.buf = ""
}

func greatestCommonPrefixLength(s1 string, s2 string) int {
	max := len(s1)
	l2 := len(s2)
	if l2 < max {
		max = l2
	}
	for i := 0; i < max; i++ {
		if s1[i] != s2[i] {
			return i - 1
		}
	}
	return max
}

func greatestCommonPrefix(prefix string, matches []string) string {
	//i.e. start = "pri", matches = ["print", "println"] -> "print"
	switch len(matches) {
	case 0:
		return ""
	case 1:
		return matches[0]
	default:
		s := matches[0]
		max := len(matches)
		greatest := len(s)
		for i := 1; i < max; i++ {
			n := greatestCommonPrefixLength(s, matches[i])
			if n < greatest {
				greatest = n
				s = s[:n+1]
			}
		}
		return s
	}
}

func (ell *EllHandler) Complete(expr string) (string, []string) {
	matches := []string{}
	addendum := ""
	exprLen := len(expr)
	prefix := ""
	funPosition := false
	//perhaps: detect if we are in a funarg position, then the "empty prefix" could show function arg usage
	if exprLen > 0 {
		i := exprLen - 1
		ch := expr[i]
		if !IsWhitespace(ch) && !IsDelimiter(ch) {
			if i > 0 {
				i--
				for {
					ch = expr[i]
					if IsWhitespace(ch) || IsDelimiter(ch) {
						funPosition = ch == '('
						prefix = expr[i+1:]
						break
					}
					i--
					if i < 0 {
						prefix = expr
						break
					}
				}
			} else {
				prefix = expr
			}
		}
	}
	candidates := map[LObject]bool{}
	if funPosition {
		for _, sym := range ell.env.Keywords() {
			str := sym.String()
			if strings.HasPrefix(str, prefix) {
				candidates[sym] = true
			}
		}
		for _, sym := range ell.env.Macros() {
			_, ok := candidates[sym]
			if !ok {
				str := sym.String()
				if strings.HasPrefix(str, prefix) {
					candidates[sym] = true
				}
			}
		}
	}
	for _, sym := range ell.env.Globals() {
		_, ok := candidates[sym]
		if !ok {
			_, ok := candidates[sym]
			if !ok {
				str := sym.String()
				if strings.HasPrefix(str, prefix) {
					if funPosition {
						val := ell.env.Global(sym)
						if IsFunction(val) {
							candidates[sym] = true
						}
					} else {
						candidates[sym] = true
					}
				}
			}
		}
	}
	for sym := range candidates {
		matches = append(matches, sym.String())

	}
	sort.Strings(matches)
	gcp := greatestCommonPrefix(prefix, matches)
	if len(gcp) > len(prefix) {
		addendum = gcp[len(prefix):]
	}
	return addendum, matches
}

func (ell *EllHandler) Prompt() string {
	prompt := ell.env.Global(Intern("*prompt*"))
	if prompt != nil {
		return prompt.String()
	} else {
		return "? "
	}
}

func HistoryFileName() string {
	return filepath.Join(os.Getenv("HOME"), ".ell_history")

}
func (ell *EllHandler) Start() []string {
	content, err := ioutil.ReadFile(HistoryFileName())
	if err != nil {
		//Println("[warning: cannot read ", HistoryFileName(), "]")
		return nil
	}
	s := strings.Split(string(content), "\n")
	var s2 []string
	for _, v := range s {
		if v != "" {
			s2 = append(s2, v)
		}
	}
	return s2
}

func (ell *EllHandler) Stop(history []string) {
	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	content := strings.Join(history, "\n") + "\n"
	err := ioutil.WriteFile(HistoryFileName(), []byte(content), 0644)
	if err != nil {
		Println("[warning: cannot write ", HistoryFileName(), "]")
	}
}

func REPL(environment LModule) {
	handler := EllHandler{environment, ""}
	err := repl.REPL(&handler)
	if err != nil {
		Println("REPL error: ", err)
	}
}
