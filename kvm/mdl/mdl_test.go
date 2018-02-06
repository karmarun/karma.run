// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package mdl

import (
	"testing"
)

func TestEither(t *testing.T) {
	{
		a := StructFromMap(map[string]Model{
			"q": String{},
			"p": Int32{},
			"t": Float{},
			"z": any,
		})
		b := StructFromMap(map[string]Model{
			"p": Int32{},  // same key, same type
			"t": String{}, // same key, different type
			"w": any,
		})
		expect := StructFromMap(map[string]Model{
			"p": Int32{}, // same key, same type
			"t": any,
		})
		out := Either(a, b, nil)
		if !out.Equals(expect) {
			t.Fatal("out != expect")
		}
	}
}
