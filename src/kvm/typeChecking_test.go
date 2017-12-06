package kvm

import (
	"kvm/mdl"
	"testing"
)

func TestCheckType(t *testing.T) {
	{
		actual := &mdl.Recursion{Label: "actual"}
		actual.Model = mdl.Struct{"test": actual}

		expected := &mdl.Recursion{Label: "expected"}
		expected.Model = mdl.Struct{"test": expected}

		if e := checkType(actual, expected); e != nil {
			t.Fatalf("case 1: %v", e)
		}
	}
	{
		actual := &mdl.Recursion{Label: "actual"}
		actual.Model = mdl.Struct{"test": mdl.Struct{"test": actual}}

		expected := &mdl.Recursion{Label: "expected"}
		expected.Model = mdl.Struct{"test": expected}

		if e := checkType(actual, expected); e != nil {
			t.Fatalf("case 2: %v", e)
		}
	}
	{
		actual := &mdl.Recursion{Label: "actual"}
		actual.Model = mdl.Struct{"test": actual}

		expected := &mdl.Recursion{Label: "expected"}
		expected.Model = mdl.Struct{"test": mdl.Struct{"test": expected}}

		if e := checkType(actual, expected); e == nil {
			t.Fatal("case 3: expected type checking error")
		}
	}
	{
		actual := &mdl.Recursion{Label: "actual"}
		actual.Model = mdl.Struct{"test": actual}

		expected := mdl.Any{}

		if e := checkType(actual, expected); e != nil {
			t.Fatal("case 4: %v", e)
		}
	}
}
