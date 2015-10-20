/*
Copyright 2015 Lee Boynton

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
	"flag"
	"github.com/boynton/ell"
	//"github.com/davecheney/profile"
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
	interactive := len(args) == 0
	ell.SetFlags(*pOptimize, *pVerbose, *pDebug, *pTrace, interactive)
	ell.Init(nil)

	if len(args) > 0 {
		if *pCompile {
			//just compile and print LAP code
			for _, filename := range args {
				lap, err := ell.CompileFile(filename)
				if err != nil {
					ell.Fatal("*** ", err)
				}
				ell.Println(lap)
			}
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
			ell.Run(args)
		}
	} else {
		if !*pNoInit {
			home := os.Getenv("HOME")
			ellini := filepath.Join(home, ".ell")
			_, err := os.Stat(ellini)
			if err == nil {
				err := ell.Load(ellini)
				if err != nil {
					ell.Fatal("*** ", err)
				}
			}
		}
		ell.ReadEvalPrintLoop()
	}
	ell.Cleanup()
}
