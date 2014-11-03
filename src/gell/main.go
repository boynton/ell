package main

import (
	"bufio"
	. "github.com/boynton/gell"
	"io"
	"os"
	"strings"
)

func main() {
	Println(len(os.Args), os.Args)
		var dr DataReader
		if len(os.Args) < 2 {
			s := "(foo 23 'bar\n(list \"foo\" 57.5))"
			dr = MakeDataReader(strings.NewReader(s))
			//Println("REPL NYI. Please provide a filename")
		} else {
			fi, err := os.Open(os.Args[1])
			if err != nil {
				panic(err)
			}
			r := bufio.NewReader(fi)
			dr = MakeDataReader(r)
		}
		data, err := dr.ReadData()
		for err == nil {
			Println("=============>", data, "is a", data.Type())
			data, err = dr.ReadData()
		}
		if err != nil {
			if err != io.EOF {
				Println("***", err)
			}
		}
	/*
	var o LObject
	n1 := LNumber(23)

	Println("Here is a number:", n1)
	pair := Cons(LNumber(23), LNumber(57))
	Println("Here is a pair (improper list):", pair)
	Println("Length of it:", Length(pair))
	Println("Length of NIL:", Length(NIL))
	lst := Cons(n1, NIL)
	lst = Cons(LNumber(57), lst)
	Println("Here is a list:", lst)
	Println("Length of it:", Length(lst))
	Println("Here is a longer list:", List(LNumber(23), lst, LNumber(57)))
	vec := Vector(LNumber(23), lst, LNumber(57))
	Println("Here is a vector:", vec)
	Println("Length of it:", Length(vec))
	Println("second element of it:", vec.Ref(1))

	//example type switch. I don't need tags!
	o = lst
	switch o.(type) {
		case LNil:
			Println("o is nil")
		case LSymbol:
			Println("o is a symbol")
		case LNumber:
			Println("o is a number")
		case LList:
			Println("o is a list")
		case LVector:
			Println("o is a vector")
        default:
            Println("o is ?")
	}

	m := Map(Intern("x"), LNumber(23), Intern("y"), LNumber(57))
	Println("Here is a map:", m)
	Println("Length of it:", Length(m))
	Println("map.Get(x):", m.Get(Intern("x")))
	Println("map.Get(y):", m.Get(Intern("y")))
	Println("map.Get(z):", m.Get(Intern("z")))

	Println("here is NIL:", NIL)
	Println("here is NIL.Type():", NIL.Type())
	Println("here is TRUE:", TRUE)
	Println("here is FALSE:", FALSE)
	Println("here is TRUE.Type():", TRUE.Type())
	Println("here is IsBoolean(TRUE):", IsBoolean(TRUE))

	Println("here is 'glorp:", Intern("glorp"))
	//Println("here is a simple cons: ", NIL)
	sym1 := Intern("true")
	sym2 := Intern("true")
	sym3 := Intern("false")
	Println("The type of 'true' symbol is ", sym1.Type())
	if IsSymbol(sym1) {
		Println("Symbol is indeed a symbol")
	} else {
		Println("Whoops, symbol type isn't what I expected")
	}
	Println("Here is the true symbol: ", sym1)
	if sym1 == sym2 {
		Println("syms that are supposed to be equal are")
	} else {
		Println("whoops: syms don't match")
	}
	if sym1 == sym3 {
		Println("whoops: syms match when they shouldn't")
	} else {
		Println("syms that are supposed to be different are")
	}
	*/
	return
}
