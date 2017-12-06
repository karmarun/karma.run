// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package main

import (
	"api"
	_ "codec/binary"
	_ "codec/json"
	"crypto/tls"
	"db"
	"encoding/base64"
	"flag"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/acme/autocert"
	"kvm"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	httpPort           string
	httpsPort          string
	letsencryptDomains string
	letsencryptEmail   string
	httpsCertFile      string
	httpsKeyFile       string
	dataPath           string
	instanceSecret     string
)

func init() {

	flag.StringVar(&httpPort, "http-port", "80", "Required. Port to serve (insecure) HTTP clients on. Overriden by environment variable: PORT or HTTP_PORT.")
	flag.StringVar(&httpsPort, "https-port", "443", "Required. Port to serve HTTPS and HTTP/2 clients on. Overriden by environment variable: HTTPS_PORT.")

	flag.StringVar(&letsencryptDomains, "letsencrypt-domains", "", "Comma-separated list of HTTPS domains to automatically secure via LetsEncrypt. Overriden by environment variable: LETSENCRYPT_DOMAINS.")
	flag.StringVar(&letsencryptEmail, "letsencrypt-email", "", "Sets the contact email for LetsEncrypt. Required if --letsencrypt-domains is set. Overriden by environment variable: LETSENCRYPT_EMAIL.")

	flag.StringVar(&httpsCertFile, "https-cert-file", "", "Path to TLS certificate. Has no effect if LetsEncrypt config if set. Overriden by environment variable: HTTPS_CERT_FILE.")
	flag.StringVar(&httpsKeyFile, "https-key-file", "", "Path to TLS private key file. Has no effect if LetsEncrypt config if set. Overriden by environment variable: HTTPS_KEY_FILE.")

	flag.StringVar(&instanceSecret, "instance-secret", "", "Instance secret as base64-encoded string, used to initialize database. Overriden by INSTANCE_SECRET.")
	flag.StringVar(&dataPath, "data-file", "karma.data", "Path to data file. Overriden by DATA_FILE.")

}

func main() {

	flag.Parse()

	{ // environment overrides
		if s := os.Getenv(`PORT`); len(s) > 0 {
			httpPort = s
		}
		if s := os.Getenv(`HTTP_PORT`); len(s) > 0 {
			httpPort = s
		}
		if s := os.Getenv(`HTTPS_PORT`); len(s) > 0 {
			httpsPort = s
		}
		if s := os.Getenv(`LETSENCRYPT_DOMAINS`); len(s) > 0 {
			letsencryptDomains = s
		}
		if s := os.Getenv(`LETSENCRYPT_EMAIL`); len(s) > 0 {
			letsencryptEmail = s
		}
		if s := os.Getenv(`HTTPS_CERT_FILE`); len(s) > 0 {
			httpsCertFile = s
		}
		if s := os.Getenv(`HTTPS_KEY_FILE`); len(s) > 0 {
			httpsKeyFile = s
		}
		if s := os.Getenv(`INSTANCE_SECRET`); len(s) > 0 {
			instanceSecret = s
		}
		if s := os.Getenv(`DATA_FILE`); len(s) > 0 {
			dataPath = s
		}
	}

	{
		secret, e := base64.StdEncoding.DecodeString(instanceSecret)
		if e != nil {
			log.Fatalln("instance secret must be base64-encoded (see --help)")
		}
		if len(secret) < 512 {
			log.Fatalf("decoded instance secret must be longer than %d bytes\n", 512)
		}
	}

	{ // publish environment
		if e := os.Setenv(`PORT`, httpPort); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`HTTP_PORT`, httpPort); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`HTTPS_PORT`, httpsPort); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`LETSENCRYPT_DOMAINS`, letsencryptDomains); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`LETSENCRYPT_EMAIL`, letsencryptEmail); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`HTTPS_CERT_FILE`, httpsCertFile); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`HTTPS_KEY_FILE`, httpsKeyFile); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`INSTANCE_SECRET`, instanceSecret); e != nil {
			log.Fatalln(e)
		}
		if e := os.Setenv(`DATA_FILE`, dataPath); e != nil {
			log.Fatalln(e)
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
	log.Println("HTTP port:", httpPort)

	httpServer, httpsServer := (*http.Server)(nil), (*http.Server)(nil)

	httpServer = &http.Server{
		Addr:    ":" + httpPort,
		Handler: http.HandlerFunc(api.HttpHandler),
	}

	httpsRedirectionHandler := http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {
		u := rq.URL
		u.Scheme = "https"
		u.Host = rq.Host
		http.Redirect(rw, rq, u.String(), http.StatusMovedPermanently)
	})

	{ // LetsEncrypt support
		if (len(letsencryptDomains) > 0 && len(letsencryptEmail) == 0) || (len(letsencryptDomains) == 0 && len(letsencryptEmail) > 0) {
			log.Fatalln("--letsencrypt-email and --letsencrypt-domains must be set together.")
		}

		if len(letsencryptDomains) > 0 {
			domains := strings.Split(letsencryptDomains, ",")
			log.Println("HTTPS port:", httpsPort)
			log.Println("LetsEncrypt domains:", domains)
			log.Println("LetsEncrypt email:", letsencryptEmail)
			m := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				Cache:      (autocert.DirCache)(`dbs/autocert-cache`),
				HostPolicy: autocert.HostWhitelist(domains...),
				Email:      letsencryptEmail,
			}
			httpsServer = &http.Server{
				Addr:      ":" + httpsPort,
				Handler:   http.HandlerFunc(api.HttpHandler),
				TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
			}
			httpServer.Handler = httpsRedirectionHandler
			httpsCertFile, httpsKeyFile = ``, ``
		}
	}

	{ // Own TLS config support
		if (len(httpsCertFile) > 0 && len(httpsKeyFile) == 0) || (len(httpsCertFile) == 0 && len(httpsKeyFile) > 0) {
			log.Fatalln("--https-cert-file and --https-key-file must be set together.")
		}

		if len(httpsCertFile) > 0 {
			httpsServer = &http.Server{
				Addr:    ":" + httpsPort,
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
			if e := httpsServer.ListenAndServeTLS(httpsCertFile, httpsKeyFile); e != http.ErrServerClosed {
				log.Fatalln("HTTPS", e.Error())
			}
		}()
		log.Println("HTTPS server started")
	}

	select {}

}
