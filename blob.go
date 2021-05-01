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

import(
	"fmt"
	
	. "github.com/boynton/ell/data"
)

var BlobType Value = Intern("<blob>")

type Blob struct {
	Value []byte
}

func (b *Blob) Type() Value {
	return BlobType
}

func (b *Blob) String() string {
	//notation.go can handle it as an instance: #<blob>"base64 string value"
	//but here, we would return a nonreadable thing
	return fmt.Sprintf("#[Blob %d bytes]", len(b.Value))
}
func (b *Blob) Equals(another Value) bool {
	return false //FIXME
}

// Blob - create a new blob, using the specified byte slice as the data. The data is not copied.
func NewBlob(bytes []byte) *Blob {
	return &Blob{Value: bytes}
}

// MakeBlob - create a new blob of the given size. It will be initialized to all zeroes
func MakeBlob(size int) *Blob {
	el := make([]byte, size)
	return NewBlob(el)
}

// EmptyBlob - a blob with no bytes
var EmptyBlob = MakeBlob(0)

// ToBlob - convert argument to a blob, if possible.
func ToBlob(obj Value) (*Blob, error) {
	switch p := obj.(type) {
	case *Blob:
		return p, nil
	case *String:
		return NewBlob([]byte(p.Value)), nil //this copies the data
	case *Vector:
		return vectorToBlob(p)
	default:
		return nil, NewError(ArgumentErrorKey, "to-blob expected <blob> or <string>, got a ", obj.Type())
	}
}

func vectorToBlob(vec *Vector) (*Blob, error) {
	el := vec.Elements
	n := len(el)
	b := make([]byte, n, n)
	for i := 0; i < n; i++ {
		val, err := AsByteValue(el[i])
		if err != nil {
			return nil, err
		}
		b[i] = val
	}
	return NewBlob(b), nil
}
