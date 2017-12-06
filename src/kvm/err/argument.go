// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"fmt"
	"kvm/val"
)

type ArgumentError struct { // implements PathedError
	Path   ErrorPath
	Child_ Error
}

func LiftArgumentError(e Error) ArgumentError {
	if ae, ok := e.(ArgumentError); ok {
		return ae
	}
	return ArgumentError{nil, e}
}

func (e ArgumentError) ErrorPath() ErrorPath {
	return e.Path
}

func (e ArgumentError) AppendPath(p ErrorPathElement, ps ...ErrorPathElement) PathedError {
	e.Path = append(append(e.Path, p), ps...)
	return e
}

func (e ArgumentError) Value() val.Union {
	return val.Union{"argumentError", e.Path.Value()}
}

func (e ArgumentError) Error() string {
	return e.String()
}

func (e ArgumentError) String() string {
	out := "Argument Error\n"
	out += "==============\n"
	out += "Location\n"
	out += "--------\n"
	out += e.Path.String() + "\n\n"
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}

func (e ArgumentError) Child() Error {
	return e.Child_
}

type FuncArgPathElement struct {
	Function string
	Argument ErrorPathElement
}

func NewFuncArgPathElement(function string, argument interface{}) FuncArgPathElement {
	arg := (ErrorPathElement)(nil)
	if i, ok := argument.(int); ok {
		arg = NumericArgPathElement(i)
	}
	if s, ok := argument.(string); ok {
		arg = NamedArgPathElement(s)
	}
	if a, ok := argument.(ErrorPathElement); ok {
		arg = a
	}
	if arg == nil {
		panic(fmt.Sprintf("NewFuncArgPathElement: unexpected %T", argument))
	}
	return FuncArgPathElement{function, arg}
}

func (p FuncArgPathElement) String() string {
	return fmt.Sprintf(`%s of function "%s"`, p.Argument.String(), p.Function)
}

func (p FuncArgPathElement) Value() val.Union {
	return val.Union{"functionArgument", val.Struct{
		"function": val.String(p.Function),
		"argument": p.Argument.Value(),
	}}
}

type NumericArgPathElement int

func (p NumericArgPathElement) String() string {
	switch p {
	case 1:
		return "first argument"
	case 2:
		return "second argument"
	case 3:
		return "third argument"
	case 4:
		return "fourth argument"
	case 5:
		return "fifth argument"
	}
	return fmt.Sprintf("argument number %d", int(p))
}

func (p NumericArgPathElement) Value() val.Union {
	return val.Union{"argumentNumber", val.Int64(p)}
}

type NamedArgPathElement string

func (p NamedArgPathElement) String() string {
	return fmt.Sprintf(`argument "%s"`, string(p))
}

func (p NamedArgPathElement) Value() val.Union {
	return val.Union{"argumentNamed", val.String(p)}
}
