package importalias

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

// spec builds an import spec with the given path literal and optional alias.
func spec(literal, alias string) *ast.ImportSpec {
	built := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: literal, ValuePos: 7}}
	if alias != "" {
		built.Name = &ast.Ident{Name: alias, NamePos: 3}
	}
	return built
}

// TestPathOfRejectsWhatIsNotAQuotedPath pins the parse guard: a well-formed
// quoted path yields its contents, and anything else — no literal at all, a
// stub too short to be quoted, an empty path — yields nothing rather than a
// bogus import path the rest of the rule would compare against.
func TestPathOfRejectsWhatIsNotAQuotedPath(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	got, ok := pathOf(spec(`"lib"`, ""))
	want.True(ok)
	want.Equal(importPath("lib"), got)

	_, ok = pathOf(&ast.ImportSpec{})
	want.False(ok, "a spec with no path names nothing")

	_, ok = pathOf(&ast.ImportSpec{Path: &ast.BasicLit{Value: `"`}})
	want.False(ok, "a literal too short to be a quoted path names nothing")

	_, ok = pathOf(&ast.ImportSpec{Path: &ast.BasicLit{Value: `""`}})
	want.False(ok, "an empty path names nothing")
}

// TestSpellingOfPrefersTheCheckersName pins the name resolution in all three
// forms: an explicit alias wins, an unaliased import takes the name the type
// checker recorded, and a path the checker knows nothing about falls back to
// the last element — the go tool's own default. A spec with no usable path has
// no spelling at all.
func TestSpellingOfPrefersTheCheckersName(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	plain := spec(`"versioned/cli/v3"`, "")
	info := &types.Info{Implicits: map[ast.Node]types.Object{
		plain: types.NewPkgName(0, nil, "cli", types.NewPackage("versioned/cli/v3", "cli")),
	}}

	want.Equal(spelling("cli"), spellingOf(info, plain), "the checker's name beats the path's last element")
	want.Equal(spelling("alias"), spellingOf(info, spec(`"lib"`, "alias")), "an explicit alias wins")
	want.Equal(spelling("lib"), spellingOf(&types.Info{}, spec(`"lib"`, "")), "an unknown path falls back")
	want.Equal(spelling(""), spellingOf(&types.Info{}, &ast.ImportSpec{}), "an unusable spec has no spelling")
}

// TestSpecPosAnchorsAtTheTextToChange pins where a finding points: at the alias
// a file wrote, and at the path when it wrote none — in both cases the text a
// reader would edit.
func TestSpecPosAnchorsAtTheTextToChange(t *testing.T) {
	t.Parallel()

	assert.Equal(t, token.Pos(3), specPos(spec(`"lib"`, "alias")))
	assert.Equal(t, token.Pos(7), specPos(spec(`"lib"`, "")))
}

// TestCountSpellingIgnoresAnUnusableSpec pins that a spec with no readable path
// contributes nothing to the tally, so it can never invent a disagreement.
func TestCountSpellingIgnoresAnUnusableSpec(t *testing.T) {
	t.Parallel()

	counts := map[importPath]tally{}
	countSpelling(&types.Info{}, counts, &ast.ImportSpec{})
	assert.Empty(t, counts)
}

// TestDeviationIsSilentOnAnUnusableSpec pins the same guard on the reporting
// side: nothing is reported for an import whose path cannot be read.
func TestDeviationIsSilentOnAnUnusableSpec(t *testing.T) {
	t.Parallel()

	_, _, ok := deviation(&types.Info{}, &ast.File{}, &ast.ImportSpec{}, norms{})
	assert.False(t, ok)
}

// TestDominantPrefersUseThenTheNaturalName pins the determinism the rule rests
// on: the spelling used most often wins; an even split goes to the name the
// package declares for itself, because reporting a PLAIN import as the
// deviation from an alias would be perverse; and a tie between two aliases with
// no natural name in play falls back to the smaller name so the answer never
// depends on the order files are read in.
func TestDominantPrefersUseThenTheNaturalName(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.Equal(spelling("errs"), dominant(tally{
		counts:  map[spelling]int{"errs": 2, "goerror": 1},
		natural: "goerror",
	}), "a spelling the module actually agreed on beats the natural name")

	want.Equal(spelling("lib"), dominant(tally{
		counts:  map[spelling]int{"lib": 1, "alias": 1},
		natural: "lib",
	}), "an even split goes to the package's own name")

	want.Equal(spelling("alpha"), dominant(tally{
		counts: map[spelling]int{"beta": 1, "alpha": 1},
	}), "with no natural name, the smaller name keeps the answer deterministic")
}

// TestNaturalOfReadsTheImportedPackagesOwnName pins where the natural spelling
// comes from in both import forms — an aliased spec still knows the name the
// package declares for itself, which is what makes the tie-break possible.
func TestNaturalOfReadsTheImportedPackagesOwnName(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	imported := types.NewPackage("versioned/cli/v3", "cli")

	plain := spec(`"versioned/cli/v3"`, "")
	plainInfo := &types.Info{Implicits: map[ast.Node]types.Object{
		plain: types.NewPkgName(0, nil, "cli", imported),
	}}
	want.Equal(spelling("cli"), naturalOf(plainInfo, plain))

	aliased := spec(`"versioned/cli/v3"`, "urfave")
	aliasedInfo := &types.Info{Defs: map[*ast.Ident]types.Object{
		aliased.Name: types.NewPkgName(0, nil, "urfave", imported),
	}}
	want.Equal(spelling("cli"), naturalOf(aliasedInfo, aliased), "an alias still knows the real name")

	want.Equal(spelling(""), naturalOf(&types.Info{}, plain), "an unknown plain import has no natural name")
	want.Equal(spelling(""), naturalOf(&types.Info{}, aliased), "an unknown aliased import has no natural name")
}

// TestSourceFilesReadTheNameTheFileCannotRewrite pins the test-file exclusion to the FileSet's own
// entry for a file. A //line directive compiles to exactly the
// AddLineColumnInfo calls made here: fset.Position reads that alternative
// information and token.File.Name() ignores it, so the disagreement below is
// the one a directive produces, built directly rather than parsed.
//
// Both directions are asserted because reading the rewritten name is wrong in
// both: it hides a deviation in compiled source and withdraws that file's vote from the norm, while holding a real test file to a norm it was excused from.
func TestSourceFilesReadTheNameTheFileCannotRewrite(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	shipped := fset.AddFile("shipped.go", -1, 100)
	shipped.AddLineColumnInfo(0, "zz_test.go", 1, 1)
	tested := fset.AddFile("shipped_test.go", -1, 100)
	tested.AddLineColumnInfo(0, "nottest.go", 1, 1)
	source := &ast.File{Package: shipped.Pos(1)}
	pass := &analysis.Pass{Fset: fset, Files: []*ast.File{source, {Package: tested.Pos(1)}}}

	assert.Equal(t, "zz_test.go", fset.Position(shipped.Pos(1)).Filename,
		"the position machinery does adopt the claimed name — without this the rest asserts nothing")
	assert.Equal(t, "nottest.go", fset.Position(tested.Pos(1)).Filename,
		"and in the other direction too")

	assert.Equal(t, []*ast.File{source}, sourceFiles(pass),
		"the agreement holds the file the go tool compiles as source, and only that one")
}
