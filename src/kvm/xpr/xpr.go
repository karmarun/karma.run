// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"fmt"
	"kvm/val"
)

type Expression interface {
	Transform(f func(Expression) Expression) Expression
}

// TransformIdentity is the identity function for Expressions
func TransformIdentity(m Expression) Expression {
	return m
}

func ExpressionFromValue(v val.Value) Expression {

	if _, ok := v.(val.Union); !ok {
		return Literal{v}
	}

	switch u := v.(val.Union); u.Case {

	case "do":
		arg := u.Value.(val.Map)
		bs := make(map[string]Expression, len(arg))
		for k, v := range arg {
			bs[k] = ExpressionFromValue(v)
		}
		return Do(bs)

	case "bind":
		return Bind(string(u.Value.(val.String)))

	case "id", "arg":
		return Argument{}

	case "currentUser":
		return CurrentUser{}

	case "zero":
		return Zero{}

	case "newBool":
		return NewBool{ExpressionFromValue(u.Value)}

	case "newInt8":
		return NewInt8{ExpressionFromValue(u.Value)}

	case "newInt16":
		return NewInt16{ExpressionFromValue(u.Value)}

	case "newInt32":
		return NewInt32{ExpressionFromValue(u.Value)}

	case "newInt64", "newInt", "floatToInt":
		return NewInt64{ExpressionFromValue(u.Value)}

	case "newUint8":
		return NewUint8{ExpressionFromValue(u.Value)}

	case "newUint16":
		return NewUint16{ExpressionFromValue(u.Value)}

	case "newUint32":
		return NewUint32{ExpressionFromValue(u.Value)}

	case "newUint64", "newUint":
		return NewUint64{ExpressionFromValue(u.Value)}

	case "newFloat", "intToFloat":
		return NewFloat{ExpressionFromValue(u.Value)}

	case "newString":
		return NewString{ExpressionFromValue(u.Value)}

	case "newDateTime":
		return NewDateTime{ExpressionFromValue(u.Value)}

	case "newRef":
		arg := u.Value.(val.Struct)
		return NewRef{ExpressionFromValue(arg["model"]), ExpressionFromValue(arg["id"])}

	case "newUnion":
		arg := u.Value.(val.Struct)
		return NewUnion{ExpressionFromValue(arg["case"]), ExpressionFromValue(arg["value"])}

	case "newList":
		arg := u.Value.(val.List)
		nod := make(NewList, len(arg), len(arg))
		for i, sub := range arg {
			nod[i] = ExpressionFromValue(sub)
		}
		return nod

	case "newTuple":
		arg := u.Value.(val.List)
		nod := make(NewTuple, len(arg), len(arg))
		for i, sub := range arg {
			nod[i] = ExpressionFromValue(sub)
		}
		return nod

	case "newMap":
		arg := u.Value.(val.Map)
		nod := make(NewMap, len(arg))
		for k, sub := range arg {
			nod[k] = ExpressionFromValue(sub)
		}
		return nod

	case "newStruct":
		arg := u.Value.(val.Map)
		nod := make(NewStruct, len(arg))
		for k, sub := range arg {
			nod[k] = ExpressionFromValue(sub)
		}
		return nod

	case "contextual", "static":
		return Literal{u.Value}

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
		if _, ok := arg["in"]; !ok {
			arg["in"] = val.Union{"arg", val.Struct{}}
		}
		return SetField{ExpressionFromValue(arg["name"]), ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["in"])}

	case "setKey":
		arg := u.Value.(val.Struct)
		if _, ok := arg["in"]; !ok {
			arg["in"] = val.Union{"arg", val.Struct{}}
		}
		return SetKey{ExpressionFromValue(arg["name"]), ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["in"])}

	case "joinStrings":
		arg := u.Value.(val.Struct)
		if _, ok := arg["separator"]; !ok {
			arg["separator"] = val.String("")
		}
		return JoinStrings{ExpressionFromValue(arg["strings"]), ExpressionFromValue(arg["separator"])}

	case "relocateRef":
		arg := u.Value.(val.Struct)
		return RelocateRef{ExpressionFromValue(arg["ref"]), ExpressionFromValue(arg["model"])}

	case "extractStrings":
		return ExtractStrings{ExpressionFromValue(u.Value)}

	case "delete":
		return Delete{ExpressionFromValue(u.Value)}

	case "createMultiple":
		arg := u.Value.(val.Struct)
		vls := arg["values"].(val.Map)
		mvs := make(map[string]Expression, len(vls))
		for k, sub := range vls {
			mvs[k] = ExpressionFromValue(sub)
		}
		return CreateMultiple{ExpressionFromValue(arg["in"]), mvs}

	case "if":
		arg := u.Value.(val.Struct)
		return If{ExpressionFromValue(arg["condition"]), ExpressionFromValue(arg["then"]), ExpressionFromValue(arg["else"])}

	case "with":
		arg := u.Value.(val.Struct)
		return With{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["return"])}

	case "update":
		arg := u.Value.(val.Struct)
		return Update{ExpressionFromValue(arg["ref"]), ExpressionFromValue(arg["value"])}

	case "create":
		arg := u.Value.(val.Struct)
		return Create{ExpressionFromValue(arg["in"]), ExpressionFromValue(arg["value"])}

	case "mapMap":
		arg := u.Value.(val.Struct)
		return MapMap{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["expression"])}

	case "mapList":
		arg := u.Value.(val.Struct)
		return MapList{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["expression"])}

	case "reduceList":
		arg := u.Value.(val.Struct)
		return ReduceList{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["expression"])}

	case "graphFlow":
		arg := u.Value.(val.Struct)
		flw := arg["flow"].(val.List)
		nod := GraphFlow{ExpressionFromValue(arg["start"]), make([]GraphFlowParam, len(flw), len(flw))}
		for i, sub := range flw {
			subArg := sub.(val.Struct)
			fwd, bwd := []Expression(nil), []Expression(nil)
			if fl, ok := subArg["forward"].(val.List); ok {
				fwd = make([]Expression, len(fl), len(fl))
				for i, sub := range fl {
					fwd[i] = ExpressionFromValue(sub)
				}
			}
			if bl, ok := subArg["backward"].(val.List); ok {
				bwd = make([]Expression, len(bl), len(bl))
				for i, sub := range bl {
					bwd[i] = ExpressionFromValue(sub)
				}
			}
			nod.Flow[i] = GraphFlowParam{ExpressionFromValue(subArg["from"]), fwd, bwd}
		}
		return nod

	case "resolveRefs":
		arg := u.Value.(val.Struct)
		mla := arg["models"].(val.List)
		mns := make([]Expression, len(mla), len(mla))
		for i, sub := range mla {
			mns[i] = ExpressionFromValue(sub)
		}
		return ResolveRefs{ExpressionFromValue(arg["value"]), mns}

	case "resolveAllRefs":
		return ResolveAllRefs{ExpressionFromValue(u.Value)}

	case "referred":
		arg := u.Value.(val.Struct)
		return Referred{ExpressionFromValue(arg["from"]), ExpressionFromValue(arg["in"])}

	case "referrers":
		arg := u.Value.(val.Struct)
		return Referrers{ExpressionFromValue(arg["of"]), ExpressionFromValue(arg["in"])}

	case "inList":
		arg := u.Value.(val.Struct)
		return InList{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["in"])}

	case "filter":
		arg := u.Value.(val.Struct)
		return Filter{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["expression"])}

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
		return AssertCase{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["case"])}

	case "isCase":
		arg := u.Value.(val.Struct)
		return IsCase{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["case"])}

	case "assertModelRef":
		arg, ok := u.Value.(val.Struct)
		if !ok {
			arg = val.Struct{
				"value": val.Null{},
				"ref":   u.Value,
			}
		}
		if arg["value"] == (val.Null{}) {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		return AssertModelRef{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["ref"])}

	case "slice":
		arg := u.Value.(val.Struct)
		if arg["value"] == (val.Null{}) {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		return Slice{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["offset"]), ExpressionFromValue(arg["length"])}

	case "searchAllRegex":
		if sarg, ok := u.Value.(val.String); ok {
			u.Value = val.Struct{
				"value":           val.Union{"arg", val.Struct{}},
				"regex":           sarg,
				"multiLine":       val.Bool(false),
				"caseInsensitive": val.Bool(false),
			}
		}
		arg := u.Value.(val.Struct)
		if arg["value"] == (val.Null{}) {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		if arg["multiLine"] == (val.Null{}) {
			arg["multiLine"] = val.Bool(false)
		}
		if arg["caseInsensitive"] == (val.Null{}) {
			arg["caseInsensitive"] = val.Bool(false)
		}
		return SearchAllRegex{
			ExpressionFromValue(arg["value"]),
			ExpressionFromValue(arg["regex"]),
			ExpressionFromValue(arg["multiLine"]),
			ExpressionFromValue(arg["caseInsensitive"]),
		}

	case "searchRegex":
		if sarg, ok := u.Value.(val.String); ok {
			u.Value = val.Struct{
				"value":           val.Union{"arg", val.Struct{}},
				"regex":           sarg,
				"multiLine":       val.Bool(false),
				"caseInsensitive": val.Bool(false),
			}
		}
		arg := u.Value.(val.Struct)
		if arg["value"] == (val.Null{}) {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		if arg["multiLine"] == (val.Null{}) {
			arg["multiLine"] = val.Bool(false)
		}
		if arg["caseInsensitive"] == (val.Null{}) {
			arg["caseInsensitive"] = val.Bool(false)
		}
		return SearchRegex{
			ExpressionFromValue(arg["value"]),
			ExpressionFromValue(arg["regex"]),
			ExpressionFromValue(arg["multiLine"]),
			ExpressionFromValue(arg["caseInsensitive"]),
		}

	case "matchRegex":
		if sarg, ok := u.Value.(val.String); ok {
			u.Value = val.Struct{
				"value":           val.Union{"arg", val.Struct{}},
				"regex":           sarg,
				"multiLine":       val.Bool(false),
				"caseInsensitive": val.Bool(false),
			}
		}
		arg := u.Value.(val.Struct)
		if arg["value"] == (val.Null{}) {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		if arg["multiLine"] == (val.Null{}) {
			arg["multiLine"] = val.Bool(false)
		}
		if arg["caseInsensitive"] == (val.Null{}) {
			arg["caseInsensitive"] = val.Bool(false)
		}
		return MatchRegex{
			ExpressionFromValue(arg["value"]),
			ExpressionFromValue(arg["regex"]),
			ExpressionFromValue(arg["multiLine"]),
			ExpressionFromValue(arg["caseInsensitive"]),
		}

	case "switchModelRef":
		arg := u.Value.(val.Struct)
		css := arg["cases"].(val.List)
		if arg["value"] == (val.Null{}) {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		nod := SwitchModelRef{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["default"]), make([]SwitchModelRefCase, len(css), len(css))}
		for i, sub := range css {
			subArg := sub.(val.Struct)
			nod.Cases[i] = SwitchModelRefCase{ExpressionFromValue(subArg["match"]), ExpressionFromValue(subArg["return"])}
		}
		return nod

	case "key":
		if arg, ok := u.Value.(val.String); ok {
			return Key{Argument{}, ExpressionFromValue(arg)}
		}
		arg := u.Value.(val.Struct)
		return Key{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["name"])}

	case "field":
		if arg, ok := u.Value.(val.String); ok {
			return Field{Argument{}, ExpressionFromValue(arg)}
		}
		arg := u.Value.(val.Struct)
		return Field{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["name"])}

	case "index":
		arg := u.Value.(val.Struct)
		return Index{ExpressionFromValue(arg["value"]), ExpressionFromValue(arg["number"])}

	case "not":
		return Not{ExpressionFromValue(u.Value)}

	case "modelOf":
		return ModelOf{ExpressionFromValue(u.Value)}

	case "metarialize":
		return Metarialize{ExpressionFromValue(u.Value)}

	case "refTo":
		return RefTo{ExpressionFromValue(u.Value)}

	case "switchType":
		arg := u.Value.(val.Struct)
		node := SwitchType{nil, make(map[string]Expression, len(arg)-1)}
		if _, ok := arg["value"]; !ok {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		node.Value = ExpressionFromValue(arg["value"])
		delete(arg, "value")
		for k, v := range arg {
			node.Cases[k] = ExpressionFromValue(v)
		}
		return node

	case "switchCase":
		arg := u.Value.(val.Struct)
		if _, ok := arg["value"]; !ok {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		cases := make(map[string]Expression)
		for k, v := range arg["cases"].(val.Map) {
			cases[k] = ExpressionFromValue(v)
		}
		return SwitchCase{
			Value: ExpressionFromValue(arg["value"]),
			Cases: cases,
		}

	case "memSort":
		arg := u.Value.(val.Struct)
		if _, ok := arg["value"]; !ok {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		return MemSort{
			Value:      ExpressionFromValue(arg["value"]),
			Expression: ExpressionFromValue(arg["expression"]),
		}

	case "mapSet":
		arg := u.Value.(val.Struct)
		if _, ok := arg["value"]; !ok {
			arg["value"] = val.Union{"arg", val.Struct{}}
		}
		return MapSet{
			Value:      ExpressionFromValue(arg["value"]),
			Expression: ExpressionFromValue(arg["expression"]),
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
		if node.Value.Primitive() {
			return node.Value
		}
		return val.Union{"static", node.Value}

	case SetField:
		return val.Union{"setField", val.Struct{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
			"in":    ValueFromExpression(node.In),
		}}
	case SetKey:
		return val.Union{"setKey", val.Struct{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
			"in":    ValueFromExpression(node.In),
		}}

	case NewBool:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newBool", arg}

	case NewInt8:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newInt8", arg}

	case NewInt16:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newInt16", arg}

	case NewInt32:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newInt32", arg}

	case NewInt64:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newInt64", arg}

	case NewUint8:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newUint8", arg}

	case NewUint16:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newUint16", arg}

	case NewUint32:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newUint32", arg}

	case NewUint64:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newUint64", arg}

	case NewFloat:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newFloat", arg}

	case NewString:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newString", arg}

	case NewDateTime:
		arg := ValueFromExpression(node.Argument)
		return val.Union{"newDateTime", arg}

	case NewRef:
		return val.Union{"newRef", val.Struct{
			"model": ValueFromExpression(node.Model),
			"id":    ValueFromExpression(node.Id),
		}}

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
		return val.Union{"joinStrings", val.Struct{
			"strings":   ValueFromExpression(node.Strings),
			"separator": ValueFromExpression(node.Separator), // TODO: separator elision in case it's ""
		}}

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
		return val.Union{"if", val.Struct{
			"condition": ValueFromExpression(node.Condition),
			"then":      ValueFromExpression(node.Then),
			"else":      ValueFromExpression(node.Else),
		}}

	case With:
		return val.Union{"with", val.Struct{
			"value":  ValueFromExpression(node.Value),
			"return": ValueFromExpression(node.Return),
		}}

	case Update:
		return val.Union{"update", val.Struct{
			"ref":   ValueFromExpression(node.Ref),
			"value": ValueFromExpression(node.Value),
		}}

	case Create:
		return val.Union{"create", val.Struct{
			"in":    ValueFromExpression(node.In),
			"value": ValueFromExpression(node.Value),
		}}

	case InList:
		return val.Union{"inList", val.Struct{
			"in":    ValueFromExpression(node.In),
			"value": ValueFromExpression(node.Value),
		}}

	case Filter:
		return val.Union{"filter", val.Struct{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		}}

	case AssertCase:
		return val.Union{"assertCase", val.Struct{
			"case":  ValueFromExpression(node.Case),
			"value": ValueFromExpression(node.Value),
		}}

	case IsCase:
		return val.Union{"isCase", val.Struct{
			"case":  ValueFromExpression(node.Case),
			"value": ValueFromExpression(node.Value),
		}}

	case MapMap:
		return val.Union{"mapMap", val.Struct{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		}}

	case MapList:
		return val.Union{"mapList", val.Struct{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		}}

	case ReduceList:
		return val.Union{"reduceList", val.Struct{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		}}

	case ResolveRefs:
		mls := make(val.List, len(node.Models), len(node.Models))
		for i, _ := range mls {
			mls[i] = ValueFromExpression(node.Models[i])
		}
		return val.Union{"resolveRefs", val.Struct{
			"value":  ValueFromExpression(node.Value),
			"models": mls,
		}}

	case Field:
		if node.Value == (Argument{}) {
			return val.Union{"field", ValueFromExpression(node.Name)}
		}
		return val.Union{"field", val.Struct{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
		}}

	case Key:
		if node.Value == (Argument{}) {
			return val.Union{"key", ValueFromExpression(node.Name)}
		}
		return val.Union{"key", val.Struct{
			"name":  ValueFromExpression(node.Name),
			"value": ValueFromExpression(node.Value),
		}}

	case Index:
		return val.Union{"index", val.Struct{
			"number": ValueFromExpression(node.Number),
			"value":  ValueFromExpression(node.Value),
		}}

	case NewList:
		return val.Union{"newList", make(val.List, len(node), len(node)).OverMap(func(i int, _ val.Value) val.Value {
			return ValueFromExpression(node[i])
		})}

	case NewTuple:
		return val.Union{"newTuple", make(val.Tuple, len(node), len(node)).OverMap(func(i int, _ val.Value) val.Value {
			return ValueFromExpression(node[i])
		})}

	case NewMap:
		values := make(val.Map, len(node))
		for k, sub := range node {
			values[k] = ValueFromExpression(sub)
		}
		return val.Union{"newMap", values}

	case NewStruct:
		values := make(val.Map, len(node))
		for k, sub := range node {
			values[k] = ValueFromExpression(sub)
		}
		return val.Union{"newStruct", values}

	case NewUnion:
		return val.Union{"newUnion", val.Struct{
			"value": ValueFromExpression(node.Value),
			"case":  ValueFromExpression(node.Case),
		}}

	case Referred:
		return val.Union{"referred", val.Struct{
			"from": ValueFromExpression(node.From),
			"in":   ValueFromExpression(node.In),
		}}

	case RelocateRef:
		return val.Union{"relocateRef", val.Struct{
			"ref":   ValueFromExpression(node.Ref),
			"model": ValueFromExpression(node.Model),
		}}

	case Referrers:
		return val.Union{"referrers", val.Struct{
			"of": ValueFromExpression(node.Of),
			"in": ValueFromExpression(node.In),
		}}

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
		values := make(val.Map, len(node.Values))
		for k, sub := range node.Values {
			values[k] = ValueFromExpression(sub)
		}
		return val.Union{"createMultiple", val.Struct{
			"in":     ValueFromExpression(node.In),
			"values": values,
		}}

	case Slice:
		arg := val.Struct{
			"offset": ValueFromExpression(node.Offset),
			"length": ValueFromExpression(node.Length),
		}
		if node.Value != (Argument{}) {
			arg["value"] = ValueFromExpression(node.Value)
		}
		return val.Union{"slice", arg}

	case MatchRegex:
		arg := val.Struct{
			"regex": ValueFromExpression(node.Regex),
		}
		if node.Value != (Argument{}) {
			arg["value"] = ValueFromExpression(node.Value)
		}
		if node.MultiLine != (Literal{val.Bool(false)}) {
			arg["multiLine"] = ValueFromExpression(node.MultiLine)
		}
		if node.CaseInsensitive != (Literal{val.Bool(false)}) {
			arg["caseInsensitive"] = ValueFromExpression(node.CaseInsensitive)
		}
		return val.Union{"matchRegex", arg}

	case AssertModelRef:
		return val.Union{"assertModelRef", val.Struct{
			"value": ValueFromExpression(node.Value),
			"ref":   ValueFromExpression(node.Ref),
		}}

	case SwitchModelRef:
		cases := make(val.List, len(node.Cases), len(node.Cases))
		for i, caze := range node.Cases {
			cases[i] = val.Struct{
				"match":  ValueFromExpression(caze.Match),
				"return": ValueFromExpression(caze.Return),
			}
		}
		return val.Union{"switchModelRef", val.Struct{
			"value":   ValueFromExpression(node.Value),
			"default": ValueFromExpression(node.Default),
			"cases":   cases,
		}}
	case GraphFlow:
		flow := make(val.List, len(node.Flow), len(node.Flow))
		for i, sub := range node.Flow {
			flow[i] = val.Struct{
				"from": ValueFromExpression(sub.From),
				"forward": make(val.List, len(sub.Forward), len(sub.Forward)).OverMap(func(i int, _ val.Value) val.Value {
					return ValueFromExpression(sub.Forward[i])
				}),
				"backward": make(val.List, len(sub.Backward), len(sub.Backward)).OverMap(func(i int, _ val.Value) val.Value {
					return ValueFromExpression(sub.Backward[i])
				}),
			}
		}
		return val.Union{"graphFlow", val.Struct{
			"start": ValueFromExpression(node.Start),
			"flow":  flow,
		}}

	case SwitchType:
		args := make(val.Struct, len(node.Cases))
		for k, subNode := range node.Cases {
			args[k] = ValueFromExpression(subNode)
		}
		args["value"] = ValueFromExpression(node.Value)
		return val.Union{"switchType", args}

	case SwitchCase:
		cases := make(val.Map, len(node.Cases))
		for k, v := range node.Cases {
			cases[k] = ValueFromExpression(v)
		}
		return val.Union{"switchCase", val.Struct{
			"value": ValueFromExpression(node.Value),
			"cases": cases,
		}}

	case MemSort:
		return val.Union{"memSort", val.Struct{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		}}

	case MapSet:
		return val.Union{"mapSet", val.Struct{
			"value":      ValueFromExpression(node.Value),
			"expression": ValueFromExpression(node.Expression),
		}}

	default:
		panic(fmt.Sprintf("unhandled case: %T", node))

	}
}
