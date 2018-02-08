// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"karma.run/kvm/val"
)

type Argument struct{}

func (x Argument) Transform(f func(Expression) Expression) Expression {
	return f(x)
}

type CurrentUser struct{}

func (x CurrentUser) Transform(f func(Expression) Expression) Expression {
	return f(x)
}

type Zero struct{}

func (x Zero) Transform(f func(Expression) Expression) Expression {
	return f(x)
}

type Literal struct {
	Value val.Value
}

func (x Literal) Transform(f func(Expression) Expression) Expression {
	return f(x)
}

type AllReferrers struct {
	Argument Expression
}

func (x AllReferrers) Transform(f func(Expression) Expression) Expression {
	return f(AllReferrers{x.Argument.Transform(f)})
}

type ReverseList struct {
	Argument Expression
}

func (x ReverseList) Transform(f func(Expression) Expression) Expression {
	return f(ReverseList{x.Argument.Transform(f)})
}

type StringToLower struct {
	Argument Expression
}

func (x StringToLower) Transform(f func(Expression) Expression) Expression {
	return f(StringToLower{x.Argument.Transform(f)})
}

type ExtractStrings struct {
	Argument Expression
}

func (x ExtractStrings) Transform(f func(Expression) Expression) Expression {
	return f(ExtractStrings{x.Argument.Transform(f)})
}

type NewBool struct {
	Argument Expression
}

func (x NewBool) Transform(f func(Expression) Expression) Expression {
	return f(NewBool{x.Argument.Transform(f)})
}

type NewInt8 struct {
	Argument Expression
}

func (x NewInt8) Transform(f func(Expression) Expression) Expression {
	return f(NewInt8{x.Argument.Transform(f)})
}

type NewInt16 struct {
	Argument Expression
}

func (x NewInt16) Transform(f func(Expression) Expression) Expression {
	return f(NewInt16{x.Argument.Transform(f)})
}

type NewInt32 struct {
	Argument Expression
}

func (x NewInt32) Transform(f func(Expression) Expression) Expression {
	return f(NewInt32{x.Argument.Transform(f)})
}

type NewInt64 struct {
	Argument Expression
}

func (x NewInt64) Transform(f func(Expression) Expression) Expression {
	return f(NewInt64{x.Argument.Transform(f)})
}

type NewUint8 struct {
	Argument Expression
}

func (x NewUint8) Transform(f func(Expression) Expression) Expression {
	return f(NewUint8{x.Argument.Transform(f)})
}

type NewUint16 struct {
	Argument Expression
}

func (x NewUint16) Transform(f func(Expression) Expression) Expression {
	return f(NewUint16{x.Argument.Transform(f)})
}

type NewUint32 struct {
	Argument Expression
}

func (x NewUint32) Transform(f func(Expression) Expression) Expression {
	return f(NewUint32{x.Argument.Transform(f)})
}

type NewUint64 struct {
	Argument Expression
}

func (x NewUint64) Transform(f func(Expression) Expression) Expression {
	return f(NewUint64{x.Argument.Transform(f)})
}

type NewFloat struct {
	Argument Expression
}

func (x NewFloat) Transform(f func(Expression) Expression) Expression {
	return f(NewFloat{x.Argument.Transform(f)})
}

type NewString struct {
	Argument Expression
}

func (x NewString) Transform(f func(Expression) Expression) Expression {
	return f(NewString{x.Argument.Transform(f)})
}

type NewDateTime struct {
	Argument Expression
}

func (x NewDateTime) Transform(f func(Expression) Expression) Expression {
	return f(NewDateTime{x.Argument.Transform(f)})
}

type IsPresent struct {
	Argument Expression
}

func (x IsPresent) Transform(f func(Expression) Expression) Expression {
	return f(IsPresent{x.Argument.Transform(f)})
}

type PresentOrZero struct {
	Argument Expression
}

func (x PresentOrZero) Transform(f func(Expression) Expression) Expression {
	return f(PresentOrZero{x.Argument.Transform(f)})
}

type AssertPresent struct {
	Argument Expression
}

func (x AssertPresent) Transform(f func(Expression) Expression) Expression {
	return f(AssertPresent{x.Argument.Transform(f)})
}

type AssertAbsent struct {
	Argument Expression
}

func (x AssertAbsent) Transform(f func(Expression) Expression) Expression {
	return f(AssertAbsent{x.Argument.Transform(f)})
}

type Model struct {
	Argument Expression
}

func (x Model) Transform(f func(Expression) Expression) Expression {
	return f(Model{x.Argument.Transform(f)})
}

type Tag struct {
	Argument Expression
}

func (x Tag) Transform(f func(Expression) Expression) Expression {
	return f(Tag{x.Argument.Transform(f)})
}

type All struct {
	Argument Expression
}

func (x All) Transform(f func(Expression) Expression) Expression {
	return f(All{x.Argument.Transform(f)})
}

type Delete struct {
	Argument Expression
}

func (x Delete) Transform(f func(Expression) Expression) Expression {
	return f(Delete{x.Argument.Transform(f)})
}

type ResolveAllRefs struct {
	Argument Expression
}

func (x ResolveAllRefs) Transform(f func(Expression) Expression) Expression {
	return f(ResolveAllRefs{x.Argument.Transform(f)})
}

type First struct {
	Argument Expression
}

func (x First) Transform(f func(Expression) Expression) Expression {
	return f(First{x.Argument.Transform(f)})
}

type Get struct {
	Argument Expression
}

func (x Get) Transform(f func(Expression) Expression) Expression {
	return f(Get{x.Argument.Transform(f)})
}

type Length struct {
	Argument Expression
}

func (x Length) Transform(f func(Expression) Expression) Expression {
	return f(Length{x.Argument.Transform(f)})
}

type Not struct {
	Argument Expression
}

func (x Not) Transform(f func(Expression) Expression) Expression {
	return f(Not{x.Argument.Transform(f)})
}

type ModelOf struct {
	Argument Expression
}

func (x ModelOf) Transform(f func(Expression) Expression) Expression {
	return f(ModelOf{x.Argument.Transform(f)})
}

type Metarialize struct {
	Argument Expression
}

func (x Metarialize) Transform(f func(Expression) Expression) Expression {
	return f(Metarialize{x.Argument.Transform(f)})
}

type RefTo struct {
	Argument Expression
}

func (x RefTo) Transform(f func(Expression) Expression) Expression {
	return f(RefTo{x.Argument.Transform(f)})
}

type NewRef struct {
	Model, Id Expression
}

func (x NewRef) Transform(f func(Expression) Expression) Expression {
	return f(NewRef{x.Model.Transform(f), x.Id.Transform(f)})
}

type With struct {
	Value, Return Expression
}

func (x With) Transform(f func(Expression) Expression) Expression {
	return f(With{x.Value.Transform(f), x.Return.Transform(f)})
}

type If struct {
	Condition, Then, Else Expression
}

func (x If) Transform(f func(Expression) Expression) Expression {
	return f(If{x.Condition.Transform(f), x.Then.Transform(f), x.Else.Transform(f)})
}

type JoinStrings struct {
	Strings, Separator Expression
}

func (x JoinStrings) Transform(f func(Expression) Expression) Expression {
	return f(JoinStrings{x.Strings.Transform(f), x.Separator.Transform(f)})
}

type Update struct {
	Ref, Value Expression
}

func (x Update) Transform(f func(Expression) Expression) Expression {
	return f(Update{x.Ref.Transform(f), x.Value.Transform(f)})
}

type Create struct {
	In, Value Expression
}

func (x Create) Transform(f func(Expression) Expression) Expression {
	return f(Create{x.In.Transform(f), x.Value.Transform(f)})
}

type InList struct {
	Value, In Expression
}

func (x InList) Transform(f func(Expression) Expression) Expression {
	return f(InList{x.Value.Transform(f), x.In.Transform(f)})
}

type Filter struct {
	Value, Expression Expression
}

func (x Filter) Transform(f func(Expression) Expression) Expression {
	return f(Filter{x.Value.Transform(f), x.Expression.Transform(f)})
}

type AssertCase struct {
	Value, Case Expression
}

func (x AssertCase) Transform(f func(Expression) Expression) Expression {
	return f(AssertCase{x.Value.Transform(f), x.Case.Transform(f)})
}

type IsCase struct {
	Value, Case Expression
}

func (x IsCase) Transform(f func(Expression) Expression) Expression {
	return f(IsCase{x.Value.Transform(f), x.Case.Transform(f)})
}

type MapMap struct {
	Value, Expression Expression
}

func (x MapMap) Transform(f func(Expression) Expression) Expression {
	return f(MapMap{x.Value.Transform(f), x.Expression.Transform(f)})
}

type MapList struct {
	Value, Expression Expression
}

func (x MapList) Transform(f func(Expression) Expression) Expression {
	return f(MapList{x.Value.Transform(f), x.Expression.Transform(f)})
}

type ReduceList struct {
	Value, Expression Expression
}

func (x ReduceList) Transform(f func(Expression) Expression) Expression {
	return f(ReduceList{x.Value.Transform(f), x.Expression.Transform(f)})
}

type ResolveRefs struct {
	Value  Expression
	Models []Expression
}

func (x ResolveRefs) Transform(f func(Expression) Expression) Expression {
	return f(ResolveRefs{x.Value.Transform(f), mapExpressions(x.Models, f)})
}

type GraphFlow struct {
	Start Expression
	Flow  []GraphFlowParam
}

// GraphFlowParam is not an Expression
type GraphFlowParam struct {
	From     Expression
	Forward  []Expression
	Backward []Expression
}

// GraphFlowParam is not an Expression
func (p GraphFlowParam) _transform(f func(Expression) Expression) GraphFlowParam {
	return GraphFlowParam{f(p.From), mapExpressions(p.Forward, f), mapExpressions(p.Backward, f)}
}

func (x GraphFlow) Transform(f func(Expression) Expression) Expression {
	ps := make([]GraphFlowParam, len(x.Flow), len(x.Flow))
	for i, p := range x.Flow {
		ps[i] = p._transform(f)
	}
	return f(GraphFlow{x.Start.Transform(f), ps})
}

type AssertModelRef struct {
	Value, Ref Expression
}

func (x AssertModelRef) Transform(f func(Expression) Expression) Expression {
	return f(AssertModelRef{x.Value.Transform(f), x.Ref.Transform(f)})
}

type SwitchModelRef struct {
	Value   Expression
	Default Expression
	Cases   []SwitchModelRefCase
}

// SwitchModelRefCase is not an Expression
type SwitchModelRefCase struct {
	Match, Return Expression
}

// SwitchModelRefCase is not an Expression
func (c SwitchModelRefCase) _transform(f func(Expression) Expression) SwitchModelRefCase {
	return SwitchModelRefCase{f(c.Match), f(c.Return)}
}

func (x SwitchModelRef) Transform(f func(Expression) Expression) Expression {
	cs := make([]SwitchModelRefCase, len(x.Cases), len(x.Cases))
	for i, c := range x.Cases {
		cs[i] = c._transform(f)
	}
	return f(SwitchModelRef{x.Value.Transform(f), x.Default.Transform(f), cs})
}

type Slice struct {
	Value  Expression
	Offset Expression
	Length Expression
}

func (x Slice) Transform(f func(Expression) Expression) Expression {
	return f(Slice{x.Value.Transform(f), x.Offset.Transform(f), x.Length.Transform(f)})
}

type SearchAllRegex struct {
	Value           Expression
	Regex           Expression
	MultiLine       Expression
	CaseInsensitive Expression
}

func (x SearchAllRegex) Transform(f func(Expression) Expression) Expression {
	return f(SearchAllRegex{x.Value.Transform(f), x.Regex.Transform(f), x.MultiLine.Transform(f), x.CaseInsensitive.Transform(f)})
}

type SearchRegex struct {
	Value           Expression
	Regex           Expression
	MultiLine       Expression
	CaseInsensitive Expression
}

func (x SearchRegex) Transform(f func(Expression) Expression) Expression {
	return f(SearchRegex{x.Value.Transform(f), x.Regex.Transform(f), x.MultiLine.Transform(f), x.CaseInsensitive.Transform(f)})
}

type MatchRegex struct {
	Value           Expression
	Regex           Expression
	MultiLine       Expression
	CaseInsensitive Expression
}

func (x MatchRegex) Transform(f func(Expression) Expression) Expression {
	return f(MatchRegex{x.Value.Transform(f), x.Regex.Transform(f), x.MultiLine.Transform(f), x.CaseInsensitive.Transform(f)})
}

type SetField struct {
	Name  Expression
	Value Expression
	In    Expression
}

func (x SetField) Transform(f func(Expression) Expression) Expression {
	return f(SetField{x.Name.Transform(f), x.Value.Transform(f), x.In.Transform(f)})
}

type SetKey struct {
	Name  Expression
	Value Expression
	In    Expression
}

func (x SetKey) Transform(f func(Expression) Expression) Expression {
	return f(SetKey{x.Name.Transform(f), x.Value.Transform(f), x.In.Transform(f)})
}

type Field struct {
	Value, Name Expression
}

func (x Field) Transform(f func(Expression) Expression) Expression {
	return f(Field{x.Value.Transform(f), x.Name.Transform(f)})
}

type Key struct {
	Value, Name Expression
}

func (x Key) Transform(f func(Expression) Expression) Expression {
	return f(Key{x.Value.Transform(f), x.Name.Transform(f)})
}

type Index struct {
	Value, Number Expression
}

func (x Index) Transform(f func(Expression) Expression) Expression {
	return f(Index{x.Value.Transform(f), x.Number.Transform(f)})
}

type NewSet []Expression

func (x NewSet) Transform(f func(Expression) Expression) Expression {
	return f(NewSet(mapExpressions(x, f)))
}

type NewList []Expression

func (x NewList) Transform(f func(Expression) Expression) Expression {
	return f(NewList(mapExpressions(x, f)))
}

type NewTuple []Expression

func (x NewTuple) Transform(f func(Expression) Expression) Expression {
	return f(NewTuple(mapExpressions(x, f)))
}

type NewMap map[string]Expression

func (x NewMap) Transform(f func(Expression) Expression) Expression {
	return f(NewMap(mapExpressionMap(x, f)))
}

type NewStruct map[string]Expression

func (x NewStruct) Transform(f func(Expression) Expression) Expression {
	return f(NewStruct(mapExpressionMap(x, f)))
}

type NewUnion struct {
	Case, Value Expression
}

func (x NewUnion) Transform(f func(Expression) Expression) Expression {
	return f(NewUnion{x.Case.Transform(f), x.Value.Transform(f)})
}

type CreateMultiple struct {
	In     Expression
	Values map[string]Expression
}

func (x CreateMultiple) Transform(f func(Expression) Expression) Expression {
	return f(CreateMultiple{x.In.Transform(f), mapExpressionMap(x.Values, f)})
}

type Referred struct {
	From, In Expression
}

func (x Referred) Transform(f func(Expression) Expression) Expression {
	return f(Referred{x.From.Transform(f), x.In.Transform(f)})
}

type Referrers struct {
	Of, In Expression
}

func (x Referrers) Transform(f func(Expression) Expression) Expression {
	return f(Referrers{x.Of.Transform(f), x.In.Transform(f)})
}

type ConcatLists [2]Expression

func (x ConcatLists) Transform(f func(Expression) Expression) Expression {
	return f(ConcatLists{f(x[0]), f(x[1])})
}

type After [2]Expression

func (x After) Transform(f func(Expression) Expression) Expression {
	return f(After{f(x[0]), f(x[1])})
}

type Before [2]Expression

func (x Before) Transform(f func(Expression) Expression) Expression {
	return f(Before{f(x[0]), f(x[1])})
}

type Equal [2]Expression

func (x Equal) Transform(f func(Expression) Expression) Expression {
	return f(Equal{f(x[0]), f(x[1])})
}

type Greater [2]Expression

func (x Greater) Transform(f func(Expression) Expression) Expression {
	return f(Greater{f(x[0]), f(x[1])})
}

type Less [2]Expression

func (x Less) Transform(f func(Expression) Expression) Expression {
	return f(Less{f(x[0]), f(x[1])})
}

type Add [2]Expression

func (x Add) Transform(f func(Expression) Expression) Expression {
	return f(Add{f(x[0]), f(x[1])})
}

type Subtract [2]Expression

func (x Subtract) Transform(f func(Expression) Expression) Expression {
	return f(Subtract{f(x[0]), f(x[1])})
}

type Multiply [2]Expression

func (x Multiply) Transform(f func(Expression) Expression) Expression {
	return f(Multiply{f(x[0]), f(x[1])})
}

type Divide [2]Expression

func (x Divide) Transform(f func(Expression) Expression) Expression {
	return f(Divide{f(x[0]), f(x[1])})
}

type And []Expression

func (x And) Transform(f func(Expression) Expression) Expression {
	return f(And(mapExpressions(x, f)))
}

type Or []Expression

func (x Or) Transform(f func(Expression) Expression) Expression {
	return f(Or(mapExpressions(x, f)))
}

type RelocateRef struct {
	Ref   Expression
	Model Expression
}

func (x RelocateRef) Transform(f func(Expression) Expression) Expression {
	return f(RelocateRef{x.Ref.Transform(f), x.Model.Transform(f)})
}

type SwitchCase struct {
	Value Expression
	Cases map[string]Expression
}

func (x SwitchCase) Transform(f func(Expression) Expression) Expression {
	return f(SwitchCase{x.Value.Transform(f), mapExpressionMap(x.Cases, f)})
}

type MapSet struct {
	Value      Expression
	Expression Expression
}

func (x MapSet) Transform(f func(Expression) Expression) Expression {
	return f(MapSet{x.Value.Transform(f), x.Expression.Transform(f)})
}

type MemSort struct {
	Value      Expression
	Expression Expression
}

func (x MemSort) Transform(f func(Expression) Expression) Expression {
	return f(MemSort{x.Value.Transform(f), x.Expression.Transform(f)})
}
