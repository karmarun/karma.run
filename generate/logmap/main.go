// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"text/template"
)

func main() {
	pkg, key, value, out := "", "", "", ""
	flag.StringVar(&pkg, "package", "", "package name")
	flag.StringVar(&key, "key", "", "key type")
	flag.StringVar(&value, "value", "", "value type")
	flag.StringVar(&out, "output", "", "output file")
	flag.Parse()
	if pkg == "" {
		log.Fatalln("missing --package")
	}
	if key == "" {
		log.Fatalln("missing --key")
	}
	if value == "" {
		log.Fatalln("missing --value")
	}
	if out == "" {
		out = "/dev/stdout"
	}
	t, e := template.New("").Parse(templateString)
	if e != nil {
		log.Fatalln(e)
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if e != nil {
		log.Fatalln(e)
	}
	defer f.Close()
	e = t.Execute(f, map[string]interface{}{
		"package": pkg,
		"key":     key,
		"value":   value,
		"type":    "logMap" + strings.Title(key) + strings.Title(value),
	})
	if e != nil {
		log.Fatalln(e)
	}
}

const templateString = `// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package {{.package}}

// This file is auto-generated using go generate. DO NOT EDIT!
// {{.type}} is a garbage-collector friendly map implementation with O(log n) operations.

import(
    "sort"
)

type {{.type}} struct {
    _keys []{{.key}}
    _vals []{{.value}}
    _sharingKeys bool
    _sharingVals bool
}

var (
    zeroKey = *new({{.key}})
    zeroValue = *new({{.value}})
)

func new{{.type}}(initialCapacity int) *{{.type}} {
    return &{{.type}}{
        _keys: make([]{{.key}}, 0, initialCapacity),
        _vals: make([]{{.value}}, 0, initialCapacity),
        _sharingKeys: false,
        _sharingVals: false,
    }
}

func (m *{{.type}}) sameKeys(w *{{.type}}) bool {
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

func (m *{{.type}}) equals(w *{{.type}}) bool {
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

func (m *{{.type}}) keys() []{{.key}} {
    if m == nil {
        return nil
    }
    keys := make([]{{.key}}, len(m._keys), cap(m._keys))
    copy(keys, m._keys)
    return keys
}

func (m *{{.type}}) values() []{{.value}} {
    if m == nil || m._vals == nil {
        return nil
    }
    values := make([]{{.value}}, len(m._vals), cap(m._vals))
    copy(values, m._vals)
    return values
}

func (m *{{.type}}) overMap(f func(k {{.key}}, v {{.value}}) {{.value}}) {
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

func (m *{{.type}}) get(k {{.key}}) ({{.value}}, bool) {
    if m == nil {
        return zeroValue, false
    }
    i := m.search(k)
    if i == len(m._keys) || m._keys[i] != k {
        return zeroValue, false
    }
    return m._vals[i], true
}

func (m *{{.type}}) set(k {{.key}}, v {{.value}}) {
    if m == nil {
        panic("set of nil")
    }
    if m._sharingKeys {
        m._keys, m._sharingKeys = m.keys(), false
    }
    i := m.search(k)
    m._keys, m._vals = append(m._keys, k), append(m._vals, v)
    if i == len(m._keys)-1 {
        return
    }
    if m._sharingVals {
        m._vals, m._sharingVals = m.values(), false
    }
    copy(m._keys[i+1:], m._keys[i:])
    copy(m._vals[i+1:], m._vals[i:])
    m._keys[i], m._vals[i] = k, v
}

func (m *{{.type}}) unset(k {{.key}}) {
    if m == nil {
        return
    }
    l := len(m._keys)
    i := m.search(k)
    if i == l || m._keys[i] != k {
        return
    }
    if i == l-1 {
        m._keys[l-1], m._vals[l-1] = zeroKey, zeroValue // let them be GC'ed
        m._keys, m._vals = m._keys[:l-1], m._vals[:l-1]
        return
    }
    if m._sharingKeys {
        m._keys, m._sharingKeys = m.keys(), false
    }
    if m._sharingVals {
        m._vals, m._sharingVals = m.values(), false
    }
    copy(m._keys[i:l-1], m._keys[i+1:])
    copy(m._vals[i:l-1], m._vals[i+1:])
    m._keys[l-1], m._vals[l-1] = zeroKey, zeroValue // let them be GC'ed
    m._keys, m._vals = m._keys[:l-1], m._vals[:l-1]
}


func (m *{{.type}}) copy() *{{.type}} {
    if m == nil {
        return m
    }
    m._sharingKeys, m._sharingVals = true, true
    return m
}

func (m *{{.type}}) forEach(f func({{.key}}, {{.value}}) bool) {
    if m == nil {
        return
    }
    for i, k := range m._keys {
        if !f(k, m._vals[i]) {
            break
        }
    }
}

func (m *{{.type}}) len() int {
    if m == nil {
        return 0
    }
    return len(m._keys)
}

func (m *{{.type}}) search(k {{.key}}) int {
    if m == nil {
        return -1
    }
    // binary search, returns index at which to insert k
    return sort.Search(len(m._keys), func(i int) bool {
        return m._keys[i] >= k
    })
}
`
