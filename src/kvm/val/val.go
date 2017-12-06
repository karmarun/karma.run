// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package val

import (
	"time"
)

type Value interface {
	Copy() Value
	Equals(Value) bool
	Transform(func(Value) Value) Value
	Primitive() bool
	Type() Type
}

// Meta is a Value that describes the attributes associated
// with a persisted Value. Encoders should never write
// it as-is. Only its Value field should be encoded.
type Meta struct {
	Id      Ref
	Model   Ref
	Created DateTime
	Updated DateTime
	Value   Value
}

func (v Meta) Transform(f func(Value) Value) Value {
	return v.Value.Transform(f)
}

func (m Meta) Equals(v Value) bool {
	n, ok := v.(Meta)
	if !ok {
		return false
	}
	return m.Id == n.Id && m.Model == n.Model && m.Created == n.Created && m.Updated == n.Updated && m.Value.Equals(n.Value)
}

func (m Meta) Copy() Value {
	return Meta{
		Id:      m.Id.Copy().(Ref),
		Model:   m.Model.Copy().(Ref),
		Created: m.Created.Copy().(DateTime),
		Updated: m.Updated.Copy().(DateTime),
		Value:   m.Value.Copy(),
	}
}

func (v Meta) Primitive() bool {
	return v.Value.Primitive()
}

type Tuple []Value

func (v Tuple) Transform(f func(Value) Value) Value {
	c := make(Tuple, len(v))
	for i, w := range v {
		c[i] = w.Transform(f)
	}
	return f(c)
}

func (l Tuple) Equals(v Value) bool {
	q, ok := v.(Tuple)
	if !ok {
		return false
	}
	if len(l) != len(q) {
		return false
	}
	for i := 0; i < len(l); i++ {
		if !l[i].Equals(q[i]) {
			return false
		}
	}
	return true
}

func (t Tuple) Copy() Value {
	return t.Transform(TransformIdentity)
}

func (l Tuple) OverMap(f func(int, Value) Value) Tuple {
	for i, v := range l {
		l[i] = f(i, v)
	}
	return l
}

func (v Tuple) Primitive() bool {
	return false
}

type List []Value

func (v List) Transform(f func(Value) Value) Value {
	c := make(List, len(v))
	for i, w := range v {
		c[i] = w.Transform(f)
	}
	return f(c)
}

func (l List) Equals(v Value) bool {
	q, ok := v.(List)
	if !ok {
		return false
	}
	if len(l) != len(q) {
		return false
	}
	for i := 0; i < len(l); i++ {
		if !l[i].Equals(q[i]) {
			return false
		}
	}
	return true
}

func (t List) Copy() Value {
	return t.Transform(TransformIdentity)
}

func (l List) Map(f func(int, Value) Value) List {
	c := make(List, len(l), len(l))
	c.OverMap(f)
	return c
}

// Like Map, but overwrites list elements in-place
func (l List) OverMap(f func(int, Value) Value) List {
	for i, v := range l {
		l[i] = f(i, v)
	}
	return l
}

func (v List) Primitive() bool {
	return false
}

type Union struct {
	Case  string
	Value Value
}

func (v Union) Transform(f func(Value) Value) Value {
	c := Union{Case: v.Case}
	c.Value = v.Value.Transform(f)
	return f(c)
}

func (u Union) Copy() Value {
	return u.Transform(TransformIdentity)
}

func (u Union) Equals(v Value) bool {
	q, ok := v.(Union)
	return ok && u.Case == q.Case && u.Value.Equals(q.Value)
}

func (v Union) Primitive() bool {
	return false
}

type Raw []byte

func (v Raw) Transform(f func(Value) Value) Value {
	return f(v.Copy())
}

func (a Raw) Copy() Value {
	c := make(Raw, len(a), len(a))
	copy(c, a)
	return c
}

func (a Raw) Equals(v Value) bool {
	q, ok := v.(Raw)
	if !ok {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != q[i] {
			return false
		}
	}
	return true
}

func (v Raw) Primitive() bool {
	return false
}

type Struct map[string]Value

func (v Struct) Transform(f func(Value) Value) Value {
	c := make(Struct, len(v))
	for k, w := range v {
		c[k] = w.Transform(f)
	}
	return f(c)
}

func (s Struct) Copy() Value {
	return s.Transform(TransformIdentity)
}

func (s Struct) Equals(v Value) bool {
	q, ok := v.(Struct)
	if !ok {
		return false
	}
	if len(q) != len(s) {
		return false
	}
	for k, v := range s {
		w, ok := q[k]
		if !ok {
			return false
		}
		if !v.Equals(w) {
			return false
		}
	}
	return true
}

func (s Struct) Keys() []string {
	keys := make([]string, 0, len(s))
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (l Struct) OverMap(f func(string, Value) Value) Struct {
	for i, v := range l {
		l[i] = f(i, v)
	}
	return l
}

func (v Struct) Primitive() bool {
	return false
}

type Map map[string]Value

func (v Map) Transform(f func(Value) Value) Value {
	c := make(Map, len(v))
	for k, w := range v {
		c[k] = w.Transform(f)
	}
	return f(c)
}

func (s Map) Copy() Value {
	return s.Transform(TransformIdentity)
}

func (s Map) Equals(v Value) bool {
	q, ok := v.(Map)
	if !ok {
		return false
	}
	if len(q) != len(s) {
		return false
	}
	for k, v := range s {
		w, ok := q[k]
		if !ok {
			return false
		}
		if !v.Equals(w) {
			return false
		}
	}
	return true
}

func (m Map) Keys() []string {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m Map) Map(f func(string, Value) Value) Map {
	c := make(Map, len(m))
	c.OverMap(f)
	return c
}

// Like Map, but overwrites map elements in-place
func (m Map) OverMap(f func(string, Value) Value) Map {
	for k, v := range m {
		m[k] = f(k, v)
	}
	return m
}

func (v Map) Primitive() bool {
	return false
}

type Float float64

func (v Float) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Float) Copy() Value {
	return x
}

func (f Float) Equals(v Value) bool {
	return f == v
}

func (v Float) Primitive() bool {
	return true
}

type Bool bool

func (v Bool) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Bool) Copy() Value {
	return x
}

func (b Bool) Equals(v Value) bool {
	return b == v
}

func (v Bool) Primitive() bool {
	return true
}

type String string

func (v String) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x String) Copy() Value {
	return x
}

func (s String) Equals(v Value) bool {
	q, ok := v.(String)
	return ok && s == q
}

func (s String) String() string {
	return string(s)
}

func (v String) Primitive() bool {
	return true
}

type Ref [2]string

func (v Ref) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Ref) Copy() Value {
	return x
}

func (s Ref) Equals(v Value) bool {
	return s == v
}

func (v Ref) Primitive() bool {
	return true
}

type DateTime struct {
	time.Time
}

func (v DateTime) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x DateTime) Copy() Value {
	return x
}

func (s DateTime) Equals(v Value) bool {
	return s == v
}

func (v DateTime) Primitive() bool {
	return true
}

type Null struct{}

func (v Null) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Null) Copy() Value {
	return x
}

func (s Null) Equals(v Value) bool {
	return s == v
}

func (v Null) Primitive() bool {
	return true
}

type Symbol string

func (v Symbol) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Symbol) Copy() Value {
	return x
}

func (s Symbol) Equals(v Value) bool {
	return s == v
}

func (v Symbol) Primitive() bool {
	return true
}

type Set map[uint64]Value // hash -> value

func (s Set) Keys() []uint64 {
	keys := make([]uint64, 0, len(s))
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (x Set) Transform(f func(Value) Value) Value {
	c := make(Set, len(x))
	for _, v := range x {
		w := v.Transform(f)
		c[Hash(w, nil).Sum64()] = w
	}
	return f(c)
}

func (x Set) Copy() Value {
	return x.Transform(TransformIdentity)
}

func (s Set) Equals(v Value) bool {
	q, ok := v.(Set)
	if !ok {
		return false
	}
	if len(q) != len(s) {
		return false
	}
	for k, _ := range s {
		if _, ok := q[k]; !ok {
			return false
		}
		// value comparison unnecessary because hash identifies object
	}
	return true
}

func (v Set) Primitive() bool {
	return false
}

type Int8 int8

func (v Int8) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Int8) Copy() Value {
	return x
}

func (i Int8) Equals(v Value) bool {
	return i == v
}

func (Int8) Primitive() bool {
	return true
}

type Int16 int16

func (v Int16) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Int16) Copy() Value {
	return x
}

func (i Int16) Equals(v Value) bool {
	return i == v
}

func (Int16) Primitive() bool {
	return true
}

type Int32 int32

func (v Int32) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Int32) Copy() Value {
	return x
}

func (i Int32) Equals(v Value) bool {
	return i == v
}

func (Int32) Primitive() bool {
	return true
}

type Int64 int64

func (v Int64) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Int64) Copy() Value {
	return x
}

func (i Int64) Equals(v Value) bool {
	return i == v
}

func (Int64) Primitive() bool {
	return true
}

type Uint8 uint8

func (v Uint8) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Uint8) Copy() Value {
	return x
}

func (i Uint8) Equals(v Value) bool {
	return i == v
}

func (Uint8) Primitive() bool {
	return true
}

type Uint16 uint16

func (v Uint16) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Uint16) Copy() Value {
	return x
}

func (i Uint16) Equals(v Value) bool {
	return i == v
}

func (Uint16) Primitive() bool {
	return true
}

type Uint32 uint32

func (v Uint32) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Uint32) Copy() Value {
	return x
}

func (i Uint32) Equals(v Value) bool {
	return i == v
}

func (Uint32) Primitive() bool {
	return true
}

type Uint64 uint64

func (v Uint64) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x Uint64) Copy() Value {
	return x
}

func (i Uint64) Equals(v Value) bool {
	return i == v
}

func (Uint64) Primitive() bool {
	return true
}

func TransformIdentity(v Value) Value {
	return v
}
