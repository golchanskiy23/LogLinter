package analyzer

import (
	"fmt"
	"go/ast"

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
	fmt.Println("Correct work")
	return nil, nil
}

func isLogCall(call *ast.CallExpr) bool {
	// TODO: Реализовать проверку для log/slog и go.uber.org/zap
	return false
}

func extractMessage(call *ast.CallExpr) (string, bool) {
	// TODO: Реализовать извлечение сообщения из аргументов вызова
	return "", false
}

func checkMessage(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	// TODO: Реализовать проверку правил:
	// 1. Строчная буква в начале
	// 2. Английский язык
	// 3. Без спецсимволов/эмодзи
	// 4. Без чувствительных данных
}
