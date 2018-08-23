// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"fmt"
	"github.com/kr/pretty"
	"karma.run/common"
	"karma.run/definitions"
	"karma.run/kvm/err"
	"karma.run/kvm/inst"
	"karma.run/kvm/mdl"
	"karma.run/kvm/val"
	"karma.run/kvm/xpr"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

func unMeta(value val.Value) val.Value {
	if mm, ok := value.(val.Meta); ok {
		return mm.Value
	}
	return value
}

type ValueScope struct {
	parent *ValueScope
	scope  map[string]val.Value
}

func NewValueScope() *ValueScope {
	return &ValueScope{nil, make(map[string]val.Value)}
}

func (s *ValueScope) Get(k string) (val.Value, bool) {
	if s == nil {
		return nil, false
	}
	if m, ok := s.scope[k]; ok {
		return m, true
	}
	return s.parent.Get(k)
}

func (s *ValueScope) Set(k string, m val.Value) {
	s.scope[k] = m
}

func (s *ValueScope) Child() *ValueScope {
	c := NewValueScope()
	c.parent = s
	return c
}

type Stack []val.Value

func NewStack(capacity int) *Stack {
	stack := make(Stack, 0, capacity)
	return &stack
}

func (s *Stack) Push(v val.Value) {
	(*s) = append((*s), v)
}

func (s *Stack) Pop() val.Value {
	l := s.Len()
	v := (*s)[l-1]
	(*s) = (*s)[:l-1]
	return v
}

func (s Stack) Len() int {
	return len(s)
}

func (s Stack) Peek() val.Value {
	if len(s) == 0 {
		return nil
	}
	return s[len(s)-1]
}

// scope may be nil, that's fine -- will be allocated when needed.
func (vm VirtualMachine) Execute(program inst.Sequence, scope *ValueScope, args ...val.Value) (val.Value, err.Error) {

	if len(program) == 0 {
		panic("empty program")
	}

	stack := NewStack(128)
	for i := len(args) - 1; i > -1; i-- {
		stack.Push(args[i])
	}

	if ct, ok := program[0].(inst.Constant); ok && len(program) == 1 {
		return ct.Value, nil
	}

	if e := vm.lazyLoadPermissions(); e != nil {
		return nil, e
	}

	for pc, pl := 0, len(program); pc < pl; pc++ {

		// { // debug
		// 	fmt.Printf("stack.Peek(): %T %v\n", stack.Peek(), stack.Peek())
		// 	fmt.Printf("instruction: %T %v\n\n", program[pc], program[pc])
		// }

		switch it := program[pc].(type) {

		case inst.Sequence:
			log.Panicln("vm.Execute: nested inst.Sequence")

		case inst.Pop:
			stack.Pop()

		case inst.Define:
			if scope == nil {
				scope = NewValueScope()
			}
			scope.Set(string(it), stack.Pop())
			stack.Push(val.Null)

		case inst.Scope:
			v, ok := scope.Get(string(it))
			if !ok {
				return nil, err.ExecutionError{
					Problem: fmt.Sprintf(`not in scope: "%s"`, it),
				}
			}
			stack.Push(v)

		case inst.CurrentUser:
			stack.Push(val.Ref{vm.UserModelId(), vm.UserID})

		case inst.Constant:
			stack.Push(it.Value)

		case inst.If:
			cont := inst.Sequence(nil)
			if stack.Pop().(val.Bool) {
				cont = it.Then
			} else {
				cont = it.Else
			}
			v, e := vm.Execute(cont, scope.Child())
			if e != nil {
				return nil, e
			}
			stack.Push(v)

		case inst.SubstringIndex:
			search := unMeta(stack.Pop()).(val.String)
			stryng := unMeta(stack.Pop()).(val.String)
			stack.Push(val.Int64(
				strings.Index(string(stryng), string(search)),
			))

		case inst.BuildList:
			ls := make(val.List, it.Length, it.Length)
			for i := it.Length - 1; i > -1; i-- {
				ls[i] = stack.Pop()
			}
			stack.Push(ls)

		case inst.BuildSet:
			st := make(val.Set, it.Length)
			for i := it.Length - 1; i > -1; i-- {
				v := stack.Pop()
				st[val.Hash(v, nil).Sum64()] = v
			}
			stack.Push(st)

		case inst.BuildTuple:
			tp := make(val.Tuple, it.Length, it.Length)
			for i := it.Length - 1; i > -1; i-- {
				tp[i] = stack.Pop()
			}
			stack.Push(tp)

		case inst.BuildMap:
			mp := val.NewMap(it.Length)
			for i, l := 0, it.Length; i < l; i++ {
				v := stack.Pop()
				k := unMeta(stack.Pop()).(val.String)
				mp.Set(string(k), v)
			}
			stack.Push(mp)

		case inst.BuildStruct:
			st := val.NewStruct(len(it.Keys))
			for i := len(it.Keys) - 1; i > -1; i-- {
				st.Set(it.Keys[i], stack.Pop())
			}
			stack.Push(st)

		case inst.BuildUnion:
			stack.Push(val.Union{Case: it.Case, Value: stack.Pop()})

		case inst.Not:
			stack.Push(!unMeta(stack.Pop()).(val.Bool))

		case inst.StringToRef:
			rf := val.Ref{it.Model, string(unMeta(stack.Pop()).(val.String))}
			if _, e := vm.Get(rf[0], rf[1]); e != nil {
				return nil, e
			}
			stack.Push(rf)

		case inst.TagExists:
			tag := unMeta(stack.Pop()).(val.String)
			mid := vm.RootBucket.Bucket(definitions.TagBucketBytes).Get([]byte(tag))
			stack.Push(val.Bool(mid != nil))

		case inst.Tag:
			tag := unMeta(stack.Pop()).(val.String)
			mid := vm.RootBucket.Bucket(definitions.TagBucketBytes).Get([]byte(tag))
			if mid == nil {
				return nil, err.ExecutionError{
					Problem: fmt.Sprintf(`tag not found: "%s"`, tag),
				}
			}
			stack.Push(val.Ref{vm.MetaModelId(), string(mid)})

		case inst.JoinStrings:
			separator := unMeta(stack.Pop()).(val.String)
			strings, e := slurpIterators(unMeta(stack.Pop()))
			if e != nil {
				return nil, e
			}
			out := val.String("")
			for i, v := range strings.(val.List) {
				if i > 0 {
					out += separator
				}
				out += v.(val.String)
			}
			stack.Push(out)

		case inst.ExtractStrings:
			value, e := slurpIterators(unMeta(stack.Pop()))
			if e != nil {
				return nil, e
			}
			strings := make(val.List, 0, 64)
			value.Transform(func(v val.Value) val.Value {
				if sv, ok := v.(val.String); ok {
					strings = append(strings, sv)
				}
				return v
			})
			// sort for some predictability, useful when joining strings afterwards
			sort.Slice(strings, func(i, j int) bool {
				return strings[i].(val.String) < strings[j].(val.String)
			})
			stack.Push(strings)

		case inst.Delete:

			rf := unMeta(stack.Pop()).(val.Ref)

			v, e := vm.Get(rf[0], rf[1])
			if e != nil {
				return nil, e
			}

			stack.Push(unMeta(v)) // remove meta information so that program can't get a ref to a now inexistent object

			mids := modelsInMigrationTree(vm.reverseMigrationTree(rf[0], nil), nil)
			mids = append(mids, rf[0])

			for _, mid := range mids {
				if vm.permissions != nil && vm.permissions.delete != nil {
					v, e := vm.Get(mid, rf[1])
					if e != nil {
						if _, ok := e.(err.ObjectNotFoundError); ok {
							continue
						}
						return nil, e
					}
					if e := vm.CheckPermission(DeletePermission, v); e != nil {
						return nil, e
					}
				}
				if e := vm.Delete(mid, rf[1]); e != nil {
					return nil, e
				}
			}

		case inst.Update:

			vl := unMeta(stack.Pop())
			rf := unMeta(stack.Pop()).(val.Ref)

			mm, e := vm.Model(rf[0])
			if e != nil {
				return nil, e
			}

			if e := mm.Validate(vl, nil); e != nil {
				return nil, e
			}

			// vl already validated against mm

			migrationMap := vm.applyMigrationTree(rf[1], map[string]*migrationNode{
				rf[0]: {
					InValue:   vl,
					Migration: xpr.ValueFromFunction(xpr.NewFunction([]string{"input"}, xpr.Scope("input"))),
					Children:  vm.migrationTree(rf[0], nil),
				},
			}, nil)

			{
				stack.Push(rf) // push a ref
			}

			for mid, v := range migrationMap {

				if mid == vm.MetaModelId() {
					return nil, &err.ExecutionError{
						`update: would lead to model mutation through migration tree.`, nil,
					}
				}

				ov, e := vm.Get(mid, rf[1])
				if e != nil {
					if _, ok := e.(err.ObjectNotFoundError); ok {
						continue // skip non-existent object in migration target
					}
					return nil, e
				}

				v.Created = ov.Created // preserve creation datestamp

				if vm.permissions != nil && vm.permissions.update != nil {
					if e := vm.CheckPermission(UpdatePermission, ov); e != nil {
						return nil, e
					}
					if e := vm.CheckPermission(UpdatePermission, v); e != nil {
						return nil, e
					}
				}

				if e := vm.Write(mid, map[string]val.Meta{rf[1]: v}); e != nil {
					return nil, e
				}

			}

		case inst.CreateMultiple:

			mm, e := vm.Model(it.Model)
			if e != nil {
				return nil, e
			}

			subArg := val.NewStruct(len(it.Values)) // key -> id as val.Ref
			for k, _ := range it.Values {
				subArg.Set(k, val.Ref{it.Model, common.RandomId()})
			}

			vs := make(map[string]val.Value, len(it.Values))
			for k, sub := range it.Values {
				w, e := vm.Execute(sub, scope.Child(), subArg)
				if e != nil {
					return nil, e
				}
				vs[k] = w
			}

			rm := make(map[string]map[string]val.Meta) // mid -> id -> value

			migrationTree := vm.migrationTree(it.Model, nil)

			for k, v := range vs {

				if e = mm.Validate(v, nil); e != nil {
					break
				}

				id := string(subArg.Field(k).(val.Ref)[1])

				if it.Model == vm.MetaModelId() {
					// in this case validating structure is not enough,
					// we have to make sure that we can build an actual model
					// from the given value
					if _, e = mdl.ModelFromValue(vm.MetaModelId(), v.(val.Union), nil); e != nil {
						break
					}
				}

				migrationMap := vm.applyMigrationTree(id, map[string]*migrationNode{
					it.Model: {
						InModel:   mm,
						InValue:   v,
						Migration: xpr.ValueFromFunction(xpr.NewFunction([]string{"input"}, xpr.Scope("input"))),
						Children:  migrationTree,
					},
				}, nil)

				for mid, v := range migrationMap {
					if _, ok := rm[mid]; !ok {
						rm[mid] = make(map[string]val.Meta, len(vs))
					}
					rm[mid][id] = v
				}

			}
			if e != nil {
				return nil, e
			}

			stack.Push(subArg)

			// TODO: should primary mid go first for more sensible model IDs in errors?
			for mid, values := range rm {
				if vm.permissions != nil && vm.permissions.create != nil {
					for _, v := range values {
						if e := vm.CheckPermission(CreatePermission, v); e != nil {
							return nil, e
						}
					}
				}
				if e := vm.Write(mid, values); e != nil {
					return nil, e
				}
			}

		case inst.All:
			mid := (unMeta(stack.Pop())).(val.Ref)[1]
			m, e := vm.Model(mid)
			if e != nil {
				return nil, e
			}
			model := vm.WrapModelInMeta(mid, m.Model)
			bucket := vm.RootBucket.Bucket([]byte(mid))
			iter := iterator(newBucketDecodingIterator(bucket, model))
			if vm.permissions != nil && vm.permissions.read != nil {
				iter = vm.newReadPermissionFilterIterator(iter)
			}
			stack.Push(iteratorValue{iter})

		case inst.LeftFoldList:
			init := stack.Pop()
			list := stack.Pop()
			switch list := list.(type) {
			case val.List:
				for _, v := range list {
					v, e := vm.Execute(it.Reducer, scope.Child(), init, v)
					if e != nil {
						return nil, e
					}
					init = v
				}
			case iteratorValue:
				e := list.iterator.forEach(func(v val.Value) err.Error {
					v, e := vm.Execute(it.Reducer, scope.Child(), init, v)
					if e != nil {
						return e
					}
					init = v
					return nil
				})
				if e != nil {
					return nil, e
				}

			default:
				log.Panicf("Execute: LeftFoldList: unexpected type on stack: %T.", list)
			}
			stack.Push(init)

		case inst.RightFoldList:
			init := stack.Pop()
			list := stack.Pop()
			if iv, ok := list.(iteratorValue); ok {
				vs, e := iteratorToList(iv.iterator)
				if e != nil {
					return nil, e
				}
				list = vs
			}
			lsvs := list.(val.List)
			for i := len(lsvs) - 1; i > -1; i-- {
				v, e := vm.Execute(it.Reducer, scope.Child(), init, lsvs[i])
				if e != nil {
					return nil, e
				}
				init = v
			}
			stack.Push(init)

		case inst.ReduceList:

			initial := unMeta(stack.Pop())

			value, e := slurpIterators(unMeta(stack.Pop()))
			if e != nil {
				return nil, e
			}

			vs := value.(val.List)

			v := initial
			for _, w := range vs {
				x, e := vm.Execute(it.Expression, scope.Child(), v, w)
				if e != nil {
					return nil, e
				}
				v = x
			}

			stack.Push(v)

		case inst.StringToLower:
			s := unMeta(stack.Pop()).(val.String)
			stack.Push(val.String(strings.ToLower(string(s))))

		case inst.ReverseList:

			var out val.List

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				reverse(ls)
				out = ls

			case iteratorValue:
				var e err.Error
				out, e = iteratorToList(ls.iterator)
				if e != nil {
					return nil, e
				}
				reverse(out)

			default:
				log.Panicf("Execute: reverseList: unexpected type on stack: %T.", ls)
			}

			stack.Push(out)

		case inst.With:
			v := stack.Pop()
			w, e := vm.Execute(it.Expression, scope.Child(), v)
			if e != nil {
				return nil, e
			}
			stack.Push(w)

		case inst.MapList:

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				cp := make(val.List, len(ls), len(ls))
				for i, value := range ls {
					mapped, e := vm.Execute(it.Expression, scope.Child(), val.Int64(i), value)
					if e != nil {
						return nil, e
					}
					cp[i] = mapped
				}
				stack.Push(cp)

			case iteratorValue:
				i := val.Int64(-1)
				stack.Push(iteratorValue{
					newMappingIterator(ls.iterator, func(v val.Value) (val.Value, err.Error) {
						i++
						return vm.Execute(it.Expression, scope.Child(), i, v)
					}),
				})

			default:
				log.Panicf("Execute: MapList: unexpected type on stack: %T.", ls)
			}

		case inst.MapSet:
			v := unMeta(stack.Pop()).(val.Set)
			c := make(val.Set, len(v))
			for _, v := range v {
				w, e := vm.Execute(it.Expression, scope.Child(), v)
				if e != nil {
					return nil, e
				}
				if _, ok := w.(iteratorValue); ok {
					c[rand.Uint64()] = w // always treat iterators as distinct
				} else {
					c[val.Hash(w, nil).Sum64()] = w
				}
			}
			stack.Push(c)

		case inst.MapMap:

			mv := unMeta(stack.Pop()).(val.Map)
			e := (err.Error)(nil)
			mapped := mv.Map(func(k string, v val.Value) val.Value {
				if e != nil {
					return nil
				}
				mapped, e_ := vm.Execute(it.Expression, scope.Child(), val.String(k), v)
				if e_ != nil {
					e = e_
				}
				return mapped
			})
			if e != nil {
				return nil, e
			}
			stack.Push(mapped)

		case inst.MapStruct:
			mv := unMeta(stack.Pop()).(val.Struct)
			e := (err.Error)(nil)
			mapped := mv.Map(func(k string, v val.Value) val.Value {
				if e != nil {
					return nil
				}
				mapped, e_ := vm.Execute(it.Expression, scope.Child(), val.String(k), v)
				if e_ != nil {
					e = e_
				}
				return mapped
			})
			if e != nil {
				return nil, e
			}
			stack.Push(mapped)

		case inst.GraphFlow:

			output := make(map[string]val.Value)
			seen := make(map[val.Ref]struct{}, 1024)
			todo := make([]val.Ref, 0, 1024)
			todo = append(todo, unMeta(stack.Pop()).(val.Ref))

			for len(todo) > 0 {

				vertex := todo[0]
				todo = todo[1:]

				v, e := vm.Get(vertex[0], vertex[1])
				if e != nil {
					return nil, e
				}

				if vm.permissions != nil && vm.permissions.read != nil {
					if e := vm.CheckPermission(ReadPermission, v); e != nil {
						return nil, e
					}
				}

				seen[vertex] = struct{}{}

				if output[vertex[0]] == nil {
					output[vertex[0]] = val.NewMap(32)
				}

				output[vertex[0]].(val.Map).Set(vertex[1], v)

				flow := it.FlowParams[vertex[0]]

				if ob := vm.RootBucket.Bucket(definitions.GraphBucketBytes).Bucket(encodeVertex(vertex[0], vertex[1])); ob != nil {
					e := ob.ForEach(func(k, _ []byte) error {
						m, i := decodeVertex(k)
						if _, ok := flow.Forward[m]; !ok {
							return nil // skip
						}
						if _, ok := seen[val.Ref{m, i}]; !ok {
							todo = append(todo, val.Ref{m, i})
						}
						return nil
					})
					if e != nil {
						log.Panicln(e)
					}
				}

				if ib := vm.RootBucket.Bucket(definitions.PhargBucketBytes).Bucket(encodeVertex(vertex[0], vertex[1])); ib != nil {
					e := ib.ForEach(func(k, _ []byte) error {
						m, i := decodeVertex(k)
						if _, ok := flow.Backward[m]; !ok {
							return nil // skip
						}
						if _, ok := seen[val.Ref{m, i}]; !ok {
							todo = append(todo, val.Ref{m, i})
						}
						return nil
					})
					if e != nil {
						log.Panicln(e)
					}
				}
			}

			stack.Push(val.StructFromMap(output))

		case inst.ResolveRefs:

			v := unMeta(stack.Pop()).Copy().Transform(func(v val.Value) val.Value {
				if r, ok := v.(val.Ref); ok {
					if _, ok := it.Models[r[0]]; ok {
						w, e := vm.Get(r[0], r[1])
						if e != nil {
							log.Panicln(e) // should not happen
						}
						return w
					}
				}
				return v
			})

			stack.Push(v)

		case inst.ResolveAllRefs:

			v := unMeta(stack.Pop()).Copy().Transform(func(v val.Value) val.Value {
				if r, ok := v.(val.Ref); ok {
					w, e := vm.Get(r[0], r[1])
					if e != nil {
						log.Panicln(e) // should not happen
					}
					return w
				}
				return v
			})

			stack.Push(v)

		case inst.RelocateRef:

			ref := unMeta(stack.Pop()).(val.Ref)

			if ref[0] == it.Model {
				stack.Push(ref)
				continue
			}

			migs := vm.RootBucket.Bucket(definitions.MigrationBucketBytes)
			todo := make([]string, 0, 16)
			todo = append(todo, ref[0])
			done := make(map[string]struct{})
			path := false

			for len(todo) > 0 {
				origin := todo[0]
				todo = todo[1:]
				if origin == it.Model {
					path = true
					break // found a path! yay
				}
				if _, ok := done[origin]; ok {
					continue
				}
				done[origin] = struct{}{}
				fb := migs.Bucket([]byte(origin))
				if fb == nil {
					continue
				}
				e := fb.ForEach(func(t, _ []byte) error {
					to := string(t)
					if _, ok := done[to]; ok {
						return nil // continue
					}
					todo = append(todo, to)
					return nil // continue
				})
				if e != nil {
					log.Panicln(e)
				}
			}

			if !path {
				return nil, err.ExecutionError{
					Problem: `relocateRef: no migration path between source and target models`,
				}
			}

			stack.Push(val.Ref{it.Model, ref[1]})

		case inst.Referred:

			from := unMeta(stack.Pop()).(val.Ref)

			gb := vm.RootBucket.Bucket(definitions.GraphBucketBytes)
			if gb == nil {
				log.Panicln("Graph bucket missing!")
			}

			key := encodeVertex(from[0], from[1])

			kb := gb.Bucket(key)
			if kb == nil {
				stack.Push((val.List)(nil))
				continue
			}

			vals := make(val.List, 0, 32)

			e := kb.ForEach(func(key, _ []byte) error {
				if m, i := decodeVertex(key); m == it.In {
					vals = append(vals, val.Ref{m, i})
				}
				return nil
			})

			if e != nil {
				log.Panicln(e)
			}

			filtered := vals[:0]
			for _, v := range vals {
				r := v.(val.Ref)
				if _, e := vm.Get(r[0], r[1]); e != nil {
					if _, ok := e.(err.PermissionDeniedError); ok {
						continue
					}
					return nil, e
				}
				filtered = append(filtered, v)
			}

			if len(vals) == 0 {
				stack.Push((val.List)(nil))
				continue
			}

			stack.Push(filtered)

		case inst.Referrers:

			of := unMeta(stack.Pop()).(val.Ref)

			pb := vm.RootBucket.Bucket(definitions.PhargBucketBytes)
			if pb == nil {
				log.Panicln("Pharg bucket missing!")
			}

			key := encodeVertex(of[0], of[1])

			kb := pb.Bucket(key)
			if kb == nil {
				stack.Push((val.List)(nil))
				continue
			}

			vals := make(val.List, 0, 32)

			e := kb.ForEach(func(key, _ []byte) error {
				if m, i := decodeVertex(key); m == it.In {
					vals = append(vals, val.Ref{m, i})
				}
				return nil
			})

			if e != nil {
				log.Panicln(e)
			}

			filtered := vals[:0]
			for _, v := range vals {
				r := v.(val.Ref)
				if _, e := vm.Get(r[0], r[1]); e != nil {
					if _, ok := e.(err.PermissionDeniedError); ok {
						continue
					}
					return nil, e
				}
				filtered = append(filtered, v)
			}

			if len(vals) == 0 {
				stack.Push((val.List)(nil))
				continue
			}

			stack.Push(filtered)

		case inst.Filter:

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				cp := make(val.List, 0, len(ls))
				for i, value := range ls {
					keep, e := vm.Execute(it.Expression, scope.Child(), val.Uint64(i), value)
					if e != nil {
						return nil, e
					}
					if keep.(val.Bool) {
						cp = append(cp, value)
					}
				}
				stack.Push(cp)

			case iteratorValue:
				i := 0
				stack.Push(iteratorValue{
					newFilterIterator(ls.iterator, func(value val.Value) (bool, err.Error) {
						v, e := vm.Execute(it.Expression, scope.Child(), val.Uint64(i), value)
						if e != nil {
							return false, e
						}
						i++
						return bool(v.(val.Bool)), nil
					}),
				})

			default:
				log.Panicf("unexpected type on stack: %T", ls)
			}

		case inst.First:

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				if len(ls) == 0 {
					return nil, err.ExecutionError{
						Problem: fmt.Sprintf("first: empty list"),
					}
				}
				stack.Push(ls[0])

			case iteratorValue:
				out := val.Value(nil)
				ir := newLimitIterator(ls.iterator, 0, 1)
				e := ir.forEach(func(v val.Value) err.Error {
					out = v
					return nil
				})
				if e != nil {
					return nil, e
				}
				if out == nil {
					return nil, err.ExecutionError{
						Problem: fmt.Sprintf("first: empty list"),
					}
				}
				stack.Push(out)

			default:
				log.Panicf("unexpected type on stack: %T", ls)
			}

		case inst.InList:

			vl := unMeta(stack.Pop())

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				out := val.Bool(false)
				for _, v := range ls {
					if v.Equals(vl) {
						out = true
					}
				}
				stack.Push(out)

			case iteratorValue:
				out := val.Bool(false)
				stop := &err.ExecutionError{} // placeholder
				e := ls.iterator.forEach(func(v val.Value) err.Error {
					if v.Equals(vl) {
						out = true
						return stop
					}
					return nil // continue
				})
				if e == stop {
					e = nil
				}
				if e != nil {
					return nil, e
				}
				stack.Push(out)

			default:
				log.Panicf("Execute: InList: unexpected type on stack: %T.", ls)
			}

		case inst.DateTimeNow:
			stack.Push(val.DateTime{time.Now()})

		case inst.OrList:

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				out := val.Bool(false)
				for _, v := range ls {
					if v.(val.Bool) {
						out = true
						break
					}
				}
				stack.Push(out)

			case iteratorValue:
				out := val.Bool(false)
				stop := &err.ExecutionError{}
				e := ls.iterator.forEach(func(v val.Value) err.Error {
					if v.(val.Bool) {
						out = true
						return stop
					}
					return nil // continue
				})
				if e == stop {
					e = nil
				}
				if e != nil {
					return nil, e
				}
				stack.Push(out)

			default:
				log.Panicf("Execute: OrList: unexpected type on stack: %T.", ls)
			}

		case inst.PresentOrConstant:
			v := unMeta(stack.Pop())
			if v == val.Null {
				stack.Push(it.Constant)
			} else {
				stack.Push(v)
			}

		case inst.AllReferrers:
			v := unMeta(stack.Pop()).(val.Ref)
			bucket := vm.RootBucket.Bucket(definitions.PhargBucketBytes).Bucket(encodeVertex(v[0], v[1]))
			ls := (val.List)(nil)
			if bucket != nil {
				ls = make(val.List, 0, bucket.Stats().KeyN)
				e := bucket.ForEach(func(k, _ []byte) error {
					m, i := decodeVertex(k)
					ls = append(ls, val.Ref{m, i})
					return nil
				})
				if e != nil {
					log.Panicf("Execute: AllReferrers: %s.", e.Error())
				}
			}
			stack.Push(ls)

		case inst.IsPresent:
			v := unMeta(stack.Pop())
			stack.Push(val.Bool(v != val.Null))

		case inst.Slice:
			length := int(unMeta(stack.Pop()).(val.Int64))
			offset := int(unMeta(stack.Pop()).(val.Int64))
			if length < 0 {
				return nil, err.ExecutionError{
					Problem: `slice: negative length`,
				}
			}
			if offset < 0 {
				return nil, err.ExecutionError{
					Problem: `slice: negative offset`,
				}
			}
			switch value := unMeta(stack.Pop()).(type) {
			case val.List:
				value = value[minInt(len(value), offset):]
				value = value[:minInt(len(value), length)]
				stack.Push(value)

			case iteratorValue:
				stack.Push(iteratorValue{
					newLimitIterator(value.iterator, offset, length),
				})
			}

		case inst.SearchAllRegex:
			v := unMeta(stack.Pop()).(val.String)
			r := it.Regex.FindAllStringIndex(string(v), -1)
			if r == nil {
				stack.Push(val.List(nil))
			} else {
				ls := make(val.List, len(r), len(r))
				for i, t := range r {
					ls[i] = val.Int64(t[0])
				}
				stack.Push(ls)
			}

		case inst.SearchRegex:
			v := unMeta(stack.Pop()).(val.String)
			r := it.Regex.FindStringIndex(string(v))
			if r == nil {
				stack.Push(val.Int64(-1))
			} else {
				stack.Push(val.Int64(r[0]))
			}
		case inst.MatchRegex:
			v := unMeta(stack.Pop()).(val.String)
			stack.Push(val.Bool(it.Regex.Match([]byte(v))))

		case inst.AssertPresent:
			v := unMeta(stack.Pop())
			if v == val.Null {
				return nil, err.ExecutionError{
					Problem: `assertPresent: value was absent`,
				}
			}
			stack.Push(v)

		case inst.Enforce:

			df := unMeta(stack.Pop())
			vl := unMeta(stack.Pop())
			if vl == val.Null {
				stack.Push(df)
			} else {
				stack.Push(vl)
			}

		case inst.ShortCircuitAnd:

			vl := unMeta(stack.Pop()).(val.Bool)
			if vl {
				if it.Arity == it.Step { // this was the last check
					stack.Push(val.Bool(true))
				}
				continue
			}
			stack.Push(val.Bool(false))
			// fast forward to last ShortCircuitAnd (outer loop increments pc thereafter)
			for {
				if and, ok := program[pc].(inst.ShortCircuitAnd); ok {
					if and.Step == it.Arity {
						break
					}
				}
				pc++
			}

		case inst.ShortCircuitOr:

			vl := unMeta(stack.Pop()).(val.Bool)
			if !vl {
				if it.Arity == it.Step { // this was the last possibility
					stack.Push(val.Bool(false))
				}
				continue
			}
			stack.Push(val.Bool(true))
			for {
				if or, ok := program[pc].(inst.ShortCircuitOr); ok {
					if or.Step == it.Arity {
						break
					}
				}
				pc++
			}

		case inst.AssertCase:
			u := unMeta(stack.Pop()).(val.Union)
			if u.Case != it.Case {
				return nil, err.ExecutionError{
					Problem: fmt.Sprintf(`assertCase: case was "%s", not "%s"`, u.Case, it.Case),
				}
			}
			stack.Push(u.Value)

		case inst.IsCase:
			u := unMeta(stack.Pop()).(val.Union)
			c := unMeta(stack.Pop()).(val.String)
			stack.Push(val.Bool(u.Case == string(c)))

		case inst.AssertModelRef:
			v := stack.Pop()
			m, ok := v.(val.Meta)
			if !ok {
				return nil, err.ExecutionError{
					Problem: `assertModelRef: value is not persistent`,
				}
			}
			if m.Model[1] != it.Model {
				return nil, err.ExecutionError{
					Problem: `assertModelRef: assertion failed`,
				}
			}
			stack.Push(v)

		case inst.SwitchModelRef:
			v := stack.Pop()
			m, ok := v.(val.Meta)
			if !ok {
				return nil, err.ExecutionError{
					Problem: `switchModelRef: value is not persistent`,
				}
			}
			c, ok := it.Cases[m.Model[1]]
			if !ok {
				c = it.Default
			}
			v, e := vm.Execute(c, scope.Child(), m)
			if e != nil {
				return nil, e
			}
			stack.Push(v)

		case inst.Key:
			k := unMeta(stack.Pop()).(val.String)
			m := unMeta(stack.Pop()).(val.Map)
			if v, ok := m.Get(string(k)); ok {
				stack.Push(v)
			} else {
				stack.Push(val.Null)
			}

		case inst.Substring:
			l := int(unMeta(stack.Pop()).(val.Int64))
			o := int(unMeta(stack.Pop()).(val.Int64))
			s := string(unMeta(stack.Pop()).(val.String))
			b := make([]rune, 0, len(s))
			i := -1
			for _, r := range s {
				i++
				if i < o {
					continue
				}
				if len(b) == l {
					break
				}
				b = append(b, r)
			}
			stack.Push(val.String(b))

		case inst.IndexTuple:
			stack.Push(unMeta(stack.Pop()).(val.Tuple)[it.Number])

		case inst.SetField:
			in := unMeta(stack.Pop()).(val.Struct)
			out := in.Copy().(val.Struct)
			out.Set(it.Field, unMeta(stack.Pop()))
			stack.Push(out)

		case inst.SetKey:
			in := unMeta(stack.Pop()).(val.Map)
			out := in.Copy().(val.Map)
			out.Set(it.Key, unMeta(stack.Pop()))
			stack.Push(out)

		case inst.Field:
			stack.Push(unMeta(stack.Pop()).(val.Struct).Field(it.Key))

		case inst.Metarialize:
			stack.Push(MaterializeMeta(stack.Pop().(val.Meta)))

		case inst.Meta:
			mv := stack.Pop().(val.Meta)
			switch it.Key {
			case "id":
				stack.Push(mv.Id)
			case "created":
				stack.Push(mv.Created)
			case "updated":
				stack.Push(mv.Updated)
			case "model":
				stack.Push(mv.Model)
			}

		case inst.MemSortFunction:

			list := (val.List)(nil)

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				list = ls.Copy().(val.List)
			case iteratorValue:
				l, e := iteratorToList(ls.iterator)
				if e != nil {
					return nil, e
				}
				list = l
			default:
				log.Panicf("Execute: MemSortFunction: unexpected type on stack: %T.", ls)
			}

			errout := (err.Error)(nil)
			sort.Slice(list, func(i, j int) bool {
				v, e := vm.Execute(it.Less, scope, list[i], list[j])
				if e != nil {
					if errout == nil {
						errout = e
					}
					return false
				}
				return bool(v.(val.Bool))
			})

			stack.Push(list)

		case inst.MemSort:

			type sortable struct {
				comparable, value val.Value
			}

			expr := it.Expression
			temp := make([]sortable, 0, 1024)

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				for _, item := range ls {
					comparable, e := vm.Execute(expr, scope.Child(), item)
					if e != nil {
						return nil, e // TODO: add context
					}
					temp = append(temp, sortable{comparable, item})
				}
			case iteratorValue:
				e := ls.iterator.forEach(func(item val.Value) err.Error {
					comparable, e := vm.Execute(expr, scope.Child(), item)
					if e != nil {
						return e
					}
					temp = append(temp, sortable{comparable, item})
					return nil // continue
				})
				if e != nil {
					return nil, e // TODO: add context
				}
			default:
				log.Panicf("Execute: MemSort: unexpected type on stack: %T.", ls)
			}

			if len(temp) == 0 {

				stack.Push(make(val.List, 0, 0))

			} else {

				var lessFunc func(i, j int) bool

				switch temp[0].comparable.(type) {
				case val.Float:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Float) < temp[j].comparable.(val.Float)
					}
				case val.Bool:
					lessFunc = func(i, j int) bool {
						return bool(!temp[i].comparable.(val.Bool) && temp[j].comparable.(val.Bool))
					}
				case val.String:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.String) < temp[j].comparable.(val.String)
					}
				case val.DateTime:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.DateTime).Time.Before(
							temp[j].comparable.(val.DateTime).Time,
						)
					}
				case val.Int8:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Int8) < temp[j].comparable.(val.Int8)
					}
				case val.Int16:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Int16) < temp[j].comparable.(val.Int16)
					}
				case val.Int32:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Int32) < temp[j].comparable.(val.Int32)
					}
				case val.Int64:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Int64) < temp[j].comparable.(val.Int64)
					}
				case val.Uint8:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Uint8) < temp[j].comparable.(val.Uint8)
					}
				case val.Uint16:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Uint16) < temp[j].comparable.(val.Uint16)
					}
				case val.Uint32:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Uint32) < temp[j].comparable.(val.Uint32)
					}
				case val.Uint64:
					lessFunc = func(i, j int) bool {
						return temp[i].comparable.(val.Uint64) < temp[j].comparable.(val.Uint64)
					}
				}

				sort.Slice(temp, lessFunc)

				out := make(val.List, len(temp), len(temp))
				for i, sortable := range temp {
					out[i] = sortable.value
				}
				stack.Push(out)
			}

		case inst.Deref:
			rf := unMeta(stack.Pop()).(val.Ref)
			v, e := vm.Get(rf[0], rf[1])
			if e != nil {
				return nil, e
			}
			stack.Push(v)

		case inst.Length:

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				stack.Push(val.Int64(len(ls)))

			case iteratorValue:
				l := ls.iterator.length()
				if l == -1 {
					count := 0
					e := ls.iterator.forEach(func(val.Value) err.Error {
						count++
						return nil
					})
					if e != nil {
						return nil, e
					}
					l = count
				}
				stack.Push(val.Int64(l))

			default:
				log.Panicf("unexpected type on stack: %T", ls)
			}

		case inst.ConcatLists:

			rhs := unMeta(stack.Pop()) // order matters
			lhs := unMeta(stack.Pop()) // order matters

			li, ri := iterator(nil), iterator(nil)

			switch lhs := lhs.(type) {
			case val.List:
				li = newListIterator(lhs)
			case iteratorValue:
				li = lhs.iterator
			default:
				log.Panicf("unexpected type on stack: %T", lhs)
			}

			switch rhs := rhs.(type) {
			case val.List:
				ri = newListIterator(rhs)
			case iteratorValue:
				ri = rhs.iterator
			default:
				log.Panicf("unexpected type on stack: %T", rhs)
			}

			stack.Push(iteratorValue{newConcatIterator(li, ri)})

		case inst.After:
			rhs := unMeta(stack.Pop()).(val.DateTime) // order matters
			lhs := unMeta(stack.Pop()).(val.DateTime) // order matters
			stack.Push(val.Bool(lhs.Time.After(rhs.Time)))

		case inst.Before:
			rhs := unMeta(stack.Pop()).(val.DateTime) // order matters
			lhs := unMeta(stack.Pop()).(val.DateTime) // order matters
			stack.Push(val.Bool(lhs.Time.Before(rhs.Time)))

		case inst.Equal:
			a, b, null := unMeta(stack.Pop()), unMeta(stack.Pop()), val.Null
			if a == null || b == null {
				stack.Push(val.Bool(a == null && b == null))
				continue
			}
			if _, ok := a.(iteratorValue); ok {
				return nil, err.ExecutionError{
					Problem: `cannot compare persistent object collections`, // TODO: not always persistent stuff
				}
			}
			if _, ok := b.(iteratorValue); ok {
				return nil, err.ExecutionError{
					Problem: `cannot compare persistent object collections`, // TODO: not always persistent stuff
				}
			}
			stack.Push(val.Bool(
				a.Equals(b),
			))

		case inst.ToString:
			v := unMeta(stack.Pop())
			s := val.String("")
			switch v := v.(type) {
			case val.Int8:
				s = val.String(strconv.FormatInt(int64(v), 10))
			case val.Int16:
				s = val.String(strconv.FormatInt(int64(v), 10))
			case val.Int32:
				s = val.String(strconv.FormatInt(int64(v), 10))
			case val.Int64:
				s = val.String(strconv.FormatInt(int64(v), 10))
			case val.Uint8:
				s = val.String(strconv.FormatUint(uint64(v), 10))
			case val.Uint16:
				s = val.String(strconv.FormatUint(uint64(v), 10))
			case val.Uint32:
				s = val.String(strconv.FormatUint(uint64(v), 10))
			case val.Uint64:
				s = val.String(strconv.FormatUint(uint64(v), 10))
			case val.String:
				s = v
			default:
				panic(fmt.Sprintf("unexpected type on stack: %T", v))
			}
			stack.Push(s)

		case inst.ToFloat:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Float{}))

		case inst.ToInt:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Int64{}))

		case inst.ToInt8:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Int8{}))

		case inst.ToInt16:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Int16{}))

		case inst.ToInt32:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Int32{}))

		case inst.ToInt64:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Int64{}))

		case inst.ToUint:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Uint64{}))

		case inst.ToUint8:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Uint8{}))

		case inst.ToUint16:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Uint16{}))

		case inst.ToUint32:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Uint32{}))

		case inst.ToUint64:
			v := unMeta(stack.Pop())
			stack.Push(convertNumericType(v, mdl.Uint64{}))

		case inst.GreaterFloat:
			rhs := stack.Pop().(val.Float)
			lhs := stack.Pop().(val.Float)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterInt:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterInt8:
			rhs := stack.Pop().(val.Int8)
			lhs := stack.Pop().(val.Int8)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterInt16:
			rhs := stack.Pop().(val.Int16)
			lhs := stack.Pop().(val.Int16)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterInt32:
			rhs := stack.Pop().(val.Int32)
			lhs := stack.Pop().(val.Int32)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterInt64:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterUint:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterUint8:
			rhs := stack.Pop().(val.Uint8)
			lhs := stack.Pop().(val.Uint8)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterUint16:
			rhs := stack.Pop().(val.Uint16)
			lhs := stack.Pop().(val.Uint16)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterUint32:
			rhs := stack.Pop().(val.Uint32)
			lhs := stack.Pop().(val.Uint32)
			stack.Push(val.Bool(lhs > rhs))

		case inst.GreaterUint64:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(val.Bool(lhs > rhs))

		case inst.LessFloat:
			rhs := stack.Pop().(val.Float)
			lhs := stack.Pop().(val.Float)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessInt:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessInt8:
			rhs := stack.Pop().(val.Int8)
			lhs := stack.Pop().(val.Int8)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessInt16:
			rhs := stack.Pop().(val.Int16)
			lhs := stack.Pop().(val.Int16)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessInt32:
			rhs := stack.Pop().(val.Int32)
			lhs := stack.Pop().(val.Int32)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessInt64:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessUint:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessUint8:
			rhs := stack.Pop().(val.Uint8)
			lhs := stack.Pop().(val.Uint8)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessUint16:
			rhs := stack.Pop().(val.Uint16)
			lhs := stack.Pop().(val.Uint16)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessUint32:
			rhs := stack.Pop().(val.Uint32)
			lhs := stack.Pop().(val.Uint32)
			stack.Push(val.Bool(lhs < rhs))

		case inst.LessUint64:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(val.Bool(lhs < rhs))

		case inst.AddFloat:
			rhs := stack.Pop().(val.Float)
			lhs := stack.Pop().(val.Float)
			stack.Push(lhs + rhs)

		case inst.AddInt:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs + rhs)

		case inst.AddInt8:
			rhs := stack.Pop().(val.Int8)
			lhs := stack.Pop().(val.Int8)
			stack.Push(lhs + rhs)

		case inst.AddInt16:
			rhs := stack.Pop().(val.Int16)
			lhs := stack.Pop().(val.Int16)
			stack.Push(lhs + rhs)

		case inst.AddInt32:
			rhs := stack.Pop().(val.Int32)
			lhs := stack.Pop().(val.Int32)
			stack.Push(lhs + rhs)

		case inst.AddInt64:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs + rhs)

		case inst.AddUint:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs + rhs)

		case inst.AddUint8:
			rhs := stack.Pop().(val.Uint8)
			lhs := stack.Pop().(val.Uint8)
			stack.Push(lhs + rhs)

		case inst.AddUint16:
			rhs := stack.Pop().(val.Uint16)
			lhs := stack.Pop().(val.Uint16)
			stack.Push(lhs + rhs)

		case inst.AddUint32:
			rhs := stack.Pop().(val.Uint32)
			lhs := stack.Pop().(val.Uint32)
			stack.Push(lhs + rhs)

		case inst.AddUint64:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs + rhs)

		case inst.SubtractFloat:
			rhs := stack.Pop().(val.Float)
			lhs := stack.Pop().(val.Float)
			stack.Push(lhs - rhs)

		case inst.SubtractInt:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs - rhs)

		case inst.SubtractInt8:
			rhs := stack.Pop().(val.Int8)
			lhs := stack.Pop().(val.Int8)
			stack.Push(lhs - rhs)

		case inst.SubtractInt16:
			rhs := stack.Pop().(val.Int16)
			lhs := stack.Pop().(val.Int16)
			stack.Push(lhs - rhs)

		case inst.SubtractInt32:
			rhs := stack.Pop().(val.Int32)
			lhs := stack.Pop().(val.Int32)
			stack.Push(lhs - rhs)

		case inst.SubtractInt64:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs - rhs)

		case inst.SubtractUint:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs - rhs)

		case inst.SubtractUint8:
			rhs := stack.Pop().(val.Uint8)
			lhs := stack.Pop().(val.Uint8)
			stack.Push(lhs - rhs)

		case inst.SubtractUint16:
			rhs := stack.Pop().(val.Uint16)
			lhs := stack.Pop().(val.Uint16)
			stack.Push(lhs - rhs)

		case inst.SubtractUint32:
			rhs := stack.Pop().(val.Uint32)
			lhs := stack.Pop().(val.Uint32)
			stack.Push(lhs - rhs)

		case inst.SubtractUint64:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs - rhs)

		case inst.MultiplyFloat:
			rhs := stack.Pop().(val.Float)
			lhs := stack.Pop().(val.Float)
			stack.Push(lhs * rhs)

		case inst.MultiplyInt:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs * rhs)

		case inst.MultiplyInt8:
			rhs := stack.Pop().(val.Int8)
			lhs := stack.Pop().(val.Int8)
			stack.Push(lhs * rhs)

		case inst.MultiplyInt16:
			rhs := stack.Pop().(val.Int16)
			lhs := stack.Pop().(val.Int16)
			stack.Push(lhs * rhs)

		case inst.MultiplyInt32:
			rhs := stack.Pop().(val.Int32)
			lhs := stack.Pop().(val.Int32)
			stack.Push(lhs * rhs)

		case inst.MultiplyInt64:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs * rhs)

		case inst.MultiplyUint:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs * rhs)

		case inst.MultiplyUint8:
			rhs := stack.Pop().(val.Uint8)
			lhs := stack.Pop().(val.Uint8)
			stack.Push(lhs * rhs)

		case inst.MultiplyUint16:
			rhs := stack.Pop().(val.Uint16)
			lhs := stack.Pop().(val.Uint16)
			stack.Push(lhs * rhs)

		case inst.MultiplyUint32:
			rhs := stack.Pop().(val.Uint32)
			lhs := stack.Pop().(val.Uint32)
			stack.Push(lhs * rhs)

		case inst.MultiplyUint64:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs * rhs)

		case inst.DivideFloat:
			rhs := stack.Pop().(val.Float)
			lhs := stack.Pop().(val.Float)
			stack.Push(lhs / rhs)

		case inst.DivideInt:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs / rhs)

		case inst.DivideInt8:
			rhs := stack.Pop().(val.Int8)
			lhs := stack.Pop().(val.Int8)
			stack.Push(lhs / rhs)

		case inst.DivideInt16:
			rhs := stack.Pop().(val.Int16)
			lhs := stack.Pop().(val.Int16)
			stack.Push(lhs / rhs)

		case inst.DivideInt32:
			rhs := stack.Pop().(val.Int32)
			lhs := stack.Pop().(val.Int32)
			stack.Push(lhs / rhs)

		case inst.DivideInt64:
			rhs := stack.Pop().(val.Int64)
			lhs := stack.Pop().(val.Int64)
			stack.Push(lhs / rhs)

		case inst.DivideUint:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs / rhs)

		case inst.DivideUint8:
			rhs := stack.Pop().(val.Uint8)
			lhs := stack.Pop().(val.Uint8)
			stack.Push(lhs / rhs)

		case inst.DivideUint16:
			rhs := stack.Pop().(val.Uint16)
			lhs := stack.Pop().(val.Uint16)
			stack.Push(lhs / rhs)

		case inst.DivideUint32:
			rhs := stack.Pop().(val.Uint32)
			lhs := stack.Pop().(val.Uint32)
			stack.Push(lhs / rhs)

		case inst.DivideUint64:
			rhs := stack.Pop().(val.Uint64)
			lhs := stack.Pop().(val.Uint64)
			stack.Push(lhs / rhs)

		case inst.SwitchCase:
			u := unMeta(stack.Pop()).(val.Union) // union value
			if c, ok := it[u.Case]; ok {
				v, e := vm.Execute(c, scope.Child(), u.Value)
				if e != nil {
					return nil, e
				}
				stack.Push(v)
				continue
			}
			return nil, err.ExecutionError{
				Problem: fmt.Sprintf(`switchCase: unexpected case: %s`, u.Case),
			}

		default:
			panic(fmt.Sprintf("unimplemented inst.Instruction: %#v", it))
		}

	}

	if stack.Len() != 1 {
		log.Panicf("Execute: stack had %d elements after execution.", stack.Len())
	}

	return stack.Pop(), nil
}

func printDebugProgram(stack *Stack, program []inst.Instruction) {
	pretty.Println("=== printDebugProgram ===")
	pretty.Println("stack:", *stack)
	pretty.Println("program:", program)
	pretty.Println("=========================")
}
