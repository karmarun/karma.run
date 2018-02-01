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

// TransformIdentity is the identity function for Expressions
func TransformIdentity(m Expression) Expression {
	return m
}

var argValue = val.Union{"arg", val.Struct{}}

func ExpressionFromValue(v val.Value) Expression {

	if _, ok := v.(val.Union); !ok {
		return Literal{v}
	}

	switch u := v.(val.Union); u.Case {

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

	case "ref":
		// case: (string, string)
		// case: (ref<metaModel>, string)   // <- needs validation
		// case: (ref<modelA>, ref<modelB>) // <- equivalent to relocateRef
		arg := u.Value.(val.Tuple)
		return RefConstructor{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "id", "arg":
		return Argument{}

	case "currentUser":
		return CurrentUser{}

	case "zero":
		return Zero{}

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

	case "tag":
		return Tag{ExpressionFromValue(u.Value)}

	case "all":
		return All{ExpressionFromValue(u.Value)}

	case "setField":
		arg := u.Value.(val.Struct)
		if arg.Field("in") == val.Null {
			arg.Set("in", argValue)
		}
		return SetField{ExpressionFromValue(arg.Field("name")), ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("in"))}

	case "setKey":
		arg := u.Value.(val.Struct)
		if arg.Field("in") == val.Null {
			arg.Set("in", argValue)
		}
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
		arg := u.Value.(val.Struct)
		vls := arg.Field("values").(val.Map)
		mvs := make(map[string]Expression, vls.Len())
		vls.ForEach(func(k string, v val.Value) bool {
			mvs[k] = ExpressionFromValue(v)
			return true
		})
		return CreateMultiple{ExpressionFromValue(arg.Field("in")), mvs}

	case "if":
		arg := u.Value.(val.Struct)
		return If{ExpressionFromValue(arg.Field("condition")), ExpressionFromValue(arg.Field("then")), ExpressionFromValue(arg.Field("else"))}

	case "with":
		arg := u.Value.(val.Struct)
		return With{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("return"))}

	case "update":
		arg := u.Value.(val.Struct)
		return Update{ExpressionFromValue(arg.Field("ref")), ExpressionFromValue(arg.Field("value"))}

	case "create":
		arg := u.Value.(val.Struct)
		return Create{ExpressionFromValue(arg.Field("in")), ExpressionFromValue(arg.Field("value"))}

	case "mapMap":
		arg := u.Value.(val.Struct)
		return MapMap{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("expression"))}

	case "mapList":
		arg := u.Value.(val.Struct)
		return MapList{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("expression"))}

	case "reduceList":
		arg := u.Value.(val.Struct)
		return ReduceList{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("expression"))}

	case "graphFlow":
		arg := u.Value.(val.Struct)
		flw := arg.Field("flow").(val.List)
		nod := GraphFlow{ExpressionFromValue(arg.Field("start")), make([]GraphFlowParam, len(flw), len(flw))}
		for i, sub := range flw {
			subArg := sub.(val.Struct)
			fwd, bwd := []Expression(nil), []Expression(nil)
			if fl, ok := subArg.Field("forward").(val.List); ok {
				fwd = make([]Expression, len(fl), len(fl))
				for i, sub := range fl {
					fwd[i] = ExpressionFromValue(sub)
				}
			}
			if bl, ok := subArg.Field("backward").(val.List); ok {
				bwd = make([]Expression, len(bl), len(bl))
				for i, sub := range bl {
					bwd[i] = ExpressionFromValue(sub)
				}
			}
			nod.Flow[i] = GraphFlowParam{ExpressionFromValue(subArg.Field("from")), fwd, bwd}
		}
		return nod

	case "resolveRefs":
		arg := u.Value.(val.Struct)
		mla := arg.Field("models").(val.List)
		mns := make([]Expression, len(mla), len(mla))
		for i, sub := range mla {
			mns[i] = ExpressionFromValue(sub)
		}
		return ResolveRefs{ExpressionFromValue(arg.Field("value")), mns}

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

	case "filter":
		arg := u.Value.(val.Struct)
		return Filter{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("expression"))}

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

	case "greater":
		arg := u.Value.(val.Tuple)
		return Greater{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "less":
		arg := u.Value.(val.Tuple)
		return Less{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "add":
		arg := u.Value.(val.Tuple)
		return Add{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "subtract":
		arg := u.Value.(val.Tuple)
		return Subtract{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "multiply":
		arg := u.Value.(val.Tuple)
		return Multiply{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "divide":
		arg := u.Value.(val.Tuple)
		return Divide{ExpressionFromValue(arg[0]), ExpressionFromValue(arg[1])}

	case "or":
		arg := u.Value.(val.List)
		if len(arg) == 1 {
			return ExpressionFromValue(arg[0])
		}
		nod := make(Or, len(arg), len(arg))
		for i, sub := range arg {
			nod[i] = ExpressionFromValue(sub)
		}
		return nod

	case "and":
		arg := u.Value.(val.List)
		nod := make(And, len(arg), len(arg))
		for i, sub := range arg {
			nod[i] = ExpressionFromValue(sub)
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
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		return AssertModelRef{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("ref"))}

	case "slice":
		arg := u.Value.(val.Struct)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		return Slice{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("offset")), ExpressionFromValue(arg.Field("length"))}

	case "searchAllRegex":
		if sarg, ok := u.Value.(val.String); ok {
			u.Value = val.StructFromMap(map[string]val.Value{
				"value":           argValue,
				"regex":           sarg,
				"multiLine":       val.Bool(false),
				"caseInsensitive": val.Bool(false),
			})
		}
		arg := u.Value.(val.Struct)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		if arg.Field("multiLine") == val.Null {
			arg.Set("multiLine", val.Bool(false))
		}
		if arg.Field("caseInsensitive") == val.Null {
			arg.Set("caseInsensitive", val.Bool(false))
		}
		return SearchAllRegex{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("regex")),
			ExpressionFromValue(arg.Field("multiLine")),
			ExpressionFromValue(arg.Field("caseInsensitive")),
		}

	case "searchRegex":
		if sarg, ok := u.Value.(val.String); ok {
			u.Value = val.StructFromMap(map[string]val.Value{
				"value":           argValue,
				"regex":           sarg,
				"multiLine":       val.Bool(false),
				"caseInsensitive": val.Bool(false),
			})
		}
		arg := u.Value.(val.Struct)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		if arg.Field("multiLine") == val.Null {
			arg.Set("multiLine", val.Bool(false))
		}
		if arg.Field("caseInsensitive") == val.Null {
			arg.Set("caseInsensitive", val.Bool(false))
		}
		return SearchRegex{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("regex")),
			ExpressionFromValue(arg.Field("multiLine")),
			ExpressionFromValue(arg.Field("caseInsensitive")),
		}

	case "matchRegex":
		if sarg, ok := u.Value.(val.String); ok {
			u.Value = val.StructFromMap(map[string]val.Value{
				"value":           argValue,
				"regex":           sarg,
				"multiLine":       val.Bool(false),
				"caseInsensitive": val.Bool(false),
			})
		}
		arg := u.Value.(val.Struct)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		if arg.Field("multiLine") == val.Null {
			arg.Set("multiLine", val.Bool(false))
		}
		if arg.Field("caseInsensitive") == val.Null {
			arg.Set("caseInsensitive", val.Bool(false))
		}
		return MatchRegex{
			ExpressionFromValue(arg.Field("value")),
			ExpressionFromValue(arg.Field("regex")),
			ExpressionFromValue(arg.Field("multiLine")),
			ExpressionFromValue(arg.Field("caseInsensitive")),
		}

	case "switchModelRef":
		arg := u.Value.(val.Struct)
		css := arg.Field("cases").(val.List)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		nod := SwitchModelRef{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("default")), make([]SwitchModelRefCase, len(css), len(css))}
		for i, sub := range css {
			subArg := sub.(val.Struct)
			nod.Cases[i] = SwitchModelRefCase{ExpressionFromValue(subArg.Field("match")), ExpressionFromValue(subArg.Field("return"))}
		}
		return nod

	case "key":
		if arg, ok := u.Value.(val.String); ok {
			return Key{Argument{}, ExpressionFromValue(arg)}
		}
		arg := u.Value.(val.Struct)
		return Key{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("name"))}

	case "field":
		if arg, ok := u.Value.(val.String); ok {
			return Field{Argument{}, ExpressionFromValue(arg)}
		}
		arg := u.Value.(val.Struct)
		return Field{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("name"))}

	case "index":
		arg := u.Value.(val.Struct)
		return Index{ExpressionFromValue(arg.Field("value")), ExpressionFromValue(arg.Field("number"))}

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
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		cases := make(map[string]Expression)
		arg.Field("cases").(val.Map).ForEach(func(k string, v val.Value) bool {
			cases[k] = ExpressionFromValue(v)
			return true
		})
		return SwitchCase{
			Value: ExpressionFromValue(arg.Field("value")),
			Cases: cases,
		}

	case "memSort":
		arg := u.Value.(val.Struct)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		return MemSort{
			Value:      ExpressionFromValue(arg.Field("value")),
			Expression: ExpressionFromValue(arg.Field("expression")),
		}

	case "mapSet":
		arg := u.Value.(val.Struct)
		if arg.Field("value") == val.Null {
			arg.Set("value", argValue)
		}
		return MapSet{
			Value:      ExpressionFromValue(arg.Field("value")),
			Expression: ExpressionFromValue(arg.Field("expression")),
		}

	default:
		panic(fmt.Sprintf("unhandled function name: %s", u.Case))

	}

}

func ValueFromExpression(x Expression) val.Value {

	switch node := x.(type) {

	case TypedExpression:
		return ValueFromExpression(node.Expression)

	case Argument:
		return val.Union{"arg", val.Struct{}}

	case CurrentUser:
		return val.Union{"currentUser", val.Struct{}}

	case Zero:
		return val.Union{"zero", val.Struct{}}

	case Literal:
		// case NewList:
		// 	return val.Union{"newList", make(val.List, len(node), len(node)).OverMap(func(i int, _ val.Value) val.Value {
		// 		return ValueFromExpression(node[i])
		// 	})}

		// case NewTuple:
		// 	return val.Union{"newTuple", make(val.Tuple, len(node), len(node)).OverMap(func(i int, _ val.Value) val.Value {
		// 		return ValueFromExpression(node[i])
		// 	})}

		// case NewMap:
		// 	values := val.NewMap(len(node))
		// 	for k, sub := range node {
		// 		values.Set(k, ValueFromExpression(sub))
		// 	}
		// 	return val.Union{"newMap", values}

		// case NewStruct:
		// 	values := val.NewMap(len(node))
		// 	for k, sub := range node {
		// 		values.Set(k, ValueFromExpression(sub))
		// 	}
		// 	return val.Union{"newStruct", values}

		// case NewUnion:
		// 	return val.Union{"newUnion", val.StructFromMap(map[string]val.Value{
		// 		"value": ValueFromExpression(node.Value),
		// 		"case":  ValueFromExpression(node.Case),
		// 	})}

		switch v := node.Value.(type) {
		case val.Meta:
			return ValueFromExpression(Literal{v.Value})

		case val.Tuple:
			return val.Union{"tuple", make(val.List, len(v), len(v)).OverMap(func(i int, v val.Value) val.Value {
				return ValueFromExpression(Literal{v})
			})}

		case val.List:
			return val.Union{"list", make(val.List, len(v), len(v)).OverMap(func(i int, v val.Value) val.Value {
				return ValueFromExpression(Literal{v})
			})}

		case val.Union:
			arg := make(val.Tuple, 2, 2)
			arg[0] = val.String(v.Case)
			arg[1] = ValueFromExpression(Literal{v.Value})
			return val.Union{"union", arg}

		case val.Struct:
			arg := val.NewMap(v.Len())
			v.ForEach(func(k string, v val.Value) bool {
				arg.Set(k, ValueFromExpression(Literal{v}))
				return true
			})
			return val.Union{"struct", arg}

		case val.Map:
			arg := val.NewMap(v.Len())
			v.ForEach(func(k string, v val.Value) bool {
				arg.Set(k, ValueFromExpression(Literal{v}))
				return true
			})
			return val.Union{"map", arg}

		case val.Set:
			arg := make(val.Set, len(v))
			for _, v := range v {
				w := ValueFromExpression(Literal{v})
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

		}

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
			"return": ValueFromExpression(node.Return),
		})}

	case Update:
		return val.Union{"update", val.StructFromMap(map[string]val.Value{
			"ref":   ValueFromExpression(node.Ref),
			"value": ValueFromExpression(node.Value),
		})}

	case Create:
		return val.Union{"create", val.StructFromMap(map[string]val.Value{
			"in":    ValueFromExpression(node.In),
			"value": ValueFromExpression(node.Value),
		})}

	case InList:
		return val.Union{"inList", val.StructFromMap(map[string]val.Value{
			"in":    ValueFromExpression(node.In),
			"value": ValueFromExpression(node.Value),
		})}

	case Filter:
		return val.Union{"filter", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		})}

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
		return val.Union{"mapMap", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		})}

	case MapList:
		return val.Union{"mapList", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		})}

	case ReduceList:
		return val.Union{"reduceList", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		})}

	case ResolveRefs:
		mls := make(val.List, len(node.Models), len(node.Models))
		for i, _ := range mls {
			mls[i] = ValueFromExpression(node.Models[i])
		}
		return val.Union{"resolveRefs", val.StructFromMap(map[string]val.Value{
			"value":  ValueFromExpression(node.Value),
			"models": mls,
		})}

	case Field:
		if node.Value == (Argument{}) {
			return val.Union{"field", ValueFromExpression(node.Name)}
		}
		return val.Union{"field", val.StructFromMap(map[string]val.Value{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
		})}

	case Key:
		if node.Value == (Argument{}) {
			return val.Union{"key", ValueFromExpression(node.Name)}
		}
		return val.Union{"key", val.StructFromMap(map[string]val.Value{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
		})}

	case Index:
		return val.Union{"index", val.StructFromMap(map[string]val.Value{
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

	case Greater:
		return val.Union{"greater", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Less:
		return val.Union{"less", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Add:
		return val.Union{"add", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Subtract:
		return val.Union{"subtract", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Multiply:
		return val.Union{"multiply", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case Divide:
		return val.Union{"divide", val.Tuple{
			ValueFromExpression(node[0]),
			ValueFromExpression(node[1]),
		}}

	case And:
		return val.Union{"and", make(val.List, len(node), len(node)).OverMap(func(i int, _ val.Value) val.Value {
			return ValueFromExpression(node[i])
		})}

	case Or:
		return val.Union{"or", make(val.List, len(node), len(node)).OverMap(func(i int, _ val.Value) val.Value {
			return ValueFromExpression(node[i])
		})}

	case CreateMultiple:
		values := val.NewMap(len(node.Values))
		for k, sub := range node.Values {
			values.Set(k, ValueFromExpression(sub))
		}
		return val.Union{"createMultiple", val.StructFromMap(map[string]val.Value{
			"in":     ValueFromExpression(node.In),
			"values": values,
		})}

	case Slice:
		arg := val.StructFromMap(map[string]val.Value{
			"offset": ValueFromExpression(node.Offset),
			"length": ValueFromExpression(node.Length),
		})
		if node.Value != (Argument{}) {
			arg.Set("value", ValueFromExpression(node.Value))
		}
		return val.Union{"slice", arg}

	case SearchRegex:
		arg := val.NewStruct(4)
		arg.Set("regex", ValueFromExpression(node.Regex))
		if node.Value != (Argument{}) {
			arg.Set("value", ValueFromExpression(node.Value))
		}
		if node.MultiLine != (Literal{val.Bool(false)}) {
			arg.Set("multiLine", ValueFromExpression(node.MultiLine))
		}
		if node.CaseInsensitive != (Literal{val.Bool(false)}) {
			arg.Set("caseInsensitive", ValueFromExpression(node.CaseInsensitive))
		}
		return val.Union{"searchRegex", arg}

	case SearchAllRegex:
		arg := val.NewStruct(4)
		arg.Set("regex", ValueFromExpression(node.Regex))
		if node.Value != (Argument{}) {
			arg.Set("value", ValueFromExpression(node.Value))
		}
		if node.MultiLine != (Literal{val.Bool(false)}) {
			arg.Set("multiLine", ValueFromExpression(node.MultiLine))
		}
		if node.CaseInsensitive != (Literal{val.Bool(false)}) {
			arg.Set("caseInsensitive", ValueFromExpression(node.CaseInsensitive))
		}
		return val.Union{"searchAllRegex", arg}

	case MatchRegex:
		arg := val.NewStruct(4)
		arg.Set("regex", ValueFromExpression(node.Regex))
		if node.Value != (Argument{}) {
			arg.Set("value", ValueFromExpression(node.Value))
		}
		if node.MultiLine != (Literal{val.Bool(false)}) {
			arg.Set("multiLine", ValueFromExpression(node.MultiLine))
		}
		if node.CaseInsensitive != (Literal{val.Bool(false)}) {
			arg.Set("caseInsensitive", ValueFromExpression(node.CaseInsensitive))
		}
		return val.Union{"matchRegex", arg}

	case AssertModelRef:
		return val.Union{"assertModelRef", val.StructFromMap(map[string]val.Value{
			"value": ValueFromExpression(node.Value),
			"ref":   ValueFromExpression(node.Ref),
		})}

	case SwitchModelRef:
		cases := make(val.List, len(node.Cases), len(node.Cases))
		for i, caze := range node.Cases {
			cases[i] = val.StructFromMap(map[string]val.Value{
				"match":  ValueFromExpression(caze.Match),
				"return": ValueFromExpression(caze.Return),
			})
		}
		return val.Union{"switchModelRef", val.StructFromMap(map[string]val.Value{
			"value":   ValueFromExpression(node.Value),
			"default": ValueFromExpression(node.Default),
			"cases":   cases,
		})}
	case GraphFlow:
		flow := make(val.List, len(node.Flow), len(node.Flow))
		for i, sub := range node.Flow {
			flow[i] = val.StructFromMap(map[string]val.Value{
				"from": ValueFromExpression(sub.From),
				"forward": make(val.List, len(sub.Forward), len(sub.Forward)).OverMap(func(i int, _ val.Value) val.Value {
					return ValueFromExpression(sub.Forward[i])
				}),
				"backward": make(val.List, len(sub.Backward), len(sub.Backward)).OverMap(func(i int, _ val.Value) val.Value {
					return ValueFromExpression(sub.Backward[i])
				}),
			})
		}
		return val.Union{"graphFlow", val.StructFromMap(map[string]val.Value{
			"start": ValueFromExpression(node.Start),
			"flow":  flow,
		})}

	case SwitchCase:
		cases := val.NewMap(len(node.Cases))
		for k, v := range node.Cases {
			cases.Set(k, ValueFromExpression(v))
		}
		return val.Union{"switchCase", val.StructFromMap(map[string]val.Value{
			"value": ValueFromExpression(node.Value),
			"cases": cases,
		})}

	case MemSort:
		return val.Union{"memSort", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		})}

	case MapSet:
		return val.Union{"mapSet", val.StructFromMap(map[string]val.Value{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		})}

	}
	panic(fmt.Sprintf("unhandled case: %T", x))
}
