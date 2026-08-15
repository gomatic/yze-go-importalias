package importalias_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	importalias "github.com/gomatic/yze-go-importalias"
)

// TestOneImportOneSpelling pins the rule in both directions: a package spelled
// two ways across a package's files is reported once, at the deviating spelling
// — and a deviation forced by a name collision is not, because demanding the
// module's spelling there would demand code that does not compile.
func TestOneImportOneSpelling(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "a")
}

// TestACollisionJustifiesTheDeviation pins the guard that keeps this rule
// honest: package c settles on `lib` for other/lib, but one file imports the
// other lib too, so that name is already taken there. The deviation is forced
// by the language, and a rule that reported it would be demanding code that
// does not compile.
func TestACollisionJustifiesTheDeviation(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "c")
}

// TestAVersionSuffixedPathIsNamedByItsPackage pins the false positive the
// obvious implementation would produce: "versioned/cli/v3" binds `cli`, not
// `v3`, so a plain import and an explicit `cli` alias agree. Reading the name
// from the path's last element would report a file for already matching the
// norm.
func TestAVersionSuffixedPathIsNamedByItsPackage(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "b")
}

// TestALeftoverNumberIsReportedOnlyWhenNothingJustifiesIt pins the refactor
// residue both ways: `lib2` with no other `lib` in the file is debris from a
// move and is reported, while the same alias beside a genuine second `lib` is
// earning its number and is silent.
func TestALeftoverNumberIsReportedOnlyWhenNothingJustifiesIt(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "d", "e")
}

// TestRegistrationIsWellFormed pins the yze wiring.
func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, importalias.Registration.Validate())
	assert.Equal(t, "yze/importalias", importalias.Registration.RuleID())
	assert.Same(t, importalias.Analyzer, importalias.Registration.Analyzer)
}

// TestADirectiveDoesNotRenameAFile pins the test-file exclusion to the name the
// FileSet holds, which the judged file cannot rewrite. A `//line` directive is a
// compiler feature for generated code: it changes what fset.Position reports and
// nothing else — the go tool still compiles the file, and go list still names it
// in GoFiles. So an exclusion resolved through Position is decided by the very
// file it is excluding, and this one is doubly load-bearing: an excluded file is
// out of the AGREEMENT as well as out of the reporting, so a forged name both
// hides a deviation and removes its vote.
//
// Package forgeline settles on the plain spelling twice over. forged.go is
// ordinary source claiming a test name and is reported anyway; forgeline_test.go
// is a real test file claiming a source name and is spared anyway.
//
// The rest of the package sits on the matcher's literal, because the identity
// is only as good as the comparison that reads it and every widening below
// survived this suite before its fixture existed. Kit's file name contains
// "_test" and does not end in "_test.go"; Helper's ends in "test.go" with no
// underscore — the left edge, and the shape of net/http/httptest/httptest.go;
// Golden's contains "_test.go" without ending in it; Cased's differs from
// "_test.go" only in letter case. All four are in GoFiles, so all four are bound
// by the agreement and reported, and each spells the import its own way so that
// adding them cannot move the norm.
func TestADirectiveDoesNotRenameAFile(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "forgeline")
}

// TestAForgedFileVotesAsWhatTheGoToolCompiled inverts the dimension forgeline
// holds constant. There the honest spelling is the majority, so no directive can
// reach the norm; here two directive-carrying files outnumber the plain import,
// which is the only region where a forged vote is observable at all.
//
// The verdict is the one the compiled identity dictates: all three files vote,
// `shed` wins, and the plain import is the deviation — identical, byte for byte,
// to the same three files with the directives deleted, which is what makes the
// directives worthless to their authors. Resolving the identity through Position
// instead drops both forged files from the tally and reports NOTHING, and so
// would any rule that let a forged file be judged but not counted: the norm it
// would be judged against would cease to exist.
//
// Whether a bare plurality is the right norm is a separate, directive-free
// question, open as importalias.norm-is-not-a-bare-vote.
func TestAForgedFileVotesAsWhatTheGoToolCompiled(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "forgevote")
}
