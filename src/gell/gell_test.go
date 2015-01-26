package main

import (
	. "github.com/boynton/gell"
	"testing"
)

func testType(t *testing.T, name string, sym LObject) {
	if !IsSymbol(sym) {
		t.Error("nil type is not a symbol:", sym)
	}
	if sym != Intern(name) {
		t.Error("type is not ", name, ":", sym)
	}
}

func testIdentical(t *testing.T, o1 LObject, o2 LObject) {
	if o1 != o2 {
		t.Error("objects should be identical but are not:", o1, "and", o2)
	}
}
func testNotIdentical(t *testing.T, o1 LObject, o2 LObject) {
	if o1 == o2 {
		t.Error("objects should not be identical but are:", o1, "and", o2)
	}
}

func TestNil(t *testing.T) {
	n1 := NIL
	testIdentical(t, n1, NIL)
	testNotIdentical(t, NIL, nil)
	testType(t, "null", NIL.Type())
	if n1 != NIL {
		t.Error("nil isn't")
	}
}

func TestBooleans(t *testing.T) {
	b1 := TRUE
	b2 := FALSE
	testType(t, "boolean", TRUE.Type())
	testType(t, "boolean", FALSE.Type())
	testIdentical(t, TRUE.Type(), FALSE.Type())
	testIdentical(t, b1, TRUE)
	testIdentical(t, b2, FALSE)
	testNotIdentical(t, b1, b2)
	if !IsBoolean(b1) {
		t.Error("boolean value isn't:", b1)
	}
	if !IsBoolean(b2) {
		t.Error("boolean value isn't:", b2)
	}
}
