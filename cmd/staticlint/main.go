package main

import (
	"strings"
	"unicode"

	"github.com/charithe/durationcheck"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	rules := []string{"ST1003", "ST1016", "SA"}

	analyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		durationcheck.Analyzer,
		ExitCheckAnalyzer,
	}

	checks := make(map[string]bool)
	for _, v := range rules {
		checks[v] = true
	}

	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			analyzers = append(analyzers, v.Analyzer)
			continue
		}

		c := strings.TrimFunc(v.Analyzer.Name, func(r rune) bool {
			return !unicode.IsLetter(r)
		})

		if checks[c] {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	multichecker.Main(analyzers...)
}
