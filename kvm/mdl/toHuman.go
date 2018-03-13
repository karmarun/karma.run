// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"fmt"
	"strings"
)

func ModelToHuman(m Model) string {
	return modelToHuman(m, 1, nil)
}

func modelToHuman(m Model, indent int, r map[*Recursion]struct{}) string {
	if r == nil {
		r = make(map[*Recursion]struct{})
	}
	m = m.Unwrap()
	switch m := m.(type) {

	case Optional:
		return "optional " + modelToHuman(m.Model, indent, r)

	case *Recursion:
		if _, ok := r[m]; ok {
			return fmt.Sprintf(`recurse(%s)`, m.Label)
		}
		r[m] = struct{}{}
		s := modelToHuman(m.Model, indent, r)
		delete(r, m)
		return s

	case Unique:
		return fmt.Sprintf(`unique %s`, modelToHuman(m.Model, indent, r))
	case Annotation:
		return fmt.Sprintf(`annotation(%s) of %s`, m.Value, modelToHuman(m.Model, indent, r))
	case Set:
		return "set of " + modelToHuman(m.Elements, indent, r)
	case List:
		return "list of " + modelToHuman(m.Elements, indent, r)
	case Map:
		return "map of " + modelToHuman(m.Elements, indent, r)
	case Tuple:
		ss := make([]string, 0, len(m))
		for _, w := range m {
			ss = append(ss, modelToHuman(w, indent, r))
		}
		return fmt.Sprintf(`tuple(%s)`, strings.Join(ss, ", "))
	case Struct:
		ks := m.Keys()
		if len(ks) == 0 {
			return "struct{}"
		}
		args := "\n"
		for i, l := 0, len(ks); i < l; i++ {
			k := ks[i]
			if i > 0 {
				args += ",\n"
			}
			args += strings.Repeat(" ", indent*2) + k + ": " + modelToHuman(m.Field(k), indent+1, r)
		}
		return fmt.Sprintf("struct {%s\n%s}", args, strings.Repeat(" ", (indent-1)*2))
	case Union:
		ks := m.Cases()
		if len(ks) == 0 {
			return "union{}"
		}
		args := "\n"
		for i, l := 0, len(ks); i < l; i++ {
			k := ks[i]
			if i > 0 {
				args += ",\n"
			}
			args += strings.Repeat(" ", indent*2) + k + ": " + modelToHuman(m.Case(k), indent+1, r)
		}
		return fmt.Sprintf("union {%s\n%s}", args, strings.Repeat(" ", (indent-1)*2))
	case Enum:
		ss := make([]string, 0, len(m))
		for k, _ := range m {
			ss = append(ss, k)
		}
		return fmt.Sprintf(`enum(%s)`, strings.Join(ss, ", "))
	case Ref:
		if m.Model == "" {
			return "ref"
		}
		return "ref to " + m.Model
	case Any:
		return "any"
	case Null:
		return "null"
	case String:
		return "string"
	case Float:
		return "float"
	case Bool:
		return "bool"
	case DateTime:
		return "dateTime"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	}
	panic(fmt.Sprintf("unhandled model type: %T", m))
}
