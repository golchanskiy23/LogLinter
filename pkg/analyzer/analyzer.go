package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	flagDisableLowercase bool
	flagDisableEnglish   bool
	flagDisableSpecial   bool
	flagDisableSensitive bool
	flagExtraKeywords    string
	flagConfigFile       string
)

func init() {
	Analyzer.Flags.BoolVar(&flagDisableLowercase, "disable-lowercase", false,
		"disable lowercase check")
	Analyzer.Flags.BoolVar(&flagDisableEnglish, "disable-english", false,
		"disable english-only check")
	Analyzer.Flags.BoolVar(&flagDisableSpecial, "disable-special", false,
		"disable special characters check")
	Analyzer.Flags.BoolVar(&flagDisableSensitive, "disable-sensitive", false,
		"disable sensitive data check")
	Analyzer.Flags.StringVar(&flagExtraKeywords, "extra-sensitive", "",
		"comma-separated extra sensitive keywords")
	Analyzer.Flags.StringVar(&flagConfigFile, "config", "",
		"path to configuration file")
}

// Config structure for future YAML support
type Config struct {
	DisableLowercase bool     `yaml:"disable-lowercase"`
	DisableEnglish   bool     `yaml:"disable-english"`
	DisableSpecial   bool     `yaml:"disable-special"`
	DisableSensitive bool     `yaml:"disable-sensitive"`
	ExtraKeywords    []string `yaml:"extra-sensitive"`
}

func loadConfig() {
	data, err := os.ReadFile(".loglint.yml")
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "disable-lowercase:") {
			if strings.Contains(line, "true") && !flagDisableLowercase {
				flagDisableLowercase = true
			}
		} else if strings.Contains(line, "disable-english:") {
			if strings.Contains(line, "true") && !flagDisableEnglish {
				flagDisableEnglish = true
			}
		} else if strings.Contains(line, "disable-special:") {
			if strings.Contains(line, "true") && !flagDisableSpecial {
				flagDisableSpecial = true
			}
		} else if strings.Contains(line, "disable-sensitive:") {
			if strings.Contains(line, "true") && !flagDisableSensitive {
				flagDisableSensitive = true
			}
		} else if strings.Contains(line, "extra-sensitive:") {
			continue
		} else if strings.HasPrefix(line, "-") {
			keyword := strings.TrimSpace(strings.TrimPrefix(line, "-"))
			keyword = strings.Trim(keyword, "\"")
			if keyword != "" && flagExtraKeywords == "" {
				flagExtraKeywords = keyword
			} else if keyword != "" {
				flagExtraKeywords += "," + keyword
			}
		}
	}
}

var Analyzer = &analysis.Analyzer{
	Name:     "loglint",
	Doc:      "checks log messages for style and security rules",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	loadConfig()

	var extraKeywords []string
	if flagExtraKeywords != "" {
		for _, kw := range strings.Split(flagExtraKeywords, ",") {
			if trimmed := strings.TrimSpace(kw); trimmed != "" {
				extraKeywords = append(extraKeywords, trimmed)
			}
		}
	}

	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !isLogCall(call) {
			return
		}

		if !flagDisableSensitive && hasSensitiveConcatenation(call, extraKeywords) {
			pass.Reportf(call.Pos(),
				"log message concatenates potentially sensitive variable")
			return
		}

		msg, ok := extractMessage(call)
		if !ok {
			return
		}
		checkMessage(pass, call, msg, extraKeywords)
	})
	return nil, nil
}

var logMethods = map[string]bool{
	"Info": true, "Error": true, "Warn": true, "Warning": true,
	"Debug": true, "Fatal": true, "Panic": true,
	"Print": true, "Printf": true, "Println": true,
}

var logPackages = map[string]bool{
	"log": true, "slog": true, "zap": true,
}

func isLogCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if !logMethods[sel.Sel.Name] {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return logPackages[ident.Name]
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
	'!': true, '?': true, ';': true, '.': true, '-': true,
}

func isEmoji(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x27BF)
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

func lowercaseFix(pass *analysis.Pass, lit *ast.BasicLit, msg string) analysis.Diagnostic {
	runes := []rune(msg)
	runes[0] = unicode.ToLower(runes[0])
	fixed := strconv.Quote(string(runes))

	return analysis.Diagnostic{
		Pos:     lit.Pos(),
		Message: fmt.Sprintf("log message must start with lowercase letter, got %q", msg),
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "convert first letter to lowercase",
			TextEdits: []analysis.TextEdit{{
				Pos:     lit.Pos(),
				End:     lit.End(),
				NewText: []byte(fixed),
			}},
		}},
	}
}

func checkLowercase(pass *analysis.Pass, call *ast.CallExpr, msg string) {
	runes := []rune(msg)
	if len(runes) == 0 {
		return
	}

	if unicode.IsUpper(runes[0]) {
		if lit, ok := call.Args[0].(*ast.BasicLit); ok {
			pass.Report(lowercaseFix(pass, lit, msg))
		} else {
			pass.Reportf(call.Pos(),
				"log message must start with lowercase letter, got %q", msg)
		}
	}
}

var sensitiveKeywords = []string{
	"password", "passwd", "secret", "token",
	"api_key", "apikey", "apitoken", "auth",
	"credential", "private_key", "privatekey",
}

func containsSensitiveKeyword(lower, keyword string) bool {
	idx := strings.Index(lower, keyword)
	if idx == -1 {
		return false
	}

	if idx > 0 && (unicode.IsLetter(rune(lower[idx-1])) || lower[idx-1] == '_') {
		return false
	}

	end := idx + len(keyword)
	if end < len(lower) && (unicode.IsLetter(rune(lower[end])) || lower[end] == '_') {
		return false
	}

	return true
}

func checkSensitive(pass *analysis.Pass, call *ast.CallExpr, msg string, extraKeywords []string) {
	lower := strings.ToLower(msg)

	// Check built-in keywords
	for _, sk := range sensitiveKeywords {
		if containsSensitiveKeyword(lower, sk) {
			pass.Reportf(call.Pos(),
				"log message may contain sensitive data (keyword %q found)", sk)
			return
		}
	}

	// Check extra keywords
	for _, sk := range extraKeywords {
		if containsSensitiveKeyword(lower, strings.ToLower(sk)) {
			pass.Reportf(call.Pos(),
				"log message may contain sensitive data (keyword %q found)", sk)
			return
		}
	}

	if hasSensitiveConcatenation(call, extraKeywords) {
		pass.Reportf(call.Pos(),
			"log message concatenates potentially sensitive variable")
	}
}

func hasSensitiveConcatenation(call *ast.CallExpr, extraKeywords []string) bool {
	if len(call.Args) == 0 {
		return false
	}

	return containsSensitiveVar(call.Args[0], extraKeywords)
}

func containsSensitiveVar(expr ast.Expr, extraKeywords []string) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			return containsSensitiveVar(e.X, extraKeywords) || containsSensitiveVar(e.Y, extraKeywords)
		}
	case *ast.Ident:
		name := strings.ToLower(e.Name)
		for _, sk := range sensitiveKeywords {
			if strings.Contains(name, sk) {
				return true
			}
		}
		for _, sk := range extraKeywords {
			if strings.Contains(name, strings.ToLower(sk)) {
				return true
			}
		}
	}
	return false
}

func checkMessage(pass *analysis.Pass, call *ast.CallExpr, msg string, extraKeywords []string) {
	if !flagDisableLowercase {
		checkLowercase(pass, call, msg)
	}
	if !flagDisableEnglish {
		checkEnglish(pass, call, msg)
	}
	if !flagDisableSpecial {
		checkSpecialChars(pass, call, msg)
	}
	if !flagDisableSensitive {
		checkSensitive(pass, call, msg, extraKeywords)
	}
}
