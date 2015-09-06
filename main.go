/*
Copyright 2014 Lee Boynton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	//	"github.com/davecheney/profile"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
)

var verbose bool
var extendedInstructions = false

// Version - this version of gell
const Version = "gell v0.1"

// EllPath is the path where the library *.ell files can be found
var EllPath string

func main() {
	EllPath = os.Getenv("ELL_PATH")
	ellini := filepath.Join(os.Getenv("HOME"), ".ell")
	if EllPath == "" {
		filename, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		EllPath = ".:" + filepath.Join(filepath.Dir(filename), "src/github.com/boynton/gell/lib")
	}
	pCompile := flag.Bool("c", false, "compile the file and output lap")
	pVerbose := flag.Bool("v", false, "verbose mode, print extra information")
	pTrace := flag.Bool("t", false, "trace VM instructions as they get executed")
	pExtended := flag.Bool("e", false, "enable extended VM instructions for common primitive operations")
	pNoInit := flag.Bool("i", false, "disable initialization from the $HOME/.ell file")
	flag.Parse()
	args := flag.Args()
	if *pVerbose {
		verbose = *pVerbose
	}
	if *pTrace {
		setTrace(*pTrace)
	}
	if *pExtended {
		extendedInstructions = *pExtended
	}
	if len(args) < 1 {
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		defer signal.Stop(interrupts)
		environment := newEnvironment("main", Ell, interrupts)
		if !*pNoInit {
			_, err := os.Stat(ellini)
			if err == nil {
				err := environment.loadModule(ellini)
				if err != nil {
					println("*** ", err)
					os.Exit(1)
				}
			}
		}
		readEvalPrintLoop(environment)
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
		environment := newEnvironment("main", Ell, nil)
		for _, filename := range args {
			if *pCompile {
				//just compile and print LAP code
				lap, err := environment.compileFile(filename)
				if err != nil {
					println("*** ", err)
					os.Exit(1)
				}
				println(lap)
			} else {
				//this executes the file
				err := environment.loadModule(filename)
				if err != nil {
					println("*** ", err)
					os.Exit(1)
				}
			}
		}
	}
}
