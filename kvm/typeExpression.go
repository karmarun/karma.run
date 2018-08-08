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
	NullModel     = mdl.Null{}
)

type ModelScope struct {
	parent *ModelScope
	scope  map[string]mdl.Model
}

func NewModelScope() *ModelScope {
	return &ModelScope{nil, make(map[string]mdl.Model)}
}

func (s *ModelScope) GetLocal(k string) (mdl.Model, bool) {
	if s == nil {
		return nil, false
	}
	m, ok := s.scope[k]
	return m, ok
}

func (s *ModelScope) Get(k string) (mdl.Model, bool) {
	if s == nil {
		return nil, false
	}
	if m, ok := s.scope[k]; ok {
		return m, true
	}
	return s.parent.Get(k)
}

func (s *ModelScope) Set(k string, m mdl.Model) {
	s.scope[k] = m
}

func (s *ModelScope) Child() *ModelScope {
	c := NewModelScope()
	c.parent = s
	return c
}

// for debugging purposes only
func (s *ModelScope) Flat() map[string]mdl.Model {
	if s == nil {
		return nil
	}
	out := make(map[string]mdl.Model)
	for k, v := range s.parent.Flat() {
		out[k] = v
	}
	for k, v := range s.scope {
		out[k] = v
	}
	return out
}

var ZeroTypedExpression = xpr.TypedExpression{}
var ZeroTypedFunction = xpr.TypedFunction{}

// scope may be nil, that's fine, scope.Child() will allocate.
func (vm VirtualMachine) TypeFunction(f xpr.Function, scope *ModelScope, expected mdl.Model) (xpr.TypedFunction, err.Error) {
	arity := len(f.Parameters())
	args := make([]mdl.Model, arity, arity)
	return vm.TypeFunctionWithArguments(f, scope, expected, args...)
}

// scope may be nil, that's fine, scope.Child() will allocate.
// nil args will be inferred (as far as possible).
func (vm VirtualMachine) TypeFunctionWithArguments(f xpr.Function, scope *ModelScope, expected mdl.Model, args ...mdl.Model) (xpr.TypedFunction, err.Error) {

	if expected == nil {
		expected = AnyModel
	}

	if _, ok := expected.(BucketModel); ok {
		panic("INVARIANT: expected is never BucketModel")
	}

	if _, ok := expected.(ConstantModel); ok {
		panic("INVARIANT: expected is never ConstantModel")
	}

	params, functionScope := f.Parameters(), scope.Child()

	if len(params) != len(args) {
		return ZeroTypedFunction, err.CompilationError{
			Problem: fmt.Sprintf(`expected function of %d parameters, have %d`, len(args), len(params)),
			Program: xpr.ValueFromFunction(f),
		}
	}

	for i, l := 0, len(params); i < l; i++ {
		functionScope.Set(params[i], args[i])
	}

	exprs := f.Expressions()

	if len(exprs) == 0 {
		return ZeroTypedFunction, err.CompilationError{
			Problem: `missing function body`,
			Program: xpr.ValueFromFunction(f),
		}
	}

	actual := mdl.Model(nil)

	typedExpressions := make([]xpr.Expression, len(exprs), len(exprs))
	for i, x := range exprs {
		subExpect := mdl.Model(AnyModel)
		if i == len(exprs)-1 {
			subExpect = expected
		}
		typed, e := vm.TypeExpression(x, functionScope, subExpect)
		if e != nil {
			return ZeroTypedFunction, e
		}
		typedExpressions[i] = typed
		actual = typed.Actual
	}

	arguments := make([]mdl.Model, len(params), len(params))
	for i, l := 0, len(params); i < l; i++ {
		m, _ := functionScope.Get(params[i])
		arguments[i] = m // may be nil, that's okay
	}

	typedFunction := xpr.TypedFunction{
		Function:  xpr.NewFunction(params, typedExpressions...),
		Arguments: arguments,
		Expected:  expected,
		Actual:    actual,
	}

	return checkTypedFunction(typedFunction)

}

// VirtualMachine.TypeExpression infers, propagates and checks type information in an xpr.Expression tree.
// It returns an equivalent tree, where each xpr.Expression is wrapped in an xpr.TypedExpression.
// Some typing decisions depend on constant expressions. This is why constant values are propagated as well.
// This is done by wrapping the pertinent types and associated values in ContantModel.
// Parameter 'expected' indicates the expected return type of the expression.
func (vm VirtualMachine) TypeExpression(node xpr.Expression, scope *ModelScope, expected mdl.Model) (xpr.TypedExpression, err.Error) {

	if expected == nil {
		expected = AnyModel
	}

	if _, ok := expected.(BucketModel); ok {
		panic("INVARIANT: expected is never BucketModel")
	}

	if _, ok := expected.(ConstantModel); ok {
		panic("INVARIANT: expected is never ConstantModel")
	}

	retNode := ZeroTypedExpression

	switch node := node.(type) {

	case xpr.StringContains:

		stryng, e := vm.TypeExpression(node.String, scope, StringModel)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.String = stryng

		search, e := vm.TypeExpression(node.Search, scope, StringModel)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Search = search

		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.SubstringIndex:

		stryng, e := vm.TypeExpression(node.String, scope, StringModel)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.String = stryng

		search, e := vm.TypeExpression(node.Search, scope, StringModel)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Search = search

		retNode = xpr.TypedExpression{node, expected, mdl.Int64{}}

	case xpr.MemSortFunction:

		list, e := vm.TypeExpression(node.List, scope, mdl.List{AnyModel})
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.List = list

		listModel := list.Actual.Concrete().(mdl.List)

		less, e := vm.TypeFunctionWithArguments(node.Less, scope, BoolModel, listModel.Elements, listModel.Elements)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Less = less

		retNode = xpr.TypedExpression{node, expected, listModel}

	case xpr.FunctionSignature:
		typed, e := vm.TypeFunction(node.Function, scope, expected)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Function = typed
		params := typed.Parameters()
		argStruct := val.NewStruct(len(params))
		for i, param := range params {
			argStruct.Set(param, mdl.ValueFromModel(vm.MetaModelId(), typed.Arguments[i], nil))
		}
		retTuple := val.Tuple{argStruct, mdl.ValueFromModel(vm.MetaModelId(), typed.Actual, nil)}
		return vm.TypeExpression(xpr.Literal{retTuple}, scope, expected)

	case xpr.GtFloat:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Float{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Float{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtInt64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtInt32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtInt16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtInt8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtUint64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtUint32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtUint16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.GtUint8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtFloat:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Float{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Float{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtInt64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtInt32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtInt16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtInt8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtUint64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtUint32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtUint16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.LtUint8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Bool{}}

	case xpr.MulFloat:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Float{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Float{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Float{}}

	case xpr.MulInt64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int64{}}

	case xpr.MulInt32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int32{}}

	case xpr.MulInt16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int16{}}

	case xpr.MulInt8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int8{}}

	case xpr.MulUint64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint64{}}

	case xpr.MulUint32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint32{}}

	case xpr.MulUint16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint16{}}

	case xpr.MulUint8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint8{}}

	case xpr.DivFloat:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Float{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Float{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Float{}}

	case xpr.DivInt64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int64{}}

	case xpr.DivInt32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int32{}}

	case xpr.DivInt16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int16{}}

	case xpr.DivInt8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int8{}}

	case xpr.DivUint64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint64{}}

	case xpr.DivUint32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint32{}}

	case xpr.DivUint16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint16{}}

	case xpr.DivUint8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint8{}}

	case xpr.SubFloat:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Float{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Float{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Float{}}

	case xpr.SubInt64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int64{}}

	case xpr.SubInt32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int32{}}

	case xpr.SubInt16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int16{}}

	case xpr.SubInt8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int8{}}

	case xpr.SubUint64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint64{}}

	case xpr.SubUint32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint32{}}

	case xpr.SubUint16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint16{}}

	case xpr.SubUint8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint8{}}

	case xpr.AddFloat:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Float{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Float{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Float{}}

	case xpr.AddInt64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int64{}}

	case xpr.AddInt32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int32{}}

	case xpr.AddInt16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int16{}}

	case xpr.AddInt8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Int8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Int8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Int8{}}

	case xpr.AddUint64:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint64{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint64{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint64{}}

	case xpr.AddUint32:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint32{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint32{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint32{}}

	case xpr.AddUint16:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint16{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint16{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint16{}}

	case xpr.AddUint8:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.Uint8{})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.Uint8{})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		retNode = xpr.TypedExpression{node, expected, mdl.Uint8{}}

	case xpr.Define:
		if _, ok := scope.GetLocal(node.Name); ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`name already defined in scope: "%s"`, node.Name),
				Program: xpr.ValueFromExpression(node),
			}
		}
		arg, e := vm.TypeExpression(node.Argument, scope, expected)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		scope.Set(node.Name, arg.Actual)
		retNode = xpr.TypedExpression{node, expected, mdl.Null{}} // define returns null

	case xpr.Scope:
		name := string(node)
		model, ok := scope.Get(name)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`undefined name "%s"`, name),
				Program: xpr.ValueFromExpression(node),
			}
		}
		if model == nil {
			model = expected
		} else {
			// note: naively doing mdl.Either would discard ConstantModel information here.
			if e := checkType(model, expected); e != nil {
				model = mdl.Either(model, expected, nil)
				if _, ok := model.(mdl.Any); ok {
					return ZeroTypedExpression, err.CompilationError{
						Problem: fmt.Sprintf(`inconsistent typing of name: "%s"`, name),
						Program: xpr.ValueFromExpression(node),
						Child_:  e,
					}
				}
			}
		}
		scope.Set(name, model)
		retNode = xpr.TypedExpression{node, expected, model}

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

		caze, e := vm.TypeExpression(node.Case, scope, StringModel)
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

		scase := string(cc.Value.(val.String))

		subExpect := mdl.Model(AnyModel)
		if eu, ok := expected.Concrete().(mdl.Union); ok {
			if m, ok := eu.Get(scase); ok {
				subExpect = m
			}
		}

		value, e := vm.TypeExpression(node.Value, scope, subExpect)
		if e != nil {
			return value, e
		}
		node.Value = value

		model := mdl.Model(nil)
		umodel := mdl.NewUnion(1)
		if ca, ok := value.Actual.(ConstantModel); ok {
			umodel.Set(scase, ca.Model)
			model = ConstantModel{umodel, val.Union{scase, ca.Value}}
		} else {
			umodel.Set(scase, value.Actual)
			model = umodel
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.NewList:

		subModel := mdl.Model(nil)
		constant := make(val.List, 0, len(node))

		for i, arg := range node {
			arg, e := vm.TypeExpression(arg, scope, AnyModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			if subModel == nil {
				subModel = arg.Actual
			} else {
				subModel = mdl.Either(subModel.Unwrap(), arg.Actual.Unwrap(), nil)
			}
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constant = append(constant, ca.Value)
			}
		}

		if len(node) == 0 {
			if em, ok := expected.Concrete().(mdl.List); ok {
				subModel = em.Elements
			} else {
				subModel = AnyModel
			}
			retNode = xpr.TypedExpression{node, expected, ConstantModel{mdl.List{subModel}, make(val.List, 0, 0)}}
		} else {
			if len(constant) == len(node) {
				retNode = xpr.TypedExpression{node, expected, ConstantModel{mdl.List{subModel}, constant}}
			} else {
				retNode = xpr.TypedExpression{node, expected, mdl.List{subModel}}
			}
		}

	case xpr.NewMap:

		subModel := mdl.Model(nil)
		constant := val.NewMap(len(node))

		for k, arg := range node {
			arg, e := vm.TypeExpression(arg, scope, AnyModel)
			if e != nil {
				return arg, e
			}
			node[k] = arg
			if subModel == nil {
				subModel = arg.Actual
			} else {
				subModel = mdl.Either(subModel.Unwrap(), arg.Actual.Unwrap(), nil)
			}
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constant.Set(k, ca.Value)
			}
		}

		if len(node) == 0 {
			if em, ok := expected.Concrete().(mdl.Map); ok {
				subModel = em.Elements
			} else {
				subModel = AnyModel
			}
			retNode = xpr.TypedExpression{node, expected, ConstantModel{mdl.Map{subModel}, val.NewMap(0)}}
		} else {
			if constant.Len() == len(node) {
				retNode = xpr.TypedExpression{node, expected, ConstantModel{mdl.List{subModel}, constant}}
			} else {
				retNode = xpr.TypedExpression{node, expected, mdl.List{subModel}}
			}
		}

		retNode = xpr.TypedExpression{node, expected, mdl.Map{subModel}}

	case xpr.NewSet:

		subModel := mdl.Model(nil)
		constant := make(val.Set, len(node))

		for i, arg := range node {
			arg, e := vm.TypeExpression(arg, scope, AnyModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			if subModel == nil {
				subModel = arg.Actual
			} else {
				subModel = mdl.Either(subModel.Unwrap(), arg.Actual.Unwrap(), nil)
			}
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constant[val.Hash(ca.Value, nil).Sum64()] = ca.Value
			}
		}

		if len(node) == 0 {
			if em, ok := expected.Concrete().(mdl.Set); ok {
				subModel = em.Elements
			} else {
				subModel = AnyModel
			}
			retNode = xpr.TypedExpression{node, expected, ConstantModel{mdl.Set{subModel}, make(val.Set, 0)}}
		} else {
			if len(constant) == len(node) {
				retNode = xpr.TypedExpression{node, expected, ConstantModel{mdl.Set{subModel}, constant}}
			} else {
				retNode = xpr.TypedExpression{node, expected, mdl.Set{subModel}}
			}
		}

	case xpr.NewStruct:

		model := mdl.NewStruct(len(node))
		constant := val.NewStruct(len(node))

		for k, arg := range node {
			arg, e := vm.TypeExpression(arg, scope, AnyModel)
			if e != nil {
				return arg, e
			}
			node[k] = arg
			model.Set(k, arg.Actual) // do not Unwrap arg.Actual (ConstantModel is relevant for typechecking)
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constant.Set(k, ca.Value)
			}
		}

		if constant.Len() == len(node) {
			retNode = xpr.TypedExpression{node, expected, ConstantModel{model, constant}}
		} else {
			retNode = xpr.TypedExpression{node, expected, model}
		}

	case xpr.NewTuple:

		model := make(mdl.Tuple, len(node))
		constant := make(val.Tuple, 0, len(node))

		for i, arg := range node {
			arg, e := vm.TypeExpression(arg, scope, AnyModel)
			if e != nil {
				return arg, e
			}
			node[i] = arg
			model[i] = arg.Actual // do not Unwrap arg.Actual (ConstantModel is relevant for typechecking)
			if ca, ok := arg.Actual.(ConstantModel); ok {
				constant = append(constant, ca.Value)
			}
		}

		if len(constant) == len(node) {
			retNode = xpr.TypedExpression{node, expected, ConstantModel{model, constant}}
		} else {
			retNode = xpr.TypedExpression{node, expected, model}
		}

	case xpr.SetField:

		name, e := vm.TypeExpression(node.Name, scope, StringModel)
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

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		in, e := vm.TypeExpression(node.In, scope, mdl.Struct{}) // any struct
		if e != nil {
			return in, e
		}
		node.In = in

		m := in.Actual.Concrete().Copy().(mdl.Struct)
		m.Set(field, value.Actual) // do not Unwrap value.Actual (ConstantModel is relevant for typechecking)

		retNode = xpr.TypedExpression{node, expected, m}

	case xpr.SetKey:
		name, e := vm.TypeExpression(node.Name, scope, StringModel)
		if e != nil {
			return name, e
		}
		node.Name = name

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		in, e := vm.TypeExpression(node.In, scope, mdl.Map{AnyModel})
		if e != nil {
			return in, e
		}
		node.In = in

		retNode = xpr.TypedExpression{node, expected, mdl.Map{value.Actual}}

	case xpr.NewRef:
		marg, e := vm.TypeExpression(node.Model, scope, StringModel)
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

		mid := string(cm.Value.(val.String))

		iarg, e := vm.TypeExpression(node.Id, scope, StringModel)
		if e != nil {
			return iarg, e
		}
		node.Id = iarg

		model := mdl.Model(mdl.Ref{mid})

		if ci, ok := iarg.Actual.(ConstantModel); ok {
			id := string(ci.Value.(val.String))
			model = ConstantModel{model, val.Ref{mid, id}}
		}

		// TODO: check that ID exists (during execution)

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.PresentOrZero:

		arg, e := vm.TypeExpression(node.Argument, scope, AnyModel)
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

		arg, e := vm.TypeExpression(node.Argument, scope, mdl.Ref{})
		if e != nil {
			return arg, e
		}
		node.Argument = arg

		retNode = xpr.TypedExpression{node, expected, mdl.List{mdl.Ref{}}}

	case xpr.IsPresent:
		arg, e := vm.TypeExpression(node.Argument, scope, AnyModel)
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
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.Optional{expected})
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
		arg, e := vm.TypeExpression(node.Argument, scope, StringModel)
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

	case xpr.TagExists:
		arg, e := vm.TypeExpression(node.Argument, scope, StringModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		model := mdl.Model(mdl.Bool{})
		if ca, ok := arg.Actual.(ConstantModel); ok {
			tag := string(ca.Value.(val.String))
			mid := vm.RootBucket.Bucket(definitions.TagBucketBytes).Get([]byte(tag))
			model = ConstantModel{model, val.Bool(mid != nil)}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Tag:
		arg, e := vm.TypeExpression(node.Argument, scope, StringModel)
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
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.Ref{vm.MetaModelId()})
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

		strings, e := vm.TypeExpression(node.Strings, scope, mdl.List{StringModel})
		if e != nil {
			return strings, e
		}
		node.Strings = strings

		separator, e := vm.TypeExpression(node.Separator, scope, StringModel)
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
		arg, e := vm.TypeExpression(node.Argument, scope, StringModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		if ca, ok := arg.Actual.(ConstantModel); ok {
			_ = ca // TODO: constant optimization
		}
		retNode = xpr.TypedExpression{node, expected, StringModel}

	case xpr.ReverseList:
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.List{AnyModel})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		retNode = xpr.TypedExpression{node, expected, arg.Actual.Unwrap()} // get rid of constant wrapper in e.g. reverseList([1,2,3])

	case xpr.ExtractStrings:
		arg, e := vm.TypeExpression(node.Argument, scope, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		if ca, ok := arg.Actual.(ConstantModel); ok {
			_ = ca // TODO: constant optimization
		}
		retNode = xpr.TypedExpression{node, expected, mdl.List{StringModel}}

	case xpr.Delete:
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.Ref{""})
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

		arg, e := vm.TypeExpression(node.Argument, scope, mdl.Any{})
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
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.List{expected})
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
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.Ref{""})
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		mid := arg.Actual.Concrete().(mdl.Ref).Model
		if mid == "" {
			retNode = xpr.TypedExpression{node, expected, BucketModel{Model: AnyModel}}
		} else {
			model, e := vm.Model(mid)
			if e != nil {
				return ZeroTypedExpression, e
			}
			retNode = xpr.TypedExpression{node, expected, model}
		}

	case xpr.Length:
		arg, e := vm.TypeExpression(node.Argument, scope, mdl.List{AnyModel})
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
		arg, e := vm.TypeExpression(node.Argument, scope, BoolModel)
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
		arg, e := vm.TypeExpression(node.Argument, scope, AnyModel)
		if e != nil {
			return arg, e
		}
		node.Argument = arg
		unwrapped := arg.Actual.Transform(UnwrapConstant).Transform(UnwrapBucket)
		modelValue := mdl.ValueFromModel(vm.MetaModelId(), unwrapped, nil)
		retNode = xpr.TypedExpression{node, expected, ConstantModel{vm.MetaModel(), modelValue}}

	case xpr.Metarialize:

		arg, e := vm.TypeExpression(node.Argument, scope, AnyModel)
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

		arg, e := vm.TypeExpression(node.Argument, scope, AnyModel)
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
		condition, e := vm.TypeExpression(node.Condition, scope, BoolModel)
		if e != nil {
			return condition, e
		}
		node.Condition = condition
		then, e := vm.TypeExpression(node.Then, scope, AnyModel)
		if e != nil {
			return then, e
		}
		node.Then = then
		elze, e := vm.TypeExpression(node.Else, scope, AnyModel)
		if e != nil {
			return elze, e
		}
		node.Else = elze
		// TODO: constant condition optimization
		retNode = xpr.TypedExpression{node, expected, mdl.Either(UnwrapConstant(UnwrapBucket(then.Actual)), UnwrapConstant(UnwrapBucket(elze.Actual)), nil)}

	case xpr.With:

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		retrn, e := vm.TypeFunctionWithArguments(node.Return, scope, expected, value.Actual)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Return = retrn

		retNode = xpr.TypedExpression{node, expected, retrn.Actual}

	case xpr.Update:
		ref, e := vm.TypeExpression(node.Ref, scope, mdl.Ref{""})
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
		value, e := vm.TypeExpression(node.Value, scope, subExpect.Model)
		if e != nil {
			return value, e
		}
		node.Value = value
		retNode = xpr.TypedExpression{node, expected, mdl.Ref{mid}}

	case xpr.Create:

		in, e := vm.TypeExpression(node.In, scope, mdl.Ref{vm.MetaModelId()})
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

		subArg := mdl.NewStruct(1)
		subArg.Set("self", mdl.Ref{mid})

		value, e := vm.TypeFunctionWithArguments(node.Value, scope, subExpect.Model, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Value = value

		retNode = xpr.TypedExpression{node, expected, mdl.Ref{mid}}

	case xpr.InList:

		in, e := vm.TypeExpression(node.In, scope, mdl.List{AnyModel})
		if e != nil {
			return in, e
		}
		node.In = in

		subExpect := in.Actual.Concrete().(mdl.List).Elements

		value, e := vm.TypeExpression(node.Value, scope, UnwrapBucket(subExpect))
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

	case xpr.FilterList:
		value, e := vm.TypeExpression(node.Value, scope, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		subArg := value.Actual.Concrete().(mdl.List).Elements
		filter, e := vm.TypeFunctionWithArguments(node.Filter, scope, BoolModel, mdl.Int64{}, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Filter = filter

		retNode = xpr.TypedExpression{node, expected, value.Actual.Unwrap()} // get rid of constant wrapper in e.g. filterList([1,2,3], () => ...)

	case xpr.AssertCase:
		caze, e := vm.TypeExpression(node.Case, scope, StringModel)
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

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
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
		caze, e := vm.TypeExpression(node.Case, scope, StringModel)
		if e != nil {
			return caze, e
		}
		node.Case = caze
		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		if _, ok := value.Actual.Concrete().(mdl.Union); !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `isCase: value argument is not a union`,
				Program: xpr.ValueFromExpression(value),
			}
		}
		node.Value = value
		// TODO: constant optimization
		retNode = xpr.TypedExpression{node, expected, BoolModel}

	case xpr.MapMap:
		value, e := vm.TypeExpression(node.Value, scope, mdl.Map{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		subArg := value.Actual.Concrete().(mdl.Map).Elements
		mapping, e := vm.TypeFunctionWithArguments(node.Mapping, scope, AnyModel, mdl.String{}, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Mapping = mapping

		retNode = xpr.TypedExpression{node, expected, mdl.Map{mapping.Actual}}

	case xpr.MapList:

		value, e := vm.TypeExpression(node.Value, scope, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		subArg := value.Actual.Concrete().(mdl.List).Elements
		mapping, e := vm.TypeFunctionWithArguments(node.Mapping, scope, AnyModel, mdl.Int64{}, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Mapping = mapping

		retNode = xpr.TypedExpression{node, expected, mdl.List{mapping.Actual}}

	case xpr.ReduceList:

		value, e := vm.TypeExpression(node.Value, scope, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		initial, e := vm.TypeExpression(node.Initial, scope, AnyModel)
		if e != nil {
			return initial, e
		}
		node.Initial = initial

		subArg := mdl.Either(
			value.Actual.Concrete().(mdl.List).Elements,
			initial.Actual.Unwrap(),
			nil, // recursion param
		)

		reducer, e := vm.TypeFunctionWithArguments(node.Reducer, scope, AnyModel, subArg, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Reducer = reducer

		retNode = xpr.TypedExpression{node, expected, reducer.Actual}

	case xpr.ResolveRefs:

		mids := make(map[string]struct{}, len(node.Models))

		for i, sub := range node.Models {
			model, e := vm.TypeExpression(sub, scope, mdl.Ref{vm.MetaModelId()})
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

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
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
		start, e := vm.TypeExpression(node.Start, scope, mdl.Ref{""})
		if e != nil {
			return start, e
		}
		node.Start = start
		mids := make(map[string]struct{})
		mids[start.Actual.Concrete().(mdl.Ref).Model] = struct{}{}
		metaId := vm.MetaModelId()
		for i, flow := range node.Flow {
			from, e := vm.TypeExpression(flow.From, scope, mdl.Ref{metaId})
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
				arg, e := vm.TypeExpression(sub, scope, mdl.Ref{metaId})
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
				arg, e := vm.TypeExpression(sub, scope, mdl.Ref{metaId})
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

		value, e := vm.TypeExpression(node.Value, scope, mdl.List{AnyModel})
		if e != nil {
			return value, e
		}
		node.Value = value

		offset, e := vm.TypeExpression(node.Offset, scope, Int64Model)
		if e != nil {
			return offset, e
		}
		node.Offset = offset

		length, e := vm.TypeExpression(node.Length, scope, Int64Model)
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

		value, e := vm.TypeExpression(node.Value, scope, StringModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		regex, e := vm.TypeExpression(node.Regex, scope, StringModel)
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

		multiLine, e := vm.TypeExpression(node.MultiLine, scope, BoolModel)
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

		caseInsensitive, e := vm.TypeExpression(node.CaseInsensitive, scope, BoolModel)
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

		value, e := vm.TypeExpression(node.Value, scope, StringModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		regex, e := vm.TypeExpression(node.Regex, scope, StringModel)
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

		multiLine, e := vm.TypeExpression(node.MultiLine, scope, BoolModel)
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

		caseInsensitive, e := vm.TypeExpression(node.CaseInsensitive, scope, BoolModel)
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

		value, e := vm.TypeExpression(node.Value, scope, StringModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		regex, e := vm.TypeExpression(node.Regex, scope, StringModel)
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

		multiLine, e := vm.TypeExpression(node.MultiLine, scope, BoolModel)
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

		caseInsensitive, e := vm.TypeExpression(node.CaseInsensitive, scope, BoolModel)
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

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		ref, e := vm.TypeExpression(node.Ref, scope, mdl.Ref{vm.MetaModelId()})
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
		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		panic("todo")

	case xpr.CreateMultiple:

		in, e := vm.TypeExpression(node.In, scope, mdl.Ref{vm.MetaModelId()})
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

		subArg := mdl.NewStruct(len(node.Values))
		for k, _ := range node.Values {
			subArg.Set(k, mdl.Ref{mid})
		}

		for k, sub := range node.Values {

			arg, e := vm.TypeFunctionWithArguments(sub, scope, subExpect, subArg)
			if e != nil {
				return ZeroTypedExpression, e
			}

			node.Values[k] = arg
		}

		retNode = xpr.TypedExpression{node, expected, subArg}

	case xpr.Field:

		subExpect := mdl.NewStruct(1)
		subExpect.Set(node.Name, expected)

		value, e := vm.TypeExpression(node.Value, scope, subExpect)
		if e != nil {
			return value, e
		}
		node.Value = value

		model := value.Actual.Concrete().(mdl.Struct).Field(node.Name)
		if cv, ok := value.Actual.(ConstantModel); ok {
			model = ConstantModel{model, cv.Value.(val.Struct).Field(node.Name)}
		}
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Key:

		name, e := vm.TypeExpression(node.Name, scope, StringModel)
		if e != nil {
			return name, e
		}
		node.Name = name

		value, e := vm.TypeExpression(node.Value, scope, mdl.Map{expected})
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

	case xpr.IndexTuple:

		value, e := vm.TypeExpression(node.Value, scope, AnyModel)
		if e != nil {
			return value, e
		}
		node.Value = value

		tupleModel, ok := value.Actual.Concrete().(mdl.Tuple)
		if !ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: `indexTuple: argument not of type tuple`,
			}

		}
		index := int(node.Number)

		if index >= len(tupleModel) {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`indexTuple: argument is tuple of %d elements, requested index %d`, len(tupleModel), index),
			}
		}

		model := tupleModel[index]
		if cv, ok := value.Actual.(ConstantModel); ok {
			model = ConstantModel{model, cv.Value.(val.Tuple)[index]}
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.Referred:
		in, e := vm.TypeExpression(node.In, scope, mdl.Ref{vm.MetaModelId()})
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
		from, e := vm.TypeExpression(node.From, scope, mdl.Ref{""})
		if e != nil {
			return from, e
		}
		node.From = from
		retNode = xpr.TypedExpression{node, expected, mdl.List{mdl.Ref{ci.Value.(val.Ref)[1]}}}

	case xpr.RelocateRef:

		ref, e := vm.TypeExpression(node.Ref, scope, mdl.Ref{})
		if e != nil {
			return ref, e
		}
		node.Ref = ref

		model, e := vm.TypeExpression(node.Model, scope, mdl.Ref{vm.MetaModelId()})
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
		in, e := vm.TypeExpression(node.In, scope, mdl.Ref{vm.MetaModelId()})
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
		of, e := vm.TypeExpression(node.Of, scope, mdl.Ref{""})
		if e != nil {
			return of, e
		}
		node.Of = of
		retNode = xpr.TypedExpression{node, expected, mdl.List{mdl.Ref{ci.Value.(val.Ref)[1]}}}

	case xpr.ConcatLists:
		lhs, e := vm.TypeExpression(node[0], scope, mdl.List{AnyModel})
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, mdl.List{AnyModel})
		if e != nil {
			return rhs, e
		}
		node[1] = rhs
		// TODO: constant optimization
		model := mdl.Either(lhs.Actual.Unwrap(), rhs.Actual.Unwrap(), nil)
		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.After:
		lhs, e := vm.TypeExpression(node[0], scope, DateTimeModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, DateTimeModel)
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
		lhs, e := vm.TypeExpression(node[0], scope, DateTimeModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, DateTimeModel)
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
		lhs, e := vm.TypeExpression(node[0], scope, AnyModel)
		if e != nil {
			return lhs, e
		}
		node[0] = lhs
		rhs, e := vm.TypeExpression(node[1], scope, AnyModel)
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

	case xpr.And:

		constants := make([]val.Bool, 0, len(node))
		for i, sub := range node {
			arg, e := vm.TypeExpression(sub, scope, BoolModel)
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
			arg, e := vm.TypeExpression(sub, scope, BoolModel)
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
				Program: xpr.ValueFromExpression(node),
			}
		}

		subExpect := mdl.NewUnion(len(node.Cases))
		for k, _ := range node.Cases {
			subExpect.Set(k, AnyModel)
		}

		value, e := vm.TypeExpression(node.Value, scope, subExpect)
		if e != nil {
			return value, e
		}
		node.Value = value

		union := value.Actual.Concrete().(mdl.Union)
		model := (mdl.Model)(nil)

		for k, sub := range node.Cases {
			fun, e := vm.TypeFunctionWithArguments(sub, scope, AnyModel, union.Case(k))
			if e != nil {
				return ZeroTypedExpression, e
			}
			if model == nil {
				model = fun.Actual
			} else {
				model = mdl.Either(model.Unwrap(), fun.Actual.Unwrap(), nil)
			}
			node.Cases[k] = fun
		}

		retNode = xpr.TypedExpression{node, expected, model}

	case xpr.MemSort:

		value, e := vm.TypeExpression(node.Value, scope, mdl.List{AnyModel})
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Value = value

		subArg := value.Actual.Concrete().(mdl.List).Elements

		order, e := vm.TypeFunctionWithArguments(node.Order, scope, AnyModel, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Order = order

		if _, ok := order.Actual.Concrete().(mdl.Any); ok {
			return ZeroTypedExpression, err.CompilationError{
				Problem: fmt.Sprintf(`memSort: expression must return unambiguous type`),
				Program: xpr.ValueFromFunction(order),
			}
		}
		retNode = xpr.TypedExpression{node, expected, UnwrapConstant(value.Actual)}

	case xpr.MapSet:

		value, e := vm.TypeExpression(node.Value, scope, mdl.Set{AnyModel})
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Value = value

		subArg := value.Actual.Concrete().(mdl.Set).Elements
		mapping, e := vm.TypeFunctionWithArguments(node.Mapping, scope, AnyModel, subArg)
		if e != nil {
			return ZeroTypedExpression, e
		}
		node.Mapping = mapping

		retNode = xpr.TypedExpression{node, expected, mdl.Set{mapping.Actual}}

	default:
		panic(fmt.Sprintf("unhandled case: %T", node))

	}

	return checkTypedExpression(retNode)
}

func checkTypedExpression(n xpr.TypedExpression) (xpr.TypedExpression, err.Error) {

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

func checkTypedFunction(n xpr.TypedFunction) (xpr.TypedFunction, err.Error) {

	// ASSUMPTION: expected is never a BucketModel

	am, em := n.Actual, n.Expected

	if e := checkType(am, em); e != nil { // means am < em
		return ZeroTypedFunction, err.CompilationError{
			Problem: `type checking failed`,
			Program: xpr.ValueFromFunction(n),
			Child_:  e,
		}
	}

	return n, nil
}

func checkArgumentTypes(f xpr.TypedFunction, args ...mdl.Model) err.Error {

	params := f.Parameters()

	if len(params) != len(args) {
		return err.CompilationError{
			Problem: fmt.Sprintf(`expected function of %d parameters, have %d`, len(args), len(params)),
			Program: xpr.ValueFromFunction(f),
		}
	}

	for i, l := 0, len(params); i < l; i++ {
		am, em := args[i], f.Arguments[i]
		if em == nil {
			continue // argument is unused
		}
		if e := checkType(am, em); e != nil {
			return err.CompilationError{
				Problem: fmt.Sprintf(`function argument type mismatch in parameter "%s"`, params[i]),
				Program: xpr.ValueFromFunction(f),
				Child_:  e,
			}
		}
	}

	return nil

}
