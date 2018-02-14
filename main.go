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
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/http2"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const version = "1.1.0"

var (
	devMode = flag.Bool("dev", false, "Puts the server in developer mode, will bind to :34265 and will not autocert")
	domains = flag.String("domain", "angrymills.net,happymills.net,sexymills.com", "A comma-seperaated list of domains to get a certificate for.")
	listen  = flag.String("listen", ":https", "The address to listen on")
	client  = &http.Client{}
	level   = 0
	m       autocert.Manager
)

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
	cLog := console.New(true)
	cLog.SetTimestampFormat(time.RFC3339)
	log.AddHandler(cLog, log.AllLevels...)

	log.Info("Starting The Josh Mills Anger Advisory System")

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/api/levels", levelHandler)
	mux.HandleFunc("/api/setlevel", setLevelHandler)
	mux.HandleFunc("/api/inclevel", increaseLevelHandler)
	mux.HandleFunc("/api/declevel", decreaseLevelHandler)
	mux.HandleFunc("/api/currentlevel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("%d", level)))
	})

	mux.HandleFunc("/api/tokens/list", listTokenHandler)
	mux.HandleFunc("/socket", webSocketHandler)

	printTokens()

	if *devMode {
		srv := &http.Server{
			Addr:    ":34265",
			Handler: mux,
		}

		log.Info("Listening on :34265")
		srv.ListenAndServe()
	}

		domainList := strings.Split(*domains, ",")
		for i, d := range domainList {
			domainList[i] = strings.TrimSpace(d)
		}

		m = autocert.Manager{
			Cache:      autocert.DirCache("certs"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domainList...),
		}

		tlsConf := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			GetCertificate:           m.GetCertificate,

			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},

			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		rootSrv := &http.Server{
			Addr:      *listen,
			Handler:   mux,
			TLSConfig: tlsConf,

			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		go http.ListenAndServe(":http", m.HTTPHandler(nil))

		log.Infof("Listening on %s", *listen)

		http2.ConfigureServer(rootSrv, &http2.Server{})
		log.Fatal(rootSrv.ListenAndServeTLS("", ""))
	}
}

func httpRedirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if _, err := os.Stat("./client" + path); err == nil {
		serveFile(w, r, "./client"+path)
	} else {
		serveFile(w, r, "./client/index.html")
	}
}
