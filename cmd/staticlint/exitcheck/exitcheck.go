// Package exitcheck содержит статический анализатор, запрещающий прямой
// вызов os.Exit в функции main пакета main.
//
// Прямой вызов os.Exit в main прерывает выполнение немедленно, минуя
// отложенные вызовы (defer), корректное завершение фоновых горутин и
// освобождение ресурсов. Это затрудняет graceful shutdown, поэтому в
// функции main рекомендуется возвращать управление обычным образом
// (например, вынести логику в отдельную функцию run, возвращающую error).
//
// Анализатор срабатывает только на прямой вызов вида os.Exit(...),
// находящийся непосредственно в теле функции main пакета main. Вызовы
// os.Exit в других функциях, в других пакетах, а также косвенные
// завершения (log.Fatal и т. п.) не проверяются.
package exitcheck

import (
	"go/ast"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
)

// Analyzer — экземпляр анализатора запрета прямого os.Exit в main.main.
// Регистрируется в multichecker наравне со стандартными анализаторами.
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

// run реализует логику анализатора: обходит файлы пакета main, находит
// функцию main и сообщает о каждом прямом вызове os.Exit в её теле.
func run(pass *analysis.Pass) (interface{}, error) {
	// Проверяем только пакет main — в остальных os.Exit допустим.
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Пропускаем сгенерированные файлы: сам код мы не писали, а
		// сгенерированный тестовый main (_testmain.go) штатно содержит
		// os.Exit(m.Run()) и не является нарушением.
		if isGenerated(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			// Интересует только функция main верхнего уровня без receiver.
			if fn.Recv != nil || fn.Name.Name != "main" {
				continue
			}
			checkMainBody(pass, fn)
		}
	}
	return nil, nil
}

// checkMainBody обходит тело функции main и сообщает о каждом прямом
// вызове os.Exit.
func checkMainBody(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isOsExitCall(call) {
			pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещён")
		}
		return true
	})
}

// isGenerated сообщает, является ли файл сгенерированным. Учитываются два
// случая: файл содержит стандартный комментарий «Code generated ... DO NOT
// EDIT» (ast.IsGenerated) либо это синтезированный тестовый main
// (_testmain.go), создаваемый инструментарием go test.
func isGenerated(pass *analysis.Pass, file *ast.File) bool {
	if ast.IsGenerated(file) {
		return true
	}
	name := pass.Fset.Position(file.Pos()).Filename
	return filepath.Base(name) == "_testmain.go"
}

// isOsExitCall сообщает, является ли выражение вызовом os.Exit,
// записанным как селектор пакета os.
func isOsExitCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != "Exit" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "os"
}
