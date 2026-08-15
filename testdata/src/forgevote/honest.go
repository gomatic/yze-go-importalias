// Package forgevote inverts the one dimension package forgeline holds constant:
// there, the honest files are the majority and the file carrying a `//line`
// directive is outvoted, so no directive can ever move the norm. Here the two
// directive-carrying files ARE the majority, which is the only region where the
// question "does a forged file get a vote?" has an observable answer.
//
// The answer this pins is that it gets exactly the vote its COMPILED identity
// earns it, and nothing more or less. All three files are in GoFiles, all three
// count, `shed` is the spelling written most often, and the odd one out is the
// deviation — the same verdict, byte for byte, that these three files produce
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
// The odd one out is deliberately an ALIAS and not the plain import, so that
// nothing here asserts the contested half. Whether a bare plurality may report a
// PLAIN import as the deviation — which dominant() already calls perverse at a
// tie, with nothing explaining why two-to-one differs in kind — is open as
// importalias.norm-is-not-a-bare-vote, and an alias-versus-alias package
// discriminates identically while staying out of it.
//
// The remedy is one the reported author can take, which is what docs/s04.md item
// 8 asks: adopt `shed` here, or respell the two files that outvoted them. Both
// are ordinary edits to ordinary source, neither blocked by anything the forgers
// wrote.
package forgevote

import (
	hedge "lib" // want `import "lib" is spelled hedge here but shed elsewhere`
)

// Honest spells the import its own way and is the file the rule holds to
// account, because two other files spell it another way.
func Honest() string { return hedge.Name }
