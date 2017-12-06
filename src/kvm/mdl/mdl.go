package mdl

import (
	"fmt"
	"kvm/err"
	"kvm/val"
	"time"
)

// Model represents a possibly nested and/or recursive data type.
type Model interface {

	// Copy copies a Model tree. It's equivalent to calling
	// Transform(mdl.TransformationIdentity)
	Copy() Model

	// Transform traverses a Model tree and returns the result
	// of mapping each of its nodes through function f.
	Transform(f func(Model) Model) Model

	// Zero returns the zero-value for the current model.
	// This is not possible with all model configurations,
	// especially not Refs. See Zeroable.
	Zero() val.Value

	// Zeroable reports wether calling Zero() on this Model is legal.
	Zeroable() bool

	// Unwrap returns the shallowest Model defined in
	// this package in a Model tree.
	Unwrap() Model

	// Concrete unwraps a Model to the shallowest Model that
	// affects its Value representation. In other words,
	// it removes semantic wrappers.
	Concrete() Model

	// Equals reports wether a Model tree equals another.
	Equals(Model) bool

	// Nullable reports wether a Model can be Null.
	Nullable() bool

	// TraverseValue traverses a Value tree v, calling function f
	// for every legal Value/Model combination therein.
	TraverseValue(v val.Value, f func(val.Value, Model))

	// Traverse is obsolete but still needed. Do not use it.
	Traverse([]string, func([]string, Model))

	// Returns the top-level Type of the Values in the set
	// of a Model. Returns TypeInvalid in case of any and or.
	ValueType() val.Type
}

const (
	FormatDateTime = time.RFC3339
)

// TransformIdentity is the identity function for Models
func TransformIdentity(m Model) Model {
	return m
}

// ValueFromModel returns a Value representation of Model model.
func ValueFromModel(metaId string, model Model, recursions map[*Recursion]struct{}) val.Value {

	if recursions == nil {
		recursions = make(map[*Recursion]struct{})
	}

	// catch top-level recursions and, if we have more than one recursion
	// involved in the model, use "recursive" constructor instead of "recursion"
	if top, ok := model.(*Recursion); ok && len(recursions) == 0 {
		recs := make(map[*Recursion]struct{})
		model.Traverse(nil, func(_ []string, m Model) {
			if r, ok := m.(*Recursion); ok {
				recs[r] = struct{}{}
			}
		})
		if len(recs) > 1 {
			mp := make(val.Map, len(recs))
			for r, _ := range recs {
				mp[r.Label] = ValueFromModel(metaId, r.Model, recs)
			}
			return val.Union{"recursive", val.Struct{
				"top":    val.String(top.Label),
				"models": mp,
			}}
		}
	}

	switch m := model.(type) {

	case Or:
		list := make(val.List, 0, 8)
		left := ValueFromModel(metaId, m[0], recursions)
		if u, ok := left.(val.Union); ok && u.Case == "or" {
			list = append(list, u.Value.(val.List)...)
		} else {
			list = append(list, left)
		}
		right := ValueFromModel(metaId, m[1], recursions)
		if u, ok := right.(val.Union); ok && u.Case == "or" {
			list = append(list, u.Value.(val.List)...)
		} else {
			list = append(list, right)
		}
		return val.Union{"or", list}

	case *Recursion:
		if _, ok := recursions[m]; ok {
			return val.Union{"recurse", val.String(m.Label)}
		}
		recursions[m] = struct{}{}
		o := val.Struct{}
		o["label"] = val.String(m.Label)
		o["model"] = ValueFromModel(metaId, m.Model, recursions)
		delete(recursions, m)
		return val.Union{"recursion", o}

	case Unique:
		return val.Union{"unique", ValueFromModel(metaId, m.Model, recursions)}

	case Annotation:
		return val.Union{
			"annotation", val.Struct{
				"value": val.String(m.Value),
				"model": ValueFromModel(metaId, m.Model, recursions),
			},
		}

	case Tuple:
		z := m
		sm := make(val.List, len(z), len(z))
		for i, m := range z {
			sm[i] = ValueFromModel(metaId, m, recursions)
		}
		return val.Union{"tuple", sm}

	case Enum:
		z := m
		ss := make(val.Set, len(z))
		for s, _ := range z {
			v := val.String(s)
			ss[val.Hash(v, nil).Sum64()] = v
		}
		return val.Union{"enum", ss}

	case List:
		return val.Union{"list", ValueFromModel(metaId, m.Elements, recursions)}

	case Set:
		return val.Union{"set", ValueFromModel(metaId, m.Elements, recursions)}

	case Map:
		return val.Union{"map", ValueFromModel(metaId, m.Elements, recursions)}

	case String:
		return val.Union{"string", val.Struct{}}

	case Struct:
		o := val.Map{}
		for k, v := range m {
			o[k] = ValueFromModel(metaId, v, recursions)
		}
		return val.Union{"struct", o}

	case Union:
		o := val.Map{}
		for k, v := range m {
			o[k] = ValueFromModel(metaId, v, recursions)
		}
		return val.Union{"union", o}

	case Null:
		return val.Union{"null", val.Struct{}}

	case Float:
		return val.Union{"float", val.Struct{}}

	case Int8:
		return val.Union{"int8", val.Struct{}}

	case Int16:
		return val.Union{"int16", val.Struct{}}

	case Int32:
		return val.Union{"int32", val.Struct{}}

	case Int64:
		return val.Union{"int64", val.Struct{}}

	case Uint8:
		return val.Union{"uint8", val.Struct{}}

	case Uint16:
		return val.Union{"uint16", val.Struct{}}

	case Uint32:
		return val.Union{"uint32", val.Struct{}}

	case Uint64:
		return val.Union{"uint64", val.Struct{}}

	case DateTime:
		return val.Union{"dateTime", val.Struct{}}

	case Bool:
		return val.Union{"bool", val.Struct{}}

	case Any:
		return val.Union{"any", val.Struct{}}

	case Ref:
		return val.Union{"ref", val.Ref{metaId, m.Model}}

	}
	panic(fmt.Sprintf(`Unhandled model: %T`, model))
}

// ModelFromValue returns a Model from its Value representation.
func ModelFromValue(metaId string, u val.Union, recursions map[string]*Recursion) (Model, err.PathedError) {
	// note: ModelFromValue may NOT call Copy or Transform on models.
	if recursions == nil {
		recursions = make(map[string]*Recursion)
	}
	switch u.Case {
	case "recursive":
		s := u.Value.(val.Struct)
		t := string(s["top"].(val.String))
		a := s["models"].(val.Map)
		for l, _ := range a {
			if _, ok := recursions[l]; ok {
				return nil, err.ModelParsingError{
					fmt.Sprintf(`recursion label already defined: %s`, l), u, nil,
				}
			}
			recursions[l] = &Recursion{Label: l}
		}
		if _, ok := recursions[t]; !ok {
			return nil, err.ModelParsingError{fmt.Sprintf(`no definition for top label: %s`, t), u, nil}
		}
		for l, _ := range a {
			m, e := ModelFromValue(metaId, a[l].(val.Union), recursions)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementStructField("models"), err.ErrorPathElementMapKey(l))
			}
			recursions[l].Model = m
		}
		r := recursions[t]
		for l, _ := range a {
			m := recursions[l]
			if couldRecurseInfinitely(m, nil) {
				return nil, err.ModelParsingError{`infinite recursion`, u, nil}
			}
			delete(recursions, l)
		}
		return r, nil

	case "annotation":
		v := u.Value.(val.Struct)
		a := string(v["value"].(val.String))
		m, e := ModelFromValue(metaId, v["model"].(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementStructField("model"))
		}
		return Annotation{Model: m, Value: a}, nil

	case "recursion":
		v := u.Value.(val.Struct)
		l := string(v["label"].(val.String))
		r := &Recursion{Label: l}
		if _, ok := recursions[l]; ok {
			return nil, err.ModelParsingError{fmt.Sprintf(`recursion label already defined: %s`, l), u, nil}
		}
		recursions[l] = r
		m, e := ModelFromValue(metaId, v["model"].(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementStructField("model"))
		}
		if couldRecurseInfinitely(m, nil) {
			return nil, err.ModelParsingError{`infinite recursion`, u, nil}
		}
		r.Model = m
		delete(recursions, l)
		return r, nil

	case "recurse":
		l := string(u.Value.(val.String))
		if _, ok := recursions[l]; !ok {
			return nil, err.ModelParsingError{fmt.Sprintf(`undefined recursion label: %s`, l), u, nil}
		}
		return recursions[l], nil

	case "enum":
		ss := u.Value.(val.Set)
		if len(ss) == 0 {
			return nil, err.ModelParsingError{`empty enum definition`, u, nil}
		}
		m := make(Enum, len(ss))
		for _, v := range ss {
			s := string(v.(val.String))
			if _, ok := m[s]; ok {
				return nil, err.ModelParsingError{fmt.Sprintf(`duplicated symbol: %s`, s), u, nil}
			}
			m[s] = struct{}{}
		}
		return m, nil

	case "set":
		w, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return Set{Elements: w}, nil

	case "list":
		w, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return List{Elements: w}, nil

	case "map":
		w, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return Map{Elements: w}, nil

	case "tuple":
		v := u.Value.(val.List)
		m := make(Tuple, len(v), len(v))
		for i, v := range v {
			w, e := ModelFromValue(metaId, v.(val.Union), recursions)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementListIndex(i))
			}
			m[i] = w
		}
		return m, nil

	case "struct":
		v := u.Value.(val.Map)
		m := Struct{}
		for k, v := range v {
			w, e := ModelFromValue(metaId, v.(val.Union), recursions)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementMapKey(k))
			}
			m[k] = w
		}
		return m, nil

	case "or":
		v := u.Value.(val.List)
		if len(v) == 0 {
			return nil, err.ModelParsingError{`empty "or"`, u, nil}
		}
		ms := make([]Model, len(v), len(v))
		for i, w := range v {
			m, e := ModelFromValue(metaId, w.(val.Union), recursions)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementListIndex(i))
			}
			ms[i] = m
		}
		return RollOr(ms), nil // NOTE: can't use Either/UnionOf here because we have half-finished *Recursions

	case "union":
		v := u.Value.(val.Map)
		m := Union{}
		for k, v := range v {
			w, e := ModelFromValue(metaId, v.(val.Union), recursions)
			if e != nil {
				return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case), err.ErrorPathElementMapKey(k))
			}
			m[k] = w
		}
		return m, nil

	case "optional":
		m, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		return Or{m, Null{}}, nil // NOTE: can't use Either/UnionOf here because we have half-finished *Recursions

	case "unique":
		m, e := ModelFromValue(metaId, u.Value.(val.Union), recursions)
		if e != nil {
			return nil, e.AppendPath(err.ErrorPathElementUnionCase(u.Case))
		}
		if _, ok := m.(Unique); ok {
			return nil, err.ModelParsingError{`nested unique`, u, nil}
		}
		return Unique{Model: m}, nil

	case "null":
		return Null{}, nil

	case "float":
		return Float{}, nil

	case "uint":
		return Uint64{}, nil

	case "int8":
		return Int8{}, nil

	case "int16":
		return Int16{}, nil

	case "int32":
		return Int32{}, nil

	case "int64":
		return Int64{}, nil

	case "uint8":
		return Uint8{}, nil

	case "uint16":
		return Uint16{}, nil

	case "uint32":
		return Uint32{}, nil

	case "uint64":
		return Uint64{}, nil

	case "int":
		return Int64{}, nil

	case "string":
		return String{}, nil

	case "dateTime":
		return DateTime{}, nil

	case "bool":
		return Bool{}, nil

	case "any":
		return Any{}, nil

	case "ref":
		return Ref{string(u.Value.(val.Ref)[1])}, nil
	}
	panic(fmt.Sprintf(`undefined model constructor: %s`, u.Case))
}

// RollOr left-reduces its argument list into an Or tree.
func RollOr(ms []Model) Model {
	if len(ms) == 0 {
		panic("mdl.RollOr of zero models")
	}
	if len(ms) == 1 {
		return ms[0]
	}
	return Or{ms[0], RollOr(ms[1:])}
}

// UnrollOr returns all individual submodels in an Or tree as a list.
func UnrollOr(m Model, c []Model) []Model {
	if o, ok := m.(Or); ok {
		return UnrollOr(o[1], UnrollOr(o[0], c))
	}
	return append(c, m)
}

func UnionOf(ms ...Model) Model {
	if len(ms) == 0 {
		panic("mdl.UnionOf zero arguments")
	}
	if len(ms) == 1 {
		return ms[0]
	}
	return Either(ms[0], RollOr(ms[1:]), nil)
}

var any = Any{}

// Either returns a (hopefully) minimal Or-combination for two given Model arguments.
func Either(l, r Model, m map[*Recursion]*Recursion) Model {

	if l == nil || r == nil {
		panic("mdl.Either called with nil")
	}

	l, r = l.Unwrap(), r.Unwrap()

	if l == any || r == any {
		return any
	}

	// TODO: can be super expensive if there's lots of Ors in the model tree
	// if l.Equals(r) {
	// 	return l
	// }

	{ // handle recursions until convergence

		lr, lok := l.(*Recursion)
		rr, rok := r.(*Recursion)
		if (lok || rok) && m == nil {
			m = make(map[*Recursion]*Recursion, 8)
		}
		switch {
		case lok && rok: // both recursions
			if lp, ok := m[lr]; ok {
				rp, ok := m[rr]
				if !ok {
					c := &Recursion{Label: rr.Label}
					m[rr] = c
					c.Model = Either(lr.Model, rr.Model, m)
					delete(m, rr)
					return c
				}
				if lp == rp {
					return lp
				}
				return Or{lp, rp}
			}
			if rp, ok := m[rr]; ok {
				lp, ok := m[lr]
				if !ok {
					c := &Recursion{Label: lr.Label}
					m[lr] = c
					c.Model = Either(lr.Model, rr.Model, m)
					delete(m, lr)
					return c
				}
				if lp == rp {
					return lp
				}
				return Or{lp, rp}
			}
			// m[lr] == m[rr] == nil
			c := &Recursion{Label: labelUnion(lr.Label, rr.Label)}
			m[lr], m[rr] = c, c
			c.Model = Either(lr.Model, rr.Model, m)
			delete(m, lr)
			delete(m, rr)
			return c

		case lok && !rok: // left is recursive, right not
			if _, ok := m[lr]; ok {
				return Either(lr.Model, r, m)
			} else {
				c := &Recursion{Label: lr.Label}
				m[lr] = c
				c.Model = Either(lr.Model, r, m)
				delete(m, lr)
				return c
			}

		case !lok && rok: // left not recursive, right is
			if _, ok := m[rr]; ok {
				return Either(l, rr.Model, m)
			} else {
				c := &Recursion{Label: rr.Label}
				m[rr] = c
				c.Model = Either(l, rr.Model, m)
				delete(m, rr)
				return c
			}
		}
	}

	{ // handle "or" cases
		lo, lok := l.(Or)
		ro, rok := r.(Or)
		switch {
		case lok && rok: // both ors
			out := (Model)(ro)
			for _, w := range UnrollOr(lo, nil) {
				out = Either(w, out, m)
			}
			return out

		case lok && !rok: // left is "or", right not
			unrolled := UnrollOr(lo, nil)
			for i, w := range unrolled {
				either := Either(w, r, m)
				if _, ok := either.(Or); ok {
					continue
				}
				unrolled[i] = either
				return Either(unrolled[0], RollOr(unrolled[1:]), m) // minimize
			}
			return RollOr(append(unrolled, r))

		case !lok && rok: // left not "or", right is
			unrolled := UnrollOr(ro, nil)
			for i, w := range unrolled {
				either := Either(l, w, m)
				if _, ok := either.(Or); ok {
					continue
				}
				unrolled[i] = either
				return Either(unrolled[0], RollOr(unrolled[1:]), m) // minimize
			}
			return RollOr(append([]Model{l}, unrolled...))

		}
	}

	{ // handle unique
		lu, lok := l.(Unique)
		ru, rok := r.(Unique)
		switch {
		case lok && rok: // both uniques
			return Either(lu.Model, ru.Model, m)
		case lok && !rok: // left is unique, right not
			return Either(lu.Model, r, m)
		case !lok && rok: // left not unique, right is
			return Either(l, ru.Model, m)
		}
	}

	{ // handle annotations
		la, lok := l.(Annotation)
		ra, rok := r.(Annotation)
		switch {
		case lok && rok: // both annotations
			return Annotation{la.Value, Annotation{ra.Value, Either(la.Model, ra.Model, m)}}
		case lok && !rok: // left is annotation, right not
			return Annotation{la.Value, Either(la.Model, r, m)}
		case !lok && rok: // left not annotation, right is
			return Annotation{ra.Value, Either(l, ra.Model, m)}
		}
	}

	switch l := l.(type) {
	case Set:
		q, ok := r.(Set)
		if !ok {
			return Or{l, r}
		}
		return Set{Either(l.Elements, q.Elements, m)}
	case List:
		q, ok := r.(List)
		if !ok {
			return Or{l, r}
		}
		return List{Either(l.Elements, q.Elements, m)}
	case Map:
		q, ok := r.(Map)
		if !ok {
			return Or{l, r}
		}
		return Map{Either(l.Elements, q.Elements, m)}
	case Tuple:
		q, ok := r.(Tuple)
		if !ok {
			return Or{l, r}
		}
		if len(l) == len(q) {
			t := make(Tuple, len(q), len(q))
			for i, _ := range l {
				t[i] = Either(l[i], q[i], m)
			}
			return t
		}
		return Or{l, r}

	case Struct:
		q, ok := r.(Struct)
		if !ok {
			return Or{l, r}
		}
		s := make(Struct, len(l)+len(q))
		for k, w := range l {
			if x, ok := q[k]; ok {
				s[k] = Either(w, x, m)
			} else {
				s[k] = w
			}
		}
		for k, x := range q {
			if _, ok := s[k]; ok {
				continue
			}
			s[k] = x
		}
		return s

	case Union:
		q, ok := r.(Union)
		if !ok {
			return Or{l, r}
		}
		u := make(Union, len(l)+len(q))
		for k, w := range l {
			if x, ok := q[k]; ok {
				u[k] = Either(w, x, m)
			} else {
				u[k] = w
			}
		}
		for k, x := range q {
			if _, ok := u[k]; ok {
				continue
			}
			u[k] = x
		}
		return u

	case Enum:
		q, ok := r.(Enum)
		if !ok {
			return Or{l, r}
		}
		e := make(Enum, len(l)+len(q))
		for k, _ := range l {
			e[k] = struct{}{}
		}
		for k, _ := range q {
			e[k] = struct{}{}
		}
		return e

	case Ref:
		q, ok := r.(Ref)
		if !ok {
			return Or{l, r}
		}
		if l.Model == "" || q.Model == "" {
			return Ref{""}
		}
		if l.Model == q.Model {
			return l
		}
		return Or{l, r}
	case Null:
		_, ok := r.(Null)
		if !ok {
			return Or{l, r}
		}
		return l
	case String:
		_, ok := r.(String)
		if !ok {
			return Or{l, r}
		}
		return l
	case Float:
		_, ok := r.(Float)
		if !ok {
			return Or{l, r}
		}
		return l
	case Bool:
		_, ok := r.(Bool)
		if !ok {
			return Or{l, r}
		}
		return l
	case DateTime:
		_, ok := r.(DateTime)
		if !ok {
			return Or{l, r}
		}
		return l
	case Int8:
		_, ok := r.(Int8)
		if !ok {
			return Or{l, r}
		}
		return l
	case Int16:
		_, ok := r.(Int16)
		if !ok {
			return Or{l, r}
		}
		return l
	case Int32:
		_, ok := r.(Int32)
		if !ok {
			return Or{l, r}
		}
		return l
	case Int64:
		_, ok := r.(Int64)
		if !ok {
			return Or{l, r}
		}
		return l
	case Uint8:
		_, ok := r.(Uint8)
		if !ok {
			return Or{l, r}
		}
		return l
	case Uint16:
		_, ok := r.(Uint16)
		if !ok {
			return Or{l, r}
		}
		return l
	case Uint32:
		_, ok := r.(Uint32)
		if !ok {
			return Or{l, r}
		}
		return l
	case Uint64:
		_, ok := r.(Uint64)
		if !ok {
			return Or{l, r}
		}
		return l
	}
	panic(fmt.Sprintf("unhandled model type: %T\n", l))
}

func labelUnion(a, b string) string {
	if a == b {
		return a
	}
	return a + " | " + b
}

func couldRecurseInfinitely(m Model, seen map[*Recursion]struct{}) bool {
	switch m := m.(type) {
	case *Recursion:
		if seen == nil {
			seen = make(map[*Recursion]struct{})
		}
		if _, ok := seen[m]; ok {
			return true
		}
		seen[m] = struct{}{}
		return couldRecurseInfinitely(m.Model, seen)
	case Unique:
		return couldRecurseInfinitely(m.Model, seen)
	case Annotation:
		return couldRecurseInfinitely(m.Model, seen)
	case Or:
		return couldRecurseInfinitely(m[0], seen) || couldRecurseInfinitely(m[1], seen)
	}
	return false
}
