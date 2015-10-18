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
	"github.com/boynton/ell"
	"flag"
	"os"
	"path/filepath"
)

func main() {
	pCompile := flag.Bool("c", false, "compile the file and output lap")
	pOptimize := flag.Bool("o", false, "optimize execution speed, should work for correct code, but doesn't check everything")
	pVerbose := flag.Bool("v", false, "verbose mode, print extra information")
	pDebug := flag.Bool("d", false, "debug mode, print extra information about compilation")
	pTrace := flag.Bool("t", false, "trace VM instructions as they get executed")
	pNoInit := flag.Bool("i", false, "disable initialization from the $HOME/.ell file")
	flag.Parse()
	args := flag.Args()

	ell.EllPath = os.Getenv("ELL_PATH")
	home := os.Getenv("HOME")
	ellini := filepath.Join(home, ".ell")
	if ell.EllPath == "" {
		ell.EllPath = "."
		homelib := filepath.Join(home, "lib/ell")
		_, err := os.Stat(homelib)
		if err == nil {
			ell.EllPath += ":" + homelib
		}
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			golibdir := filepath.Join(gopath, "src/github.com/boynton/gell/lib")
			_, err := os.Stat(golibdir)
			if err == nil {
				ell.EllPath += ":" + golibdir
			}
		}
	}
	ell.InitEnvironment()
	if len(args) < 1 {
//		interactive = true
		if !*pNoInit {
			_, err := os.Stat(ellini)
			if err == nil {
				err := ell.LoadModule(ellini)
				if err != nil {
					ell.Fatal("*** ", err)
				}
			}
		}
		ell.SetFlags(*pOptimize, *pVerbose, *pDebug, *pTrace, true)
/*		if *pOptimize {
			ell.SetOptimize(*pOptimize)
//			optimize = *pOptimize
		}
		if *pVerbose {
			ell.SetVerbose(*pVerbose)
//			verbose = *pVerbose
		}
		if *pDebug {
			ell.SetDebug(*pDebug)
//			debug = *pDebug
		}
		if *pTrace {
			ell.SetTrace(*pTrace)
//			trace = *pTrace
		}
*/
		ell.ReadEvalPrintLoop()
	} else {
		ell.SetFlags(*pOptimize, *pVerbose, *pDebug, *pTrace, false)
/*
		if *pOptimize {
			optimize = *pOptimize
		}
		if *pVerbose {
			verbose = *pVerbose
		}
		if *pDebug {
			debug = *pDebug
		}
		if *pTrace {
			trace = *pTrace
		}
		interactive = false
*/
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
				lap, err := ell.CompileFile(filename)
				if err != nil {
					ell.Fatal("*** ", err)
				}
				println(lap)
			} else {
				//this executes the file
				err := ell.LoadModule(filename)
				if err != nil {
					ell.Fatal("*** ", err.Error())
				}
			}
		}
	}
	ell.CleanupEnvironment()
}
