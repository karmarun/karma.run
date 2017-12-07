// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
//
// Package codec/karma is DEPRECATED and should not be used. It exists for historic reasons and will
// be removed in the near future.
package karma

import (
	"encoding/binary"
	"fmt"
	"karma.run/kvm/val"
	"math"
	"time"
	"unsafe"
)

type Type byte

const (
	TypeTuple Type = iota
	TypeList
	TypeUnion
	TypeRaw
	TypeStruct
	TypeMap
	TypeInt // kept for backward-compatibility
	TypeFloat
	TypeBool
	TypeString
	TypeRef
	TypeDateTime
	TypeNull
	TypeSymbol
	TypeSet
	TypeUint // kept for backward-compatibility
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
)

func Encode(v val.Value, bs []byte) []byte {

	if bs == nil {
		bs = make([]byte, 0, 1024)
	}

	switch w := v.(type) {

	case val.Meta:
		return Encode(w.Value, bs)

	case val.Tuple:
		bs = append(bs, byte(TypeTuple))
		bs = encodeUint(uint64(len(w)), bs)
		for _, x := range w {
			bs = Encode(x, bs)
		}
		return bs

	case val.List:
		bs = append(bs, byte(TypeList))
		bs = encodeUint(uint64(len(w)), bs)
		for _, x := range w {
			bs = Encode(x, bs)
		}
		return bs

	case val.Set:
		bs = append(bs, byte(TypeSet))
		bs = encodeUint(uint64(len(w)), bs)
		for _, x := range w {
			bs = Encode(x, bs)
		}
		return bs

	case val.Union:
		bs = append(bs, byte(TypeUnion))
		bs = encodeString(w.Case, bs)
		return Encode(w.Value, bs)

	case val.Raw:
		bs = append(bs, byte(TypeRaw))
		return encodeString(string(w), bs)

	case val.Struct:
		bs = append(bs, byte(TypeStruct))
		bs = encodeUint(uint64(len(w)), bs)
		for k, x := range w {
			bs = encodeString(k, bs)
			bs = Encode(x, bs)
		}
		return bs

	case val.Map:
		bs = append(bs, byte(TypeMap))
		bs = encodeUint(uint64(len(w)), bs)
		for k, x := range w {
			bs = encodeString(k, bs)
			bs = Encode(x, bs)
		}
		return bs

	case val.Float:
		bs = append(bs, byte(TypeFloat))
		return encodeUint(math.Float64bits(float64(w)), bs)

	case val.Bool:
		bs = append(bs, byte(TypeBool))
		if w {
			return append(bs, 1)
		}
		return append(bs, 0)

	case val.Symbol:
		bs = append(bs, byte(TypeSymbol))
		return encodeString(string(w), bs)

	case val.String:
		bs = append(bs, byte(TypeString))
		return encodeString(string(w), bs)

	case val.Ref:
		bs = append(bs, byte(TypeRef))
		return append(bs, (w[0] + w[1])...)

	case val.DateTime:
		bs = append(bs, byte(TypeDateTime))
		return encodeInt(w.UnixNano(), bs)

	case val.Null:
		return append(bs, byte(TypeNull))

	case val.Int8:
		bs = append(bs, byte(TypeInt8))
		return append(bs, byte(w))

	case val.Uint8:
		bs = append(bs, byte(TypeUint8))
		return append(bs, byte(w))

	case val.Int16:
		bs = append(bs, byte(TypeInt16))
		cs := *(*[2]byte)((unsafe.Pointer)(&w))
		return append(bs, cs[0], cs[1])

	case val.Uint16:
		bs = append(bs, byte(TypeUint16))
		cs := *(*[2]byte)((unsafe.Pointer)(&w))
		return append(bs, cs[0], cs[1])

	case val.Int32:
		bs = append(bs, byte(TypeInt32))
		cs := *(*[4]byte)((unsafe.Pointer)(&w))
		return append(bs, cs[0], cs[1], cs[2], cs[3])

	case val.Uint32:
		bs = append(bs, byte(TypeUint32))
		cs := *(*[4]byte)((unsafe.Pointer)(&w))
		return append(bs, cs[0], cs[1], cs[2], cs[3])

	case val.Int64:
		bs = append(bs, byte(TypeInt64))
		cs := *(*[8]byte)((unsafe.Pointer)(&w))
		return append(bs, cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7])

	case val.Uint64:
		bs = append(bs, byte(TypeUint64))
		cs := *(*[8]byte)((unsafe.Pointer)(&w))
		return append(bs, cs[0], cs[1], cs[2], cs[3], cs[4], cs[5], cs[6], cs[7])

	}
	panic(fmt.Sprintf("%T", v))
}

func encodeInt(n int64, bs []byte) []byte {
	tmp := [10]byte{}
	ln := binary.PutVarint(tmp[:], n)
	bs = append(bs, tmp[:ln]...)
	return bs
}

func encodeUint(n uint64, bs []byte) []byte {
	tmp := [10]byte{}
	ln := binary.PutUvarint(tmp[:], n)
	bs = append(bs, tmp[:ln]...)
	return bs
}

func encodeString(s string, bs []byte) []byte {
	bs = encodeUint(uint64(len(s)), bs)
	bs = append(bs, s...)
	return bs
}

func Decode(bs []byte) (val.Value, []byte) {
	c := Type(bs[0])
	bs = bs[1:]
	switch c {
	case TypeTuple:
		n, bs := decodeUint(bs)
		l := int(n)
		v := make(val.Tuple, l, l)
		for i := 0; i < l; i++ {
			v[i], bs = Decode(bs)
		}
		return v, bs

	case TypeList:
		n, bs := decodeUint(bs)
		l := int(n)
		v := make(val.List, l, l)
		for i := 0; i < l; i++ {
			v[i], bs = Decode(bs)
		}
		return v, bs

	case TypeSet:
		n, bs := decodeUint(bs)
		l := int(n)
		v := make(val.Set, l)
		for i := 0; i < l; i++ {
			w, cs := Decode(bs)
			v[val.Hash(w, nil).Sum64()] = w
			bs = cs
		}
		return v, bs

	case TypeUnion:
		s, bs := decodeString(bs)
		v, bs := Decode(bs)
		return val.Union{s, v}, bs

	case TypeRaw:
		s, bs := decodeString(bs)
		return val.Raw(s), bs

	case TypeStruct:
		n, bs := decodeUint(bs)
		l := int(n)
		v := make(val.Struct, l)
		k, w := "", (val.Value)(nil)
		for i := 0; i < l; i++ {
			k, bs = decodeString(bs)
			w, bs = Decode(bs)
			v[k] = w
		}
		return v, bs

	case TypeMap:
		n, bs := decodeUint(bs)
		l := int(n)
		v := make(val.Map, l)
		k, w := "", (val.Value)(nil)
		for i := 0; i < l; i++ {
			k, bs = decodeString(bs)
			w, bs = Decode(bs)
			v[k] = w
		}
		return v, bs

	case TypeFloat:
		n, bs := decodeUint(bs)
		return val.Float(math.Float64frombits(n)), bs

	case TypeBool:
		if bs[0] == 1 {
			return val.Bool(true), bs[1:]
		}
		return val.Bool(false), bs[1:]

	case TypeString:
		s, bs := decodeString(bs)
		return val.String(s), bs

	case TypeSymbol:
		s, bs := decodeString(bs)
		return val.String(s), bs

	case TypeRef:
		a, b, c := bs[:16], bs[16:32], bs[32:]
		return val.Ref{string(a), string(b)}, c

	case TypeDateTime:
		n, bs := decodeInt(bs)
		t := time.Unix(n/1000000000, n%1000000000)
		return val.DateTime{t}, bs

	case TypeNull:
		return val.Null{}, bs

	case TypeInt:
		n, bs := decodeInt(bs)
		return val.Int64(n), bs

	case TypeUint:
		n, bs := decodeUint(bs)
		return val.Uint64(n), bs

	case TypeInt8:
		n := (val.Int8)(bs[0])
		return n, bs[1:]

	case TypeUint8:
		n := (val.Uint8)(bs[0])
		return n, bs[1:]

	case TypeInt16:
		n := *(*val.Int16)((unsafe.Pointer)(&bs[0]))
		return n, bs[2:]

	case TypeUint16:
		n := *(*val.Uint16)((unsafe.Pointer)(&bs[0]))
		return n, bs[2:]

	case TypeInt32:
		n := *(*val.Int32)((unsafe.Pointer)(&bs[0]))
		return n, bs[4:]

	case TypeUint32:
		n := *(*val.Uint32)((unsafe.Pointer)(&bs[0]))
		return n, bs[4:]

	case TypeInt64:
		n := *(*val.Int64)((unsafe.Pointer)(&bs[0]))
		return n, bs[8:]

	case TypeUint64:
		n := *(*val.Uint64)((unsafe.Pointer)(&bs[0]))
		return n, bs[8:]

	}
	panic(fmt.Sprintf("%v", c))
}

func decodeInt(bs []byte) (int64, []byte) {
	n, l := binary.Varint(bs)
	return n, bs[l:]
}

func decodeUint(bs []byte) (uint64, []byte) {
	n, l := binary.Uvarint(bs)
	return n, bs[l:]
}

func decodeString(bs []byte) (string, []byte) {
	l, bs := decodeUint(bs)
	return string(bs[:l]), bs[l:]
}
