// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"github.com/karmarun/karma.run/kvm/val"
	"strconv"
	"time"
)

type Recursion struct {
	Label        string
	Model        Model
	traverseFlag bool       // not thread safe
	transPointer *Recursion // not thread safe
}

func (r *Recursion) Transform(f func(Model) Model) Model {
	if r.Model == nil {
		panic("mdl.*Recursion.Transform() called during model building")
	}
	if r.transPointer != nil {
		return r.transPointer
	}
	c := &Recursion{Label: r.Label}
	r.transPointer = c
	c.Model = r.Model.Transform(f)
	r.transPointer = nil
	return f(c)
}

func (r *Recursion) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, r)
	r.Model.TraverseValue(j, f)
}

func (r *Recursion) Traverse(p []string, f func([]string, Model)) {
	if r.traverseFlag {
		return
	}
	r.traverseFlag = true
	f(p, r)
	r.Model.Traverse(append(p, "model"), f)
	r.traverseFlag = false
}

func (r *Recursion) Copy() Model {
	return r.Transform(TransformIdentity)
}

func (r *Recursion) Zero() val.Value {
	return r.Model.Zero()
}

func (m *Recursion) Concrete() Model {
	return m.Model.Concrete()
}

func (m *Recursion) Equals(n Model) bool {
	if q, ok := n.(*Recursion); ok {
		return m == q
	}
	return false
}

type List struct {
	Elements Model
}

func (m List) Transform(f func(Model) Model) Model {
	c := List{}
	c.Elements = m.Elements.Transform(f)
	return f(c)
}

func (l List) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, l)
	if u, ok := j.(val.List); ok {
		for _, w := range u {
			l.Elements.TraverseValue(w, f)
		}
	}
}

func (l List) Copy() Model {
	return l.Transform(TransformIdentity)
}

func (r List) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
	r.Elements.Traverse(append(p, "elements"), f)
}

func (r List) Zero() val.Value {
	return make(val.List, 0, 0)
}

func (m List) Concrete() Model {
	return m
}

func (m List) Equals(n Model) bool {
	if q, ok := n.(List); ok {
		return m.Elements.Equals(q.Elements)
	}
	return false
}

type Map struct {
	Elements Model
}

func (m Map) Transform(f func(Model) Model) Model {
	c := Map{}
	c.Elements = m.Elements.Transform(f)
	return f(c)
}

func (m Map) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, m)
	if u, ok := j.(val.Map); ok {
		for _, w := range u {
			m.Elements.TraverseValue(w, f)
		}
	}
}

func (l Map) Copy() Model {
	return l.Transform(TransformIdentity)
}

func (r Map) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
	r.Elements.Traverse(append(p, "elements"), f)
}

func (r Map) Zero() val.Value {
	return make(val.Map, 0)
}

func (m Map) Concrete() Model {
	return m
}

func (m Map) Equals(n Model) bool {
	if q, ok := n.(Map); ok {
		return m.Elements.Equals(q.Elements)
	}
	return false
}

type String struct{}

func (r String) Zero() val.Value {
	return val.String("")
}

func (m String) Transform(f func(Model) Model) Model {
	return f(String{})
}

func (s String) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, s)
}

func (s String) Copy() Model {
	return s
}

func (r String) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m String) Concrete() Model {
	return m
}

func (m String) Equals(n Model) bool {
	_, ok := n.(String)
	return ok
}

type Float struct{}

func (r Float) Zero() val.Value {
	return val.Float(0)
}

func (m Float) Transform(f func(Model) Model) Model {
	return f(Float{})
}

func (s Float) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, s)
}

func (f Float) Copy() Model {
	return f
}

func (r Float) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Float) Concrete() Model {
	return m
}

func (m Float) Equals(n Model) bool {
	_, ok := n.(Float)
	return ok
}

type Struct map[string]Model

func (r Struct) Zero() val.Value {
	v := make(val.Struct, len(r))
	for i, m := range r {
		v[i] = m.Zero()
	}
	return v
}

func (m Struct) Transform(f func(Model) Model) Model {
	c := make(Struct, len(m))
	for k, v := range m {
		c[k] = v.Transform(f)
	}
	return f(c)
}

func (s Struct) Keys() []string {
	keys := make([]string, 0, len(s))
	for k, _ := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s Struct) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, s)
	if u, ok := j.(val.Struct); ok {
		for k, m := range s {
			if _, ok := u[k]; !ok {
				continue
			}
			m.TraverseValue(u[k], f)
		}
	}
}

func (s Struct) Copy() Model {
	return s.Transform(TransformIdentity)
}

func (r Struct) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
	z := r
	for k, m := range z {
		m.Traverse(append(p, k), f)
	}
}

func (m Struct) Concrete() Model {
	return m
}

func (m Struct) Equals(n Model) bool {
	if q, ok := n.(Struct); ok {
		for k, m := range m {
			n, ok := q[k]
			if !ok || !m.Equals(n) {
				return false
			}
		}
		return true
	}
	return false
}

type Union map[string]Model

func (m Union) Zero() val.Value {
	for k, w := range m {
		if w.Zeroable() {
			return val.Union{k, w.Zero()}
		}
	}
	panic("never reached")
}

func (m Union) Transform(f func(Model) Model) Model {
	c := make(Union, len(m))
	for k, v := range m {
		c[k] = v.Transform(f)
	}
	return f(c)
}

func (s Union) Cases() []string {
	cases := make([]string, 0, len(s))
	for k, _ := range s {
		cases = append(cases, k)
	}
	return cases
}

func (s Union) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, s)
	if u, ok := j.(val.Union); ok {
		if m, ok := s[u.Case]; ok {
			m.TraverseValue(u.Value, f)
		}
	}
}

func (s Union) Copy() Model {
	return s.Transform(TransformIdentity)
}

func (r Union) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
	z := r
	for k, m := range z {
		m.Traverse(append(p, k), f)
	}
}

func (m Union) Concrete() Model {
	return m
}

func (m Union) Equals(n Model) bool {
	if q, ok := n.(Union); ok {
		for k, m := range m {
			n, ok := q[k]
			if !ok || !m.Equals(n) {
				return false
			}
		}
		return true
	}
	return false
}

type Bool struct{}

func (r Bool) Zero() val.Value {
	return val.Bool(false)
}

func (m Bool) Transform(f func(Model) Model) Model {
	return f(Bool{})
}

func (b Bool) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, b)
}

func (m Bool) Copy() Model {
	return m
}

func (r Bool) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Bool) Concrete() Model {
	return m
}

func (m Bool) Equals(n Model) bool {
	_, ok := n.(Bool)
	return ok
}

// Any behaves like a top type
type Any struct{}

func (r Any) Zero() val.Value {
	panic("zero called on mdl.Any")
}

func (m Any) Transform(f func(Model) Model) Model {
	return f(Any{})
}

func (a Any) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, a)
}

func (m Any) Copy() Model {
	return m
}

func (r Any) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Any) Concrete() Model {
	return m
}

func (m Any) Equals(n Model) bool {
	_, ok := n.(Any)
	return ok
}

type Ref struct {
	Model string
}

func (r Ref) Zero() val.Value {
	panic("Zero called on mdl.Ref")
}

func (m Ref) Transform(f func(Model) Model) Model {
	return f(Ref{m.Model})
}

func (r Ref) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, r)
}

func (r Ref) Copy() Model {
	return r
}

func (r Ref) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m Ref) Concrete() Model {
	return m
}

func (m Ref) Equals(n Model) bool {
	if q, ok := n.(Ref); ok {
		return q.Model == m.Model
	}
	return false
}

type Unique struct {
	Model Model
}

func (r Unique) Zero() val.Value {
	return r.Model.Zero()
}

func (m Unique) Transform(f func(Model) Model) Model {
	c := Unique{}
	c.Model = m.Model.Transform(f)
	return f(c)
}

func (o Unique) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, o)
	o.Model.TraverseValue(j, f)
}

func (o Unique) Copy() Model {
	return o.Transform(TransformIdentity)
}

func (r Unique) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
	r.Model.Traverse(append(p, "model"), f)
}

func (m Unique) Concrete() Model {
	return m.Model.Concrete()
}

func (m Unique) Equals(n Model) bool {
	if q, ok := n.(Unique); ok {
		return m.Model.Equals(q.Model)
	}
	return false
}

type DateTime struct{}

func (r DateTime) Zero() val.Value {
	return val.DateTime{time.Time{}}
}

func (m DateTime) Transform(f func(Model) Model) Model {
	return f(DateTime{})
}

func (o DateTime) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, o)
}

func (o DateTime) Copy() Model {
	return o
}

func (r DateTime) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (m DateTime) Concrete() Model {
	return m
}

func (m DateTime) Equals(n Model) bool {
	_, ok := n.(DateTime)
	return ok
}

type Tuple []Model

func (r Tuple) Zero() val.Value {
	v := make(val.Tuple, len(r), len(r))
	for i, m := range r {
		v[i] = m.Zero()
	}
	return v
}

func (m Tuple) Transform(f func(Model) Model) Model {
	c := make(Tuple, len(m), len(m))
	for k, v := range m {
		c[k] = v.Transform(f)
	}
	return f(c)
}

func (s Tuple) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, s)
	if u, ok := j.(val.Tuple); ok {
		for i, l := 0, minInt(len(s), len(u)); i < l; i++ {
			s[i].TraverseValue(u[i], f)
		}
	}
}

func (s Tuple) Copy() Model {
	return s.Transform(TransformIdentity)
}

func (s Tuple) Traverse(p []string, f func([]string, Model)) {
	f(p, s)
	for i, m := range s {
		m.Traverse(append(p, strconv.Itoa(i)), f)
	}
}

func (m Tuple) Concrete() Model {
	return m
}

func (m Tuple) Equals(n Model) bool {
	if q, ok := n.(Tuple); ok {
		if len(m) != len(q) {
			return false
		}
		for i, m := range m {
			if !m.Equals(q[i]) {
				return false
			}
		}
		return true
	}
	return false
}

type Annotation struct {
	Value string
	Model Model
}

func (r Annotation) Zero() val.Value {
	return r.Model.Zero()
}

func (m Annotation) Transform(f func(Model) Model) Model {
	c := Annotation{Value: m.Value}
	c.Model = m.Model.Transform(f)
	return f(c)
}

func (w Annotation) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, w)
	w.Model.TraverseValue(j, f)
}

func (m Annotation) Copy() Model {
	return m.Transform(TransformIdentity)
}

func (w Annotation) Traverse(p []string, f func([]string, Model)) {
	f(p, w)
	w.Model.Traverse(append(p, "model"), f)
}

func (m Annotation) Concrete() Model {
	return m.Model
}

func (m Annotation) Equals(n Model) bool {
	if q, ok := n.(Annotation); ok {
		return m.Model.Equals(q.Model)
	}
	return false
}

type Or [2]Model

func (m Or) Transform(f func(Model) Model) Model {

	return Or{m[0].Transform(f), m[1].Transform(f)}

	// note: unable to call Either() here, leads to bugs
	//       due to incomplete *Recursion values in transition
	//       from being a submodel-less transPointer to becoming
	//       the return value of Transform

}

func (w Or) TraverseValue(j val.Value, f func(val.Value, Model)) {
	w[0].TraverseValue(j, f)
	w[1].TraverseValue(j, f)
}

func (m Or) Copy() Model {
	return m.Transform(TransformIdentity)
}

func (w Or) Traverse(p []string, f func([]string, Model)) {
	f(p, w)
	w[0].Traverse(append(p, "0"), f)
	w[1].Traverse(append(p, "1"), f)
}

func (r Or) Zero() val.Value {
	if r[0].Zeroable() {
		return r[0].Zero()
	}
	return r[1].Zero()
}

func (m Or) Concrete() Model {
	return m
}

func (m Or) Equals(n Model) bool {
	if q, ok := n.(Or); ok {
		return (m[0].Equals(q[0]) && m[1].Equals(q[1])) || (m[0].Equals(q[1]) && m[1].Equals(q[0]))
	}
	return false
}

type Enum map[string]struct{}

func (m Enum) Zero() val.Value {
	k := ""
	for a, _ := range m {
		k = a
		break
	}
	return val.Symbol(k)
}

func (m Enum) Transform(f func(Model) Model) Model {
	c := make(Enum, len(m))
	for i, s := range m {
		c[i] = s
	}
	return f(c)
}

func (m Enum) TraverseValue(v val.Value, f func(val.Value, Model)) {
	f(v, m)
}

func (s Enum) Copy() Model {
	return s.Transform(TransformIdentity)
}

func (s Enum) Traverse(p []string, f func([]string, Model)) {
	f(p, s)
}

func (m Enum) Concrete() Model {
	return m
}

func (m Enum) Equals(n Model) bool {
	if q, ok := n.(Enum); ok {
		if len(m) != len(q) {
			return false
		}
		for s, _ := range m {
			if _, ok := q[s]; !ok {
				return false
			}
		}
		return true
	}
	return false
}

type Set struct {
	Elements Model
}

func (m Set) Transform(f func(Model) Model) Model {
	c := Set{}
	c.Elements = m.Elements.Transform(f)
	return f(c)
}

func (l Set) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, l)
	if u, ok := j.(val.Set); ok {
		for _, w := range u {
			l.Elements.TraverseValue(w, f)
		}
	}
}

func (l Set) Copy() Model {
	return l.Transform(TransformIdentity)
}

func (r Set) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
	r.Elements.Traverse(append(p, "elements"), f)
}

func (r Set) Zero() val.Value {
	return make(val.Set, 0)
}

func (m Set) Concrete() Model {
	return m
}

func (m Set) Equals(n Model) bool {
	if q, ok := n.(Set); ok {
		return m.Elements.Equals(q.Elements)
	}
	return false
}

type Null struct{}

func (m Null) Transform(f func(Model) Model) Model {
	return f(Null{})
}

func (l Null) TraverseValue(j val.Value, f func(val.Value, Model)) {
	f(j, l)
}

func (l Null) Copy() Model {
	return l
}

func (r Null) Traverse(p []string, f func([]string, Model)) {
	f(p, r)
}

func (r Null) Zero() val.Value {
	return val.Null{}
}

func (m Null) Concrete() Model {
	return m
}

func (m Null) Equals(n Model) bool {
	return n == m
}

func (r *Recursion) Nullable() bool {
	return r.Model.Nullable() // TODO use recursion lock
}

func (Any) Nullable() bool {
	return false
}

func (o Or) Nullable() bool {
	return o[0].Nullable() || o[1].Nullable()
}

func (u Unique) Nullable() bool {
	return u.Model.Nullable()
}

func (a Annotation) Nullable() bool {
	return a.Model.Nullable()
}

func (Tuple) Nullable() bool {
	return false
}

func (Enum) Nullable() bool {
	return false
}

func (List) Nullable() bool {
	return false
}

func (Map) Nullable() bool {
	return false
}

func (Union) Nullable() bool {
	return false
}

func (Struct) Nullable() bool {
	return false
}

func (Set) Nullable() bool {
	return false
}

func (String) Nullable() bool {
	return false
}

func (Null) Nullable() bool {
	return true
}

func (Float) Nullable() bool {
	return false
}

func (DateTime) Nullable() bool {
	return false
}

func (Bool) Nullable() bool {
	return false
}

func (Ref) Nullable() bool {
	return false
}

func (r *Recursion) Zeroable() bool {
	return r.Model.Zeroable() // TODO use recursion lock
}

func (Any) Zeroable() bool {
	return false
}

func (o Or) Zeroable() bool {
	return o[0].Zeroable() || o[1].Zeroable()
}

func (u Unique) Zeroable() bool {
	return u.Model.Zeroable()
}

func (a Annotation) Zeroable() bool {
	return a.Model.Zeroable()
}

func (m Tuple) Zeroable() bool {
	for _, w := range m {
		if !w.Zeroable() {
			return false
		}
	}
	return true
}

func (Enum) Zeroable() bool {
	return true
}

func (List) Zeroable() bool {
	return true // empty case
}

func (Map) Zeroable() bool {
	return true // empty case
}

func (Set) Zeroable() bool {
	return true // empty case
}

func (m Union) Zeroable() bool {
	for _, w := range m {
		if w.Zeroable() {
			return true
		}
	}
	return false
}

func (m Struct) Zeroable() bool {
	for _, w := range m {
		if !w.Zeroable() {
			return false
		}
	}
	return true
}

func (String) Zeroable() bool {
	return true
}

func (Null) Zeroable() bool {
	return true
}

func (Float) Zeroable() bool {
	return true
}

func (DateTime) Zeroable() bool {
	return true
}

func (Bool) Zeroable() bool {
	return true
}

func (Ref) Zeroable() bool {
	return false
}

func (m *Recursion) Unwrap() Model {
	return m
}

func (m Any) Unwrap() Model {
	return m
}

func (m Or) Unwrap() Model {
	return m
}

func (m Unique) Unwrap() Model {
	return m
}

func (m Annotation) Unwrap() Model {
	return m
}

func (m Tuple) Unwrap() Model {
	return m
}

func (m Enum) Unwrap() Model {
	return m
}

func (m List) Unwrap() Model {
	return m
}

func (m Map) Unwrap() Model {
	return m
}

func (m Set) Unwrap() Model {
	return m
}

func (m Union) Unwrap() Model {
	return m
}

func (m Struct) Unwrap() Model {
	return m
}

func (m String) Unwrap() Model {
	return m
}

func (m Null) Unwrap() Model {
	return m
}

func (m Float) Unwrap() Model {
	return m
}

func (m DateTime) Unwrap() Model {
	return m
}

func (m Bool) Unwrap() Model {
	return m
}

func (m Ref) Unwrap() Model {
	return m
}
