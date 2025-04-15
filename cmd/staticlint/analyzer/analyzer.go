package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "metricsAnalyzer",
	Doc:  "Analyzer searches for os.Exit calls in main func of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Ищем функцию main.
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				return true
			}

			// Теперь ищем вызовы os.Exit в теле функции main
			ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if ident, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
					if ident.Sel.Name == "Exit" {
						pass.Reportf(callExpr.Pos(), "call to os.Exit in main function")
					}
				}
				return true
			})
			// Останавливаем поиск, если нашли функцию main.
			return false
		})
	}
	return nil, nil
}
