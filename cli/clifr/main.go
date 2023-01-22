package main

import (
	"encoding/gob"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/avamsi/clifr"
	"github.com/avamsi/ergo"
	"github.com/sanity-io/litter"
	"golang.org/x/exp/maps"
)

func onlyElement[T any](s []T) T {
	if len(s) == 1 {
		return s[0]
	}
	panic(fmt.Sprintf("want exactly 1 element, got %s", litter.Sdump(s)))
}

func importPath(dir string) string {
	cmd := exec.Command("go", "list")
	cmd.Dir = dir
	return strings.TrimSpace(string(ergo.Must1(cmd.Output())))
}

func docFromField(f *ast.Field) string {
	d := strings.TrimSpace(f.Doc.Text())
	if d == "" {
		d = strings.TrimSpace(f.Comment.Text())
	}
	return d
}

type doc struct {
	Long, Short, Usage string
}

const (
	shortDirective = "//clifr:short"
	usageDirective = "//clifr:usage"
)

func newDoc(cg *ast.CommentGroup) doc {
	if cg == nil {
		return doc{}
	}
	short, usage := "", ""
	for _, c := range cg.List {
		if strings.HasPrefix(c.Text, shortDirective) {
			if short == "" {
				short = strings.TrimSpace(strings.TrimPrefix(c.Text, shortDirective))
			} else {
				panic(fmt.Sprintf("want exactly 1 %s directive, got %s", shortDirective, litter.Sdump(cg)))
			}
		} else if strings.HasPrefix(c.Text, usageDirective) {
			if usage == "" {
				usage = strings.TrimSpace(strings.TrimPrefix(c.Text, usageDirective))
			} else {
				panic(fmt.Sprintf("want exactly 1 %s directive, got %s", usageDirective, litter.Sdump(cg)))
			}
		}
	}
	return doc{Long: strings.TrimSpace(cg.Text()), Short: short, Usage: usage}
}

func populateDocsForType(pid string, t ast.Expr, docs map[string]doc) {
	if st, ok := t.(*ast.StructType); ok {
		for _, f := range st.Fields.List {
			for _, n := range f.Names {
				id := pid + "." + n.Name
				docs[id] = doc{Long: docFromField(f)}
			}
		}
	}
}

func populateDocs(dir string, docs map[string]doc) {
	pkgs := ergo.Must1(parser.ParseDir(token.NewFileSet(), dir, nil, parser.ParseComments))
	if len(pkgs) == 0 {
		return
	}
	pkg := onlyElement(maps.Values(pkgs))
	pkgPath := "main"
	if pkg.Name != "main" {
		pkgPath = importPath(dir)
	}

	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			switch dt := d.(type) {
			case *ast.GenDecl:
				for _, s := range dt.Specs {
					switch st := s.(type) {
					case *ast.TypeSpec:
						id := pkgPath + "." + st.Name.Name
						docs[id] = newDoc(dt.Doc)
						populateDocsForType(id, st.Type, docs)
					}
				}
			case *ast.FuncDecl:
				if dt.Recv == nil {
					// We're only interested in methods for now.
					continue
				}
				pid := pkgPath + "." + onlyElement(dt.Recv.List).Type.(*ast.Ident).Name
				id := pid + "." + dt.Name.Name
				docs[id] = newDoc(dt.Doc)
				for _, p := range dt.Type.Params.List {
					populateDocsForType(id, p.Type, docs)
				}
			}

		}
	}
}

type Clifr struct{}

func (Clifr) Docs(opts struct{ Out string }) {
	docs := map[string]doc{}
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
	clifr.Execute(Clifr{})
}
