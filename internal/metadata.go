package internal

import (
	"bytes"
	"encoding/gob"
	"go/ast"
	"reflect"
	"strings"
	"unicode"

	"github.com/avamsi/ergo"
	"github.com/avamsi/ergo/assert"
	"github.com/sanity-io/litter"
)

// Note: it's important that all fields of RawMetadata be exported, otherwise
// gob won't be able to encode / decode them correctly.
type RawMetadata struct {
	Doc        string
	Directives map[string]string
	Comment    string
	Params     []string
	Children   map[string]*RawMetadata
}

func DecodeAsRawMetadata(b []byte) *RawMetadata {
	var rmd RawMetadata
	assert.Nil(gob.NewDecoder(bytes.NewReader(b)).Decode(&rmd))
	return &rmd
}

const directivePrefix = "//cli:"

func (rmd *RawMetadata) SetDoc(doc *ast.CommentGroup) {
	if doc == nil {
		return
	}
	rmd.Doc = strings.TrimSpace(doc.Text())
	rmd.Directives = map[string]string{}
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, directivePrefix) {
			continue
		}
		d, value, _ := strings.Cut(comment.Text, " ")
		d = strings.TrimPrefix(d, directivePrefix)
		if _, ok := rmd.Directives[d]; ok {
			ergo.Panicf("more than one %v directive: %v", d, litter.Sdump(doc))
		}
		rmd.Directives[d] = strings.TrimSpace(value)
	}
}

func (rmd *RawMetadata) SetComment(comment *ast.CommentGroup) {
	rmd.Comment = strings.TrimSpace(comment.Text())
}

func (rmd *RawMetadata) Child(name string) *RawMetadata {
	if rmd.Children == nil {
		rmd.Children = map[string]*RawMetadata{}
	}
	child, ok := rmd.Children[name]
	if !ok {
		child = &RawMetadata{}
		rmd.Children[name] = child
	}
	return child
}

func (rmd *RawMetadata) Encode() []byte {
	var b bytes.Buffer
	assert.Nil(gob.NewEncoder(&b).Encode(rmd))
	return b.Bytes()
}

type Metadata struct {
	root     *Metadata
	raw      *RawMetadata
	children map[string]*Metadata
}

func DecodeAsMetadata(b []byte) *Metadata {
	md := &Metadata{raw: DecodeAsRawMetadata(b)}
	md.root = md
	return md
}

func (md *Metadata) Lookup(pkgPath, name string) *Metadata {
	if md == nil {
		return nil
	}
	return md.root.Child(pkgPath).Child(name)
}

func (md *Metadata) LookupType(t reflect.Type) *Metadata {
	return md.Lookup(t.PkgPath(), t.Name())
}

func (md *Metadata) Aliases() []string {
	if md == nil {
		return nil
	}
	var (
		tmp     = strings.Split(md.raw.Directives["aliases"], ",")
		aliases []string
	)
	for _, alias := range tmp {
		if alias := strings.TrimSpace(alias); alias != "" {
			aliases = append(aliases, alias)
		}
	}
	return aliases
}

func (md *Metadata) Long() string {
	if md == nil {
		return ""
	}
	return md.raw.Doc
}

func (md *Metadata) Short() string {
	if md == nil {
		return ""
	}
	if short, ok := md.raw.Directives["short"]; ok {
		return short
	}
	if md.raw.Comment != "" {
		return md.raw.Comment
	}
	// Auto generate a short description from the long description.
	var (
		long = md.Long()
		i    = strings.Index(long, "\n\n")
	)
	if i != -1 {
		long = long[:i]
	}
	long = strings.Join(strings.Fields(long), " ")
	if long == "" {
		return ""
	}
	rs := []rune(long)
	rs[0] = unicode.ToUpper(rs[0])
	if l := len(rs); l > 80 {
		rs = append(rs[:77], []rune("...")...)
	} else if rs[l-1] == '.' && !strings.HasSuffix(long, "..") {
		// Clip the period at the end, by (Cobra's) convention.
		rs = rs[:l-1]
	}
	return string(rs)
}

func (md *Metadata) Usage(name string, args []ParamType) string {
	if md == nil {
		return strings.ToLower(name)
	}
	if usage, ok := md.raw.Directives["usage"]; ok {
		return usage
	}
	return strings.ToLower(name) + ParamsUsage(md.raw.Params, args)
}

func (md *Metadata) Child(name string) *Metadata {
	if md == nil {
		return nil
	}
	if md.children == nil {
		md.children = map[string]*Metadata{}
	}
	child, ok := md.children[name]
	if !ok {
		child = &Metadata{root: md.root, raw: md.raw.Child(name)}
		md.children[name] = child
	}
	return child
}
