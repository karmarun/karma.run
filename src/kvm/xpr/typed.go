package xpr

import (
	"kvm/mdl"
)

type TypedExpression struct {
	Expression
	Expected mdl.Model
	Actual   mdl.Model
}

func (x TypedExpression) Transform(f func(Expression) Expression) Expression {
	return f(TypedExpression{f(x.Expression), x.Expected, x.Actual})
}
