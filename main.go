// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/acme/autocert"
	"karma.run/api"
	_ "karma.run/codec/binary"
	_ "karma.run/codec/json"
	"karma.run/config"
	"karma.run/db"
	"karma.run/kvm"
	"log"
	"net/http"
	"strings"
)

func main() {

	flag.Parse()

	{
		if len(config.InstanceSecret) == 0 {
			bs := make([]byte, 512, 512)
			if n, _ := rand.Read(bs); n < len(bs) {
				log.Fatalln("error generating instance secret: system entropy too low. see --help to pass one by hand.")
			}
			config.InstanceSecret = base64.StdEncoding.EncodeToString(bs)
			log.Println("no instance secret specified, using generated:", config.InstanceSecret)
		} else {
			secret, e := base64.StdEncoding.DecodeString(config.InstanceSecret)
			if e != nil {
				log.Fatalln("instance secret must be base64-encoded (see --help)")
			}
			if len(secret) < 512 {
				log.Fatalf("decoded instance secret must be longer than %d bytes\n", 512)
			}
		}
	}

	{ // init database if necessary

		db, e := db.Open()
		if e != nil {
			log.Fatalln(e)
		}
		rootBytes := []byte(`root`)
		e = db.Update(func(tx *bolt.Tx) error {
			if tx.Bucket(rootBytes) != nil {
				log.Println("data file already initialized")
				return nil
			}
			log.Println("initializing data file...")
			rb, e := tx.CreateBucket(rootBytes)
			if e != nil {
				return e
			}
			log.Println("initialized data file")
			return (&kvm.VirtualMachine{RootBucket: rb}).InitDB()
		})
		if e != nil {
			log.Fatalln(e)
		}
	}

	log.Println("starting karma.run...")
	log.Println("HTTP port:", config.HttpPort)

	httpServer, httpsServer := (*http.Server)(nil), (*http.Server)(nil)

	httpServer = &http.Server{
		Addr:    ":" + config.HttpPort,
		Handler: http.HandlerFunc(api.HttpHandler),
	}

	httpsRedirectionHandler := http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		u := rq.URL
		u.Scheme = "https"
		u.Host = rq.Host
		http.Redirect(rw, rq, u.String(), http.StatusMovedPermanently)
	})

	{ // LetsEncrypt support
		if (len(config.LetsencryptDomains) > 0 && len(config.LetsencryptEmail) == 0) || (len(config.LetsencryptDomains) == 0 && len(config.LetsencryptEmail) > 0) {
			log.Fatalln("--letsencrypt-email and --letsencrypt-domains must be set together.")
		}
		if (len(config.LetsencryptDomains) > 0 && len(config.LetsencryptCacheDir) == 0) || (len(config.LetsencryptDomains) == 0 && len(config.LetsencryptCacheDir) > 0) {
			log.Fatalln("--letsencrypt-cache-dir and --letsencrypt-domains must be set together.")
		}

		if len(config.LetsencryptDomains) > 0 {
			domains := strings.Split(config.LetsencryptDomains, ",")
			log.Println("HTTPS port:", config.HttpsPort)
			log.Println("LetsEncrypt domains:", domains)
			log.Println("LetsEncrypt email:", config.LetsencryptEmail)
			m := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				Cache:      (autocert.DirCache)(config.LetsencryptCacheDir),
				HostPolicy: autocert.HostWhitelist(domains...),
				Email:      config.LetsencryptEmail,
			}
			httpsServer = &http.Server{
				Addr:      ":" + config.HttpsPort,
				Handler:   http.HandlerFunc(api.HttpHandler),
				TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
			}
			httpServer.Handler = m.HTTPHandler(httpsRedirectionHandler)
			config.HttpsCertFile, config.HttpsKeyFile = ``, ``
		}
	}

	{ // Own TLS config support
		if (len(config.HttpsCertFile) > 0 && len(config.HttpsKeyFile) == 0) || (len(config.HttpsCertFile) == 0 && len(config.HttpsKeyFile) > 0) {
			log.Fatalln("--https-cert-file and --https-key-file must be set together.")
		}

		if len(config.HttpsCertFile) > 0 {
			httpsServer = &http.Server{
				Addr:    ":" + config.HttpsPort,
				Handler: http.HandlerFunc(api.HttpHandler),
			}
			httpServer.Handler = httpsRedirectionHandler
		}
	}

	go func() {
		if e := httpServer.ListenAndServe(); e != http.ErrServerClosed {
			log.Fatalln("HTTP", e.Error())
		}
	}()
	log.Println("HTTP server started")

	if httpsServer != nil {
		go func() {
			if e := httpsServer.ListenAndServeTLS(config.HttpsCertFile, config.HttpsKeyFile); e != http.ErrServerClosed {
				log.Fatalln("HTTPS", e.Error())
			}
		}()
		log.Println("HTTPS server started")
	}

	select {}

}
