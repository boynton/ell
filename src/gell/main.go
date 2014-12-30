package main

import (
	. "github.com/boynton/gell"
	//	"github.com/davecheney/profile"
	"flag"
	"os"
	"os/signal"
)

func main() {
	pCompile := flag.Bool("c", false, "compile the file and output lap")
	pVerbose := flag.Bool("v", false, "verbose mode, print extra information")
	pTrace := flag.Bool("t", false, "trace VM instructions as they get executed")
	pExtended := flag.Bool("e", false, "enable extended VM instructions for common primitive operations")
	flag.Parse()
	args := flag.Args()
	if *pVerbose {
		SetVerbose(*pVerbose)
	}
	if *pTrace {
		SetTrace(*pTrace)
	}
	if *pExtended {
		EnableExtendedInstructions(*pExtended)
	}
	if len(args) < 1 {
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		defer signal.Stop(interrupts)
		environment := NewEnvironment("main", Ell, interrupts)
		REPL(environment)
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
		environment := NewEnvironment("main", Ell, nil)
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
