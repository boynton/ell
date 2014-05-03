package gell

import (
	"fmt"
)

func Println(args ...interface{}) {
	fmt.Println(args)
}

func Exec(expr Object) (Object, Error) {
	return nil, Error{"NYI"}
}
