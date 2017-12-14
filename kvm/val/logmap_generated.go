// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package val

// This file is auto-generated using go generate. DO NOT EDIT!
// logMapStringValue is a garbage-collector friendly map implementation with O(log n) operations.

import(
    "sort"
)

type logMapStringValue struct {
    _keys []string
    _vals []Value
}

func newlogMapStringValue(initialCapacity int) logMapStringValue {
    return logMapStringValue{
        _keys: make([]string, 0, initialCapacity),
        _vals: make([]Value, 0, initialCapacity),
    }
}

func (m logMapStringValue) sameKeys(w logMapStringValue) bool {
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

func (m logMapStringValue) equals(w logMapStringValue) bool {
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

func (m logMapStringValue) keys() []string {
    if m._keys == nil {
        return nil
    }
    keys := make([]string, len(m._keys), len(m._keys))
    copy(keys, m._keys)
    return keys
}

func (m logMapStringValue) values() []Value {
    if m._vals == nil {
        return nil
    }
    values := make([]Value, len(m._vals), len(m._vals))
    copy(values, m._vals)
    return values
}

func (m logMapStringValue) overMap(f func(k string, v Value) Value) {
    for i, k := range m._keys {
        m._vals[i] = f(k, m._vals[i])
    }
}

func (m logMapStringValue) get(k string) (Value, bool) {
    i := m.search(k)
    if i == len(m._keys) || m._keys[i] != k {
        return nil, false
    }
    return m._vals[i], true
}

func (m *logMapStringValue) set(k string, v Value) {
    i := m.search(k)
    m._keys, m._vals = append(m._keys, k), append(m._vals, v)
    copy(m._keys[i+1:], m._keys[i:])
    copy(m._vals[i+1:], m._vals[i:])
    m._keys[i], m._vals[i] = k, v
}

func (m *logMapStringValue) unset(k string) {
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


func (m logMapStringValue) copyFunc(f func(Value) Value) logMapStringValue {
    if m._keys == nil {
        return logMapStringValue{}
    }
    keys := make([]string, len(m._keys), cap(m._keys))
    copy(keys, m._keys)
    values := make([]Value, len(m._vals), cap(m._vals))
    for i, v := range m._vals {
        values[i] = f(v)
    }
    return logMapStringValue{keys, values}
}

func (m logMapStringValue) forEach(f func(string, Value) bool) {
    for i, k := range m._keys {
        if !f(k, m._vals[i]) {
            break
        }
    }
}

func (m logMapStringValue) len() int {
    return len(m._keys)
}

func (m logMapStringValue) search(k string) int {
    // binary search, returns index at which to insert k
    return sort.Search(len(m._keys), func(i int) bool {
        return m._keys[i] >= k
    })
}
