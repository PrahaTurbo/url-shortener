package main

import (
	"strings"

	"github.com/charithe/durationcheck"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	styleRules := []string{"ST1003", "ST1016"}
	staticRules := []string{"SA"}

	analyzers := []*analysis.Analyzer{
		printf.Analyzer,        // printf.Analyzer: Checks whether Printf family functions are used correctly.
		shadow.Analyzer,        // shadow.Analyzer: Checks for shadowed variables.
		structtag.Analyzer,     // structtag.Analyzer: Check that struct field tags conform to reflect.StructTag conventions.
		errcheck.Analyzer,      // errcheck.Analyzer: Checks that error return values are used.
		durationcheck.Analyzer, // durationcheck.Analyzer: Check for two durations multiplied together.
		ExitCheckAnalyzer,      // ExitCheckAnalyzer: Checks for direct os.Exit calls in the main function.
	}

	for _, v := range stylecheck.Analyzers {
		for _, rule := range styleRules {
			if strings.Contains(v.Analyzer.Name, rule) {
				analyzers = append(analyzers, v.Analyzer)
			}
		}
	}

	for _, v := range staticcheck.Analyzers {
		for _, rule := range staticRules {
			if strings.Contains(v.Analyzer.Name, rule) {
				analyzers = append(analyzers, v.Analyzer)
			}
		}
	}

	multichecker.Main(analyzers...)
}
