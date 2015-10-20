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

package ell

import (
	"fmt"
)

//
// this type is similar to what an extension type (outside the ell package) would look like:
//   the Value field of LOB stores a pointer to the types specific data

var ChannelType = Intern("<channel>")

type channel struct {
	name    string
	bufsize int
	channel chan *LOB // non-nil for channels
}

func (ch *channel) String() string {
	s := "#[channel"
	if ch.name != "" {
		s += " " + ch.name
	}
	if ch.bufsize > 0 {
		s += fmt.Sprintf(" [%d]", ch.bufsize)
	}
	if ch.channel == nil {
		s += " CLOSED"
	}
	return s + "]"
}

func Channel(bufsize int, name string) *LOB {
	return NewObject(ChannelType, &channel{name: name, bufsize: bufsize, channel: make(chan *LOB, bufsize)})
}

func ChannelValue(obj *LOB) chan *LOB {
	if obj.Value == nil {
		return nil
	}
	v, _ := obj.Value.(*channel)
	return v.channel
}

func CloseChannel(obj *LOB) {
	v, _ := obj.Value.(*channel)
	if v != nil {
		c := v.channel
		if c != nil {
			v.channel = nil
			close(c)
		}
	}
}
