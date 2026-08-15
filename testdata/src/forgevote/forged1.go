//line vote1_test.go:1
package forgevote

import (
	shed "lib"
)

// ForgedOne is ordinary compiled source — go list reports forged1.go in GoFiles
// — claiming a test name it does not have. Its vote counts because the file the
// go tool compiled is a source file, and it is not reported because it agrees
// with the spelling its own vote helped settle on. Nothing here is a concession
// to the directive: delete the line above and this file behaves identically.
func ForgedOne() string { return shed.Name }
