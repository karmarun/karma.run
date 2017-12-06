// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"kvm/err"
	"kvm/mdl"
	"kvm/val"
)

type TypeInferenceError struct {
	Want mdl.Model
	Have val.Value // literal
	Path err.ErrorPath
}

func (e TypeInferenceError) ErrorPath() err.ErrorPath {
	return e.Path
}

func (e TypeInferenceError) AppendPath(a err.ErrorPathElement, b ...err.ErrorPathElement) err.PathedError {
	e.Path = append(append(e.Path, a), b...)
	return e
}

func (e TypeInferenceError) Value() val.Union {
	return val.Union{"typeInferenceError", val.Struct{
		"want": mdl.ValueFromModel("TODO:metaID", e.Want, nil),
		"have": e.Have,
		"path": e.Path.Value(),
	}}
}
func (e TypeInferenceError) Error() string {
	return e.String()
}
func (e TypeInferenceError) String() string {
	out := "Type Inference Error\n"
	out += "====================\n"
	if len(e.Path) > 0 {
		out += "Location\n"
		out += "--------\n"
		out += e.Path.String() + "\n\n"
	}
	out += "Expected\n"
	out += "--------\n"
	out += ModelToHuman(e.Want) + "\n\n"
	out += "Actual\n"
	out += "--------\n"
	out += err.ValueToHuman(e.Have) + "\n\n"
	return out
}
func (e TypeInferenceError) Child() err.Error {
	return nil
}

func (e TypeInferenceError) Zero() bool {
	return e.Want == nil && e.Have == nil && e.Path == nil
}

// inferType takes a value and a "loose" model and returns the tightest possible
// model for the value, taking into account the expected model type (which can be any).
func inferType(value val.Value, expected mdl.Model) (mdl.Model, err.Error) {
	m, es := _inferType(value, expected)
	if len(es) == 0 {
		return m, nil
	}
	if len(es) == 1 {
		return nil, es[0]
	}
	ls := make(err.ErrorList, len(es), len(es))
	for i, e := range es {
		ls[i] = e
	}
	return nil, ls
}

func _inferType(value val.Value, expected mdl.Model) (mdl.Model, []TypeInferenceError) {

	switch m := expected.Concrete().(type) {

	case mdl.Or:
		left, e1 := _inferType(value, m[0])
		if e1 == nil {
			return left, nil
		}
		right, e2 := _inferType(value, m[1])
		if e2 == nil {
			return right, nil
		}
		return nil, mergeTypeInferenceErrors(e1, e2)

	case mdl.Any:
		return mdl.TightestModelForValue(value), nil

	case mdl.Set:
		v, ok := value.(val.Set)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if len(v) == 0 {
			return m, nil
		}
		ms := make([]mdl.Model, 0, len(v))
		for _, w := range v {
			wm, es := _inferType(w, m.Elements)
			if len(es) > 0 {
				return nil, overMapTypeInferenceErrorsPaths(es, func(p err.ErrorPath) err.ErrorPath {
					return append(p, err.ErrorPathElementSetItem{})
				})
			}
			ms = append(ms, wm)
		}
		return mdl.Set{mdl.UnionOf(ms...)}, nil

	case mdl.List:
		v, ok := value.(val.List)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if len(v) == 0 {
			return m, nil
		}
		ms := make([]mdl.Model, 0, len(v))
		for i, w := range v {
			wm, es := _inferType(w, m.Elements)
			if len(es) > 0 {
				return nil, overMapTypeInferenceErrorsPaths(es, func(p err.ErrorPath) err.ErrorPath {
					return append(p, err.ErrorPathElementListIndex(i))
				})
			}
			ms = append(ms, wm)
		}
		return mdl.List{mdl.UnionOf(ms...)}, nil

	case mdl.Map:
		v, ok := value.(val.Map)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if len(v) == 0 {
			return m, nil
		}
		ms := make([]mdl.Model, 0, len(v))
		for k, w := range v {
			wm, es := _inferType(w, m.Elements)
			if len(es) > 0 {
				return nil, overMapTypeInferenceErrorsPaths(es, func(p err.ErrorPath) err.ErrorPath {
					return append(p, err.ErrorPathElementMapKey(k))
				})
			}
			ms = append(ms, wm)
		}
		return mdl.Map{mdl.UnionOf(ms...)}, nil

	case mdl.Tuple:
		v, ok := value.(val.Tuple)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if len(v) != len(m) {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		out := make(mdl.Tuple, len(v), len(v))
		for i, w := range v {
			wm, es := _inferType(w, m[i])
			if len(es) > 0 {
				return nil, overMapTypeInferenceErrorsPaths(es, func(p err.ErrorPath) err.ErrorPath {
					return append(p, err.ErrorPathElementTupleIndex(i))
				})
			}
			out[i] = wm
		}
		return out, nil

	case mdl.Struct:
		v, ok := value.(val.Struct)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		out := make(mdl.Struct, len(v))
		for k, w := range v {
			if _, ok := m[k]; !ok {
				return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
			}
			wm, es := _inferType(w, m[k])
			if len(es) > 0 {
				return nil, overMapTypeInferenceErrorsPaths(es, func(p err.ErrorPath) err.ErrorPath {
					return append(p, err.ErrorPathElementStructField(k))
				})
			}
			out[k] = wm
		}
		return out, nil

	case mdl.Union:
		v, ok := value.(val.Union)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if _, ok := m[v.Case]; !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		out := make(mdl.Union, 1)
		wm, es := _inferType(v.Value, m[v.Case])
		if len(es) > 0 {
			return nil, overMapTypeInferenceErrorsPaths(es, func(p err.ErrorPath) err.ErrorPath {
				return append(p, err.ErrorPathElementUnionCase(v.Case))
			})
		}
		out[v.Case] = wm
		return out, nil

	case mdl.Enum:
		v, ok := value.(val.Symbol)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if _, ok := m[string(v)]; !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		out := make(mdl.Enum, 1)
		out[string(v)] = struct{}{}
		return out, nil

	case mdl.Ref:
		v, ok := value.(val.Ref)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		if m.Model != "" && v[0] != m.Model {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return mdl.Ref{v[0]}, nil

	case mdl.Null:
		_, ok := value.(val.Null)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil

	case mdl.String:
		_, ok := value.(val.String)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil

	case mdl.Float:
		_, ok := value.(val.Float)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil

	case mdl.Bool:
		_, ok := value.(val.Bool)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil

	case mdl.DateTime:
		_, ok := value.(val.DateTime)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Int8:
		_, ok := value.(val.Int8)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Int16:
		_, ok := value.(val.Int16)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Int32:
		_, ok := value.(val.Int32)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Int64:
		_, ok := value.(val.Int64)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Uint8:
		_, ok := value.(val.Uint8)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Uint16:
		_, ok := value.(val.Uint16)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Uint32:
		_, ok := value.(val.Uint32)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	case mdl.Uint64:
		_, ok := value.(val.Uint64)
		if !ok {
			return nil, []TypeInferenceError{TypeInferenceError{expected, value, nil}}
		}
		return m, nil
	}
	panic(fmt.Sprintf("unhandled model type: %T", expected.Concrete()))
}

func overMapTypeInferenceErrorsPaths(es []TypeInferenceError, f func(p err.ErrorPath) err.ErrorPath) []TypeInferenceError {
	return overMapTypeInferenceErrors(es, func(e TypeInferenceError) TypeInferenceError {
		e.Path = f(e.Path)
		return e
	})
}

func overMapTypeInferenceErrors(es []TypeInferenceError, f func(e TypeInferenceError) TypeInferenceError) []TypeInferenceError {
	for i, e := range es {
		es[i] = f(e)
	}
	return es
}

func mergeTypeInferenceErrors(left, right []TypeInferenceError) []TypeInferenceError {
	if len(left) == 0 && len(right) == 0 {
		return nil
	}
	zero := TypeInferenceError{}
	merged := make([]TypeInferenceError, 0, len(left)+len(right))
	unmatchedLeft := left[:0] // reuse memory, avoid allocation
	for _, l := range left {
		match := -1
		for i, r := range right {
			if l.Path.Equals(r.Path) {
				match = i
				break
			}
		}
		if match == -1 {
			unmatchedLeft = append(unmatchedLeft, l)
		} else {
			r := right[match]
			merged = append(merged, TypeInferenceError{mdl.Either(l.Want, r.Want, nil), l.Have, l.Path}) // note: same path implies same .Have
			right[match] = zero
		}
	}
	unmatchedRight := right[:0] // reuse memory, avoid allocation
	for _, r := range right {
		if r.Zero() {
			continue
		}
		unmatchedRight = append(unmatchedRight, r)
	}
	return append(append(merged, unmatchedLeft...), unmatchedRight...)
}
