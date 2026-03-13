package main

import (
	"github.com/golchanskiy23/loglint/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

var AnalyzerPlugin analyzerPlugin

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{analyzer.Analyzer}
}
