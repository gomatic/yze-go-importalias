// Package a spells its imports inconsistently across files.
package a

import (
	"lib"
)

// First reads the plainly-imported package.
func First() string { return lib.Name }
