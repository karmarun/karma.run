package xpr

import (
	"kvm/err"
)

// transforms bind/do into with/return
func EliminateDoNotation(x Expression) (Expression, err.Error) {

	e := (err.Error)(nil)

	needsWork := false

	x = x.Transform(func(x Expression) Expression {
		if e != nil {
			return x
		}
		if do, ok := x.(Do); ok {
			x, e = eliminateDoInstance(do)
			needsWork = true
		}
		if _, ok := x.(Bind); ok {
			needsWork = true
		}
		return x
	})

	if e != nil {
		return nil, e
	}

	if !needsWork {
		return x, nil // avoid traversing expression tree again
	}

	x.Transform(func(x Expression) Expression {
		if e != nil {
			return x
		}
		if _, ok := x.(Do); ok {
			panic("uneliminated do (should not be possible)")
		}
		if _, ok := x.(Bind); ok {
			e = err.DoNotationError{"bind used outside of do block"}
		}
		return x
	})

	if e != nil {
		return nil, e
	}

	return x, nil

}

func eliminateDoInstance(do Do) (Expression, err.Error) {

	if _, ok := do["return"]; !ok {
		return nil, err.DoNotationError{`do missing "return" binding`}
	}

	// binding string -> depends on binding keys
	dependencies := make(map[string]map[string]struct{}, len(do))

	for name, body := range do {

		deps := make(map[string]struct{}, 8)

		body.Transform(func(x Expression) Expression {
			if bind, ok := x.(Bind); ok {
				deps[string(bind)] = struct{}{}
			}
			return x
		})

		dependencies[name] = deps

	}

	order := sortDependencies(dependencies, nil)
	if order == nil {
		return nil, err.DoNotationError{`cyclic bindings in do block`}
	}

	for _, name := range order {
		do[name] = do[name].Transform(func(x Expression) Expression {
			if bind, ok := x.(Bind); ok {
				return do[string(bind)]
			}
			return x
		})
	}

	return do["return"], nil

}

// return value of nil means there is no order
func sortDependencies(todo map[string]map[string]struct{}, done map[string]struct{}) []string {

	if done == nil {
		done = make(map[string]struct{}, len(todo))
	}

	order := make([]string, 0, len(todo))

	for len(todo) > 0 {
		progress := false
		for name, deps := range todo {
			if allLeftKeysInRightKeys(deps, done) {
				order = append(order, name)
				delete(todo, name)
				done[name] = struct{}{}
				progress = true
			}
		}
		if !progress {
			return nil
		}
	}

	return order

}

func allLeftKeysInRightKeys(left, right map[string]struct{}) bool {
	for k, _ := range left {
		if _, ok := right[k]; !ok {
			return false
		}
	}
	return true
}

// func transformDoNotation(n ast.Node, bindings map[string]ast.Node) (ast.Node, err.Error) {
// 	if bindings == nil {
// 		bindings = make(map[string]ast.Node)
// 	}
// 	e := (err.Error)(nil)
// 	n = ast.Map(n, func(n ast.Node) ast.Node {
// 		if e != nil {
// 			return n // short-circuit on error
// 		}
// 		if bind, ok := n.(ast.Bind); ok {
// 			if rt, ok := bindings[string(bind)]; ok {
// 				return rt
// 			}
// 			e = err.CompilationError{
// 				Problem: fmt.Sprintf(`do/bind: undefined binding: %s`, string(bind)),
// 			}
// 			return n
// 		}
// 		if do, ok := n.(ast.Do); ok {
// 			rt, ok := do["return"]
// 			if !ok {
// 				e = err.CompilationError{
// 					Problem: `do/bind: missing "return" binding`,
// 				}
// 				return n
// 			}
// 			delete(do, "return")
// 			for k, sub := range do {
// 				bindings[k] = sub
// 			}
// 			n, e = transformDoNotation(rt, bindings)
// 			return n
// 		}
// 		return n
// 	})
// 	return n, e
// }
