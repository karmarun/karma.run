// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"github.com/boltdb/bolt"
	"github.com/karmarun/karma.run/codec/karma.v2"
	"github.com/karmarun/karma.run/kvm/err"
	"github.com/karmarun/karma.run/kvm/val"
)

type Iterator interface {
	Reset() err.Error
	Next() (val.Value, err.Error) // (nil, nil) == exhausted
}

func IterForEach(i Iterator, f func(val.Value) (bool, err.Error)) err.Error {
	if e := i.Reset(); e != nil {
		return e
	}
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

func (i *listIterator) Reset() err.Error {
	i.cursor = 0
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

type concatIterator struct {
	l, r Iterator
}

func (i concatIterator) Reset() err.Error {
	le, re := i.r.Reset(), i.r.Reset()
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

// adapter for Iterators that implements val.Value
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
	skip, pass  int
}

func newLimitIterator(sub Iterator, skip, pass int) *limitIterator {
	return &limitIterator{sub, skip, pass, skip, pass}
}

func (i *limitIterator) Reset() err.Error {
	if e := i.SubIterator.Reset(); e != nil {
		return e
	}
	i.skip, i.pass = i.Skip, i.Pass
	return nil
}

func (i *limitIterator) Next() (val.Value, err.Error) {
	for i.skip > 0 {
		i.SubIterator.Next()
		i.skip--
	}
	if i.pass == 0 {
		return nil, nil
	}
	i.pass--
	return i.SubIterator.Next()
}

type mapListIterator struct {
	SubIterator Iterator
	MapFunc     func(val.Value) (val.Value, err.Error)
}

func (i mapListIterator) Reset() err.Error {
	return i.SubIterator.Reset()
}

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

type filterIterator struct {
	SubIterator Iterator
	FilterFunc  func(val.Value) (bool, err.Error)
}

func (i filterIterator) Reset() err.Error {
	return i.SubIterator.Reset()
}

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

type bucketIterator struct {
	VM     VirtualMachine
	Mid    string
	Model  BucketModel
	cursor *bolt.Cursor // initialized in Reset
	first  bool
}

func (i *bucketIterator) Reset() err.Error {

	bk := i.VM.RootBucket.Bucket([]byte(i.Mid))
	if bk == nil {
		return err.ModelNotFoundError{
			err.ObjectNotFoundError{
				Ref: val.Ref{i.VM.MetaModelId(), i.Mid},
			},
		}
	}

	i.cursor = bk.Cursor()
	i.first = true
	return nil
}

func (i *bucketIterator) Next() (val.Value, err.Error) {

	kb, vb := ([]byte)(nil), ([]byte)(nil)

	if i.cursor == nil {
		if e := i.Reset(); e != nil {
			return nil, e
		}
	}

	if i.first {
		kb, vb = i.cursor.First() // must be invoked at beginning
		i.first = false
	} else {
		kb, vb = i.cursor.Next()
	}

	if kb == nil {
		return nil, nil // finished
	}

	v, _ := karma.Decode(vb, i.VM.WrapModelInMeta(i.Mid, i.Model.Model))

	return DematerializeMeta(v.(val.Struct)), nil
}

type permissionIterator struct {
	VM          VirtualMachine
	SubIterator Iterator
}

func (i permissionIterator) Reset() err.Error {
	return i.SubIterator.Reset()
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
