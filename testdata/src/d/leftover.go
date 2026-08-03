// Package d carries numbered aliases: one left over from a move, one still
// earning its number.
package d

import (
	lib2 "lib" // want `is aliased lib2, but nothing here is called lib`
)

// Leftover reads through a number nothing justifies.
func Leftover() string { return lib2.Name }
