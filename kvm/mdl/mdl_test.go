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
			t.Fatalf("%#v\n", out)
		}
	}
	{
		a := Tuple{String{}, Float{}, Int16{}, Any{}}
		b := Tuple{Int32{}, Float{}, Any{}}
		expect := Tuple{any, Float{}, any}
		out := Either(a, b, nil)
		if !out.Equals(expect) {
			t.Fatalf("%#v\n", out)
		}
	}
	{
		a := Int64{}
		b := Optional{Int64{}}
		expect := b
		out := Either(a, b, nil)
		if !out.Equals(expect) {
			t.Fatalf("%#v\n", out)
		}
	}
	{
		a := List{Int64{}}
		b := List{Float{}}
		expect := List{any}
		out := Either(a, b, nil)
		if !out.Equals(expect) {
			t.Fatalf("%#v\n", out)
		}
	}
}
