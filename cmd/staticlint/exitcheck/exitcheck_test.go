package exitcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/bluegopher/go-musthave-metrics-tpl/cmd/staticlint/exitcheck"
)

// TestExitCheck прогоняет анализатор на тестовых пакетах из testdata и
// сверяет диагностику с комментариями // want.
func TestExitCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitcheck.Analyzer, "mainpkg", "otherpkg")
}
