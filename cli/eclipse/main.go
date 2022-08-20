package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/avamsi/ergo"
	"golang.org/x/exp/maps"
)

func onlyElement[T any](s []T) T {
	if l := len(s); l != 1 {
		log.Panicf("want: exactly 1 element; got: %d", l)
	}
	return s[0]
}

type packageT struct {
	ImportPath string
	Name       string
	Module     struct {
		Dir string
	}
}

const genCodeDefault = `// DO NOT EDIT (or do, whatever; I'm a sign, not a cop -- ¯\_(ツ)_/¯).
// Auto-generated by github.com/avamsi/eclipse/cli/eclipse.

package main

import (
	"github.com/avamsi/eclipse"
)

const _github_io_avamsi_eclipse_cli_docs = "e30="

func main() {
	eclipse.Execute(_github_io_avamsi_eclipse_cli_docs)
}
`

var (
	docsRe    = regexp.MustCompile(`const _github_io_avamsi_eclipse_cli_docs = (".*")`)
	eclipseRe = regexp.MustCompile(`eclipse.Execute\((.*)\)`)
)

var cmd = flag.String("cmd", "", "TODO")

func populateDocs(pkg packageT, docs map[string]string) {
	cwd := ergo.Check1(os.Getwd())
	astPkg := onlyElement(maps.Values(ergo.Check1(parser.ParseDir(token.NewFileSet(), cwd, nil, parser.ParseComments))))
	if astPkg.Name != pkg.Name {
		log.Panicf("want: %s; got: %s", pkg.Name, astPkg.Name)
	}
	for _, tipe := range doc.New(astPkg, pkg.ImportPath, doc.AllDecls).Types {
		if tipe.Name != *cmd {
			continue
		}
		parentID := pkg.ImportPath + "." + *cmd
		docs[parentID] = strings.TrimSpace(tipe.Doc)
		for _, field := range onlyElement(tipe.Decl.Specs).(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
			docs[parentID+"."+onlyElement(field.Names).Name] = strings.TrimSpace(field.Doc.Text())
		}
		for _, method := range tipe.Methods {
			parentID := parentID + "." + method.Name
			docs[parentID] = strings.TrimSpace(method.Doc)
			for _, param := range method.Decl.Type.Params.List {
				if paramType, ok := param.Type.(*ast.StructType); ok {
					for _, field := range paramType.Fields.List {
						docs[parentID+"."+onlyElement(field.Names).Name] = strings.TrimSpace(field.Doc.Text())
					}
				}
			}
		}
	}
}

func updateDocs(pkg packageT, genCode string) string {
	docs := make(map[string]string)
	rawDocs := ergo.Check1(strconv.Unquote(docsRe.FindStringSubmatch(genCode)[1]))
	ergo.Check0(json.Unmarshal(ergo.Check1(base64.StdEncoding.DecodeString(rawDocs)), &docs))
	populateDocs(pkg, docs)
	rawDocs = base64.StdEncoding.EncodeToString(ergo.Check1(json.Marshal(docs)))
	return docsRe.ReplaceAllLiteralString(genCode, fmt.Sprintf("const _github_io_avamsi_eclipse_cli_docs = %#v", rawDocs))
}

func updateEclipseParams(pkg packageT, genCode string) string {
	param := *cmd + "{}"
	if pkg.Name != "main" {
		param = pkg.Name + "." + param
		if !strings.Contains(genCode, pkg.ImportPath) {
			genCode = strings.Replace(genCode, "import (", fmt.Sprintf("import (\n\t\"%s\"", pkg.ImportPath), 1)
		}
	}
	eclipseParams := eclipseRe.FindStringSubmatch(genCode)[1]
	if !strings.Contains(eclipseParams, param) {
		eclipseParams = fmt.Sprintf("%s, %s", eclipseParams, param)
		genCode = eclipseRe.ReplaceAllLiteralString(genCode, fmt.Sprintf("eclipse.Execute(%s)", eclipseParams))
	}
	return genCode
}

func main() {
	flag.Parse()
	var pkg packageT
	json.Unmarshal(ergo.Check1(exec.Command("go", "list", "-json").Output()), &pkg)
	// TODO: is "main" really this special?
	if pkg.Name == "main" {
		pkg.ImportPath = "main"
	}
	genPath := filepath.Join(pkg.Module.Dir, "main.go")
	genCode := genCodeDefault
	{
		genFile, err := os.Open(genPath)
		if err == nil {
			defer genFile.Close()
			genCode = string(ergo.Check1(io.ReadAll(genFile)))
		} else {
			if !errors.Is(err, os.ErrNotExist) {
				panic(err)
			}
		}
	}
	genCode = updateEclipseParams(pkg, updateDocs(pkg, genCode))
	{
		genFile := ergo.Check1(os.OpenFile(genPath, os.O_WRONLY|os.O_CREATE, 0644))
		defer genFile.Close()
		ergo.Check0(genFile.Truncate(0))
		ergo.Check1(genFile.Seek(0, 0))
		ergo.Check1(genFile.WriteString(genCode))
	}
	ergo.Check0(exec.Command("go", "get", "github.com/avamsi/eclipse").Run())
	ergo.Check0(exec.Command("go", "mod", "tidy").Run())
}
