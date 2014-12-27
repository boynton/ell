package main

import (
	. "github.com/boynton/gell"
	//	"github.com/davecheney/profile"
	"flag"
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
		pCompile := flag.Bool("c", false, "compile the file and output lap")
		pVerbose := flag.Bool("v", false, "verbose mode, print extra information")
		flag.Parse()
		args := flag.Args()
		filename := args[0]
		var prims map[string]Primitive
		if strings.HasSuffix(filename, ".scm") {
			prims = SchemePrimitives()
		} else { //assume Ell
			prims = EllPrimitives()
		}
		if *pVerbose {
			SetVerbose(true)
		}
		if *pCompile {
			code, err := LoadModule(filename, prims)
			if err != nil {
				Println("*** ", err)
				os.Exit(1)
			} else {
				Println(code.Decompile())
				os.Exit(0)
			}
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
