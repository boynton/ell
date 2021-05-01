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
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	. "github.com/boynton/ell/data"
)

var HTTPErrorKey = Intern("http-error:")

func httpClientOperation(method string, url string, headers *Struct, data *String) (*Struct, error) {
	client := &http.Client{}
	var bodyReader io.Reader
	bodyLen := 0
	if data != nil {
		tmp := []byte(data.Value)
		bodyLen = len(tmp)
		if bodyLen > 0 {
			bodyReader = bytes.NewBuffer(tmp)
		}
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		for k, v := range headers.Bindings {
			ks := StringValue(k.ToValue())
			if p, ok := v.(*List); ok {
				vs := p.Car.String()
				req.Header.Set(ks, vs)
				for p.Cdr != EmptyList {
					p = p.Cdr
					req.Header.Add(ks, p.Car.String())
				}
			} else {
				req.Header.Set(ks, v.String())
			}
		}
	}
	if bodyLen > 0 {
		req.Header.Set("Content-Length", fmt.Sprint(bodyLen))
	}
	res, err := client.Do(req)
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err == nil {
			s := NewStruct()
			Put(s, Intern("status:"), Integer(res.StatusCode))
			bodyLen := len(bodyBytes)
			if bodyLen > 0 {
				Put(s, Intern("body:"), NewBlob(bodyBytes))
			}
			if len(res.Header) > 0 {
				headers = NewStruct()
				for k, v := range res.Header {
					var values []Value
					for _, val := range v {
						values = append(values, NewString(val))
					}
					Put(headers, NewString(k), ListFromValues(values))
				}
				Put(s, Intern("headers:"), headers)
			}
			return s, nil
		}
	}
	return nil, err
}

func httpServer(port int, handler *Function) (Value, error) {
	glue := func(w http.ResponseWriter, r *http.Request) {
		headers := NewStruct()
		for k, v := range r.Header {
			var values []Value
			for _, val := range v {
				values = append(values, NewString(val))
			}
			Put(headers, NewString(k), ListFromValues(values))
		}
		var body Value
		method := strings.ToUpper(r.Method)
		switch method {
		case "POST", "PUT":
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Cannot decode body for " + r.Method + " request"))
				return
			}
			body = NewBlob(bodyBytes)
		}
		req, _ := MakeStruct([]Value{Intern("headers:"), headers})
		if body != nil {
			Put(req, Intern("body:"), body)
		}
		Put(req, Intern("method:"), NewString(method))
		Put(req, Intern("path:"), NewString(r.URL.Path))
		if r.URL.Scheme != "" {
			Put(req, Intern("scheme:"), NewString(r.URL.Scheme))
		}
		if r.URL.User != nil {
			//this is a *url.Userinfo
		}
		if r.URL.Host != "" {
			Put(req, Intern("host:"), NewString(r.URL.Host))
		}
		if r.URL.RawQuery != "" {
			Put(req, Intern("query:"), NewString(r.URL.RawQuery))
		}
		args := []Value{req}
		res, err := exec(handler.code, args)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		p, ok := res.(*Struct)
		if !ok {
			//		if !IsStruct(res) {
			w.WriteHeader(500)
			w.Write([]byte("Handler did not return a struct"))
			return
		}
		headers2 := p.Get(Intern("headers:"))
		body = p.Get(Intern("body:"))
		status := p.Get(Intern("status:"))
		if headers, ok := headers2.(*Struct); ok {
			//fix: multiple values for a header
			for k, v := range headers.Bindings {
				ks := headerString(k.ToValue())
				vs := v.String()
				w.Header().Set(ks, vs)
			}
		}
		if s, ok := body.(*String); ok {
			bodylen := len(s.Value)
			w.Header().Set("Content-length", fmt.Sprint(bodylen))
			if status != nil {
				nstatus, _ := AsIntValue(status)
				if nstatus != 0 && nstatus != 200 {
					w.WriteHeader(nstatus)
				}
			}
			if bodylen > 0 {
				w.Write([]byte(s.Value))
			}
		} else {
			if status != nil {
				nstatus := IntValue(status)
				if nstatus != 0 && nstatus != 200 {
					w.WriteHeader(nstatus)
				}
			}
		}
	}
	http.HandleFunc("/", glue)
	//if verbose {
	fmt.Printf("[web server running at http://localhost:%d]\n", port)
	//}
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	//no way to stop it
	return Null, nil
}

func headerString(obj Value) string {
	switch p := obj.(type) {
	case *String:
		return p.Value
	case *Symbol:
		return p.Text
	case *Keyword:
		return p.Name()
	default:
		s, err := ToString(obj)
		if err != nil {
			return TypeNameOf(obj)
		}
		return s.Value
	}
}

var ConnectionType = Intern("<connection>")

type Connection struct {
	Name string
	In *Channel
	Out *Channel
	Con net.Conn
}

func (c *Connection) Type() Value {
	return ConnectionType
}
func (c *Connection) Equals(another Value) bool {
	return false
}
func (c *Connection) String() string {
	return "#[Connection]"
}

func NewConnection(con net.Conn, endpoint string) Value {
	inchan := NewChannel(10, "input")
	outchan := NewChannel(10, "output")
	go tcpReader(con, inchan)
	go tcpWriter(con, outchan)
	name := fmt.Sprintf("connection on %s", endpoint)
	return &Connection{
		Name: name,
		In: inchan,
		Out: outchan,
		Con: con,
	}
}

func closeConnection(obj Value) {
	if p, ok := obj.(*Connection); ok {
		if p.Con != nil {
			CloseChannel(p.In)
			CloseChannel(p.Out)
			p.Con.Close()
			p.Con = nil;
		}
	}
}

// MaxFrameSize is an arbitrary limit to the tcp server framesize, to avoid trouble
const MaxFrameSize = 1000000

func tcpReader(conn net.Conn, inchan Value) {
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
		packet := NewBlob(buf)
		ch := ChannelValue(inchan)
		if ch != nil {
			ch <- packet
		}
	}
}

func tcpWriter(con net.Conn, outchan Value) {
	for {
		var packet Value
		ch := ChannelValue(outchan)
		if ch != nil {
			packet = <-ch
		}
		if packet == nil {
			return
		}
		if p, ok := packet.(*String); ok {
			data := []byte(p.Value)
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
}

func tcpListener(listener net.Listener, acceptChannel Value, endpoint string) (Value, error) {
	for {
		con, err := listener.Accept()
		if err != nil {
			return nil, err
		}
		ch := ChannelValue(acceptChannel)
		if ch != nil {
			ch <- NewConnection(con, endpoint)
		}
	}
}
