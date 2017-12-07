// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package kvm

import (
	"github.com/boltdb/bolt"
	"github.com/karmarun/karma.run/kvm/err"
	"github.com/karmarun/karma.run/kvm/val"
)

// adapter for iterators that implements val.Value
type iteratorValue struct {
	iterator iterator
}

func (iteratorValue) Equals(v val.Value) bool {
	panic("iteratorValue.Equals called")
}

func (iteratorValue) Copy() val.Value {
	panic("iteratorValue.Copy called")
}

func (iteratorValue) Type() val.Type {
	panic("iteratorValue.Type called")
}

func (v iteratorValue) Transform(f func(val.Value) val.Value) val.Value {
	return f(iteratorValue{
		newMappingIterator(v.iterator, func(v val.Value) (val.Value, err.Error) {
			return v.Transform(f), nil
		}),
	})
}

func (iteratorValue) Primitive() bool {
	return false
}

type iterator interface {

	// if f returns a non-nil error, forEach stops and returns it
	forEach(f func(val.Value) err.Error) err.Error

	// if length returns -1, it's up to the caller to count the iterator
	length() int
}

type listIterator struct {
	list val.List
}

func newListIterator(l val.List) listIterator {
	return listIterator{l}
}

func (i listIterator) forEach(f func(val.Value) err.Error) err.Error {
	for _, v := range i.list {
		if e := f(v); e != nil {
			return e
		}
	}
	return nil
}

func (i listIterator) length() int {
	return len(i.list)
}

type concatIterator struct {
	left, right iterator
}

func newConcatIterator(left, right iterator) concatIterator {
	return concatIterator{left, right}
}

func (i concatIterator) forEach(f func(val.Value) err.Error) err.Error {
	e := i.left.forEach(f)
	if e != nil {
		return e
	}
	return i.right.forEach(f)
}

func (i concatIterator) length() int {
	l, r := i.left.length(), i.right.length()
	if l == -1 || r == -1 {
		return -1
	}
	return l + r
}

type limitIterator struct {
	sub  iterator
	skip int
	pass int
}

func newLimitIterator(sub iterator, skip, pass int) limitIterator {
	return limitIterator{sub, skip, pass}
}

func (i limitIterator) forEach(f func(val.Value) err.Error) err.Error {
	skipped, passed := 0, 0
	stop := &err.ExecutionError{} // placeholder
	e := i.sub.forEach(func(v val.Value) err.Error {
		if skipped < i.skip {
			skipped++
			return nil // continue
		}
		if passed < i.pass {
			if e := f(v); e != nil {
				return e
			}
			passed++
			return nil // continue
		}
		return stop
	})
	if e == stop {
		e = nil
	}
	return e
}

func (i limitIterator) length() int {
	return -1
}

type mappingIterator struct {
	sub iterator
	fnc func(val.Value) (val.Value, err.Error)
}

func newMappingIterator(i iterator, f func(val.Value) (val.Value, err.Error)) mappingIterator {
	return mappingIterator{i, f}
}

func (i mappingIterator) forEach(f func(val.Value) err.Error) err.Error {
	return i.sub.forEach(func(v val.Value) err.Error {
		m, e := i.fnc(v)
		if e != nil {
			return e
		}
		return f(m)
	})
}

func (i mappingIterator) length() int {
	return i.sub.length()
}

type filterIterator struct {
	sub iterator
	fnc func(val.Value) (bool, err.Error)
}

func newFilterIterator(sub iterator, f func(val.Value) (bool, err.Error)) filterIterator {
	return filterIterator{sub, f}
}

func (i filterIterator) forEach(f func(val.Value) err.Error) err.Error {
	return i.sub.forEach(func(v val.Value) err.Error {
		m, e := i.fnc(v)
		if e != nil {
			return e
		}
		if m {
			return f(v)
		}
		return nil
	})
}

func (i filterIterator) length() int {
	return -1
}

// bucketRefIterator yields val.Refs to the elements in a bucket
type bucketRefIterator struct {
	mid    string
	bucket *bolt.Bucket
}

func newBucketRefIterator(mid string, bucket *bolt.Bucket) bucketRefIterator {
	return bucketRefIterator{mid, bucket}
}

func (i bucketRefIterator) forEach(f func(val.Value) err.Error) err.Error {
	c := i.bucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		if e := f(val.Ref{i.mid, string(k)}); e != nil {
			return e
		}
	}
	return nil
}

func (i bucketRefIterator) length() int {
	return i.bucket.Stats().KeyN
}
