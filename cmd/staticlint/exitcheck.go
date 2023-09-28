package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for direct os.Exit calls in main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, inspectNode(pass))
	}
	return nil, nil
}

func inspectNode(pass *analysis.Pass) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		if fn, ok := node.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			inspectStmts(fn, pass)
		}

		return true
	}
}

func inspectStmts(fn *ast.FuncDecl, pass *analysis.Pass) {
	for _, stmt := range fn.Body.List {
		if es, ok := stmt.(*ast.ExprStmt); ok {
			if ce, ok := es.X.(*ast.CallExpr); ok {
				inspectExpr(ce, pass)
			}
		}
	}
}

func inspectExpr(ce *ast.CallExpr, pass *analysis.Pass) {
	if se, ok := ce.Fun.(*ast.SelectorExpr); ok {
		if id, ok := se.X.(*ast.Ident); ok && id.Name == "os" && se.Sel.Name == "Exit" {
			pass.Reportf(ce.Pos(), "os.Exit() should not be used in main function")
		}
	}
}
