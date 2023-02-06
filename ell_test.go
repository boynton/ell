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
package ell

import (
	"testing"

	. "github.com/boynton/ell/data"
)

func testType(t *testing.T, name string, sym Value) {
	if sym != Intern(name) {
		t.Error("type is not ", name, ":", sym)
	}
}

func testIdentical(t *testing.T, o1 Value, o2 Value) {
	if o1 != o2 {
		t.Error("objects should be identical but are not:", o1, "and", o2)
	}
}

func testNotIdentical(t *testing.T, o1 Value, o2 Value) {
	if o1 == o2 {
		t.Error("objects should not be identical but are:", o1, "and", o2)
	}
}

func TestNull(t *testing.T) {
	n1 := Null
	testIdentical(t, n1, Null)
	testNotIdentical(t, Null, nil)
	testType(t, "<null>", Null.Type())
	if n1 != Null {
		t.Error("nil isn't Null")
	}
}

func TestBooleans(t *testing.T) {
	b1 := True
	b2 := False
	testType(t, "<boolean>", True.Type())
	testType(t, "<boolean>", False.Type())
	testIdentical(t, True.Type(), False.Type())
	testIdentical(t, b1, True)
	testIdentical(t, b2, False)
	testNotIdentical(t, b1, b2)
	testType(t, "<boolean>", b1.Type())
	testType(t, "<boolean>", b2.Type())
}
