// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"fmt"
	bolt "github.com/coreos/bbolt"
	"karma.run/codec"
	"karma.run/kvm"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
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
		RestApiPostHttpHandler(rw, rq)

	case http.MethodPut:
		RestApiPutHttpHandler(rw, rq)

	case http.MethodDelete:
		RestApiDeleteHttpHandler(rw, rq)

	default:
		cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
		writeError(rw, cdc, err.HumanReadableError{
			err.RequestError{
				Problem: fmt.Sprintf("invalid HTTP method requested: %s. supported are: GET, POST, PUT and DELETE.", rq.Method),
			},
		})
	}
}

func RestApiPutHttpHandler(rw http.ResponseWriter, rq *http.Request) {
	segments := pathSegments(rq.URL.Path)[1:] // drop "rest" prefix
	switch len(segments) {
	case 0: // PUT /
		http.NotFound(rw, rq)
		return

	case 1: // PUT /{resource}
		http.NotFound(rw, rq)
		return

	case 2: // PUT /{resource}/{id}
		RestApiPutResourceIdHttpHandler(segments[0], segments[1], rw, rq)
		return

	default:
		// error?
	}

}
func RestApiPostHttpHandler(rw http.ResponseWriter, rq *http.Request) {
	segments := pathSegments(rq.URL.Path)[1:] // drop "rest" prefix
	switch len(segments) {
	case 0: // POST /
		http.NotFound(rw, rq)
		return

	case 1: // POST /{resource}
		RestApiPostResourceHttpHandler(segments[0], rw, rq)
		return

	case 2: // POST /{resource}/{id}
		http.NotFound(rw, rq)
		return

	default:
		// error?
	}
}

func RestApiGetHttpHandler(rw http.ResponseWriter, rq *http.Request) {
	segments := pathSegments(rq.URL.Path)[1:] // drop "rest" prefix
	switch len(segments) {
	case 0: // GET /
		http.NotFound(rw, rq)
		return

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

func RestApiDeleteHttpHandler(rw http.ResponseWriter, rq *http.Request) {
	segments := pathSegments(rq.URL.Path)[1:] // drop "rest" prefix
	switch len(segments) {
	case 0: // DELETE /

	case 1: // DELETE /{resource}
		RestApiDeleteResourceHttpHandler(segments[0], rw, rq)

	case 2: // DELETE /{resource}/{id}
		RestApiDeleteResourceIdHttpHandler(segments[0], segments[1], rw, rq)
		return

	default:
		// error?
	}
}

// GET /{resource}
// query arguments:
// - length   int  amount of results
// - offset   int  amount to skip
// - metadata bool whether to metarialize
// TODO: sort
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

	rw.WriteHeader(http.StatusOK)
	rw.Write(cdc.Encode(list))

}

// PUT /{resource}/{id}
func RestApiPutResourceIdHttpHandler(resource, id string, rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	uid := rq.Context().Value(ContextKeyUserId).(string)

	outValue := (val.Value)(nil)
	payload := payloadFromRequest(rq)

	e := dtbs.Batch(func(tx *bolt.Tx) error {

		rb := tx.Bucket([]byte(`root`))
		if rb == nil {
			return err.InternalError{Problem: `database uninitialized`}
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

		modelRef, _, ke := vm.CompileAndExecuteExpression(modelExpr)
		if ke != nil {
			return ke
		}

		model, ke := vm.Model(modelRef.(val.Ref)[1])
		if ke != nil {
			return ke
		}

		value, ke := cdc.Decode(payload, model.Unwrap())
		if ke != nil {
			return ke
		}

		updateExpr := xpr.Update{
			Ref:   xpr.Literal{val.Ref{modelRef.(val.Ref)[1], id}},
			Value: xpr.Literal{value},
		}

		retVal, _, ke := vm.CompileAndExecuteExpression(updateExpr)
		if ke != nil {
			return ke
		}

		outValue = retVal
		return nil

	})

	if e != nil {
		ke, ok := e.(err.Error)
		if !ok {
			log.Println(e)
			ke = err.InternalError{Problem: `internal error`}
		}
		if _, ok := ke.(err.HumanReadableError); !ok {
			ke = err.HumanReadableError{ke}
		}
		writeError(rw, cdc, ke)
		return
	}

	rw.Write(cdc.Encode(outValue))

}

// POST /{resource}
func RestApiPostResourceHttpHandler(resource string, rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	uid := rq.Context().Value(ContextKeyUserId).(string)

	outValue := (val.Value)(nil)
	payload := payloadFromRequest(rq)

	e := dtbs.Batch(func(tx *bolt.Tx) error {

		rb := tx.Bucket([]byte(`root`))
		if rb == nil {
			return err.InternalError{Problem: `database uninitialized`}
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

		modelRef, _, ke := vm.CompileAndExecuteExpression(modelExpr)
		if ke != nil {
			return ke
		}

		model, ke := vm.Model(modelRef.(val.Ref)[1])
		if ke != nil {
			return ke
		}

		values, ke := cdc.Decode(payload, mdl.List{model.Unwrap()})
		if ke != nil {
			return ke
		}

		valueList := values.(val.List)

		funcMap := make(map[string]xpr.Function, len(valueList))
		for i, value := range valueList {
			k := strconv.Itoa(i)
			funcMap[k] = xpr.NewFunction([]string{"_"}, xpr.Literal{value})
		}

		createExpr := xpr.CreateMultiple{
			In:     modelExpr,
			Values: funcMap,
		}

		createdMap, _, ke := vm.CompileAndExecuteExpression(createExpr)
		if ke != nil {
			return ke
		}

		createdMap.(val.Struct).ForEach(func(k string, v val.Value) bool {
			i, _ := strconv.Atoi(k)
			valueList[i] = v
			return true
		})

		outValue = valueList
		return nil

	})

	if e != nil {
		ke, ok := e.(err.Error)
		if !ok {
			log.Println(e)
			ke = err.InternalError{Problem: `internal error`}
		}
		if _, ok := ke.(err.HumanReadableError); !ok {
			ke = err.HumanReadableError{ke}
		}
		writeError(rw, cdc, ke)
		return
	}

	rw.Write(cdc.Encode(outValue))

}

func maxInt64(a, b val.Int64) val.Int64 {
	if a > b {
		return a
	}
	return b
}

// DELETE /{resource}
func RestApiDeleteResourceHttpHandler(resource string, rw http.ResponseWriter, rq *http.Request) {
	RestApiDeleteResourceIdHttpHandler(`_model`, resource, rw, rq)
}

// DELETE /{resource}/{id}
func RestApiDeleteResourceIdHttpHandler(resource, id string, rw http.ResponseWriter, rq *http.Request) {

	cdc := rq.Context().Value(ContextKeyCodec).(codec.Interface)
	dtbs := rq.Context().Value(ContextKeyDatabase).(*bolt.DB)
	uid := rq.Context().Value(ContextKeyUserId).(string)

	outValue := (val.Value)(nil)

	e := dtbs.Batch(func(tx *bolt.Tx) error {

		rb := tx.Bucket([]byte(`root`))
		if rb == nil {
			return err.InternalError{Problem: `database uninitialized`}
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

		modelRef, _, ke := vm.CompileAndExecuteExpression(modelExpr)
		if ke != nil {
			return ke
		}

		targetRef := val.Ref{modelRef.(val.Ref)[1], id}

		oldVal, _, ke := vm.CompileAndExecuteExpression(xpr.Delete{
			xpr.Literal{targetRef},
		})

		outValue = oldVal

		return ke

	})

	if e != nil {
		ke, ok := e.(err.Error)
		if !ok {
			log.Println(e)
			ke = err.InternalError{Problem: `internal error`}
		}
		if _, ok := ke.(err.HumanReadableError); !ok {
			ke = err.HumanReadableError{ke}
		}
		writeError(rw, cdc, ke)
		return
	}

	rw.Write(cdc.Encode(outValue))

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
