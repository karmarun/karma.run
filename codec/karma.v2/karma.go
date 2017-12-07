// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package karma

import (
	"encoding/binary"
	"fmt"
	"karma.run/codec/karma"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"math"
	"sort"
	"time"
)

func Encode(v val.Value, m mdl.Model) []byte {
	return encode(v, m, make([]byte, 0, 1024*4))
}

func encode(v val.Value, m mdl.Model, bs []byte) []byte {
	m = m.Concrete()
	switch m := m.(type) {
	case mdl.Or:
		t := v.Type()
		mm := disambiguateModel(t, m)
		if mm == nil {
			panic("no model found for value in mdl.Or: " + t.String() + " vs " + m.ValueType().String())
		}
		bs = writeUint64(uint64(t), bs)
		return encode(v, mm, bs)

	case mdl.Any:
		return karma.Encode(v, bs)

	case mdl.Set:
		v := v.(val.Set)
		bs = writeUint32(uint32(len(v)), bs)
		for _, w := range v {
			bs = encode(w, m.Elements, bs)
		}
		return bs

	case mdl.List:
		v := v.(val.List)
		bs = writeUint32(uint32(len(v)), bs)
		for _, w := range v {
			bs = encode(w, m.Elements, bs)
		}
		return bs

	case mdl.Map:
		v := v.(val.Map)
		bs = writeUint32(uint32(len(v)), bs)
		for k, w := range v {
			bs = writeString(k, bs)
			bs = encode(w, m.Elements, bs)
		}
		return bs

	case mdl.Tuple:
		v := v.(val.Tuple)
		for i, w := range v {
			bs = encode(w, m[i], bs)
		}
		return bs

	case mdl.Struct:
		v := v.(val.Struct)
		ks := m.Keys()
		sort.Strings(ks)
		for _, k := range ks {
			w, ok := v[k]
			if !ok {
				w = val.Null{} // m[k] is (null|something)
			}
			bs = encode(w, m[k], bs)
		}
		return bs

	case mdl.Union:
		v := v.(val.Union)
		bs = writeString(v.Case, bs)
		return encode(v.Value, m[v.Case], bs)

	case mdl.Enum:
		v := v.(val.Symbol)
		return writeString(string(v), bs)

	case mdl.Ref:
		v := v.(val.Ref)
		return append(bs, v[1]...)

	case mdl.Null:
		return bs

	case mdl.String:
		v := v.(val.String)
		return writeString(string(v), bs)

	case mdl.Float:
		v := v.(val.Float)
		x := math.Float64bits(float64(v))
		return writeUint64(x, bs)

	case mdl.Bool:
		v := v.(val.Bool)
		if v {
			return append(bs, 1)
		}
		return append(bs, 0)

	case mdl.DateTime:
		v := v.(val.DateTime)
		t, e := v.MarshalBinary()
		if e != nil {
			panic(e)
		}
		if len(t) > 255 {
			panic("time.Time binary encoding unexpectedly long")
		}
		return append(append(bs, uint8(len(t))), t...)

	case mdl.Int8:
		v := v.(val.Int8)
		return writeInt8(int8(v), bs)

	case mdl.Int16:
		v := v.(val.Int16)
		return writeInt16(int16(v), bs)

	case mdl.Int32:
		v := v.(val.Int32)
		return writeInt32(int32(v), bs)

	case mdl.Int64:
		v := v.(val.Int64)
		return writeInt64(int64(v), bs)

	case mdl.Uint8:
		v := v.(val.Uint8)
		return writeUint8(uint8(v), bs)

	case mdl.Uint16:
		v := v.(val.Uint16)
		return writeUint16(uint16(v), bs)

	case mdl.Uint32:
		v := v.(val.Uint32)
		return writeUint32(uint32(v), bs)

	case mdl.Uint64:
		v := v.(val.Uint64)
		return writeUint64(uint64(v), bs)

	}
	panic(fmt.Sprintf("unhandled model: %T", m))
}

func Decode(bs []byte, m mdl.Model) (val.Value, []byte) {
	return decode(bs, m)
}

func decode(bs []byte, m mdl.Model) (val.Value, []byte) {
	m = m.Concrete()
	switch m := m.(type) {
	case mdl.Or:
		ts, bs := readUint64(bs)
		mm := disambiguateModel(val.Type(ts), m)
		if mm == nil {
			panic("no model found for type marker in mdl.Or")
		}
		return decode(bs, mm)

	case mdl.Any:
		return karma.Decode(bs)

	case mdl.Set:
		l, bs := readUint32(bs)
		v := make(val.Set, l)
		for i, l := 0, int(l); i < l; i++ {
			w, cs := decode(bs, m.Elements)
			v[val.Hash(w, nil).Sum64()] = w
			bs = cs
		}
		return v, bs

	case mdl.List:
		l, bs := readUint32(bs)
		v := make(val.List, l, l)
		for i, l := 0, int(l); i < l; i++ {
			w, cs := decode(bs, m.Elements)
			v[i], bs = w, cs
		}
		return v, bs

	case mdl.Map:
		l, bs := readUint32(bs)
		v := make(val.Map, l)
		for i, l := 0, int(l); i < l; i++ {
			k, cs := readString(bs)
			w, cs := decode(cs, m.Elements)
			v[k], bs = w, cs
		}
		return v, bs

	case mdl.Tuple:
		l := len(m)
		v := make(val.Tuple, l, l)
		for i := 0; i < l; i++ {
			v[i], bs = decode(bs, m[i])
		}
		return v, bs

	case mdl.Struct:
		ks := m.Keys()
		sort.Strings(ks)
		v := make(val.Struct, len(ks))
		for _, k := range ks {
			v[k], bs = decode(bs, m[k])
		}
		return v, bs

	case mdl.Union:
		c, bs := readString(bs)
		v := val.Union{Case: c}
		v.Value, bs = decode(bs, m[c])
		return v, bs

	case mdl.Enum:
		s, bs := readString(bs)
		return val.Symbol(s), bs

	case mdl.Ref:
		r, bs := string(bs[:16]), bs[16:]
		return val.Ref{m.Model, r}, bs

	case mdl.Null:
		return val.Null{}, bs

	case mdl.String:
		s, bs := readString(bs)
		return val.String(s), bs

	case mdl.Float:
		bits, bs := readUint64(bs)
		f := math.Float64frombits(bits)
		return val.Float(f), bs

	case mdl.Bool:
		return val.Bool(bs[0] == 1), bs[1:]

	case mdl.DateTime:
		l, bs := int(bs[0]), bs[1:]
		ts, bs := bs[:l], bs[l:]
		t := &time.Time{}
		if e := t.UnmarshalBinary(ts); e != nil {
			panic(e)
		}
		return val.DateTime{*t}, bs

	case mdl.Int8:
		return val.Int8(bs[0]), bs[1:]

	case mdl.Int16:
		x, bs := readInt16(bs)
		return val.Int16(x), bs

	case mdl.Int32:
		x, bs := readInt32(bs)
		return val.Int32(x), bs

	case mdl.Int64:
		x, bs := readInt64(bs)
		return val.Int64(x), bs

	case mdl.Uint8:
		return val.Uint8(bs[0]), bs[1:]

	case mdl.Uint16:
		x, bs := readUint16(bs)
		return val.Uint16(x), bs

	case mdl.Uint32:
		x, bs := readUint32(bs)
		return val.Uint32(x), bs

	case mdl.Uint64:
		x, bs := readUint64(bs)
		return val.Uint64(x), bs

	}
	panic(fmt.Sprintf("unhandled model: %T", m))
}

func readUint64(bs []byte) (uint64, []byte) {
	x, n := binary.Uvarint(bs)
	return uint64(x), bs[n:]
}

func readUint32(bs []byte) (uint32, []byte) {
	x, n := binary.Uvarint(bs)
	return uint32(x), bs[n:]
}

func readUint16(bs []byte) (uint16, []byte) {
	h, l := bs[0], bs[1]
	return uint16(h)<<8 | uint16(l), bs[2:]
}

func readInt64(bs []byte) (int64, []byte) {
	x, n := binary.Varint(bs)
	return int64(x), bs[n:]
}

func readInt32(bs []byte) (int32, []byte) {
	x, n := binary.Varint(bs)
	return int32(x), bs[n:]
}

func readInt16(bs []byte) (int16, []byte) {
	h, l := bs[0], bs[1]
	return int16(h)<<8 | int16(l), bs[2:]
}

func readString(bs []byte) (string, []byte) {
	l, bs := readUint32(bs)
	return string(bs[:l]), bs[l:]
}

func writeString(s string, bs []byte) []byte {
	bs = writeUint32(uint32(len(s)), bs)
	return append(bs, s...)
}

func writeInt64(x int64, bs []byte) []byte {
	t := [binary.MaxVarintLen64]byte{}
	n := binary.PutVarint(t[:], int64(x))
	return append(bs, t[:n]...)
}

func writeInt32(x int32, bs []byte) []byte {
	t := [binary.MaxVarintLen32]byte{}
	n := binary.PutVarint(t[:], int64(x))
	return append(bs, t[:n]...)
}

func writeInt16(x int16, bs []byte) []byte {
	h, l := uint8(x>>8), uint8(x)
	return append(bs, h, l)
}

func writeInt8(x int8, bs []byte) []byte {
	return append(bs, uint8(x))
}

func writeUint64(x uint64, bs []byte) []byte {
	t := [binary.MaxVarintLen64]byte{}
	n := binary.PutUvarint(t[:], uint64(x))
	return append(bs, t[:n]...)
}

func writeUint32(x uint32, bs []byte) []byte {
	t := [binary.MaxVarintLen32]byte{}
	n := binary.PutUvarint(t[:], uint64(x))
	return append(bs, t[:n]...)
}

func writeUint16(x uint16, bs []byte) []byte {
	h, l := uint8(x>>8), uint8(x)
	return append(bs, h, l)
}

func writeUint8(x uint8, bs []byte) []byte {
	return append(bs, x)
}

func disambiguateModel(t val.Type, m mdl.Model) mdl.Model {
	m = m.Concrete()
	// NOTE: given that Ors can be non-minimal, this check
	//       must come before m.ValueType() == t
	if or, ok := m.(mdl.Or); ok {
		l := disambiguateModel(t, or[0])
		r := disambiguateModel(t, or[1])
		if l == nil {
			return r
		}
		if r == nil {
			return l
		}
		return mdl.Either(l, r, nil)
	}
	if m.ValueType() == t {
		return m
	}
	if _, ok := m.(mdl.Any); ok {
		return m
	}
	return nil
}
