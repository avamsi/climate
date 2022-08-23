package main

import (
	"encoding/gob"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/avamsi/eclipse"
	"github.com/avamsi/ergo"
	"golang.org/x/exp/maps"
)

func onlyElement[T any](s []T) T {
	if len(s) != 1 {
		log.Panicf("want: exactly 1 element; got: '%#v'", s)
	}
	return s[0]
}

func importPath(dir string) string {
	cmd := exec.Command("go", "list")
	cmd.Dir = dir
	return strings.TrimSpace(string(ergo.Must1(cmd.Output())))
}

func docFromField(field *ast.Field) string {
	doc := strings.TrimSpace(field.Doc.Text())
	if doc == "" {
		doc = strings.TrimSpace(field.Comment.Text())
	}
	return doc
}

func populateDocsForStructType(parentID string, t ast.Expr, docs map[string]string) {
	if st, ok := t.(*ast.StructType); ok {
		for _, field := range st.Fields.List {
			for _, name := range field.Names {
				docs[parentID+"."+name.Name] = docFromField(field)
			}
		}
	}
}

func populateDocs(dir string, docs map[string]string) {
	pkgs := ergo.Must1(parser.ParseDir(token.NewFileSet(), dir, nil, parser.ParseComments))
	if len(pkgs) == 0 {
		return
	}
	pkg := onlyElement(maps.Values(pkgs))
	pkgPath := "main"
	if pkg.Name != "main" {
		pkgPath = importPath(dir)
	}
	for _, docDotType := range doc.New(pkg, "TODO", doc.AllDecls).Types {
		parentID := pkgPath + "." + docDotType.Name
		docs[parentID] = strings.TrimSpace(docDotType.Doc)
		astDotType := onlyElement(docDotType.Decl.Specs).(*ast.TypeSpec).Type
		populateDocsForStructType(parentID, astDotType, docs)
		for _, method := range docDotType.Methods {
			parentID := parentID + "." + method.Name
			docs[parentID] = strings.TrimSpace(method.Doc)
			for _, param := range method.Decl.Type.Params.List {
				populateDocsForStructType(parentID, param.Type, docs)
			}
		}
	}
}

type Eclipse struct{}

func (Eclipse) Docs(opts struct{ Out string }) {
	docs := map[string]string{}
	cwd := ergo.Must1(os.Getwd())
	populateDocs(cwd, docs)
	filepath.WalkDir(cwd, func(path string, d fs.DirEntry, err error) error {
		ergo.Must0(err)
		if d.IsDir() {
			populateDocs(path, docs)
		}
		return nil
	})
	file := ergo.Must1(os.OpenFile(opts.Out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644))
	defer file.Close()
	ergo.Must0(gob.NewEncoder(file).Encode(docs))
}

func main() {
	eclipse.Execute(Eclipse{})
}
