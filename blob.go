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

// Blob - create a new blob, using the specified byte slice as the data. The data is not copied.
func Blob(bytes []byte) *LOB {
	b := new(LOB)
	b.Type = BlobType
	b.Value = bytes
	return b
}

// MakeBlob - create a new blob of the given size. It will be initialized to all zeroes
func MakeBlob(size int) *LOB {
	el := make([]byte, size)
	return Blob(el)
}

// EmptyBlob - a blob with no bytes
var EmptyBlob = MakeBlob(0)

// ToBlob - convert argument to a blob, if possible.
func ToBlob(obj *LOB) (*LOB, error) {
	switch obj.Type {
	case BlobType:
		return obj, nil
	case StringType:
		return Blob([]byte(obj.text)), nil //this copies the data
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
		val, err := AsByteValue(el[i])
		if err != nil {
			return nil, err
		}
		b[i] = val
	}
	return Blob(b), nil
}
