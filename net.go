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
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	//	"net/http"
)

var TcpConnectionType = Intern("<tcp-connection>")

func Connection(con net.Conn, endpoint string) *LOB {
	inchan := Channel(10, "input")
	outchan := Channel(10, "output")
	go tcpReader(con, inchan)
	go tcpWriter(con, outchan)
	name := fmt.Sprintf("connection on %s", endpoint)
	connection := new(LOB)
	connection.Type = TcpConnectionType
	s, _ := Struct([]*LOB{Intern("input:"), inchan, Intern("output:"), outchan, Intern("name:"), String(name)})
	connection.car = s
	connection.Value = con
	return connection
}

func closeConnection(conobj *LOB) {
	if conobj.Value != nil {
		inchan, err := Get(conobj, Intern("input:"))
		if err == nil {
			CloseChannel(inchan)
		}
		outchan, err := Get(conobj, Intern("output:"))
		if err == nil {
			CloseChannel(outchan)
		}
		con, ok := conobj.Value.(net.Conn)
		if ok {
			con.Close()
			conobj.Value = nil
		}
	}
}

// MaxFrameSize is an arbitrary limit to the tcp server framesize, to avoid trouble
const MaxFrameSize = 1000000

func tcpReader(conn net.Conn, inchan *LOB) {
	r := bufio.NewReader(conn)
	for {
		count, err := binary.ReadVarint(r)
		if err != nil {
			CloseChannel(inchan)
			return
		}
		if count < 0 || count > MaxFrameSize {
			println("Bad frame size: ", count)
			CloseChannel(inchan)
			return
		}
		buf := make([]byte, count, count)
		cur := buf[:]
		remaining := int(count)
		offset := 0
		for remaining > 0 {
			n, err := r.Read(cur)
			if err != nil {
				CloseChannel(inchan)
				return
			}
			remaining -= n
			offset += n
			cur = buf[offset:]
		}
		packet := Blob(buf)
		ch := ChannelValue(inchan)
		if ch != nil {
			ch <- packet
		}
	}
}

func tcpWriter(con net.Conn, outchan *LOB) {
	for {
		var packet *LOB
		ch := ChannelValue(outchan)
		if ch != nil {
			packet = <-ch
		}
		if packet == nil {
			return
		}
		data := []byte(packet.text)
		count := len(data)
		header := make([]byte, 8)
		n := binary.PutVarint(header, int64(count))
		n, err := con.Write(header[:n])
		if err != nil {
			CloseChannel(outchan)
			return
		}
		n, err = con.Write([]byte(data))
		if n != len(data) || err != nil {
			CloseChannel(outchan)
			return
		}
	}
}

func tcpListener(listener net.Listener, acceptChannel *LOB, endpoint string) (*LOB, error) {
	for {
		con, err := listener.Accept()
		if err != nil {
			return nil, err
		}
		ch := ChannelValue(acceptChannel)
		if ch != nil {
			ch <- Connection(con, endpoint)
		}
	}
}
