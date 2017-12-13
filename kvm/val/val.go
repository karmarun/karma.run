// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package val

import (
	"sort"
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

type Struct struct{ valueMap }

func NewStruct(capacity int) Struct {
	return Struct{newValueMap(capacity)}
}

func StructFromMap(m map[string]Value) Struct {
	v := NewStruct(len(m))
	for k, w := range m {
		v.Set(k, w)
	}
	return v
}

func (v Struct) Len() int {
	return v.valueMap.Len()
}

func (v Struct) Field(k string) Value {
	return v.valueMap.Get(k)
}

func (v Struct) Get(k string) (Value, bool) {
	w := v.valueMap.Get(k)
	if w == nil {
		return nil, false
	}
	return w, true
}

func (v *Struct) Set(k string, w Value) {
	v.valueMap = v.valueMap.Set(k, w)
}

func (v *Struct) Delete(k string) {
	v.valueMap = v.valueMap.Unset(k)
}

func (v Struct) Transform(f func(Value) Value) Value {
	v.valueMap.OverMap(func(k string, v Value) Value {
		return f(v)
	})
	return f(v)
}

func (v Struct) ForEach(f func(string, Value) bool) {
	v.valueMap.ForEach(f)
}

func (v Struct) Copy() Value {
	return Struct{v.valueMap.Copy()}
}

func (v Struct) Equals(w Value) bool {
	x, ok := w.(Struct)
	if !ok {
		return false
	}
	return v.valueMap.Equals(x.valueMap)
}

func (v Struct) Keys() []string {
	return v.valueMap.Keys()
}

func (v Struct) Map(f func(string, Value) Value) Struct {
	c := v.valueMap.Copy()
	c.OverMap(f)
	return Struct{c}
}

func (v Struct) OverMap(f func(string, Value) Value) Struct {
	v.valueMap.OverMap(f)
	return v
}

func (v Struct) Primitive() bool {
	return false
}

type Map map[string]Value

func (v Map) Transform(f func(Value) Value) Value {
	for k, w := range v {
		v[k] = w.Transform(f)
	}
	return f(v)
}

func (v Map) Copy() Value {
	c := make(Map, len(v))
	for k, w := range v {
		c[k] = w.Copy()
	}
	return c
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

func (v Map) Map(f func(string, Value) Value) Map {
	return v.Copy().(Map).OverMap(f)
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

type valueMap struct {
	keys   []string
	values []Value
}

func newValueMap(capacity int) valueMap {
	return valueMap{
		keys:   make([]string, 0, capacity),
		values: make([]Value, 0, capacity),
	}
}

func (m valueMap) Equals(w valueMap) bool {
	if len(m.keys) != len(w.keys) {
		return false
	}
	for i, _ := range m.keys {
		if m.keys[i] != w.keys[i] {
			return false
		}
	}
	for i, _ := range m.values {
		if !m.values[i].Equals(w.values[i]) {
			return false
		}
	}
	return true
}

func (m valueMap) Keys() []string {
	keys := make([]string, len(m.keys), len(m.keys))
	copy(keys, m.keys)
	return keys
}

func (m valueMap) OverMap(f func(k string, v Value) Value) {
	for i, k := range m.keys {
		m.values[i] = f(k, m.values[i])
	}
}

func (m valueMap) Get(k string) Value {
	i := m.search(k)
	if i == len(m.keys) || m.keys[i] != k {
		return nil
	}
	return m.values[i]
}

func (m valueMap) Set(k string, v Value) valueMap {
	i := m.search(k)
	m.keys, m.values = append(m.keys, ""), append(m.values, nil)
	copy(m.keys[i+1:], m.keys[i:])
	copy(m.values[i+1:], m.values[i:])
	m.keys[i], m.values[i] = k, v
	return m
}

func (m valueMap) Unset(k string) valueMap {
	i := m.search(k)
	if i == len(m.keys) || m.keys[i] != k {
		return m
	}
	l := len(m.keys)
	copy(m.keys[i:l-1], m.keys[i+1:])
	copy(m.values[i:l-1], m.values[i+1:])
	m.keys[l-1], m.values[l-1] = "", nil // let them be GC'ed
	m.keys, m.values = m.keys[:l-1], m.values[:l-1]
	return m
}

func (m valueMap) Copy() valueMap {
	keys, values := ([]string)(nil), ([]Value)(nil)
	if m.keys != nil {
		keys = make([]string, len(m.keys), cap(m.keys))
		copy(keys, m.keys)
	}
	if m.values != nil {
		values = make([]Value, len(m.values), cap(m.values))
		copy(values, m.values)
	}
	return valueMap{keys, values}
}

func (m valueMap) ForEach(f func(string, Value) bool) {
	for i, k := range m.keys {
		if !f(k, m.values[i]) {
			break
		}
	}
}

func (m valueMap) Len() int {
	return len(m.keys)
}

// The return value is the index to insert k
// if k is not present (it can be len(m.keys)).
func (m valueMap) search(k string) int {
	return sort.SearchStrings(m.keys, k)
}
