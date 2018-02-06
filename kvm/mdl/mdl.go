// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"fmt"
	"karma.run/kvm/err"
	"karma.run/kvm/val"
	"time"
)

// Model represents a possibly nested and/or recursive data type.
type Model interface {

	// Copy copies a Model tree. It's equivalent to calling
	// Transform(mdl.TransformationIdentity)
	Copy() Model

	// Transform traverses a Model tree and returns the result
	// of mapping each of its nodes through function f.
	Transform(f func(Model) Model) Model

	// Zero returns the zero-value for the current model.
	// This is not possible with all model configurations,
	// especially not Refs. See Zeroable.
	Zero() val.Value

	// Zeroable reports wether calling Zero() on this Model is legal.
	Zeroable() bool

	// Unwrap returns the shallowest Model defined in
	// this package in a Model tree.
	Unwrap() Model

	// Concrete unwraps a Model to the shallowest Model that
	// affects its Value representation. In other words,
	// it removes semantic wrappers.
	Concrete() Model

	// Equals reports wether a Model tree equals another.
	Equals(Model) bool

	// Nullable reports wether a Model can be Null.
	Nullable() bool

	// TraverseValue traverses a Value tree v, calling function f
	// for every legal Value/Model combination therein.
	TraverseValue(v val.Value, f func(val.Value, Model))

	// Traverse is obsolete but still needed. Do not use it.
	Traverse([]string, func([]string, Model))

	// Returns the top-level Type of the Values in the set
	// of a Model. Returns TypeInvalid in case of Any and Or.
	ValueType() val.Type
}

const (
	FormatDateTime = time.RFC3339
)

// ValueFromModel returns a Value representation of Model model.
func ValueFromModel(metaId string, model Model, recursions map[*Recursion]struct{}) val.Value {

	if recursions == nil {
		recursions = make(map[*Recursion]struct{})
	}

	// catch top-level recursions and, if we have more than one recursion
	// involved in the model, use "recursive" constructor instead of "recursion"
	if top, ok := model.(*Recursion); ok && len(recursions) == 0 {
		recs := make(map[*Recursion]struct{})
		model.Traverse(nil, func(_ []string, m Model) {
			if r, ok := m.(*Recursion); ok {
				recs[r] = struct{}{}
			}
		})
		if len(recs) > 1 {
			mp := val.NewMap(len(recs))
			for r, _ := range recs {
				mp.Set(r.Label, ValueFromModel(metaId, r.Model, recs))
			}
			return val.Union{"recursive", val.StructFromMap(map[string]val.Value{
				"top":    val.String(top.Label),
				"models": mp,
			})}
		}
	}

	switch m := model.(type) {

	case *Recursion:
		if _, ok := recursions[m]; ok {
			return val.Union{"recurse", val.String(m.Label)}
		}
		recursions[m] = struct{}{}
		o := val.StructFromMap(map[string]val.Value{
			"label": val.String(m.Label),
			"model": ValueFromModel(metaId, m.Model, recursions),
		})
		delete(recursions, m)
		return val.Union{"recursion", o}

	case Optional:
		return val.Union{"optional", ValueFromModel(metaId, m.Model, recursions)}

	case Unique:
		return val.Union{"unique", ValueFromModel(metaId, m.Model, recursions)}

	case Annotation:
		return val.Union{
			"annotation", val.StructFromMap(map[string]val.Value{
				"value": val.String(m.Value),
				"model": ValueFromModel(metaId, m.Model, recursions),
			}),
		}

	case Tuple:
		z := m
		sm := make(val.List, len(z), len(z))
		for i, m := range z {
			sm[i] = ValueFromModel(metaId, m, recursions)
		}
		return val.Union{"tuple", sm}

	case Enum:
		z := m
		ss := make(val.Set, len(z))
		for s, _ := range z {
			v := val.String(s)
			ss[val.Hash(v, nil).Sum64()] = v
		}
		return val.Union{"enum", ss}

	case List:
		return val.Union{"list", ValueFromModel(metaId, m.Elements, recursions)}

	case Set:
		return val.Union{"set", ValueFromModel(metaId, m.Elements, recursions)}

	case Map:
		return val.Union{"map", ValueFromModel(metaId, m.Elements, recursions)}

	case String:
		return val.Union{"string", val.Struct{}}

	case Struct:
		o := val.NewMap(m.Len())
		m.ForEach(func(k string, m Model) bool {
			o.Set(k, ValueFromModel(metaId, m, recursions))
			return true
		})
		return val.Union{"struct", o}

	case Union:
		o := val.NewMap(m.Len())
		m.ForEach(func(k string, v Model) bool {
			o.Set(k, ValueFromModel(metaId, v, recursions))
			return true
		})
		return val.Union{"union", o}

	case Null:
		return val.Union{"null", val.Struct{}}

	case Float:
		return val.Union{"float", val.Struct{}}

	case Int8:
		return val.Union{"int8", val.Struct{}}

	case Int16:
		return val.Union{"int16", val.Struct{}}

	case Int32:
		return val.Union{"int32", val.Struct{}}

	case Int64:
		return val.Union{"int64", val.Struct{}}

	case Uint8:
		return val.Union{"uint8", val.Struct{}}

	case Uint16:
		return val.Union{"uint16", val.Struct{}}

	case Uint32:
		return val.Union{"uint32", val.Struct{}}

	case Uint64:
		return val.Union{"uint64", val.Struct{}}

	case DateTime:
		return val.Union{"dateTime", val.Struct{}}

	case Bool:
		return val.Union{"bool", val.Struct{}}

	case Any:
		return val.Union{"any", val.Struct{}}

	case Ref:
		return val.Union{"ref", val.Ref{metaId, m.Model}}

	}
	panic(fmt.Sprintf(`Unhandled model: %T`, model))
}

// ModelFromValue returns a Model from its Value representation.
func ModelFromValue(metaId string, u val.Union, recursions map[string]*Recursion) (Model, err.PathedError) {

	// note: ModelFromValue may NOT call Copy or Transform on models.
	if recursions == nil {
		recursions = make(map[string]*Recursion)
	}
	switch u.Case {
	case "recursive":
		s := u.Value.(val.Struct)
		t := string(s.Field("top").(val.String))
		a := s.Field("models").(val.Map)
		e := (err.PathedError)(nil)
		a.ForEach(func(l string, _ val.Value) bool {
			if _, ok := recursions[l]; ok {
				e = err.ModelParsingError{
					fmt.Sprintf(`recursion label already defined: %s`, l), u, nil,
				}
				return false
			}
			recursions[l] = NewRecursion(l)
			return true
		})
		if e != nil {
			return nil, e
		}
		if _, ok := recursions[t]; !ok {
			return nil, err.ModelParsingError{fmt.Sprintf(`no definition for top label: %s`, t), u, nil}
		}
		a.ForEach(func(l string, v val.Value) bool {
			m, e_ := ModelFromValue(metaId, v.(val.Union), recursions)
			if e_ != nil {
				e = e_.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementStructField("models"), err.ErrorPathElementMapKey(l))
				return false
			}
			recursions[l].Model = m
			return true
		})
		if e != nil {
			return nil, e
		}
		r := recursions[t]
		a.ForEach(func(l string, v val.Value) bool {
			m := recursions[l]
			if couldRecurseInfinitely(m, nil) {
				e = err.ModelParsingError{`infinite recursion`, u, nil}
				return false
			}
			delete(recursions, l)
			return true
		})
		if e != nil {
			return nil, e
		}
		return r, nil

	case "annotation":
		v := u.Value.(val.Struct)
		a := string(v.Field("value").(val.String))
		m, e := ModelFromValue(metaId, v.Field("model").(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementStructField("model"))
		}
		return Annotation{Model: m, Value: a}, nil

	case "recursion":
		v := u.Value.(val.Struct)
		l := string(v.Field("label").(val.String))
		r := NewRecursion(l)
		if _, ok := recursions[l]; ok {
			return nil, err.ModelParsingError{fmt.Sprintf(`recursion label already defined: %s`, l), u, nil}
		}
		recursions[l] = r
		m, e := ModelFromValue(metaId, v.Field("model").(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementStructField("model"))
		}
		if couldRecurseInfinitely(m, nil) {
			return nil, err.ModelParsingError{`infinite recursion`, u, nil}
		}
		r.Model = m
		delete(recursions, l)
		return r, nil

	case "recurse":
		l := string(u.Value.(val.String))
		if _, ok := recursions[l]; !ok {
			return nil, err.ModelParsingError{fmt.Sprintf(`undefined recursion label: %s`, l), u, nil}
		}
		return recursions[l], nil

	case "enum":
		ss := u.Value.(val.Set)
		if len(ss) == 0 {
			return nil, err.ModelParsingError{`empty enum definition`, u, nil}
		}
		m := make(Enum, len(ss))
		for _, v := range ss {
			s := string(v.(val.String))
			if _, ok := m[s]; ok {
				return nil, err.ModelParsingError{fmt.Sprintf(`duplicated symbol: %s`, s), u, nil}
			}
			m[s] = struct{}{}
		}
		return m, nil

	case "set":
		w, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return Set{Elements: w}, nil

	case "list":
		w, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return List{Elements: w}, nil

	case "map":
		w, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return Map{Elements: w}, nil

	case "tuple":
		v := u.Value.(val.List)
		m := make(Tuple, len(v), len(v))
		for i, v := range v {
			w, e := ModelFromValue(metaId, v.(val.Union), recursions)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementListIndex(i))
			}
			m[i] = w
		}
		return m, nil

	case "struct":
		v := u.Value.(val.Map)
		m := NewStruct(v.Len())
		e := (err.PathedError)(nil)
		v.ForEach(func(k string, v val.Value) bool {
			w, e_ := ModelFromValue(metaId, v.(val.Union), recursions)
			if e_ != nil {
				e = e_.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementMapKey(k))
				return false
			}
			m.Set(k, w)
			return true
		})
		if e != nil {
			return nil, e
		}
		return m, nil

	case "union":
		v := u.Value.(val.Map)
		m := NewUnion(v.Len())
		e := (err.PathedError)(nil)
		v.ForEach(func(k string, v val.Value) bool {
			w, e_ := ModelFromValue(metaId, v.(val.Union), recursions)
			if e_ != nil {
				e = e_.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementMapKey(k))
				return false
			}
			m.Set(k, w)
			return true
		})
		if e != nil {
			return nil, e
		}
		return m, nil

	case "optional":
		m, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		if _, ok := m.(Optional); ok {
			return nil, err.ModelParsingError{`nested optional`, u, nil} // TODO: make this check more robust
		}
		return Optional{m}, nil

	case "unique":
		m, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		if _, ok := m.(Unique); ok {
			return nil, err.ModelParsingError{`nested unique`, u, nil} // TODO: make this check more robust
		}
		return Unique{Model: m}, nil

	case "null":
		return Null{}, nil

	case "float":
		return Float{}, nil

	case "uint":
		return Uint64{}, nil

	case "int8":
		return Int8{}, nil

	case "int16":
		return Int16{}, nil

	case "int32":
		return Int32{}, nil

	case "int64":
		return Int64{}, nil

	case "uint8":
		return Uint8{}, nil

	case "uint16":
		return Uint16{}, nil

	case "uint32":
		return Uint32{}, nil

	case "uint64":
		return Uint64{}, nil

	case "int":
		return Int64{}, nil

	case "string":
		return String{}, nil

	case "dateTime":
		return DateTime{}, nil

	case "bool":
		return Bool{}, nil

	case "ref":
		return Ref{string(u.Value.(val.Ref)[1])}, nil
	}
	panic(fmt.Sprintf(`undefined model constructor: %s`, u.Case))
}

var any = Any{}

// Either returns a (hopefully) minimal Or-combination for two given Model arguments.
func Either(l, r Model, m map[*Recursion]*Recursion) Model {

	if l == nil || r == nil {
		panic("mdl.Either called with nil")
	}

	l, r = l.Concrete(), r.Concrete() // discard semantic wrappers

	if l == any || r == any {
		return any
	}

	if l.Equals(r) {
		return l
	}

	if _, ok := l.(Ref); ok {
		if _, ok := r.(Ref); ok {
			return Ref{} // any ref
		}
	}

	{ // optional<L> | R = optional<L|R>
		if lo, ok := l.(Optional); ok {
			ro, ok := r.(Optional)
			if !ok {
				return NewOptional(Either(lo.Model, r, m))
			}
			return NewOptional(Either(lo.Model, ro.Model, m))
		}
		if ro, ok := r.(Optional); ok {
			return NewOptional(Either(l, ro.Model, m))
		}
	}

	{ // L | null = optional<L>
		if _, ok := l.(Null); ok {
			return Optional{r}
		}
		if _, ok := r.(Null); ok {
			return Optional{l}
		}
	}

	switch l := l.(type) {

	case List:
		r, ok := r.(List)
		if !ok {
			return any
		}
		return List{Either(l.Elements, r.Elements, m)}

	case Map:
		r, ok := r.(Map)
		if !ok {
			return any
		}
		return Map{Either(l.Elements, r.Elements, m)}

	case Set:
		r, ok := r.(Set)
		if !ok {
			return any
		}
		return Set{Either(l.Elements, r.Elements, m)}

	case Union:
		r, ok := r.(Union)
		if !ok {
			return any
		}
		u := NewUnion(l.Len() + r.Len())
		l.ForEach(func(k string, w Model) bool {
			u.Set(k, w)
			return true
		})
		r.ForEach(func(k string, w Model) bool {
			q, ok := u.Get(k)
			if !ok {
				u.Set(k, w)
				return true
			}
			u.Set(k, Either(q, w, m))
			return true
		})
		return u

	case Enum:
		r, ok := r.(Enum)
		if !ok {
			return any
		}
		e := make(Enum, minInt(len(l), len(r)))
		for k, _ := range l {
			if _, ok := r[k]; ok {
				e[k] = struct{}{}
			}
		}
		return e

	case Struct:
		r, ok := r.(Struct)
		if !ok {
			return any
		}
		s := NewStruct(l.Len() + r.Len())
		l.ForEach(func(k string, w Model) bool {
			q, ok := r.Get(k)
			if !ok {
				return true
			}
			s.Set(k, Either(q, w, m))
			return true
		})
		return s

	case Tuple:
		r, ok := r.(Tuple)
		if !ok {
			return any
		}
		ln := minInt(len(l), len(r))
		t := make(Tuple, ln, ln)
		longer, shorter := l, r
		if len(l) == ln {
			longer, shorter = r, l
		}
		for i, w := range shorter {
			t[i] = Either(w, longer[i], m)
		}
		return t

	}

	return any
}

func labelUnion(a, b string) string {
	if a == b {
		return a
	}
	return a + " | " + b
}

func couldRecurseInfinitely(m Model, seen map[*Recursion]struct{}) bool {
	switch m := m.(type) {
	case *Recursion:
		if seen == nil {
			seen = make(map[*Recursion]struct{})
		}
		if _, ok := seen[m]; ok {
			return true
		}
		seen[m] = struct{}{}
		return couldRecurseInfinitely(m.Model, seen)
	case Unique:
		return couldRecurseInfinitely(m.Model, seen)
	case Annotation:
		return couldRecurseInfinitely(m.Model, seen)
	}
	return false
}

func UnionOf(ms ...Model) Model {
	base := ms[0]
	for _, m := range ms[1:] {
		base = Either(base, m, nil)
	}
	return base
}
