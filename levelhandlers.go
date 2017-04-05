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
	"net/http"
	"strconv"
)

type socketMessage struct {
	Type string
	Data interface{}
}

func setLevelHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")
	if token == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("no token provided"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	attr, authed := isTokenAuthed(token)
	if !authed {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("token is not authed"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Infof("Got authed token %s with note '%s'", token, attr.Note)
	lvlstr := r.Header.Get("New-Level")
	if lvlstr == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("you must provide a New-Level header"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newlvl, err := strconv.Atoi(lvlstr)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("error processing New-Level: " + err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	numlvls := getNumLevels()
	if newlvl > numlvls-1 {
		newlvl = newlvl - 1
	}

	level = newlvl

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Level set successfully"))

	go webSocketPool.broadcastMessage(&socketMessage{Type: "levelupdate", Data: map[string]interface{}{"level": level}})
	return

}

func increaseLevelHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")
	if token == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("no token provided"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	_, authed := isTokenAuthed(token)
	if !authed {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("token is not authed"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	level++
	numlvls := getNumLevels()
	if level > numlvls-1 {
		level = numlvls - 1
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Level set successfully"))
	go webSocketPool.broadcastMessage(&socketMessage{Type: "levelupdate", Data: map[string]interface{}{"level": level}})
	return

}

func decreaseLevelHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Token")
	if token == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("no token provided"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	attr, authed := isTokenAuthed(token)
	if !authed {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("token is not authed"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	level--
	if level < 0 {
		level = 0
	}
	log.Infof("%s setting updating level to %d", attr.Note, level)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Level set successfully"))
	go webSocketPool.broadcastMessage(&socketMessage{Type: "levelupdate", Data: map[string]interface{}{"level": level}})
	return

}

func levelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	serveFile(w, r, "levels.json")
}
