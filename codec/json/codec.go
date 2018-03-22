// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
//
// Package codec/json is DEPRECATED and should not be used. It exists for historic reasons and will
// be removed in the near future.
package json

import (
	"encoding/json"
	"fmt"
	"karma.run/codec"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"log"
	"reflect"
	"sort"
	"time"
)

func init() {
	codec.Register("json", func() codec.Interface { return JsonCodec{} })
}

type JsonCodec struct{}

func (dec JsonCodec) Decode(data []byte, model mdl.Model) (val.Value, err.Error) {
	return Decode(data, model, nil)
}

func (dec JsonCodec) Encode(v val.Value) []byte {
	return Encode(v)
}

type JSON []byte

func (j JSON) MarshalJSON() ([]byte, error) {
	return []byte(j), nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	(*j) = append((*j)[:0], data...)
	return nil
}

func (j JSON) String() string {
	return string(j)
}

func Encode(value val.Value) JSON {
	return encode(value, make(JSON, 0, 4096))
}

func encode(value val.Value, buffer JSON) JSON {
	if value == nil {
		log.Panicln("json/codec.encode: value == nil")
	}
	if value == val.Null {
		return append(buffer, JSON(`null`)...)
	}
	switch v := value.(type) {
	case val.Meta:
		return encode(v.Value, nil)

	case val.Tuple:
		buffer = append(buffer, '[')
		for i, w := range v {
			if i > 0 {
				buffer = append(buffer, ',')
			}
			buffer = encode(w, buffer)
		}
		return append(buffer, ']')

	case val.Set:
		buffer = append(buffer, '[')
		for i, w := range v {
			if i > 0 {
				buffer = append(buffer, ',')
			}
			buffer = encode(w, buffer)
		}
		return append(buffer, ']')

	case val.List:
		buffer = append(buffer, '[')
		for i, w := range v {
			if i > 0 {
				buffer = append(buffer, ',')
			}
			buffer = encode(w, buffer)
		}
		return append(buffer, ']')

	case val.Union:
		buffer = append(buffer, '{')
		buffer = append(buffer, mustMarshal(v.Case)...)
		buffer = append(buffer, ':')
		buffer = encode(v.Value, buffer)
		return append(buffer, '}')

	case val.Struct:
		buffer = append(buffer, '{')
		first := true
		v.ForEach(func(k string, q val.Value) bool {
			if q == val.Null {
				return true // omit optional null elements in structs
			}
			if first {
				first = false
			} else {
				buffer = append(buffer, ',')
			}
			buffer = append(buffer, mustMarshal(k)...)
			buffer = append(buffer, ':')
			buffer = encode(q, buffer)
			return true
		})
		return append(buffer, '}')

	case val.Map:
		buffer = append(buffer, '{')
		first := true
		v.ForEach(func(k string, q val.Value) bool {
			if first {
				first = false
			} else {
				buffer = append(buffer, ',')
			}
			buffer = append(buffer, mustMarshal(k)...)
			buffer = append(buffer, ':')
			buffer = encode(q, buffer)
			return true
		})
		return append(buffer, '}')

	case val.Raw:
		return append(buffer, JSON(v)...)
	case val.Float:
		return append(buffer, mustMarshal(v)...)
	case val.Bool:
		return append(buffer, mustMarshal(v)...)
	case val.Symbol:
		return append(buffer, mustMarshal(v)...)
	case val.String:
		return append(buffer, mustMarshal(v)...)
	case val.Ref:
		return append(buffer, mustMarshal(v[1])...)
	case val.DateTime:
		return append(buffer, mustMarshal(v.Format(mdl.FormatDateTime))...)
	case val.Int8:
		return append(buffer, mustMarshal(v)...)
	case val.Int16:
		return append(buffer, mustMarshal(v)...)
	case val.Int32:
		return append(buffer, mustMarshal(v)...)
	case val.Int64:
		return append(buffer, mustMarshal(v)...)
	case val.Uint8:
		return append(buffer, mustMarshal(v)...)
	case val.Uint16:
		return append(buffer, mustMarshal(v)...)
	case val.Uint32:
		return append(buffer, mustMarshal(v)...)
	case val.Uint64:
		return append(buffer, mustMarshal(v)...)
	}
	panic(fmt.Sprintf(`JSON encoding unimplemented for type: %T`, value))
}

func Decode(data JSON, model mdl.Model, path []string) (val.Value, err.OffsetError) {
	v, e := decode(data, model)
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
		return nil, err.CodecError{"json", len(data) - len(e.Input), e}
	}
	return v, nil
}

func decode(data JSON, model mdl.Model) (val.Value, err.Error) {
	if data == nil {
		log.Panicln("data == nil")
	}
	switch m := model.(type) {

	case mdl.Null:
		if len(data) != 4 || string(data) != "null" {
			return nil, err.InputParsingError{"expected null", data}
		}
		return val.Null, nil

	case mdl.Or:
		v, e0 := decode(data, m[0])
		if e0 == nil {
			return v, nil
		}
		v, e1 := decode(data, m[1])
		if e1 == nil {
			return v, nil
		}
		es := make(err.ErrorList, 0, 32)
		if es0, ok := e0.(err.ErrorList); ok {
			es = append(es, es0...)
		} else {
			es = append(es, e0.(err.InputParsingError))
		}
		if es1, ok := e1.(err.ErrorList); ok {
			es = append(es, es1...)
		} else {
			es = append(es, e1.(err.InputParsingError))
		}
		return nil, es

	case mdl.Annotation:
		return decode(data, m.Model)

	case *mdl.Recursion:
		return decode(data, m.Model)

	case mdl.Unique:
		return decode(data, m.Model)

	case mdl.Set:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		u := ([]JSON)(nil)
		if e := json.Unmarshal(data, &u); e != nil {
			return nil, mapJsonError(e, data)
		}
		v := make(val.Set, len(u))
		for _, q := range u {
			w, e := decode(q, m.Elements)
			if e != nil {
				return nil, e
			}
			v[val.Hash(w, nil).Sum64()] = w
		}
		return v, nil

	case mdl.List:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		u := ([]JSON)(nil)
		if e := json.Unmarshal(data, &u); e != nil {
			return nil, mapJsonError(e, data)
		}
		v := make(val.List, len(u), len(u))
		for i, q := range u {
			w, e := decode(q, m.Elements)
			if e != nil {
				return nil, e
			}
			v[i] = w
		}
		return v, nil

	case mdl.Map:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		u := (map[string]JSON)(nil)
		if e := json.Unmarshal(data, &u); e != nil {
			if e := mapJsonError(e, data); e != nil {
				return nil, e
			}
			return nil, mapJsonError(e, data)
		}
		v := val.NewMap(len(u))
		for k, q := range u {
			w, e := decode(q, m.Elements)
			if e != nil {
				return nil, e
			}
			v.Set(k, w)
		}
		return v, nil

	case mdl.Tuple:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		u := ([]JSON)(nil)
		if e := json.Unmarshal(data, &u); e != nil {
			return nil, mapJsonError(e, data)
		}
		if len(u) != len(m) {
			return nil, err.InputParsingError{fmt.Sprintf(`expected array of length %d, have %d`, len(m), len(u)), data}
		}
		v := make(val.Tuple, len(u), len(u))
		for i, q := range u {
			w, e := decode(q, m[i])
			if e != nil {
				return nil, e
			}
			v[i] = w
		}
		return v, nil

	case mdl.Struct:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		u := make(map[string]JSON, m.Len())
		if e := json.Unmarshal(data, &u); e != nil {
			return nil, mapJsonError(e, data)
		}
		e := (err.Error)(nil)
		m.ForEach(func(k string, m mdl.Model) bool {
			if m.Nullable() {
				return true // allow optional elements to be omitted
			}
			if _, ok := u[k]; !ok {
				e = err.InputParsingError{fmt.Sprintf(`missing key "%s" in object`, k), data}
				return false
			}
			return true
		})
		if e != nil {
			return nil, e
		}
		for k, _ := range u {
			if _, ok := m.Get(k); !ok {
				return nil, err.InputParsingError{fmt.Sprintf(`unknown key "%s" in object`, k), data}
			}
		}
		v := val.NewStruct(len(u))
		for k, q := range u {
			w, e := decode(q, m.Field(k))
			if e != nil {
				return nil, e
			}
			v.Set(k, w)
		}
		m.ForEach(func(k string, _ mdl.Model) bool {
			if _, ok := v.Get(k); !ok {
				v.Set(k, val.Null) // omitted optional elements should be Null in struct val
			}
			return true
		})
		return v, nil

	case mdl.Union:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		s := m
		u := make(map[string]JSON, 1)
		if e := json.Unmarshal(data, &u); e != nil {
			return nil, mapJsonError(e, data)
		}
		if len(u) != 1 {
			return nil, err.InputParsingError{fmt.Sprintf(`expected 1 key in object, have %d`, len(u)), data}
		}
		k, j := "", JSON(nil)
		for a, b := range u {
			k, j = a, b
			break
		}
		sk, ok := s.Get(k)
		if !ok {
			return nil, err.InputParsingError{fmt.Sprintf(`unknown union case in object: "%s"`, k), data}
		}
		w, e := decode(j, sk)
		if e != nil {
			return nil, e
		}
		v := val.Union{Case: k, Value: w}
		return v, nil

	case mdl.String:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.String("")
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Enum:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := ""
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		if _, ok := m[v]; ok {
			return val.Symbol(v), nil
		}
		return nil, err.InputParsingError{fmt.Sprintf(`unknown enum symbol: "%s"`, v), data}

	case mdl.Float:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Float(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Bool:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Bool(false)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Any:
		return val.Raw(data), nil

	case mdl.Ref:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := ""
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return val.Ref{m.Model, v}, nil

	case mdl.DateTime:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := ""
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		t, e := time.Parse(mdl.FormatDateTime, v)
		if e != nil {
			return nil, err.InputParsingError{fmt.Sprintf(`invalid dateTime string: %s`, v), data}
		}
		return val.DateTime{t}, nil

	case mdl.Int8:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Int8(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Int16:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Int16(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Int32:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Int32(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Int64:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Int64(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Uint8:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Uint8(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Uint16:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Uint16(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Uint32:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Uint32(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	case mdl.Uint64:
		if isNull(data) {
			return nil, err.InputParsingError{"unexpected null", data}
		}
		v := val.Uint64(0)
		if e := json.Unmarshal(data, &v); e != nil {
			return nil, mapJsonError(e, data)
		}
		return v, nil

	}
	panic(fmt.Sprintf(`JSON parsing unimplemented for model %T `, model))
}

func isNull(x JSON) bool {
	return x == nil || (len(x) == 4 && (x[0] == 'n' && x[1] == 'u' && x[2] == 'l' && x[3] == 'l'))
}

func mustMarshal(v interface{}) []byte {
	bs, e := json.Marshal(v)
	if e != nil {
		log.Panicln(e.Error())
	}
	return bs
}

func mapJsonError(e error, bs []byte) err.Error {
	switch e := e.(type) {
	case *json.SyntaxError:
		return err.InputParsingError{e.Error(), bs[e.Offset:]}
	case *json.UnmarshalTypeError:
		expected := "?"
		switch e.Type.Kind() {
		case reflect.Map:
			expected = "object"
		case reflect.Slice:
			expected = "array"
		case reflect.Bool:
			expected = "boolean"
		case reflect.String:
			expected = "string"
		case reflect.Float64:
			expected = "number"
		case reflect.Int8:
			expected = "number (that fits int8)"
		case reflect.Int16:
			expected = "number (that fits int16)"
		case reflect.Int32:
			expected = "number (that fits int32)"
		case reflect.Int64:
			expected = "number (that fits int64)"
		case reflect.Uint8:
			expected = "number (that fits uint8)"
		case reflect.Uint16:
			expected = "number (that fits uint16)"
		case reflect.Uint32:
			expected = "number (that fits uint32)"
		case reflect.Uint64:
			expected = "number (that fits uint64)"
		}
		return err.InputParsingError{fmt.Sprintf(`expected %s, have %s`, expected, e.Value), bs[e.Offset-1:]}
	}
	return err.InputParsingError{`error parsing json`, bs}
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
