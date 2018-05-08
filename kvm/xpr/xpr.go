// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"fmt"
	"karma.run/kvm/val"
)

type Expression interface {
	Transform(f func(Expression) Expression) Expression
}

type Function interface {
	Parameters() []string
	Expressions() []Expression
}

// TransformIdentity is the identity function for Expressions
func TransformIdentity(m Expression) Expression {
	return m
}

func FunctionFromValue(v val.Value) Function {

	switch u := v.(val.Union); u.Case {

	case "function":
		def := u.Value.(val.Tuple)
		args := def[0].(val.List)
		exprs := def[1].(val.List)
		node := function{
			args:  make([]string, len(args), len(args)),
			exprs: make([]Expression, len(exprs), len(exprs)),
		}
		for i, v := range args {
			node.args[i] = string(v.(val.String))
		}
		for i, v := range exprs {
			node.exprs[i] = ExpressionFromValue(v)
		}
		return node

	default:
		panic(fmt.Sprintf("undefined case: %s", u.Case))
	}

}

func ExpressionFromValue(v val.Value) Expression {

	if _, ok := v.(val.Union); !ok {
		return Literal{v}
	}

	switch u := v.(val.Union); u.Case {

	case "null", // convenience primitive constructors
		"bool",
		"dateTime",
		"string",
		"float",
		"int8",
		"int16",
		"int32",
		"int64",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"symbol":
		return DataExpressionFromValue(u)

	case "signature":
		return FunctionSignature{FunctionFromValue(u.Value)}

	case "data":
		return DataExpressionFromValue(u.Value)

	case "define":
		args := u.Value.(val.Tuple)
		name := args[0].(val.String)
		return Define{string(name), ExpressionFromValue(args[1])}

	case "scope":
		return Scope(u.Value.(val.String))

	case "gtFloat":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtFloat{lhs, rhs}

	case "gtInt64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtInt64{lhs, rhs}

	case "gtInt32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtInt32{lhs, rhs}

	case "gtInt16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtInt16{lhs, rhs}

	case "gtInt8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtInt8{lhs, rhs}

	case "gtUint64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtUint64{lhs, rhs}

	case "gtUint32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtUint32{lhs, rhs}

	case "gtUint16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtUint16{lhs, rhs}

	case "gtUint8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return GtUint8{lhs, rhs}

	case "ltFloat":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtFloat{lhs, rhs}

	case "ltInt64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtInt64{lhs, rhs}

	case "ltInt32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtInt32{lhs, rhs}

	case "ltInt16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtInt16{lhs, rhs}

	case "ltInt8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtInt8{lhs, rhs}

	case "ltUint64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtUint64{lhs, rhs}

	case "ltUint32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtUint32{lhs, rhs}

	case "ltUint16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtUint16{lhs, rhs}

	case "ltUint8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return LtUint8{lhs, rhs}

	case "mulFloat":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulFloat{lhs, rhs}

	case "mulInt64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulInt64{lhs, rhs}

	case "mulInt32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulInt32{lhs, rhs}

	case "mulInt16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulInt16{lhs, rhs}

	case "mulInt8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulInt8{lhs, rhs}

	case "mulUint64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulUint64{lhs, rhs}

	case "mulUint32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulUint32{lhs, rhs}

	case "mulUint16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulUint16{lhs, rhs}

	case "mulUint8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return MulUint8{lhs, rhs}

	case "divFloat":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivFloat{lhs, rhs}

	case "divInt64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivInt64{lhs, rhs}

	case "divInt32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivInt32{lhs, rhs}

	case "divInt16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivInt16{lhs, rhs}

	case "divInt8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivInt8{lhs, rhs}

	case "divUint64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivUint64{lhs, rhs}

	case "divUint32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivUint32{lhs, rhs}

	case "divUint16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivUint16{lhs, rhs}

	case "divUint8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return DivUint8{lhs, rhs}

	case "subInt64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubInt64{lhs, rhs}

	case "subInt32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubInt32{lhs, rhs}

	case "subInt16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubInt16{lhs, rhs}

	case "subInt8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubInt8{lhs, rhs}

	case "subUint64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubUint64{lhs, rhs}

	case "subUint32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubUint32{lhs, rhs}

	case "subUint16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubUint16{lhs, rhs}

	case "subUint8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubUint8{lhs, rhs}

	case "subFloat":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return SubFloat{lhs, rhs}

	case "addFloat":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddFloat{lhs, rhs}

	case "addInt64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddInt64{lhs, rhs}

	case "addInt32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddInt32{lhs, rhs}

	case "addInt16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddInt16{lhs, rhs}

	case "addInt8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddInt8{lhs, rhs}

	case "addUint64":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddUint64{lhs, rhs}

	case "addUint32":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddUint32{lhs, rhs}

	case "addUint16":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddUint16{lhs, rhs}

	case "addUint8":
		args := u.Value.(val.Tuple)
		lhs := ExpressionFromValue(args[0])
		rhs := ExpressionFromValue(args[1])
		return AddUint8{lhs, rhs}

	case "currentUser":
		return CurrentUser{}

	case "zero":
		return Zero{}

	case "allReferrers":
		return AllReferrers{ExpressionFromValue(u.Value)}

	case "isPresent":
		return IsPresent{ExpressionFromValue(u.Value)}

	case "presentOrZero":
		return PresentOrZero{ExpressionFromValue(u.Value)}

	case "assertPresent":
		return AssertPresent{ExpressionFromValue(u.Value)}

	case "model":
		return Model{ExpressionFromValue(u.Value)}

	case "stringToLower":
		return StringToLower{ExpressionFromValue(u.Value)}

	case "reverseList":
		return ReverseList{ExpressionFromValue(u.Value)}

	case "tagExists":
		return TagExists{ExpressionFromValue(u.Value)}

	case "tag":
		return Tag{ExpressionFromValue(u.Value)}

	case "all":
		return All{ExpressionFromValue(u.Value)}

	case "setField":
		arg := u.Value.(val.Struct)
		return SetField{ExpressionFromValue(arg.Field("name")), ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("in"))}

	case "setKey":
		arg := u.Value.(val.Struct)
		return SetKey{ExpressionFromValue(arg.Field("name")), ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("in"))}

	case "joinStrings":
		arg := u.Value.(val.Struct)
		if arg.Field("separator") == val.Null {
			arg.Set("separator", val.String(""))
		}
		return JoinStrings{ExpressionFromValue(arg.Field("strings")), ExpressionFromValue(arg.Field("separator"))}

	case "relocateRef":
		arg := u.Value.(val.Struct)
		return RelocateRef{ExpressionFromValue(arg.Field("ref")), ExpressionFromValue(arg.Field("model"))}

	case "extractStrings":
		return ExtractStrings{ExpressionFromValue(u.Value)}

	case "delete":
		return Delete{ExpressionFromValue(u.Value)}

	case "createMultiple":
		args := u.Value.(val.Tuple)
		vls := args[1].(val.Map)
		mvs := make(map[string]Function, vls.Len())
		vls.ForEach(func(k string, v val.Value) bool {
			mvs[k] = FunctionFromValue(v)
			return true
		})
		return CreateMultiple{ExpressionFromValue(args[0]), mvs}

	case "if":
		arg := u.Value.(val.Struct)
		return If{ExpressionFromValue(arg.Field("condition")), ExpressionFromValue(arg.Field("then")), ExpressionFromValue(arg.Field("else"))}

	case "with":
		arg := u.Value.(val.Struct)
		return With{ExpressionFromValue(arg.Field("value")), FunctionFromValue(arg.Field("return"))}

	case "update":
		arg := u.Value.(val.Struct)
		return Update{ExpressionFromValue(arg.Field("ref")), ExpressionFromValue(arg.Field("value"))}

	case "create":
		arg := u.Value.(val.Tuple)
		return Create{ExpressionFromValue(arg[0]), FunctionFromValue(arg[1])}

	case "mapMap":
		arg := u.Value.(val.Tuple)
		return MapMap{ExpressionFromValue(arg[0]), FunctionFromValue(arg[1])}

	case "mapList":
		arg := u.Value.(val.Tuple)
		return MapList{ExpressionFromValue(arg[0]), FunctionFromValue(arg[1])}

	case "mapSet":
		arg := u.Value.(val.Tuple)
		return MapSet{ExpressionFromValue(arg[0]), FunctionFromValue(arg[1])}

	case "reduceList":
		arg := u.Value.(val.Struct)
		return ReduceList{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("initial")),
			FunctionFromValue(arg.Field("reducer")),
		}

	case "graphFlow":
		arg := u.Value.(val.Struct)
		flw := arg.Field("flow").(val.Set)
		nod := GraphFlow{ExpressionFromValue(arg.Field("start")), make([]GraphFlowParam, 0, len(flw))}
		for _, sub := range flw {
			subArg := sub.(val.Struct)
			fwd, bwd := []Expression(nil), []Expression(nil)
			if fl, ok := subArg.Field("forward").(val.Set); ok {
				fwd = make([]Expression, 0, len(fl))
				for _, sub := range fl {
					fwd = append(fwd, ExpressionFromValue(sub))
				}
			}
			if bl, ok := subArg.Field("backward").(val.Set); ok {
				bwd = make([]Expression, 0, len(bl))
				for _, sub := range bl {
					bwd = append(bwd, ExpressionFromValue(sub))
				}
			}
			nod.Flow = append(nod.Flow, GraphFlowParam{ExpressionFromValue(subArg.Field("from")), fwd, bwd})
		}
		return nod

	case "resolveRefs":
		args := u.Value.(val.Tuple)
		mla := args[1].(val.Set)
		mns := make([]Expression, 0, len(mla))
		for _, sub := range mla {
			mns = append(mns, ExpressionFromValue(sub))
		}
		return ResolveRefs{ExpressionFromValue(args[0]), mns}

	case "resolveAllRefs":
		return ResolveAllRefs{ExpressionFromValue(u.Value)}

	case "referred":
		arg := u.Value.(val.Struct)
		return Referred{ExpressionFromValue(arg.Field("from")), ExpressionFromValue(arg.Field("in"))}

	case "referrers":
		arg := u.Value.(val.Struct)
		return Referrers{ExpressionFromValue(arg.Field("of")), ExpressionFromValue(arg.Field("in"))}

	case "inList":
		arg := u.Value.(val.Struct)
		return InList{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("in"))}

	case "filterList":
		arg := u.Value.(val.Tuple)
		return FilterList{ExpressionFromValue(arg[0]), FunctionFromValue(arg[1])}

	case "first":
		return First{ExpressionFromValue(u.Value)}

	case "get", "deref":
		return Get{ExpressionFromValue(u.Value)}

	case "concatLists":
		arg := u.Value.(val.Tuple)
		return ConcatLists{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "after":
		arg := u.Value.(val.Tuple)
		return After{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "before":
		arg := u.Value.(val.Tuple)
		return Before{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "equal":
		arg := u.Value.(val.Tuple)
		return Equal{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "or":
		arg := u.Value.(val.Set)
		if len(arg) == 1 {
			for _, sub := range arg {
				return ExpressionFromValue(sub)
			}
		}
		nod := make(Or, 0, len(arg))
		for _, sub := range arg {
			nod = append(nod, ExpressionFromValue(sub))
		}
		return nod

	case "and":
		arg := u.Value.(val.Set)
		nod := make(And, 0, len(arg))
		for _, sub := range arg {
			nod = append(nod, ExpressionFromValue(sub))
		}
		return nod

	case "length":
		return Length{ExpressionFromValue(u.Value)}

	case "assertCase":
		arg := u.Value.(val.Struct)
		return AssertCase{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("case"))}

	case "isCase":
		arg := u.Value.(val.Struct)
		return IsCase{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("case"))}

	case "assertModelRef":
		arg, ok := u.Value.(val.Struct)
		if !ok {
			arg = val.StructFromMap(map[string]val.Value{
				"value": val.Null,
				"ref":   u.Value,
			})
		}
		return AssertModelRef{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("ref"))}

	case "slice":
		arg := u.Value.(val.Struct)
		return Slice{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("offset")),
			ExpressionFromValue(arg.Field("length")),
		}

	case "searchAllRegex":
		arg := u.Value.(val.Struct)
		return SearchAllRegex{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("regex")),
			ExpressionFromValue(arg.Field("multiLine")),
			ExpressionFromValue(arg.Field("caseInsensitive")),
		}

	case "searchRegex":
		arg := u.Value.(val.Struct)
		return SearchRegex{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("regex")),
			ExpressionFromValue(arg.Field("multiLine")),
			ExpressionFromValue(arg.Field("caseInsensitive")),
		}

	case "matchRegex":
		arg := u.Value.(val.Struct)
		return MatchRegex{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("regex")),
			ExpressionFromValue(arg.Field("multiLine")),
			ExpressionFromValue(arg.Field("caseInsensitive")),
		}

	case "switchModelRef":
		arg := u.Value.(val.Struct)
		css := arg.Field("cases").(val.Set)
		nod := SwitchModelRef{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("default")), make([]SwitchModelRefCase, 0, len(css))}
		for _, sub := range css {
			subArg := sub.(val.Struct)
			nod.Cases = append(nod.Cases, SwitchModelRefCase{ExpressionFromValue(subArg.Field("match")), ExpressionFromValue(subArg.Field("return"))})
		}
		return nod

	case "key":
		arg := u.Value.(val.Tuple)
		return Key{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "field":
		arg := u.Value.(val.Tuple)
		return Field{string(arg[0].(val.String)), ExpressionFromValue(arg[1])}

	case "indexTuple":
		arg := u.Value.(val.Struct)
		return IndexTuple{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("number"))}

	case "not":
		return Not{ExpressionFromValue(u.Value)}

	case "modelOf":
		return ModelOf{ExpressionFromValue(u.Value)}

	case "metarialize":
		return Metarialize{ExpressionFromValue(u.Value)}

	case "refTo":
		return RefTo{ExpressionFromValue(u.Value)}

	case "switchCase":
		arg := u.Value.(val.Struct)
		cases := make(map[string]Function)
		arg.Field("cases").(val.Map).ForEach(func(k string, v val.Value) bool {
			cases[k] = FunctionFromValue(v)
			return true
		})
		return SwitchCase{
			Value:   ExpressionFromValue(arg.Field("value")),
			Default: ExpressionFromValue(arg.Field("default")),
			Cases:   cases,
		}

	case "memSort":
		arg := u.Value.(val.Struct)
		return MemSort{
			Value: ExpressionFromValue(arg.Field("value")),
			Order: FunctionFromValue(arg.Field("expression")),
		}

	default:
		panic(fmt.Sprintf("unhandled expression: %s", u.Case))

	}

}

func DataExpressionFromValue(v val.Value) Expression {

	switch u := v.(val.Union); u.Case {

	case "expr":
		return ExpressionFromValue(u.Value)

	case "null":
		return Literal{u.Value}

	case "bool":
		return Literal{u.Value}

	case "dateTime":
		return Literal{u.Value}

	case "string":
		return Literal{u.Value}

	case "float":
		return Literal{u.Value}

	case "int8":
		return Literal{u.Value}

	case "int16":
		return Literal{u.Value}

	case "int32":
		return Literal{u.Value}

	case "int64":
		return Literal{u.Value}

	case "uint8":
		return Literal{u.Value}

	case "uint16":
		return Literal{u.Value}

	case "uint32":
		return Literal{u.Value}

	case "uint64":
		return Literal{u.Value}

	case "symbol":
		return Literal{val.Symbol(u.Value.(val.String))}

	case "union":
		arg := u.Value.(val.Tuple)
		return NewUnion{Literal{arg[0].(val.String)}, DataExpressionFromValue(arg[1])}

	case "map":
		arg := u.Value.(val.Map)
		ret := make(NewMap, arg.Len())
		arg.ForEach(func(k string, w val.Value) bool {
			ret[k] = DataExpressionFromValue(w)
			return true
		})
		return ret

	case "set":
		arg := u.Value.(val.Set)
		ret := make(NewSet, 0, len(arg))
		for _, w := range arg {
			ret = append(ret, DataExpressionFromValue(w))
		}
		return ret

	case "tuple":
		arg := u.Value.(val.List)
		ret := make(NewTuple, len(arg))
		for i, w := range arg {
			ret[i] = DataExpressionFromValue(w)
		}
		return ret

	case "list":
		arg := u.Value.(val.List)
		ret := make(NewList, len(arg))
		for i, w := range arg {
			ret[i] = DataExpressionFromValue(w)
		}
		return ret

	case "struct":
		arg := u.Value.(val.Map)
		ret := make(NewStruct, arg.Len())
		arg.ForEach(func(k string, w val.Value) bool {
			ret[k] = DataExpressionFromValue(w)
			return true
		})
		return ret

	case "ref":
		arg := u.Value.(val.Tuple)
		return NewRef{Literal{arg[0]}, Literal{arg[1]}}

	default:
		panic(fmt.Sprintf("unhandled constructor: %s", u.Case))
	}

}

func ValueFromFunction(f Function) val.Value {

	args, exprs := f.Parameters(), f.Expressions()

	argVals := make(val.List, len(args), len(args))
	for i, a := range args {
		argVals[i] = val.String(a)
	}

	exprVals := make(val.List, len(exprs), len(exprs))
	for i, x := range exprs {
		exprVals[i] = ValueFromExpression(x)
	}

	return val.Union{"function", val.Tuple{argVals, exprVals}}

}

func ValueFromExpression(x Expression) val.Value {

	switch node := x.(type) {

	case TypedExpression:
		return ValueFromExpression(node.Expression)

	case FunctionSignature:
		return val.Union{"signature", ValueFromFunction(node.Function)}

	case GtFloat:
		return val.Union{"gtFloat", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtInt64:
		return val.Union{"gtInt64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtInt32:
		return val.Union{"gtInt32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtInt16:
		return val.Union{"gtInt16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtInt8:
		return val.Union{"gtInt8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtUint64:
		return val.Union{"gtUint64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtUint32:
		return val.Union{"gtUint32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtUint16:
		return val.Union{"gtUint16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case GtUint8:
		return val.Union{"gtUint8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtFloat:
		return val.Union{"ltFloat", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtInt64:
		return val.Union{"ltInt64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtInt32:
		return val.Union{"ltInt32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtInt16:
		return val.Union{"ltInt16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtInt8:
		return val.Union{"ltInt8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtUint64:
		return val.Union{"ltUint64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtUint32:
		return val.Union{"ltUint32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtUint16:
		return val.Union{"ltUint16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case LtUint8:
		return val.Union{"ltUint8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case MulFloat:
		return val.Union{"mulFloat", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulInt64:
		return val.Union{"mulInt64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulInt32:
		return val.Union{"mulInt32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulInt16:
		return val.Union{"mulInt16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulInt8:
		return val.Union{"mulInt8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulUint64:
		return val.Union{"mulUint64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulUint32:
		return val.Union{"mulUint32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulUint16:
		return val.Union{"mulUint16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case MulUint8:
		return val.Union{"mulUint8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivFloat:
		return val.Union{"divFloat", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivInt64:
		return val.Union{"divInt64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivInt32:
		return val.Union{"divInt32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivInt16:
		return val.Union{"divInt16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivInt8:
		return val.Union{"divInt8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivUint64:
		return val.Union{"divUint64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivUint32:
		return val.Union{"divUint32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivUint16:
		return val.Union{"divUint16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case DivUint8:
		return val.Union{"divUint8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case SubFloat:
		return val.Union{"subFloat", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case SubInt64:
		return val.Union{"subInt64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubInt32:
		return val.Union{"subInt32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubInt16:
		return val.Union{"subInt16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubInt8:
		return val.Union{"subInt8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubUint64:
		return val.Union{"subUint64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubUint32:
		return val.Union{"subUint32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubUint16:
		return val.Union{"subUint16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case SubUint8:
		return val.Union{"subUint8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddFloat:
		return val.Union{"addFloat", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddInt64:
		return val.Union{"addInt64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddInt32:
		return val.Union{"addInt32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddInt16:
		return val.Union{"addInt16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddInt8:
		return val.Union{"addInt8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddUint64:
		return val.Union{"addUint64", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddUint32:
		return val.Union{"addUint32", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddUint16:
		return val.Union{"addUint16", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}
	case AddUint8:
		return val.Union{"addUint8", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Literal:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewRef:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewStruct:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewList:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewMap:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewUnion:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewSet:
		return val.Union{"data", DataValueFromExpression(node)}

	case NewTuple:
		return val.Union{"data", DataValueFromExpression(node)}

	case CurrentUser:
		return val.Union{"currentUser", val.Struct{}}

	case Zero:
		return val.Union{"zero", val.Struct{}}

	case SetField:
		return val.Union{"setField", val.StructFromMap(map[string]val.Value{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
			"in":    ValueFromExpression(node.In),
		})}
	case SetKey:
		return val.Union{"setKey", val.StructFromMap(map[string]val.Value{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
			"in":    ValueFromExpression(node.In),
		})}

	case AllReferrers:
		return val.Union{"allReferrers", ValueFromExpression(node.Argument)}

	case PresentOrZero:
		return val.Union{"presentOrZero", ValueFromExpression(node.Argument)}

	case IsPresent:
		return val.Union{"isPresent", ValueFromExpression(node.Argument)}

	case AssertPresent:
		return val.Union{"assertPresent", ValueFromExpression(node.Argument)}

	case Model:
		return val.Union{"model", ValueFromExpression(node.Argument)}

	case Tag:
		return val.Union{"tag", ValueFromExpression(node.Argument)}

	case TagExists:
		return val.Union{"tagExists", ValueFromExpression(node.Argument)}

	case All:
		return val.Union{"all", ValueFromExpression(node.Argument)}

	case JoinStrings:
		return val.Union{"joinStrings", val.StructFromMap(map[string]val.Value{
			"strings":   ValueFromExpression(node.Strings),
			"separator": ValueFromExpression(node.Separator), // TODO: separator elision in case it's ""
		})}

	case StringToLower:
		return val.Union{"stringToLower", ValueFromExpression(node.Argument)}

	case ReverseList:
		return val.Union{"reverseList", ValueFromExpression(node.Argument)}

	case ExtractStrings:
		return val.Union{"extractStrings", ValueFromExpression(node.Argument)}

	case Delete:
		return val.Union{"delete", ValueFromExpression(node.Argument)}

	case ResolveAllRefs:
		return val.Union{"resolveAllRefs", ValueFromExpression(node.Argument)}

	case First:
		return val.Union{"first", ValueFromExpression(node.Argument)}

	case Get:
		return val.Union{"get", ValueFromExpression(node.Argument)}

	case Length:
		return val.Union{"length", ValueFromExpression(node.Argument)}

	case Not:
		return val.Union{"not", ValueFromExpression(node.Argument)}

	case ModelOf:
		return val.Union{"modelOf", ValueFromExpression(node.Argument)}

	case Metarialize:
		return val.Union{"metarialize", ValueFromExpression(node.Argument)}

	case RefTo:
		return val.Union{"refTo", ValueFromExpression(node.Argument)}

	case If:
		return val.Union{"if", val.StructFromMap(map[string]val.Value{
			"condition": ValueFromExpression(node.Condition),
			"then":      ValueFromExpression(node.Then),
			"else":      ValueFromExpression(node.Else),
		})}

	case With:
		return val.Union{"with", val.StructFromMap(map[string]val.Value{
			"value":  ValueFromExpression(node.Value),
			"return": ValueFromFunction(node.Return),
		})}

	case Update:
		return val.Union{"update", val.StructFromMap(map[string]val.Value{
			"ref":   ValueFromExpression(node.Ref),
			"value": ValueFromExpression(node.Value),
		})}

	case Create:
		return val.Union{"create", val.Tuple{ValueFromExpression(node.In), ValueFromFunction(node.Value)}}

	case InList:
		return val.Union{"inList", val.StructFromMap(map[string]val.Value{
			"in":    ValueFromExpression(node.In),
			"value": ValueFromExpression(node.Value),
		})}

	case FilterList:
		return val.Union{"filterList", val.Tuple{ValueFromExpression(node.Value), ValueFromFunction(node.Filter)}}

	case AssertCase:
		return val.Union{"assertCase", val.StructFromMap(map[string]val.Value{
			"case":  ValueFromExpression(node.Case),
			"value": ValueFromExpression(node.Value),
		})}

	case IsCase:
		return val.Union{"isCase", val.StructFromMap(map[string]val.Value{
			"case":  ValueFromExpression(node.Case),
			"value": ValueFromExpression(node.Value),
		})}

	case MapMap:
		return val.Union{"mapMap", val.Tuple{ValueFromExpression(node.Value), ValueFromFunction(node.Mapping)}}

	case MapList:
		return val.Union{"mapList", val.Tuple{ValueFromExpression(node.Value), ValueFromFunction(node.Mapping)}}

	case MapSet:
		return val.Union{"mapSet", val.Tuple{ValueFromExpression(node.Value), ValueFromFunction(node.Mapping)}}

	case ReduceList:
		return val.Union{"reduceList", val.StructFromMap(map[string]val.Value{
			"value":   ValueFromExpression(node.Value),
			"initial": ValueFromExpression(node.Initial),
			"reducer": ValueFromFunction(node.Reducer),
		})}

	case ResolveRefs:
		mls := make(val.Set, len(node.Models))
		for _, v := range node.Models {
			w := ValueFromExpression(v)
			mls[val.Hash(w, nil).Sum64()] = w
		}
		return val.Union{"resolveRefs", val.Tuple{ValueFromExpression(node.Value), mls}}

	case Field:
		return val.Union{"field", val.Tuple{val.String(node.Name), ValueFromExpression(node.Value)}}

	case Key:
		return val.Union{"key", val.Tuple{ValueFromExpression(node.Name), ValueFromExpression(node.Value)}}

	case IndexTuple:
		return val.Union{"indexTuple", val.StructFromMap(map[string]val.Value{
			"number": ValueFromExpression(node.Number),
			"value":  ValueFromExpression(node.Value),
		})}

	case Referred:
		return val.Union{"referred", val.StructFromMap(map[string]val.Value{
			"from": ValueFromExpression(node.From),
			"in":   ValueFromExpression(node.In),
		})}

	case RelocateRef:
		return val.Union{"relocateRef", val.StructFromMap(map[string]val.Value{
			"ref":   ValueFromExpression(node.Ref),
			"model": ValueFromExpression(node.Model),
		})}

	case Referrers:
		return val.Union{"referrers", val.StructFromMap(map[string]val.Value{
			"of": ValueFromExpression(node.Of),
			"in": ValueFromExpression(node.In),
		})}

	case After:
		return val.Union{"after", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Before:
		return val.Union{"before", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Equal:
		return val.Union{"equal", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case And:
		arg := make(val.Set, len(node))
		for _, v := range node {
			w := ValueFromExpression(v)
			arg[val.Hash(w, nil).Sum64()] = w
		}
		return val.Union{"and", arg}

	case Or:
		arg := make(val.Set, len(node))
		for _, v := range node {
			w := ValueFromExpression(v)
			arg[val.Hash(w, nil).Sum64()] = w
		}
		return val.Union{"or", arg}

	case CreateMultiple:
		values := val.NewMap(len(node.Values))
		for k, sub := range node.Values {
			values.Set(k, ValueFromFunction(sub))
		}
		return val.Union{"createMultiple", val.Tuple{
			ValueFromExpression(node.In),
			values,
		}}

	case Slice:
		arg := val.StructFromMap(map[string]val.Value{
			"offset": ValueFromExpression(node.Offset),
			"length": ValueFromExpression(node.Length),
			"value":  ValueFromExpression(node.Value),
		})
		return val.Union{"slice", arg}

	case SearchRegex:
		arg := val.NewStruct(4)
		arg.Set("regex", ValueFromExpression(node.Regex))
		arg.Set("value", ValueFromExpression(node.Value))
		arg.Set("multiLine", ValueFromExpression(node.MultiLine))
		arg.Set("caseInsensitive", ValueFromExpression(node.CaseInsensitive))
		return val.Union{"searchRegex", arg}

	case SearchAllRegex:
		arg := val.NewStruct(4)
		arg.Set("regex", ValueFromExpression(node.Regex))
		arg.Set("value", ValueFromExpression(node.Value))
		arg.Set("multiLine", ValueFromExpression(node.MultiLine))
		arg.Set("caseInsensitive", ValueFromExpression(node.CaseInsensitive))
		return val.Union{"searchAllRegex", arg}

	case MatchRegex:
		arg := val.NewStruct(4)
		arg.Set("regex", ValueFromExpression(node.Regex))
		arg.Set("value", ValueFromExpression(node.Value))
		arg.Set("multiLine", ValueFromExpression(node.MultiLine))
		arg.Set("caseInsensitive", ValueFromExpression(node.CaseInsensitive))
		return val.Union{"matchRegex", arg}

	case AssertModelRef:
		return val.Union{"assertModelRef", val.StructFromMap(map[string]val.Value{
			"value": ValueFromExpression(node.Value),
			"ref":   ValueFromExpression(node.Ref),
		})}

	case SwitchModelRef:
		cases := make(val.Set, len(node.Cases))
		for _, caze := range node.Cases {
			w := val.StructFromMap(map[string]val.Value{
				"match":  ValueFromExpression(caze.Match),
				"return": ValueFromExpression(caze.Return),
			})
			cases[val.Hash(w, nil).Sum64()] = w
		}
		return val.Union{"switchModelRef", val.StructFromMap(map[string]val.Value{
			"value":   ValueFromExpression(node.Value),
			"default": ValueFromExpression(node.Default),
			"cases":   cases,
		})}
	case GraphFlow:
		flow := make(val.Set, len(node.Flow))
		for _, sub := range node.Flow {
			fwd := make(val.Set, len(sub.Forward))
			bwd := make(val.Set, len(sub.Backward))
			w := val.StructFromMap(map[string]val.Value{
				"from":     ValueFromExpression(sub.From),
				"forward":  fwd,
				"backward": bwd,
			})
			for _, v := range sub.Forward {
				w := ValueFromExpression(v)
				fwd[val.Hash(w, nil).Sum64()] = w
			}
			for _, v := range sub.Backward {
				w := ValueFromExpression(v)
				bwd[val.Hash(w, nil).Sum64()] = w
			}
			flow[val.Hash(w, nil).Sum64()] = w
		}
		return val.Union{"graphFlow", val.StructFromMap(map[string]val.Value{
			"start": ValueFromExpression(node.Start),
			"flow":  flow,
		})}

	case SwitchCase:
		cases := val.NewMap(len(node.Cases))
		for k, v := range node.Cases {
			cases.Set(k, ValueFromFunction(v))
		}
		return val.Union{"switchCase", val.StructFromMap(map[string]val.Value{
			"value":   ValueFromExpression(node.Value),
			"default": ValueFromExpression(node.Default),
			"cases":   cases,
		})}

	case MemSort:
		return val.Union{"memSort", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromFunction(node.Order),
		})}

	case Scope:
		return val.Union{"scope", val.String(node)}

	case Define:
		return val.Union{"define", val.Tuple{val.String(node.Name), ValueFromExpression(node.Argument)}}

	}
	panic(fmt.Sprintf("unhandled case: %T", x))
}

func DataValueFromExpression(x Expression) val.Value {

	switch node := x.(type) {
	case TypedExpression:
		return DataValueFromExpression(node.Expression)

	case NewRef:
		arg := make(val.Tuple, 2, 2)
		arg[0], arg[1] = DataValueFromExpression(node.Model), DataValueFromExpression(node.Id)
		return val.Union{"ref", arg}

	case NewStruct:
		arg := val.NewMap(len(node))
		for k, w := range node {
			arg.Set(k, DataValueFromExpression(w))
		}
		return val.Union{"struct", arg}

	case NewList:
		arg := make(val.List, len(node))
		for i, w := range node {
			arg[i] = DataValueFromExpression(w)
		}
		return val.Union{"list", arg}

	case NewMap:
		arg := val.NewMap(len(node))
		for k, w := range node {
			arg.Set(k, DataValueFromExpression(w))
		}
		return val.Union{"map", arg}

	case NewUnion:
		arg := make(val.Tuple, 2)
		arg[0] = DataValueFromExpression(node.Case)
		arg[1] = DataValueFromExpression(node.Value)
		return val.Union{"union", arg}

	case NewSet:
		arg := make(val.Set, len(node))
		for _, w := range node {
			v := DataValueFromExpression(w)
			arg[val.Hash(v, nil).Sum64()] = v
		}
		return val.Union{"set", arg}

	case NewTuple:
		arg := make(val.List, len(node))
		for i, w := range node {
			arg[i] = DataValueFromExpression(w)
		}
		return val.Union{"tuple", arg}

	case Literal:
		switch v := node.Value.(type) {

		case val.Meta:
			return DataValueFromExpression(Literal{v.Value})

		case val.Tuple:
			return val.Union{"tuple", make(val.List, len(v), len(v)).OverMap(func(i int, v val.Value) val.Value {
				return DataValueFromExpression(Literal{v})
			})}

		case val.List:
			return val.Union{"list", make(val.List, len(v), len(v)).OverMap(func(i int, v val.Value) val.Value {
				return DataValueFromExpression(Literal{v})
			})}

		case val.Union:
			arg := make(val.Tuple, 2, 2)
			arg[0] = val.String(v.Case)
			arg[1] = DataValueFromExpression(Literal{v.Value})
			return val.Union{"union", arg}

		case val.Struct:
			arg := val.NewMap(v.Len())
			v.ForEach(func(k string, v val.Value) bool {
				arg.Set(k, DataValueFromExpression(Literal{v}))
				return true
			})
			return val.Union{"struct", arg}

		case val.Map:
			arg := val.NewMap(v.Len())
			v.ForEach(func(k string, v val.Value) bool {
				arg.Set(k, DataValueFromExpression(Literal{v}))
				return true
			})
			return val.Union{"map", arg}

		case val.Set:
			arg := make(val.Set, len(v))
			for _, v := range v {
				w := DataValueFromExpression(Literal{v})
				arg[val.Hash(w, nil).Sum64()] = w
			}
			return val.Union{"set", arg}

		case val.Float:
			return val.Union{"float", v}

		case val.Bool:
			return val.Union{"bool", v}

		case val.String:
			return val.Union{"string", v}

		case val.Ref:
			return val.Union{"ref", v}

		case val.DateTime:
			return val.Union{"dateTime", v}

		case val.Symbol:
			return val.Union{"symbol", v}

		case val.Int8:
			return val.Union{"int8", v}

		case val.Int16:
			return val.Union{"int16", v}

		case val.Int32:
			return val.Union{"int32", v}

		case val.Int64:
			return val.Union{"int64", v}

		case val.Uint8:
			return val.Union{"uint8", v}

		case val.Uint16:
			return val.Union{"uint16", v}

		case val.Uint32:
			return val.Union{"uint32", v}

		case val.Uint64:
			return val.Union{"uint64", v}

		default:
			panic(fmt.Sprintf("unhandled literal type %T", node.Value))
		}
	}
	return val.Union{"expr", ValueFromExpression(x)}
}
