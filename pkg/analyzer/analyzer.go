package analyzer

import (
	"go/ast"
	"go/token"
	"strconv"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "loglint",
	Doc:      "checks log messages for style and security rules",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !isLogCall(call) {
			return
		}
		msg, ok := extractMessage(call)
		if !ok {
			return
		}
		checkMessage(pass, call, msg)
	})
	return nil, nil
}

var logMethods = map[string]bool{
	"Info": true, "Error": true, "Warn": true, "Warning": true,
	"Debug": true, "Fatal": true, "Panic": true,
}

var logPackages = map[string]bool{
	"log": true, "slog": true, "zap": true,
}

func isLogCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return logMethods[sel.Sel.Name]
}

func extractMessage(call *ast.CallExpr) (string, bool) {
	if len(call.Args) == 0 {
		return "", false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	msg, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return msg, true
}

func checkLowercase(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	runes := []rune(msg)
	if len(runes) == 0 {
		return
	}
	if unicode.IsUpper(runes[0]) {
		pass.Reportf(call.Pos(),
			"log message must start with lowercase letter, got %q", msg)
	}
}

func checkMessage(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	checkLowercase(pass, call, msg)
}
