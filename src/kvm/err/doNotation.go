// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"github.com/karmarun/karma.run/kvm/val"
)

type DoNotationError struct {
	Problem string
}

func (e DoNotationError) Value() val.Union {
	return val.Union{"doNotationError", val.String(e.Problem)}
}
func (e DoNotationError) Error() string {
	return e.String()
}
func (e DoNotationError) String() string {
	out := "Do Notation Error\n"
	out += "=================\n"
	out += "Problem\n"
	out += "-------\n"
	out += e.Problem + "\n\n"
	return out
}
func (e DoNotationError) Child() Error {
	return nil
}
