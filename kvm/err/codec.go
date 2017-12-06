// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"fmt"
	"github.com/karmarun/karma.run/kvm/val"
)

type OffsetError interface {
	Error
	Offset() int
	SetOffset(int) OffsetError
}

type CodecError struct {
	Name    string // the name of the codec
	Offset_ int
	Child_  Error
}

func (e CodecError) SetOffset(o int) OffsetError {
	e.Offset_ = o
	return e
}
func (e CodecError) Offset() int {
	return e.Offset_
}
func (e CodecError) Value() val.Union {
	return val.Union{"codecError", val.Struct{
		"name":   val.String(e.Name),
		"offset": val.Int64(e.Offset_),
	}}
}
func (e CodecError) Error() string {
	return e.String()
}
func (e CodecError) String() string {
	out := "Codec Error\n"
	out += "===========\n"
	out += "Codec\n"
	out += "-----\n"
	out += e.Name + "\n\n"
	out += "Offset\n"
	out += "------\n"
	out += fmt.Sprintf("%d\n\n", e.Offset_)
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e CodecError) Child() Error {
	return e.Child_
}

// general input parsing errors
type InputParsingError struct {
	Problem string
	Input   []byte
}

func (e InputParsingError) Value() val.Union {
	return val.Union{"inputParsingError", val.Struct{
		"problem": val.String(e.Problem),
		"input":   val.String(e.Input),
	}}
}
func (e InputParsingError) Error() string {
	return e.String()
}
func (e InputParsingError) String() string {
	out := "Input Parsing Error\n"
	out += "===================\n"
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n\n"
	out += "Input\n"
	out += "-------\n"
	out += string(e.Input) + "\n\n"
	return out
}
func (e InputParsingError) Child() Error {
	return nil
}
