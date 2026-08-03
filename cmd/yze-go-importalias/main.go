// Command yze-go-importalias runs the importalias analyzer as a standalone go/analysis
// checker (text and -json output, and usable as a `go vet -vettool`).
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	importalias "github.com/gomatic/yze-go-importalias"
)

// run is the analysis entry point, indirected so the binary's wiring is testable
// without invoking the real driver (which loads packages and exits the process).
var run = singlechecker.Main

func main() { run(importalias.Analyzer) }
