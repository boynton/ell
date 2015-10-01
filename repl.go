package main

import (
	"errors"
	"github.com/boynton/repl"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ellHandler struct {
	buf string
}

func (ell *ellHandler) Eval(expr string) (string, bool, error) {
	//return result, needMore, error
	for checkInterrupt() {
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
		lexpr, err := read(whole, nil)
		ell.buf = ""
		if err == nil {
			val, err := eval(lexpr)
			if err == nil {
				result := ""
				result = "= " + write(val)
				return result, false, nil
			}
			return "", false, err
		}
		return "", false, err
	}
}

func (ell *ellHandler) Reset() {
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

func (ell *ellHandler) completePrefix(expr string) (string, bool) {
	prefix := ""
	funPosition := false
	exprLen := len(expr)
	if exprLen > 0 {
		i := exprLen - 1
		ch := expr[i]
		if !isWhitespace(ch) && !isDelimiter(ch) {
			if i > 0 {
				i--
				for {
					ch = expr[i]
					if isWhitespace(ch) || isDelimiter(ch) {
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
	return prefix, funPosition
}

func (ell *ellHandler) Complete(expr string) (string, []string) {
	matches := []string{}
	addendum := ""
	prefix, funPosition := ell.completePrefix(expr)
	candidates := map[LOB]bool{}
	if funPosition {
		for _, sym := range getKeywords() {
			str := sym.String()
			if strings.HasPrefix(str, prefix) {
				candidates[sym] = true
			}
		}
		for _, sym := range macros() {
			_, ok := candidates[sym]
			if !ok {
				str := sym.String()
				if strings.HasPrefix(str, prefix) {
					candidates[sym] = true
				}
			}
		}
	}
	for _, sym := range getGlobals() {
		_, ok := candidates[sym]
		if !ok {
			_, ok := candidates[sym]
			if !ok {
				str := sym.String()
				if strings.HasPrefix(str, prefix) {
					if funPosition {
						val := global(sym)
						if isFunction(val) {
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

func (ell *ellHandler) Prompt() string {
	prompt := global(internSymbol("*prompt*"))
	if prompt != nil {
		return prompt.String()
	}
	return "? "
}

func historyFileName() string {
	return filepath.Join(os.Getenv("HOME"), ".ell_history")

}
func (ell *ellHandler) Start() []string {
	content, err := ioutil.ReadFile(historyFileName())
	if err != nil {
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

func (ell *ellHandler) Stop(history []string) {
	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	content := strings.Join(history, "\n") + "\n"
	err := ioutil.WriteFile(historyFileName(), []byte(content), 0644)
	if err != nil {
		println("[warning: cannot write ", historyFileName(), "]")
	}
}

func readEvalPrintLoop() {
	handler := ellHandler{""}
	err := repl.REPL(&handler)
	if err != nil {
		println("REPL error: ", err)
	}
}

func exit(code int) {
	repl.Exit(code)
}
