// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"karma.run/kvm/val"
)

type function struct {
	args  []string
	exprs []Expression
}

func NewFunction(args []string, exprs ...Expression) Function {
	return function{args, exprs}
}

func (f function) Parameters() []string {
	return f.args
}

func (f function) Expressions() []Expression {
	return f.exprs
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

type FunctionSignature struct {
	Function Function
}

func (x FunctionSignature) Transform(f func(Expression) Expression) Expression {
	return f(FunctionSignature{x.Function})
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
	Value  Expression
	Return Function
}

func (x With) Transform(f func(Expression) Expression) Expression {
	return f(With{x.Value.Transform(f), x.Return})
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
	In    Expression
	Value Function
}

func (x Create) Transform(f func(Expression) Expression) Expression {
	return f(Create{x.In.Transform(f), x.Value})
}

type InList struct {
	Value, In Expression
}

func (x InList) Transform(f func(Expression) Expression) Expression {
	return f(InList{x.Value.Transform(f), x.In.Transform(f)})
}

type FilterList struct {
	Value  Expression
	Filter Function
}

func (x FilterList) Transform(f func(Expression) Expression) Expression {
	return f(FilterList{x.Value.Transform(f), x.Filter})
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
	Value   Expression
	Mapping Function
}

func (x MapMap) Transform(f func(Expression) Expression) Expression {
	return f(MapMap{x.Value.Transform(f), x.Mapping})
}

type MapList struct {
	Value   Expression
	Mapping Function
}

func (x MapList) Transform(f func(Expression) Expression) Expression {
	return f(MapList{x.Value.Transform(f), x.Mapping})
}

type ReduceList struct {
	Value   Expression
	Initial Expression
	Reducer Function
}

func (x ReduceList) Transform(f func(Expression) Expression) Expression {
	return f(ReduceList{x.Value.Transform(f), x.Initial.Transform(f), x.Reducer})
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
	Name  string
	Value Expression
}

func (x Field) Transform(f func(Expression) Expression) Expression {
	return f(Field{x.Name, x.Value.Transform(f)})
}

type Key struct {
	Name, Value Expression
}

func (x Key) Transform(f func(Expression) Expression) Expression {
	return f(Key{x.Name.Transform(f), x.Value.Transform(f)})
}

type IndexTuple struct {
	Value, Number Expression
}

func (x IndexTuple) Transform(f func(Expression) Expression) Expression {
	return f(IndexTuple{x.Value.Transform(f), x.Number.Transform(f)})
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
	Values map[string]Function
}

func (x CreateMultiple) Transform(f func(Expression) Expression) Expression {
	return f(CreateMultiple{x.In.Transform(f), x.Values})
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
	return f(ConcatLists{x[0].Transform(f), x[1].Transform(f)})
}

type After [2]Expression

func (x After) Transform(f func(Expression) Expression) Expression {
	return f(After{x[0].Transform(f), x[1].Transform(f)})
}

type Before [2]Expression

func (x Before) Transform(f func(Expression) Expression) Expression {
	return f(Before{x[0].Transform(f), x[1].Transform(f)})
}

type Equal [2]Expression

func (x Equal) Transform(f func(Expression) Expression) Expression {
	return f(Equal{x[0].Transform(f), x[1].Transform(f)})
}

type SubFloat [2]Expression

func (x SubFloat) Transform(f func(Expression) Expression) Expression {
	return f(SubFloat{x[0].Transform(f), x[1].Transform(f)})
}

type SubInt64 [2]Expression

func (x SubInt64) Transform(f func(Expression) Expression) Expression {
	return f(SubInt64{x[0].Transform(f), x[1].Transform(f)})
}

type SubInt32 [2]Expression

func (x SubInt32) Transform(f func(Expression) Expression) Expression {
	return f(SubInt32{x[0].Transform(f), x[1].Transform(f)})
}

type SubInt16 [2]Expression

func (x SubInt16) Transform(f func(Expression) Expression) Expression {
	return f(SubInt16{x[0].Transform(f), x[1].Transform(f)})
}

type SubInt8 [2]Expression

func (x SubInt8) Transform(f func(Expression) Expression) Expression {
	return f(SubInt8{x[0].Transform(f), x[1].Transform(f)})
}

type SubUint64 [2]Expression

func (x SubUint64) Transform(f func(Expression) Expression) Expression {
	return f(SubUint64{x[0].Transform(f), x[1].Transform(f)})
}

type SubUint32 [2]Expression

func (x SubUint32) Transform(f func(Expression) Expression) Expression {
	return f(SubUint32{x[0].Transform(f), x[1].Transform(f)})
}

type SubUint16 [2]Expression

func (x SubUint16) Transform(f func(Expression) Expression) Expression {
	return f(SubUint16{x[0].Transform(f), x[1].Transform(f)})
}

type SubUint8 [2]Expression

func (x SubUint8) Transform(f func(Expression) Expression) Expression {
	return f(SubUint8{x[0].Transform(f), x[1].Transform(f)})
}

type MulFloat [2]Expression

func (x MulFloat) Transform(f func(Expression) Expression) Expression {
	return f(MulFloat{x[0].Transform(f), x[1].Transform(f)})
}

type MulInt64 [2]Expression

func (x MulInt64) Transform(f func(Expression) Expression) Expression {
	return f(MulInt64{x[0].Transform(f), x[1].Transform(f)})
}

type MulInt32 [2]Expression

func (x MulInt32) Transform(f func(Expression) Expression) Expression {
	return f(MulInt32{x[0].Transform(f), x[1].Transform(f)})
}

type MulInt16 [2]Expression

func (x MulInt16) Transform(f func(Expression) Expression) Expression {
	return f(MulInt16{x[0].Transform(f), x[1].Transform(f)})
}

type MulInt8 [2]Expression

func (x MulInt8) Transform(f func(Expression) Expression) Expression {
	return f(MulInt8{x[0].Transform(f), x[1].Transform(f)})
}

type MulUint64 [2]Expression

func (x MulUint64) Transform(f func(Expression) Expression) Expression {
	return f(MulUint64{x[0].Transform(f), x[1].Transform(f)})
}

type MulUint32 [2]Expression

func (x MulUint32) Transform(f func(Expression) Expression) Expression {
	return f(MulUint32{x[0].Transform(f), x[1].Transform(f)})
}

type MulUint16 [2]Expression

func (x MulUint16) Transform(f func(Expression) Expression) Expression {
	return f(MulUint16{x[0].Transform(f), x[1].Transform(f)})
}

type MulUint8 [2]Expression

func (x MulUint8) Transform(f func(Expression) Expression) Expression {
	return f(MulUint8{x[0].Transform(f), x[1].Transform(f)})
}

type DivFloat [2]Expression

func (x DivFloat) Transform(f func(Expression) Expression) Expression {
	return f(DivFloat{x[0].Transform(f), x[1].Transform(f)})
}

type DivInt64 [2]Expression

func (x DivInt64) Transform(f func(Expression) Expression) Expression {
	return f(DivInt64{x[0].Transform(f), x[1].Transform(f)})
}

type DivInt32 [2]Expression

func (x DivInt32) Transform(f func(Expression) Expression) Expression {
	return f(DivInt32{x[0].Transform(f), x[1].Transform(f)})
}

type DivInt16 [2]Expression

func (x DivInt16) Transform(f func(Expression) Expression) Expression {
	return f(DivInt16{x[0].Transform(f), x[1].Transform(f)})
}

type DivInt8 [2]Expression

func (x DivInt8) Transform(f func(Expression) Expression) Expression {
	return f(DivInt8{x[0].Transform(f), x[1].Transform(f)})
}

type DivUint64 [2]Expression

func (x DivUint64) Transform(f func(Expression) Expression) Expression {
	return f(DivUint64{x[0].Transform(f), x[1].Transform(f)})
}

type DivUint32 [2]Expression

func (x DivUint32) Transform(f func(Expression) Expression) Expression {
	return f(DivUint32{x[0].Transform(f), x[1].Transform(f)})
}

type DivUint16 [2]Expression

func (x DivUint16) Transform(f func(Expression) Expression) Expression {
	return f(DivUint16{x[0].Transform(f), x[1].Transform(f)})
}

type DivUint8 [2]Expression

func (x DivUint8) Transform(f func(Expression) Expression) Expression {
	return f(DivUint8{x[0].Transform(f), x[1].Transform(f)})
}

type AddFloat [2]Expression

func (x AddFloat) Transform(f func(Expression) Expression) Expression {
	return f(AddFloat{x[0].Transform(f), x[1].Transform(f)})
}

type AddInt64 [2]Expression

func (x AddInt64) Transform(f func(Expression) Expression) Expression {
	return f(AddInt64{x[0].Transform(f), x[1].Transform(f)})
}

type AddInt32 [2]Expression

func (x AddInt32) Transform(f func(Expression) Expression) Expression {
	return f(AddInt32{x[0].Transform(f), x[1].Transform(f)})
}

type AddInt16 [2]Expression

func (x AddInt16) Transform(f func(Expression) Expression) Expression {
	return f(AddInt16{x[0].Transform(f), x[1].Transform(f)})
}

type AddInt8 [2]Expression

func (x AddInt8) Transform(f func(Expression) Expression) Expression {
	return f(AddInt8{x[0].Transform(f), x[1].Transform(f)})
}

type AddUint64 [2]Expression

func (x AddUint64) Transform(f func(Expression) Expression) Expression {
	return f(AddUint64{x[0].Transform(f), x[1].Transform(f)})
}

type AddUint32 [2]Expression

func (x AddUint32) Transform(f func(Expression) Expression) Expression {
	return f(AddUint32{x[0].Transform(f), x[1].Transform(f)})
}

type AddUint16 [2]Expression

func (x AddUint16) Transform(f func(Expression) Expression) Expression {
	return f(AddUint16{x[0].Transform(f), x[1].Transform(f)})
}

type AddUint8 [2]Expression

func (x AddUint8) Transform(f func(Expression) Expression) Expression {
	return f(AddUint8{x[0].Transform(f), x[1].Transform(f)})
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
	Value   Expression
	Default Expression
	Cases   map[string]Function
}

func (x SwitchCase) Transform(f func(Expression) Expression) Expression {
	return f(SwitchCase{x.Value.Transform(f), x.Default.Transform(f), x.Cases})
}

type MapSet struct {
	Value   Expression
	Mapping Function
}

func (x MapSet) Transform(f func(Expression) Expression) Expression {
	return f(MapSet{x.Value.Transform(f), x.Mapping})
}

type MemSort struct {
	Value Expression
	Order Function
}

func (x MemSort) Transform(f func(Expression) Expression) Expression {
	return f(MemSort{x.Value.Transform(f), x.Order})
}

type Define struct {
	Name     string
	Argument Expression
}

func (x Define) Transform(f func(Expression) Expression) Expression {
	return f(Define{x.Name, x.Argument.Transform(f)})
}

type Scope string

func (x Scope) Transform(f func(Expression) Expression) Expression {
	return f(x)
}
