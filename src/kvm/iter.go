// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"codec/karma.v2"
	"github.com/boltdb/bolt"
	"kvm/err"
	"kvm/val"
)

type Iterator interface {
	Init() err.Error
	Next() (val.Value, err.Error) // (nil, nil) == exhausted
	Close()
}

func IterForEach(i Iterator, f func(val.Value) (bool, err.Error)) err.Error {
	for v, e := i.Next(); ; v, e = i.Next() {
		if e != nil {
			return e
		}
		if v == nil {
			return nil // done
		}
		if c, e := f(v); !c || e != nil {
			return e
		}
	}
	return nil
}

type listIterator struct {
	list   val.List
	cursor int
}

func newListIterator(l val.List) *listIterator {
	return &listIterator{l, 0}
}

func (i listIterator) Init() err.Error {
	return nil
}

func (i *listIterator) Next() (val.Value, err.Error) {
	if i.cursor >= len(i.list) {
		return nil, nil
	}
	v := i.list[i.cursor]
	i.cursor++
	return v, nil
}

func (i listIterator) Close() {
}

type concatIterator struct {
	l, r Iterator
}

func (i concatIterator) Init() err.Error {
	le, re := i.r.Init(), i.r.Init()
	if le == nil {
		return re
	}
	if re == nil {
		return le
	}
	return err.ErrorList{le, re}
}
func (i concatIterator) Next() (val.Value, err.Error) {
	lv, le := i.l.Next()
	if le != nil {
		return lv, le
	}
	if lv != nil {
		return lv, nil
	}
	rv, re := i.r.Next()
	if re != nil {
		return rv, re
	}
	return rv, nil
}
func (i concatIterator) Close() {
	i.l.Close()
	i.r.Close()
}

// convenience adapter for Iterators that implements val.Value
type IteratorVal struct {
	Iterator Iterator
}

func (IteratorVal) Equals(v val.Value) bool {
	panic("IteratorVal.Equals called")
}

func (IteratorVal) Copy() val.Value {
	panic("IteratorVal.Copy called")
}

func (IteratorVal) Type() val.Type {
	panic("IteratorVal.Type called")
}

func (v IteratorVal) Transform(f func(val.Value) val.Value) val.Value {
	return f(IteratorVal{
		mapListIterator{
			SubIterator: v.Iterator,
			MapFunc: func(v val.Value) (val.Value, err.Error) {
				return v.Transform(f), nil
			},
		},
	})
}

func (IteratorVal) Primitive() bool {
	return false
}

type limitIterator struct {
	SubIterator Iterator
	Skip, Pass  int
}

func (*limitIterator) Init() err.Error { return nil }

func (i *limitIterator) Next() (val.Value, err.Error) {
	for i.Skip > 0 {
		i.SubIterator.Next()
		i.Skip--
	}
	if i.Pass == 0 {
		return nil, nil
	}
	i.Pass--
	return i.SubIterator.Next()
}

func (i limitIterator) Close() {
	i.SubIterator.Close()
}

type mapListIterator struct {
	SubIterator Iterator
	MapFunc     func(val.Value) (val.Value, err.Error)
}

func (mapListIterator) Init() err.Error { return nil }

func (i mapListIterator) Next() (val.Value, err.Error) {

	v, e := i.SubIterator.Next()

	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, nil
	}

	return i.MapFunc(v)

}

func (i mapListIterator) Close() {
	i.SubIterator.Close()
}

type filterIterator struct {
	SubIterator Iterator
	FilterFunc  func(val.Value) (bool, err.Error)
}

func (filterIterator) Init() err.Error { return nil }

func (i filterIterator) Next() (val.Value, err.Error) {

	for {
		v, e := i.SubIterator.Next()

		if e != nil {
			return nil, e
		}

		if v == nil {
			return nil, nil
		}

		keep, e := i.FilterFunc(v)
		if e != nil {
			return nil, e
		}

		if !keep {
			continue
		}

		return v, nil
	}

}

func (i filterIterator) Close() {
	i.SubIterator.Close()
}

type bucketIterator struct {
	VM     VirtualMachine
	Mid    string
	Model  BucketModel
	Cursor *bolt.Cursor // initialized in Init
}

func (i *bucketIterator) Init() err.Error {

	if i.Cursor != nil {
		return nil // already initialized
	}

	bk := i.VM.RootBucket.Bucket([]byte(i.Mid))
	if bk == nil {
		return err.ModelNotFoundError{
			err.ObjectNotFoundError{
				Ref: val.Ref{i.VM.MetaModelId(), i.Mid},
			},
		}
	}

	i.Cursor = bk.Cursor()

	return nil
}

func (i *bucketIterator) Next() (val.Value, err.Error) {

	kb, vb := ([]byte)(nil), ([]byte)(nil)

	if i.Cursor == nil {
		if e := i.Init(); e != nil {
			return nil, e
		}
		kb, vb = i.Cursor.First() // must be invoked at beginning
	} else {
		kb, vb = i.Cursor.Next()
	}

	if kb == nil {
		return nil, nil // finished
	}

	v, _ := karma.Decode(vb, i.VM.WrapModelInMeta(i.Mid, i.Model.Model))

	return DematerializeMeta(v.(val.Struct)), nil
}

func (i *bucketIterator) Close() {
	if i.Cursor == nil {
		return
	}
	i.Cursor = nil
}

type permissionIterator struct {
	VM          VirtualMachine
	SubIterator Iterator
}

func (permissionIterator) Init() err.Error {
	return nil
}

func (i permissionIterator) Next() (val.Value, err.Error) {

_continue: // recursion could easily overflow stack

	v, e := i.SubIterator.Next()

	if e != nil {
		return nil, e
	}

	if v == nil {
		return nil, nil
	}

	if mv, ok := v.(val.Meta); ok && i.VM.permissions != nil && i.VM.permissions.read != nil {
		if e := i.VM.CheckPermission(ReadPermission, mv); e != nil {
			if _, ok := e.(err.PermissionDeniedError); ok {
				goto _continue
			}
			return nil, e
		}
	}

	return v, nil

}

func (i permissionIterator) Close() {
	i.SubIterator.Close()
}
