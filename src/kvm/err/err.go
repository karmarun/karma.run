package err

import (
	"kvm/val"
)

type Error interface {
	Value() val.Union // serializable
	Error() string    // should be proxy to String() (to implement error interface)
	String() string   // human readable string
	Child() Error     // may be nil
}
