//line zz_test.go:1
package forgeline

import (
	shed "lib" // want `import "lib" is spelled shed here but lib elsewhere`
)

// Forged is ordinary compiled source — go list reports forged.go in GoFiles —
// and the directive above is the only thing claiming a test name for it. Test
// files are excluded from the agreement as well as from the reporting, so
// reading the claimed name would drop this deviation out of the rule entirely
// AND let it stop counting toward the norm the other files must meet.
func Forged() string { return shed.Name }
