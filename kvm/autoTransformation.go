// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
)

// findAutoTransformation tries to infer a transformation function from the source model to the target model.
func findAutoTransformation(source, target mdl.Model) (xpr.Function, err.PathedError) {
	t, e := _findAutoTransformation(source, target)
	if e != nil {
		return nil, e
	}
	return xpr.NewFunction([]string{"source"}, t), nil
}

func _findAutoTransformation(source, target mdl.Model) (xpr.Expression, err.PathedError) {

	if source.Equals(target) {
		return xpr.Scope("source"), nil
	}

	if _, ok := source.(*mdl.Recursion); ok {
		return nil, NewAutoTransformationError(`source model is recursive.`, source, target)
	}

	if _, ok := target.(*mdl.Recursion); ok {
		return nil, NewAutoTransformationError(`target model is recursive.`, source, target)
	}

	source, target = source.Concrete(), target.Concrete()

	if _, ok := source.(mdl.Optional); !ok {
		if target, ok := target.(mdl.Optional); ok {
			sub, e := _findAutoTransformation(source, target.Model)
			if e != nil {
				return xpr.Literal{val.Null}, nil
			}
			return sub, nil
		}
	}

	switch source := source.(type) {

	case mdl.Optional:
		if target, ok := target.(mdl.Optional); ok {
			sub, e := _findAutoTransformation(source.Model, target.Model)
			if e != nil {
				return nil, e
			}
			return xpr.If{
				Condition: xpr.IsPresent{xpr.Scope("source")},
				Then: xpr.With{
					Value:  xpr.AssertPresent{xpr.Scope("source")},
					Return: xpr.NewFunction([]string{"source"}, sub),
				},
				Else: xpr.Literal{val.Null},
			}, nil
		}
		if target.Zeroable() {
			sub, e := _findAutoTransformation(source.Model, target)
			if e != nil {
				return nil, e
			}
			return xpr.If{
				Condition: xpr.IsPresent{xpr.Scope("source")},
				Then: xpr.With{
					Value:  xpr.AssertPresent{xpr.Scope("source")},
					Return: xpr.NewFunction([]string{"source"}, sub),
				},
				Else: xpr.Literal{target.Zero()},
			}, nil
		}
		return nil, NewAutoTransformationError(`source is optional but target is neither optional nor zeroable.`, source, target)

	case mdl.Tuple:
		if target, ok := target.(mdl.Tuple); ok {
			switch {
			case len(source) == len(target):
				out := make(xpr.NewTuple, len(source), len(source))
				for i, l := 0, len(source); i < l; i++ {
					sub, e := _findAutoTransformation(source[i], target[i])
					if e != nil {
						return nil, e.AppendPath(err.ErrorPathElementTupleIndex(i))
					}
					out[i] = xpr.With{
						Value:  xpr.IndexTuple{xpr.Scope("source"), val.Int64(i)},
						Return: xpr.NewFunction([]string{"source"}, sub),
					}
				}
				return out, nil

			case len(source) > len(target):
				clip := len(target)
				out := make(xpr.NewTuple, clip, clip)
				for i, l := 0, clip; i < l; i++ {
					sub, e := _findAutoTransformation(source[i], target[i])
					if e != nil {
						return nil, e.AppendPath(err.ErrorPathElementTupleIndex(i))
					}
					out[i] = xpr.With{
						Value:  xpr.IndexTuple{xpr.Scope("source"), val.Int64(i)},
						Return: xpr.NewFunction([]string{"source"}, sub),
					}
				}
				return out, nil

			case len(source) < len(target):
				prefix := len(source)
				out := make(xpr.NewTuple, prefix, len(target))
				for i, l := 0, prefix; i < l; i++ {
					sub, e := _findAutoTransformation(source[i], target[i])
					if e != nil {
						return nil, e.AppendPath(err.ErrorPathElementTupleIndex(i))
					}
					out[i] = xpr.With{
						Value:  xpr.IndexTuple{xpr.Scope("source"), val.Int64(i)},
						Return: xpr.NewFunction([]string{"source"}, sub),
					}
				}
				for i, l := prefix, len(target); i < l; i++ {
					if !target[i].Zeroable() {
						e := NewAutoTransformationError(`target tuple longer than source and higher elements not zeroable`, source, target)
						return nil, e.AppendPath(err.ErrorPathElementTupleIndex(i))
					}
					out = append(out, xpr.Literal{target[i].Zero()})
				}
				return out, nil
			}
		}
		return nil, NewAutoTransformationError(`source is tuple but target is not`, source, target)

	case mdl.List:
		if target, ok := target.(mdl.List); ok {
			sub, e := _findAutoTransformation(source.Elements, target.Elements)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementListElements{})
			}
			return xpr.MapList{
				Value:   xpr.Scope("source"),
				Mapping: xpr.NewFunction([]string{"i", "source"}, sub),
			}, nil
		}
		return nil, NewAutoTransformationError(`source is list but target is not`, source, target)

	case mdl.Set:
		if target, ok := target.(mdl.Set); ok {
			sub, e := _findAutoTransformation(source.Elements, target.Elements)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementSetElements{})
			}
			return xpr.MapSet{
				Value:   xpr.Scope("source"),
				Mapping: xpr.NewFunction([]string{"i", "source"}, sub),
			}, nil
		}
		return nil, NewAutoTransformationError(`source is set but target is not`, source, target)

	case mdl.Map:
		if target, ok := target.(mdl.Map); ok {
			sub, e := _findAutoTransformation(source.Elements, target.Elements)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementMapElements{})
			}
			return xpr.MapMap{
				Value:   xpr.Scope("source"),
				Mapping: xpr.NewFunction([]string{"i", "source"}, sub),
			}, nil
		}
		return nil, NewAutoTransformationError(`source is map but target is not`, source, target)

	case mdl.Struct:
		if target, ok := target.(mdl.Struct); ok {
			out := make(xpr.NewStruct, maxInt(source.Len(), target.Len()))
			errout := (err.PathedError)(nil)
			source.ForEach(func(field string, sourceElement mdl.Model) bool {
				targetElement, ok := target.Get(field)
				if !ok {
					return true
				}
				sub, e := _findAutoTransformation(sourceElement, targetElement)
				if e != nil {
					errout = e.AppendPath(err.ErrorPathElementStructField(field))
					return false
				}
				out[field] = xpr.With{
					Value:  xpr.Field{field, xpr.Scope("source")},
					Return: xpr.NewFunction([]string{"source"}, sub),
				}
				return true
			})
			if errout != nil {
				return nil, errout
			}
			// deal with target fields not present in source
			target.ForEach(func(field string, targetElement mdl.Model) bool {
				sourceElement, ok := source.Get(field)
				if ok {
					return true
				}
				if targetElement.Zeroable() {
					out[field] = xpr.Literal{targetElement.Zero()}
					return true
				}
				e := NewAutoTransformationError(`new field in target struct not zeroable`, sourceElement, targetElement)
				errout = e.AppendPath(err.ErrorPathElementStructField(field))
				return false
			})
			if errout != nil {
				return nil, errout
			}
			return out, nil
		}
		return nil, NewAutoTransformationError(`source is struct but target is not`, source, target)

	case mdl.Union:
		if target, ok := target.(mdl.Union); ok {
			out := make(map[string]xpr.Function, maxInt(source.Len(), target.Len()))
			errout := (err.PathedError)(nil)
			source.ForEach(func(caze string, sourceElement mdl.Model) bool {
				targetElement, ok := target.Get(caze)
				if !ok {
					errout = NewAutoTransformationError(`source union has cases not defined in target union`, source, target)
					return false
				}
				sub, e := _findAutoTransformation(sourceElement, targetElement)
				if e != nil {
					errout = e.AppendPath(err.ErrorPathElementUnionCase(caze))
					return false
				}
				out[caze] = xpr.NewFunction([]string{"source"}, xpr.NewUnion{
					Case:  caze,
					Value: sub,
				})
				return true
			})
			return xpr.SwitchCase{
				Value: xpr.Scope("source"),
				Cases: out,
			}, nil
		}
		return nil, NewAutoTransformationError(`source is union but target is not`, source, target)

	case mdl.Enum:
		if target, ok := target.(mdl.Enum); ok {
			cover := true
			for key, _ := range source {
				_, ok := target[key]
				cover = cover && ok
				if !cover {
					break
				}
			}
			if !cover {
				return nil, NewAutoTransformationError(`source enum contains cases not present in target enum`, source, target)
			}
			return xpr.Scope("source"), nil
		}
		return nil, NewAutoTransformationError(`source is enum but target is not`, source, target)

	case mdl.Ref:
		if target, ok := target.(mdl.Ref); ok {
			if source.Model == target.Model {
				return xpr.Scope("source"), nil
			}
			return xpr.RelocateRef{
				Ref:   xpr.Scope("source"),
				Model: xpr.Model{xpr.Literal{val.String(target.Model)}},
			}, nil
		}
		return nil, NewAutoTransformationError(`source is ref but target is not`, source, target)

	case mdl.Null:
		return nil, NewAutoTransformationError(`source is null but target is not`, source, target)

	case mdl.Any:
		// NOTE: this case doesn't happen in practice
		return nil, NewAutoTransformationError(`source is any but target is not`, source, target)

	case mdl.String: // string -> string case caught in function prelude
		return nil, NewAutoTransformationError(`source is string but target is not`, source, target)

	case mdl.Float:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is float but target is not`, source, target)

	case mdl.Int8:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is int8 but target is not`, source, target)

	case mdl.Int16:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is int16 but target is not`, source, target)

	case mdl.Int32:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is int32 but target is not`, source, target)

	case mdl.Int64:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is int64 but target is not`, source, target)

	case mdl.Uint8:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is uint8 but target is not`, source, target)

	case mdl.Uint16:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is uint16 but target is not`, source, target)

	case mdl.Uint32:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is uint32 but target is not`, source, target)

	case mdl.Uint64:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is uint64 but target is not`, source, target)

	case mdl.DateTime:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is dateTime but target is not`, source, target)

	case mdl.Bool:
		// TODO: numeric conversion functions
		return nil, NewAutoTransformationError(`source is bool but target is not`, source, target)

	}
	panic(fmt.Sprintf(`unhandled source model: %T`, source))
}

type AutoTransformationError struct {
	Problem string
	Source  mdl.Model
	Target  mdl.Model
	Path    err.ErrorPath
}

func NewAutoTransformationError(problem string, source, target mdl.Model) AutoTransformationError {
	return AutoTransformationError{problem, source, target, make(err.ErrorPath, 0, 8)}
}

func (e AutoTransformationError) Value() val.Union {
	params := val.NewStruct(2)
	params.Set("problem", val.String(e.Problem))
	params.Set("problem", e.Path.Value())
	return val.Union{"autoMigrationError", params}
}
func (e AutoTransformationError) Error() string {
	return e.Problem
}
func (e AutoTransformationError) String() string {
	out := "Automigration Error\n"
	out += "===================\n"
	if len(e.Path) > 0 {
		out += "Location\n"
		out += "--------\n"
		out += e.Path.String() + "\n\n"
	}
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n"
	return out
}
func (e AutoTransformationError) Child() err.Error {
	return nil
}
func (e AutoTransformationError) ErrorPath() err.ErrorPath {
	return e.Path
}
func (e AutoTransformationError) AppendPath(a err.ErrorPathElement, b ...err.ErrorPathElement) err.PathedError {
	e.Path = append(append(e.Path, a), b...)
	return e
}
