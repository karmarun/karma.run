// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"kvm/mdl"
	"strings"
)

// TODO: make recursion secure
func ModelToHuman(m mdl.Model) string {
	return _ModelToHuman(m, nil)
}

func _ModelToHuman(m mdl.Model, r map[*mdl.Recursion]struct{}) string {
	if r == nil {
		r = make(map[*mdl.Recursion]struct{})
	}
	switch m := m.(type) {
	case *mdl.Recursion:
		if _, ok := r[m]; ok {
			return `recursion`
		}
		r[m] = struct{}{}
		s := _ModelToHuman(m.Model, r)
		delete(r, m)
		return s

	case mdl.Unique:
		return fmt.Sprintf(`unique{%s}`, m.Model)
	case mdl.Annotation:
		return fmt.Sprintf(`annotation{%s}`, m.Model)
	case mdl.Or:
		l, r := _ModelToHuman(m[0], r), _ModelToHuman(m[1], r)
		return l + " | " + r
	case mdl.Set:
		return "set of " + _ModelToHuman(m.Elements, r)
	case mdl.List:
		return "list of " + _ModelToHuman(m.Elements, r)
	case mdl.Map:
		return "map of " + _ModelToHuman(m.Elements, r)
	case mdl.Tuple:
		ss := make([]string, 0, len(m))
		for _, w := range m {
			ss = append(ss, _ModelToHuman(w, r))
		}
		return fmt.Sprintf(`tuple(%s)`, strings.Join(ss, ", "))
	case mdl.Struct:
		ks := m.Keys()
		args := ""
		for i, l := 0, len(ks); i < l; i++ {
			k := ks[i]
			if i > 0 {
				args += ", "
			}
			args += k + ": " + _ModelToHuman(m[k], r)
		}
		return fmt.Sprintf("struct{%s}", args)
	case mdl.Union:
		ks := m.Cases()
		args := ""
		for i, l := 0, len(ks); i < l; i++ {
			k := ks[i]
			if i > 0 {
				args += ", "
			}
			args += k + ": " + _ModelToHuman(m[k], r)
		}
		return fmt.Sprintf("union{%s}", args)
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
