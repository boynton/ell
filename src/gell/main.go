package main

import (
	. "github.com/boynton/gell"
	//	"github.com/davecheney/profile"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		Println("REPL NYI. Please provide a filename")
	} else {
		/*
			if len(os.Args) > 2 {
				cfg := profile.Config{
					CPUProfile:     true,
					ProfilePath:    ".",  // store profiles in current directory
					NoShutdownHook: true, // do not hook SIGINT
				}
				defer profile.Start(&cfg).Stop()
			}
		*/
		filename := os.Args[1]
		var prims Primitives
		if strings.HasSuffix(filename, ".ell") {
			prims = NewEllPrimitives()
		} else if strings.HasSuffix(filename, ".scm") {
			prims = NewSchemePrimitives()
		}
		val, err := RunModule(filename, prims)
		if err != nil {
			Println("*** ", err)
		} else if val != nil {
			//nil is used ot mean no value. Different thatn NIL, which means an Ell nil value
			//Println("returned ", val)
		}
	}
}
