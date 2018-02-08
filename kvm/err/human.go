// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"fmt"
	"karma.run/kvm/val"
	"strings"
	"time"
)

type HumanReadableError struct {
	Error_ Error
}

func (e HumanReadableError) Value() val.Union {
	return val.Union{"humanReadableError", val.StructFromMap(map[string]val.Value{
		"human":   val.String(e.Error_.String()),
		"machine": e.Error_.Value(),
	})}
}
func (e HumanReadableError) Error() string {
	return e.String()
}
func (e HumanReadableError) String() string {
	out := "Human Readable Error\n"
	out += "====================\n"
	out += e.Error_.String() + "\n"
	return out
}
func (e HumanReadableError) Child() Error {
	return nil
}

const indentUnit = "  "

func ProgramToHuman(v val.Value, indent int) string {
	if v == nil {
		return "(unknown)"
	}
	u, ok := v.(val.Union)
	if !ok {
		return ValueToHuman(v)
	}
	indentation := strings.Repeat(indentUnit, indent)
	switch u.Case {
	case "struct":
		subIndentation := strings.Repeat(indentUnit, indent+1)
		arg := u.Value.(val.Map)
		if arg.Len() == 0 {
			return "struct{}"
		}
		args := "\n"
		keys := arg.Keys()
		for i, l := 0, len(keys); i < l; i++ {
			k := keys[i]
			args += subIndentation + k + ": " + ProgramToHuman(arg.Key(k), indent+1) + ",\n"
		}
		return fmt.Sprintf("struct {%s%s}", args, indentation)
	case "list":
		subIndentation := strings.Repeat(indentUnit, indent+1)
		arg := u.Value.(val.List)
		if len(arg) == 0 {
			return "list[]"
		}
		args := "\n"
		for i, l := 0, len(arg); i < l; i++ {
			args += subIndentation + ProgramToHuman(arg[i], indent+1) + ",\n"
		}
		return fmt.Sprintf("list [%s%s]", args, indentation)

	case "set":
		subIndentation := strings.Repeat(indentUnit, indent+1)
		arg := u.Value.(val.Set)
		if len(arg) == 0 {
			return "set{}"
		}
		args := "\n"
		for _, sub := range arg {
			args += subIndentation + ProgramToHuman(sub, indent+1) + ",\n"
		}
		return fmt.Sprintf("set {%s%s}", args, indentation)

	case "map":
		subIndentation := strings.Repeat(indentUnit, indent+1)
		arg := u.Value.(val.Map)
		if arg.Len() == 0 {
			return "map{}"
		}
		args := "\n"
		keys := arg.Keys()
		for i, l := 0, len(keys); i < l; i++ {
			k := keys[i]
			args += subIndentation + fmt.Sprintf(`"%s" => `, k) + ProgramToHuman(arg.Key(k), indent+1) + ",\n"
		}
		return fmt.Sprintf("map {%s%s}", args, indentation)

	case "union":
		arg := u.Value.(val.Tuple)
		caze := string(arg[0].(val.Union).Value.(val.String))
		return fmt.Sprintf(`union(%s: %s)`, caze, ProgramToHuman(arg[1], indent))
	}
	out := u.Case
	a := u.Value
	if s, ok := a.(val.Struct); ok {
		if s.Len() == 0 {
			return out + `()`
		}
		as := make([]string, 0, s.Len())
		s.ForEach(func(k string, a val.Value) bool {
			as = append(as, fmt.Sprintf("\n%s%s = %s", indentation+indentUnit, k, ProgramToHuman(a, indent+2)))
			return true
		})
		return out + fmt.Sprintf("(%s\n%s)", strings.Join(as, ", "), indentation)
	}
	if l, ok := a.(val.List); ok {
		if len(l) == 0 {
			return out + `()`
		}
		as := make([]string, 0, len(l))
		for _, a := range l {
			as = append(as, ProgramToHuman(a, indent+1))
		}
		return out + fmt.Sprintf("(%s)", strings.Join(as, ", "))
	}
	if t, ok := a.(val.Tuple); ok {
		if len(t) == 0 {
			return out + `()`
		}
		as := make([]string, 0, len(t))
		for _, a := range t {
			as = append(as, ProgramToHuman(a, indent+1))
		}
		return out + fmt.Sprintf("(%s)", strings.Join(as, ", "))
	}
	return out + fmt.Sprintf(`(%s)`, ValueToHuman(a))

}
func ValueToHuman(v val.Value) string {
	if v == val.Null {
		return `null`
	}
	switch v := v.(type) {
	case val.Meta:
		return ValueToHuman(v.Value)
	case val.Tuple:
		return "..." // TODO
	case val.Union:
		return "..." // TODO
	case val.Struct:
		return "..." // TODO
	case val.Map:
		return "..." // TODO
	case val.List:
		return "..." // TODO
	case val.Set:
		return "..." // TODO
	case val.Ref:
		return fmt.Sprintf(`%s/%s`, v[0], v[1])
	case val.Raw:
		return `<raw bytes>`
	case val.Symbol:
		return fmt.Sprintf(`%s`, v)
	case val.Bool:
		if v {
			return `true`
		}
		return `false`
	case val.DateTime:
		return v.Format(time.RFC3339)
	case val.Float:
		return fmt.Sprintf(`%f`, v)
	case val.String:
		return fmt.Sprintf(`"%s"`, v)
	case val.Int8:
		return fmt.Sprintf(`%d`, v)
	case val.Int16:
		return fmt.Sprintf(`%d`, v)
	case val.Int32:
		return fmt.Sprintf(`%d`, v)
	case val.Int64:
		return fmt.Sprintf(`%d`, v)
	case val.Uint8:
		return fmt.Sprintf(`%d`, v)
	case val.Uint16:
		return fmt.Sprintf(`%d`, v)
	case val.Uint32:
		return fmt.Sprintf(`%d`, v)
	case val.Uint64:
		return fmt.Sprintf(`%d`, v)
	}
	return "value"
	panic(fmt.Sprintf(`unhandled value type: %T`, v))
}
