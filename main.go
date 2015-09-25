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
	"unsafe"
)

var verbose bool
var interactive bool

// Version - this version of gell
const Version = "gell v0.2"

// EllPath is the path where the library *.ell files can be found
var EllPath string

func fatal(args ...interface{}) {
	println(args...)
	exit(1)
}

var interrupts chan os.Signal

func main() {
	pCompile := flag.Bool("c", false, "compile the file and output lap")
	pVerbose := flag.Bool("v", false, "verbose mode, print extra information")
	pTrace := flag.Bool("t", false, "trace VM instructions as they get executed")
	pNoInit := flag.Bool("i", false, "disable initialization from the $HOME/.ell file")
	flag.Parse()
	args := flag.Args()

	EllPath = os.Getenv("ELL_PATH")
	home := os.Getenv("HOME")
	ellini := filepath.Join(home, ".ell")
	if EllPath == "" {
		EllPath = "."
		homelib := filepath.Join(home, "lib/ell")
		_, err := os.Stat(homelib)
		if err == nil {
			EllPath += ":" + homelib
		}
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			golibdir := filepath.Join(gopath, "src/github.com/boynton/gell/lib")
			_, err := os.Stat(golibdir)
			if err == nil {
				EllPath += ":" + golibdir
			}
		}
	}
	if *pVerbose {
		verbose = *pVerbose
	}
	if *pTrace {
		setTrace(*pTrace)
	}
	initEnvironment()
	if len(args) < 1 {
		interactive = true
		if !*pNoInit {
			_, err := os.Stat(ellini)
			if err == nil {
				err := loadModule(ellini)
				if err != nil {
					fatal("*** ", err)
				}
			}
		}
		interrupts = make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		defer signal.Stop(interrupts)
		readEvalPrintLoop()
	} else {
		interactive = false
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
		for _, filename := range args {
			if *pCompile {
				//just compile and print LAP code
				lap, err := compileFile(filename)
				if err != nil {
					fatal("*** ", err)
				}
				println(lap)
			} else {
				//this executes the file
				err := loadModule(filename)
				if err != nil {
					fatal("*** ", err)
				}
			}
		}
	}
}
