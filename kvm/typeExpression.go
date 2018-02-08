// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"karma.run/definitions"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"regexp"
)

var (
	AnyModel      = mdl.Any{}
	BoolModel     = mdl.Bool{}
	Int8Model     = mdl.Int8{}
	Int16Model    = mdl.Int16{}
	Int32Model    = mdl.Int32{}
	Int64Model    = mdl.Int64{}
	Uint8Model    = mdl.Uint8{}
	Uint16Model   = mdl.Uint16{}
	Uint32Model   = mdl.Uint32{}
	Uint64Model   = mdl.Uint64{}
	FloatModel    = mdl.Float{}
	StringModel   = mdl.String{}
	DateTimeModel = mdl.DateTime{}
	IntegerModel  = mdl.UnionOf([]mdl.Model{ // TODO: this is now Any
		Int8Model,
		Int16Model,
		Int32Model,
		Int64Model,
		Uint8Model,
		Uint16Model,
		Uint32Model,
		Uint64Model,
	}...)
	NumericModel = mdl.UnionOf([]mdl.Model{
		FloatModel,
		Int8Model,
		Int16Model,
		Int32Model,
		Int64Model,
		Uint8Model,
		Uint16Model,
		Uint32Model,
		Uint64Model,
	}...)
	NullModel     = mdl.Null{}
	SortableModel = mdl.UnionOf([]mdl.Model{
		FloatModel,
		BoolModel,
		StringModel,
		DateTimeModel,
		Int8Model,
		Int16Model,
		Int32Model,
		Int64Model,
		Uint8Model,
		Uint16Model,
		Uint32Model,
		Uint64Model,
	}...)
)

var ZeroTypedExpression = xpr.TypedExpression{}

// VirtualMachine.TypeExpression infers, propagates and checks type information in an xpr.Expression tree.
// It returns an equivalent tree, where each xpr.Expression is wrapped in an xpr.TypedExpression.
// Some typing decisions depend on constant expressions. This is why constant values are propagated as well.
// This is done by wrapping the pertinent types and associated values in ContantModel.
// Parameter argument describes the expression's argument's type.
// Parameter expected indicates the expected return type of the expression.
func (vm VirtualMachine) TypeExpression(node xpr.Expression, argument, expected mdl.Model) (xpr.TypedExpression, err.Error) {

	if expected == nil {
		expected = AnyModel
	}

	if argument == nil {
		argument = AnyModel
	}

	if _, ok := expected.(BucketModel); ok {
		panic("INVARIANT: expected is never BucketModel")
	}

	if _, ok := expected.(ConstantModel); ok {
		panic("INVARIANT: expected is never ConstantModel")
	}

	retNode := ZeroTypedExpression

	switch node := node.(type) {

	case xpr.Argument:
		retNode = xpr.TypedExpression{node, expected, argument}

	case xpr.CurrentUser:
		// NOTE: do not bake current user ref into AST as constant (because of compiler cache).
		retNode = xpr.TypedExpression{node, expected, mdl.Ref{vm.UserModelId()}}

	case xpr.Zero:
		if !expected.Zeroable() {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `no zeroable type in scope`,
				Program: val.Union{"zero", val.Struct{}},
			}
		}
		retNode = xpr.TypedExpression{node, expected, ConstantModel{expected, expected.Zero()}}

	case xpr.NewUnion:

		caze, e := vm.TypeExpression(node.Case, argument, StringModel)
		if e != nil {
			return caze, e
		}
		node.Case = caze

		cc, ok := caze.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `union: case must be constant expression`,
				Program: xpr.ValueFromExpression(node),
			}
		}

		caseString := string(cc.Value.(val.String))

		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		model := mdl.NewUnion(1)
		model.Set(caseString, value.Actual.Unwrap())

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.NewList:

		subModel := mdl.Model(nil)

		for i, arg := range node {
			arg, e := vm.TypeExpression(arg, argument, AnyModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			if subModel == nil {
				subModel = arg.Actual.Unwrap()
			} else {
				subModel = mdl.Either(subModel, arg.Actual.Unwrap(), nil)
			}
		}

		if len(node) == 0 {
			if em, ok := expected.Unwrap().(mdl.List); ok {
				subModel = em.Elements
			} else {
				subModel = AnyModel
			}
		}

		retNode = xpr.TypedExpression{node, expected, mdl.List{subModel}}

	case xpr.NewMap:

		subModel := mdl.Model(nil)

		for k, arg := range node {
			arg, e := vm.TypeExpression(arg, argument, AnyModel)
			if e != nil {
				return arg, e
			}
			node[k] = arg
			if subModel == nil {
				subModel = arg.Actual.Unwrap()
			} else {
				subModel = mdl.Either(subModel, arg.Actual.Unwrap(), nil)
			}
		}

		if len(node) == 0 {
			if em, ok := expected.Unwrap().(mdl.Map); ok {
				subModel = em.Elements
			} else {
				subModel = AnyModel
			}
		}

		retNode = xpr.TypedExpression{node, expected, mdl.Map{subModel}}

	case xpr.NewSet:

		subModel := mdl.Model(nil)

		for i, arg := range node {
			arg, e := vm.TypeExpression(arg, argument, AnyModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			if subModel == nil {
				subModel = arg.Actual.Unwrap()
			} else {
				subModel = mdl.Either(subModel, arg.Actual.Unwrap(), nil)
			}
		}

		if len(node) == 0 {
			if em, ok := expected.Unwrap().(mdl.Set); ok {
				subModel = em.Elements
			} else {
				subModel = AnyModel
			}
		}

		retNode = xpr.TypedExpression{node, expected, mdl.Set{subModel}}

	case xpr.NewStruct:

		model := mdl.NewStruct(len(node))

		for k, arg := range node {
			arg, e := vm.TypeExpression(arg, argument, AnyModel)
			if e != nil {
				return arg, e
			}
			node[k] = arg
			model.Set(k, arg.Actual.Unwrap())
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.NewTuple:

		model := make(mdl.Tuple, len(node))

		for i, arg := range node {
			arg, e := vm.TypeExpression(arg, argument, AnyModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			model[i] = arg.Actual.Unwrap()
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.SetField:

		name, e := vm.TypeExpression(node.Name, argument, StringModel)
		if e != nil {
			return name, e
		}
		node.Name = name

		field := ""
		if cn, ok := name.Actual.(ConstantModel); ok {
			field = string(cn.Value.(val.String))
		} else {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `setField: name must be constant expression`,
				Program: xpr.ValueFromExpression(node),
			}
		}

		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		in, e := vm.TypeExpression(node.In, argument, mdl.Struct{}) // any struct
		if e != nil {
			return in, e
		}
		node.In = in

		m := in.Actual.Concrete().Copy().(mdl.Struct)
		m.Set(field, value.Actual.Unwrap())

		retNode = xpr.TypedExpression{node, expected, m}

	case xpr.SetKey:
		name, e := vm.TypeExpression(node.Name, argument, StringModel)
		if e != nil {
			return name, e
		}
		node.Name = name

		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		in, e := vm.TypeExpression(node.In, argument, mdl.Map{AnyModel})
		if e != nil {
			return in, e
		}
		node.In = in

		retNode = xpr.TypedExpression{node, expected, mdl.Map{value.Actual.Unwrap()}}

	case xpr.NewRef:
		marg, e := vm.TypeExpression(node.Model, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return marg, e
		}
		node.Model = marg

		cm, ok := marg.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `ref: model argument must be constant expression`,
				Program: xpr.ValueFromExpression(marg),
			}
		}

		mid := cm.Value.(val.Ref)[1]

		iarg, e := vm.TypeExpression(node.Id, argument, StringModel)
		if e != nil {
			return iarg, e
		}
		node.Id = iarg

		// TODO: check that ID exists (during execution)

		retNode = xpr.TypedExpression{node, expected, mdl.Ref{mid}}

	case xpr.PresentOrZero:

		arg, e := vm.TypeExpression(node.Argument, argument, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg

		_, ok := arg.Actual.Concrete().(mdl.Optional)
		alwaysPresent := !ok

		if alwaysPresent {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`presentOrZero: argument is always present`),
				Program: xpr.ValueFromExpression(arg),
			}
		}

		model := deoptionalize(arg.Actual.Concrete())
		if model.Concrete() == NullModel {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`presentOrZero: argument is never present`),
				Program: xpr.ValueFromExpression(arg),
			}
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.AllReferrers:

		arg, e := vm.TypeExpression(node.Argument, argument, mdl.Ref{})
		if e != nil {
			return arg, e
		}
		node.Argument = arg

		retNode = xpr.TypedExpression{node, expected, mdl.List{mdl.Ref{}}}

	case xpr.IsPresent:
		arg, e := vm.TypeExpression(node.Argument, argument, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		model := mdl.Model(BoolModel)
		if deoptionalize(arg.Actual) == NullModel {
			model = ConstantModel{model, val.Bool(false)}
		} else if ca, ok := arg.Actual.(ConstantModel); ok {
			model = ConstantModel{model, val.Bool(ca.Value != val.Null)}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.AssertPresent:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.Optional{expected})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		model := deoptionalize(UnwrapBucket(arg.Actual))
		if model == NullModel {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `assertPresent: value was absent`,
				Program: xpr.ValueFromExpression(arg),
			}
		} else if ca, ok := model.(ConstantModel); ok {
			if ca.Value == val.Null {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `assertPresent: value was absent`,
					Program: xpr.ValueFromExpression(arg),
				}
			}
			model = ConstantModel{model, ca.Value}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Model:
		arg, e := vm.TypeExpression(node.Argument, argument, StringModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		metaId := vm.MetaModelId()
		model := mdl.Model(mdl.Ref{metaId})
		if ca, ok := arg.Actual.(ConstantModel); ok {
			model = ConstantModel{model, val.Ref{metaId, string(ca.Value.(val.String))}}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Tag:
		arg, e := vm.TypeExpression(node.Argument, argument, StringModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		metaId := vm.MetaModelId()
		model := mdl.Model(mdl.Ref{metaId})
		if ca, ok := arg.Actual.(ConstantModel); ok {
			tag := string(ca.Value.(val.String))
			mid := vm.RootBucket.Bucket(definitions.TagBucketBytes).Get([]byte(tag))
			if mid == nil {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `tag: not found`,
					Program: xpr.ValueFromExpression(arg),
				}
			}
			model = ConstantModel{model, val.Ref{metaId, string(mid)}}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.All:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		if ca, ok := arg.Actual.(ConstantModel); ok {
			mid := ca.Value.(val.Ref)[1]
			model, e := vm.Model(mid)
			if e != nil {
				return ZeroTypedExpression, e
			}
			retNode = xpr.TypedExpression{node, expected, mdl.List{model}} // model is BucketModel
		} else {
			retNode = xpr.TypedExpression{node, expected, mdl.List{AnyModel}}
		}

	case xpr.JoinStrings:

		strings, e := vm.TypeExpression(node.Strings, argument, mdl.List{StringModel})
		if e != nil {
			return strings, e
		}
		node.Strings = strings

		separator, e := vm.TypeExpression(node.Separator, argument, StringModel)
		if e != nil {
			return separator, e
		}
		node.Separator = separator

		if cs, ok := strings.Actual.(ConstantModel); ok {
			if cp, ok := separator.Actual.(ConstantModel); ok {
				_, _ = cs, cp // TODO: constant optimization
			}
		}

		retNode = xpr.TypedExpression{node, expected, StringModel}

	case xpr.StringToLower:
		arg, e := vm.TypeExpression(node.Argument, argument, StringModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		if ca, ok := arg.Actual.(ConstantModel); ok {
			_ = ca // TODO: constant optimization
		}
		retNode = xpr.TypedExpression{node, expected, StringModel}

	case xpr.ReverseList:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.List{AnyModel})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		retNode = xpr.TypedExpression{node, expected, arg.Actual}

	case xpr.ExtractStrings:
		arg, e := vm.TypeExpression(node.Argument, argument, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		if ca, ok := arg.Actual.(ConstantModel); ok {
			_ = ca // TODO: constant optimization
		}
		retNode = xpr.TypedExpression{node, expected, mdl.List{StringModel}}

	case xpr.Delete:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.Ref{""})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		mid := arg.Actual.Concrete().(mdl.Ref).Model
		model, e := vm.Model(mid)
		if e != nil {
			return ZeroTypedExpression, e
		}
		unwrapped := UnwrapBucket(model)
		retNode = xpr.TypedExpression{node, expected, unwrapped}

	case xpr.Literal:
		actual, e := inferType(node.Value, expected)
		if e != nil {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `literal type mismatch`,
				Program: node.Value,
				Child_:  e,
			}
		}
		retNode = xpr.TypedExpression{xpr.Literal{node.Value}, expected, ConstantModel{actual, node.Value}}

	case xpr.ResolveAllRefs:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.Any{})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		rewritten := arg.Actual.Copy().Transform(func(m mdl.Model) mdl.Model {
			if e != nil {
				return m
			}
			if rf, ok := m.(mdl.Ref); ok {
				m, e2 := vm.Model(rf.Model)
				if e2 != nil {
					e = e2
				}
				return UnwrapBucket(m)
			}
			return m
		})
		retNode = xpr.TypedExpression{node, expected, rewritten}

	case xpr.First:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.List{expected})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		model := arg.Actual.Concrete().(mdl.List).Elements
		if ca, ok := arg.Actual.(ConstantModel); ok {
			ls := ca.Value.(val.List)
			if len(ls) == 0 {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `first: empty list`,
					Program: xpr.ValueFromExpression(arg),
				}
			}
			model = ConstantModel{ca.Model.(mdl.List).Elements, ls[0]}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Get:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.Ref{""})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		mid := arg.Actual.Concrete().(mdl.Ref).Model
		model, e := vm.Model(mid)
		if e != nil {
			return ZeroTypedExpression, e
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Length:
		arg, e := vm.TypeExpression(node.Argument, argument, mdl.List{AnyModel})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		model := mdl.Model(Int64Model)
		if ca, ok := arg.Actual.(ConstantModel); ok {
			model = ConstantModel{model, val.Int64(len(ca.Value.(val.List)))}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Not:
		arg, e := vm.TypeExpression(node.Argument, argument, BoolModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		model := mdl.Model(BoolModel)
		if ca, ok := arg.Actual.(ConstantModel); ok {
			model = ConstantModel{model, !ca.Value.(val.Bool)}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.ModelOf:
		arg, e := vm.TypeExpression(node.Argument, argument, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		unwrapped := arg.Actual.Transform(UnwrapConstant).Transform(UnwrapBucket)
		modelValue := mdl.ValueFromModel(vm.MetaModelId(), unwrapped, nil)
		retNode = xpr.TypedExpression{node, expected, ConstantModel{vm.MetaModel(), modelValue}}

	case xpr.Metarialize:
		arg, e := vm.TypeExpression(node.Argument, argument, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		ba, ok := arg.Actual.(BucketModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `metarialize: argument is not persistent`,
				Program: xpr.ValueFromExpression(arg),
			}
		}
		model := vm.WrapModelInMeta(ba.Bucket, ba.Model)
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.RefTo:
		arg, e := vm.TypeExpression(node.Argument, argument, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		ba, ok := arg.Actual.(BucketModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `refTo: argument is not persistent`,
				Program: xpr.ValueFromExpression(arg),
			}
		}
		model := mdl.Ref{ba.Bucket}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.If:
		condition, e := vm.TypeExpression(node.Condition, argument, BoolModel)
		if e != nil {
			return condition, e
		}
		node.Condition = condition
		then, e := vm.TypeExpression(node.Then, argument, AnyModel)
		if e != nil {
			return then, e
		}
		node.Then = then
		elze, e := vm.TypeExpression(node.Else, argument, AnyModel)
		if e != nil {
			return elze, e
		}
		node.Else = elze
		// TODO: constant condition optimization
		retNode = xpr.TypedExpression{node, expected, mdl.Either(UnwrapConstant(UnwrapBucket(then.Actual)), UnwrapConstant(UnwrapBucket(elze.Actual)), nil)}

	case xpr.With:
		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value
		retrn, e := vm.TypeExpression(node.Return, value.Actual, expected)
		if e != nil {
			return retrn, e
		}
		node.Return = retrn
		retNode = xpr.TypedExpression{node, expected, retrn.Actual}

	case xpr.Update:
		ref, e := vm.TypeExpression(node.Ref, argument, mdl.Ref{""})
		if e != nil {
			return ref, e
		}
		node.Ref = ref
		mid := ref.Actual.Concrete().(mdl.Ref).Model
		if mid == vm.MetaModelId() {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `update: models are immutable`,
				Program: xpr.ValueFromExpression(ref),
			}
		}
		subExpect, e := vm.Model(mid)
		if e != nil {
			return ZeroTypedExpression, e
		}
		value, e := vm.TypeExpression(node.Value, argument, subExpect.Concrete())
		if e != nil {
			return value, e
		}
		node.Value = value
		retNode = xpr.TypedExpression{node, expected, mdl.Ref{mid}}

	case xpr.Create:
		in, e := vm.TypeExpression(node.In, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return in, e
		}
		node.In = in
		ci, ok := in.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `create: model argument must be constant expression`,
				Program: xpr.ValueFromExpression(in),
			}
		}
		mid := ci.Value.(val.Ref)[1]
		subExpect, e := vm.Model(mid)
		if e != nil {
			return ZeroTypedExpression, e
		}
		value, e := vm.TypeExpression(node.Value, argument, UnwrapBucket(subExpect))
		if e != nil {
			return value, e
		}
		node.Value = value
		retNode = xpr.TypedExpression{node, expected, mdl.Ref{mid}}

	case xpr.InList:

		in, e := vm.TypeExpression(node.In, argument, mdl.List{AnyModel})
		if e != nil {
			return in, e
		}
		node.In = in

		subExpect := in.Actual.Concrete().(mdl.List).Elements

		value, e := vm.TypeExpression(node.Value, argument, UnwrapBucket(subExpect))
		if e != nil {
			return value, e
		}
		node.Value = value

		model := mdl.Model(BoolModel)
		if cv, ok := value.Actual.(ConstantModel); ok {
			if cl, ok := in.Actual.(ConstantModel); ok {
				found := val.Bool(false)
				for _, v := range cl.Value.(val.List) {
					if cv.Value.Equals(v) {
						found = true
						break
					}
				}
				model = ConstantModel{model, found}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Filter:
		value, e := vm.TypeExpression(node.Value, argument, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value
		subArg := value.Actual.Concrete().(mdl.List).Elements
		expression, e := vm.TypeExpression(node.Expression, subArg, BoolModel)
		if e != nil {
			return expression, e
		}
		node.Expression = expression
		if ce, ok := expression.Actual.(ConstantModel); ok {
			v := ce.Value.(val.Bool)
			if v {
				retNode = value
			} else {
				retNode = xpr.TypedExpression{node, expected, ConstantModel{value.Actual.Unwrap(), make(val.List, 0, 0)}}
			}
		} else {
			retNode = xpr.TypedExpression{node, expected, value.Actual.Unwrap()}
		}

	case xpr.AssertCase:
		caze, e := vm.TypeExpression(node.Case, argument, StringModel)
		if e != nil {
			return caze, e
		}
		node.Case = caze
		cc, ok := caze.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `assertCase: case argument must be constant expression`,
				Program: xpr.ValueFromExpression(caze),
			}
		}
		caseString := string(cc.Value.(val.String))

		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		um, ok := value.Actual.Concrete().(mdl.Union)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `assertCase: value argument is not a union`,
				Program: xpr.ValueFromExpression(value),
				// C: val.Map{"model": mdl.ValueFromModel(vm.MetaModelId(), value.Actual.Concrete(), nil)},
			}
		}

		model, ok := um.Get(caseString)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`assertCase: value never has case specified case`),
				Program: xpr.ValueFromExpression(value),
				// C: val.Map{"model": mdl.ValueFromModel(vm.MetaModelId(), um.Concrete(), nil)},
			}
		}
		if cv, ok := value.Actual.(ConstantModel); ok {
			uv := cv.Value.(val.Union)
			if uv.Case != caseString {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `assertCase: assertion failed`,
					Program: xpr.ValueFromExpression(value),
					// C: val.Map{"case": val.String(caseString)},
				}
			}
			model = ConstantModel{cv.Model.(mdl.Union).Case(caseString), uv.Value}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.IsCase:
		caze, e := vm.TypeExpression(node.Case, argument, StringModel)
		if e != nil {
			return caze, e
		}
		node.Case = caze
		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		if _, ok := value.Actual.Concrete().(mdl.Union); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `isCase: value argument is not a union`,
				Program: xpr.ValueFromExpression(value),
				// C: val.Map{"model": mdl.ValueFromModel(vm.MetaModelId(), value.Actual.Concrete(), nil)},
			}
		}
		node.Value = value
		// TODO: constant optimization
		retNode = xpr.TypedExpression{node, expected, BoolModel}

	case xpr.MapMap:
		value, e := vm.TypeExpression(node.Value, argument, mdl.Map{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value
		subArg := mdl.StructFromMap(map[string]mdl.Model{
			"key":   StringModel,
			"value": value.Actual.Concrete().(mdl.Map).Elements,
		})
		expression, e := vm.TypeExpression(node.Expression, subArg, AnyModel)
		if e != nil {
			return expression, e
		}
		node.Expression = expression
		retNode = xpr.TypedExpression{node, expected, mdl.Map{expression.Actual}}

	case xpr.MapList:

		value, e := vm.TypeExpression(node.Value, argument, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		subArg := value.Actual.Concrete().(mdl.List).Elements

		expression, e := vm.TypeExpression(node.Expression, subArg, AnyModel)
		if e != nil {
			return expression, e
		}
		node.Expression = expression

		retNode = xpr.TypedExpression{node, expected, mdl.List{expression.Actual}}

	case xpr.ReduceList:

		value, e := vm.TypeExpression(node.Value, argument, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		// TODO: constant optimization

		subArg := value.Actual.Concrete().(mdl.List).Elements
		bottom, typeOK := mdl.Model(nil), false
		for i := 0; i < 100; i++ { // 100 iterations as sanity
			if bottom != nil {
				subArg = mdl.Either(subArg, bottom, nil)
			}
			expression, e := vm.TypeExpression(node.Expression, mdl.Tuple{subArg, subArg}, expected)
			if e != nil {
				return expression, e
			}
			subType := expression.Actual.Concrete()
			if bottom == nil {
				bottom = subType
			} else {
				bottom = mdl.Either(bottom, subType, nil)
			}
			if e := checkType(subType, bottom); e == nil {
				node.Expression = expression
				retNode = xpr.TypedExpression{node, expected, bottom}
				typeOK = true
				break
			}
		}
		if !typeOK {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `reduceList: no sensible return type could be found`,
				Program: xpr.ValueFromExpression(node.Expression.(xpr.TypedExpression)),
			}
		}

	case xpr.ResolveRefs:

		mids := make(map[string]struct{}, len(node.Models))

		for i, sub := range node.Models {
			model, e := vm.TypeExpression(sub, argument, mdl.Ref{vm.MetaModelId()})
			if e != nil {
				return model, e
			}
			node.Models[i] = model
			cm, ok := model.Actual.(ConstantModel)
			if !ok {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `resolveRefs: model arguments must be constant expressions`,
					Program: xpr.ValueFromExpression(model),
				}
			}
			mids[cm.Value.(val.Ref)[1]] = struct{}{}
		}

		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		rewritten := value.Actual.Copy().Transform(func(m mdl.Model) mdl.Model {
			if e != nil {
				return m
			}
			if rf, ok := m.(mdl.Ref); ok {
				if _, ok := mids[rf.Model]; ok {
					m, e2 := vm.Model(rf.Model)
					if e2 != nil {
						e = e2
					}
					return UnwrapBucket(m)
				}
			}
			return m
		})

		retNode = xpr.TypedExpression{node, expected, rewritten}

	case xpr.GraphFlow:
		start, e := vm.TypeExpression(node.Start, argument, mdl.Ref{""})
		if e != nil {
			return start, e
		}
		node.Start = start
		mids := make(map[string]struct{})
		mids[start.Actual.Concrete().(mdl.Ref).Model] = struct{}{}
		metaId := vm.MetaModelId()
		for i, flow := range node.Flow {
			from, e := vm.TypeExpression(flow.From, argument, mdl.Ref{metaId})
			if e != nil {
				return from, e
			}
			flow.From = from
			fc, ok := from.Actual.(ConstantModel)
			if !ok {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `graphFlow: from arguments must be constant expressions`,
					Program: xpr.ValueFromExpression(from),
				}
			}
			mids[fc.Value.(val.Ref)[1]] = struct{}{}
			for i, sub := range flow.Forward {
				arg, e := vm.TypeExpression(sub, argument, mdl.Ref{metaId})
				if e != nil {
					return arg, e
				}
				ca, ok := arg.Actual.(ConstantModel)
				if !ok {
					return ZeroTypedExpression, err.CompilationError{
						Problem: `graphFlow: forward arguments must be constant expressions`,
						Program: xpr.ValueFromExpression(arg),
					}
				}
				mids[ca.Value.(val.Ref)[1]] = struct{}{}
				flow.Forward[i] = arg
			}
			for i, sub := range flow.Backward {
				arg, e := vm.TypeExpression(sub, argument, mdl.Ref{metaId})
				if e != nil {
					return arg, e
				}
				ca, ok := arg.Actual.(ConstantModel)
				if !ok {
					return ZeroTypedExpression, err.CompilationError{
						Problem: `graphFlow: backward arguments must be constant expressions`,
						Program: xpr.ValueFromExpression(arg),
					}
				}
				mids[ca.Value.(val.Ref)[1]] = struct{}{}
				flow.Backward[i] = arg
			}
			node.Flow[i] = flow
		}
		strct := mdl.NewStruct(len(mids))
		for k, _ := range mids {
			model, e := vm.Model(k)
			if e != nil {
				return ZeroTypedExpression, e
			}
			strct.Set(k, mdl.Map{model})
		}
		retNode = xpr.TypedExpression{node, expected, strct}

	case xpr.Slice:

		value, e := vm.TypeExpression(node.Value, argument, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		offset, e := vm.TypeExpression(node.Offset, argument, Int64Model)
		if e != nil {
			return offset, e
		}
		node.Offset = offset

		length, e := vm.TypeExpression(node.Length, argument, Int64Model)
		if e != nil {
			return length, e
		}
		node.Length = length

		if cv, ok := value.Actual.(ConstantModel); ok {
			if co, ok := offset.Actual.(ConstantModel); ok {
				if cl, ok := length.Actual.(ConstantModel); ok {
					_, _, _ = cv, co, cl // TODO: constant optimization
				}
			}
		}

		retNode = xpr.TypedExpression{node, expected, UnwrapConstant(value.Actual)}

	case xpr.SearchAllRegex:

		value, e := vm.TypeExpression(node.Value, argument, StringModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		regex, e := vm.TypeExpression(node.Regex, argument, StringModel)
		if e != nil {
			return regex, e
		}
		node.Regex = regex

		cr, ok := regex.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchAllRegex: regex argument must be constant expression`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		if _, e := regexp.Compile(string(cr.Value.(val.String))); e != nil {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchAllRegex: regex does not compile`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		multiLine, e := vm.TypeExpression(node.MultiLine, argument, BoolModel)
		if e != nil {
			return multiLine, e
		}
		node.MultiLine = multiLine

		if _, ok := multiLine.Actual.(ConstantModel); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchAllRegex: multiLine argument must be constant expression`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		caseInsensitive, e := vm.TypeExpression(node.CaseInsensitive, argument, BoolModel)
		if e != nil {
			return caseInsensitive, e
		}
		node.CaseInsensitive = caseInsensitive

		if _, ok := caseInsensitive.Actual.(ConstantModel); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchAllRegex: caseInsensitive argument must be constant expression compile`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		retNode = xpr.TypedExpression{node, expected, mdl.List{Int64Model}}

	case xpr.SearchRegex:

		value, e := vm.TypeExpression(node.Value, argument, StringModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		regex, e := vm.TypeExpression(node.Regex, argument, StringModel)
		if e != nil {
			return regex, e
		}
		node.Regex = regex

		cr, ok := regex.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchRegex: regex argument must be constant expression`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		if _, e := regexp.Compile(string(cr.Value.(val.String))); e != nil {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchRegex: regex does not compile`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		multiLine, e := vm.TypeExpression(node.MultiLine, argument, BoolModel)
		if e != nil {
			return multiLine, e
		}
		node.MultiLine = multiLine

		if _, ok := multiLine.Actual.(ConstantModel); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchRegex: multiLine argument must be constant expression`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		caseInsensitive, e := vm.TypeExpression(node.CaseInsensitive, argument, BoolModel)
		if e != nil {
			return caseInsensitive, e
		}
		node.CaseInsensitive = caseInsensitive

		if _, ok := caseInsensitive.Actual.(ConstantModel); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `searchRegex: caseInsensitive argument must be constant expression compile`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		retNode = xpr.TypedExpression{node, expected, Int64Model}

	case xpr.MatchRegex:

		value, e := vm.TypeExpression(node.Value, argument, StringModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		regex, e := vm.TypeExpression(node.Regex, argument, StringModel)
		if e != nil {
			return regex, e
		}
		node.Regex = regex

		cr, ok := regex.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `matchRegex: regex argument must be constant expression`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		if _, e := regexp.Compile(string(cr.Value.(val.String))); e != nil {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `matchRegex: regex does not compile`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		multiLine, e := vm.TypeExpression(node.MultiLine, argument, BoolModel)
		if e != nil {
			return multiLine, e
		}
		node.MultiLine = multiLine

		if _, ok := multiLine.Actual.(ConstantModel); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `matchRegex: multiLine argument must be constant expression`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		caseInsensitive, e := vm.TypeExpression(node.CaseInsensitive, argument, BoolModel)
		if e != nil {
			return caseInsensitive, e
		}
		node.CaseInsensitive = caseInsensitive

		if _, ok := caseInsensitive.Actual.(ConstantModel); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `matchRegex: caseInsensitive argument must be constant expression compile`,
				Program: xpr.ValueFromExpression(regex),
			}
		}

		retNode = xpr.TypedExpression{node, expected, BoolModel}

	case xpr.AssertModelRef:

		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		ref, e := vm.TypeExpression(node.Ref, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return ref, e
		}
		node.Ref = ref

		cr, ok := ref.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `assertModelRef: ref argument must be constant expression`,
				Program: xpr.ValueFromExpression(ref),
			}
		}

		mid := cr.Value.(val.Ref)[1]
		model, e := vm.Model(mid)
		if e != nil {
			return ZeroTypedExpression, e
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.SwitchModelRef:
		value, e := vm.TypeExpression(node.Value, argument, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		// NOTE: checking for object persistence in $value is futile.

		deflt, e := vm.TypeExpression(node.Default, argument, AnyModel)
		if e != nil {
			return deflt, e
		}
		node.Default = deflt

		retModel := deflt.Actual
		metaId := vm.MetaModelId()

		for i, caze := range node.Cases {

			match, e := vm.TypeExpression(caze.Match, argument, mdl.Ref{metaId})
			if e != nil {
				return match, e
			}
			caze.Match = match

			cm, ok := match.Actual.(ConstantModel)
			if !ok {
				return ZeroTypedExpression, err.CompilationError{
					Problem: `switchModelRef: match arguments must be constant expressions`,
					Program: xpr.ValueFromExpression(match),
				}
			}

			mid := cm.Value.(val.Ref)[1]
			m, e := vm.Model(mid)
			if e != nil {
				return ZeroTypedExpression, e
			}

			retrn, e := vm.TypeExpression(caze.Return, m, AnyModel)
			if e != nil {
				return retrn, e
			}
			caze.Return = retrn

			retModel = mdl.Either(retModel, retrn.Actual.Concrete(), nil)

			node.Cases[i] = caze
		}

		retNode = xpr.TypedExpression{node, expected, retModel}

	case xpr.CreateMultiple:
		in, e := vm.TypeExpression(node.In, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return in, e
		}
		node.In = in
		ci, ok := in.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `createMultiple: model argument must be constant expression`,
				Program: xpr.ValueFromExpression(in),
			}
		}
		mid := ci.Value.(val.Ref)[1]
		subExpect := mdl.Model(nil)
		{
			m, e := vm.Model(mid)
			if e != nil {
				return ZeroTypedExpression, e
			}
			subExpect = UnwrapBucket(m)
		}
		for k, sub := range node.Values {
			arg, e := vm.TypeExpression(sub, argument, subExpect)
			if e != nil {
				return arg, e
			}
			node.Values[k] = arg
		}
		model := mdl.Map{mdl.Ref{mid}}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Field:

		name, e := vm.TypeExpression(node.Name, argument, StringModel)
		if e != nil {
			return name, e
		}
		node.Name = name

		cn, ok := name.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `field: name argument must be constant expression`,
				Program: xpr.ValueFromExpression(name),
			}
		}

		field := string(cn.Value.(val.String))

		subExpect := mdl.NewStruct(1)
		subExpect.Set(field, expected)

		value, e := vm.TypeExpression(node.Value, argument, subExpect)
		if e != nil {
			return value, e
		}
		node.Value = value

		model := value.Actual.Concrete().(mdl.Struct).Field(field)
		if cv, ok := value.Actual.(ConstantModel); ok {
			model = ConstantModel{model, cv.Value.(val.Struct).Field(field)}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Key:
		name, e := vm.TypeExpression(node.Name, argument, StringModel)
		if e != nil {
			return name, e
		}
		node.Name = name
		value, e := vm.TypeExpression(node.Value, argument, mdl.Map{expected})
		if e != nil {
			return value, e
		}
		node.Value = value
		model := mdl.Model(mdl.Optional{value.Actual.Concrete().(mdl.Map).Elements})
		if cv, ok := value.Actual.(ConstantModel); ok {
			if cn, ok := name.Actual.(ConstantModel); ok {
				ov := cv.Value.(val.Map).Key(string(cn.Value.(val.String)))
				if ov == nil {
					ov = val.Null
				}
				model = ConstantModel{model, ov}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Index: // TODO: Index for list access?
		number, e := vm.TypeExpression(node.Number, argument, Int64Model)
		if e != nil {
			return number, e
		}
		node.Number = number
		cn, ok := number.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `index: number argument must be constant expression`,
				Program: xpr.ValueFromExpression(number),
			}
		}
		index := int(cn.Value.(val.Int64))
		arity := index + 1
		if at, ok := argument.(mdl.Tuple); ok {
			arity = len(at)
		}
		subExpect := make(mdl.Tuple, arity, arity)
		for i, _ := range subExpect {
			subExpect[i] = AnyModel
		}
		value, e := vm.TypeExpression(node.Value, argument, &subExpect)
		if e != nil {
			return value, e
		}
		node.Value = value
		model := value.Actual.Concrete().(mdl.Tuple)[index]
		if cv, ok := value.Actual.(ConstantModel); ok {
			model = ConstantModel{model, cv.Value.(val.Tuple)[index]}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Referred:
		in, e := vm.TypeExpression(node.In, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return in, e
		}
		node.In = in
		ci, ok := in.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `referred: model argument must be constant expression`,
				Program: xpr.ValueFromExpression(in),
			}
		}
		from, e := vm.TypeExpression(node.From, argument, mdl.Ref{""})
		if e != nil {
			return from, e
		}
		node.From = from
		retNode = xpr.TypedExpression{node, expected, mdl.List{mdl.Ref{ci.Value.(val.Ref)[1]}}}

	case xpr.RelocateRef:

		ref, e := vm.TypeExpression(node.Ref, argument, mdl.Ref{})
		if e != nil {
			return ref, e
		}
		node.Ref = ref

		model, e := vm.TypeExpression(node.Model, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return ref, e
		}
		node.Model = model

		cm, ok := model.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `relocateRef: model argument must be constant expression`,
				Program: xpr.ValueFromExpression(model),
			}
		}

		mid := cm.Value.(val.Ref)[1]

		if cr, ok := ref.Actual.(ConstantModel); ok {
			// TODO: constant optimization
			_ = cr
		}

		retNode = xpr.TypedExpression{node, expected, mdl.Ref{mid}}

	case xpr.Referrers:
		in, e := vm.TypeExpression(node.In, argument, mdl.Ref{vm.MetaModelId()})
		if e != nil {
			return in, e
		}
		node.In = in
		ci, ok := in.Actual.(ConstantModel)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `referrers: in argument must be constant expression`,
				Program: xpr.ValueFromExpression(in),
			}
		}
		of, e := vm.TypeExpression(node.Of, argument, mdl.Ref{""})
		if e != nil {
			return of, e
		}
		node.Of = of
		retNode = xpr.TypedExpression{node, expected, mdl.List{mdl.Ref{ci.Value.(val.Ref)[1]}}}

	case xpr.ConcatLists:
		lhs, e := vm.TypeExpression(node[0], argument, mdl.List{AnyModel})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, mdl.List{AnyModel})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		// TODO: constant optimization
		model := mdl.Either(lhs.Actual.Unwrap(), rhs.Actual.Unwrap(), nil)
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.After:
		lhs, e := vm.TypeExpression(node[0], argument, DateTimeModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, DateTimeModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		model := mdl.Model(BoolModel)
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value.(val.DateTime), rc.Value.(val.DateTime)
				model = ConstantModel{model, val.Bool(lv.Time.After(rv.Time))}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Before:
		lhs, e := vm.TypeExpression(node[0], argument, DateTimeModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, DateTimeModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		model := mdl.Model(BoolModel)
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value.(val.DateTime), rc.Value.(val.DateTime)
				model = ConstantModel{model, val.Bool(lv.Time.Before(rv.Time))}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Equal:
		lhs, e := vm.TypeExpression(node[0], argument, AnyModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, AnyModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		model := mdl.Model(BoolModel)
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value, rc.Value
				model = ConstantModel{model, val.Bool(lv.Equals(rv))}
			}
		}
		// TODO: compare models (one might be an Or, or Optional of the other)
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Greater:
		lhs, e := vm.TypeExpression(node[0], argument, NumericModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, NumericModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		if lac, rac := lhs.Actual.Concrete(), rhs.Actual.Concrete(); lac != rac {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `greater: comparing distinct types`,
				Program: xpr.ValueFromExpression(rhs),
			}
		}
		model := mdl.Model(BoolModel)
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				out := val.Bool(false)
				lv, rv := lc.Value, rc.Value
				switch lv := lv.(type) {
				case val.Int8:
					out = lv > rv.(val.Int8)
				case val.Int16:
					out = lv > rv.(val.Int16)
				case val.Int32:
					out = lv > rv.(val.Int32)
				case val.Int64:
					out = lv > rv.(val.Int64)
				case val.Uint8:
					out = lv > rv.(val.Uint8)
				case val.Uint16:
					out = lv > rv.(val.Uint16)
				case val.Uint32:
					out = lv > rv.(val.Uint32)
				case val.Uint64:
					out = lv > rv.(val.Uint64)
				case val.Float:
					out = lv > rv.(val.Float)
				default:
					panic(fmt.Sprintf("greater(%T, %T)\n", lv, rv))
				}
				model = ConstantModel{model, out}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Less:
		lhs, e := vm.TypeExpression(node[0], argument, NumericModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, NumericModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		if lac, rac := lhs.Actual.Concrete(), rhs.Actual.Concrete(); lac != rac {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `less: comparing distinct types`,
				Program: xpr.ValueFromExpression(rhs),
			}
		}
		model := mdl.Model(BoolModel)
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				out := val.Bool(false)
				lv, rv := lc.Value, rc.Value
				switch lv := lv.(type) {
				case val.Int8:
					out = lv < rv.(val.Int8)
				case val.Int16:
					out = lv < rv.(val.Int16)
				case val.Int32:
					out = lv < rv.(val.Int32)
				case val.Int64:
					out = lv < rv.(val.Int64)
				case val.Uint8:
					out = lv < rv.(val.Uint8)
				case val.Uint16:
					out = lv < rv.(val.Uint16)
				case val.Uint32:
					out = lv < rv.(val.Uint32)
				case val.Uint64:
					out = lv < rv.(val.Uint64)
				case val.Float:
					out = lv < rv.(val.Float)
				default:
					panic(fmt.Sprintf("less(%T, %T)\n", lv, rv))
				}
				model = ConstantModel{model, out}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Add:
		lhs, e := vm.TypeExpression(node[0], argument, NumericModel)
		if e != nil {
			return lhs, err.LiftArgumentError(e).AppendPath(err.NewFuncArgPathElement("add", 1))
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, NumericModel)
		if e != nil {
			return rhs, err.LiftArgumentError(e).AppendPath(err.NewFuncArgPathElement("add", 2))
		}
		node[1] = rhs
		if lac, rac := lhs.Actual.Concrete(), rhs.Actual.Concrete(); lac != rac {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `add: distinct types`,
				Program: xpr.ValueFromExpression(rhs),
			}
		}
		model := lhs.Actual.Concrete()
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value, rc.Value
				switch lv := lv.(type) {
				case val.Int8:
					model = ConstantModel{model, lv + rv.(val.Int8)}
				case val.Int16:
					model = ConstantModel{model, lv + rv.(val.Int16)}
				case val.Int32:
					model = ConstantModel{model, lv + rv.(val.Int32)}
				case val.Int64:
					model = ConstantModel{model, lv + rv.(val.Int64)}
				case val.Uint8:
					model = ConstantModel{model, lv + rv.(val.Uint8)}
				case val.Uint16:
					model = ConstantModel{model, lv + rv.(val.Uint16)}
				case val.Uint32:
					model = ConstantModel{model, lv + rv.(val.Uint32)}
				case val.Uint64:
					model = ConstantModel{model, lv + rv.(val.Uint64)}
				case val.Float:
					model = ConstantModel{model, lv + rv.(val.Float)}
				default:
					panic(fmt.Sprintf("add(%T, %T)\n", lv, rv))
				}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Subtract:
		lhs, e := vm.TypeExpression(node[0], argument, NumericModel)
		if e != nil {
			return lhs, err.LiftArgumentError(e).AppendPath(err.NewFuncArgPathElement("subtract", 1))
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, NumericModel)
		if e != nil {
			return rhs, err.LiftArgumentError(e).AppendPath(err.NewFuncArgPathElement("subtract", 2))
		}
		node[1] = rhs
		if lac, rac := lhs.Actual.Concrete(), rhs.Actual.Concrete(); lac != rac {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `subtract: distinct types`,
				Program: xpr.ValueFromExpression(rhs),
			}
		}
		model := lhs.Actual.Concrete()
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value, rc.Value
				switch lv := lv.(type) {
				case val.Int8:
					model = ConstantModel{model, lv - rv.(val.Int8)}
				case val.Int16:
					model = ConstantModel{model, lv - rv.(val.Int16)}
				case val.Int32:
					model = ConstantModel{model, lv - rv.(val.Int32)}
				case val.Int64:
					model = ConstantModel{model, lv - rv.(val.Int64)}
				case val.Uint8:
					model = ConstantModel{model, lv - rv.(val.Uint8)}
				case val.Uint16:
					model = ConstantModel{model, lv - rv.(val.Uint16)}
				case val.Uint32:
					model = ConstantModel{model, lv - rv.(val.Uint32)}
				case val.Uint64:
					model = ConstantModel{model, lv - rv.(val.Uint64)}
				case val.Float:
					model = ConstantModel{model, lv - rv.(val.Float)}
				default:
					panic(fmt.Sprintf("subtract(%T, %T)\n", lv, rv))
				}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Multiply:
		lhs, e := vm.TypeExpression(node[0], argument, NumericModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, NumericModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		if lac, rac := lhs.Actual.Concrete(), rhs.Actual.Concrete(); lac != rac {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `multiply: distinct types`,
				Program: xpr.ValueFromExpression(rhs),
			}
		}
		model := lhs.Actual.Concrete()
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value, rc.Value
				switch lv := lv.(type) {
				case val.Int8:
					model = ConstantModel{model, lv * rv.(val.Int8)}
				case val.Int16:
					model = ConstantModel{model, lv * rv.(val.Int16)}
				case val.Int32:
					model = ConstantModel{model, lv * rv.(val.Int32)}
				case val.Int64:
					model = ConstantModel{model, lv * rv.(val.Int64)}
				case val.Uint8:
					model = ConstantModel{model, lv * rv.(val.Uint8)}
				case val.Uint16:
					model = ConstantModel{model, lv * rv.(val.Uint16)}
				case val.Uint32:
					model = ConstantModel{model, lv * rv.(val.Uint32)}
				case val.Uint64:
					model = ConstantModel{model, lv * rv.(val.Uint64)}
				case val.Float:
					model = ConstantModel{model, lv * rv.(val.Float)}
				default:
					panic(fmt.Sprintf("multiply(%T, %T)\n", lv, rv))
				}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Divide:
		lhs, e := vm.TypeExpression(node[0], argument, NumericModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], argument, NumericModel)
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		if lac, rac := lhs.Actual.Concrete(), rhs.Actual.Concrete(); lac != rac {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `divide: distinct types`,
				Program: xpr.ValueFromExpression(rhs),
			}
		}
		model := lhs.Actual.Concrete()
		if lc, ok := lhs.Actual.(ConstantModel); ok {
			if rc, ok := rhs.Actual.(ConstantModel); ok {
				lv, rv := lc.Value, rc.Value
				switch lv := lv.(type) {
				case val.Int8:
					model = ConstantModel{model, lv / rv.(val.Int8)}
				case val.Int16:
					model = ConstantModel{model, lv / rv.(val.Int16)}
				case val.Int32:
					model = ConstantModel{model, lv / rv.(val.Int32)}
				case val.Int64:
					model = ConstantModel{model, lv / rv.(val.Int64)}
				case val.Uint8:
					model = ConstantModel{model, lv / rv.(val.Uint8)}
				case val.Uint16:
					model = ConstantModel{model, lv / rv.(val.Uint16)}
				case val.Uint32:
					model = ConstantModel{model, lv / rv.(val.Uint32)}
				case val.Uint64:
					model = ConstantModel{model, lv / rv.(val.Uint64)}
				case val.Float:
					model = ConstantModel{model, lv / rv.(val.Float)}
				default:
					panic(fmt.Sprintf("divide(%T, %T)\n", lv, rv))
				}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.And:

		constants := make([]val.Bool, 0, len(node))
		for i, sub := range node {
			arg, e := vm.TypeExpression(sub, argument, BoolModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constants = append(constants, ca.Value.(val.Bool))
			}
		}
		model := mdl.Model(BoolModel)
		{
			out := val.Bool(true)
			for _, b := range constants {
				if !b {
					out = false
				}
			}
			if !out || len(constants) == len(node) {
				model = ConstantModel{model, out}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Or:

		constants := make([]val.Bool, 0, len(node))
		for i, sub := range node {
			arg, e := vm.TypeExpression(sub, argument, BoolModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constants = append(constants, ca.Value.(val.Bool))
			}
		}
		model := mdl.Model(BoolModel)
		{
			out := val.Bool(false)
			for _, b := range constants {
				if b {
					out = true
				}
			}
			if out || len(constants) == len(node) {
				model = ConstantModel{model, out}
			}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.SwitchCase:

		if len(node.Cases) == 0 {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`switchCase: zero cases specified`),
				// TODO: program
			}
		}

		subExpect := mdl.NewUnion(len(node.Cases))
		for k, _ := range node.Cases {
			subExpect.Set(k, AnyModel)
		}

		value, e := vm.TypeExpression(node.Value, argument, subExpect)
		if e != nil {
			return value, e
		}
		node.Value = value

		valueModel := value.Actual.Concrete().(mdl.Union)
		model := mdl.Model(nil)

		for k, caze := range node.Cases {
			subNode, e := vm.TypeExpression(caze, valueModel.Case(k), AnyModel)
			if e != nil {
				return subNode, e
			}
			node.Cases[k] = subNode
			if model == nil {
				model = subNode.Actual.Concrete() // TODO: .Concrete() correct?
			} else {
				model = mdl.Either(model, subNode.Actual.Concrete(), nil)
			}
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.MemSort:

		value, e := vm.TypeExpression(node.Value, argument, mdl.List{AnyModel})
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Value = value

		valueModel := value.Actual.Concrete().(mdl.List)

		expression, e := vm.TypeExpression(node.Expression, valueModel.Elements, SortableModel)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Expression = expression

		if _, ok := expression.Actual.Concrete().(mdl.Any); ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`memSort: expression must return unambiguous type`),
				Program: xpr.ValueFromExpression(expression),
			}
		}

		retNode = xpr.TypedExpression{node, expected, UnwrapConstant(value.Actual)}

	case xpr.MapSet:

		value, e := vm.TypeExpression(node.Value, argument, mdl.Set{AnyModel})
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Value = value

		valueModel := value.Actual.Concrete().(mdl.Set)

		expression, e := vm.TypeExpression(node.Expression, valueModel.Elements, AnyModel)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Expression = expression

		retNode = xpr.TypedExpression{node, expected, mdl.Set{expression.Actual}}

	default:
		panic(fmt.Sprintf("unhandled case: %T", node))

	}

	return checkTypedNode(retNode)
}

func checkTypedNode(n xpr.TypedExpression) (xpr.TypedExpression, err.Error) {

	// ASSUMPTION: expected is never a BucketModel

	am, em := n.Actual, n.Expected

	if e := checkType(am, em); e != nil { // means am < em
		return ZeroTypedExpression, err.CompilationError{
			Problem: `type checking failed`,
			Program: xpr.ValueFromExpression(n),
			Child_:  e,
		}
	}

	return n, nil
}
