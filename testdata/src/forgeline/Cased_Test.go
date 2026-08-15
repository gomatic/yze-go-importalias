package forgeline

import (
	sheaf "lib" // want `import "lib" is spelled sheaf here but lib elsewhere`
)

// Cased differs from a test file's name only in letter case. The go tool's own
// check is `strings.HasSuffix(name, "_test.go")` and is case-sensitive, so this
// is ordinary compiled source — verified on a case-INSENSITIVE darwin
// filesystem, where `go list` still reports it in GoFiles and never in
// TestGoFiles — and it is bound by the agreement.
//
// Case is the third dimension this package's names would otherwise hold
// constant. Folding the name before matching is the ordinary instinct of anyone
// who has been bitten by a Windows or macOS path.
//
// Each boundary fixture spells the import its own way rather than sharing one
// deviant alias, so that adding them cannot move the norm the package settled
// on: lib keeps a plurality of two against four spellings used once each.
func Cased() string { return sheaf.Name }
