// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"fmt"
	"kvm/val"
)

type ErrorList []Error

func (a ErrorList) OverMap(f func(Error) Error) ErrorList {
	for i, b := range a {
		a[i] = f(b)
	}
	return a
}

func (e ErrorList) Value() val.Union {
	l := make(val.List, len(e), len(e))
	for i, e := range e {
		l[i] = e.Value()
	}
	return val.Union{"errorList", l}
}
func (e ErrorList) Error() string {
	return e.String()
}
func (e ErrorList) String() string {
	out := ""
	for _, e := range e {
		out += e.String() + "\n\n"
	}
	return out
}
func (e ErrorList) Child() Error {
	return nil
}

type ObjectNotFoundError struct {
	Ref    val.Ref
	Child_ Error
}

func (e ObjectNotFoundError) Value() val.Union {
	return val.Union{"objectNotFoundError", e.Ref}
}
func (e ObjectNotFoundError) Error() string {
	return e.String()
}
func (e ObjectNotFoundError) String() string {
	out := "Object Not Found Error\n"
	out += "======================\n"
	out += "Reference\n"
	out += "---------\n"
	out += fmt.Sprintf("model: %s, id: %s\n\n", e.Ref[0], e.Ref[1])
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e ObjectNotFoundError) Child() Error {
	return e.Child_
}

type ModelNotFoundError struct {
	ObjectNotFoundError
}

func (e ModelNotFoundError) Value() val.Union {
	return val.Union{"modelNotFoundError", e.Ref}
}
func (e ModelNotFoundError) Error() string {
	return e.String()
}
func (e ModelNotFoundError) String() string {
	out := "Model Not Found Error\n"
	out += "=====================\n"
	out += "Reference\n"
	out += "---------\n"
	out += fmt.Sprintf("model: %s, id: %s\n\n", e.Ref[0], e.Ref[1])
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}

type PermissionDeniedError struct {
	Child_ Error
}

func (e PermissionDeniedError) Value() val.Union {
	return val.Union{"permissionDeniedError", val.Struct{}}
}
func (e PermissionDeniedError) Error() string {
	return e.String()
}
func (e PermissionDeniedError) String() string {
	out := "Permission Denied Error\n"
	out += "=======================\n\n"
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e PermissionDeniedError) Child() Error {
	return e.Child_
}

type CompilationError struct {
	Problem string
	Program val.Value
	Child_  Error
}

func (e CompilationError) Value() val.Union {
	if e.Program == nil {
		e.Program = val.String("(unknown)")
	}
	return val.Union{"compilationError", val.Struct{
		"problem": val.String(e.Problem),
		"program": e.Program,
	}}
}
func (e CompilationError) Error() string {
	return e.String()
}
func (e CompilationError) String() string {
	out := "Compilation Error\n"
	out += "=================\n"
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n\n"
	out += "Program\n"
	out += "-------\n"
	out += ProgramToHuman(e.Program, 0) + "\n\n"
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e CompilationError) Child() Error {
	return e.Child_
}

type ExecutionError struct {
	Problem string
	Child_  Error
}

func (e ExecutionError) Value() val.Union {
	return val.Union{"executionError", val.Struct{
		"problem": val.String(e.Problem),
	}}
}
func (e ExecutionError) Error() string {
	return e.String()
}
func (e ExecutionError) String() string {
	out := "Execution Error\n"
	out += "=================\n"
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n\n"
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e ExecutionError) Child() Error {
	return e.Child_
}

type DatabaseDoesNotExistError struct {
	Name string
}

func (e DatabaseDoesNotExistError) Value() val.Union {
	return val.Union{"databaseDoesNotExistError", val.String(e.Name)}
}
func (e DatabaseDoesNotExistError) Error() string {
	return e.String()
}
func (e DatabaseDoesNotExistError) String() string {
	out := "Database Does Not Exist Error\n"
	out += "=============================\n"
	out += "Name\n"
	out += "----\n"
	out += e.Name + "\n\n"
	return out
}
func (e DatabaseDoesNotExistError) Child() Error {
	return nil
}

type InternalError struct {
	Problem string
	Child_  Error
}

func (e InternalError) Value() val.Union {
	return val.Union{"internalError", val.String(e.Problem)}
}
func (e InternalError) Error() string {
	return e.String()
}
func (e InternalError) String() string {
	out := "Internal Error\n"
	out += "==============\n"
	out += "Please excuse us, karma slipped and fell on its face.\n\n"
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e InternalError) Child() Error {
	return e.Child_
}

type RequestError struct {
	Problem string
	Child_  Error
}

func (e RequestError) Value() val.Union {
	return val.Union{"requestError", val.String(e.Problem)}
}
func (e RequestError) Error() string {
	return e.String()
}
func (e RequestError) String() string {
	out := "Request Error\n"
	out += "=============\n"
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n\n"
	if e.Child_ != nil {
		out += e.Child_.String()
	}
	return out
}
func (e RequestError) Child() Error {
	return e.Child_
}
