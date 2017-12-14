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
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func unMeta(value val.Value) val.Value {
	if mm, ok := value.(val.Meta); ok {
		return mm.Value
	}
	return value
}

var StackPool = &sync.Pool{
	New: func() interface{} {
		s := make(Stack, 0, 32)
		return &s
	},
}

func (vm VirtualMachine) Execute(program inst.Sequence, input val.Value) (val.Value, err.Error) {

	if len(program) == 0 {
		panic("empty program")
	}

	db := vm.RootBucket

	if ct, ok := program[0].(inst.Constant); ok && len(program) == 1 {
		return ct.Value, nil
	}

	if e := vm.lazyLoadPermissions(); e != nil {
		return nil, e
	}

	stack := StackPool.Get().(*Stack)

	defer func() {
		s := *stack
		for i, _ := range s {
			s[i] = nil
		}
		s = s[:0]
		StackPool.Put(&s)
	}()

	// pretty.Println(input)
	// printDebugProgram(stack, program)

	for pc, pl := 0, len(program); pc < pl; pc++ {

		switch it := program[pc].(type) {

		case inst.Sequence:
			log.Panicln("flattenSequences missed an inst.Sequence ;P")

		case inst.DebugPrintStack:
			pretty.Println("input:", input)
			pretty.Println("stack:", *stack)

		case inst.Identity:
			if input == nil {
				return nil, err.ExecutionError{
					`id/arg: no input in this context`,
					nil,
				}
			}
			stack.Push(input)

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
			v, e := vm.Execute(cont, input)
			if e != nil {
				return nil, e
			}
			stack.Push(v)

		case inst.BuildList:
			ls := make(val.List, it.Length, it.Length)
			for i := it.Length - 1; i > -1; i-- {
				ls[i] = stack.Pop()
			}
			stack.Push(ls)

		case inst.BuildTuple:
			ls := make(val.Tuple, it.Length, it.Length)
			for i := it.Length - 1; i > -1; i-- {
				ls[i] = stack.Pop()
			}
			stack.Push(ls)

		case inst.BuildMap:
			ls := val.NewMap(it.Length)
			for i, l := 0, it.Length; i < l; i++ {
				v := stack.Pop()
				k := unMeta(stack.Pop()).(val.String)
				ls.Set(string(k), v)
			}
			stack.Push(ls)

		case inst.BuildStruct:
			v := val.NewStruct(len(it.Keys))
			for i := len(it.Keys) - 1; i > -1; i-- {
				v.Set(it.Keys[i], stack.Pop())
			}
			stack.Push(v)

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

		case inst.Tag:
			tag := unMeta(stack.Pop()).(val.String)
			mid := db.Bucket(definitions.TagBucketBytes).Get([]byte(tag))
			if mid == nil {
				return nil, err.ExecutionError{
					fmt.Sprintf(`tag: "%s" not found`, tag),
					nil,
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

			// vl already validated against mm

			migrationMap := vm.applyMigrationTree(rf[1], map[string]*migrationNode{
				rf[0]: {
					InModel:   mm, // correct, even if it seems weird
					InValue:   vl,
					Migration: val.Union{Case: "id", Value: val.Struct{}},
					Children:  vm.migrationTree(rf[0], nil),
				},
			}, nil)

			{
				stack.Push(rf) // push a ref
			}

			for mid, v := range migrationMap {

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

		case inst.PopToInput:
			input = stack.Pop().Copy()

		case inst.CreateMultiple:

			mm, e := vm.Model(it.Model)
			if e != nil {
				return nil, e
			}

			vs := unMeta(stack.Pop()).(val.Map)        // vm's elements already validated against mm
			im := make(map[string]val.Value, vs.Len()) // key -> id as val.Ref
			rm := make(map[string]map[string]val.Meta) // mid -> id -> value

			// allocate some random IDs
			vs.ForEach(func(k string, v val.Value) bool {
				im[k] = val.Ref{it.Model, common.RandomId()}
				return true
			})

			migrationTree := vm.migrationTree(it.Model, nil)

			vs.ForEach(func(k string, v val.Value) bool {

				id := string(im[k].(val.Ref)[1])

				v = unMeta(v).Transform(func(v val.Value) val.Value {
					if r, ok := v.(val.Ref); ok {
						if id, ok := im[r[1]]; ok {
							return id
						}
					}
					return v
				})

				if it.Model == vm.MetaModelId() {
					// in this case validating structure is not enough,
					// we have to make sure that we can build an actual model
					// from the given value
					if _, e = mdl.ModelFromValue(vm.MetaModelId(), v.(val.Union), nil); e != nil {
						return false
					}
				}

				migrationMap := vm.applyMigrationTree(id, map[string]*migrationNode{
					it.Model: {
						InModel:   mm,
						InValue:   v,
						Migration: val.Union{Case: "id", Value: val.Struct{}},
						Children:  migrationTree,
					},
				}, nil)

				for mid, v := range migrationMap {
					if _, ok := rm[mid]; !ok {
						rm[mid] = make(map[string]val.Meta, vs.Len())
					}
					rm[mid][id] = v
				}

				return true
			})
			if e != nil {
				return nil, e
			}

			stack.Push(val.MapFromMap(im))

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
			bucket := vm.RootBucket.Bucket([]byte(mid))
			bkir := newBucketRefIterator(mid, bucket)
			drir := vm.newDerefMappingIterator(bkir)
			prir := vm.newReadPermissionFilterIterator(drir)
			stack.Push(iteratorValue{prir})

		case inst.MapStruct:
			mv := unMeta(stack.Pop()).(val.Struct)
			e := (err.Error)(nil)
			temp := val.NewStruct(2)
			mapped := mv.Map(func(k string, v val.Value) val.Value {
				if e != nil {
					return nil
				}
				temp.Set("field", val.String(k))
				temp.Set("value", v)
				mapped, e_ := vm.Execute(it.Expression, temp)
				if e_ != nil {
					e = e_
				}
				return mapped
			})
			if e != nil {
				return nil, e
			}
			stack.Push(mapped)

		case inst.MapMap:

			mv := unMeta(stack.Pop()).(val.Map)
			e := (err.Error)(nil)
			temp := val.NewStruct(2)
			mapped := mv.Map(func(k string, v val.Value) val.Value {
				if e != nil {
					return nil
				}
				temp.Set("key", val.String(k))
				temp.Set("value", v)
				mapped, e_ := vm.Execute(it.Expression, temp)
				if e_ != nil {
					e = e_
				}
				return mapped
			})
			if e != nil {
				return nil, e
			}
			stack.Push(mapped)

		case inst.ReduceList:

			value, e := slurpIterators(unMeta(stack.Pop()))
			if e != nil {
				return nil, e
			}

			vs := value.(val.List)
			if len(vs) == 0 {
				return nil, err.ExecutionError{
					`reduce: empty list value`,
					nil,
				}
			}

			v := vs[0]
			for _, w := range vs[1:] {
				x, e := vm.Execute(it.Expression, val.Tuple{v, w})
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
				log.Panicf("Execute: MapList: unexpected type on stack: %T.", ls)
			}

			stack.Push(out)

		case inst.MapList:

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				cp := make(val.List, len(ls), len(ls))
				for i, value := range ls {
					mapped, e := vm.Execute(it.Expression, value)
					if e != nil {
						return nil, e
					}
					cp[i] = mapped
				}
				stack.Push(cp)

			case iteratorValue:
				stack.Push(iteratorValue{
					newMappingIterator(ls.iterator, func(v val.Value) (val.Value, err.Error) {
						return vm.Execute(it.Expression, v)
					}),
				})

			default:
				log.Panicf("Execute: MapList: unexpected type on stack: %T.", ls)
			}

		case inst.GraphFlow:

			ov := make(map[string]val.Value)
			in := make([]val.Ref, 0, 1024)
			sn := make(map[val.Ref]struct{}, 1024)
			in = append(in, unMeta(stack.Pop()).(val.Ref))

			for len(in) > 0 {

				cv := in[0]
				in = in[1:]

				v, e := vm.Get(cv[0], cv[1])
				if e != nil {
					return nil, e
				}

				if vm.permissions != nil && vm.permissions.read != nil {
					if e := vm.CheckPermission(ReadPermission, v); e != nil {
						return nil, e
					}
				}

				sn[cv] = struct{}{}
				if ov[cv[0]] == nil {
					ov[cv[0]] = val.NewMap(32)
				}
				{
					first := ov[cv[0]].(val.Map)
					first.Set(cv[1], v)
				}

				flow := it.FlowParams[cv[0]]

				if ob := db.Bucket(definitions.GraphBucketBytes).Bucket(encodeVertex(cv[0], cv[1])); ob != nil {
					e := ob.ForEach(func(k, _ []byte) error {
						m, i := decodeVertex(k)
						if _, ok := flow.Forward[m]; !ok {
							return nil // skip
						}
						if _, ok := sn[val.Ref{m, i}]; !ok {
							in = append(in, val.Ref{m, i})
						}
						return nil
					})
					if e != nil {
						log.Panicln(e)
					}
				}

				if ib := db.Bucket(definitions.PhargBucketBytes).Bucket(encodeVertex(cv[0], cv[1])); ib != nil {
					e := ib.ForEach(func(k, _ []byte) error {
						m, i := decodeVertex(k)
						if _, ok := flow.Backward[m]; !ok {
							return nil // skip
						}
						if _, ok := sn[val.Ref{m, i}]; !ok {
							in = append(in, val.Ref{m, i})
						}
						return nil
					})
					if e != nil {
						log.Panicln(e)
					}
				}
			}

			stack.Push(val.StructFromMap(ov))

		case inst.ResolveRefs:

			v := unMeta(stack.Pop()).Transform(func(v val.Value) val.Value {
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

			v := unMeta(stack.Pop()).Transform(func(v val.Value) val.Value {
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

			migs := db.Bucket(definitions.MigrationBucketBytes)
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
					`relocateRef: no migration path between source and target models`,
					nil,
				}
			}

			stack.Push(val.Ref{it.Model, ref[1]})

		case inst.Referred:

			from := unMeta(stack.Pop()).(val.Ref)

			gb := db.Bucket(definitions.GraphBucketBytes)
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

			pb := db.Bucket(definitions.PhargBucketBytes)
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
				for _, value := range ls {
					keep, e := vm.Execute(it.Expression, value)
					if e != nil {
						return nil, e
					}
					if keep.(val.Bool) {
						cp = append(cp, value)
					}
				}
				stack.Push(cp)

			case iteratorValue:
				stack.Push(iteratorValue{
					newFilterIterator(ls.iterator, func(value val.Value) (bool, err.Error) {
						v, e := vm.Execute(it.Expression, value)
						if e != nil {
							return false, e
						}
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
						fmt.Sprintf("first: empty list"),
						nil,
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
						fmt.Sprintf("first: empty list"),
						nil,
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

		case inst.IsPresent:
			v := unMeta(stack.Pop())
			stack.Push(val.Bool(v != val.Null))

		case inst.Slice:
			length := int(unMeta(stack.Pop()).(val.Int64))
			offset := int(unMeta(stack.Pop()).(val.Int64))
			if length < 0 {
				return nil, err.ExecutionError{
					`slice: negative length`,
					nil,
				}
			}
			if offset < 0 {
				return nil, err.ExecutionError{
					`slice: negative offset`,
					nil,
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
					fmt.Sprintf(`assertPresent: value was absent`),
					nil,
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
					fmt.Sprintf(`assertCase: case is "%s", not "%s"`, u.Case, it.Case),
					nil,
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
					`assertModelRef: value is not persistent`,
					nil,
					// C: val.Map{"value": v},
				}
			}
			if m.Model[1] != it.Model {
				return nil, err.ExecutionError{
					`assertModelRef: assertion failed`,
					nil,
					// C: val.Map{
					//  "expected": val.Ref{m.Model[0], it.Model},
					//  "actual":   m.Model,
					// },
				}
			}
			stack.Push(v)

		case inst.SwitchModelRef:
			v := stack.Pop()
			m, ok := v.(val.Meta)
			if !ok {
				return nil, err.ExecutionError{
					`switchModelRef: value is not persistent`,
					nil,
					// C: val.Map{"value": v},
				}
			}
			c, ok := it.Cases[m.Model[1]]
			if !ok {
				c = it.Default
			}
			v, e := vm.Execute(c, m)
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
			in.Set(it.Field, unMeta(stack.Pop()))
			stack.Push(in)

		case inst.SetKey:
			in := unMeta(stack.Pop()).(val.Map)
			in.Set(it.Key, unMeta(stack.Pop()))
			stack.Push(in)

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

		case inst.MemSort:

			type sortable struct {
				comparable, value val.Value
			}

			expr := it.Expression
			temp := make([]sortable, 0, 1024)

			switch ls := unMeta(stack.Pop()).(type) {
			case val.List:
				for _, item := range ls {
					comparable, e := vm.Execute(expr, item)
					if e != nil {
						return nil, e // TODO: add context
					}
					temp = append(temp, sortable{comparable, item})
				}
			case iteratorValue:
				e := ls.iterator.forEach(func(item val.Value) err.Error {
					comparable, e := vm.Execute(expr, item)
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
					`cannot compare persistent object collections`, // TODO: not always persistent stuff
					nil,
				}
			}
			if _, ok := b.(iteratorValue); ok {
				return nil, err.ExecutionError{
					`cannot compare persistent object collections`, // TODO: not always persistent stuff
					nil,
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

		case inst.SwitchType:
			v := unMeta(stack.Pop())
			k := valueTypeKey(v)
			c, ok := it[k]
			if !ok {
				return nil, err.ExecutionError{
					fmt.Sprintf(`switchType: unexpected type: %s`, k),
					nil,
					// C: val.Map{
					//  "value": v,
					// },
				}
			}
			v, e := vm.Execute(c, v)
			if e != nil {
				return nil, e
			}
			stack.Push(v)

		case inst.SwitchCase:
			v := unMeta(stack.Pop())
			u := v.(val.Union)
			c, ok := it[u.Case]
			if !ok {
				return nil, err.ExecutionError{
					fmt.Sprintf(`switchCase: unexpected case: %s`, u.Case),
					nil,
					// C: val.Map{
					//  "value": v,
					// },
				}
			}
			v, e := vm.Execute(c, u.Value)
			if e != nil {
				return nil, e
			}
			stack.Push(v)

		case inst.MapSet:
			v := unMeta(stack.Pop()).(val.Set)
			c := make(val.Set, len(v))
			for _, v := range v {
				w, e := vm.Execute(it.Expression, v)
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

		default:
			panic(fmt.Sprintf("unimplemented inst.Instruction: %T: %v", it, it))
		}

	}

	if stack.Len() != 1 {
		log.Panicf("Execute: stack had %d elements after execution: %s.", stack.Len(), pretty.Sprint(stack))
	}

	return stack.Pop(), nil
}

func printDebugProgram(stack *Stack, program []inst.Instruction) {
	pretty.Println("=== printDebugProgram ===")
	pretty.Println("stack:", *stack)
	pretty.Println("program:", program)
	pretty.Println("=========================")
}

func flattenSequences(it inst.Instruction, bf []inst.Instruction) inst.Sequence {
	if bf == nil {
		bf = make([]inst.Instruction, 0, 32)
	}
	if sq, ok := it.(inst.Sequence); ok {
		for _, is := range sq {
			bf = flattenSequences(is, bf)
		}
		return bf
	}
	return append(bf, it)
}
