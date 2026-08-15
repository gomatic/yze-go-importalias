//line vote2_test.go:1
package forgevote

import (
	shed "lib"
)

// ForgedTwo is the second of the pair. Two are needed rather than one: a single
// deviating file cannot outvote the plain import, and a tie goes to the name the
// package declares for itself, so only a strict majority of forged files puts
// the question — can a file that claims to be a test move the norm the source
// files are held to? — anywhere it can be observed.
func ForgedTwo() string { return shed.Name }
