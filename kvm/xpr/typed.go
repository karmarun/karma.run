// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package xpr

import (
	"karma.run/kvm/mdl"
)

type TypedExpression struct {
	Expression
	Expected mdl.Model
	Actual   mdl.Model
}

func (x TypedExpression) Transform(f func(Expression) Expression) Expression {
	return f(TypedExpression{f(x.Expression), x.Expected, x.Actual})
}
