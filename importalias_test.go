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
func TestADirectiveDoesNotRenameAFile(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), importalias.Analyzer, "forgeline")
}
