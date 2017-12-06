// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

func mapExpressions(xs []Expression, f func(Expression) Expression) []Expression {
	out := make([]Expression, len(xs), len(xs))
	for i, x := range xs {
		out[i] = f(x)
	}
	return out
}

func mapExpressionMap(xs map[string]Expression, f func(Expression) Expression) map[string]Expression {
	out := make(map[string]Expression, len(xs))
	for k, x := range xs {
		out[k] = f(x)
	}
	return out
}
