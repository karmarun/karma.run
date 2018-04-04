package api

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"karma.run/codec"
	"karma.run/kvm"
	"karma.run/kvm/err"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"log"
	"net/http"
	"strconv"
	"strings"
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
			err.RequestError{
				Problem: fmt.Sprintf("invalid HTTP method requested: %s. supported are: GET, POST, PUT and DELETE.", rq.Method),
			},
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
// query arguments:
// - length   int  amount of results
// - offset   int  amount to skip
// - metadata bool whether to materialize
func RestApiGetResourceHttpHandler(resource string, rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	uid := rq.Context().Value(ContextKeyUserId).(string)

	tx, e := dtbs.Begin(false)
	if e != nil {
		log.Panicln(e)
	}
	defer tx.Rollback()

	rb := tx.Bucket([]byte(`root`))
	if rb == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{Problem: `database uninitialized`}.Value()))
		return
	}

	vm := &kvm.VirtualMachine{RootBucket: rb, UserID: uid}

	resourceLit := xpr.Literal{val.String(resource)}

	isTag, _, ke := vm.CompileAndExecuteExpression(xpr.TagExists{resourceLit})
	if ke != nil {
		log.Panicln(ke)
	}

	modelExpr := xpr.Expression(nil)

	if isTag.(val.Bool) {
		modelExpr = xpr.Tag{resourceLit}
	} else {
		modelExpr = xpr.Model{resourceLit}
	}

	listExpr := xpr.Expression(xpr.All{modelExpr})
	totalExpr := xpr.Length{listExpr}

	offset, length := val.Int64(0), val.Int64(100) // defaults

	if p, ok := rq.URL.Query()["offset"]; ok && len(p) > 0 {
		o, e := strconv.ParseInt(p[0], 10, 64)
		if e != nil || o < 0 {
			writeError(rw, cdc, err.HumanReadableError{err.RequestError{
				Problem: fmt.Sprintf(`offset parameter must be a positive integer, have: %s`, p[0]),
			}})
			return
		}
		offset = val.Int64(o)
	}

	if p, ok := rq.URL.Query()["length"]; ok && len(p) > 0 {
		l, e := strconv.ParseInt(p[0], 10, 64)
		if e != nil || l < 0 || l > 1000 {
			writeError(rw, cdc, err.HumanReadableError{err.RequestError{
				Problem: fmt.Sprintf(`length parameter must be a positive integer less than 1001, have: %s`, p[0]),
			}})
			return
		}
		length = val.Int64(l)
	}

	listExpr = xpr.Slice{
		Value:  listExpr,
		Offset: xpr.Literal{offset},
		Length: xpr.Literal{length},
	}

	if _, ok := rq.URL.Query()["metadata"]; ok {
		listExpr = xpr.MapList{
			Value:   listExpr,
			Mapping: xpr.NewFunction([]string{"index", "value"}, xpr.Metarialize{xpr.Scope("value")}),
		}
	}

	value, _, ke := vm.CompileAndExecuteExpression(xpr.NewTuple{listExpr, totalExpr})
	if ke != nil {
		writeError(rw, cdc, err.HumanReadableError{ke})
		return
	}

	list, total := value.(val.Tuple)[0].(val.List), value.(val.Tuple)[1].(val.Int64)

	linkHeader := make([]string, 0, 4)
	if (offset + length) < total {

		// set last link header
		lastOffset := val.Int64(0)
		for lastOffset < total-length {
			lastOffset += length
		}

		// set next link header
		nextOffset := offset + length

		query := rq.URL.Query()

		query.Set("offset", strconv.FormatInt(int64(lastOffset), 10))
		rq.URL.RawQuery = query.Encode()
		linkHeader = append(linkHeader, fmt.Sprintf(`<%s>; rel="last"`, rq.URL.String()))

		query.Set("offset", strconv.FormatInt(int64(nextOffset), 10))
		rq.URL.RawQuery = query.Encode()
		linkHeader = append(linkHeader, fmt.Sprintf(`<%s>; rel="next"`, rq.URL.String()))
	}
	if offset > 0 {

		// set prev link header
		prevOffset := maxInt64(0, offset-length)

		// set first link header
		firstOffset := 0

		query := rq.URL.Query()

		query.Set("offset", strconv.FormatInt(int64(prevOffset), 10))
		rq.URL.RawQuery = query.Encode()
		linkHeader = append(linkHeader, fmt.Sprintf(`<%s>; rel="prev"`, rq.URL.String()))

		query.Set("offset", strconv.FormatInt(int64(firstOffset), 10))
		rq.URL.RawQuery = query.Encode()
		linkHeader = append(linkHeader, fmt.Sprintf(`<%s>; rel="first"`, rq.URL.String()))
	}

	if len(linkHeader) > 0 {
		rw.Header().Set(`Link`, strings.Join(linkHeader, `, `))
	}

	// Link: <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=15>; rel="next",
	//   <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=34>; rel="last",
	//   <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=1>; rel="first",
	//   <https://api.github.com/search/code?q=addClass+user%3Amozilla&page=13>; rel="prev"

	rw.Write(cdc.Encode(list))

}

func maxInt64(a, b val.Int64) val.Int64 {
	if a > b {
		return a
	}
	return b
}

// GET /{resource}/{id}
func RestApiGetResourceIdHttpHandler(resource, id string, rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	uid := rq.Context().Value(ContextKeyUserId).(string)

	tx, e := dtbs.Begin(false)
	if e != nil {
		log.Panicln(e)
	}
	defer tx.Rollback()

	rb := tx.Bucket([]byte(`root`))
	if rb == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(cdc.Encode(err.InternalError{Problem: `database uninitialized`}.Value()))
		return
	}

	vm := &kvm.VirtualMachine{RootBucket: rb, UserID: uid}

	resRef, _, ke := vm.CompileAndExecuteExpression(xpr.Tag{xpr.Literal{val.String(resource)}})
	if ke != nil {
		resRef = val.Ref{vm.MetaModelId(), resource}
	}

	idRef, _, ke := vm.CompileAndExecuteExpression(xpr.Tag{xpr.Literal{val.String(id)}})
	if ke != nil {
		idRef = val.Ref{vm.MetaModelId(), id}
	}

	valExpr := xpr.Expression(xpr.Get{
		xpr.NewRef{
			Model: xpr.Literal{val.String(resRef.(val.Ref)[1])},
			Id:    xpr.Literal{val.String(idRef.(val.Ref)[1])},
		},
	})

	if _, ok := rq.URL.Query()["metadata"]; ok {
		valExpr = xpr.Metarialize{valExpr}
	}

	value, _, ke := vm.CompileAndExecuteExpression(valExpr)
	if ke != nil {
		writeError(rw, cdc, err.HumanReadableError{ke})
		return
	}

	rw.Write(cdc.Encode(value))
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
