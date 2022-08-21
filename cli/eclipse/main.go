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
	if l := len(s); l != 1 {
		log.Panicf("want: exactly 1 element; got: %d", l)
	}
	return s[0]
}

func importPath(dir string) string {
	cmd := exec.Command("go", "list")
	cmd.Dir = dir
	return strings.TrimSpace(string(ergo.Check1(cmd.Output())))
}

func docFromField(field *ast.Field) string {
	doc := strings.TrimSpace(field.Doc.Text())
	if doc == "" {
		doc = strings.TrimSpace(field.Comment.Text())
	}
	return doc
}

func populateDocs(dir string, docs map[string]string) {
	pkg := onlyElement(maps.Values(ergo.Check1(parser.ParseDir(token.NewFileSet(), dir, nil, parser.ParseComments))))
	pkgPath := "main"
	if pkg.Name != "main" {
		pkgPath = importPath(dir)
	}
	for _, tipe := range doc.New(pkg, "TODO", doc.AllDecls).Types {
		parentID := pkgPath + "." + tipe.Name
		docs[parentID] = strings.TrimSpace(tipe.Doc)
		for _, field := range onlyElement(tipe.Decl.Specs).(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
			docs[parentID+"."+onlyElement(field.Names).Name] = docFromField(field)
		}
		for _, method := range tipe.Methods {
			parentID := parentID + "." + method.Name
			docs[parentID] = strings.TrimSpace(method.Doc)
			for _, param := range method.Decl.Type.Params.List {
				if paramType, ok := param.Type.(*ast.StructType); ok {
					for _, field := range paramType.Fields.List {
						docs[parentID+"."+onlyElement(field.Names).Name] = docFromField(field)
					}
				}
			}
		}
	}
}

type Eclipse struct{}

func (Eclipse) Docs(opts struct{ Out string }) {
	docs := map[string]string{}
	cwd := ergo.Check1(os.Getwd())
	populateDocs(cwd, docs)
	filepath.WalkDir(cwd, func(path string, d fs.DirEntry, err error) error {
		ergo.Check0(err)
		if d.IsDir() {
			populateDocs(path, docs)
		}
		return nil
	})
	file := ergo.Check1(os.OpenFile(opts.Out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644))
	defer file.Close()
	ergo.Check0(gob.NewEncoder(file).Encode(docs))
}

func main() {
	eclipse.Execute(Eclipse{})
}
