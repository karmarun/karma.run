// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package config

import (
	"flag"
	"os"
)

var (
	HttpPort            string = "80"  // explicit default
	HttpsPort           string = "443" // explicit default
	LetsencryptDomains  string
	LetsencryptEmail    string
	LetsencryptCacheDir string
	HttpsCertFile       string
	HttpsKeyFile        string
	InstanceSecret      string
	DataFile            string
)

func init() {
	flag.StringVar(
		&HttpPort,
		"http-port",
		getenv("KARMA_HTTP_PORT", HttpPort),
		"Required. Port to serve (insecure) HTTP clients on.",
	)
	flag.StringVar(
		&HttpsPort,
		"https-port",
		getenv("KARMA_HTTPS_PORT", HttpsPort),
		"Required. Port to serve HTTPS and HTTP/2 clients on.",
	)
	flag.StringVar(
		&LetsencryptDomains,
		"letsencrypt-domains",
		getenv("KARMA_LETSENCRYPT_DOMAINS", LetsencryptDomains),
		"Comma-separated list of HTTPS domains to automatically secure via LetsEncrypt.",
	)
	flag.StringVar(
		&LetsencryptEmail,
		"letsencrypt-email",
		getenv("KARMA_LETSENCRYPT_EMAIL", LetsencryptEmail),
		"Sets the contact email for LetsEncrypt. Required if --letsencrypt-domains is set.",
	)
	flag.StringVar(
		&LetsencryptCacheDir,
		"letsencrypt-cache-dir",
		getenv("KARMA_LETSENCRYPT_CACHE_DIR", LetsencryptCacheDir),
		"Sets the LetsEncrypt file cache location. Required if --letsencrypt-domains is set.",
	)
	flag.StringVar(
		&HttpsCertFile,
		"https-cert-file",
		getenv("KARMA_HTTPS_CERT_FILE", HttpsCertFile),
		"Path to TLS certificate. Has no effect if LetsEncrypt config if set.",
	)
	flag.StringVar(
		&HttpsKeyFile,
		"https-key-file",
		getenv("KARMA_HTTPS_KEY_FILE", HttpsKeyFile),
		"Path to TLS private key file. Has no effect if LetsEncrypt config if set.",
	)
	flag.StringVar(
		&InstanceSecret,
		"instance-secret",
		getenv("KARMA_INSTANCE_SECRET", InstanceSecret),
		"Instance secret as base64-encoded string, used to initialize database.",
	)
	flag.StringVar(
		&DataFile,
		"data-file",
		getenv("KARMA_DATA_FILE", DataFile),
		"Path to data file.",
	)
}

func getenv(key string, deflt string) string {
	v := os.Getenv(key)
	if v == "" {
		return deflt
	}
	return v
}
