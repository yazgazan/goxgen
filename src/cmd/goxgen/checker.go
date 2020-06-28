package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"unsafe"

	"golang.org/x/tools/go/packages"
)

const (
	HTMLTypeName            = "HTML"
	ComponentOrHTMLTypeName = "ComponentOrHTML"
	MarkupListTypeName      = "MarkupList"
	ApplyerTypeName         = "Applyer"
)

func (g *GoxTranspiler) loadPackages() (pkgs map[string]*packages.Package, loaded map[string]string) {
	loaded = map[string]string{}
	pkgs = map[string]*packages.Package{}

	for path := range g.files {
		dir := filepath.Dir(path)
		if _, ok := loaded[dir]; ok {
			continue
		}

		pp, err := packages.Load(&packages.Config{
			Mode:       packages.NeedName | packages.NeedModule | packages.NeedTypes | packages.NeedImports,
			BuildFlags: []string{"-tags=gox"},
		}, "./"+dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load package %q: %v\n", path, err)
			continue
		}
		if len(pp) == 0 {
			continue
		}

		pkgs[pp[0].PkgPath] = pp[0]
		loaded[dir] = pp[0].PkgPath
	}

	return pkgs, loaded
}

func (g *GoxTranspiler) CheckTags() {
	pkgs, loaded := g.loadPackages()
	// Load packages
	// walk each files
	for filename, file := range g.files {
		file := file
		filePkg := pkgs[loaded[filepath.Dir(filename)]]

		fn := &Visitor{}
		fn.fn = func(goxexpr *ast.GoxExpr) (w ast.Visitor) {
			location := file.Tokens.Position(goxexpr.Otag).String()

			switch t := goxexpr.TagName.(type) {
			default:
				panic(fmt.Sprintf("unexpected tag type %T", t))
			case *ast.Ident: // local struct
				CheckStruct(location, filePkg, t.Name, goxexpr)
			case *ast.SelectorExpr:
				x, ok := t.X.(*ast.Ident)
				if ok {
					resolved := resolveSelector(pkgs, file.Imports, x.Name)
					if resolved != nil {
						// fmt.Printf("struct %q.%s\n", resolved.PkgPath, t.Sel.Name)
						CheckStruct(location, resolved, t.Sel.Name, goxexpr)
					}
				}
			case *ast.CallExpr:
				CheckCallExpr(location, file, filePkg, pkgs, t, goxexpr)
			}
			// fmt.Printf("%#v\n", goxexpr.TagName)
			return fn
		}

		ast.Walk(fn, file)
	}
}

func CheckCallExpr(location string, file *ast.File, filePkg *packages.Package, pkgs map[string]*packages.Package, callExpr *ast.CallExpr, goxexpr *ast.GoxExpr) {
	switch t2 := callExpr.Fun.(type) {
	default:
		panic(fmt.Sprintf("unexpected function type %T", t2))
	case *ast.Ident:
		CheckCall(location, filePkg, t2.Name, goxexpr)
	case *ast.SelectorExpr:
		x, ok := t2.X.(*ast.Ident)
		if ok {
			resolved := resolveSelector(pkgs, file.Imports, x.Name)
			if resolved != nil {
				CheckCall(location, resolved, t2.Sel.Name, goxexpr)
			} else {
				fmt.Fprintf(os.Stderr, "%s: Info: cannot resolve %s\n", location, x.Name)
			}
			return
		}
		panic(fmt.Sprintf("unexpected function type (.X) %T", t2.X))
	}
}

func CheckStruct(location string, pkg *packages.Package, name string, goxexpr *ast.GoxExpr) {
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		fmt.Fprintf(os.Stderr, "%s: Warning: %s not found in %q\n", location, name, pkg.PkgPath)
		return
	}
	structT, ok := Struct(obj.Type())
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: Warning: %s is not a struct\n", location, ObjName(obj))
	}
	fields := StructFields(structT)
	for _, attr := range goxexpr.Attrs {
		_, ok := fields[attr.Lhs.Name]
		if !ok {
			fmt.Fprintf(os.Stderr, "%s: Warning: unknown field %q in gox component of type %s\n", location, attr.Lhs.Name, ObjName(obj))
			continue
		}
	}

	if len(goxexpr.X) != 0 {
		bodyField, ok := fields["Body"]
		if !ok {
			fmt.Fprintf(os.Stderr, "%s: Warning: gox component of type %s does not accept a body\n", location, ObjName(obj))
		}

		named, ok := Named(bodyField)
		if !ok {
			fmt.Fprintf(os.Stderr, "%s: Warning: field Body (type %s) in gox component of type %s does not look like a body type\n", location, TypeName(bodyField), ObjName(obj))
			return
		}

		switch named.Obj().Name() {
		default:
			fmt.Fprintf(os.Stderr, "%s: Warning: field Body (type %s) in gox component of type %s does not look like a body type\n", location, TypeName(bodyField), ObjName(obj))
		case HTMLTypeName, ComponentOrHTMLTypeName:
		}
	}
}

func StructFields(t *types.Struct) map[string]types.Type {
	m := make(map[string]types.Type, t.NumFields())
	for i := 0; i < t.NumFields(); i++ {
		field := t.Field(i)
		m[field.Name()] = field.Type()
	}

	return m
}

func CheckCall(location string, pkg *packages.Package, name string, goxexpr *ast.GoxExpr) {
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		fmt.Fprintf(os.Stderr, "%s: Warning: %s not found in %q\n", location, name, pkg.PkgPath)
		return
	}
	funcObj, ok := obj.(*types.Func)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: Warning: %s is not a function\n", location, ObjName(obj))
		return
	}
	funcType := funcObj.Type().(*types.Signature)

	params := funcType.Params()

	if len(goxexpr.X) > 0 {
		if params.Len() == 0 {
			fmt.Fprintf(os.Stderr, "%s: Warning: function %s does not accept a body\n", location, ObjName(obj))
			return
		}

		bodyArg := params.At(params.Len() - 1)
		CheckCallBody(location, ObjName(obj), bodyArg.Type())
	}

	if len(goxexpr.X) > 0 && len(goxexpr.Attrs) > 0 {
		if params.Len() < 2 {
			fmt.Fprintf(os.Stderr, "%s: Warning: function %s does not accept attributes\n", location, ObjName(obj))
			return
		}

		attrsArg := params.At(params.Len() - 2)
		CheckCallAttrs(location, ObjName(obj), attrsArg.Type())
	} else if len(goxexpr.Attrs) > 0 {
		if params.Len() == 0 {
			fmt.Fprintf(os.Stderr, "%s: Warning: function %s does not accept attributes\n", location, ObjName(obj))
			return
		}

		attrsArg := params.At(params.Len() - 1)
		CheckCallAttrs(location, ObjName(obj), attrsArg.Type())
	}
}

func TypeName(typ types.Type) string {
	return types.TypeString(typ, func(pkg *types.Package) string {
		return pkg.Name()
	})
}

func ObjName(obj types.Object) string {
	return fmt.Sprintf("%s.%s", obj.Pkg().Name(), obj.Name())
}

func CheckCallBody(location, funcName string, typ types.Type) {
	named, ok := Named(typ)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: Warning: body argument of %s (type %s) does not look like a body type\n", location, funcName, TypeName(typ))
		return
	}

	switch named.Obj().Name() {
	default:
		fmt.Fprintf(os.Stderr, "%s: Warning: body argument of %s (type %s) does not look like a body type\n", location, funcName, TypeName(typ))
	case HTMLTypeName, ComponentOrHTMLTypeName:
	}
}

func CheckCallAttrs(location, funcName string, typ types.Type) {
	named, ok := Named(typ)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: Warning: attributes argument of %s (type %s) does not look like a body type\n", location, funcName, TypeName(typ))
		return
	}

	switch named.Obj().Name() {
	default:
		fmt.Fprintf(os.Stderr, "%s: Warning: attributes argument of %s (type %s) does not look like a body type\n", location, funcName, TypeName(typ))
	case MarkupListTypeName, ApplyerTypeName:
	}
}

func Named(typ types.Type) (*types.Named, bool) {
	switch t := typ.(type) {
	default:
		return nil, false
	case *types.Named:
		return t, true
	case *types.Pointer:
		return Named(t.Elem())
	}
}

func Struct(typ types.Type) (*types.Struct, bool) {
	switch t := typ.(type) {
	default:
		return nil, false
	case *types.Struct:
		return t, true
	case *types.Named:
		return Struct(t.Underlying())
	}
}

func resolveSelector(pkgs map[string]*packages.Package, imports []*ast.ImportSpec, name string) *packages.Package {
	for _, _import := range imports {
		if _import.Name != nil {
			if name == _import.Name.Name {
				p, err := strconv.Unquote(_import.Path.Value)
				if err != nil {
					panic(err)
				}

				return pkgs[p] // can be nil

			}
			continue
		}

		p, err := strconv.Unquote(_import.Path.Value)
		if err != nil {
			panic(err)
		}

		pkg, ok := pkgs[p]
		if !ok {
			continue
		}
		if name == pkg.Name {
			return pkg
		}
	}

	return nil
}

type Visitor struct {
	visited map[unsafe.Pointer]struct{}
	fn      func(node *ast.GoxExpr) (w ast.Visitor)
}

func (v *Visitor) Visit(node ast.Node) (w ast.Visitor) {
	if v.visited == nil {
		v.visited = map[unsafe.Pointer]struct{}{}
	}
	goxexpr, ok := node.(*ast.GoxExpr)
	if !ok {
		return v
	}

	ptr := unsafe.Pointer(goxexpr)
	if _, ok := v.visited[ptr]; ok {
		return v
	}
	v.visited[ptr] = struct{}{}

	if !ast.IsGoxComponent(goxexpr.TagName) {
		return v
	}

	return v.fn(goxexpr)
}

// type VisitorFunc func(node ast.Node) (w ast.Visitor)

// func (fn VisitorFunc) Visit(node ast.Node) (w ast.Visitor) {
// 	return fn(node)
// }
