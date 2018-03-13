// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package val

import (
	"fmt"
	"strings"
)

type Type uint64

const (
	TypeTuple Type = 1 << iota
	TypeList
	TypeUnion
	TypeStruct
	TypeMap
	TypeFloat
	TypeBool
	TypeString
	TypeRef
	TypeDateTime
	TypeNull
	TypeSymbol
	TypeSet
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	lastType // internal marker
)

const AnyType = TypeTuple |
	TypeList |
	TypeUnion |
	TypeStruct |
	TypeMap |
	TypeFloat |
	TypeBool |
	TypeString |
	TypeRef |
	TypeDateTime |
	TypeNull |
	TypeSymbol |
	TypeSet |
	TypeInt8 |
	TypeInt16 |
	TypeInt32 |
	TypeInt64 |
	TypeUint8 |
	TypeUint16 |
	TypeUint32 |
	TypeUint64

func (t Type) String() string {
	if t == 0 {
		return "unknown"
	}
	buf := make([]string, 0, 64)
	for i := Type(0); i < lastType; i++ {
		q := (Type(1) << i) & t
		if q != 0 {
			buf = append(buf, typeToString(q))
		}
	}
	return strings.Join(buf, "|")
}

func typeToString(t Type) string {
	switch t {
	case TypeTuple:
		return "tuple"
	case TypeList:
		return "list"
	case TypeUnion:
		return "union"
	case TypeStruct:
		return "struct"
	case TypeMap:
		return "map"
	case TypeFloat:
		return "float"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeRef:
		return "ref"
	case TypeDateTime:
		return "dateTime"
	case TypeNull:
		return "null"
	case TypeSymbol:
		return "symbol"
	case TypeSet:
		return "set"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"
	}
	panic(fmt.Sprintf("unhandled Type: %b", uint64(t)))
}

func (m Meta) Type() Type {
	return m.Value.Type()
}

func (Tuple) Type() Type {
	return TypeTuple
}

func (List) Type() Type {
	return TypeList
}

func (Union) Type() Type {
	return TypeUnion
}

func (Struct) Type() Type {
	return TypeStruct
}

func (Map) Type() Type {
	return TypeMap
}

func (Float) Type() Type {
	return TypeFloat
}

func (Bool) Type() Type {
	return TypeBool
}

func (String) Type() Type {
	return TypeString
}

func (Ref) Type() Type {
	return TypeRef
}

func (DateTime) Type() Type {
	return TypeDateTime
}

func (null) Type() Type {
	return TypeNull
}

func (Symbol) Type() Type {
	return TypeSymbol
}

func (Set) Type() Type {
	return TypeSet
}

func (Int8) Type() Type {
	return TypeInt8
}

func (Int16) Type() Type {
	return TypeInt16
}

func (Int32) Type() Type {
	return TypeInt32
}

func (Int64) Type() Type {
	return TypeInt64
}

func (Uint8) Type() Type {
	return TypeUint8
}

func (Uint16) Type() Type {
	return TypeUint16
}

func (Uint32) Type() Type {
	return TypeUint32
}

func (Uint64) Type() Type {
	return TypeUint64
}
