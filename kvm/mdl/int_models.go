// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"karma.run/kvm/val"
)

type Int8 struct{}

func (r Int8) Zero() val.Value {
	return val.Int8(0)
}

func (m Int8) Transform(f func(Model) Model) Model {
	return f(Int8{})
}

func (i Int8) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Int8) Copy() Model {
	return i
}

func (r Int8) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Int8) Union(w Model) Model {
	if _, ok := w.(Int8); ok {
		return Int8{} // e.g. int8 | int8 = int8
	}
	return Or{m, w}
}

type Int16 struct{}

func (r Int16) Zero() val.Value {
	return val.Int16(0)
}

func (m Int16) Transform(f func(Model) Model) Model {
	return f(Int16{})
}

func (i Int16) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Int16) Copy() Model {
	return i
}

func (r Int16) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Int16) Union(w Model) Model {
	if _, ok := w.(Int16); ok {
		return Int16{} // e.g. int16 | int16 = int16
	}
	return Or{m, w}
}

type Int32 struct{}

func (r Int32) Zero() val.Value {
	return val.Int32(0)
}

func (m Int32) Transform(f func(Model) Model) Model {
	return f(Int32{})
}

func (i Int32) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Int32) Copy() Model {
	return i
}

func (r Int32) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Int32) Union(w Model) Model {
	if _, ok := w.(Int32); ok {
		return Int32{} // e.g. int32 | int32 = int32
	}
	return Or{m, w}
}

type Int64 struct{}

func (r Int64) Zero() val.Value {
	return val.Int64(0)
}

func (m Int64) Transform(f func(Model) Model) Model {
	return f(Int64{})
}

func (i Int64) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Int64) Copy() Model {
	return i
}

func (r Int64) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Int64) Union(w Model) Model {
	if _, ok := w.(Int64); ok {
		return Int64{} // e.g. int64 | int64 = int64
	}
	return Or{m, w}
}

type Uint8 struct{}

func (r Uint8) Zero() val.Value {
	return val.Uint8(0)
}

func (m Uint8) Transform(f func(Model) Model) Model {
	return f(Uint8{})
}

func (i Uint8) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Uint8) Copy() Model {
	return i
}

func (r Uint8) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Uint8) Union(w Model) Model {
	if _, ok := w.(Uint8); ok {
		return Uint8{} // e.g. uint8 | uint8 = uint8
	}
	return Or{m, w}
}

type Uint16 struct{}

func (r Uint16) Zero() val.Value {
	return val.Uint16(0)
}

func (m Uint16) Transform(f func(Model) Model) Model {
	return f(Uint16{})
}

func (i Uint16) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Uint16) Copy() Model {
	return i
}

func (r Uint16) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Uint16) Union(w Model) Model {
	if _, ok := w.(Uint16); ok {
		return Uint16{} // e.g. uint16 | uint16 = uint16
	}
	return Or{m, w}
}

type Uint32 struct{}

func (r Uint32) Zero() val.Value {
	return val.Uint32(0)
}

func (m Uint32) Transform(f func(Model) Model) Model {
	return f(Uint32{})
}

func (i Uint32) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Uint32) Copy() Model {
	return i
}

func (r Uint32) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Uint32) Union(w Model) Model {
	if _, ok := w.(Uint32); ok {
		return Uint32{} // e.g. uint32 | uint32 = uint32
	}
	return Or{m, w}
}

type Uint64 struct{}

func (r Uint64) Zero() val.Value {
	return val.Uint64(0)
}

func (m Uint64) Transform(f func(Model) Model) Model {
	return f(Uint64{})
}

func (i Uint64) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, i)
}

func (i Uint64) Copy() Model {
	return i
}

func (r Uint64) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Uint64) Union(w Model) Model {
	if _, ok := w.(Uint64); ok {
		return Uint64{} // e.g. uint64 | uint64 = uint64
	}
	return Or{m, w}
}

func (m Int8) Concrete() Model {
	return m
}
func (m Int16) Concrete() Model {
	return m
}
func (m Int32) Concrete() Model {
	return m
}
func (m Int64) Concrete() Model {
	return m
}

func (m Uint8) Concrete() Model {
	return m
}
func (m Uint16) Concrete() Model {
	return m
}
func (m Uint32) Concrete() Model {
	return m
}
func (m Uint64) Concrete() Model {
	return m
}

func (m Int8) Equals(n Model) bool {
	_, ok := n.(Int8)
	return ok
}
func (m Int16) Equals(n Model) bool {
	_, ok := n.(Int16)
	return ok
}
func (m Int32) Equals(n Model) bool {
	_, ok := n.(Int32)
	return ok
}
func (m Int64) Equals(n Model) bool {
	_, ok := n.(Int64)
	return ok
}

func (m Uint8) Equals(n Model) bool {
	_, ok := n.(Uint8)
	return ok
}
func (m Uint16) Equals(n Model) bool {
	_, ok := n.(Uint16)
	return ok
}
func (m Uint32) Equals(n Model) bool {
	_, ok := n.(Uint32)
	return ok
}
func (m Uint64) Equals(n Model) bool {
	_, ok := n.(Uint64)
	return ok
}

func (Int8) Nullable() bool {
	return false
}

func (Int16) Nullable() bool {
	return false
}

func (Int32) Nullable() bool {
	return false
}

func (Int64) Nullable() bool {
	return false
}

func (Uint8) Nullable() bool {
	return false
}

func (Uint16) Nullable() bool {
	return false
}

func (Uint32) Nullable() bool {
	return false
}

func (Uint64) Nullable() bool {
	return false
}

func (Int8) Zeroable() bool {
	return true
}

func (Int16) Zeroable() bool {
	return true
}

func (Int32) Zeroable() bool {
	return true
}

func (Int64) Zeroable() bool {
	return true
}

func (Uint8) Zeroable() bool {
	return true
}

func (Uint16) Zeroable() bool {
	return true
}

func (Uint32) Zeroable() bool {
	return true
}

func (Uint64) Zeroable() bool {
	return true
}

func (m Int8) Unwrap() Model {
	return m
}

func (m Int16) Unwrap() Model {
	return m
}

func (m Int32) Unwrap() Model {
	return m
}

func (m Int64) Unwrap() Model {
	return m
}

func (m Uint8) Unwrap() Model {
	return m
}

func (m Uint16) Unwrap() Model {
	return m
}

func (m Uint32) Unwrap() Model {
	return m
}

func (m Uint64) Unwrap() Model {
	return m
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
