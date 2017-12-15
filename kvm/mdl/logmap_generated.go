// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

// This file is auto-generated using go generate. DO NOT EDIT!
// logMapStringModel is a garbage-collector friendly map implementation with O(log n) operations.

import(
    "sort"
)

type logMapStringModel struct {
    _keys []string
    _vals []Model
    _sharingKeys bool
    _sharingVals bool
}

var (
    zeroKey = *new(string)
    zeroValue = *new(Model)
)

func newlogMapStringModel(initialCapacity int) *logMapStringModel {
    return &logMapStringModel{
        _keys: make([]string, 0, initialCapacity),
        _vals: make([]Model, 0, initialCapacity),
        _sharingKeys: false,
        _sharingVals: false,
    }
}

func (m *logMapStringModel) sameKeys(w *logMapStringModel) bool {
    if m == nil {
        return w == nil || len(w._keys) == 0
    }
    if len(m._keys) != len(w._keys) {
        return false
    }
    for i, l := 0, len(m._keys); i < l; i++ {
        if m._keys[i] != w._keys[i] {
            return false
        }
    }
    return true
}

func (m *logMapStringModel) equals(w *logMapStringModel) bool {
    if m == nil {
        return w == nil
    }
    if m == w {
        return true
    }
    if !m.sameKeys(w) {
        return false
    }
    for i, l := 0, len(m._keys); i < l; i++ {
        if m._vals[i] != w._vals[i] {
            return false
        }
    }
    return true
}

func (m *logMapStringModel) keys() []string {
    if m == nil {
        return nil
    }
    keys := make([]string, len(m._keys), cap(m._keys))
    copy(keys, m._keys)
    return keys
}

func (m *logMapStringModel) values() []Model {
    if m == nil || m._vals == nil {
        return nil
    }
    values := make([]Model, len(m._vals), cap(m._vals))
    copy(values, m._vals)
    return values
}

func (m *logMapStringModel) overMap(f func(k string, v Model) Model) {
    if m == nil {
        return
    }
    if m._sharingVals {
        m._vals, m._sharingVals = m.values(), false
    }
    for i, k := range m._keys {
        m._vals[i] = f(k, m._vals[i])
    }
}

func (m *logMapStringModel) get(k string) (Model, bool) {
    if m == nil {
        return zeroValue, false
    }
    i := m.search(k)
    if i == len(m._keys) || m._keys[i] != k {
        return zeroValue, false
    }
    return m._vals[i], true
}

func (m *logMapStringModel) set(k string, v Model) {
    if m == nil {
        panic("set of nil")
    }
    if m._sharingKeys {
        m._keys, m._sharingKeys = m.keys(), false
    }
    if m._sharingVals {
        m._vals, m._sharingVals = m.values(), false
    }
    i := m.search(k)
    m._keys, m._vals = append(m._keys, zeroKey), append(m._vals, zeroValue)
    copy(m._keys[i+1:], m._keys[i:])
    copy(m._vals[i+1:], m._vals[i:])
    m._keys[i], m._vals[i] = k, v
}

func (m *logMapStringModel) unset(k string) {
    if m == nil {
        return
    }
    i := m.search(k)
    if i == len(m._keys) || m._keys[i] != k {
        return
    }
    if m._sharingKeys {
        m._keys, m._sharingKeys = m.keys(), false
    }
    if m._sharingVals {
        m._vals, m._sharingVals = m.values(), false
    }
    l := len(m._keys)
    copy(m._keys[i:l-1], m._keys[i+1:])
    copy(m._vals[i:l-1], m._vals[i+1:])
    m._keys[l-1], m._vals[l-1] = zeroKey, zeroValue // let them be GC'ed
    m._keys, m._vals = m._keys[:l-1], m._vals[:l-1]
}


func (m *logMapStringModel) copy() *logMapStringModel {
    if m == nil {
        return m
    }
    m._sharingKeys, m._sharingVals = true, true
    return m
}

func (m *logMapStringModel) forEach(f func(string, Model) bool) {
    if m == nil {
        return
    }
    for i, k := range m._keys {
        if !f(k, m._vals[i]) {
            break
        }
    }
}

func (m *logMapStringModel) len() int {
    if m == nil {
        return 0
    }
    return len(m._keys)
}

func (m *logMapStringModel) search(k string) int {
    if m == nil {
        return -1
    }
    // binary search, returns index at which to insert k
    return sort.Search(len(m._keys), func(i int) bool {
        return m._keys[i] >= k
    })
}
