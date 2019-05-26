// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
)

type TypeCheckingError struct {
	Want mdl.Model
	Have mdl.Model
	Path err.ErrorPath
}

// TypeCheckingError implements err.Error
var _ err.Error = TypeCheckingError{}

func (e TypeCheckingError) ErrorPath() err.ErrorPath {
	return e.Path
}

func (e TypeCheckingError) AppendPath(a err.ErrorPathElement, b ...err.ErrorPathElement) err.PathedError {
	e.Path = append(append(e.Path, a), b...)
	return e
}

func (e TypeCheckingError) Value() val.Union {
	return val.Union{"typeCheckingError", val.StructFromMap(map[string]val.Value{
		"want": mdl.ValueFromModel("", e.Want, nil), // TODO: metaID
		"have": mdl.ValueFromModel("", e.Have, nil), // TODO: metaID
		"path": e.Path.Value(),
	})}
}
func (e TypeCheckingError) Error() string {
	return e.String()
}
func (e TypeCheckingError) String() string {
	out := "Type Checking Error\n"
	out += "===================\n"
	if len(e.Path) > 0 {
		out += "Location\n"
		out += "--------\n"
		out += e.Path.String() + "\n\n"
	}
	out += "Expected\n"
	out += "--------\n"
	out += mdl.ModelToHuman(e.Want) + "\n\n"
	out += "Actual\n"
	out += "--------\n"
	out += mdl.ModelToHuman(e.Have) + "\n\n"
	return out
}
func (e TypeCheckingError) Child() err.Error {
	return nil
}

func (e TypeCheckingError) Zero() bool {
	return e.Want == nil && e.Have == nil && e.Path == nil
}

// checkType takes two models: "actual" and "expected" and returns nil if actual fits expected.
func checkType(actual, expected mdl.Model) err.Error {
	return _checkType(actual, expected, nil)
}

func _checkType(actual, expected mdl.Model, recs map[[2]*mdl.Recursion]struct{}) err.PathedError {

	expected = expected.Unwrap()

	if expected == (mdl.Any{}) {
		return nil
	}

	if ca, ok := actual.(ConstantModel); ok {
		e := expected.Validate(ca.Value, nil)
		if e == nil {
			return nil
		}
		return TypeCheckingError{expected, actual.Unwrap(), nil}
	}

	actual = actual.Unwrap()

	{ // handle recursions until convergence

		lr, lok := actual.(*mdl.Recursion)
		rr, rok := expected.(*mdl.Recursion)
		if (lok || rok) && recs == nil {
			recs = make(map[[2]*mdl.Recursion]struct{})
		}
		switch {
		case lok && rok: // both recursions
			k, ik := [2]*mdl.Recursion{lr, rr}, [2]*mdl.Recursion{rr, lr}
			if _, ok := recs[k]; ok {
				return nil // treat as equal
			}
			if _, ok := recs[ik]; ok {
				return nil // treat as equal (swapped order)
			}
			recs[k] = struct{}{}
			e := _checkType(lr.Model, rr.Model, recs)
			delete(recs, k)
			if e != nil {
				e := e.(TypeCheckingError)
				if e.Have.Equals(lr.Model) {
					e.Have = lr
				}
				if e.Want.Equals(rr.Model) {
					e.Want = rr
				}
			}
			return e

		case lok && !rok: // actual is recursive, expected not
			return TypeCheckingError{expected, actual, nil}

		case !lok && rok: // actual not recursive, expected is
			e := _checkType(actual, rr.Model, recs)
			if e != nil {
				e := e.(TypeCheckingError)
				if e.Want.Equals(rr.Model) {
					e.Want = rr
				}
			}
			return e
		}
	}

	{ // handle annotations
		la, lok := actual.(mdl.Annotation)
		ra, rok := expected.(mdl.Annotation)
		switch {
		case lok && rok: // both annotations
			return _checkType(la.Model, ra.Model, recs)
		case lok && !rok: // actual is annotation, expected is not
			return _checkType(la.Model, expected, recs)
		case !lok && rok: // actual not annotation, expected is
			return _checkType(actual, ra.Model, recs)
		}
	}

	switch expected := expected.(type) {

	case mdl.Optional:
		if _, ok := actual.(mdl.Null); ok {
			return nil
		}
		if a, ok := actual.(mdl.Optional); ok {
			return _checkType(a.Model, expected.Model, recs)
		}
		return _checkType(actual, expected.Model, recs)

	case mdl.Set:
		a, ok := actual.(mdl.Set)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		if e := _checkType(a.Elements, expected.Elements, recs); e != nil {
			return e.AppendPath(err.ErrorPathElementSetElements{})
		}
		return nil

	case mdl.List:
		a, ok := actual.(mdl.List)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		if e := _checkType(a.Elements, expected.Elements, recs); e != nil {
			return e.AppendPath(err.ErrorPathElementListElements{})
		}
		return nil

	case mdl.Map:
		a, ok := actual.(mdl.Map)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		if e := _checkType(a.Elements, expected.Elements, recs); e != nil {
			return e.AppendPath(err.ErrorPathElementMapElements{})
		}
		return nil

	case mdl.Tuple:
		a, ok := actual.(mdl.Tuple)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		if len(a) > len(expected) {
			return TypeCheckingError{expected, actual, nil}
		}
		for i, l := 0, minInt(len(a), len(expected)); i < l; i++ {
			if e := _checkType(a[i], expected[i], recs); e != nil {
				return e.AppendPath(err.ErrorPathElementTupleIndex(i))
			}
		}
		return nil

	case mdl.Struct:
		a, ok := actual.(mdl.Struct)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		e := (err.PathedError)(nil)
		expected.ForEach(func(k string, m mdl.Model) bool {
			ak, ok := a.Get(k)
			if !ok {
				if m.Nullable() {
					return true
				}
				e = TypeCheckingError{expected, actual, nil}
				return false
			}
			if e_ := _checkType(ak, m, recs); e_ != nil {
				e = e_.AppendPath(err.ErrorPathElementStructField(k))
				return false
			}
			return true
		})
		if e != nil {
			return e
		}
		return nil

	case mdl.Union:
		a, ok := actual.(mdl.Union)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		e := (err.PathedError)(nil)
		a.ForEach(func(k string, ak mdl.Model) bool {
			ek, ok := expected.Get(k)
			if !ok {
				e = TypeCheckingError{expected, actual, nil}
				return false
			}
			if e_ := _checkType(ak, ek, recs); e_ != nil {
				e = e_.AppendPath(err.ErrorPathElementUnionCase(k))
				return false
			}
			return true
		})
		if e != nil {
			return e
		}
		return nil

	case mdl.Enum:
		a, ok := actual.(mdl.Enum)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		for k, _ := range a {
			if _, ok := expected[k]; !ok {
				return TypeCheckingError{expected, actual, nil}
			}
		}
		return nil

	case mdl.Ref:
		a, ok := actual.(mdl.Ref)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		if expected.Model != "" && a.Model != expected.Model {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Null:
		_, ok := actual.(mdl.Null)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.String:
		_, ok := actual.(mdl.String)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Float:
		_, ok := actual.(mdl.Float)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Bool:
		_, ok := actual.(mdl.Bool)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.DateTime:
		_, ok := actual.(mdl.DateTime)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Int8:
		_, ok := actual.(mdl.Int8)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Int16:
		_, ok := actual.(mdl.Int16)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Int32:
		_, ok := actual.(mdl.Int32)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Int64:
		_, ok := actual.(mdl.Int64)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Uint8:
		_, ok := actual.(mdl.Uint8)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Uint16:
		_, ok := actual.(mdl.Uint16)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Uint32:
		_, ok := actual.(mdl.Uint32)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	case mdl.Uint64:
		_, ok := actual.(mdl.Uint64)
		if !ok {
			return TypeCheckingError{expected, actual, nil}
		}
		return nil

	}
	panic(fmt.Sprintf("unhandled model type: %T", expected))
}

func overMapTypeCheckingErrorsPaths(es []TypeCheckingError, f func(p err.ErrorPath) err.ErrorPath) []TypeCheckingError {
	return overMapTypeCheckingErrors(es, func(e TypeCheckingError) TypeCheckingError {
		e.Path = f(e.Path)
		return e
	})
}

func overMapTypeCheckingErrors(es []TypeCheckingError, f func(e TypeCheckingError) TypeCheckingError) []TypeCheckingError {
	for i, e := range es {
		es[i] = f(e)
	}
	return es
}

func mergeTypeCheckingErrors(left, right []TypeCheckingError) []TypeCheckingError {
	if len(left) == 0 && len(right) == 0 {
		return nil
	}
	zero := TypeCheckingError{}
	merged := make([]TypeCheckingError, 0, len(left)+len(right))
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
			merged = append(merged, TypeCheckingError{mdl.Either(l.Want, r.Want, nil), mdl.Either(l.Have, r.Have, nil), l.Path})
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
