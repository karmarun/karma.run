// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
//
// Package codec/json is DEPRECATED and should not be used. It exists for historic reasons and will
// be removed in the near future.
package json

import (
	"bytes"
	ej "encoding/json"
	"fmt"
	"karma.run/codec"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"log"
	"sort"
	"strconv"
	"time"
)

func init() {
	codec.Register("json", func() codec.Interface { return JsonCodec{} })
}

type JsonCodec struct{}

func (dec JsonCodec) Decode(json []byte, model mdl.Model) (val.Value, err.Error) {
	return Decode(json, model, nil)
}

func (dec JsonCodec) Encode(v val.Value) []byte {
	return Encode(v)
}

type JSON []byte

func (j JSON) MarshalJSON() ([]byte, error) {
	return []byte(j), nil
}

func (j *JSON) UnmarshalJSON(json []byte) error {
	(*j) = append((*j)[:0], json...)
	return nil
}

func (j JSON) String() string {
	return string(j)
}

func Encode(value val.Value) JSON {
	return encode(value, make(JSON, 0, 1024*4))
}

func encode(value val.Value, cache JSON) JSON {
	if value == nil {
		log.Panicln("json/codec.Encode: value == nil")
	}
	if value == val.Null {
		return JSON(`null`)
	}
	switch v := value.(type) {
	case val.Meta:
		return encode(v.Value, cache)

	case val.Tuple:
		bs := cache
		bs = append(bs, '[')
		for i, w := range v {
			if i > 0 {
				bs = append(bs, ',')
			}
			bs = encode(w, bs)
		}
		bs = append(bs, ']')
		return bs

	case val.Set:
		bs := cache
		bs = append(bs, '[')
		i := 0
		for _, w := range v {
			if i > 0 {
				bs = append(bs, ',')
			}
			bs = encode(w, bs)
			i++
		}
		bs = append(bs, ']')
		return bs

	case val.List:
		bs := cache
		bs = append(bs, '[')
		for i, w := range v {
			if i > 0 {
				bs = append(bs, ',')
			}
			bs = encode(w, bs)
		}
		bs = append(bs, ']')
		return bs

	case val.Union:
		bs := cache
		bs = append(bs, '[')
		cs, _ := ej.Marshal(v.Case)
		bs = append(bs, cs...)
		bs = append(bs, ',')
		bs = encode(v.Value, bs)
		bs = append(bs, ']')
		return bs

	case val.Struct:
		bs := cache
		bs = append(bs, '{')
		first := true
		v.ForEach(func(k string, v val.Value) bool {
			if !first {
				bs = append(bs, ',')
			}
			cs, _ := ej.Marshal(k)
			bs = append(bs, cs...)
			bs = append(bs, ':')
			bs = encode(v, bs)
			first = false
			return true
		})
		bs = append(bs, '}')
		return bs

	case val.Map:
		bs := cache
		bs = append(bs, '{')
		first := true
		v.ForEach(func(k string, v val.Value) bool {
			if !first {
				bs = append(bs, ',')
			}
			cs, _ := ej.Marshal(k)
			bs = append(bs, cs...)
			bs = append(bs, ':')
			bs = encode(v, bs)
			first = false
			return true
		})
		bs = append(bs, '}')
		return bs

	case val.String:
		bs, _ := ej.Marshal(v)
		return append(cache, bs...)

	case val.Ref:
		bs := append(cache, '[')
		cs, _ := ej.Marshal(v[0])
		bs = append(bs, cs...)
		bs = append(bs, ',')
		cs, _ = ej.Marshal(v[1])
		bs = append(bs, cs...)
		bs = append(bs, ']')
		return bs

	case val.DateTime:

	case val.Int8:
		return append(cache, JSON(strconv.FormatInt(int64(v), 10))...)

	case val.Int16:
		return append(cache, JSON(strconv.FormatInt(int64(v), 10))...)

	case val.Int32:
		return append(cache, JSON(strconv.FormatInt(int64(v), 10))...)

	case val.Int64:
		return append(cache, JSON(strconv.FormatInt(int64(v), 10))...)

	case val.Uint8:
		return append(cache, JSON(strconv.FormatUint(uint64(v), 10))...)

	case val.Uint16:
		return append(cache, JSON(strconv.FormatUint(uint64(v), 10))...)

	case val.Uint32:
		return append(cache, JSON(strconv.FormatUint(uint64(v), 10))...)

	case val.Uint64:
		return append(cache, JSON(strconv.FormatUint(uint64(v), 10))...)

	case val.Float:
		return append(cache, strconv.FormatFloat(float64(v), 'g', -1, 64)...)

	case val.Bool:
		if v {
			return append(cache, "true"...)
		}
		return append(cache, "false"...)

	case val.Symbol:
		bs, _ := ej.Marshal(v)
		return append(cache, bs...)

	}
	panic(fmt.Sprintf(`JSON encoding unimplemented for type: %T`, value))
}

func Decode(json JSON, model mdl.Model, path []string) (val.Value, err.OffsetError) {
	v, _, e := decode(json, model)
	if e != nil {
		if es, ok := e.(err.ErrorList); ok {
			// following approach pretty good for most cases
			sort.Slice(es, func(i, j int) bool {
				l, r := es[i].(err.InputParsingError), es[j].(err.InputParsingError)
				return len(l.Input) < len(r.Input)
			})
			e = es[0]
		}
		e := e.(err.InputParsingError)
		return nil, err.CodecError{"json", len(json) - len(e.Input), e}
	}
	return v, nil
}

func decode(json JSON, model mdl.Model) (val.Value, JSON, err.Error) {
	if json == nil {
		log.Panicln("json == nil")
	}
	json = skipWhiteSpace(json)
	switch m := model.(type) {
	case mdl.Null:
		json, e := readLiteral("null", json)
		if e != nil {
			return nil, json, e
		}
		return val.Null, json, nil

	case mdl.Optional:
		json, e := readLiteral("null", json)
		if e == nil {
			return val.Null, json, nil
		}
		return decode(json, m.Model)

	case mdl.Annotation:
		return decode(json, m.Model)

	case *mdl.Recursion:
		return decode(json, m.Model)

	case mdl.Unique:
		return decode(json, m.Model)

	case mdl.Set:
		vs, json, e := decodeArray(json, m.Elements)
		if e != nil {
			return nil, json, e
		}
		w := make(val.Set, len(vs))
		for _, v := range vs {
			w[val.Hash(v, nil).Sum64()] = v
		}
		return w, json, nil

	case mdl.List:
		vs, json, e := decodeArray(json, m.Elements)
		if e != nil {
			return nil, json, e
		}
		return val.List(vs), json, e

	case mdl.Tuple:
		vs, json, e := decodeTuple(json, m)
		if e != nil {
			return nil, json, e
		}
		return vs, json, nil

	case mdl.Union:
		json, e := readLiteral(`[`, json)
		if e != nil {
			return nil, json, e
		}
		caze, json, e := readString(skipWhiteSpace(json))
		if e != nil {
			return nil, json, e
		}
		element, ok := m.Get(caze)
		if !ok {
			return nil, json, err.InputParsingError{
				Problem: fmt.Sprintf(`undefined union case "%s"`, caze),
				Input:   json,
			}
		}
		json, e = readLiteral(`,`, skipWhiteSpace(json))
		if e != nil {
			return nil, json, e
		}
		value, json, e := decode(json, element)
		if e != nil {
			return nil, json, e
		}
		json, _ = readLiteral(`,`, skipWhiteSpace(json)) // optional trailing comma
		json, e = readLiteral(`]`, skipWhiteSpace(json))
		if e != nil {
			return nil, json, e
		}
		return val.Union{Case: caze, Value: value}, json, nil

	case mdl.String:
		str, json, e := readString(json)
		if e != nil {
			return nil, json, e
		}
		return val.String(str), json, nil

	case mdl.Enum:
		str, json, e := readString(json)
		if e != nil {
			return nil, json, e
		}
		if _, ok := m[str]; !ok {
			return nil, json, err.InputParsingError{
				Problem: fmt.Sprintf(`undefined enum case "%s"`, str),
				Input:   json,
			}
		}
		return val.Symbol(str), json, e

	case mdl.Bool:
		json, e := readLiteral("true", json)
		if e == nil {
			return val.Bool(true), json, nil
		}
		json, e = readLiteral("false", json)
		if e == nil {
			return val.Bool(false), json, nil
		}
		return nil, json, err.InputParsingError{
			Problem: `expected boolean true or false`,
			Input:   json,
		}

	case mdl.Map:
		vs, json, e := decodeObject(json, m.Elements)
		if e != nil {
			return nil, json, e
		}
		w := val.NewMap(len(vs))
		for k, v := range vs {
			w.Set(k, v)
		}
		return w, json, nil

	case mdl.Struct:
		vs, json, e := decodeStruct(json, m)
		if e != nil {
			return nil, json, e
		}
		return vs, json, nil

	case mdl.Ref:
		str, json, e := readString(json)
		if e != nil {
			return nil, json, e
		}
		return val.Ref{m.Model, str}, json, nil

	case mdl.DateTime:
		str, json, e := readString(json)
		if e != nil {
			return nil, json, e
		}
		t, e_ := time.Parse(time.RFC3339, str)
		if e_ != nil {
			return nil, json, err.InputParsingError{
				Problem: `malformed datetime format (must follow RFC3339)`,
				Input:   json,
			}
		}
		return val.DateTime{t}, json, nil

	case mdl.Int8:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseInt(string(n), 10, 8)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type int8`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Int8(x), json, nil

	case mdl.Int16:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseInt(string(n), 10, 16)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type int16`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Int16(x), json, nil

	case mdl.Int32:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseInt(string(n), 10, 32)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type int32`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Int32(x), json, nil

	case mdl.Int64:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseInt(string(n), 10, 64)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type int64`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Int64(x), json, nil

	case mdl.Uint8:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseUint(string(n), 10, 8)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type uint8`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Uint8(x), json, nil

	case mdl.Uint16:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseUint(string(n), 10, 16)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type uint16`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Uint16(x), json, nil

	case mdl.Uint32:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseUint(string(n), 10, 32)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type uint32`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Uint32(x), json, nil

	case mdl.Uint64:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseUint(string(n), 10, 64)
		if e_ != nil {
			ne := e_.(*strconv.NumError)
			if ne.Err == strconv.ErrRange {
				return nil, json, err.InputParsingError{
					Problem: `integer too large for type uint64`,
					Input:   json,
				}
			}
			return nil, json, err.InputParsingError{
				Problem: `malformed integer`,
				Input:   json,
			}
		}
		return val.Uint64(x), json, nil

	case mdl.Float:
		n, json, e := readJsonNumber(json)
		if e != nil {
			return nil, json, e
		}
		x, e_ := strconv.ParseFloat(string(n), 64)
		if e_ != nil {
			return nil, json, err.InputParsingError{
				Problem: `malformed float`,
				Input:   json,
			}
		}
		return val.Float(x), json, nil

	case mdl.Any:
		return nil, json, err.InputParsingError{
			Problem: `decoding typeless value not possible. This is a bug, please report it.`,
			Input:   json,
		}

	}
	panic(fmt.Sprintf(`JSON parsing unimplemented for model %T `, model))
}

// postcondition: returns intact JSON on error
func readJsonNumber(json JSON) (JSON, JSON, err.Error) {
	input := json
	if e := assertNonEmpty(json); e != nil {
		return nil, input, e
	}
	if json[0] == '-' {
		json = json[1:]
		if e := assertNonEmpty(json); e != nil {
			return nil, input, e
		}
	}
	if json[0] == '0' {
		json = json[1:]
		if len(json) == 0 {
			return input[:len(input)-len(json)], json, nil
		}
		goto dotDecimals
	}
	if json[0] < '1' || json[0] > '9' {
		return nil, input, err.InputParsingError{
			Problem: fmt.Sprintf(`expected digit between 1 and 9, found "%s"`, string(json[0])),
			Input:   json,
		}
	}
	json = json[1:]
	if len(json) == 0 {
		return input[:len(input)-len(json)], json, nil
	}
	for json[0] >= '0' && json[0] <= '9' {
		json = json[1:]
		if len(json) == 0 {
			return input[:len(input)-len(json)], json, nil
		}
	}
dotDecimals:
	if json[0] != '.' {
		goto exponent
	}
	json = json[1:]
	if len(json) == 0 {
		return input[:len(input)-len(json)], json, nil
	}
	for json[0] >= '0' && json[0] <= '9' {
		json = json[1:]
		if len(json) == 0 {
			return input[:len(input)-len(json)], json, nil
		}
	}
exponent:
	if json[0] != 'e' && json[0] != 'E' {
		return input[:len(input)-len(json)], json, nil
	}
	json = json[1:]
	if len(json) == 0 {
		return input[:len(input)-len(json)], json, nil
	}
	if json[0] == '-' || json[0] == '+' {
		json = json[1:]
		if len(json) == 0 {
			return input[:len(input)-len(json)], json, nil
		}
	}
	for json[0] >= '0' && json[0] <= '9' {
		json = json[1:]
		if len(json) == 0 {
			return input[:len(input)-len(json)], json, nil
		}
	}
	return input[:len(input)-len(json)], json, nil
}

// postcondition: returns intact JSON on error
func readString(json JSON) (string, JSON, err.Error) {
	input := json
	jstr, json, e := readJsonString(json)
	if e != nil {
		return "", input, e
	}
	value := ""
	if e := ej.Unmarshal(jstr, &value); e != nil {
		return "", input, err.InputParsingError{
			Problem: `malformed string`,
			Input:   input,
		}
	}
	return value, json, nil
}

// postcondition: returns intact JSON on error
func readJsonString(json JSON) (JSON, JSON, err.Error) {
	input := json
	json, e := readLiteral(`"`, json)
	if e != nil {
		return nil, input, e
	}
	escape := false
	for {
		if e := assertNonEmpty(json); e != nil {
			return nil, input, e
		}
		if json[0] == '"' && !escape {
			json = json[1:]
			break
		}
		if json[0] == '\\' && !escape {
			escape = true
			json = json[1:]
			continue
		}
		escape = false
		json = json[1:]
	}
	return input[:len(input)-len(json)], json, nil
}

// allows trailing commas
func decodeTuple(json JSON, tuple mdl.Tuple) (val.Tuple, JSON, err.Error) {
	json, e := readLiteral(`[`, json)
	if e != nil {
		return nil, json, e
	}
	vs := make(val.Tuple, 0, len(tuple))
	for i, l := 0, len(tuple); i < l; i++ {
		v, temp, e := decode(json, tuple[i])
		if e != nil {
			return nil, temp, e
		}
		vs, json = append(vs, v), temp
		if i == (l - 1) {
			json, _ = readLiteral(`,`, skipWhiteSpace(json)) // allow optional trailing comma
			break
		}
		json, e = readLiteral(`,`, skipWhiteSpace(json))
		if e != nil {
			return nil, json, e
		}
	}
	// closing bracket
	json, e = readLiteral(`]`, skipWhiteSpace(json))
	if e != nil {
		return nil, json, e
	}
	return vs, json, nil
}

// allows trailing commas
func decodeArray(json JSON, model mdl.Model) ([]val.Value, JSON, err.Error) {
	json, e := readLiteral(`[`, json)
	if e != nil {
		return nil, json, e
	}
	vs := make([]val.Value, 0, 16)
	for {
		if json, e := readLiteral(`]`, skipWhiteSpace(json)); e == nil {
			return vs, json, nil
		}
		v, temp, e := decode(json, model)
		if e != nil {
			return nil, temp, e
		}
		vs, json = append(vs, v), temp
		if temp, e := readLiteral(`,`, skipWhiteSpace(json)); e == nil {
			json = temp
			continue
		}
		json, e = readLiteral(`]`, skipWhiteSpace(json))
		if e != nil {
			return nil, json, e
		}
		return vs, json, nil
	}
	panic("never reached")
}

// allows trailing commas
func decodeStruct(json JSON, strct mdl.Struct) (val.Value, JSON, err.Error) {
	json, e := readLiteral(`{`, json)
	if e != nil {
		return nil, json, e
	}
	vs := val.NewStruct(strct.Len())
	for {
		if json, e = readLiteral(`}`, skipWhiteSpace(json)); e == nil {
			break
		}
		str, temp, e := readString(skipWhiteSpace(json))
		if e != nil {
			return nil, temp, e
		}
		element, ok := strct.Get(str)
		if !ok {
			return nil, json, err.InputParsingError{
				Problem: fmt.Sprintf(`undefined struct field "%s"`, str),
				Input:   json,
			}
		}
		json, e = readLiteral(`:`, skipWhiteSpace(temp))
		if e != nil {
			return nil, json, e
		}
		v, temp, e := decode(json, element)
		if e != nil {
			return nil, temp, e
		}
		vs.Set(str, v)
		json = temp
		if temp, e := readLiteral(`,`, skipWhiteSpace(json)); e == nil {
			json = temp
			continue
		}
		json, e = readLiteral(`}`, skipWhiteSpace(json))
		if e != nil {
			break
		}
		return vs, json, nil
	}
	for _, k := range strct.Keys() {
		if _, ok := vs.Get(k); ok {
			continue
		}
		m, _ := strct.Get(k)
		if m.Nullable() {
			vs.Set(k, val.Null)
			continue
		}
		return nil, json, err.InputParsingError{
			Problem: fmt.Sprintf(`missing non-optional field "%s" in struct`, k),
			Input:   json,
		}
	}

	return vs, json, nil
}

// allows trailing commas
func decodeObject(json JSON, model mdl.Model) (map[string]val.Value, JSON, err.Error) {
	json, e := readLiteral(`{`, json)
	if e != nil {
		return nil, json, e
	}
	vs := make(map[string]val.Value, 16)
	for {
		if json, e := readLiteral(`}`, skipWhiteSpace(json)); e == nil {
			return vs, json, nil
		}
		str, temp, e := readString(skipWhiteSpace(json))
		if e != nil {
			return nil, temp, e
		}
		json, e = readLiteral(`:`, skipWhiteSpace(temp))
		if e != nil {
			return nil, json, e
		}
		v, temp, e := decode(json, model)
		if e != nil {
			return nil, temp, e
		}
		vs[str], json = v, temp
		if temp, e := readLiteral(`,`, skipWhiteSpace(json)); e == nil {
			json = temp
			continue
		}
		json, e = readLiteral(`}`, skipWhiteSpace(json))
		if e != nil {
			return nil, json, e
		}
		return vs, json, nil
	}
	panic("never reached")
}

func skipWhiteSpace(json JSON) JSON {
	for len(json) > 0 && (json[0] == '\t' || json[0] == '\n' || json[0] == '\r' || json[0] == ' ') {
		json = json[1:]
	}
	if len(json) > 1 && json[0] == '/' && json[1] == '/' {
		json = json[2:]
		for len(json) > 0 && json[0] != '\n' {
			json = json[1:]
		}
		if len(json) > 0 {
			json = json[1:] // skip last \n if there is one
		}
		return skipWhiteSpace(json)
	}
	return json
}

// postcondition: returns intact JSON on error
func readLiteral(lit string, json JSON) (JSON, err.Error) {
	bs := JSON(lit)
	if len(bs) > len(json) {
		return json, err.InputParsingError{
			Problem: fmt.Sprintf(`expected "%s", input too short`, bs),
			Input:   []byte(json),
		}
	}
	cs := json[:len(bs)]
	if !bytes.Equal(bs, cs) {
		return json, err.InputParsingError{
			Problem: fmt.Sprintf(`expected "%s" but found "%s"`, bs, cs),
			Input:   []byte(json),
		}
	}
	return json[len(bs):], nil
}

func assertNonEmpty(json JSON) err.Error {
	if len(json) == 0 {
		return err.InputParsingError{
			Problem: `unexpected end of input`,
			Input:   []byte(json),
		}
	}
	return nil
}

func isNull(x JSON) bool {
	return x == nil || (len(x) == 4 && (x[0] == 'n' && x[1] == 'u' && x[2] == 'l' && x[3] == 'l'))
}

func reduceInputParsingErrorList(es err.ErrorList) err.ErrorList {
	mp := make(map[string]err.ErrorList, len(es))
	for _, e := range es {
		e := e.(err.InputParsingError)
		k := string(e.Input)
		mp[k] = append(mp[k], e)
	}
	es = es[:0] // reuse memory
	for _, l := range mp {
		es = append(es, mergeInputParsingErrors(l))
	}
	return es
}

func mergeInputParsingErrors(es err.ErrorList) err.InputParsingError {
	out := err.InputParsingError{}
	for i, e := range es {
		e := e.(err.InputParsingError)
		if i == 0 {
			out.Problem = e.Problem
		} else {
			out.Problem += " or " + e.Problem
		}
		out.Input = e.Input // same for all
	}
	return out
}
