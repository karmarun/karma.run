// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	bolt "github.com/coreos/bbolt"
	"golang.org/x/crypto/bcrypt"
	"karma.run/codec"
	"karma.run/config"
	"karma.run/definitions"
	"karma.run/kvm"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"log"
	"net/http"
	"sync"
	"time"
)

var AuthRequestModel = mdl.StructFromMap(map[string]mdl.Model{
	"username": mdl.String{},
	"password": mdl.String{},
})

const (
	ivLength    = 32
	tokenExpiry = time.Minute * 15
)

var loginLock = &sync.Mutex{}

func AuthHttpHandler(rw http.ResponseWriter, rq *http.Request) {

	loginLock.Lock()
	defer loginLock.Unlock() // no concurrent login attempts to brute-force passwords

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)

	hmacKey := []byte(config.InstanceSecret)

	if sig, e := signatureFromRequest(rq); len(sig) > 0 && e == nil { // user provided signature, try to renew it
		userId, e := tenref(sig, hmacKey)
		if e != nil {
			rw.WriteHeader(http.StatusForbidden)
			rw.Write(cdc.Encode(err.RequestError{`failed to decode user signature`, nil}.Value()))
			return
		}
		rw.Write(cdc.Encode(val.String(base64.RawURLEncoding.EncodeToString(fernet(userId, hmacKey)))))
		return
	}

	payload := payloadFromRequest(rq)

	atv, ke := cdc.Decode([]byte(payload), AuthRequestModel)
	if ke != nil {
		writeError(rw, cdc, err.HumanReadableError{ke})
		return
	}

	atr := atv.(val.Struct)
	username, password := string(atr.Field("username").(val.String)), string(atr.Field("password").(val.String))

	tx, e := dtbs.Begin(false)
	if e != nil {
		log.Panicln(e)
	}
	defer tx.Rollback()

	rb := tx.Bucket([]byte(`root`))
	if rb == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{`database uninitialized`, nil}.Value()))
		return
	}

	if username == "admin" && password == config.InstanceSecret {
		buf := fernet(rb.Get(definitions.RootUserBytes), hmacKey)
		rw.Write(cdc.Encode(val.String(base64.RawURLEncoding.EncodeToString(buf))))
		return
	}

	findUser := xpr.NewFunction(nil, xpr.Metarialize{
		xpr.First{
			xpr.FilterList{
				Value: xpr.All{
					xpr.Tag{
						xpr.Literal{val.String("_user")},
					},
				},
				Filter: xpr.NewFunction([]string{"i", "user"}, xpr.Equal{
					xpr.Literal{val.String(username)},
					xpr.Field{"username", xpr.Scope("user")},
				}),
			},
		},
	})

	vm := &kvm.VirtualMachine{RootBucket: rb}

	typedFindUser, e := vm.TypeFunction(findUser, nil, nil)
	if e != nil {
		panic(e)
	}

	mv, ke := vm.Execute(vm.CompileFunction(typedFindUser), nil)
	if ke != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	us := mv.(val.Struct).Field("value").(val.Struct)

	if e := bcrypt.CompareHashAndPassword([]byte(us.Field("password").(val.String)), []byte(password)); e != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write(cdc.Encode(err.PermissionDeniedError{}.Value()))
		return
	}

	buf := fernet([]byte(mv.(val.Struct).Field("id").(val.Ref)[1]), hmacKey)
	rw.Write(cdc.Encode(val.String(base64.RawURLEncoding.EncodeToString(buf))))
}

// fernet^-1
func tenref(sig, hmacKey []byte) ([]byte, err.Error) {

	if len(sig) != 88 {
		return nil, err.InputParsingError{`invalid token length`, sig}
	}

	ts := sig[:8]
	iv := sig[8 : 8+ivLength]
	id := sig[8+ivLength : len(sig)-32]
	sg := sig[len(sig)-32:]

	{
		te := time.Unix(int64(binary.LittleEndian.Uint64(ts)), 0)
		if te.Before(time.Now().Add(-1 * tokenExpiry)) {
			return nil, err.InputParsingError{`token expired`, sig}
		}
	}

	{
		hash := hmac.New(sha512.New512_256, hmacKey)
		hash.Write(ts)
		hash.Write(iv)
		hash.Write(id)
		if !hmac.Equal(hash.Sum(nil), sg) {
			return nil, err.InputParsingError{`invalid token`, sig}
		}
	}

	{ // decrypt id
		block, _ := aes.NewCipher(iv) // error never happens as long as ivLength is valid
		for i := 0; i < len(id); i += aes.BlockSize {
			block.Decrypt(id[i:], id[i:])
		}
	}

	return bytes.TrimRight(id, string([]byte{0})), nil

}

// INVARIANT: id must be exactly 16 bytes long
func fernet(id, hmacKey []byte) []byte {

	if len(id) != 16 {
		panic(fmt.Sprintf("fernet: len(id) != 16: %s", id))
	}

	ts := make([]byte, 8, 8)
	iv := RandIv(ivLength) // AES-256
	sg := ([]byte)(nil)

	binary.LittleEndian.PutUint64(ts, uint64(time.Now().Unix()))

	{ // encrypt id
		block, _ := aes.NewCipher(iv) // ignore error about IV length
		src, dst := id, make([]byte, len(id), len(id))
		for i := 0; i < len(src); i += block.BlockSize() {
			block.Encrypt(dst[i:], src[i:])
		}
		id = dst
	}

	{ // generate signature
		hash := hmac.New(sha512.New512_256, hmacKey)
		hash.Write(ts)
		hash.Write(iv)
		hash.Write(id)
		sg = hash.Sum(nil)
	}

	buf := make([]byte, 0, len(ts)+len(iv)+len(id)+len(sg))
	buf = append(buf, ts...)
	buf = append(buf, iv...)
	buf = append(buf, id...)
	buf = append(buf, sg...)

	return buf

}

func ceilToMultOf(n, m int) int {
	if n%m == 0 {
		return n
	}
	return n + (m - (n % m))
}

func RandIv(ln int) []byte {
	rd, iv := 0, make([]byte, ln, ln)
	for rd < len(iv) {
		n, e := rand.Read(iv[rd:])
		if e != nil {
			time.Sleep(time.Millisecond) // allow some entropy gathering
		}
		rd += n
	}
	return iv
}
