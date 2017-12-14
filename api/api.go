// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/boltdb/bolt"
	"io"
	"karma.run/codec"
	"karma.run/db"
	"karma.run/definitions"
	"karma.run/kvm"
	"karma.run/kvm/err"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"sync"

	_ "net/http/pprof"
)

var version = `1.4.0`

type Payload []byte

func (p Payload) Close() {
	copy(p, ZeroPayload)
	PayloadPool.Put(p[:MaxPayloadBytes])
}

const MaxPayloadBytes = 1 * 1024 * 1024 // 1MB

var (
	PayloadPool = &sync.Pool{
		New: func() interface{} {
			return make(Payload, MaxPayloadBytes, MaxPayloadBytes)
		},
	}
	ZeroPayload = make(Payload, MaxPayloadBytes, MaxPayloadBytes)
)

type TxType byte

const (
	TxTypeNone TxType = iota
	TxTypeRead
	TxTypeWrite
)

const (
	DocsPrefix = `docs`
	AuthPrefix = `auth`

	ExportPrefix               = `admin/export`
	ImportPrefix               = `admin/import`
	ResetPrefix                = `admin/reset`
	RotateInstanceSecretPrefix = `admin/rotate_instance_secret`
)

const (
	SignatureHeader = `X-Karma-Signature`
	CodecHeader     = `X-Karma-Codec`
	SecretHeader    = `X-Karma-Secret`
)

func HttpHandler(rw http.ResponseWriter, rq *http.Request) {

	if len(os.Getenv("PPROF")) > 0 && strings.HasPrefix(rq.URL.Path, "/debug/pprof") {
		http.DefaultServeMux.ServeHTTP(rw, rq)
		return
	}

	// CORS headers for browsers
	rw.Header().Set("Access-Control-Allow-Headers", rq.Header.Get("Access-Control-Request-Headers"))
	rw.Header().Set("Access-Control-Allow-Methods", rq.Header.Get("Access-Control-Request-Method"))
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	if rq.Method == http.MethodOptions {
		return // CORS pre-flight
	}

	path := strings.Trim(path.Clean(rq.URL.Path), "/")

	if rq.Method == http.MethodGet && path == "" { // k8s health checks
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`karma.run ` + version))
		return
	}

	if len(os.Getenv("DOCS_PATH")) > 0 && len(path) >= len(DocsPrefix) && path[:len(DocsPrefix)] == DocsPrefix {
		path = path[len(DocsPrefix):]
		path = os.Getenv("DOCS_PATH") + path
		http.ServeFile(rw, rq, path)
		return
	}

	cdc := codec.Get(rq.Header.Get(CodecHeader))
	if cdc == nil {
		msg := fmt.Sprintf(`invalid codec requested (%s header). available codecs: %s`, CodecHeader, strings.Join(codec.Available(), ", "))
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(msg))
		return
	}

	rq = rq.WithContext(context.WithValue(rq.Context(), ContextKeyCodec, cdc))

	dtbs, e := db.Open()
	if e != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{Problem: "failed opening database"}.Value()))
		log.Println(e)
		return
	}

	rq = rq.WithContext(context.WithValue(rq.Context(), ContextKeyDatabase, dtbs))

	if len(path) >= len(RotateInstanceSecretPrefix) && path[:len(RotateInstanceSecretPrefix)] == RotateInstanceSecretPrefix {
		RotateInstanceSecretHttpHandler(rw, rq)
		return
	}

	if len(path) >= len(AuthPrefix) && path[:len(AuthPrefix)] == AuthPrefix {
		AuthHttpHandler(rw, rq)
		return
	}

	sig, e := signatureFromRequest(rq)
	if e != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.RequestError{`failed to decode user signature`, nil}.Value()))
		return
	}

	userId, ke := tenref(sig, secretFromEnvironment())
	if ke != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{ke}.Value()))
		return
	}

	rq = rq.WithContext(context.WithValue(rq.Context(), ContextKeyUserId, string(userId)))

	if len(path) >= len(ResetPrefix) && path[:len(ResetPrefix)] == ResetPrefix {
		ResetHttpHandler(rw, rq)
		return
	}

	if len(path) >= len(ExportPrefix) && path[:len(ExportPrefix)] == ExportPrefix {
		ExportHttpHandler(rw, rq)
		return
	}

	if len(path) >= len(ImportPrefix) && path[:len(ImportPrefix)] == ImportPrefix {
		ImportHttpHandler(rw, rq)
		return
	}

	if len(path) > 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	payload := payloadFromRequest(rq)
	defer payload.Close()

	expr, ke := cdc.Decode(payload, xpr.LanguageModel)
	if ke != nil {
		writeError(rw, cdc, err.HumanReadableError{ke})
		return
	}

	txt := TxTypeRead
	expr.Transform(func(v val.Value) val.Value {
		if u, ok := v.(val.Union); ok {
			switch u.Case {
			case "create",
				"delete",
				"update",
				"createMultiple":
				txt = TxTypeWrite
			}
		}
		return v
	})

	tx, e := dtbs.Begin(txt == TxTypeWrite)
	if e != nil {
		panic(e)
	}

	defer tx.Rollback()

	defer func() {
		if v := recover(); v != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			switch e := v.(type) {
			case err.Error:
				writeError(rw, cdc, err.HumanReadableError{e})
			case error:
				log.Printf("%#v\n", expr)
				log.Println(e.Error())
			default:
				log.Printf("%#v\n", expr)
				log.Printf("%#v\n", v)
			}
			debug.PrintStack()
		}
	}()

	bk := tx.Bucket([]byte(`root`))
	if bk == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`database uninitialized`, nil}.Value()))
		return
	}

	vm := &kvm.VirtualMachine{RootBucket: bk, Codec: cdc, UserID: string(userId)}

	if txt == TxTypeWrite {
		if e := vm.UpdateModels(); e != nil {
			log.Panicln(e)
		}
	}

	res, _, ke := vm.ParseCompileAndExecute(expr, nil, nil)
	if ke != nil {
		writeError(rw, cdc, err.HumanReadableError{ke})
		return
	}

	rw.Write(cdc.Encode(res))

	if tx.Writable() {
		if e := tx.Commit(); e != nil {
			log.Panicln(e)
		}
	}

}

func writeError(rw http.ResponseWriter, cdc codec.Interface, e err.Error) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write(cdc.Encode(e.Value()))
	return
}

func stringsContain(ss []string, s string) bool {
	for _, t := range ss {
		if t == s {
			return true
		}
	}
	return false
}

func payloadFromRequest(rq *http.Request) Payload {
	defer rq.Body.Close()
	return payloadFromReader(rq.Body)
}

func payloadFromReader(r io.Reader) Payload {
	payload := PayloadPool.Get().(Payload)
	readLength := 0
	for readLength < MaxPayloadBytes {
		n, e := r.Read(payload[readLength:])
		readLength += n
		if e == io.EOF {
			break // we're done
		}
	}
	return payload[:readLength]
}

func signatureFromRequest(rq *http.Request) ([]byte, error) {
	return base64.StdEncoding.DecodeString(rq.Header.Get(SignatureHeader))
}

func secretFromEnvironment() []byte {
	bs, _ := base64.StdEncoding.DecodeString(os.Getenv("INSTANCE_SECRET"))
	return bs
}

func adminUserIdFromDatabase(db *bolt.DB) ([]byte, err.Error) {
	tx, e := db.Begin(false)
	if e != nil {
		return nil, err.InternalError{`failed opening database transaction`, nil}
	}
	defer tx.Rollback()
	bk := tx.Bucket([]byte(`root`))
	if bk == nil {
		return nil, err.InternalError{`database uninitialized`, nil}
	}
	return bk.Get(definitions.RootUserBytes), nil
}
