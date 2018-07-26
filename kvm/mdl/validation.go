// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"github.com/kr/pretty"
	"karma.run/kvm/err"
	"karma.run/kvm/val"
)

type ValidationError struct {
	Want Model
	Have val.Value
	Path err.ErrorPath
}

// ValidationError implements err.Error
var _ err.Error = ValidationError{}

func (e ValidationError) ErrorPath() err.ErrorPath {
	return e.Path
}

func (e ValidationError) AppendPath(a err.ErrorPathElement, b ...err.ErrorPathElement) err.PathedError {
	e.Path = append(append(e.Path, a), b...)
	return e
}

func (e ValidationError) Value() val.Union {
	return val.Union{"ValidationError", val.StructFromMap(map[string]val.Value{
		"want": ValueFromModel("", e.Want, nil), // TODO: metaID
		"have": e.Have,
		"path": e.Path.Value(),
	})}
}
func (e ValidationError) Error() string {
	return e.String()
}
func (e ValidationError) String() string {
	out := "Validation Error\n"
	out += "================\n"
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
func (e ValidationError) Child() err.Error {
	return nil
}

func (m *Recursion) Validate(v val.Value, p err.ErrorPath) err.Error {
	e := m.Model.Validate(v, p)
	if e != nil {
		if e := e.(ValidationError); e.Want == m.Model {
			e.Want = m
		}
		return e
	}
	return e
}
func (m Optional) Validate(v val.Value, p err.ErrorPath) err.Error {
	if v == val.Null {
		return nil
	}
	return m.Model.Validate(v, p)
}
func (m Any) Validate(v val.Value, p err.ErrorPath) err.Error {
	return nil
}
func (m Unique) Validate(v val.Value, p err.ErrorPath) err.Error {
	return m.Model.Validate(v, p)
}
func (m Annotation) Validate(v val.Value, p err.ErrorPath) err.Error {
	return m.Model.Validate(v, p)
}
func (m Tuple) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Tuple)
	if !ok {
		return ValidationError{m, v, p}
	}
	if len(w) != len(m) {
		return ValidationError{m, v, p}
	}
	for i, l := 0, len(m); i < l; i++ {
		if e := m[i].Validate(w[i], append(p, err.ErrorPathElementTupleIndex(i))); e != nil {
			return e
		}
	}
	return nil
}
func (m Enum) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Symbol)
	if !ok {
		return ValidationError{m, v, p}
	}
	if _, ok = m[string(w)]; !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m List) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.List)
	if !ok {
		return ValidationError{m, v, p}
	}
	for i, l := 0, len(w); i < l; i++ {
		if e := m.Elements.Validate(w[i], append(p, err.ErrorPathElementListIndex(i))); e != nil {
			return e
		}
	}
	return nil
}
func (m Map) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Map)
	if !ok {
		return ValidationError{m, v, p}
	}
	for _, k := range w.Keys() {
		if e := m.Elements.Validate(w.Key(k), append(p, err.ErrorPathElementMapKey(k))); e != nil {
			return e
		}
	}
	return nil
}
func (m Set) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Set)
	if !ok {
		return ValidationError{m, v, p}
	}
	for _, v := range w {
		if e := m.Elements.Validate(v, append(p, err.ErrorPathElementSetItem{})); e != nil {
			return e
		}
	}
	return nil
}
func (m Union) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Union)
	if !ok {
		return ValidationError{m, v, p}
	}
	q, ok := m.Get(w.Case)
	if !ok {
		return ValidationError{m, v, p}
	}
	return q.Validate(w.Value, append(p, err.ErrorPathElementUnionCase(w.Case)))
}

var _ = pretty.Println

func (m Struct) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Struct)
	if !ok {
		return ValidationError{m, v, p}
	}
	for _, k := range m.Keys() {
		x, ok := w.Get(k)
		if !ok {
			x = val.Null
		}
		if e := m.Field(k).Validate(x, append(p, err.ErrorPathElementStructField(k))); e != nil {
			return e
		}
	}
	return nil
}
func (m String) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.String); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Null) Validate(v val.Value, p err.ErrorPath) err.Error {
	if v != val.Null {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Float) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Float); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m DateTime) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.DateTime); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Bool) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Bool); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Ref) Validate(v val.Value, p err.ErrorPath) err.Error {
	w, ok := v.(val.Ref)
	if !ok {
		return ValidationError{m, v, p}
	}
	if m.Model == "" { // ("any ref" case)
		return nil
	}
	if w[0] != m.Model {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Int8) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Int8); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Int16) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Int16); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Int32) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Int32); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Int64) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Int64); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Uint8) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Uint8); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Uint16) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Uint16); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Uint32) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Uint32); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
func (m Uint64) Validate(v val.Value, p err.ErrorPath) err.Error {
	if _, ok := v.(val.Uint64); !ok {
		return ValidationError{m, v, p}
	}
	return nil
}
