// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package binary

import (
	"fmt"
	"karma.run/codec"
	"karma.run/kvm/err"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"math"
	"time"
)

type Type byte

const (
	TypeTuple    Type = 0
	TypeList     Type = 1
	TypeUnion    Type = 2
	TypeStruct   Type = 3
	TypeMap      Type = 4
	TypeFloat    Type = 5
	TypeBool     Type = 6
	TypeString   Type = 7
	TypeRef      Type = 8
	TypeDateTime Type = 9
	TypeNull     Type = 10
	TypeSymbol   Type = 11
	TypeSet      Type = 12
	TypeInt8     Type = 13
	TypeInt16    Type = 14
	TypeInt32    Type = 15
	TypeInt64    Type = 16
	TypeUint8    Type = 17
	TypeUint16   Type = 18
	TypeUint32   Type = 19
	TypeUint64   Type = 20
)

func (t Type) String() string {
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
	return "unknown"
}

var sharedInstance = JstrCodec{}

func init() {
	codec.Register("binary", func() codec.Interface { return sharedInstance })
}

type JstrCodec struct{}

func (dec JstrCodec) Decode(data []byte, model mdl.Model) (val.Value, err.Error) {
	return Decode(data, model)
}

func (dec JstrCodec) Encode(v val.Value) []byte {
	return Encode(v)
}

func Encode(v val.Value) []byte {
	buf := make([]byte, 0, 1024*4)
	buf = encode(v, buf)
	return buf
}

func encode(v val.Value, buf []byte) []byte {
	switch v := v.(type) {

	case val.Null:
		return append(buf, byte(TypeNull))

	case val.Meta:
		return encode(v.Value, buf)

	case val.Tuple:
		buf = append(buf, byte(TypeTuple))
		buf = writeLength(len(v), buf)
		for _, w := range v {
			buf = encode(w, buf)
		}
		return buf

	case val.Set:
		buf = append(buf, byte(TypeSet))
		buf = writeLength(len(v), buf)
		for _, w := range v {
			buf = encode(w, buf)
		}
		return buf

	case val.List:
		buf = append(buf, byte(TypeList))
		buf = writeLength(len(v), buf)
		for _, w := range v {
			buf = encode(w, buf)
		}
		return buf

	case val.Union:
		buf = append(buf, byte(TypeUnion))
		buf = writeString(v.Case, buf)
		return encode(v.Value, buf)

	case val.Struct:
		buf = append(buf, byte(TypeStruct))
		buf = writeLength(len(v), buf)
		for k, w := range v {
			buf = writeString(k, buf)
			buf = encode(w, buf)
		}
		return buf

	case val.Map:
		buf = append(buf, byte(TypeMap))
		buf = writeLength(len(v), buf)
		for k, w := range v {
			buf = writeString(k, buf)
			buf = encode(w, buf)
		}
		return buf

	case val.Float:
		buf = append(buf, byte(TypeFloat))
		return writeUint64(math.Float64bits(float64(v)), buf)

	case val.Bool:
		buf = append(buf, byte(TypeBool))
		if v {
			return append(buf, 't')
		}
		return append(buf, 'f')

	case val.Symbol:
		buf = append(buf, byte(TypeSymbol))
		return writeString(string(v), buf)

	case val.String:
		buf = append(buf, byte(TypeString))
		return writeString(string(v), buf)

	case val.Ref:
		buf = append(buf, byte(TypeRef))
		buf = writeString(v[0], buf)
		return writeString(v[1], buf)

	case val.DateTime:
		buf = append(buf, byte(TypeDateTime))
		return writeString(v.Time.Format(time.RFC3339), buf)

	case val.Int8:
		buf = append(buf, byte(TypeInt8))
		return writeInt8(int8(v), buf)

	case val.Int16:
		buf = append(buf, byte(TypeInt16))
		return writeInt16(int16(v), buf)

	case val.Int32:
		buf = append(buf, byte(TypeInt32))
		return writeInt32(int32(v), buf)

	case val.Int64:
		buf = append(buf, byte(TypeInt64))
		return writeInt64(int64(v), buf)

	case val.Uint8:
		buf = append(buf, byte(TypeUint8))
		return writeUint8(uint8(v), buf)

	case val.Uint16:
		buf = append(buf, byte(TypeUint16))
		return writeUint16(uint16(v), buf)

	case val.Uint32:
		buf = append(buf, byte(TypeUint32))
		return writeUint32(uint32(v), buf)

	case val.Uint64:
		buf = append(buf, byte(TypeUint64))
		return writeUint64(uint64(v), buf)
	}

	panic(fmt.Sprintf(`unhandled type: %T`, v))
}

func Decode(data []byte, _ mdl.Model) (val.Value, err.Error) {
	v, d, e := decode(data)
	if e != nil {
		return nil, err.CodecError{"binary", len(data) - len(d), e}
	}
	return v, nil
}

func decode(data []byte) (val.Value, []byte, err.Error) {

	r, data, e := readBytes(1, data)
	if e != nil {
		return nil, data, e
	}

	switch t := Type(r[0]); t {
	case TypeTuple:
		l, data, e := readLength(data)
		if e != nil {
			return nil, data, e
		}
		v := make(val.Tuple, l, l)
		for i := 0; i < l; i++ {
			w, d, e := decode(data)
			if e != nil {
				return nil, d, e
			}
			v[i], data = w, d
		}
		return v, data, nil

	case TypeList:
		l, data, e := readLength(data)
		if e != nil {
			return nil, data, e
		}
		v := make(val.List, l, l)
		for i := 0; i < l; i++ {
			w, d, e := decode(data)
			if e != nil {
				return nil, d, e
			}
			v[i], data = w, d
		}
		return v, data, nil

	case TypeUnion:
		caze, data, e := readString(data)
		if e != nil {
			return nil, data, e
		}
		value, data, e := decode(data)
		if e != nil {
			return nil, data, e
		}
		return val.Union{caze, value}, data, nil

	case TypeStruct:
		l, data, e := readLength(data)
		if e != nil {
			return nil, data, e
		}
		v := make(val.Struct, l)
		for i := 0; i < l; i++ {
			field, d, e := readString(data)
			if e != nil {
				return nil, d, e
			}
			data = d
			value, d, e := decode(data)
			if e != nil {
				return nil, d, e
			}
			data = d
			v[field] = value
		}
		return v, data, nil

	case TypeMap:
		l, data, e := readLength(data)
		if e != nil {
			return nil, data, e
		}
		v := make(val.Map, l)
		for i := 0; i < l; i++ {
			field, d, e := readString(data)
			if e != nil {
				return nil, d, e
			}
			data = d
			value, d, e := decode(data)
			if e != nil {
				return nil, d, e
			}
			data = d
			v[field] = value
		}
		return v, data, nil

	case TypeInt64:
		n, data, e := readInt64(data)
		if e != nil {
			return nil, data, e
		}
		return val.Int64(n), data, nil

	case TypeInt32:
		n, data, e := readInt32(data)
		if e != nil {
			return nil, data, e
		}
		return val.Int32(n), data, nil

	case TypeInt16:
		n, data, e := readInt16(data)
		if e != nil {
			return nil, data, e
		}
		return val.Int16(n), data, nil

	case TypeInt8:
		n, data, e := readInt8(data)
		if e != nil {
			return nil, data, e
		}
		return val.Int8(n), data, nil

	case TypeBool:
		r, data, e := readBytes(1, data)
		if e != nil {
			return nil, data, e
		}
		if r[0] == 't' {
			return val.Bool(true), data, nil
		}
		if r[0] == 'f' {
			return val.Bool(false), data, nil
		}
		return nil, data, err.InputParsingError{fmt.Sprintf(`expected 't' or 'f', got: %s`, string(r[0])), data}

	case TypeString:
		s, data, e := readString(data)
		if e != nil {
			return nil, data, e
		}
		return val.String(s), data, nil

	case TypeRef:
		m, data, e := readString(data)
		if e != nil {
			return nil, data, e
		}
		i, data, e := readString(data)
		if e != nil {
			return nil, data, e
		}
		return (val.Ref{m, i}), data, nil // parenthesis resolve ambiguity

	case TypeDateTime:
		s, data, e := readString(data)
		if e != nil {
			return nil, data, e
		}
		t, te := time.Parse(time.RFC3339, s)
		if te != nil {
			return nil, data, err.InputParsingError{fmt.Sprintf(`invalid ISO dateTime string: %s`, s), data}
		}
		return val.DateTime{t}, data, nil

	case TypeNull:
		return val.Null{}, data, nil

	case TypeSymbol:
		s, data, e := readString(data)
		if e != nil {
			return nil, data, e
		}
		return val.Symbol(s), data, e

	case TypeSet:
		l, data, e := readLength(data)
		if e != nil {
			return nil, data, e
		}
		v := make(val.Set, l)
		for i := 0; i < l; i++ {
			w, d, e := decode(data)
			if e != nil {
				return nil, d, e
			}
			v[val.Hash(w, nil).Sum64()], data = w, d
		}
		return v, data, nil

	case TypeUint64:
		n, data, e := readUint64(data)
		if e != nil {
			return nil, data, e
		}
		return val.Uint64(n), data, e

	case TypeUint32:
		n, data, e := readUint32(data)
		if e != nil {
			return nil, data, e
		}
		return val.Uint32(n), data, e

	case TypeUint16:
		n, data, e := readUint16(data)
		if e != nil {
			return nil, data, e
		}
		return val.Uint16(n), data, e

	case TypeUint8:
		n, data, e := readUint8(data)
		if e != nil {
			return nil, data, e
		}
		return val.Uint8(n), data, e

	case TypeFloat:
		n, data, e := readUint64(data)
		if e != nil {
			return nil, data, e
		}
		return val.Float(math.Float64frombits(n)), data, nil
	}
	return nil, data, err.InputParsingError{fmt.Sprintf(`invalid type specifier: %d`, r), data}
}

func readBytes(n int, data []byte) ([]byte, []byte, err.Error) {
	if len(data) < n {
		return nil, data, err.InputParsingError{`unexpected EOF`, data}
	}
	return data[:n], data[n:], nil
}

func readLength(data []byte) (int, []byte, err.Error) {
	r, data, e := readUint32(data)
	if e != nil {
		return 0, data, e
	}
	l := int(r)
	if e := validateLength(l, data); e != nil {
		return 0, data, e
	}
	return l, data, nil
}

func readString(data []byte) (string, []byte, err.Error) {
	l, data, e := readLength(data)
	if e != nil {
		return "", data, e
	}
	if e := validateLength(l, data); e != nil {
		return "", data, e
	}
	return string(data[:l]), data[l:], nil
}

func writeString(s string, buf []byte) []byte {
	bs := []byte(s)
	return append(writeLength(len(bs), buf), bs...)
}

func writeLength(l int, buf []byte) []byte {
	return writeUint32(uint32(l), buf)
}

func validateLength(l int, data []byte) err.Error {
	if l < 0 {
		return err.InputParsingError{fmt.Sprintf(`negative length: %d`, l), data}
	}
	if l > len(data) {
		return err.InputParsingError{fmt.Sprintf(`length exceeds input bounds: %d`, l), data}
	}
	return nil
}

func writeInt64(i int64, buf []byte) []byte {
	return writeUint64(uint64(i), buf)
}

func writeInt32(i int32, buf []byte) []byte {
	return writeUint32(uint32(i), buf)
}

func writeInt16(i int16, buf []byte) []byte {
	return writeUint16(uint16(i), buf)
}

func writeInt8(i int8, buf []byte) []byte {
	return writeUint8(uint8(i), buf)
}

func writeUint64(u uint64, buf []byte) []byte {
	buf = append(buf, byte(u>>56))
	buf = append(buf, byte(u>>48))
	buf = append(buf, byte(u>>40))
	buf = append(buf, byte(u>>32))
	buf = append(buf, byte(u>>24))
	buf = append(buf, byte(u>>16))
	buf = append(buf, byte(u>>8))
	buf = append(buf, byte(u))
	return buf
}

func writeUint32(u uint32, buf []byte) []byte {
	buf = append(buf, byte(u>>24))
	buf = append(buf, byte(u>>16))
	buf = append(buf, byte(u>>8))
	buf = append(buf, byte(u))
	return buf
}

func writeUint16(u uint16, buf []byte) []byte {
	buf = append(buf, byte(u>>8))
	buf = append(buf, byte(u))
	return buf
}

func writeUint8(u uint8, buf []byte) []byte {
	buf = append(buf, byte(u))
	return buf
}

func readInt64(data []byte) (int64, []byte, err.Error) {
	u, data, e := readUint64(data)
	if e != nil {
		return 0, data, e
	}
	return int64(u), data, nil
}

func readInt32(data []byte) (int32, []byte, err.Error) {
	u, data, e := readUint32(data)
	if e != nil {
		return 0, data, e
	}
	return int32(u), data, nil
}

func readInt16(data []byte) (int16, []byte, err.Error) {
	u, data, e := readUint16(data)
	if e != nil {
		return 0, data, e
	}
	return int16(u), data, nil
}

func readInt8(data []byte) (int8, []byte, err.Error) {
	u, data, e := readUint8(data)
	if e != nil {
		return 0, data, e
	}
	return int8(u), data, nil
}

func readUint64(data []byte) (uint64, []byte, err.Error) {
	bs, data, e := readBytes(8, data)
	if e != nil {
		return 0, data, e
	}
	return uint64(bs[7])<<56 |
		uint64(bs[6])<<48 |
		uint64(bs[5])<<40 |
		uint64(bs[4])<<32 |
		uint64(bs[3])<<24 |
		uint64(bs[2])<<16 |
		uint64(bs[1])<<8 |
		uint64(bs[0]), data, e
}

func readUint32(data []byte) (uint32, []byte, err.Error) {
	bs, data, e := readBytes(4, data)
	if e != nil {
		return 0, data, e
	}
	return uint32(bs[3])<<24 |
		uint32(bs[2])<<16 |
		uint32(bs[1])<<8 |
		uint32(bs[0]), data, e
}

func readUint16(data []byte) (uint16, []byte, err.Error) {
	bs, data, e := readBytes(2, data)
	if e != nil {
		return 0, data, e
	}
	return uint16(bs[1])<<8 |
		uint16(bs[0]), data, e
}

func readUint8(data []byte) (uint8, []byte, err.Error) {
	bs, data, e := readBytes(1, data)
	if e != nil {
		return 0, data, e
	}
	return uint8(bs[0]), data, e
}
