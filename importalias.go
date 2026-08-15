// Package importalias provides a go/analysis analyzer enforcing one spelling
// per import across a module: if a package is aliased anywhere, it is aliased
// identically everywhere.
//
// Two files naming the same import differently is a maintainability defect and
// nothing else — a reader who has learned that `errs` means go-error has to
// re-learn it as `goerr` in the next file, and a grep for one spelling silently
// misses the other. The standard states it as an invariant, and it is a pure
// syntactic fact over a module's import specs, which is what makes it a gate
// rather than a judgment call.
//
// The one legitimate reason to differ is a COLLISION: a file importing two
// packages whose names would clash must alias one of them. A deviation is
// therefore reported only when adopting the module's own spelling would leave
// that file conflict-free — so the analyzer never asks for code that does not
// compile.
package importalias

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
)

const (
	message = "import %q is spelled %s here but %s elsewhere in this module; one import, one spelling"
	//nolint:lll // one sentence, and breaking it would split the explanation from the name it explains.
	messageNumbered = "import %q is aliased %s, but nothing here is called %s; the number is left over from a move — drop it"
)

// numbered matches an alias that is a package's own name followed by digits —
// the shape goimports and gopls generate to break a collision.
var numbered = regexp.MustCompile(`^[0-9]+$`)

// Analyzer reports imports spelled inconsistently across the module.
var Analyzer = &analysis.Analyzer{
	Name: "importalias",
	Doc: "reports an import path spelled with different aliases across a module, " +
		"except where a name collision in the file requires one",
	Run: run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "importalias",
	Categories: []goyze.Category{"structure"},
	URL:        "https://docs.gomatic.dev/yze/importalias",
	Analyzer:   Analyzer,
}

// run reports every import whose spelling differs from the module's own.
//
// Scope is the pass: a package at a time. That is deliberately narrower than
// "the module" — a divergence between two packages the driver analyses
// separately falls outside it — and it buys two things. Every file of a package
// is seen together, which is where a reader suffers the inconsistency; and the
// judgement rests on the files the driver loaded rather than on a directory
// walk's guess about which ones count.
func run(pass *analysis.Pass) (any, error) {
	source := sourceFiles(pass)
	norms := normsOf(pass.TypesInfo, source)
	for _, file := range source {
		reportFile(pass, file, norms)
	}
	return nil, nil
}

// sourceFiles is the pass's non-test files.
//
// A test file may alias an import for local clarity, and holding it to the
// source's spelling pits two standards against each other: the cliapp layout
// MANDATES the alias `domain` in a command's source, while the test beside it
// reasonably calls the same package `creation` to say which one it is driving.
// Test files are excluded from the agreement itself, rather than only from the
// reporting: a spelling only a test uses would otherwise drag the norm away
// from what the source files settled on.
//
// The name comes from the FileSet's own entry, never from a Position: Position
// applies //line directives, so a file could rename itself out of the agreement
// with one comment line while the go tool went on compiling it as ordinary
// source — hiding its own deviation and withdrawing its vote from the norm the
// rest of the package is held to. A decision ABOUT a file must read something
// that file cannot rewrite.
func sourceFiles(pass *analysis.Pass) []*ast.File {
	kept := make([]*ast.File, 0, len(pass.Files))
	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			kept = append(kept, file)
		}
	}
	return kept
}

// norms is the agreed spelling for each import path, absent when a path is
// spelled only one way (nothing to disagree with) or has no majority.
type norms map[importPath]spelling

// tally counts how often each spelling is used for one path, alongside the
// name the package declares for itself — the spelling a plain import binds.
type tally struct {
	counts  map[spelling]int
	natural spelling
}

// normsOf resolves each import path's spelling across the files, keeping only
// the paths that are spelled more than one way.
func normsOf(info *types.Info, files []*ast.File) norms {
	counts := map[importPath]tally{}
	for _, file := range files {
		for _, spec := range file.Imports {
			countSpelling(info, counts, spec)
		}
	}
	resolved := norms{}
	for at, counted := range counts {
		if len(counted.counts) > 1 {
			resolved[at] = dominant(counted)
		}
	}
	return resolved
}

// countSpelling folds one import spec into the tally.
func countSpelling(info *types.Info, counts map[importPath]tally, spec *ast.ImportSpec) {
	at, ok := pathOf(spec)
	if !ok {
		return
	}
	counted, seen := counts[at]
	if !seen {
		counted = tally{counts: map[spelling]int{}}
	}
	counted.counts[spellingOf(info, spec)]++
	if natural := naturalOf(info, spec); natural != "" {
		counted.natural = natural
	}
	counts[at] = counted
}

// reportFile reports each of the file's imports that deviates from the norm
// without a collision to justify it.
func reportFile(pass *analysis.Pass, file *ast.File, agreed norms) {
	for _, spec := range file.Imports {
		if at, want, ok := deviation(pass.TypesInfo, file, spec, agreed); ok {
			pass.Reportf(specPos(spec), message, string(at), spellingOf(pass.TypesInfo, spec), want)
			continue
		}
		if at, natural, ok := leftoverNumber(pass.TypesInfo, file, spec); ok {
			pass.Reportf(specPos(spec), messageNumbered, string(at), spellingOf(pass.TypesInfo, spec), string(natural))
		}
	}
}

// leftoverNumber reports whether spec carries a numbered alias no longer
// serving any purpose — `abc2` where nothing else in the file is called `abc`.
//
// This is the residue of a move. The tools mint `abc2` when two packages named
// `abc` meet in one file; when one of them later moves away the collision goes
// with it, and the digit survives forever, naming a package after an argument
// it no longer has with anything. It is reported only when the plain name is
// genuinely free in this file, so the fix is always to delete the alias.
func leftoverNumber(info *types.Info, file *ast.File, spec *ast.ImportSpec) (importPath, spelling, bool) {
	at, ok := pathOf(spec)
	if spec.Name == nil || !ok {
		return "", "", false
	}
	natural := naturalOf(info, spec)
	suffix, cut := strings.CutPrefix(spec.Name.Name, string(natural))
	if natural == "" || !cut || !numbered.MatchString(suffix) {
		return "", "", false
	}
	if justified(info, file, spec, natural) {
		return "", "", false
	}
	return at, natural, true
}

// deviation reports whether spec breaks the agreed spelling with no collision
// to justify it, yielding the path and the spelling the rest of the package
// uses.
func deviation(
	info *types.Info, file *ast.File, spec *ast.ImportSpec, agreed norms,
) (importPath, spelling, bool) {
	at, ok := pathOf(spec)
	if !ok {
		return "", "", false
	}
	want, disputed := agreed[at]
	if !disputed || spellingOf(info, spec) == want || justified(info, file, spec, want) {
		return "", "", false
	}
	return at, want, true
}

// justified reports whether the file must spell this import its own way: taking
// the module's spelling would collide with another import in the same file, so
// demanding it would demand code that does not compile.
func justified(info *types.Info, file *ast.File, spec *ast.ImportSpec, want spelling) bool {
	for _, other := range file.Imports {
		if other != spec && spellingOf(info, other) == want {
			return true
		}
	}
	return false
}

// specPos anchors a finding at the name the file chose when it wrote one, and
// at the path otherwise — always at the text a reader would change.
func specPos(spec *ast.ImportSpec) token.Pos {
	if spec.Name != nil {
		return spec.Name.Pos()
	}
	return spec.Path.Pos()
}
