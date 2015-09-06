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
	"testing"
)

func testType(t *testing.T, name string, sym lob) {
	if !isSymbol(sym) {
		t.Error("nil type is not a symbol:", sym)
	}
	if sym != intern(name) {
		t.Error("type is not ", name, ":", sym)
	}
}

func testIdentical(t *testing.T, o1 lob, o2 lob) {
	if o1 != o2 {
		t.Error("objects should be identical but are not:", o1, "and", o2)
	}
}
func testNotIdentical(t *testing.T, o1 lob, o2 lob) {
	if o1 == o2 {
		t.Error("objects should not be identical but are:", o1, "and", o2)
	}
}

func TestNil(t *testing.T) {
	n1 := NIL
	testIdentical(t, n1, NIL)
	testNotIdentical(t, NIL, nil)
	testType(t, "null", NIL.typeSymbol())
	if n1 != NIL {
		t.Error("nil isn't")
	}
}

func TestBooleans(t *testing.T) {
	b1 := TRUE
	b2 := FALSE
	testType(t, "boolean", TRUE.typeSymbol())
	testType(t, "boolean", FALSE.typeSymbol())
	testIdentical(t, TRUE.typeSymbol(), FALSE.typeSymbol())
	testIdentical(t, b1, TRUE)
	testIdentical(t, b2, FALSE)
	testNotIdentical(t, b1, b2)
	if !isBoolean(b1) {
		t.Error("boolean value isn't:", b1)
	}
	if !isBoolean(b2) {
		t.Error("boolean value isn't:", b2)
	}
}
