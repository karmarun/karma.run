
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
}

func newlogMapStringModel(initialCapacity int) logMapStringModel {
    return logMapStringModel{
        _keys:   make([]string, 0, initialCapacity),
        _vals: make([]Model, 0, initialCapacity),
    }
}

func (m logMapStringModel) sameKeys(w logMapStringModel) bool {
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

func (m logMapStringModel) equals(w logMapStringModel) bool {
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

func (m logMapStringModel) keys() []string {
    if m._keys == nil {
        return nil
    }
    keys := make([]string, len(m._keys), len(m._keys))
    copy(keys, m._keys)
    return keys
}

func (m logMapStringModel) values() []Model {
    if m._vals == nil {
        return nil
    }
    values := make([]Model, len(m._vals), len(m._vals))
    copy(values, m._vals)
    return values
}

func (m logMapStringModel) overMap(f func(k string, v Model) Model) {
    for i, k := range m._keys {
        m._vals[i] = f(k, m._vals[i])
    }
}

func (m logMapStringModel) get(k string) (Model, bool) {
    i := m.search(k)
    if i == len(m._keys) || m._keys[i] != k {
        return nil, false
    }
    return m._vals[i], true
}

func (m *logMapStringModel) set(k string, v Model) {
    i := m.search(k)
    m._keys, m._vals = append(m._keys, k), append(m._vals, v)
    copy(m._keys[i+1:], m._keys[i:])
    copy(m._vals[i+1:], m._vals[i:])
    m._keys[i], m._vals[i] = k, v
}

func (m *logMapStringModel) unset(k string) {
    i := m.search(k)
    if i == len(m._keys) || m._keys[i] != k {
        return
    }
    l := len(m._keys)
    copy(m._keys[i:l-1], m._keys[i+1:])
    copy(m._vals[i:l-1], m._vals[i+1:])
    m._keys[l-1], m._vals[l-1] = "", nil // let them be GC'ed
    m._keys, m._vals = m._keys[:l-1], m._vals[:l-1]
}


func (m logMapStringModel) copyFunc(f func(Model) Model) logMapStringModel {
    if m._keys == nil {
        return logMapStringModel{}
    }
    keys := make([]string, len(m._keys), cap(m._keys))
    copy(keys, m._keys)
    values := make([]Model, len(m._vals), cap(m._vals))
    for i, v := range m._vals {
        values[i] = f(v)
    }
    return logMapStringModel{keys, values}
}

func (m logMapStringModel) forEach(f func(string, Model) bool) {
    for i, k := range m._keys {
        if !f(k, m._vals[i]) {
            break
        }
    }
}

func (m logMapStringModel) len() int {
    return len(m._keys)
}

func (m logMapStringModel) search(k string) int {
    // binary search, returns index at which to insert k
    return sort.Search(len(m._keys), func(i int) bool {
        return m._keys[i] >= k
    })
}
