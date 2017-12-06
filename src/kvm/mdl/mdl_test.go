package mdl

import (
	"testing"
)

func TestEitherOr(t *testing.T) {
	{
		l := Int64{}
		r := String{}
		actual := Either(l, r, nil)
		expect := Or{Int64{}, String{}}
		if actual != expect {
			t.Fatalf("case 1: %#v\n", actual)
		}
	}
	{
		l := Int64{}
		r := Or{String{}, Float{}}
		actual := Either(l, r, nil)
		expect := Or{Int64{}, Or{String{}, Float{}}}
		if actual != expect {
			t.Fatalf("case 2: %#v\n", actual)
		}
	}
	{
		l := Or{String{}, Float{}}
		r := Int64{}
		actual := Either(l, r, nil)
		expect := Or{String{}, Or{Float{}, Int64{}}}
		if actual != expect {
			t.Fatalf("case 3: %#v\n", actual)
		}
	}
	{
		l := Or{String{}, Float{}}
		r := Or{Float{}, Int64{}}
		actual := Either(l, r, nil)
		expect := Or{String{}, Or{Float{}, Int64{}}}
		if actual != expect {
			t.Fatalf("case 4: %#v\n", actual)
		}
	}
	{
		l := &Recursion{Label: "l"}
		l.Model = List{l}
		r := &Recursion{Label: "r"}
		r.Model = List{r}
		actual, ok := Either(l, r, nil).(*Recursion)
		if !ok {
			t.Fatal("expected a *Recursion")
		}
		list, ok := actual.Model.(List)
		if !ok {
			t.Fatal("expected a List")
		}
		if list.Elements != actual {
			t.Fatalf("case 5: %#v\n", actual)
		}
	}
	{
		l := String{}
		r := Or{String{}, String{}}
		actual := Either(l, r, nil)
		expect := String{}
		if actual != expect {
			t.Fatalf("case 6: %#v\n", actual)
		}
	}
	{
		l := Union{"left": Struct{}}
		r := Union{"right": Struct{}}
		actual := Either(l, r, nil)
		expect := Union{"left": Struct{}, "right": Struct{}}
		if !actual.Equals(expect) {
			t.Fatalf("case 7: %#v\n", actual)
		}
	}
}
