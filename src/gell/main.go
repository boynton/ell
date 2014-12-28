package main

import (
	. "github.com/boynton/gell"
	//	"github.com/davecheney/profile"
	"flag"
	"os"
	//	"strings"
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
		if *pVerbose {
			SetVerbose(true)
		}
		environment := NewEnvironment("main", EllPrimitiveFunctions(), EllPrimitiveMacros())
		for _, filename := range args {
			if *pCompile {
				//just compile and print LAP code
				lap, err := environment.CompileFile(filename)
				if err != nil {
					Println("*** ", err)
					os.Exit(1)
				}
				Println(lap)
			} else {
				//this executes the file
				err := environment.LoadModule(filename)
				if err != nil {
					Println("*** ", err)
					os.Exit(1)
				}
			}
		}
	}
}
