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
func findAutoTransformation(source, target mdl.Model) (xpr.Expression, err.Error) {

	// remove transient wrappers
	source, target = source.Unwrap(), target.Unwrap()

	{ // TODO: recursion support
		_, lr := source.(*mdl.Recursion)
		_, rr := target.(*mdl.Recursion)
		if lr || rr {
			return nil, err.ExecutionError{Problem: `auto transformations with recursive models is unsupported (for now)`}
		}
	}

	// only possible target value
	if _, ok := target.(mdl.Null); ok {
		return xpr.Literal{val.Null}, nil
	}
	if target, ok := target.(mdl.Struct); ok && target.Len() == 0 {
		return xpr.NewStruct{}, nil
	}

	// simplifies output expression identical submodels
	if source.Equals(target) {
		return xpr.Argument{}, nil
	}

	switch source := source.(type) {

	case mdl.Null:
		if _, ok := target.(mdl.Null); !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Literal{val.Null}, nil

	case mdl.Set:
		target, ok := target.(mdl.Set)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		elements, e := findAutoTransformation(source.Elements, target.Elements)
		if e != nil {
			return nil, e
		}
		return xpr.MapSet{Value: xpr.Argument{}, Expression: elements}, nil

	case mdl.List:
		target, ok := target.(mdl.List)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		elements, e := findAutoTransformation(source.Elements, target.Elements)
		if e != nil {
			return nil, e
		}
		return xpr.MapList{Value: xpr.Argument{}, Expression: elements}, nil

	case mdl.Map:
		target, ok := target.(mdl.Map)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		elements, e := findAutoTransformation(source.Elements, target.Elements)
		if e != nil {
			return nil, e
		}
		return xpr.MapMap{Value: xpr.Argument{}, Expression: elements}, nil

	case mdl.Tuple:
		target, ok := target.(mdl.Tuple)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		if len(source) < len(target) {
			return nil, err.ExecutionError{
				fmt.Sprintf(`cannot infer mapping from short tuple to longer one`),
				nil,
				// C: val.Map{
				//  "source": mdl.ValueFromModel("meta", source, nil),
				//  "target": mdl.ValueFromModel("meta", target, nil),
				// },
			}
		}
		args := make([]xpr.Expression, len(target))
		for i, l := 0, len(target); i < l; i++ {
			arg, e := findAutoTransformation(source[i], target[i])
			if e != nil {
				return nil, e
			}
			args[i] = arg
		}
		return xpr.NewTuple(args), nil

	case mdl.Struct:
		target, ok := target.(mdl.Struct)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		args := make(map[string]xpr.Expression, target.Len())
		e := (err.Error)(nil)
		target.ForEach(func(k string, targetElement mdl.Model) bool {
			sourceElement := source.Field(k)
			if sourceElement == nil {
				if _, ok := targetElement.(mdl.Ref); ok {
					e = err.ExecutionError{
						fmt.Sprintf(`cannot infer mapping to new ref field in struct`),
						nil,
						// C: val.Map{
						//  "source": mdl.ValueFromModel("meta", source, nil),
						//  "target": mdl.ValueFromModel("meta", target, nil),
						// },
					}
					return false
				}
				args[k] = xpr.Literal{targetElement.Zero()}
			} else {
				arg, e_ := findAutoTransformation(sourceElement, targetElement)
				if e_ != nil {
					e = e_
					return false
				}
				args[k] = xpr.With{
					Value:  xpr.Field{Value: xpr.Argument{}, Name: xpr.Literal{val.String(k)}},
					Return: arg,
				}
			}
			return true
		})
		if e != nil {
			return nil, e
		}
		return xpr.NewStruct(args), nil

	case mdl.Union:
		target, ok := target.(mdl.Union)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		cases := make(map[string]xpr.Expression, source.Len())
		e := (err.Error)(nil)
		source.ForEach(func(k string, sourceElement mdl.Model) bool {
			targetElement := target.Case(k)
			if targetElement == nil {
				e = err.ExecutionError{
					fmt.Sprintf(`cannot infer mapping for union case "%s"`, k),
					nil,
					// C: val.Map{
					//  "source": mdl.ValueFromModel("meta", source, nil),
					//  "target": mdl.ValueFromModel("meta", target, nil),
					// },
				}
				return false
			}
			arg, e_ := findAutoTransformation(sourceElement, targetElement)
			if e_ != nil {
				e = e_
				return false
			}
			cases[k] = xpr.NewUnion{
				Case:  xpr.Literal{val.String(k)},
				Value: arg,
			}
			return true
		})
		if e != nil {
			return nil, e
		}
		return xpr.SwitchCase{Value: xpr.Argument{}, Cases: cases}, nil

	case mdl.String:
		_, ok := target.(mdl.String)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil

	case mdl.Enum:
		_, ok := target.(mdl.Enum)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		if e := checkType(source, target); e != nil {
			return nil, err.ExecutionError{
				fmt.Sprintf(`cannot infer mapping for incompatible enums`),
				nil,
				// C: val.Map{
				//  "source": mdl.ValueFromModel("meta", source, nil),
				//  "target": mdl.ValueFromModel("meta", target, nil),
				// },
			}
		}
		return xpr.Argument{}, nil

	case mdl.Float:
		_, ok := target.(mdl.Float)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil

	case mdl.Bool:
		_, ok := target.(mdl.Bool)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil

	case mdl.Any:
		_, ok := target.(mdl.Any)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil

	case mdl.DateTime:
		_, ok := target.(mdl.DateTime)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Int8:
		_, ok := target.(mdl.Int8)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Int16:
		_, ok := target.(mdl.Int16)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Int32:
		_, ok := target.(mdl.Int32)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Int64:
		_, ok := target.(mdl.Int64)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Uint8:
		_, ok := target.(mdl.Uint8)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Uint16:
		_, ok := target.(mdl.Uint16)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Uint32:
		_, ok := target.(mdl.Uint32)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Uint64:
		_, ok := target.(mdl.Uint64)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		return xpr.Argument{}, nil
	case mdl.Ref:
		target, ok := target.(mdl.Ref)
		if !ok {
			return nil, newAutoTransformationError(source, target)
		}
		if source.Model == target.Model {
			return xpr.Argument{}, nil
		}
		return xpr.RelocateRef{
			Ref:   xpr.Argument{},
			Model: xpr.Model{xpr.Literal{val.String(target.Model)}},
		}, nil
	}
	panic(fmt.Sprintf(`unhandled case: %T`, source))
}

func newAutoTransformationError(source, target mdl.Model) err.Error {
	return err.ExecutionError{
		fmt.Sprintf(`cannot infer mapping from %s to %s`, modelTypeKey(source), modelTypeKey(target)),
		nil,
	}
}

func modelTypeKey(m mdl.Model) string {
	switch m.Concrete().(type) {
	case mdl.Null:
		return "null"
	case mdl.Set:
		return "set"
	case mdl.List:
		return "list"
	case mdl.Map:
		return "map"
	case mdl.Tuple:
		return "tuple"
	case mdl.Struct:
		return "struct"
	case mdl.Union:
		return "union"
	case mdl.String:
		return "string"
	case mdl.Enum:
		return "enum"
	case mdl.Float:
		return "float"
	case mdl.Bool:
		return "bool"
	case mdl.Any:
		return "any"
	case mdl.Ref:
		return "ref"
	case mdl.DateTime:
		return "dateTime"
	case mdl.Int8:
		return "int8"
	case mdl.Int16:
		return "int16"
	case mdl.Int32:
		return "int32"
	case mdl.Int64:
		return "int64"
	case mdl.Uint8:
		return "uint8"
	case mdl.Uint16:
		return "uint16"
	case mdl.Uint32:
		return "uint32"
	case mdl.Uint64:
		return "uint64"
	}
	panic(fmt.Sprintf(`unhandled modelTypeKey case: %T`, m))
}
