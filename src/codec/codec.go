// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package codec

import (
	"github.com/karmarun/karma.run/kvm/err"
	"github.com/karmarun/karma.run/kvm/mdl"
	"github.com/karmarun/karma.run/kvm/val"
	"log"
)

type Instantiator func() Interface

type Interface interface {
	Decode([]byte, mdl.Model) (val.Value, err.Error)
	Encode(val.Value) []byte
}

// Not thread-safe
var registry = make(map[string]Instantiator)

func Register(key string, itr Instantiator) {
	if _, ok := registry[key]; ok {
		log.Panicf(`Codec already registered for key: %s`, key)
	}
	registry[key] = itr
}

func Available() []string {
	decs := make([]string, 0, len(registry))
	for k, _ := range registry {
		decs = append(decs, k)
	}
	return decs
}

func Get(key string) Interface {
	i := registry[key]
	if i == nil {
		return nil
	}
	return i()
}
