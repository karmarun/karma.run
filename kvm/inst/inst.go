// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package inst

import (
	"karma.run/kvm/val"
	"regexp"
)

type Instruction interface {
	_inst() // private interface
}

type Slice struct{}

type Substring struct{}

type DateTimeNow struct{}

type DateTimeDiff struct{}

type StringToLower struct{}

type AssertPresent struct{}

type Metarialize struct{}

type Length struct{}

type Pop struct{}

type PopToInput struct{}

type Not struct{}

type Tag struct{}

type ReverseList struct{}

type Equal struct{}

type PresentOrConstant struct {
	Constant val.Value
}

type IsPresent struct{}

type Before struct{}

type ConcatLists struct{}

type After struct{}

type AddInts struct{}

type AddFloats struct{}

type All struct{}

type First struct{}

type StringToRef struct {
	Model string
}

type Identity struct{}

type CurrentUser struct{}

type OrList struct{}

type InList struct{}

type Enforce struct{}

type Constant struct {
	val.Value
}

type SearchAllRegex struct {
	Regex *regexp.Regexp
}
type SearchRegex struct {
	Regex *regexp.Regexp
}
type MatchRegex struct {
	Regex *regexp.Regexp
}

type If struct {
	Then, Else Sequence
}

type AssertCase struct {
	Case string
}

type IsCase struct{}

type ShortCircuitAnd struct {
	Arity, Step int
}

type ShortCircuitOr struct {
	Arity, Step int
}

type UnsetKey struct {
	Key string
}

type UnsetField struct {
	Key string
}

type SetKey struct {
	Key string
}

type SetField struct {
	Field string
}

type Key struct{}

type Field struct {
	Key string
}

type IndexTuple struct {
	Number int
}

type Meta struct {
	Key string
}

type Limit struct {
	Offset int
	Length int
}

type BuildList struct {
	Length int
}

type BuildMap struct {
	Length int
}

type BuildStruct struct {
	Keys []string
}

type BuildUnion struct {
	Case string
}

type BuildTuple struct {
	Length int
}

type Referrers struct {
	In string
}

type Referred struct {
	In string
}

type ResolveRefs struct {
	Models map[string]struct{}
}
type ResolveAllRefs struct{}

type Deref struct{}

type CreateMultiple struct {
	Model string
}

type Update struct{}

type JoinStrings struct{}

type ExtractStrings struct{}

type Delete struct{}

type Sequence []Instruction

type TraverseGraph struct {
	Direction  string
	StartModel string
	EdgeFilter Sequence
}

type ReduceList struct {
	Expression Sequence
}

type MapList struct {
	Expression Sequence
}

type MapMap struct {
	Expression Sequence
}

type MapStruct struct {
	Expression Sequence
}

type Filter struct {
	Expression Sequence
}

type GraphFlowParam struct {
	Forward, Backward map[string]struct{}
}

type GraphFlow struct {
	FlowParams map[string]GraphFlowParam
}

type AssertModelRef struct {
	Model string
}

type RelocateRef struct {
	Model string
}

type SwitchModelRef struct {
	Cases   map[string]Sequence
	Default Sequence
}

type DebugPrintStack struct{}

type ToFloat struct{}
type ToInt struct{}
type ToInt8 struct{}
type ToInt16 struct{}
type ToInt32 struct{}
type ToInt64 struct{}
type ToUint struct{}
type ToUint8 struct{}
type ToUint16 struct{}
type ToUint32 struct{}
type ToUint64 struct{}
type ToString struct{}

type GreaterFloat struct{}
type GreaterInt struct{}
type GreaterInt8 struct{}
type GreaterInt16 struct{}
type GreaterInt32 struct{}
type GreaterInt64 struct{}
type GreaterUint struct{}
type GreaterUint8 struct{}
type GreaterUint16 struct{}
type GreaterUint32 struct{}
type GreaterUint64 struct{}
type LessFloat struct{}
type LessInt struct{}
type LessInt8 struct{}
type LessInt16 struct{}
type LessInt32 struct{}
type LessInt64 struct{}
type LessUint struct{}
type LessUint8 struct{}
type LessUint16 struct{}
type LessUint32 struct{}
type LessUint64 struct{}
type AddFloat struct{}
type AddInt struct{}
type AddInt8 struct{}
type AddInt16 struct{}
type AddInt32 struct{}
type AddInt64 struct{}
type AddUint struct{}
type AddUint8 struct{}
type AddUint16 struct{}
type AddUint32 struct{}
type AddUint64 struct{}
type SubtractFloat struct{}
type SubtractInt struct{}
type SubtractInt8 struct{}
type SubtractInt16 struct{}
type SubtractInt32 struct{}
type SubtractInt64 struct{}
type SubtractUint struct{}
type SubtractUint8 struct{}
type SubtractUint16 struct{}
type SubtractUint32 struct{}
type SubtractUint64 struct{}
type MultiplyFloat struct{}
type MultiplyInt struct{}
type MultiplyInt8 struct{}
type MultiplyInt16 struct{}
type MultiplyInt32 struct{}
type MultiplyInt64 struct{}
type MultiplyUint struct{}
type MultiplyUint8 struct{}
type MultiplyUint16 struct{}
type MultiplyUint32 struct{}
type MultiplyUint64 struct{}
type DivideFloat struct{}
type DivideInt struct{}
type DivideInt8 struct{}
type DivideInt16 struct{}
type DivideInt32 struct{}
type DivideInt64 struct{}
type DivideUint struct{}
type DivideUint8 struct{}
type DivideUint16 struct{}
type DivideUint32 struct{}
type DivideUint64 struct{}

type SwitchType map[string]Sequence
type SwitchCase map[string]Sequence
type MapSet struct {
	Expression Sequence
}

type MemSort struct {
	Expression Sequence
}

func (Identity) _inst()          {}
func (AddFloats) _inst()         {}
func (AddInts) _inst()           {}
func (All) _inst()               {}
func (BuildList) _inst()         {}
func (BuildMap) _inst()          {}
func (BuildStruct) _inst()       {}
func (BuildUnion) _inst()        {}
func (Constant) _inst()          {}
func (Equal) _inst()             {}
func (Field) _inst()             {}
func (Filter) _inst()            {}
func (First) _inst()             {}
func (Deref) _inst()             {}
func (Length) _inst()            {}
func (Limit) _inst()             {}
func (Meta) _inst()              {}
func (Sequence) _inst()          {}
func (StringToRef) _inst()       {}
func (Delete) _inst()            {}
func (Update) _inst()            {}
func (Metarialize) _inst()       {}
func (MapList) _inst()           {}
func (Tag) _inst()               {}
func (Pop) _inst()               {}
func (Referrers) _inst()         {}
func (ShortCircuitOr) _inst()    {}
func (ShortCircuitAnd) _inst()   {}
func (OrList) _inst()            {}
func (Enforce) _inst()           {}
func (InList) _inst()            {}
func (Not) _inst()               {}
func (SetField) _inst()          {}
func (UnsetField) _inst()        {}
func (SwitchCase) _inst()        {}
func (Key) _inst()               {}
func (AssertCase) _inst()        {}
func (SetKey) _inst()            {}
func (UnsetKey) _inst()          {}
func (AssertPresent) _inst()     {}
func (DateTimeNow) _inst()       {}
func (DateTimeDiff) _inst()      {}
func (MapMap) _inst()            {}
func (TraverseGraph) _inst()     {}
func (Substring) _inst()         {}
func (MapStruct) _inst()         {}
func (CreateMultiple) _inst()    {}
func (BuildTuple) _inst()        {}
func (GraphFlow) _inst()         {}
func (Before) _inst()            {}
func (After) _inst()             {}
func (ResolveRefs) _inst()       {}
func (ResolveAllRefs) _inst()    {}
func (Referred) _inst()          {}
func (ReduceList) _inst()        {}
func (IndexTuple) _inst()        {}
func (PopToInput) _inst()        {}
func (SwitchModelRef) _inst()    {}
func (IsPresent) _inst()         {}
func (DebugPrintStack) _inst()   {}
func (CurrentUser) _inst()       {}
func (GreaterFloat) _inst()      {}
func (GreaterInt) _inst()        {}
func (GreaterInt8) _inst()       {}
func (GreaterInt16) _inst()      {}
func (GreaterInt32) _inst()      {}
func (GreaterInt64) _inst()      {}
func (GreaterUint) _inst()       {}
func (GreaterUint8) _inst()      {}
func (GreaterUint16) _inst()     {}
func (GreaterUint32) _inst()     {}
func (GreaterUint64) _inst()     {}
func (LessFloat) _inst()         {}
func (LessInt) _inst()           {}
func (LessInt8) _inst()          {}
func (LessInt16) _inst()         {}
func (LessInt32) _inst()         {}
func (LessInt64) _inst()         {}
func (LessUint) _inst()          {}
func (LessUint8) _inst()         {}
func (LessUint16) _inst()        {}
func (LessUint32) _inst()        {}
func (LessUint64) _inst()        {}
func (AddFloat) _inst()          {}
func (AddInt) _inst()            {}
func (AddInt8) _inst()           {}
func (AddInt16) _inst()          {}
func (AddInt32) _inst()          {}
func (AddInt64) _inst()          {}
func (AddUint) _inst()           {}
func (AddUint8) _inst()          {}
func (AddUint16) _inst()         {}
func (AddUint32) _inst()         {}
func (AddUint64) _inst()         {}
func (SubtractFloat) _inst()     {}
func (SubtractInt) _inst()       {}
func (SubtractInt8) _inst()      {}
func (SubtractInt16) _inst()     {}
func (SubtractInt32) _inst()     {}
func (SubtractInt64) _inst()     {}
func (SubtractUint) _inst()      {}
func (SubtractUint8) _inst()     {}
func (SubtractUint16) _inst()    {}
func (SubtractUint32) _inst()    {}
func (SubtractUint64) _inst()    {}
func (MultiplyFloat) _inst()     {}
func (MultiplyInt) _inst()       {}
func (MultiplyInt8) _inst()      {}
func (MultiplyInt16) _inst()     {}
func (MultiplyInt32) _inst()     {}
func (MultiplyInt64) _inst()     {}
func (MultiplyUint) _inst()      {}
func (MultiplyUint8) _inst()     {}
func (MultiplyUint16) _inst()    {}
func (MultiplyUint32) _inst()    {}
func (MultiplyUint64) _inst()    {}
func (DivideFloat) _inst()       {}
func (DivideInt) _inst()         {}
func (DivideInt8) _inst()        {}
func (DivideInt16) _inst()       {}
func (DivideInt32) _inst()       {}
func (DivideInt64) _inst()       {}
func (DivideUint) _inst()        {}
func (DivideUint8) _inst()       {}
func (DivideUint16) _inst()      {}
func (DivideUint32) _inst()      {}
func (DivideUint64) _inst()      {}
func (If) _inst()                {}
func (AssertModelRef) _inst()    {}
func (ToFloat) _inst()           {}
func (ToInt) _inst()             {}
func (ToInt8) _inst()            {}
func (ToInt16) _inst()           {}
func (ToInt32) _inst()           {}
func (ToInt64) _inst()           {}
func (ToUint) _inst()            {}
func (ToUint8) _inst()           {}
func (ToUint16) _inst()          {}
func (ToUint32) _inst()          {}
func (ToUint64) _inst()          {}
func (MatchRegex) _inst()        {}
func (Slice) _inst()             {}
func (ExtractStrings) _inst()    {}
func (JoinStrings) _inst()       {}
func (RelocateRef) _inst()       {}
func (SwitchType) _inst()        {}
func (MapSet) _inst()            {}
func (IsCase) _inst()            {}
func (PresentOrConstant) _inst() {}
func (MemSort) _inst()           {}
func (ReverseList) _inst()       {}
func (StringToLower) _inst()     {}
func (ConcatLists) _inst()       {}
func (SearchAllRegex) _inst()    {}
func (SearchRegex) _inst()       {}
func (ToString) _inst()          {}
