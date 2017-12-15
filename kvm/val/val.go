// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package val

import (
	"time"
)

//go:generate go run ../../generate/logmap/main.go --package val --key string --value Value --output logmap_generated.go

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

// Meta.Transform does NOT call f on itself because
// after a transformation, the object is not
// the same as persisted.
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
	for i, w := range v {
		v[i] = w.Transform(f)
	}
	return f(v)
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

func (v Tuple) Copy() Value {
	c := make(Tuple, len(v), len(v))
	for i, w := range v {
		c[i] = w.Copy()
	}
	return c
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
	for i, w := range v {
		v[i] = w.Transform(f)
	}
	return f(v)
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

func (v List) Copy() Value {
	c := make(List, len(v), len(v))
	for i, w := range v {
		c[i] = w.Copy()
	}
	return c
}

func (l List) Map(f func(int, Value) Value) List {
	return l.Copy().(List).OverMap(f)
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
	return f(Union{v.Case, v.Value.Transform(f)})
}

func (v Union) Copy() Value {
	return Union{v.Case, v.Value.Copy()}
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
	return f(v)
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

type Struct struct{ lm *logMapStringValue }

func NewStruct(capacity int) Struct {
	return Struct{newlogMapStringValue(capacity)}
}

func StructFromMap(m map[string]Value) Struct {
	v := NewStruct(len(m))
	for k, w := range m {
		v.Set(k, w)
	}
	return v
}

func (v Struct) Len() int {
	return v.lm.len()
}

func (v Struct) Field(k string) Value {
	w, ok := v.lm.get(k)
	if !ok {
		return nil
	}
	return w
}

func (v Struct) Get(k string) (Value, bool) {
	return v.lm.get(k)
}

func (v *Struct) Set(k string, w Value) {
	v.lm.set(k, w)
}

func (v *Struct) Delete(k string) {
	v.lm.unset(k)
}

func (v Struct) Transform(f func(Value) Value) Value {
	v.lm.overMap(func(k string, v Value) Value {
		return v.Transform(f)
	})
	return f(v)
}

func (v Struct) ForEach(f func(string, Value) bool) {
	v.lm.forEach(f)
}

func (v Struct) Copy() Value {
	c := v.lm.copy()
	c.overMap(func(k string, v Value) Value {
		return v.Copy()
	})
	return Struct{c}
}

func (v Struct) Equals(w Value) bool {
	x, ok := w.(Struct)
	if !ok {
		return false
	}
	if !v.lm.sameKeys(x.lm) {
		return false
	}
	eq := true
	v.lm.forEach(func(k string, v Value) bool {
		w, _ := x.lm.get(k)
		eq = v.Equals(w)
		return eq
	})
	return eq
}

func (v Struct) Keys() []string {
	return v.lm.keys()
}

func (v Struct) Map(f func(string, Value) Value) Struct {
	c := v.Copy().(Struct)
	c.lm.overMap(f)
	return c
}

func (v Struct) OverMap(f func(string, Value) Value) {
	v.lm.overMap(f)
}

func (v Struct) Primitive() bool {
	return false
}

type Map struct{ lm *logMapStringValue }

func NewMap(capacity int) Map {
	return Map{newlogMapStringValue(capacity)}
}

func MapFromMap(m map[string]Value) Map {
	v := NewMap(len(m))
	for k, w := range m {
		v.Set(k, w)
	}
	return v
}

func (v Map) Len() int {
	return v.lm.len()
}

func (v Map) Key(k string) Value {
	w, ok := v.lm.get(k)
	if !ok {
		return nil
	}
	return w
}

func (v Map) Get(k string) (Value, bool) {
	return v.lm.get(k)
}

func (v *Map) Set(k string, w Value) {
	v.lm.set(k, w)
}

func (v *Map) Delete(k string) {
	v.lm.unset(k)
}

func (v Map) Transform(f func(Value) Value) Value {
	v.lm.overMap(func(k string, v Value) Value {
		return v.Transform(f)
	})
	return f(v)
}

func (v Map) ForEach(f func(string, Value) bool) {
	v.lm.forEach(f)
}

func (v Map) Copy() Value {
	c := v.lm.copy()
	c.overMap(func(k string, v Value) Value {
		return v.Copy()
	})
	return Map{c}
}

func (v Map) Equals(w Value) bool {
	x, ok := w.(Map)
	if !ok {
		return false
	}
	if !v.lm.sameKeys(x.lm) {
		return false
	}
	eq := true
	v.lm.forEach(func(k string, v Value) bool {
		w, _ := x.lm.get(k)
		eq = v.Equals(w)
		return eq
	})
	return eq
}

func (v Map) Keys() []string {
	return v.lm.keys()
}

func (v Map) Map(f func(string, Value) Value) Map {
	c := v.Copy().(Map)
	c.lm.overMap(f)
	return c
}

func (v Map) OverMap(f func(string, Value) Value) {
	v.lm.overMap(f)
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

var Null = null{}

type null struct{}

func (v null) Transform(f func(Value) Value) Value {
	return f(v)
}

func (x null) Copy() Value {
	return x
}

func (s null) Equals(v Value) bool {
	return s == v
}

func (v null) Primitive() bool {
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
	for h, v := range x {
		w := v.Transform(f)
		delete(x, h)
		h = Hash(w, nil).Sum64()
		x[h] = w
	}
	return f(x)
}

func (x Set) Copy() Value {
	c := make(Set, len(x))
	for k, v := range x {
		c[k] = v.Copy()
	}
	return c
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
