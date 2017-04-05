// Copyright (c) 2017 Henry Slawniak <https://henry.computer/>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"github.com/go-playground/log"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type socketConnection struct {
	mu   sync.RWMutex
	conn *websocket.Conn
	send chan interface{}
}

func (c *socketConnection) reader() {
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		log.Debugf("message from %s, %d:\"%s\"", c.conn.RemoteAddr().String(), msgType, msg)
	}
	c.conn.Close()
}

func (c *socketConnection) writer() {
	for {
		for msg := range c.send {
			c.mu.Lock()
			err := c.conn.WriteJSON(msg)
			c.mu.Unlock()
			if err != nil {
				log.Error(err)
				break
			}
		}
	}
	c.conn.Close()
}

type socketConnectionPool struct {
	mu          sync.RWMutex
	connections []*socketConnection
}

var webSocketPool = socketConnectionPool{
	mu:          sync.RWMutex{},
	connections: []*socketConnection{},
}

func (p *socketConnectionPool) registerConn(c *socketConnection) {
	log.Debugf("registering socket for %s", c.conn.RemoteAddr())
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections = append(p.connections, c)
}

func (p *socketConnectionPool) unregisterConn(c *socketConnection) {
	log.Debugf("unregistering socket for %s", c.conn.RemoteAddr())
	p.mu.Lock()
	defer p.mu.Unlock()
	out := []*socketConnection{}
	for _, conn := range p.connections {
		if conn != c {
			out = append(out, conn)
		}
	}
	p.connections = out
}

func (p *socketConnectionPool) broadcastMessage(msg interface{}) {
	log.Debug(msg)
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, c := range p.connections {
		c.send <- msg
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Socket opened from %s", r.RemoteAddr)

	socket := &socketConnection{
		mu:   sync.RWMutex{},
		conn: conn,
		send: make(chan interface{}),
	}

	webSocketPool.registerConn(socket)
	defer webSocketPool.unregisterConn(socket)

	go socket.writer()
	socket.send <- &socketMessage{Type: "levelupdate", Data: map[string]interface{}{"level": level}}

	socket.reader()
}
