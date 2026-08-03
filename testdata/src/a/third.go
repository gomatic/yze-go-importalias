package a

import (
	"lib"
	otherlib "other/lib"
)

// Third imports two packages whose names collide, so aliasing one is forced —
// the deviation is justified and nothing is reported.
func Third() string { return lib.Name + otherlib.Other }
