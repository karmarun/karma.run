// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.

package api

import (
	"compress/gzip"
	"encoding/base64"
	bolt "github.com/coreos/bbolt"
	"io"
	"karma.run/codec"
	"karma.run/config"
	"karma.run/db"
	"karma.run/kvm"
	"karma.run/kvm/err"
	"karma.run/kvm/val"
	"log"
	"net/http"
	"os"
	"sync"
)

func ExportHttpHandler(rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	userId := rq.Context().Value(ContextKeyUserId).(string)

	adminId, ke := adminUserIdFromDatabase(dtbs)
	if ke != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`unable to read database`, ke}.Value()))
		return
	}

	if string(adminId) != userId {
		log.Printf(`unauthorized database export request by user %s: %#v`, userId, *rq)
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	e := func() error {
		ow, e := gzip.NewWriterLevel(rw, gzip.BestCompression)
		if e != nil {
			return e
		}
		tx, e := dtbs.Begin(false)
		if e != nil {
			return e
		}
		defer tx.Rollback()
		rw.Header().Set(`Content-Type`, `application/octet-stream`)
		rw.Header().Set(`Content-Disposition`, `attachment; filename="`+config.DataFile+`"`)
		rw.Header().Set(`Content-Encoding`, `gzip`)
		_, e = tx.WriteTo(ow)
		if e != nil {
			return e
		}
		e = ow.Close()
		return e
	}()

	if e != nil {
		log.Println(e)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`export failed`, nil}.Value()))
	}
}

func ImportHttpHandler(rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	userId := rq.Context().Value(ContextKeyUserId).(string)

	adminId, ke := adminUserIdFromDatabase(dtbs)
	if ke != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`unable to read database`, ke}.Value()))
		return
	}

	if string(adminId) != userId {
		log.Printf(`unauthorized database export request by user %s: %#v`, userId, *rq)
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	e := db.WhileClosed(func() error {

		f, e := os.OpenFile(config.DataFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, db.Perm)
		if e != nil {
			return e
		}
		ir, e := gzip.NewReader(rq.Body)
		if e != nil {
			return e
		}
		_, e = io.Copy(f, ir)
		if e != nil {
			log.Panicln(e.Error())
		}
		_ = ir.Close()
		_ = f.Close()
		_ = rq.Body.Close()
		kvm.ClearCompilerCache()
		return nil
	})

	if e != nil {
		log.Println(e)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`import failed`, nil}.Value()))
	}

}

func RotateInstanceSecretHttpHandler(rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	secret := rq.Header.Get(SecretHeader)

	if string(secret) != config.InstanceSecret {
		log.Printf(`unauthorized attempt to rotate instance secret: %#v`, *rq)
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	newSecret := base64.StdEncoding.EncodeToString(RandIv(512))

	config.InstanceSecret = newSecret

	log.Println("instance secret rotated:", newSecret)

}

const (
	MinDatabaseSecretLength = 512
)

var resetLock = &sync.Mutex{}

func ResetHttpHandler(rw http.ResponseWriter, rq *http.Request) {

	resetLock.Lock()
	defer resetLock.Unlock() // no concurrent reset requests

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	userId := rq.Context().Value(ContextKeyUserId).(string)

	adminId, ke := adminUserIdFromDatabase(dtbs)
	if ke != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`unable to read database`, ke}.Value()))
		return
	}

	if string(adminId) != userId {
		log.Printf(`unauthorized database reset request by user %s: %#v`, userId, *rq)
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	e := dtbs.Update(func(tx *bolt.Tx) error {
		_ = tx.DeleteBucket([]byte(`root`))
		rb, e := tx.CreateBucket([]byte(`root`))
		if e != nil {
			return e
		}
		return (&kvm.VirtualMachine{RootBucket: rb}).InitDB()
	})

	if e != nil {
		if ke, ok := e.(err.Error); ok {
			writeError(rw, cdc, err.HumanReadableError{ke})
			return
		}
		log.Panicln(e)
	}

	msg := "instance reset successful"

	log.Println(msg)
	rw.Write(cdc.Encode(val.String(msg)))

}
