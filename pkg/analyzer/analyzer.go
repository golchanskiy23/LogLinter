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

var forbiddenChars = map[rune]bool{
	'!': true, '?': true, ';': true,
}

func isEmoji(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1FAFF) || // Misc Symbols, Emoticons, etc.
		(r >= 0x2600 && r <= 0x27BF) // Misc Symbols block
}

func checkSpecialChars(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	for _, r := range msg {
		if isEmoji(r) {
			pass.Reportf(call.Pos(),
				"log message must not contain emoji, found %q", string(r))
			return
		}
		if forbiddenChars[r] {
			pass.Reportf(call.Pos(),
				"log message must not contain special character %q", string(r))
			return
		}
	}
}

func isAllowedRune(r rune) bool {
	return r <= 127
}

func checkEnglish(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	for _, r := range msg {
		if unicode.IsLetter(r) && !isAllowedRune(r) {
			pass.Reportf(call.Pos(),
				"log message must be in English, found non-Latin character %q", string(r))
			return
		}
	}
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
	// Правило 1: Строчная буква в начале
	checkLowercase(pass, call, msg)

	// Правило 2: Только английский язык
	checkEnglish(pass, call, msg)

	// Правило 3: Без спецсимволов и эмодзи
	checkSpecialChars(pass, call, msg)
}
