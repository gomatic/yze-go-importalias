package c

import (
	"lib"
	otherlib "other/lib"
)

// Forced must spell other/lib differently: `lib` is already taken here by the
// other package, so demanding the norm would demand code that does not compile.
// Nothing is reported.
func Forced() string { return lib.Name + otherlib.Other }
