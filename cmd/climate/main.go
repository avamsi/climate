package main

import (
	"errors"
	"go/ast"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	_ "embed"

	"github.com/avamsi/ergo/assert"
	"github.com/sanity-io/litter"
	"golang.org/x/tools/go/packages"

	"github.com/avamsi/climate"
	"github.com/avamsi/climate/internal"
)

func parseFunc(f *ast.FuncDecl, pkgMd *internal.RawMetadata) {
	parentMd := pkgMd
	if f.Recv != nil {
		assert.Truef(len(f.Recv.List) == 1,
			"not exactly one receiver: %s", litter.Sdump(f.Recv.List))
		recv := f.Recv.List[0]
		// We only support pointer receivers, skip others.
		e, ok := recv.Type.(*ast.StarExpr)
		if !ok {
			return
		}
		parentMd = pkgMd.Child(e.X.(*ast.Ident).Name)
	}
	md := parentMd.Child(f.Name.Name)
	md.SetDoc(f.Doc)
	for _, param := range f.Type.Params.List {
		for _, n := range param.Names {
			md.Params = append(md.Params, n.Name)
		}
	}
}

func parseType(g *ast.GenDecl, pkgMd *internal.RawMetadata) {
	for _, spec := range g.Specs {
		spec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		s, ok := spec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		structMd := pkgMd.Child(spec.Name.Name)
		structMd.SetDoc(g.Doc)
		for _, f := range s.Fields.List {
			for _, n := range f.Names {
				md := structMd.Child(n.Name)
				md.SetDoc(f.Doc)
				md.SetComment(f.Comment)
			}
		}
	}
}

func parsePkg(pkg *packages.Package, rootMd *internal.RawMetadata) {
	pkgPath := pkg.PkgPath
	if pkg.Name == "main" {
		pkgPath = "main"
	}
	pkgMd := rootMd.Child(pkgPath)
	for node := range pkg.TypesInfo.Scopes {
		file, ok := node.(*ast.File)
		if !ok {
			continue
		}
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				parseFunc(decl, pkgMd)
			case *ast.GenDecl:
				parseType(decl, pkgMd)
			}
		}
	}
}

func pkgDir(pkg *packages.Package) string {
	if len(pkg.GoFiles) > 0 {
		return filepath.Dir(pkg.GoFiles[0])
	}
	return ""
}

type parseOptions struct {
	Out   string // output file to write metadata to
	Debug bool   // whether to print metadata
}

// climate recursively parses the metadata of all Go packages in the current
// directory and its subdirectories, and writes it to the given output file.
//
//climate:usage climate [opts]
func parse(opts *parseOptions) {
	var (
		rootMd internal.RawMetadata
		mode   = (packages.NeedName | packages.NeedFiles |
			packages.NeedTypes | packages.NeedTypesInfo)
		cfg     = &packages.Config{Mode: mode}
		pkgs    = assert.Ok(packages.Load(cfg, "./..."))
		rootDir = assert.Ok(filepath.Abs(assert.Ok(os.Getwd())))
	)
	for _, pkg := range pkgs {
		if pkg.Name == "main" && pkgDir(pkg) != rootDir {
			// Skip non-root main packages.
			continue
		}
		parsePkg(pkg, &rootMd)
	}
	if opts.Debug {
		litter.Dump(rootMd)
	}
	// gob encoding is not deterministic with maps, which means the encoded
	// metadata (and hence the output file) will keep changing on each
	// invocation even if nothing about the metadata actually changed, which is
	// not ideal (causes VCS noise, for example). To avoid this, we compare the
	// new metadata with the old one and write only if they are different.
	oldEncoded, err := os.ReadFile(opts.Out)
	if !errors.Is(err, fs.ErrNotExist) { // if exists
		assert.Nil(err)
		oldDecoded := internal.DecodeAsRawMetadata(oldEncoded)
		if reflect.DeepEqual(rootMd, *oldDecoded) {
			return
		}
	}
	assert.Nil(os.WriteFile(opts.Out, rootMd.Encode(), 0o644))
}

//go:generate go run github.com/avamsi/climate/cmd/climate --out=md.climate
//go:embed md.climate
var md []byte

func main() {
	os.Exit(climate.Run(climate.Func(parse), climate.Metadata(md)))
}
