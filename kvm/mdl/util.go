// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"fmt"
	"karma.run/kvm/val"
)

// NOTE: Abstain from using directly!
//       Consider calling kvm.inferType instead.
func TightestModelForValue(v val.Value) Model {
	if v == val.Null {
		return Null{}
	}
	switch v := v.(type) {
	case val.List:
		if len(v) == 0 {
			return List{Any{}}
		}
		ms := make([]Model, 0, len(v))
		for _, w := range v {
			ms = append(ms, TightestModelForValue(w))
		}
		return List{UnionOf(ms...)}

	case val.Map:
		if v.Len() == 0 {
			return Map{Any{}}
		}
		ms := make([]Model, 0, v.Len())
		v.ForEach(func(k string, v val.Value) bool {
			ms = append(ms, TightestModelForValue(v))
			return true
		})
		return Map{UnionOf(ms...)}

	case val.Struct:
		s := NewStruct(v.Len())
		v.ForEach(func(k string, w val.Value) bool {
			s.Set(k, TightestModelForValue(w))
			return true
		})
		return s

	case val.Set:
		if len(v) == 0 {
			return Set{Any{}}
		}
		ms := make([]Model, 0, len(v))
		for _, w := range v {
			ms = append(ms, TightestModelForValue(w))
		}
		return Set{UnionOf(ms...)}

	case val.Tuple:
		m := make(Tuple, len(v), len(v))
		for i, w := range v {
			m[i] = TightestModelForValue(w)
		}
		return m

	case val.Symbol:
		return Enum{string(v): struct{}{}}
	case val.Bool:
		return Bool{}
	case val.DateTime:
		return DateTime{}
	case val.Float:
		return Float{}
	case val.String:
		return String{}
	case val.Union:
		u := NewUnion(1)
		u.Set(v.Case, TightestModelForValue(v.Value))
		return u
	case val.Int8:
		return Int8{}
	case val.Int16:
		return Int16{}
	case val.Int32:
		return Int32{}
	case val.Int64:
		return Int64{}
	case val.Uint8:
		return Uint8{}
	case val.Uint16:
		return Uint16{}
	case val.Uint32:
		return Uint32{}
	case val.Uint64:
		return Uint64{}
	case val.Ref:
		return Ref{v[0]}
	}
	panic(fmt.Sprintf("TightestModelForValue undefined for type %T\n", v))
}

func copyStringSlice(ss []string) []string {
	cp := make([]string, len(ss), len(ss))
	for i, _ := range ss {
		cp[i] = ss[i]
	}
	return cp
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
