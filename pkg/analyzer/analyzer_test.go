package analyzer_test

import (
	"testing"

	"github.com/golchanskiy23/loglint/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestLowerCase(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "loglint/lowercase")
}

func TestEnglish(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "loglint/english")
}

func TestSpecialSymbols(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "loglint/special")
}

func TestSensitiveWords(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "loglint/sensitive")
}
