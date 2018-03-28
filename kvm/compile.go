// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"karma.run/kvm/inst"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"regexp"
)

func (vm VirtualMachine) CompileFunction(f xpr.TypedFunction) inst.Sequence {

	expressions := f.Expressions()

	instructions := make(inst.Sequence, 0, (len(expressions) * 2))

	for _, p := range f.Parameters() {
		instructions = append(instructions, inst.Define(p), inst.Pop{})
	}

	last := len(expressions) - 1

	for i, x := range expressions {
		instructions = vm.CompileExpression(x.(xpr.TypedExpression), instructions)
		if i < last {
			instructions = append(instructions, inst.Pop{}) // discard intermediate stack values
		}
	}

	return instructions
}

func (vm VirtualMachine) CompileExpression(typed xpr.TypedExpression, prev inst.Sequence) inst.Sequence {

	if prev == nil {
		prev = make(inst.Sequence, 0, 64)
	}

	if ca, ok := typed.Actual.(ConstantModel); ok {
		return append(prev, inst.Constant{ca.Value})
	}

	switch node := typed.Expression.(type) {

	case xpr.Zero:
		panic("vm.CompileExpression: uneliminated xpr.Zero") // should be gone by now

	case xpr.ModelOf:
		panic("vm.CompileExpression: uneliminated xpr.ModelOf") // should be gone by now

	case xpr.Define:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Define(node.Name))

	case xpr.Scope:
		return append(prev, inst.Scope(node))

	case xpr.GtFloat:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterFloat{})

	case xpr.GtInt64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterInt64{})

	case xpr.GtInt32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterInt32{})

	case xpr.GtInt16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterInt16{})

	case xpr.GtInt8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterInt8{})

	case xpr.GtUint64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterUint64{})

	case xpr.GtUint32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterUint32{})

	case xpr.GtUint16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterUint16{})

	case xpr.GtUint8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.GreaterUint8{})

	case xpr.LtFloat:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessFloat{})

	case xpr.LtInt64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessInt64{})

	case xpr.LtInt32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessInt32{})

	case xpr.LtInt16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessInt16{})

	case xpr.LtInt8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessInt8{})

	case xpr.LtUint64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessUint64{})

	case xpr.LtUint32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessUint32{})

	case xpr.LtUint16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessUint16{})

	case xpr.LtUint8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.LessUint8{})

	case xpr.MulFloat:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyFloat{})

	case xpr.MulInt64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyInt64{})

	case xpr.MulInt32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyInt32{})

	case xpr.MulInt16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyInt16{})

	case xpr.MulInt8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyInt8{})

	case xpr.MulUint64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyUint64{})

	case xpr.MulUint32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyUint32{})

	case xpr.MulUint16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyUint16{})

	case xpr.MulUint8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.MultiplyUint8{})

	case xpr.DivFloat:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideFloat{})

	case xpr.DivInt64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideInt64{})

	case xpr.DivInt32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideInt32{})

	case xpr.DivInt16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideInt16{})

	case xpr.DivInt8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideInt8{})

	case xpr.DivUint64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideUint64{})

	case xpr.DivUint32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideUint32{})

	case xpr.DivUint16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideUint16{})

	case xpr.DivUint8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.DivideUint8{})

	case xpr.SubFloat:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractFloat{})

	case xpr.SubInt64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractInt64{})

	case xpr.SubInt32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractInt32{})

	case xpr.SubInt16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractInt16{})

	case xpr.SubInt8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractInt8{})

	case xpr.SubUint64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractUint64{})

	case xpr.SubUint32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractUint32{})

	case xpr.SubUint16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractUint16{})

	case xpr.SubUint8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.SubtractUint8{})

	case xpr.AddFloat:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddFloat{})

	case xpr.AddInt64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddInt64{})

	case xpr.AddInt32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddInt32{})

	case xpr.AddInt16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddInt16{})

	case xpr.AddInt8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddInt8{})

	case xpr.AddUint64:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddUint64{})

	case xpr.AddUint32:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddUint32{})

	case xpr.AddUint16:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddUint16{})

	case xpr.AddUint8:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.AddUint8{})

	case xpr.CurrentUser:
		return append(prev, inst.CurrentUser{})

	case xpr.Literal:
		return append(prev, inst.Constant{node.Value})

	// case xpr.NewBool:
	// 	return vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)

	// case xpr.NewInt8:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToInt8{})

	// case xpr.NewInt16:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToInt16{})

	// case xpr.NewInt32:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToInt32{})

	// case xpr.NewInt64:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToInt64{})

	// case xpr.NewUint8:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToUint8{})

	// case xpr.NewUint16:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToUint16{})

	// case xpr.NewUint32:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToUint32{})

	// case xpr.NewUint64:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToUint64{})

	// case xpr.NewFloat:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToFloat{})

	// case xpr.NewString:
	// 	prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
	// 	return append(prev, inst.ToString{})

	// case xpr.NewDateTime:
	// 	return vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)

	case xpr.NewRef:
		prev = vm.CompileExpression(node.Id.(xpr.TypedExpression), prev)
		return append(prev, inst.StringToRef{typed.Actual.(mdl.Ref).Model})

	case xpr.PresentOrZero:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.PresentOrConstant{typed.Actual.Zero()})

	case xpr.AllReferrers:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.AllReferrers{})

	case xpr.IsPresent:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.IsPresent{})

	case xpr.AssertPresent:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.AssertPresent{})

	case xpr.Model:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.StringToRef{typed.Actual.(mdl.Ref).Model})

	case xpr.Tag:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Tag{})

	case xpr.All:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.All{})

	case xpr.StringToLower:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.StringToLower{})

	case xpr.ReverseList:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.ReverseList{})

	case xpr.ExtractStrings:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.ExtractStrings{})

	case xpr.Delete:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Delete{})

	case xpr.ResolveAllRefs:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.ResolveAllRefs{})

	case xpr.First:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.First{})

	case xpr.Get:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Deref{})

	case xpr.Length:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Length{})

	case xpr.Not:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Not{})

	case xpr.Metarialize:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Metarialize{})

	case xpr.RefTo:
		prev = vm.CompileExpression(node.Argument.(xpr.TypedExpression), prev)
		return append(prev, inst.Meta{"id"})

	case xpr.JoinStrings:
		prev = vm.CompileExpression(node.Strings.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Separator.(xpr.TypedExpression), prev)
		return append(prev, inst.JoinStrings{})

	case xpr.If:
		prev = vm.CompileExpression(node.Condition.(xpr.TypedExpression), prev)
		return append(prev, inst.If{
			Then: vm.CompileExpression(node.Then.(xpr.TypedExpression), nil),
			Else: vm.CompileExpression(node.Else.(xpr.TypedExpression), nil),
		})

	case xpr.With:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.With{
			vm.CompileFunction(node.Return.(xpr.TypedFunction)),
		})

	case xpr.Update:
		prev = vm.CompileExpression(node.Ref.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.Update{})

	case xpr.Create:
		return append(prev, inst.CreateMultiple{
			Model: typed.Actual.(mdl.Ref).Model,
			Values: map[string]inst.Sequence{
				"self": vm.CompileFunction(node.Value.(xpr.TypedFunction)),
			},
		}, inst.Constant{val.String("self")}, inst.Key{})

	case xpr.InList:
		prev = vm.CompileExpression(node.In.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.InList{})

	case xpr.FilterList:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.Filter{
			vm.CompileFunction(node.Filter.(xpr.TypedFunction)),
		})

	case xpr.AssertCase:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.AssertCase{
			string(node.Case.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)),
		})

	case xpr.IsCase:
		prev = vm.CompileExpression(node.Case.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.IsCase{})

	case xpr.MapMap:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.MapMap{
			vm.CompileFunction(node.Mapping.(xpr.TypedFunction)),
		})

	case xpr.MapList:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.MapList{
			vm.CompileFunction(node.Mapping.(xpr.TypedFunction)),
		})

	case xpr.MapSet:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.MapSet{
			vm.CompileFunction(node.Mapping.(xpr.TypedFunction)),
		})

	case xpr.ReduceList:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Initial.(xpr.TypedExpression), prev)
		return append(prev, inst.ReduceList{
			vm.CompileFunction(node.Reducer.(xpr.TypedFunction)),
		})

	case xpr.ResolveRefs:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		mrefs := make(map[string]struct{}, len(node.Models))
		for _, sub := range node.Models {
			mref := sub.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
			mrefs[mref[1]] = struct{}{}
		}
		return append(prev, inst.ResolveRefs{mrefs})

	case xpr.SetField:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.In.(xpr.TypedExpression), prev)
		return append(prev, inst.SetField{
			string(node.Name.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)),
		})

	case xpr.SetKey:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.In.(xpr.TypedExpression), prev)
		return append(prev, inst.SetKey{
			string(node.Name.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)),
		})

	case xpr.Field:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.Field{node.Name})

	case xpr.Key:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Name.(xpr.TypedExpression), prev)
		return append(prev, inst.Key{})

	case xpr.IndexTuple:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.IndexTuple{
			int(node.Number.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Int64)),
		})

	case xpr.NewList:
		for _, sub := range node {
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
		}
		return append(prev, inst.BuildList{len(node)})

	case xpr.NewSet:
		for _, sub := range node {
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
		}
		return append(prev, inst.BuildSet{len(node)})

	case xpr.NewTuple:
		for _, sub := range node {
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
		}
		return append(prev, inst.BuildTuple{len(node)})

	case xpr.NewMap:
		for k, sub := range node {
			prev = append(prev, inst.Constant{val.String(k)})
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
		}
		return append(prev, inst.BuildMap{len(node)})

	case xpr.NewStruct:
		ks := make([]string, 0, len(node))
		for k, sub := range node {
			ks = append(ks, k)
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
		}
		return append(prev, inst.BuildStruct{ks})

	case xpr.NewUnion:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.BuildUnion{
			string(node.Case.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)),
		})

	case xpr.Referred:
		mref := node.In.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		prev = vm.CompileExpression(node.From.(xpr.TypedExpression), prev)
		return append(prev, inst.Referred{mref[1]})

	case xpr.RelocateRef:
		mref := node.Model.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		prev = vm.CompileExpression(node.Ref.(xpr.TypedExpression), prev)
		return append(prev, inst.RelocateRef{mref[1]})

	case xpr.Referrers:
		mref := node.In.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		prev = vm.CompileExpression(node.Of.(xpr.TypedExpression), prev)
		return append(prev, inst.Referrers{mref[1]})

	case xpr.ConcatLists:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.ConcatLists{})

	case xpr.After:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.After{})

	case xpr.Before:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.Before{})

	case xpr.Equal:
		prev = vm.CompileExpression(node[0].(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node[1].(xpr.TypedExpression), prev)
		return append(prev, inst.Equal{})

	case xpr.And:
		for i, sub := range node {
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
			prev = append(prev, inst.ShortCircuitAnd{
				Arity: len(node),
				Step:  (i + 1),
			})
		}
		return prev

	case xpr.Or:
		for i, sub := range node {
			prev = vm.CompileExpression(sub.(xpr.TypedExpression), prev)
			prev = append(prev, inst.ShortCircuitOr{
				Arity: len(node),
				Step:  (i + 1),
			})
		}
		return prev

	case xpr.CreateMultiple:
		cm := inst.CreateMultiple{
			Model:  node.In.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)[1],
			Values: make(map[string]inst.Sequence, len(node.Values)),
		}
		for k, sub := range node.Values {
			cm.Values[k] = vm.CompileFunction(sub.(xpr.TypedFunction))
		}
		return append(prev, cm)

	case xpr.Slice:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Offset.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Length.(xpr.TypedExpression), prev)
		return append(prev, inst.Slice{})

	case xpr.SearchAllRegex:
		regex := string(node.Regex.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String))
		if node.MultiLine.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?m)` + regex
		}
		if node.CaseInsensitive.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?i)` + regex
		}
		r := regexp.MustCompile(regex) // compilation previously checked
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.SearchAllRegex{r})

	case xpr.SearchRegex:
		regex := string(node.Regex.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String))
		if node.MultiLine.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?m)` + regex
		}
		if node.CaseInsensitive.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?i)` + regex
		}
		r := regexp.MustCompile(regex) // compilation previously checked
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.SearchRegex{r})

	case xpr.MatchRegex:
		regex := string(node.Regex.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String))
		if node.MultiLine.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?m)` + regex
		}
		if node.CaseInsensitive.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?i)` + regex
		}
		r := regexp.MustCompile(regex) // compilation previously checked
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.MatchRegex{r})

	case xpr.AssertModelRef:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.AssertModelRef{
			node.Ref.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)[1],
		})

	case xpr.SwitchModelRef:
		cases := make(map[string]inst.Sequence, len(node.Cases))
		for _, sub := range node.Cases {
			mref := sub.Match.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
			cases[mref[1]] = vm.CompileExpression(sub.Return.(xpr.TypedExpression), nil)
		}
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.SwitchModelRef{
			Cases:   cases,
			Default: vm.CompileExpression(node.Default.(xpr.TypedExpression), nil),
		})

	case xpr.GraphFlow:
		flow := make(map[string]inst.GraphFlowParam, len(node.Flow))
		for _, sub := range node.Flow {

			mref := sub.From.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)

			var param inst.GraphFlowParam
			if subFlow, ok := flow[mref[1]]; ok {
				param = subFlow
			} else {
				param = inst.GraphFlowParam{
					Forward:  make(map[string]struct{}, len(sub.Forward)),
					Backward: make(map[string]struct{}, len(sub.Backward)),
				}
			}

			for _, subSub := range sub.Forward {
				ref := subSub.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
				param.Forward[ref[1]] = struct{}{}
			}
			for _, subSub := range sub.Backward {
				ref := subSub.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
				param.Backward[ref[1]] = struct{}{}
			}

			flow[mref[1]] = param

		}
		prev = vm.CompileExpression(node.Start.(xpr.TypedExpression), prev)
		return append(prev, inst.GraphFlow{flow})

	case xpr.SwitchCase:
		switchCase := make(inst.SwitchCase, len(node.Cases))
		for k, v := range node.Cases {
			switchCase[k] = vm.CompileFunction(v.(xpr.TypedFunction))
		}
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		prev = vm.CompileExpression(node.Default.(xpr.TypedExpression), prev)
		return append(prev, switchCase)

	case xpr.MemSort:
		prev = vm.CompileExpression(node.Value.(xpr.TypedExpression), prev)
		return append(prev, inst.MemSort{
			vm.CompileFunction(node.Order.(xpr.TypedFunction)),
		})

	}
	panic(fmt.Sprintf("unhandled case: %T", typed.Expression))

}
