package err

import (
	"kvm/val"
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
