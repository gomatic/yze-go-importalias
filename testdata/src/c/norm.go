// Package c establishes `lib` as the spelling for other/lib, then imports both
// packages together in one file — where taking that spelling is impossible.
package c

import (
	lib "other/lib"
)

// Norm sets the module's spelling for other/lib.
func Norm() string { return lib.Other }
