package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var Analyzer = &analysis.Analyzer{
	Name:             "temporal_linter",
	Doc:              "reports bad usage of the Temporal logger",
	Run:              run,
	RunDespiteErrors: true,
}

func main() {
	singlechecker.Main(Analyzer)
}

// https://cs.opensource.google/go/x/tools/+/refs/tags/v0.26.0:go/analysis/passes/printf/printf.go;l=985
var printFormatRE = regexp.MustCompile(`%` + flagsRE + numOptRE + `\.?` + numOptRE + indexOptRE + verbRE)

const (
	flagsRE    = `[+\-#]*`
	indexOptRE = `(\[[0-9]+\])?`
	numOptRE   = `([0-9]+|` + indexOptRE + `\*)?`
	verbRE     = `[bcdefgopqstvxEFGTUX]`
)

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			be, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			se, ok := (be.Fun).(*ast.SelectorExpr)
			if !ok {
				return true
			}

			exp, ok := se.X.(*ast.Ident)
			if !ok {
				return true
			}

			x := strings.ToLower(exp.String())

			// Look for common name of a logger
			if x != "log" && x != "l" && x != "logger" {
				return false
			}

			// Look for Temporal logger
			if !slices.Contains([]string{"Debug", "Info", "Warn", "Error"}, se.Sel.Name) {
				return true
			}

			if len(be.Args)%2 == 0 {
				pass.Reportf(be.Pos(), "invalid log usage: %s", render(pass.Fset, be))
			}

			for _, arg := range be.Args {
				basicLit, ok := arg.(*ast.BasicLit)
				if !ok {
					return false
				}

				if basicLit.Kind != token.STRING {
					return false
				}

				s := basicLit.Value

				// Ignore trailing % character
				// The % in "abc 0.0%" couldn't be a formatting directive.
				s = strings.TrimSuffix(s, "%")
				if strings.Contains(s, "%") {
					m := printFormatRE.FindStringSubmatch(s)

					if m != nil {
						pass.Reportf(be.Pos(), "%s.%s call has possible logf-style formatting directive %s", exp, se.Sel.Name, m[0])
					}
				}

			}

			return true
		})
	}

	return nil, nil
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
