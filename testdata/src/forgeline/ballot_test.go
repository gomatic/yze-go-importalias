package forgeline

import (
	shed "lib"
)

// castTwo is the second such use. Two are needed rather than one, because the
// norm is decided by a strict majority and `lib` wins a tie by being the name
// the package binds for itself.
func castTwo() string { return shed.Name }
