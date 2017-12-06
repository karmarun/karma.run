// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package api

import (
	"codec"
	"compress/gzip"
	"db"
	"encoding/base64"
	"github.com/boltdb/bolt"
	"io"
	"kvm"
	"kvm/err"
	"kvm/val"
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

	e := db.WhileClosed(func() {

		f, e := os.Open(os.Getenv("DATA_FILE"))
		if e != nil {
			rw.WriteHeader(http.StatusNotFound)
			// ?
			return
		}
		ow := gzip.NewWriter(rw)
		_, e = io.Copy(ow, f)
		if e != nil {
			log.Panicln(e.Error())
		}
		ow.Close()
		f.Close()

	})

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

	e := db.WhileClosed(func() {

		f, e := os.OpenFile(os.Getenv("DATA_FILE"), os.O_RDWR|os.O_TRUNC|os.O_CREATE, db.Perm)
		if e != nil {
			rw.WriteHeader(http.StatusNotFound)
			// ?
			return
		}
		ir, e := gzip.NewReader(rq.Body)
		if e != nil {
			rw.WriteHeader(http.StatusBadRequest)
			// ?
			return
		}
		_, e = io.Copy(f, ir)
		if e != nil {
			log.Panicln(e.Error())
		}
		ir.Close()
		f.Close()
		rq.Body.Close()
		kvm.ClearCompilerCache()
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

	if string(secret) != os.Getenv("INSTANCE_SECRET") {
		log.Printf(`unauthorized attempt to rotate instance secret: %#v`, *rq)
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	newSecret := base64.StdEncoding.EncodeToString(RandIv(512))
	if e := os.Setenv("INSTANCE_SECRET", newSecret); e != nil {
		log.Panicln(e)
	}

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
		log.Panicln(e)
	}

	msg := "instance reset successful"

	log.Println(msg)
	rw.Write(cdc.Encode(val.String(msg)))

}
