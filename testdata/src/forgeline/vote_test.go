package forgeline

import (
	shed "lib"
)

// castOne is the first of two test-only uses of the deviating spelling. They
// exist to prove the exclusion is from the AGREEMENT and not merely from the
// reporting: together with forged.go they make `shed` the majority spelling of
// every file the driver loaded, so a tally that counted test files would move
// the norm off `lib` and start reporting first.go and second.go instead.
func castOne() string { return shed.Name }
