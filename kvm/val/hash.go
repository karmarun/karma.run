// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package val

import (
	"fmt"
	"hash"
	"hash/fnv"
	"sort"
	"unsafe"
)

func Hash(v Value, h hash.Hash64) hash.Hash64 {
	if h == nil {
		h = fnv.New64()
	}
	switch v := v.(type) {
	case Meta:
		h.Write([]byte(`meta`))
		h = Hash(v.Id, h)
		h = Hash(v.Model, h)
		h = Hash(v.Created, h)
		h = Hash(v.Updated, h)
		h = Hash(v.Value, h)
		return h
	case Tuple:
		h.Write([]byte(`tuple`))
		for _, w := range v {
			h = Hash(w, h)
		}
		return h
	case List:
		h.Write([]byte(`list`))
		for _, w := range v {
			h = Hash(w, h)
		}
		return h
	case Union:
		h.Write([]byte(`union`))
		h.Write([]byte(v.Case))
		h = Hash(v.Value, h)
		return h
	case Raw:
		h.Write([]byte(`raw`))
		h.Write([]byte(v))
		return h
	case Struct:
		h.Write([]byte(`struct`))
		v.ForEach(func(k string, v Value) bool {
			h.Write([]byte(k))
			h = Hash(v, h)
			return true
		})
		return h
	case Map:
		h.Write([]byte(`map`))
		ks := v.Keys()
		sort.Strings(ks)
		for _, k := range ks {
			h.Write([]byte(k))
			h = Hash(v[k], h)
		}
		return h
	case Float:
		h.Write([]byte(`float`))
		x := float64(v)
		b := *(*[8]byte)((unsafe.Pointer)(&x))
		h.Write(b[:])
		return h
	case Bool:
		h.Write([]byte(`bool`))
		if v {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
		return h
	case String:
		h.Write([]byte(`string`))
		h.Write([]byte(v))
		return h
	case Ref:
		h.Write([]byte(`ref`))
		h.Write([]byte(v[0] + v[1]))
		return h
	case DateTime:
		h.Write([]byte(`datetime`))
		h.Write([]byte(v.Time.String()))
		return h
	case null:
		h.Write([]byte(`null`))
		return h
	case Symbol:
		h.Write([]byte(`symbol`))
		h.Write([]byte(v))
		return h
	case Set:
		h.Write([]byte(`set`))
		ks := v.Keys()
		sort.Slice(ks, func(i, j int) bool {
			return ks[i] < ks[j]
		})
		for _, k := range ks {
			h = Hash(v[k], h)
		}
		return h

	case Int8:
		h.Write([]byte(`int8`))
		h.Write([]byte{byte(v)})
		return h

	case Int16:
		h.Write([]byte(`int16`))
		h.Write([]byte{byte(v >> 8), byte(v)})
		return h

	case Int32:
		h.Write([]byte(`int32`))
		h.Write([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
		return h

	case Int64:
		h.Write([]byte(`int64`))
		h.Write([]byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
		return h

	case Uint8:
		h.Write([]byte(`uint8`))
		h.Write([]byte{byte(v)})
		return h

	case Uint16:
		h.Write([]byte(`uint16`))
		h.Write([]byte{byte(v >> 8), byte(v)})
		return h

	case Uint32:
		h.Write([]byte(`uint32`))
		h.Write([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
		return h

	case Uint64:
		h.Write([]byte(`uint64`))
		h.Write([]byte{byte(v >> 56), byte(v >> 48), byte(v >> 40), byte(v >> 32), byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})
		return h

	}
	panic(fmt.Sprintf("unhandled type: %T", v))
}
