// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"karma.run/kvm/mdl"
	"strings"
)

func ModelToHuman(m mdl.Model) string {
	return modelToHuman(m, 1, nil)
}

func modelToHuman(m mdl.Model, indent int, r map[*mdl.Recursion]struct{}) string {
	if r == nil {
		r = make(map[*mdl.Recursion]struct{})
	}
	m = m.Unwrap()
	switch m := m.(type) {
	case *mdl.Recursion:
		if _, ok := r[m]; ok {
			return `recursion`
		}
		r[m] = struct{}{}
		s := modelToHuman(m.Model, indent, r)
		delete(r, m)
		return s

	case mdl.Unique:
		return fmt.Sprintf(`unique %s`, modelToHuman(m.Model, indent, r))
	case mdl.Annotation:
		return fmt.Sprintf(`annotation(%s) of %s`, m.Value, modelToHuman(m.Model, indent, r))
	case mdl.Set:
		return "set of " + modelToHuman(m.Elements, indent, r)
	case mdl.List:
		return "list of " + modelToHuman(m.Elements, indent, r)
	case mdl.Map:
		return "map of " + modelToHuman(m.Elements, indent, r)
	case mdl.Tuple:
		ss := make([]string, 0, len(m))
		for _, w := range m {
			ss = append(ss, modelToHuman(w, indent, r))
		}
		return fmt.Sprintf(`tuple(%s)`, strings.Join(ss, ", "))
	case mdl.Struct:
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
		return fmt.Sprintf("struct{%s\n%s}", args, strings.Repeat(" ", (indent-1)*2))
	case mdl.Union:
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
		return fmt.Sprintf("union{%s\n%s}", args, strings.Repeat(" ", (indent-1)*2))
	case mdl.Enum:
		ss := make([]string, 0, len(m))
		for k, _ := range m {
			ss = append(ss, k)
		}
		return fmt.Sprintf(`enum(%s)`, strings.Join(ss, ", "))
	case mdl.Ref:
		return "ref to " + m.Model
	case mdl.Any:
		return "any"
	case mdl.Null:
		return "null"
	case mdl.String:
		return "string"
	case mdl.Float:
		return "float"
	case mdl.Bool:
		return "bool"
	case mdl.DateTime:
		return "dateTime"
	case mdl.Int8:
		return "int8"
	case mdl.Int16:
		return "int16"
	case mdl.Int32:
		return "int32"
	case mdl.Int64:
		return "int64"
	case mdl.Uint8:
		return "uint8"
	case mdl.Uint16:
		return "uint16"
	case mdl.Uint32:
		return "uint32"
	case mdl.Uint64:
		return "uint64"
	}
	panic(fmt.Sprintf("unhandled model type: %T", m))
}
