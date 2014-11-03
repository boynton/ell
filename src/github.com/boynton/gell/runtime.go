package gell

import (
	"fmt"
)

func Println(args ...interface{}) {
	fmt.Println(args)
}

func Exec(expr LObject) (LObject, LError) {
	return NIL, LError{"NYI"}
}
