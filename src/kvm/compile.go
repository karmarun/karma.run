// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"kvm/inst"
	"kvm/mdl"
	"kvm/val"
	"kvm/xpr"
	"regexp"
)

// PRECONDITION: $typed obtained from error-less vm.TypeAST call
func (vm VirtualMachine) Compile(typed xpr.TypedExpression) inst.Instruction {

	if ca, ok := typed.Actual.(ConstantModel); ok {
		return inst.Constant{ca.Value}
	}

	switch node := typed.Expression.(type) {
	case xpr.Argument:
		return inst.Identity{}

	case xpr.CurrentUser:
		return inst.CurrentUser{}

	case xpr.Zero:
		panic("vm.Compile: unfolded zero() call")

	case xpr.Literal:
		return inst.Constant{node.Value}

	case xpr.NewBool:
		return vm.Compile(node.Argument.(xpr.TypedExpression))

	case xpr.NewInt8:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToInt8{}}

	case xpr.NewInt16:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToInt16{}}

	case xpr.NewInt32:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToInt32{}}

	case xpr.NewInt64:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToInt64{}}

	case xpr.NewUint8:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToUint8{}}

	case xpr.NewUint16:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToUint16{}}

	case xpr.NewUint32:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToUint32{}}

	case xpr.NewUint64:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToUint64{}}

	case xpr.NewFloat:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		return inst.Sequence{asm, inst.ToFloat{}}

	case xpr.NewString:
		arg := node.Argument.(xpr.TypedExpression)
		asm := vm.Compile(arg)
		switch arg.Actual.Concrete().(type) {
		case mdl.String:
			return asm
		case mdl.Ref:
			return inst.Sequence{asm, inst.StringToRef{}}
		default:
			panic(fmt.Sprintf("%T", arg.Actual.Concrete()))
		}

	case xpr.NewDateTime:
		return vm.Compile(node.Argument.(xpr.TypedExpression))

	case xpr.NewRef:
		return inst.Sequence{vm.Compile(node.Id.(xpr.TypedExpression)), inst.StringToRef{typed.Actual.(mdl.Ref).Model}}

	case xpr.PresentOrZero:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.PresentOrConstant{typed.Actual.Zero()}}

	case xpr.IsPresent:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.IsPresent{}}

	case xpr.AssertPresent:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.AssertPresent{}}

	case xpr.Model:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.StringToRef{typed.Actual.(mdl.Ref).Model}}

	case xpr.Tag:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Tag{}}

	case xpr.All:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.All{}}

	case xpr.JoinStrings:
		return inst.Sequence{vm.Compile(node.Strings.(xpr.TypedExpression)), vm.Compile(node.Separator.(xpr.TypedExpression)), inst.JoinStrings{}}

	case xpr.StringToLower:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.StringToLower{}}

	case xpr.ReverseList:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.ReverseList{}}

	case xpr.ExtractStrings:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.ExtractStrings{}}

	case xpr.Delete:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Delete{}}

	case xpr.ResolveAllRefs:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.ResolveAllRefs{}}

	case xpr.First:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.First{}}

	case xpr.Get:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Deref{}}

	case xpr.Length:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Length{}}

	case xpr.Not:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Not{}}

	case xpr.ModelOf:
		panic("vm.Compile: uneliminated ast.ModelOf")

	case xpr.Metarialize:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Metarialize{}}

	case xpr.RefTo:
		return inst.Sequence{vm.Compile(node.Argument.(xpr.TypedExpression)), inst.Meta{"id"}}

	case xpr.If:
		condition := vm.Compile(node.Condition.(xpr.TypedExpression))
		then := vm.Compile(node.Then.(xpr.TypedExpression))
		elze := vm.Compile(node.Else.(xpr.TypedExpression))
		return inst.Sequence{condition, inst.If{Then: then, Else: elze}}

	case xpr.With:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		retrn := vm.Compile(node.Return.(xpr.TypedExpression))
		return inst.Sequence{value, inst.PopToInput{}, retrn}

	case xpr.Update:
		ref := vm.Compile(node.Ref.(xpr.TypedExpression))
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		return inst.Sequence{ref, value, inst.Update{}}

	case xpr.Create:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		return inst.Sequence{
			inst.Constant{val.String("self")},
			value,
			inst.BuildMap{1},
			inst.CreateMultiple{typed.Actual.(mdl.Ref).Model},
			inst.Constant{val.String("self")},
			inst.Key{},
		}

	case xpr.InList:
		in := vm.Compile(node.In.(xpr.TypedExpression))
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		return inst.Sequence{in, value, inst.InList{}}

	case xpr.Filter:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		expression := vm.Compile(node.Expression.(xpr.TypedExpression))
		return inst.Sequence{value, inst.Filter{expression}}

	case xpr.AssertCase:
		caze := node.Case.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		return inst.Sequence{value, inst.AssertCase{string(caze)}}

	case xpr.IsCase:
		caze := vm.Compile(node.Case.(xpr.TypedExpression))
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		return inst.Sequence{caze, value, inst.IsCase{}}

	case xpr.MapMap:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		expression := vm.Compile(node.Expression.(xpr.TypedExpression))
		return inst.Sequence{value, inst.MapMap{expression}}

	case xpr.MapList:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		expression := vm.Compile(node.Expression.(xpr.TypedExpression))
		return inst.Sequence{value, inst.MapList{expression}}

	case xpr.ReduceList:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		expression := vm.Compile(node.Expression.(xpr.TypedExpression))
		return inst.Sequence{value, inst.ReduceList{expression}}

	case xpr.ResolveRefs:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		mrefs := make(map[string]struct{}, len(node.Models))
		for _, sub := range node.Models {
			mref := sub.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
			mrefs[mref[1]] = struct{}{}
		}
		return inst.Sequence{value, inst.ResolveRefs{mrefs}}

	case xpr.SetField:
		name := node.Name.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		in := vm.Compile(node.In.(xpr.TypedExpression))
		return inst.Sequence{value, in, inst.SetField{string(name)}}
	case xpr.SetKey:
		name := node.Name.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		in := vm.Compile(node.In.(xpr.TypedExpression))
		return inst.Sequence{value, in, inst.SetKey{string(name)}}
	case xpr.Field:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		name := node.Name.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)
		return inst.Sequence{value, inst.Field{string(name)}}

	case xpr.Key:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		name := vm.Compile(node.Name.(xpr.TypedExpression))
		return inst.Sequence{value, name, inst.Key{}}

	case xpr.Index:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		number := node.Number.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Int64)
		return inst.Sequence{value, inst.IndexTuple{int(number)}}

	case xpr.NewList:
		is := make(inst.Sequence, 0, len(node)+1)
		for _, sub := range node {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, arg)
		}
		return append(is, inst.BuildList{len(node)})

	case xpr.NewTuple:
		is := make(inst.Sequence, 0, len(node)+1)
		for _, sub := range node {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, arg)
		}
		return append(is, inst.BuildTuple{len(node)})

	case xpr.NewMap:
		is := make(inst.Sequence, 0, len(node)*2+1)
		for k, sub := range node {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, inst.Constant{val.String(k)}, arg)
		}
		return append(is, inst.BuildMap{len(node)})

	case xpr.NewStruct:
		is := make(inst.Sequence, 0, len(node)+1)
		ks := make([]string, 0, len(node))
		for k, sub := range node {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, arg)
			ks = append(ks, k)
		}
		return append(is, inst.BuildStruct{ks})

	case xpr.NewUnion:
		caze := node.Case.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String)
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		return inst.Sequence{value, inst.BuildUnion{string(caze)}}

	case xpr.Referred:
		mref := node.In.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		from := vm.Compile(node.From.(xpr.TypedExpression))
		return inst.Sequence{from, inst.Referred{mref[1]}}

	case xpr.RelocateRef:
		mref := node.Model.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		ref := vm.Compile(node.Ref.(xpr.TypedExpression))
		return inst.Sequence{ref, inst.RelocateRef{mref[1]}}

	case xpr.Referrers:
		mref := node.In.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		of := vm.Compile(node.Of.(xpr.TypedExpression))
		return inst.Sequence{of, inst.Referrers{mref[1]}}

	case xpr.ConcatLists:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		return inst.Sequence{lhs, rhs, inst.ConcatLists{}}

	case xpr.After:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		return inst.Sequence{lhs, rhs, inst.After{}}

	case xpr.Before:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		return inst.Sequence{lhs, rhs, inst.Before{}}

	case xpr.Equal:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		return inst.Sequence{lhs, rhs, inst.Equal{}}

	case xpr.Greater:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		switch node[0].(xpr.TypedExpression).Actual.Concrete().(type) {
		case mdl.Float:
			return inst.Sequence{lhs, rhs, inst.GreaterFloat{}}
		case mdl.Int8:
			return inst.Sequence{lhs, rhs, inst.GreaterInt8{}}
		case mdl.Int16:
			return inst.Sequence{lhs, rhs, inst.GreaterInt16{}}
		case mdl.Int32:
			return inst.Sequence{lhs, rhs, inst.GreaterInt32{}}
		case mdl.Int64:
			return inst.Sequence{lhs, rhs, inst.GreaterInt64{}}
		case mdl.Uint8:
			return inst.Sequence{lhs, rhs, inst.GreaterUint8{}}
		case mdl.Uint16:
			return inst.Sequence{lhs, rhs, inst.GreaterUint16{}}
		case mdl.Uint32:
			return inst.Sequence{lhs, rhs, inst.GreaterUint32{}}
		case mdl.Uint64:
			return inst.Sequence{lhs, rhs, inst.GreaterUint64{}}
		}
	case xpr.Less:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		switch node[0].(xpr.TypedExpression).Actual.Concrete().(type) {
		case mdl.Float:
			return inst.Sequence{lhs, rhs, inst.LessFloat{}}
		case mdl.Int8:
			return inst.Sequence{lhs, rhs, inst.LessInt8{}}
		case mdl.Int16:
			return inst.Sequence{lhs, rhs, inst.LessInt16{}}
		case mdl.Int32:
			return inst.Sequence{lhs, rhs, inst.LessInt32{}}
		case mdl.Int64:
			return inst.Sequence{lhs, rhs, inst.LessInt64{}}
		case mdl.Uint8:
			return inst.Sequence{lhs, rhs, inst.LessUint8{}}
		case mdl.Uint16:
			return inst.Sequence{lhs, rhs, inst.LessUint16{}}
		case mdl.Uint32:
			return inst.Sequence{lhs, rhs, inst.LessUint32{}}
		case mdl.Uint64:
			return inst.Sequence{lhs, rhs, inst.LessUint64{}}
		}
	case xpr.Add:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		switch node[0].(xpr.TypedExpression).Actual.Concrete().(type) {
		case mdl.Float:
			return inst.Sequence{lhs, rhs, inst.AddFloat{}}
		case mdl.Int8:
			return inst.Sequence{lhs, rhs, inst.AddInt8{}}
		case mdl.Int16:
			return inst.Sequence{lhs, rhs, inst.AddInt16{}}
		case mdl.Int32:
			return inst.Sequence{lhs, rhs, inst.AddInt32{}}
		case mdl.Int64:
			return inst.Sequence{lhs, rhs, inst.AddInt64{}}
		case mdl.Uint8:
			return inst.Sequence{lhs, rhs, inst.AddUint8{}}
		case mdl.Uint16:
			return inst.Sequence{lhs, rhs, inst.AddUint16{}}
		case mdl.Uint32:
			return inst.Sequence{lhs, rhs, inst.AddUint32{}}
		case mdl.Uint64:
			return inst.Sequence{lhs, rhs, inst.AddUint64{}}
		}
	case xpr.Subtract:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		switch node[0].(xpr.TypedExpression).Actual.Concrete().(type) {
		case mdl.Float:
			return inst.Sequence{lhs, rhs, inst.SubtractFloat{}}
		case mdl.Int8:
			return inst.Sequence{lhs, rhs, inst.SubtractInt8{}}
		case mdl.Int16:
			return inst.Sequence{lhs, rhs, inst.SubtractInt16{}}
		case mdl.Int32:
			return inst.Sequence{lhs, rhs, inst.SubtractInt32{}}
		case mdl.Int64:
			return inst.Sequence{lhs, rhs, inst.SubtractInt64{}}
		case mdl.Uint8:
			return inst.Sequence{lhs, rhs, inst.SubtractUint8{}}
		case mdl.Uint16:
			return inst.Sequence{lhs, rhs, inst.SubtractUint16{}}
		case mdl.Uint32:
			return inst.Sequence{lhs, rhs, inst.SubtractUint32{}}
		case mdl.Uint64:
			return inst.Sequence{lhs, rhs, inst.SubtractUint64{}}
		}
	case xpr.Multiply:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		switch node[0].(xpr.TypedExpression).Actual.Concrete().(type) {
		case mdl.Float:
			return inst.Sequence{lhs, rhs, inst.MultiplyFloat{}}
		case mdl.Int8:
			return inst.Sequence{lhs, rhs, inst.MultiplyInt8{}}
		case mdl.Int16:
			return inst.Sequence{lhs, rhs, inst.MultiplyInt16{}}
		case mdl.Int32:
			return inst.Sequence{lhs, rhs, inst.MultiplyInt32{}}
		case mdl.Int64:
			return inst.Sequence{lhs, rhs, inst.MultiplyInt64{}}
		case mdl.Uint8:
			return inst.Sequence{lhs, rhs, inst.MultiplyUint8{}}
		case mdl.Uint16:
			return inst.Sequence{lhs, rhs, inst.MultiplyUint16{}}
		case mdl.Uint32:
			return inst.Sequence{lhs, rhs, inst.MultiplyUint32{}}
		case mdl.Uint64:
			return inst.Sequence{lhs, rhs, inst.MultiplyUint64{}}
		}
	case xpr.Divide:
		lhs := vm.Compile(node[0].(xpr.TypedExpression))
		rhs := vm.Compile(node[1].(xpr.TypedExpression))
		switch node[0].(xpr.TypedExpression).Actual.Concrete().(type) {
		case mdl.Float:
			return inst.Sequence{lhs, rhs, inst.DivideFloat{}}
		case mdl.Int8:
			return inst.Sequence{lhs, rhs, inst.DivideInt8{}}
		case mdl.Int16:
			return inst.Sequence{lhs, rhs, inst.DivideInt16{}}
		case mdl.Int32:
			return inst.Sequence{lhs, rhs, inst.DivideInt32{}}
		case mdl.Int64:
			return inst.Sequence{lhs, rhs, inst.DivideInt64{}}
		case mdl.Uint8:
			return inst.Sequence{lhs, rhs, inst.DivideUint8{}}
		case mdl.Uint16:
			return inst.Sequence{lhs, rhs, inst.DivideUint16{}}
		case mdl.Uint32:
			return inst.Sequence{lhs, rhs, inst.DivideUint32{}}
		case mdl.Uint64:
			return inst.Sequence{lhs, rhs, inst.DivideUint64{}}
		}

	case xpr.And:
		is := make(inst.Sequence, 0, len(node)*2)
		for i, sub := range node {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, arg, inst.ShortCircuitAnd{Arity: len(node), Step: (i + 1)})
		}
		return is

	case xpr.Or:
		is := make(inst.Sequence, 0, len(node)*2)
		for i, sub := range node {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, arg, inst.ShortCircuitOr{Arity: len(node), Step: (i + 1)})
		}
		return is

	case xpr.CreateMultiple:
		is := make(inst.Sequence, 0, len(node.Values)*2+2)
		for k, sub := range node.Values {
			arg := vm.Compile(sub.(xpr.TypedExpression))
			is = append(is, inst.Constant{val.String(k)}, arg)
		}
		mref := node.In.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		return append(is, inst.BuildMap{len(node.Values)}, inst.CreateMultiple{mref[1]})

	case xpr.Slice:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		offset := vm.Compile(node.Offset.(xpr.TypedExpression))
		length := vm.Compile(node.Length.(xpr.TypedExpression))
		return inst.Sequence{value, offset, length, inst.Slice{}}

	case xpr.SearchAllRegex:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		regex := string(node.Regex.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String))
		if node.MultiLine.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?m)` + regex
		}
		if node.CaseInsensitive.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?i)` + regex
		}
		r := regexp.MustCompile(regex) // compilation previously checked
		return inst.Sequence{value, inst.SearchAllRegex{r}}

	case xpr.SearchRegex:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		regex := string(node.Regex.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String))
		if node.MultiLine.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?m)` + regex
		}
		if node.CaseInsensitive.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?i)` + regex
		}
		r := regexp.MustCompile(regex) // compilation previously checked
		return inst.Sequence{value, inst.SearchRegex{r}}

	case xpr.MatchRegex:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		regex := string(node.Regex.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.String))
		if node.MultiLine.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?m)` + regex
		}
		if node.CaseInsensitive.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Bool) {
			regex = `(?i)` + regex
		}
		r := regexp.MustCompile(regex) // compilation previously checked
		return inst.Sequence{value, inst.MatchRegex{r}}

	case xpr.AssertModelRef:
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		mref := node.Ref.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
		return inst.Sequence{value, inst.AssertModelRef{mref[1]}}

	case xpr.SwitchModelRef:
		cases := make(map[string]inst.Instruction, len(node.Cases))
		for _, sub := range node.Cases {
			mref := sub.Match.(xpr.TypedExpression).Actual.(ConstantModel).Value.(val.Ref)
			cases[mref[1]] = vm.Compile(sub.Return.(xpr.TypedExpression))
		}
		value := vm.Compile(node.Value.(xpr.TypedExpression))
		deflt := vm.Compile(node.Default.(xpr.TypedExpression))
		return inst.Sequence{value, inst.SwitchModelRef{Cases: cases, Default: deflt}}

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
		start := vm.Compile(node.Start.(xpr.TypedExpression))
		return inst.Sequence{start, inst.GraphFlow{flow}}

	case xpr.Do:
		panic("vm.Compile: uneliminiated xpr.Do")

	case xpr.Bind:
		panic("vm.Compile: uneliminiated xpr.Bind")

	case xpr.SwitchType:
		instruction := make(inst.SwitchType, len(node.Cases))
		for k, v := range node.Cases {
			instruction[k] = vm.Compile(v.(xpr.TypedExpression))
		}
		return inst.Sequence{vm.Compile(node.Value.(xpr.TypedExpression)), instruction}

	case xpr.SwitchCase:
		instruction := make(inst.SwitchCase, len(node.Cases))
		for k, v := range node.Cases {
			instruction[k] = vm.Compile(v.(xpr.TypedExpression))
		}
		return inst.Sequence{vm.Compile(node.Value.(xpr.TypedExpression)), instruction}

	case xpr.MemSort:
		return inst.Sequence{
			vm.Compile(node.Value.(xpr.TypedExpression)),
			inst.MemSort{vm.Compile(node.Expression.(xpr.TypedExpression))},
		}

	case xpr.MapSet:
		return inst.Sequence{
			vm.Compile(node.Value.(xpr.TypedExpression)),
			inst.MapSet{vm.Compile(node.Expression.(xpr.TypedExpression))},
		}

	}
	panic(fmt.Sprintf("unhandled case: %T", typed))

}
