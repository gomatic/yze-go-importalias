package e

import (
	"lib"
	lib2 "other/lib"
)

// Justified keeps its number: both packages are called lib, so the collision
// is real and the digit is doing work.
func Justified() string { return lib.Name + lib2.Other }
