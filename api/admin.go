// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.

package api

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	bolt "github.com/coreos/bbolt"
	"io"
	"io/ioutil"
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
		zw := zip.NewWriter(rw)
		fw, e := zw.Create(config.DataFile)
		if e != nil {
			return e
		}
		tx, e := dtbs.Begin(false)
		if e != nil {
			return e
		}
		defer tx.Rollback()
		rw.Header().Set(`Content-Type`, `application/zip`)
		rw.Header().Set(`Content-Disposition`, `attachment; filename="`+config.DataFile+`.zip"`)
		_, e = tx.WriteTo(fw)
		if e != nil {
			return e
		}
		return zw.Close()
	}()

	if e != nil {
		log.Println(e)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`export failed`, nil}.Value()))
	}
}

const maxImportSize = 1024 * 1024 * 1024 // in bytes

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

		temp, e := ioutil.TempFile("", "karma_import_*")
		if e != nil {
			// todo
		}

		defer os.Remove(temp.Name())

		l, e := io.Copy(temp, io.LimitReader(rq.Body, maxImportSize))
		if e != nil {
			return e
		}

		rq.Body.Close()

		if l == maxImportSize {
			return fmt.Errorf(`import too big. max size in bytes: %d`, maxImportSize)
		}

		zr, e := zip.NewReader(temp, l)
		if e != nil {
			return e
		}

		if len(zr.File) == 0 {
			return fmt.Errorf(`empty zip file`)
		}

		if len(zr.File) > 1 {
			return fmt.Errorf(`zip file contains %d files, expected 1`, len(zr.File))
		}

		fr, e := zr.File[0].Open()
		if e != nil {
			return e
		}

		defer fr.Close()

		f, e := os.OpenFile(config.DataFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, db.Perm)
		if e != nil {
			return e
		}
		if _, e = io.Copy(f, fr); e != nil {
			return e
		}
		if e := f.Close(); e != nil {
			return e
		}

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
