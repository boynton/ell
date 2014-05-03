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
		dr = MakeDataReader(strings.NewReader(s), make(Namespace))
		//Println("REPL NYI. Please provide a filename")
	} else {
		fi, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		r := bufio.NewReader(fi)
		dr = MakeDataReader(r, make(Namespace))
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
	return
}
