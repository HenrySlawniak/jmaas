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
	"encoding/gob"
	"github.com/go-playground/log"
	"math/rand"
	"os"
)

type tokenAttr struct {
	Level int
	Note  string
}

type tokenList map[string]tokenAttr

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func printTokens() {
	f, err := os.OpenFile("tokens.gob", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tokens := tokenList{}
	decoder := gob.NewDecoder(f)
	decoder.Decode(&tokens)

	log.Debug(tokens)
}

func addNewAuthedToken(note string) string {
	f, err := os.OpenFile("tokens.gob", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	tokens := tokenList{}
	decoder := gob.NewDecoder(f)
	decoder.Decode(&tokens)
	f.Close()

	token := randStringRunes(25)
	tokens[token] = tokenAttr{Level: 1, Note: note}

	f, err = os.OpenFile("tokens.gob", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	encoder := gob.NewEncoder(f)
	encoder.Encode(tokens)
	f.Close()

	return token
}

func isTokenAuthed(token string) (tokenAttr, bool) {
	f, err := os.OpenFile("tokens.gob", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tokens := tokenList{}
	decoder := gob.NewDecoder(f)
	decoder.Decode(&tokens)

	attr, exists := tokens[token]
	return attr, exists && attr.Level > 0
}
