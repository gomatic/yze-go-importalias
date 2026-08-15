// Package forgevote inverts the one dimension package forgeline holds constant:
// there, the honest files are the majority and the file carrying a `//line`
// directive is outvoted, so no directive can ever move the norm. Here the two
// directive-carrying files ARE the majority, which is the only region where the
// question "does a forged file get a vote?" has an observable answer.
//
// The answer this pins is that it gets exactly the vote its COMPILED identity
// earns it, and nothing more or less. All three files are in GoFiles, all three
// count, `shed` is the spelling written most often, and the plain import is the
// deviation — the same verdict, byte for byte, that the same three files produce
// with the directives deleted. That equality was measured, and it is the point:
// the directives buy their authors nothing at all.
//
// It is also what makes this case discriminate where forgeline cannot. Resolving
// the identity through Position drops both forged files from the tally, leaving
// one spelling, no disagreement and TOTAL SILENCE — a package with a live
// inconsistency reported nowhere. So would any rule that excluded a forged file
// from the vote while still judging it: the norm it would be judged against
// would cease to exist. That is the trap in "a forged file may not vote", and
// this case is what fails when someone tries it.
//
// The remedy is one the reported author can actually take, which is the
// question docs/s04.md item 8 asks of every finding. It is not "find two
// comments nothing reports": the directives are irrelevant to the verdict, the
// deviation named is a spelling, and the author either adopts `shed` here or
// respells the two files that outvoted them — both ordinary edits to ordinary
// source, neither of them blocked by anything the forgers wrote.
//
// What is NOT settled here is whether a bare plurality is the right norm at all.
// dominant() already refuses to call a plain import the deviation at a TIE, on
// the ground that doing so would be perverse; nothing explains why two-to-one is
// different in kind. That is a question about the rule, not about directives,
// and it is open as importalias.norm-is-not-a-bare-vote. Should it be answered
// the other way, this fixture's want moves to the two aliasing files and the
// directive assertion above is unaffected.
package forgevote

import (
	"lib" // want `import "lib" is spelled lib here but shed elsewhere`
)

// Honest imports the package the way the language does, and is the file the
// rule holds to account.
func Honest() string { return lib.Name }
