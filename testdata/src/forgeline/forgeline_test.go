//line nottest.go:1
package forgeline

import (
	shelf "lib"
)

// helper is the other direction: a real test file that a directive tells the
// position machinery to call nottest.go. A test may name an import for local
// clarity, so it is out of the agreement and nothing is reported here — and its
// spelling does not drag the norm the source files settled on.
func helper() string { return shelf.Name }
