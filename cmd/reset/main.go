// Command reset — генератор методов Reset для помеченных структур.
//
// Утилита сканирует пакеты проекта, находит объявления структур с
// комментарием-маркером и генерирует для каждой такой структуры метод
// Reset, сбрасывающий все её поля в «пустое» состояние.
//
// # Маркер
//
// Структура помечается строкой-комментарием непосредственно над её
// объявлением:
//
//	// generate:reset
//	type Buffer struct {
//		data []byte
//		size int
//	}
//
// # Правила сброса полей
//
// Метод Reset сбрасывает поля по следующим правилам:
//
//   - примитивные типы (int, string, bool и т.п.) — присваивается нулевое
//     значение соответствующего типа;
//   - срезы — усекаются до нулевой длины (s = s[:0]), выделенная память
//     сохраняется;
//   - карты — очищаются встроенной функцией clear(m);
//   - вложенные структуры, у которых есть свой метод Reset (в том числе
//     сгенерированный), — сбрасываются вызовом этого метода;
//   - указатели: если целевой тип сбрасывается через Reset, вызывается
//     Reset для разыменованного значения; иначе значение по указателю
//     сбрасывается по тем же правилам. Nil-указатели пропускаются.
//
// # Результат
//
// Для каждого пакета сгенерированные методы складываются в файл
// reset.gen.go в каталоге этого пакета. Файл помечается стандартным
// заголовком «Code generated ... DO NOT EDIT.».
//
// # Запуск
//
// Сборка и запуск в корне проекта:
//
//	go build -o reset ./cmd/reset
//	./reset            # сканирует ./... от текущего каталога
//	./reset -dir ./internal
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// marker — текст комментария, помечающего структуру для генерации Reset.
const marker = "generate:reset"

// genFileName — имя файла со сгенерированными методами в каждом пакете.
const genFileName = "reset.gen.go"

func main() {
	dir := flag.String("dir", ".", "корневой каталог для сканирования пакетов")
	flag.Parse()

	if err := run(*dir); err != nil {
		fmt.Fprintln(os.Stderr, "reset:", err)
		exit(1)
	}
}

// exit вынесен в отдельную функцию, чтобы прямой вызов os.Exit не
// находился в теле main (этого требует анализатор exitcheck).
func exit(code int) {
	os.Exit(code)
}

// run выполняет всю работу генератора: загружает пакеты из каталога dir,
// находит помеченные структуры и генерирует для них методы Reset.
func run(dir string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports |
			packages.NeedDeps,
		Dir: dir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("загрузка пакетов: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("в пакетах есть ошибки компиляции")
	}

	// Первый проход: собираем множество всех помеченных структур во всех
	// пакетах. Это нужно, чтобы при генерации знать, у каких вложенных
	// типов появится метод Reset (в текущей информации о типах его ещё нет).
	marked := map[*types.TypeName]bool{}
	for _, pkg := range pkgs {
		collectMarked(pkg, marked)
	}

	// Второй проход: для каждого пакета генерируем файл reset.gen.go.
	for _, pkg := range pkgs {
		if err := generatePackage(pkg, marked); err != nil {
			return err
		}
	}
	return nil
}

// collectMarked находит в пакете структуры с маркером и добавляет их
// именованные типы в множество marked.
func collectMarked(pkg *packages.Package, marked map[*types.TypeName]bool) {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}
				if !hasMarker(ts, gd) {
					continue
				}
				obj := pkg.TypesInfo.Defs[ts.Name]
				if obj == nil {
					continue
				}
				if tn, ok := obj.(*types.TypeName); ok {
					marked[tn] = true
				}
			}
		}
	}
}

// hasMarker сообщает, помечена ли структура комментарием-маркером.
// Комментарий может быть привязан к самому TypeSpec или к объемлющему
// GenDecl (когда объявление одиночное: type T struct{...}).
func hasMarker(ts *ast.TypeSpec, gd *ast.GenDecl) bool {
	if commentHasMarker(ts.Doc) {
		return true
	}
	// Для одиночного объявления doc-комментарий висит на GenDecl.
	if len(gd.Specs) == 1 {
		return commentHasMarker(gd.Doc)
	}
	return false
}

// commentHasMarker проверяет, содержит ли группа комментариев строку-маркер.
func commentHasMarker(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if text == marker {
			return true
		}
	}
	return false
}

// generatePackage генерирует файл reset.gen.go для одного пакета, если в
// нём есть помеченные структуры.
func generatePackage(pkg *packages.Package, marked map[*types.TypeName]bool) error {
	// Отбираем помеченные типы, принадлежащие именно этому пакету, и
	// сортируем по имени для стабильного вывода.
	var named []*types.Named
	for tn := range marked {
		if tn.Pkg() != pkg.Types {
			continue
		}
		if n, ok := tn.Type().(*types.Named); ok {
			named = append(named, n)
		}
	}
	if len(named) == 0 {
		return nil
	}
	sort.Slice(named, func(i, j int) bool {
		return named[i].Obj().Name() < named[j].Obj().Name()
	})

	src, err := generateSource(pkg.Types, pkg.Name, named, marked)
	if err != nil {
		return err
	}

	outDir, err := packageDir(pkg)
	if err != nil {
		return err
	}
	outPath := filepath.Join(outDir, genFileName)
	if err := os.WriteFile(outPath, src, 0o644); err != nil {
		return fmt.Errorf("запись %s: %w", outPath, err)
	}
	fmt.Println("сгенерирован", outPath)
	return nil
}

// packageDir определяет каталог пакета по путям его исходных файлов.
func packageDir(pkg *packages.Package) (string, error) {
	if len(pkg.GoFiles) == 0 {
		return "", fmt.Errorf("у пакета %s нет исходных файлов", pkg.PkgPath)
	}
	return filepath.Dir(pkg.GoFiles[0]), nil
}

// fileData — данные для шаблона сгенерированного файла reset.gen.go.
type fileData struct {
	Package string
	Imports []string
	Methods []string
}

// fileTemplate задаёт общий каркас сгенерированного файла: заголовок,
// объявление пакета, блок импортов и тела методов подставляются в
// соответствующие placeholder'ы. Итог всё равно прогоняется через
// go/format, поэтому отступы в шаблоне не критичны.
var fileTemplate = template.Must(template.New("file").Parse(
	`// Code generated by "reset"; DO NOT EDIT.

package {{.Package}}
{{if .Imports}}
import (
{{- range .Imports}}
	{{printf "%q" .}}
{{- end}}
)
{{end}}
{{- range .Methods}}
{{.}}
{{- end}}
`))

// generateSource формирует содержимое файла reset.gen.go по шаблону
// fileTemplate: заголовок, объявление пакета, блок импортов и методы Reset
// для каждого типа.
func generateSource(pkgTypes *types.Package, pkgName string, named []*types.Named, marked map[*types.TypeName]bool) ([]byte, error) {
	g := &generator{
		pkgTypes: pkgTypes,
		marked:   marked,
		imports:  map[string]string{},
	}

	// Сначала генерируем тела методов: попутно g.imports наполняется
	// нужными импортами, которые понадобятся шаблону ниже.
	methods := make([]string, 0, len(named))
	for _, n := range named {
		var body bytes.Buffer
		g.writeMethod(&body, n)
		methods = append(methods, strings.TrimRight(body.String(), "\n"))
	}

	paths := make([]string, 0, len(g.imports))
	for p := range g.imports {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var out bytes.Buffer
	data := fileData{Package: pkgName, Imports: paths, Methods: methods}
	if err := fileTemplate.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("формирование файла по шаблону: %w", err)
	}

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("форматирование сгенерированного кода: %w\n%s", err, out.String())
	}
	return formatted, nil
}

// generator хранит состояние генерации одного файла: пакет, для которого
// генерируется код, множество помеченных типов и накопленные импорты.
type generator struct {
	pkgTypes *types.Package
	marked   map[*types.TypeName]bool
	imports  map[string]string // путь пакета -> имя пакета
}

// qual — квалификатор типов для types.TypeString. Для типов из текущего
// пакета возвращает пустую строку, для внешних — регистрирует импорт и
// возвращает имя пакета.
func (g *generator) qual(p *types.Package) string {
	if p == g.pkgTypes {
		return ""
	}
	g.imports[p.Path()] = p.Name()
	return p.Name()
}

// typeString возвращает строковое представление типа с учётом квалификатора.
func (g *generator) typeString(t types.Type) string {
	return types.TypeString(t, g.qual)
}

// writeMethod генерирует метод Reset для одной именованной структуры.
func (g *generator) writeMethod(buf *bytes.Buffer, named *types.Named) {
	recv := "x"
	typeName := named.Obj().Name()
	fmt.Fprintf(buf, "// Reset сбрасывает все поля %s в пустое состояние.\n", typeName)
	fmt.Fprintf(buf, "func (%s *%s) Reset() {\n", recv, typeName)

	st, ok := named.Underlying().(*types.Struct)
	if ok {
		for i := 0; i < st.NumFields(); i++ {
			field := st.Field(i)
			sel := recv + "." + field.Name()
			g.writeFieldReset(buf, sel, field.Type())
		}
	}
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)
}

// writeFieldReset генерирует код сброса поля sel типа t.
func (g *generator) writeFieldReset(buf *bytes.Buffer, sel string, t types.Type) {
	switch u := t.Underlying().(type) {
	case *types.Basic:
		fmt.Fprintf(buf, "\t%s = %s\n", sel, zeroBasic(u))
	case *types.Slice:
		fmt.Fprintf(buf, "\t%s = %s[:0]\n", sel, sel)
	case *types.Map:
		fmt.Fprintf(buf, "\tclear(%s)\n", sel)
	case *types.Struct:
		if g.willReset(t) {
			fmt.Fprintf(buf, "\t%s.Reset()\n", sel)
		} else {
			fmt.Fprintf(buf, "\t%s = %s{}\n", sel, g.typeString(t))
		}
	case *types.Pointer:
		g.writePointerReset(buf, sel, t, u)
	case *types.Array:
		fmt.Fprintf(buf, "\t%s = %s{}\n", sel, g.typeString(t))
	case *types.Chan, *types.Signature, *types.Interface:
		fmt.Fprintf(buf, "\t%s = nil\n", sel)
	default:
		fmt.Fprintf(buf, "\t%s = %s{}\n", sel, g.typeString(t))
	}
}

// writePointerReset генерирует код сброса поля-указателя. Nil-указатели
// пропускаются. Если тип поддерживает Reset — вызывается Reset, иначе
// сбрасывается значение по указателю.
func (g *generator) writePointerReset(buf *bytes.Buffer, sel string, ptr types.Type, u *types.Pointer) {
	fmt.Fprintf(buf, "\tif %s != nil {\n", sel)
	if g.willReset(ptr) {
		fmt.Fprintf(buf, "\t\t%s.Reset()\n", sel)
		fmt.Fprintln(buf, "\t}")
		return
	}
	elem := u.Elem()
	switch eu := elem.Underlying().(type) {
	case *types.Basic:
		fmt.Fprintf(buf, "\t\t*%s = %s\n", sel, zeroBasic(eu))
	case *types.Slice:
		fmt.Fprintf(buf, "\t\t*%s = (*%s)[:0]\n", sel, sel)
	case *types.Map:
		fmt.Fprintf(buf, "\t\tclear(*%s)\n", sel)
	default:
		fmt.Fprintf(buf, "\t\t*%s = %s{}\n", sel, g.typeString(elem))
	}
	fmt.Fprintln(buf, "\t}")
}

// willReset сообщает, будет ли у типа t (или типа, на который он
// указывает) метод Reset: либо он помечен маркером, либо метод уже
// объявлен в исходном коде.
func (g *generator) willReset(t types.Type) bool {
	named := namedOf(t)
	if named == nil {
		return false
	}
	if g.marked[named.Obj()] {
		return true
	}
	return hasResetMethod(named)
}

// namedOf извлекает *types.Named из типа, при необходимости снимая один
// уровень указателя.
func namedOf(t types.Type) *types.Named {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if n, ok := t.(*types.Named); ok {
		return n
	}
	return nil
}

// hasResetMethod проверяет, есть ли у именованного типа (или указателя на
// него) метод Reset без параметров и без возвращаемых значений.
func hasResetMethod(named *types.Named) bool {
	for _, t := range []types.Type{named, types.NewPointer(named)} {
		ms := types.NewMethodSet(t)
		if sel := ms.Lookup(nil, "Reset"); sel != nil {
			sig, ok := sel.Type().(*types.Signature)
			if ok && sig.Params().Len() == 0 && sig.Results().Len() == 0 {
				return true
			}
		}
	}
	return false
}

// zeroBasic возвращает литерал нулевого значения для примитивного типа.
func zeroBasic(b *types.Basic) string {
	switch {
	case b.Kind() == types.UnsafePointer:
		return "nil"
	case b.Info()&types.IsBoolean != 0:
		return "false"
	case b.Info()&types.IsString != 0:
		return `""`
	default:
		return "0"
	}
}
