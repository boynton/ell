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
//	"strings"
)

func newBlob(bytes []byte) *LOB {
	b := newLOB(BlobType)
	b.text = string(bytes)
	return b
}

func makeBlob(size int) *LOB {
	el := make([]byte, size)
	return newBlob(el)
}

// EmptyBlob - a blob with no bytes
var EmptyBlob = makeBlob(0)

func toBlob(obj *LOB) (*LOB, error) {
	switch obj.Type {
	case BlobType:
		return obj, nil
	case StringType:
		return newBlob([]byte(obj.text)), nil //this copies the data
	case VectorType:
		return vectorToBlob(obj)
	default:
		return nil, Error(ArgumentErrorKey, "to-blob expected <blob> or <string>, got a ", obj.Type)
	}
}

func vectorToBlob(obj *LOB) (*LOB, error) {
	el := obj.elements
	n := len(el)
	b := make([]byte, n, n)
	for i := 0; i < n; i++ {
		val, err := byteValue(el[i])
		if err != nil {
			return nil, err
		}
		b[i] = val
	}
	return newBlob(b), nil
}
