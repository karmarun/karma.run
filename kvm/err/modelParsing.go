// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"github.com/karmarun/karma.run/kvm/val"
)

type ModelParsingError struct {
	Problem string
	Input   val.Value
	Path    ErrorPath
}

func (e ModelParsingError) ErrorPath() ErrorPath {
	return e.Path
}
func (e ModelParsingError) AppendPath(a ErrorPathElement, b ...ErrorPathElement) PathedError {
	e.Path = append(append(e.Path, a), b...)
	return e
}

func (e ModelParsingError) Value() val.Union {
	return val.Union{"modelParsingError", val.Struct{
		"problem": val.String(e.Problem),
		"input":   e.Input,
		"path":    e.Path.Value(),
	}}
}
func (e ModelParsingError) Error() string {
	return e.String()
}
func (e ModelParsingError) String() string {
	out := "Model Parsing Error\n"
	out += "===================\n"
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n\n"
	out += "Location\n"
	out += "--------\n"
	out += e.Path.String() + "\n\n"
	out += "Input\n"
	out += "-----\n"
	out += ValueToHuman(e.Input) + "\n"
	return out
}
func (e ModelParsingError) Child() Error {
	return nil
}
