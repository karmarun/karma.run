package common

import (
	"codec/json"
	"definitions"
	"kvm/mdl"
	"kvm/val"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// internal use only
var internalMetaModel, _ = mdl.ModelFromValue("internal", definitions.NewMetaModelValue("internal").(val.Union), nil)

func QuickModel(data string) mdl.Model {
	v, e := json.Decode(json.JSON(data), internalMetaModel, nil)
	if e != nil {
		log.Panicf("%#v\n%s\n", e, data)
	}
	m, ke := mdl.ModelFromValue("internal", v.(val.Union), nil)
	if ke != nil {
		log.Panicf("%#v\n", ke.Error())
	}
	return m
}

// func SetCorsHeaders(rw http.ResponseWriter, rq *http.Request) {
// 	rw.Header().Set("Access-Control-Allow-Headers", "Expect, Accept, Content-Type, X-Ijt, X-Karma-Database, X-Karma-Codec, X-Karma-Signature, X-Karma-Secret")
// 	rw.Header().Set("Access-Control-Allow-Methods", "GET, POST")
// 	rw.Header().Set("Access-Control-Allow-Origin", "*")
// }

func SetJsonHeader(rw http.ResponseWriter, rq *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func LogRequest(rw http.ResponseWriter, rq *http.Request) {
	log.Println(rq.Method, rq.URL.String())
}

const alphabet = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`

func BytesToBaseAlphabet(bs []byte) string {
	cs := make([]byte, len(bs), len(bs))
	for i, b := range bs {
		cs[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(cs)
}

func IntToBaseAlphabet(i int) string {
	if i < 0 {
		i = -i
	}
	s := make([]byte, 0, 24)
	for {
		s = append(s, alphabet[i%len(alphabet)])
		i = i / len(alphabet)
		if i == 0 {
			break
		}
	}
	return string(s)
}

func Uint64ToBaseAlphabet(i uint64) string {
	s := make([]byte, 0, 24)
	l := uint64(len(alphabet))
	for {
		s = append(s, alphabet[i%l])
		i = i / l
		if i == 0 {
			break
		}
	}
	return string(s)
}

func RandomId() string {
	bs := make([]byte, 16, 16)
	for i, _ := range bs {
		bs[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(bs)
}

func HttpError(rw http.ResponseWriter, code int) {
	http.Error(rw, http.StatusText(code), code)
}

const excelAlphabet = "abcdefghijklmnopqrstuvwxyz"

func ExcelVariableName(n int) string {
	name := ""
	n++
	for n > 0 {
		n--
		name = string(excelAlphabet[n%len(excelAlphabet)]) + name
		n /= len(excelAlphabet)
	}
	return name
}
