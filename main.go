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
	"crypto/sha512"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"io/ioutil"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const version = "1.1.0"

var devMode = flag.Bool("dev", false, "Puts the server in developer mode, will bind to :34265 and will not autocert")
var domains = flag.String("domain", "angrymills.net", "A comma-seperaated list of domains to get a certificate for.")
var client = &http.Client{}
var level = 0

func init() {
	gob.Register(tokenList{})
	rand.Seed(time.Now().UnixNano())
}

func getNumLevels() int {
	lvls := map[int]map[string]interface{}{}
	f, err := os.Open("levels.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.Decode(&lvls)
	return len(lvls)
}

func main() {
	flag.Parse()
	cLog := console.New()
	cLog.SetTimestampFormat(time.RFC3339)
	log.RegisterHandler(cLog, log.AllLevels...)

	log.Info("Starting The Josh Mills Anger Advisory System")

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler())
	mux.HandleFunc("/api/levels", levelHandler())
	mux.HandleFunc("/api/setlevel", setLevelHandler())
	mux.HandleFunc("/api/inclevel", increaseLevelHandler())
	mux.HandleFunc("/api/declevel", decreaseLevelHandler())
	mux.HandleFunc("/api/currentlevel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("%d", level)))
	})

	addNewAuthedToken("autogen")

	printTokens()

	if *devMode {
		srv := &http.Server{
			Addr:    ":34265",
			Handler: mux,
		}

		log.Info("Listening on :34265")
		srv.ListenAndServe()
	} else {
		httpSrv := &http.Server{
			Addr:    ":http",
			Handler: http.HandlerFunc(httpRedirectHandler),
		}

		go httpSrv.ListenAndServe()

		domainList := strings.Split(*domains, ",")
		for i, d := range domainList {
			domainList[i] = strings.TrimSpace(d)
		}

		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domainList...),
			Cache:      autocert.DirCache("certs"),
		}

		rootSrv := &http.Server{
			Addr:      ":https",
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
			Handler:   mux,
		}

		log.Info("Listening on :https")

		http2.ConfigureServer(rootSrv, &http2.Server{})
		rootSrv.ListenAndServeTLS("", "")
	}
}

func httpRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
}

func indexHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if _, err := os.Stat("./client" + path); err == nil {
			serveFile(w, r, "./client"+path)
		} else {
			serveFile(w, r, "./client/index.html")
		}
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	if path == "./client/" {
		path = "./client/index.html"
	}
	content, sum, mod, err := readFile(path)
	if err != nil {
		http.Error(w, "Could not read file", http.StatusInternalServerError)
		fmt.Printf("%s:%s\n", path, err.Error())
		return
	}
	if strings.Contains(path, ".html") {
		if pusher, ok := w.(http.Pusher); ok {
			if err := pusher.Push("/static/style.css", nil); err != nil {
				log.Warnf("Failed to push: %v", err)
			}
			if err := pusher.Push("/static/arrow.png", nil); err != nil {
				log.Warnf("Failed to push: %v", err)
			}
		}
	}
	mime := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("Last-Modified", mod.Format(time.RFC1123))
	if r.Header.Get("If-None-Match") == sum {
		w.WriteHeader(http.StatusNotModified)
		w.Header().Set("ETag", sum)
		return
	}
	w.Header().Set("ETag", sum)
	w.Write(content)
}

func readFile(path string) ([]byte, string, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", time.Now(), err
	}
	defer f.Close()

	stat, err := os.Stat(path)
	if err != nil {
		return nil, "", time.Now(), err
	}

	cont, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, "", time.Now(), err
	}

	return cont, fmt.Sprintf("%x", sha512.Sum512(cont)), stat.ModTime(), nil
}
