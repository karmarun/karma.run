// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package err

import (
	"karma.run/kvm/val"
)

type Error interface {
	Value() val.Union // serializable
	Error() string    // should be proxy to String() (to implement error interface)
	String() string   // human readable string
	Child() Error     // may be nil
}
