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
)

func httpClientOperation(method string, url string, headers *Object, data *Object) (*Object, error) {
	client := &http.Client{}
	var bodyReader io.Reader
	bodyLen := 0
	if data != nil {
		tmp := []byte(data.text)
		bodyLen = len(tmp)
		if bodyLen > 0 {
			bodyReader = bytes.NewBuffer(tmp)
		}
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if headers != nil {
		for k, v := range headers.bindings {
			ks := k.toObject().text
			if v.Type == ListType {
				vs := v.car.String()
				req.Header.Set(ks, vs)
				for v.cdr != EmptyList {
					v = v.cdr
					req.Header.Add(ks, v.car.String())
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
			s := MakeStruct(3)
			Put(s, Intern("status:"), Number(float64(res.StatusCode)))
			bodyLen := len(bodyBytes)
			if bodyLen > 0 {
				Put(s, Intern("body:"), Blob(bodyBytes))
			}
			if len(res.Header) > 0 {
				headers = MakeStruct(len(res.Header))
				for k, v := range res.Header {
					var values []*Object
					for _, val := range v {
						values = append(values, String(val))
					}
					Put(headers, String(k), ListFromValues(values))
				}
				Put(s, Intern("headers:"), headers)
			}
			return s, nil
		}
	}
	return nil, err
}

func httpServer(port int, handler *Object) (*Object, error) {
	glue := func(w http.ResponseWriter, r *http.Request) {
		headers := MakeStruct(10)
		for k, v := range r.Header {
			var values []*Object
			for _, val := range v {
				values = append(values, String(val))
			}
			Put(headers, String(k), ListFromValues(values))
		}
		var body *Object
		method := strings.ToUpper(r.Method)
		switch method {
		case "POST", "PUT":
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Cannot decode body for " + r.Method + " request"))
				return
			}
			body = Blob(bodyBytes)
		}
		req, _ := Struct([]*Object{Intern("headers:"), headers})
		if body != nil {
			Put(req, Intern("body:"), body)
		}
		Put(req, Intern("method:"), String(method))
		Put(req, Intern("path:"), String(r.URL.Path))
		if r.URL.Scheme != "" {
			Put(req, Intern("scheme:"), String(r.URL.Scheme))
		}
		if r.URL.User != nil {
			//this is a *url.Userinfo
		}
		if r.URL.Host != "" {
			Put(req, Intern("host:"), String(r.URL.Host))
		}
		if r.URL.RawQuery != "" {
			Put(req, Intern("query:"), String(r.URL.RawQuery))
		}
		args := []*Object{req}
		res, err := exec(handler.code, args)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		if !IsStruct(res) {
			w.WriteHeader(500)
			w.Write([]byte("Handler did not return a struct"))
			return
		}
		headers = structGet(res, Intern("headers:"))
		body = structGet(res, Intern("body:"))
		status := structGet(res, Intern("status:"))
		if IsStruct(headers) {
			//fix: multiple values for a header
			for k, v := range headers.bindings {
				ks := headerString(k.toObject())
				vs := v.String()
				w.Header().Set(ks, vs)
			}
		}
		if IsString(body) {
			bodylen := len(body.text)
			w.Header().Set("Content-length", fmt.Sprint(bodylen))
			if status != nil {
				nstatus := int(status.fval)
				if nstatus != 0 && nstatus != 200 {
					w.WriteHeader(nstatus)
				}
			}
			if bodylen > 0 {
				w.Write([]byte(body.text))
			}
		} else {
			if status != nil {
				nstatus := int(status.fval)
				if nstatus != 0 && nstatus != 200 {
					w.WriteHeader(nstatus)
				}
			}
		}
	}
	http.HandleFunc("/", glue)
	//if verbose {
	println("[web server running at ", fmt.Sprintf(":%d", port), "]")
	//}
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	//no way to stop it
	return Null, nil
}

func headerString(obj *Object) string {
	switch obj.Type {
	case StringType, SymbolType:
		return obj.text
	case KeywordType:
		return unkeywordedString(obj)
	default:
		s, err := ToString(obj)
		if err != nil {
			return typeNameString(obj.Type.text)
		}
		return s.text
	}
}

var TcpConnectionType = Intern("<tcp-connection>")

func Connection(con net.Conn, endpoint string) *Object {
	inchan := Channel(10, "input")
	outchan := Channel(10, "output")
	go tcpReader(con, inchan)
	go tcpWriter(con, outchan)
	name := fmt.Sprintf("connection on %s", endpoint)
	connection := new(Object)
	connection.Type = TcpConnectionType
	s, _ := Struct([]*Object{Intern("input:"), inchan, Intern("output:"), outchan, Intern("name:"), String(name)})
	connection.car = s
	connection.Value = con
	return connection
}

func closeConnection(conobj *Object) {
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

func tcpReader(conn net.Conn, inchan *Object) {
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

func tcpWriter(con net.Conn, outchan *Object) {
	for {
		var packet *Object
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

func tcpListener(listener net.Listener, acceptChannel *Object, endpoint string) (*Object, error) {
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
