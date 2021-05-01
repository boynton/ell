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
	"fmt"
	. "github.com/boynton/ell/data"
)

// this type is similar to what an extension type (outside the ell package) would look like:
// the Value field of Object stores a pointer to the types specific data

// ChannelType - the type of Ell's channel object
var ChannelType Value = Intern("<channel>")

type Channel struct {
	name    string
	bufsize int
	channel chan Value // non-nil for channels
}

func (ch *Channel) Type() Value {
	return ChannelType
}

func (ch *Channel) String() string {
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

func (ch1 *Channel) Equals(another Value) bool {
	if ch2, ok := another.(*Channel); ok {
		return ch1 == ch2
	}
	return false
}

// Channel - create a new channel with the given buffer size and name
func NewChannel(bufsize int, name string) *Channel {
	return &Channel{name: name, bufsize: bufsize, channel: make(chan Value, bufsize)}
}

// ChannelValue - return the Go channel object for the Ell channel
func ChannelValue(obj Value) chan Value {
	if v, ok := obj.(*Channel); ok {
		return v.channel
	}
	return nil
}

// CloseChannel - close the channel object
func CloseChannel(obj Value) {
	if v, ok := obj.(*Channel); ok {
		c := v.channel
		if c != nil {
			v.channel = nil
			close(c)
		}
	}
}
