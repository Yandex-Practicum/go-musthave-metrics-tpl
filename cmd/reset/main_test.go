package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// typeCheck разбирает и проверяет типы исходного текста пакета,
// возвращая пакет типов и информацию о нём.
func typeCheck(t *testing.T, src string) (*types.Package, *types.Info, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "sample.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("разбор исходника: %v", err)
	}
	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("sample", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatalf("проверка типов: %v", err)
	}
	return pkg, info, file
}

// namedByName находит именованный тип по имени в пакете типов.
func namedByName(t *testing.T, pkg *types.Package, name string) *types.Named {
	t.Helper()
	obj := pkg.Scope().Lookup(name)
	if obj == nil {
		t.Fatalf("тип %s не найден", name)
	}
	n, ok := obj.Type().(*types.Named)
	if !ok {
		t.Fatalf("%s не является именованным типом", name)
	}
	return n
}

// TestGenerateSource проверяет, что для структуры со всеми поддерживаемыми
// категориями полей генерируются корректные строки сброса и что
// результат компилируется вместе с исходником.
func TestGenerateSource(t *testing.T) {
	src := `package sample

// Inner имеет собственный метод Reset и НЕ помечен маркером.
type Inner struct {
	v int
}

func (i *Inner) Reset() { i.v = 0 }

// generate:reset
type Sample struct {
	i     int
	str   string
	flag  bool
	strP  *string
	s     []int
	m     map[string]string
	inner Inner
	child *Sample
	ip    *int
}
`
	pkg, _, _ := typeCheck(t, src)
	sample := namedByName(t, pkg, "Sample")

	// Помечаем Sample как подлежащий генерации.
	marked := map[*types.TypeName]bool{sample.Obj(): true}

	out, err := generateSource(pkg, pkg.Name(), []*types.Named{sample}, marked)
	if err != nil {
		t.Fatalf("generateSource: %v", err)
	}
	got := string(out)

	wantLines := []string{
		"func (x *Sample) Reset() {",
		"x.i = 0",
		`x.str = ""`,
		"x.flag = false",
		"if x.strP != nil {",
		`*x.strP = ""`,
		"x.s = x.s[:0]",
		"clear(x.m)",
		"x.inner.Reset()",
		"if x.child != nil {",
		"x.child.Reset()",
		"if x.ip != nil {",
		"*x.ip = 0",
	}
	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Errorf("в сгенерированном коде нет строки %q\n--- код ---\n%s", want, got)
		}
	}

	// Сгенерированный код должен компилироваться вместе с исходником.
	combined := src + "\n" + stripHeaderAndPackage(got)
	if _, _, _ = typeCheck(t, combined); t.Failed() {
		t.Logf("объединённый код:\n%s", combined)
	}
}

// stripHeaderAndPackage убирает из сгенерированного файла строку package и
// заголовок, оставляя только методы, чтобы приклеить их к исходнику для
// повторной проверки типов.
func stripHeaderAndPackage(gen string) string {
	lines := strings.Split(gen, "\n")
	var body []string
	skip := true
	for _, l := range lines {
		if skip {
			if strings.HasPrefix(l, "package ") {
				skip = false
			}
			continue
		}
		body = append(body, l)
	}
	return strings.Join(body, "\n")
}

// TestWillReset проверяет распознавание типов, у которых будет метод Reset.
func TestWillReset(t *testing.T) {
	src := `package sample

// generate:reset
type Marked struct{ a int }

type Manual struct{ b int }

func (m *Manual) Reset() { m.b = 0 }

type Plain struct{ c int }
`
	pkg, _, _ := typeCheck(t, src)
	marked := map[*types.TypeName]bool{
		namedByName(t, pkg, "Marked").Obj(): true,
	}
	g := &generator{pkgTypes: pkg, marked: marked, imports: map[string]string{}}

	cases := []struct {
		name string
		want bool
	}{
		{"Marked", true},
		{"Manual", true},
		{"Plain", false},
	}
	for _, c := range cases {
		n := namedByName(t, pkg, c.name)
		if got := g.willReset(n); got != c.want {
			t.Errorf("willReset(%s) = %v, ожидалось %v", c.name, got, c.want)
		}
		// Указатель на тип должен вести себя так же.
		if got := g.willReset(types.NewPointer(n)); got != c.want {
			t.Errorf("willReset(*%s) = %v, ожидалось %v", c.name, got, c.want)
		}
	}
}

// TestZeroBasic проверяет литералы нулевых значений примитивов.
func TestZeroBasic(t *testing.T) {
	src := `package sample
type T struct {
	i int
	s string
	b bool
	f float64
}
`
	pkg, _, _ := typeCheck(t, src)
	st := namedByName(t, pkg, "T").Underlying().(*types.Struct)
	want := map[string]string{"i": "0", "s": `""`, "b": "false", "f": "0"}
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		b := f.Type().Underlying().(*types.Basic)
		if got := zeroBasic(b); got != want[f.Name()] {
			t.Errorf("zeroBasic(%s) = %q, ожидалось %q", f.Name(), got, want[f.Name()])
		}
	}
}

// TestRunIntegration создаёт временный модуль с помеченной структурой,
// запускает генератор и проверяет, что файл reset.gen.go создан и содержит
// корректный метод.
func TestRunIntegration(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "go.mod"), "module samplemod\n\ngo 1.26\n")
	writeFile(t, filepath.Join(dir, "sample.go"), `package samplemod

// generate:reset
type Buffer struct {
	data []byte
	size int
	tags map[string]int
}
`)

	if err := run(dir); err != nil {
		t.Fatalf("run: %v", err)
	}

	genPath := filepath.Join(dir, genFileName)
	data, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("сгенерированный файл не создан: %v", err)
	}
	gen := string(data)
	for _, want := range []string{
		"func (x *Buffer) Reset() {",
		"x.data = x.data[:0]",
		"x.size = 0",
		"clear(x.tags)",
		"DO NOT EDIT.",
	} {
		if !strings.Contains(gen, want) {
			t.Errorf("в файле нет %q\n--- файл ---\n%s", want, gen)
		}
	}
}

// TestRunNoMarked проверяет, что при отсутствии помеченных структур файл не
// создаётся.
func TestRunNoMarked(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module plainmod\n\ngo 1.26\n")
	writeFile(t, filepath.Join(dir, "plain.go"), `package plainmod

type NoMarker struct{ x int }
`)

	if err := run(dir); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, genFileName)); !os.IsNotExist(err) {
		t.Fatalf("файл reset.gen.go не должен был создаваться, err=%v", err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("запись %s: %v", path, err)
	}
}
