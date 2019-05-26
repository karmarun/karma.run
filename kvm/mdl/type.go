// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"karma.run/kvm/val"
)

func (m *Recursion) ValueType() val.Type {
	return m.Model.ValueType()
}

func (m Annotation) ValueType() val.Type {
	return m.Model.ValueType()
}

func (m Optional) ValueType() val.Type {
	return m.Model.ValueType()
}

func (Any) ValueType() val.Type {
	return val.AnyType
}

func (Tuple) ValueType() val.Type {
	return val.TypeTuple
}

func (List) ValueType() val.Type {
	return val.TypeList
}

func (Union) ValueType() val.Type {
	return val.TypeUnion
}

func (Struct) ValueType() val.Type {
	return val.TypeStruct
}

func (Map) ValueType() val.Type {
	return val.TypeMap
}

func (Float) ValueType() val.Type {
	return val.TypeFloat
}

func (Bool) ValueType() val.Type {
	return val.TypeBool
}

func (String) ValueType() val.Type {
	return val.TypeString
}

func (Ref) ValueType() val.Type {
	return val.TypeRef
}

func (DateTime) ValueType() val.Type {
	return val.TypeDateTime
}

func (Null) ValueType() val.Type {
	return val.TypeNull
}

func (Enum) ValueType() val.Type {
	return val.TypeSymbol
}

func (Set) ValueType() val.Type {
	return val.TypeSet
}

func (Int8) ValueType() val.Type {
	return val.TypeInt8
}

func (Int16) ValueType() val.Type {
	return val.TypeInt16
}

func (Int32) ValueType() val.Type {
	return val.TypeInt32
}

func (Int64) ValueType() val.Type {
	return val.TypeInt64
}

func (Uint8) ValueType() val.Type {
	return val.TypeUint8
}

func (Uint16) ValueType() val.Type {
	return val.TypeUint16
}

func (Uint32) ValueType() val.Type {
	return val.TypeUint32
}

func (Uint64) ValueType() val.Type {
	return val.TypeUint64
}
