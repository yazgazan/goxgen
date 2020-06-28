package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/goxgen"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func _() {
	// Ensures we are using patched go/... packages
	goxgen.Goxgen()
}

const (
	goxExtension = ".gox"
)

func main() {
	var goxTargetPackage string

	flag.StringVar(&goxTargetPackage, "target-package", "gox", "package to use for html generation (i.e gox, vecty)")
	flag.Parse()

	if len(os.Args) > 1 {
		for _, dir := range flag.Args() {
			Transpile(goxTargetPackage, dir)
		}
	} else {
		Transpile("vecty", "goxtests")
	}
}

type GoxTranspiler struct {
	cfg  *printer.Config
	fset *token.FileSet

	files map[string]*ast.File
}

func (g *GoxTranspiler) TranspileFile(path string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	name := f.Name()

	if f.IsDir() || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, goxExtension) {
		return nil
	}

	fmt.Printf("Transpiling %s\n", path)

	file, err := parser.ParseFile(g.fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Can't parse file", err)
		return err
	}

	// fmt.Println(file.Name.Name) // package name

	// var fn VisitorFunc

	// fn = func(node ast.Node) (w ast.Visitor) {
	// 	goxexpr, ok := node.(*ast.GoxExpr)
	// 	if !ok {
	// 		return fn
	// 	}

	// 	if !ast.IsGoxComponent(goxexpr.TagName) {
	// 		return fn
	// 	}

	// 	fmt.Printf("%#v\n", goxexpr.TagName)
	// 	return fn
	// }

	// ast.Walk(fn, file)

	ofname := path[:len(path)-1] // lol
	g.files[ofname] = file

	// // cfg.Fprint(os.Stdout, fset, file)
	of, err := os.Create(ofname)
	g.cfg.Fprint(of, g.fset, file)

	if err != nil {
		fmt.Printf("Failed with error: %v", err)
		log.Fatalf("ParseFile(%s): %v", name, err)
		return err
	}

	return nil
}

func Transpile(goxTargetPackage, directory string) {
	folders, err := ListFolders(directory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	goxT := &GoxTranspiler{
		cfg: &printer.Config{
			Mode:             printer.GoxToGo | printer.RawFormat,
			GoxTargetPackage: goxTargetPackage,
		},
		fset:  token.NewFileSet(),
		files: map[string]*ast.File{},
	}

	for _, path := range folders {
		if filepath.IsAbs(path) {
			path = "./" + filepath.Join(".", path)
		}

		err = WalkFiles(path, goxT.TranspileFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	goxT.CheckTags()
}

func ListFolders(path string) ([]string, error) {
	m := map[string]struct{}{}

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if filepath.Ext(path) != goxExtension {
			return nil
		}

		m[filepath.Dir(path)] = struct{}{}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ss := make([]string, 0, len(m))
	for p := range m {
		ss = append(ss, p)
	}

	return ss, nil
}

func WalkFiles(path string, fn filepath.WalkFunc) error {
	d, err := os.Open(path)
	if err != nil {
		return fn("", nil, err)
	}
	defer d.Close()

	ffi, err := d.Readdir(-1)
	if err != nil {
		return fn("", nil, err)
	}

	for _, fi := range ffi {
		if fi.IsDir() || filepath.Ext(fi.Name()) != ".gox" {
			continue
		}

		err = fn(filepath.Join(path, fi.Name()), fi, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
