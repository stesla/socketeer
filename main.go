// Copyright (c) 2012, Samuel Tesla <samuel.tesla@gmail.com>
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//  * Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.

//  * Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the
//    distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

var ListenAddr *string = flag.String("http", ":6060", "[addr]:port")
var LineEnding *string = flag.String("line", "\r\n", "line ending")

func init() {
	flag.Parse()
}

func main() {
	http.Handle("/bytes", handleConnection(bytewise))
	http.Handle("/lines", handleConnection(linewise))
	if err := http.ListenAndServe(*ListenAddr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type handlerFunc func(down *websocket.Conn, up net.Conn)

func handleConnection(handler handlerFunc) http.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		params, perr := url.ParseQuery(ws.Config().Location.RawQuery)
		host := params.Get("host")
		port := params.Get("port")
		if perr != nil || host == "" || port == "" {
			ws.WriteClose(4000)
			return
		}
		conn, cerr := net.Dial("tcp", host + ":" + port)
		if cerr != nil {
			ws.WriteClose(4001)
			return
		}
		defer conn.Close()
		handler(ws, conn)
	})
}

func bytewise(down *websocket.Conn, up net.Conn) {
	exit := make(chan bool)
	go func() { io.Copy(down, up); exit <- true }()
	go func() { io.Copy(up, down); exit <- true }()
	<-exit
}

func linewise(down *websocket.Conn, up net.Conn) {
	params, _ := url.ParseQuery(down.Config().Location.RawQuery)
	lineEnding := *LineEnding
	if val := params.Get("lineEnding"); val != "" {
		lineEnding = val
	}
	exit := make(chan bool)
	go func() { io.Copy(down, up); exit <- true }()
	go func() {
		for {
			var msg string
			if rerr := websocket.Message.Receive(down, &msg); rerr != nil {
				break;
			}
			if _, werr := up.Write([]byte(msg + lineEnding)); werr != nil {
				break;
			}
		}
		exit <- true
	}()
	<-exit
}
