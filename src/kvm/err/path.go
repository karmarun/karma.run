// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"fmt"
	"kvm/val"
	"strings"
)

type PathedError interface {
	Error
	ErrorPath() ErrorPath
	AppendPath(ErrorPathElement, ...ErrorPathElement) PathedError
}

type ErrorPath []ErrorPathElement

func (p ErrorPath) String() string {
	out := ""
	for i, _ := range p {
		l := p[len(p)-i-1]
		if i > 0 {
			out += "\n" + strings.Repeat("  ", i)
		}
		out += l.String()
	}
	return out
}

func (p ErrorPath) Value() val.List {
	l := make(val.List, len(p), len(p))
	for i, loc := range p {
		l[i] = loc.Value()
	}
	return l
}

func (p ErrorPath) Equals(q ErrorPath) bool {
	if len(p) != len(q) {
		return false
	}
	for i, _ := range p {
		if p[i] != q[i] {
			return false
		}
	}
	return true
}

type ErrorPathElement interface {
	String() string
	Value() val.Union
}

type ErrorPathElementUnionCase string

func (l ErrorPathElementUnionCase) String() string {
	return fmt.Sprintf(`union case "%s"`, string(l))
}

func (l ErrorPathElementUnionCase) Value() val.Union {
	return val.Union{"unionCase", val.String(l)}
}

type ErrorPathElementStructField string

func (l ErrorPathElementStructField) String() string {
	return fmt.Sprintf(`struct field "%s"`, string(l))
}

func (l ErrorPathElementStructField) Value() val.Union {
	return val.Union{"structField", val.String(l)}
}

type ErrorPathElementMapKey string

func (l ErrorPathElementMapKey) String() string {
	return fmt.Sprintf(`map key "%s"`, string(l))

}

func (l ErrorPathElementMapKey) Value() val.Union {
	return val.Union{"mapKey", val.String(l)}
}

type ErrorPathElementMapElements struct{}

func (l ErrorPathElementMapElements) String() string {
	return `map elements`
}

func (l ErrorPathElementMapElements) Value() val.Union {
	return val.Union{"mapElements", val.Struct{}}
}

type ErrorPathElementSetElements struct{}

func (l ErrorPathElementSetElements) String() string {
	return `set elements`
}

func (l ErrorPathElementSetElements) Value() val.Union {
	return val.Union{"setElements", val.Struct{}}
}

type ErrorPathElementListElements struct{}

func (l ErrorPathElementListElements) String() string {
	return `list elements`
}

func (l ErrorPathElementListElements) Value() val.Union {
	return val.Union{"listElements", val.Struct{}}
}

type ErrorPathElementTupleIndex int

func (l ErrorPathElementTupleIndex) String() string {
	return fmt.Sprintf(`tuple index %d`, int(l))

}

func (l ErrorPathElementTupleIndex) Value() val.Union {
	return val.Union{"tupleIndex", val.Int64(l)}
}

type ErrorPathElementListIndex int

func (l ErrorPathElementListIndex) String() string {
	return fmt.Sprintf(`list index %d`, int(l))
}

func (l ErrorPathElementListIndex) Value() val.Union {
	return val.Union{"listIndex", val.Int64(l)}
}

type ErrorPathElementSetItem struct{}

func (l ErrorPathElementSetItem) String() string {
	return `set elements`
}

func (l ErrorPathElementSetItem) Value() val.Union {
	return val.Union{"setItem", val.Struct{}}
}
