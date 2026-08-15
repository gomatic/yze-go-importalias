// Package forgeline settles on the plain spelling in two ordinary files, so the
// norm is not in doubt, and then asks who gets held to it.
package forgeline

import (
	"lib"
)

// First reads the plainly-imported package.
func First() string { return lib.Name }
