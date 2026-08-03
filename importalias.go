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
	"path"
	"strings"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
)

const message = "import %q is spelled %s here but %s elsewhere in this module; one import, one spelling"

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

// importPath is a package's quoted import path, unquoted.
type importPath string

// spelling is the name an import is referred to by in a file: an explicit
// alias, or the package's own name when the import carries none.
type spelling string

// run reports every import whose spelling differs from the module's own.
//
// Scope is the pass: a package at a time. That is deliberately narrower than
// "the module" — a divergence between two packages the driver analyses
// separately falls outside it — and it buys two things. Every file of a package
// is seen together, which is where a reader suffers the inconsistency; and the
// judgement rests on the files the driver loaded rather than on a directory
// walk's guess about which ones count.
func run(pass *analysis.Pass) (any, error) {
	norms := normsOf(pass.TypesInfo, pass.Files)
	for _, file := range pass.Files {
		reportFile(pass, file, norms)
	}
	return nil, nil
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

// naturalOf is the name the imported package declares for itself — what a
// plain import binds, whether or not this particular spec aliases it.
func naturalOf(info *types.Info, spec *ast.ImportSpec) spelling {
	if spec.Name == nil {
		if named, ok := info.Implicits[spec].(*types.PkgName); ok {
			return spelling(named.Name())
		}
		return ""
	}
	if named, ok := info.Defs[spec.Name].(*types.PkgName); ok {
		return spelling(named.Imported().Name())
	}
	return ""
}

// dominant is the spelling a path is written with most often.
//
// Ties go to the name the package declares for itself: an even split between a
// plain import and an alias is a module that has not decided, and the language's
// own default is the answer least likely to surprise a reader — reporting the
// PLAIN import as the deviation would be perverse. A tie between two aliases,
// where no natural name is in play, falls back to the lexicographically smaller
// so the answer never depends on the order files are read in.
func dominant(counted tally) spelling {
	var best spelling
	for name, count := range counted.counts {
		if best == "" || count > counted.counts[best] ||
			(count == counted.counts[best] && prefer(name, best, counted.natural)) {
			best = name
		}
	}
	return best
}

// prefer breaks a tie between two equally-used spellings: the package's own
// name wins, otherwise the lexicographically smaller.
func prefer(name, best, natural spelling) bool {
	if natural != "" && (name == natural || best == natural) {
		return name == natural
	}
	return name < best
}

// reportFile reports each of the file's imports that deviates from the norm
// without a collision to justify it.
func reportFile(pass *analysis.Pass, file *ast.File, agreed norms) {
	for _, spec := range file.Imports {
		if at, want, ok := deviation(pass.TypesInfo, file, spec, agreed); ok {
			pass.Reportf(specPos(spec), message, string(at), spellingOf(pass.TypesInfo, spec), want)
		}
	}
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

// pathOf unquotes an import spec's path, reporting false for one that is not a
// plain quoted string.
func pathOf(spec *ast.ImportSpec) (importPath, bool) {
	if spec.Path == nil || len(spec.Path.Value) < 2 {
		return "", false
	}
	unquoted := strings.Trim(spec.Path.Value, `"`)
	if unquoted == "" {
		return "", false
	}
	return importPath(unquoted), true
}

// spellingOf is the name a file refers to an import by: its alias, or the
// package's own name when the import carries none.
//
// The declared name comes from the type checker, never from the path's last
// element — those disagree exactly where it would matter most. Importing
// "github.com/urfave/cli/v3" plainly binds `cli`, not `v3`, so guessing from
// the path would read an explicit `cli` alias as a disagreement with the
// identical plain import and report a file for matching the norm it already
// matches. A path the checker has no name for falls back to the last element,
// which is the go tool's own default for a well-formed path.
//
// The blank and dot forms are spellings in their own right and compared as
// such: a package imported for effect in one file and by name in another is
// exactly the divergence worth reporting, since the two do different things.
func spellingOf(info *types.Info, spec *ast.ImportSpec) spelling {
	if spec.Name != nil {
		return spelling(spec.Name.Name)
	}
	if named, ok := info.Implicits[spec].(*types.PkgName); ok {
		return spelling(named.Name())
	}
	at, ok := pathOf(spec)
	if !ok {
		return ""
	}
	return spelling(path.Base(string(at)))
}
