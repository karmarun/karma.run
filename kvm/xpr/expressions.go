// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"github.com/karmarun/karma.run/kvm/val"
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

type ReverseList struct {
	Argument Expression
}

func (x ReverseList) Transform(f func(Expression) Expression) Expression {
	return f(ReverseList{f(x.Argument)})
}

type StringToLower struct {
	Argument Expression
}

func (x StringToLower) Transform(f func(Expression) Expression) Expression {
	return f(StringToLower{f(x.Argument)})
}

type ExtractStrings struct {
	Argument Expression
}

func (x ExtractStrings) Transform(f func(Expression) Expression) Expression {
	return f(ExtractStrings{f(x.Argument)})
}

type NewBool struct {
	Argument Expression
}

func (x NewBool) Transform(f func(Expression) Expression) Expression {
	return f(NewBool{f(x.Argument)})
}

type NewInt8 struct {
	Argument Expression
}

func (x NewInt8) Transform(f func(Expression) Expression) Expression {
	return f(NewInt8{f(x.Argument)})
}

type NewInt16 struct {
	Argument Expression
}

func (x NewInt16) Transform(f func(Expression) Expression) Expression {
	return f(NewInt16{f(x.Argument)})
}

type NewInt32 struct {
	Argument Expression
}

func (x NewInt32) Transform(f func(Expression) Expression) Expression {
	return f(NewInt32{f(x.Argument)})
}

type NewInt64 struct {
	Argument Expression
}

func (x NewInt64) Transform(f func(Expression) Expression) Expression {
	return f(NewInt64{f(x.Argument)})
}

type NewUint8 struct {
	Argument Expression
}

func (x NewUint8) Transform(f func(Expression) Expression) Expression {
	return f(NewUint8{f(x.Argument)})
}

type NewUint16 struct {
	Argument Expression
}

func (x NewUint16) Transform(f func(Expression) Expression) Expression {
	return f(NewUint16{f(x.Argument)})
}

type NewUint32 struct {
	Argument Expression
}

func (x NewUint32) Transform(f func(Expression) Expression) Expression {
	return f(NewUint32{f(x.Argument)})
}

type NewUint64 struct {
	Argument Expression
}

func (x NewUint64) Transform(f func(Expression) Expression) Expression {
	return f(NewUint64{f(x.Argument)})
}

type NewFloat struct {
	Argument Expression
}

func (x NewFloat) Transform(f func(Expression) Expression) Expression {
	return f(NewFloat{f(x.Argument)})
}

type NewString struct {
	Argument Expression
}

func (x NewString) Transform(f func(Expression) Expression) Expression {
	return f(NewString{f(x.Argument)})
}

type NewDateTime struct {
	Argument Expression
}

func (x NewDateTime) Transform(f func(Expression) Expression) Expression {
	return f(NewDateTime{f(x.Argument)})
}

type IsPresent struct {
	Argument Expression
}

func (x IsPresent) Transform(f func(Expression) Expression) Expression {
	return f(IsPresent{f(x.Argument)})
}

type PresentOrZero struct {
	Argument Expression
}

func (x PresentOrZero) Transform(f func(Expression) Expression) Expression {
	return f(PresentOrZero{f(x.Argument)})
}

type AssertPresent struct {
	Argument Expression
}

func (x AssertPresent) Transform(f func(Expression) Expression) Expression {
	return f(AssertPresent{f(x.Argument)})
}

type AssertAbsent struct {
	Argument Expression
}

func (x AssertAbsent) Transform(f func(Expression) Expression) Expression {
	return f(AssertAbsent{f(x.Argument)})
}

type Model struct {
	Argument Expression
}

func (x Model) Transform(f func(Expression) Expression) Expression {
	return f(Model{f(x.Argument)})
}

type Tag struct {
	Argument Expression
}

func (x Tag) Transform(f func(Expression) Expression) Expression {
	return f(Tag{f(x.Argument)})
}

type All struct {
	Argument Expression
}

func (x All) Transform(f func(Expression) Expression) Expression {
	return f(All{f(x.Argument)})
}

type Delete struct {
	Argument Expression
}

func (x Delete) Transform(f func(Expression) Expression) Expression {
	return f(Delete{f(x.Argument)})
}

type ResolveAllRefs struct {
	Argument Expression
}

func (x ResolveAllRefs) Transform(f func(Expression) Expression) Expression {
	return f(ResolveAllRefs{f(x.Argument)})
}

type First struct {
	Argument Expression
}

func (x First) Transform(f func(Expression) Expression) Expression {
	return f(First{f(x.Argument)})
}

type Get struct {
	Argument Expression
}

func (x Get) Transform(f func(Expression) Expression) Expression {
	return f(Get{f(x.Argument)})
}

type Length struct {
	Argument Expression
}

func (x Length) Transform(f func(Expression) Expression) Expression {
	return f(Length{f(x.Argument)})
}

type Not struct {
	Argument Expression
}

func (x Not) Transform(f func(Expression) Expression) Expression {
	return f(Not{f(x.Argument)})
}

type ModelOf struct {
	Argument Expression
}

func (x ModelOf) Transform(f func(Expression) Expression) Expression {
	return f(ModelOf{f(x.Argument)})
}

type Metarialize struct {
	Argument Expression
}

func (x Metarialize) Transform(f func(Expression) Expression) Expression {
	return f(Metarialize{f(x.Argument)})
}

type RefTo struct {
	Argument Expression
}

func (x RefTo) Transform(f func(Expression) Expression) Expression {
	return f(RefTo{f(x.Argument)})
}

type NewRef struct {
	Model, Id Expression
}

func (x NewRef) Transform(f func(Expression) Expression) Expression {
	return f(NewRef{f(x.Model), f(x.Id)})
}

type With struct {
	Value, Return Expression
}

func (x With) Transform(f func(Expression) Expression) Expression {
	return f(With{f(x.Value), f(x.Return)})
}

type If struct {
	Condition, Then, Else Expression
}

func (x If) Transform(f func(Expression) Expression) Expression {
	return f(If{f(x.Condition), f(x.Then), f(x.Else)})
}

type JoinStrings struct {
	Strings, Separator Expression
}

func (x JoinStrings) Transform(f func(Expression) Expression) Expression {
	return f(JoinStrings{f(x.Strings), f(x.Separator)})
}

type Update struct {
	Ref, Value Expression
}

func (x Update) Transform(f func(Expression) Expression) Expression {
	return f(Update{f(x.Ref), f(x.Value)})
}

type Create struct {
	In, Value Expression
}

func (x Create) Transform(f func(Expression) Expression) Expression {
	return f(Create{f(x.In), f(x.Value)})
}

type InList struct {
	Value, In Expression
}

func (x InList) Transform(f func(Expression) Expression) Expression {
	return f(InList{f(x.Value), f(x.In)})
}

type Filter struct {
	Value, Expression Expression
}

func (x Filter) Transform(f func(Expression) Expression) Expression {
	return f(Filter{f(x.Value), f(x.Expression)})
}

type AssertCase struct {
	Value, Case Expression
}

func (x AssertCase) Transform(f func(Expression) Expression) Expression {
	return f(AssertCase{f(x.Value), f(x.Case)})
}

type IsCase struct {
	Value, Case Expression
}

func (x IsCase) Transform(f func(Expression) Expression) Expression {
	return f(IsCase{f(x.Value), f(x.Case)})
}

type MapMap struct {
	Value, Expression Expression
}

func (x MapMap) Transform(f func(Expression) Expression) Expression {
	return f(MapMap{f(x.Value), f(x.Expression)})
}

type MapList struct {
	Value, Expression Expression
}

func (x MapList) Transform(f func(Expression) Expression) Expression {
	return f(MapList{f(x.Value), f(x.Expression)})
}

type ReduceList struct {
	Value, Expression Expression
}

func (x ReduceList) Transform(f func(Expression) Expression) Expression {
	return f(ReduceList{f(x.Value), f(x.Expression)})
}

type ResolveRefs struct {
	Value  Expression
	Models []Expression
}

func (x ResolveRefs) Transform(f func(Expression) Expression) Expression {
	return f(ResolveRefs{f(x.Value), mapExpressions(x.Models, f)})
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
	return f(GraphFlow{f(x.Start), ps})
}

type AssertModelRef struct {
	Value, Ref Expression
}

func (x AssertModelRef) Transform(f func(Expression) Expression) Expression {
	return f(AssertModelRef{f(x.Value), f(x.Ref)})
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
	return f(SwitchModelRef{f(x.Value), f(x.Default), cs})
}

type Slice struct {
	Value  Expression
	Offset Expression
	Length Expression
}

func (x Slice) Transform(f func(Expression) Expression) Expression {
	return f(Slice{f(x.Value), f(x.Offset), f(x.Length)})
}

type SearchAllRegex struct {
	Value           Expression
	Regex           Expression
	MultiLine       Expression
	CaseInsensitive Expression
}

func (x SearchAllRegex) Transform(f func(Expression) Expression) Expression {
	return f(SearchAllRegex{f(x.Value), f(x.Regex), f(x.MultiLine), f(x.CaseInsensitive)})
}

type SearchRegex struct {
	Value           Expression
	Regex           Expression
	MultiLine       Expression
	CaseInsensitive Expression
}

func (x SearchRegex) Transform(f func(Expression) Expression) Expression {
	return f(SearchRegex{f(x.Value), f(x.Regex), f(x.MultiLine), f(x.CaseInsensitive)})
}

type MatchRegex struct {
	Value           Expression
	Regex           Expression
	MultiLine       Expression
	CaseInsensitive Expression
}

func (x MatchRegex) Transform(f func(Expression) Expression) Expression {
	return f(MatchRegex{f(x.Value), f(x.Regex), f(x.MultiLine), f(x.CaseInsensitive)})
}

type SetField struct {
	Name  Expression
	Value Expression
	In    Expression
}

func (x SetField) Transform(f func(Expression) Expression) Expression {
	return f(SetField{f(x.Name), f(x.Value), f(x.In)})
}

type SetKey struct {
	Name  Expression
	Value Expression
	In    Expression
}

func (x SetKey) Transform(f func(Expression) Expression) Expression {
	return f(SetKey{f(x.Name), f(x.Value), f(x.In)})
}

type Field struct {
	Value, Name Expression
}

func (x Field) Transform(f func(Expression) Expression) Expression {
	return f(Field{f(x.Value), f(x.Name)})
}

type Key struct {
	Value, Name Expression
}

func (x Key) Transform(f func(Expression) Expression) Expression {
	return f(Key{f(x.Value), f(x.Name)})
}

type Index struct {
	Value, Number Expression
}

func (x Index) Transform(f func(Expression) Expression) Expression {
	return f(Index{f(x.Value), f(x.Number)})
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
	return f(NewUnion{f(x.Case), f(x.Value)})
}

type CreateMultiple struct {
	In     Expression
	Values map[string]Expression
}

func (x CreateMultiple) Transform(f func(Expression) Expression) Expression {
	return f(CreateMultiple{f(x.In), mapExpressionMap(x.Values, f)})
}

type Referred struct {
	From, In Expression
}

func (x Referred) Transform(f func(Expression) Expression) Expression {
	return f(Referred{f(x.From), f(x.In)})
}

type Referrers struct {
	Of, In Expression
}

func (x Referrers) Transform(f func(Expression) Expression) Expression {
	return f(Referrers{f(x.Of), f(x.In)})
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
	return f(RelocateRef{f(x.Ref), f(x.Model)})
}

type SwitchType struct {
	Value Expression
	Cases map[string]Expression
}

func (x SwitchType) Transform(f func(Expression) Expression) Expression {
	return f(SwitchType{f(x.Value), mapExpressionMap(x.Cases, f)})
}

type SwitchCase struct {
	Value Expression
	Cases map[string]Expression
}

func (x SwitchCase) Transform(f func(Expression) Expression) Expression {
	return f(SwitchCase{f(x.Value), mapExpressionMap(x.Cases, f)})
}

type MapSet struct {
	Value      Expression
	Expression Expression
}

func (x MapSet) Transform(f func(Expression) Expression) Expression {
	return f(MapSet{f(x.Value), f(x.Expression)})
}

type MemSort struct {
	Value      Expression
	Expression Expression
}

func (x MemSort) Transform(f func(Expression) Expression) Expression {
	return f(MemSort{f(x.Value), f(x.Expression)})
}

type Do map[string]Expression

func (x Do) Transform(f func(Expression) Expression) Expression {
	return f(Do(mapExpressionMap(x, f)))
}

type Bind string

func (x Bind) Transform(f func(Expression) Expression) Expression {
	return f(x)
}
