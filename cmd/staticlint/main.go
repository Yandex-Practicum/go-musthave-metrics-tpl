// Command staticlint — статический анализатор (multichecker) проекта
// go-musthave-metrics-tpl. Он объединяет в одном исполняемом файле
// множество анализаторов и запускается поверх пакетов проекта.
//
// # Состав multichecker
//
// В сборку включены следующие группы анализаторов:
//
//  1. Стандартные анализаторы golang.org/x/tools/go/analysis/passes —
//     набор проверок, используемых в том числе командой go vet
//     (printf, shadow-независимые проверки, copylock, httpresponse и др.).
//  2. Все анализаторы класса SA пакета staticcheck.io
//     (honnef.co/go/tools/staticcheck) — обнаружение реальных ошибок и
//     подозрительных конструкций.
//  3. Анализаторы прочих классов staticcheck.io: класс S
//     (honnef.co/go/tools/simple, упрощение кода) и класс ST
//     (honnef.co/go/tools/stylecheck, вопросы стиля).
//  4. Публичные сторонние анализаторы:
//     - bodyclose (github.com/timakin/bodyclose) — проверяет, что тело
//     http.Response.Body закрывается;
//     - ineffassign (github.com/gordonklaus/ineffassign) — находит
//     присваивания, значения которых нигде не используются.
//  5. Собственный анализатор exitcheck — запрещает прямой вызов os.Exit
//     в функции main пакета main.
//
// # Запуск
//
// Сборка исполняемого файла:
//
//	go build -o staticlint ./cmd/staticlint
//
// Запуск на всех пакетах проекта:
//
//	./staticlint ./...
//
// Запуск на конкретных пакетах:
//
//	./staticlint ./internal/... ./cmd/server/... ./cmd/agent/...
//
// Список включённых анализаторов и справку по каждому можно получить так:
//
//	./staticlint help
//	./staticlint help exitcheck
//
// Multichecker возвращает ненулевой код выхода, если хотя бы один
// анализатор нашёл проблему, что удобно использовать в CI.
package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/bluegopher/go-musthave-metrics-tpl/cmd/staticlint/exitcheck"
)

func main() {
	checks := standardPasses()
	checks = append(checks, staticcheckAnalyzers()...)
	checks = append(checks, thirdPartyAnalyzers()...)
	checks = append(checks, exitcheck.Analyzer)

	multichecker.Main(checks...)
}

// standardPasses возвращает стандартные анализаторы из
// golang.org/x/tools/go/analysis/passes.
func standardPasses() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}
}

// staticcheckAnalyzers возвращает все анализаторы класса SA пакета
// staticcheck, а также по одному представителю прочих классов: S (simple)
// и ST (stylecheck).
func staticcheckAnalyzers() []*analysis.Analyzer {
	var checks []*analysis.Analyzer

	// Все анализаторы класса SA — поиск ошибок.
	for _, a := range staticcheck.Analyzers {
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			checks = append(checks, a.Analyzer)
		}
	}

	// Класс S (simple) — упрощение кода.
	for _, a := range simple.Analyzers {
		checks = append(checks, a.Analyzer)
	}

	// Класс ST (stylecheck) — стиль кода.
	for _, a := range stylecheck.Analyzers {
		checks = append(checks, a.Analyzer)
	}

	return checks
}

// thirdPartyAnalyzers возвращает публичные сторонние анализаторы.
func thirdPartyAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		bodyclose.Analyzer,
		ineffassign.Analyzer,
	}
}
