package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		sortslice.Analyzer,
		appends.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		MainOSExitCheckAnalyzer,
	}

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	styleChecks := map[string]bool{
		"ST1000": true,
	}

	for _, v := range stylecheck.Analyzers {
		if styleChecks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	simpleChecks := map[string]bool{
		"S1001": true,
	}
	for _, v := range simple.Analyzers {
		if simpleChecks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}

var MainOSExitCheckAnalyzer = &analysis.Analyzer{
	Name: "mainosexitcheck",
	Doc:  "check for call of os.Exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	expr := func(expr *ast.SelectorExpr) {
		x, ok := expr.X.(*ast.Ident)
		if !ok || x.Name != "os" {
			return
		}

		if expr.Sel.Name != "Exit" {
			return
		}

		pass.Reportf(x.Pos(), "os.Exit not allowed call in main func")
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				return x.Name.Name == "main"
			case *ast.SelectorExpr:
				expr(x)
			}
			return true
		})
	}
	return nil, nil
}
