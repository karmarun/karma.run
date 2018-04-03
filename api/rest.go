package api

import (
	"bytes"
	"fmt"
	"karma.run/codec"
	"karma.run/kvm/err"
	"net/http"
)

func RestApiHttpHandler(rw http.ResponseWriter, rq *http.Request) {
	switch rq.Method {
	case http.MethodGet:
		RestApiGetHttpHandler(rw, rq)

	case http.MethodPost:

	case http.MethodPut:

	case http.MethodDelete:

	default:
		cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
		writeError(rw, cdc, err.HumanReadableError{
			err.RequestError{fmt.Sprintf("invalid HTTP method requested: %s. supported are: GET, POST, PUT and DELETE.", rq.Method), nil},
		})
	}
}

func RestApiGetHttpHandler(rw http.ResponseWriter, rq *http.Request) {

	segments := pathSegments(rq.URL.Path)[1:] // drop "rest" prefix

	switch len(segments) {
	case 0: // GET /
		// swagger spec?

	case 1: // GET /{resource}
		RestApiGetResourceHttpHandler(segments[0], rw, rq)
		return

	case 2: // GET /{resource}/{id}
		RestApiGetResourceIdHttpHandler(segments[0], segments[1], rw, rq)
		return

	default:
		// error?
	}

}

// GET /{resource}
func RestApiGetResourceHttpHandler(resource string, rw http.ResponseWriter, rq *http.Request) {

}

// GET /{resource}/{id}
func RestApiGetResourceIdHttpHandler(resource, id string, rw http.ResponseWriter, rq *http.Request) {

}

// "/rest//foo//bar///" -> ["rest", "foo", "bar"]
func pathSegments(path string) []string {

	bs := []byte(path)

	temp := bs[:0]
	for slash := true; len(bs) > 0; bs = bs[1:] {
		if slash && bs[0] == '/' {
			continue
		}
		slash = (bs[0] == '/')
		temp = append(temp, bs[0])
	}
	bs = temp

	cut := len(bs) - 1
	for cut >= 0 && bs[cut] == '/' {
		cut--
	}
	bs = bs[:cut+1]

	bss := bytes.Split(bs, []byte{'/'})
	ss := make([]string, len(bss), len(bss))
	for i, bs := range bss {
		ss[i] = string(bs)
	}
	return ss
}
