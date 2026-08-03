package importalias

import (
	"go/ast"
	"go/types"
	"path"
	"strings"
)

// Resolving what a file calls an import, and what the package agreed to call
// it. The name always comes from the type checker rather than the import
// path's last element: "urfave/cli/v3" binds cli, not v3.

// importPath is a package's quoted import path, unquoted.
type importPath string

// spelling is the name an import is referred to by in a file: an explicit
// alias, or the package's own name when the import carries none.
type spelling string

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
